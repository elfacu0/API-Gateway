package apy

import (
	"encoding/json"
	"fmt"
	"gateway/pkg/auth"
	"gateway/pkg/storage"
	"gateway/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type FormErrors struct {
	Name   string
	Url    string
	Method string
}

type ResError struct {
	Status  int
	Message string
}

type Metric struct {
	Url      string `json:"url"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Requests int    `json:"requests"`
}

// RateLimit affects all request and not only request from one IP
type Endpoint struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Url           string `json:"url"`
	Method        string `json:"method"`
	Cache         string `json:"cache"`
	EnableCache   bool   `json:"enable-cache"`
	EnableAuth    bool   `json:"enable-auth"`
	RateLimit     int    `json:"rate-limit"`
	LastTime      int    `json:"last-time"`
	Requests      int    `json:"requests"`
	TotalRequests int    `json:"total-requests"`
}

type Apy struct {
	App       *gin.Engine
	Endpoints map[string]*Endpoint `json:"endpoints"`
}

const RATE_LIMIT_DURATION = 15 * 60 * 60

func HandleRateLimit(endpoint *Endpoint) ResError {
	if endpoint.RateLimit > 0 && (endpoint.LastTime+RATE_LIMIT_DURATION) < int(time.Now().Unix()) {
		endpoint.ResetRequest()
	}

	if endpoint.RateLimit > 0 && endpoint.Requests >= endpoint.RateLimit {
		return ResError{Status: http.StatusTooManyRequests, Message: "Too many Request"}
	}

	return ResError{}
}

func HandleAuth(endpoint *Endpoint, c *gin.Context) ResError {
	if endpoint.EnableAuth {
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			return ResError{Status: http.StatusUnauthorized, Message: "Bearer Token cannot be empty"}
		}

		bearerToken := strings.Split(token, " ")
		if len(bearerToken) != 2 {
			return ResError{Status: http.StatusBadRequest, Message: "Bearer Token Is Invalid"}
		}

		token = bearerToken[1]

		err := auth.ValidJwtToken(token)
		return ResError{Status: http.StatusUnauthorized, Message: err.Error()}
	}
	return ResError{}
}

func MiddleWare(endpoints map[string]*Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(strings.Split(path, "/")) > 1 {
			path = "/" + strings.Split(path, "/")[1]
		}
		fmt.Println(path)
		method := c.Request.Method
		id := utils.ID(path, method)

		endpoint, ok := endpoints[id]

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Endpoint not found",
			})
			return
		}

		if err := HandleRateLimit(endpoint); err.Status != 0 {
			c.JSON(err.Status, gin.H{
				"error": err.Message,
			})
			return
		}

		if err := HandleAuth(endpoint, c); err.Status != 0 {
			c.JSON(err.Status, gin.H{
				"error": err.Message,
			})
			return
		}

		c.Set("ok", false)

		c.Next()

		endpoint.IncReqCounter()
	}
}

func (a *Apy) AddEndpoint(e Endpoint) {
	id := utils.ID(e.Path, e.Method)
	a.Endpoints[id] = &e
	a.App.Handle(e.Method, e.Path, func(c *gin.Context) {
		if _, ok := c.Get("ok"); !ok {
			return
		}

		if res, err := a.Fetch(c); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"body": res,
			})
		}
	})
}

func SaveEndpoint(e Endpoint) {
	b, err := json.Marshal(e)
	if err != nil {
		return
	}

	jsonEndpoint := string(b)
	id := utils.ID(e.Path, e.Method)
	storage.Save(id, jsonEndpoint)
}

func (a *Apy) LoadEnpoints() {
	keys, err := storage.Keys()
	if err != nil {
		return
	}

	for _, key := range keys {
		var endpoint Endpoint
		jsonData, err := storage.Load(key)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(jsonData), &endpoint)
		if err != nil {
			continue
		}
		if endpoint.Url != "" {
			a.AddEndpoint(endpoint)
		}
	}
}

func (a *Apy) DeleteEndpoint(path, method string) error {
	id := utils.ID(path, method)
	if err := storage.Delete(id); err != nil {
		return err
	}
	delete(a.Endpoints, id)
	return nil
}

func (a *Apy) Fetch(c *gin.Context) (string, error) {
	method, path := c.Request.Method, c.Request.URL.Path
	id := utils.ID(path, method)
	endpoint := a.Endpoints[id]

	url := endpoint.GetUrl()

	if endpoint.EnableCache && endpoint.Cache != "" {
		return endpoint.Cache, nil
	}

	var (
		res *http.Response
		err error
	)

	req, err := http.NewRequest(method, url, c.Request.Body)

	if err != nil {
		return "", err
	}

	client := &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if endpoint.EnableCache {
		endpoint.SetCache(string(body))
	}

	return string(body), nil
}

func (e *Endpoint) GetUrl() string {
	return e.Url
}

func (e *Endpoint) SetCache(body string) {
	e.Cache = body
	SaveEndpoint(*e)
}

func (e *Endpoint) ResetRequest() {
	e.Requests = 0
	e.LastTime = int(time.Now().Unix())
}

func (e *Endpoint) IncReqCounter() {
	e.Requests++
	e.TotalRequests++
	SaveEndpoint(*e)
}

func (a *Apy) EnableDelEndpoints() {
	path := "/delete"
	id := utils.ID(path, http.MethodDelete)
	a.Endpoints[id] = &Endpoint{Name: "Delete Endpoint", Path: path, Method: http.MethodDelete}

	a.App.DELETE("/delete/:method", func(c *gin.Context) {
		path := c.Param("path")
		method := c.Request.Method
		if err := a.DeleteEndpoint(path, method); err != nil {
			c.JSON(http.StatusTeapot, gin.H{
				"error": "Failed to Delete endpoint",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Endpoint Deleted Successfully",
			"path":    path,
		})
	})
}

func (a *Apy) EnableAuthEndpoint() {
	path := "/auth"
	id := utils.ID(path, http.MethodGet)
	a.Endpoints[id] = &Endpoint{Name: "Auth", Path: path, Method: http.MethodGet}

	a.App.GET(path, func(c *gin.Context) {
		token, err := auth.CreateJwtToken()
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, token)
	})
}

func (a *Apy) Run() {
	a.EnableAuthEndpoint()
	a.App.Run()
}

func (a *Apy) EnableMetrics() {
	path := "/metrics"
	id := utils.ID(path, http.MethodGet)
	a.Endpoints[id] = &Endpoint{Name: "Metrics", Path: path, Method: http.MethodGet}

	a.App.GET(path, func(c *gin.Context) {
		metrics := make(map[string]Metric)
		for _, endpoint := range a.Endpoints {
			metrics[endpoint.Name] = Metric{Url: endpoint.Url, Path: endpoint.Path, Method: endpoint.Method, Requests: endpoint.TotalRequests}
		}
		c.JSON(200, metrics)
	})
}

func ParseForm(c *gin.Context) (Endpoint, FormErrors) {
	errors := FormErrors{}
	name := c.PostForm("name")
	url := c.PostForm("url")
	method := c.PostForm("method")
	rateLimit, _ := strconv.Atoi(c.PostForm("rate-limit"))
	enableCache := c.PostForm("enable-cache") != ""
	enableAuth := c.PostForm("enable-auth") != ""
	path := "/" + utils.RandomStr()

	if name == "" {
		errors.Name = "Name cannot be empty."
	}
	if url == "" {
		errors.Url = "Url cannot be empty."
	}
	if method == "" {
		errors.Method = "Method cannot be empty. "
	}
	if method != http.MethodGet && method != http.MethodPost && method != http.MethodDelete {
		errors.Method += "Method not allowed. "
	}

	return Endpoint{Name: name, Path: path, Url: url, Method: method, RateLimit: rateLimit, EnableCache: enableCache, EnableAuth: enableAuth}, errors
}

func (a *Apy) EnableNewEndpoints() {
	path := "/add"
	id := utils.ID(path, http.MethodPost)
	a.Endpoints[id] = &Endpoint{Name: "Add Endpoint", Path: path, Method: http.MethodPost}

	a.App.POST(path, func(c *gin.Context) {
		endpoint, errors := ParseForm(c)
		if errors != (FormErrors{}) {
			c.JSON(http.StatusBadRequest, errors)
			return
		}
		a.AddEndpoint(endpoint)
		SaveEndpoint(endpoint)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Endpoint Created Successfully",
			"path":    endpoint.Path,
		})
	})
}

func (a *Apy) Init() {
	a.App = gin.Default()
	a.Endpoints = make(map[string]*Endpoint)
	a.App.Use(MiddleWare(a.Endpoints))
}
