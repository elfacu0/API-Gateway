package apy

import (
	"gateway/pkg/auth"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Metric struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	Requests int    `json:"requests"`
}

// RateLimit affects all request and not only request from one IP
type Endpoint struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Cache       string `json:"cache"`
	EnableCache bool   `json:"enable-cache"`
	EnableAuth  bool   `json:"enable-auth"`
	RateLimit   int    `json:"rate-limit"`
	LastTime    int    `json:"last-time"`
	Requests    int    `json:"requests"`
}

type Route struct {
	Name      string               `json:"name"`
	Path      string               `json:"path"`
	Endpoints map[string]*Endpoint `json:"endpoints"`
}

type Apy struct {
	App   *gin.Engine
	Route Route
}

const RATE_LIMIT_DURATION = 15 * 60 * 60

func MiddleWare(endpoints map[string]*Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		endpoint, ok := endpoints[path+method]

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Endpoint not found",
			})
			c.Set("error", true)
			return
		}

		if endpoint.RateLimit > 0 && (endpoint.LastTime+RATE_LIMIT_DURATION) < int(time.Now().Unix()) {
			endpoint.ResetRequest()
		}

		if endpoint.RateLimit > 0 && endpoint.Requests >= endpoint.RateLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many Request",
			})
			c.Set("error", true)
			return
		}

		if endpoint.EnableAuth {
			token := strings.Split(c.Request.Header.Get("Authorization"), " ")[1]
			err := auth.ValidJwtToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				c.Set("error", true)
				return
			}
		}

		c.Set("error", false)

		c.Next()

		endpoint.IncReqCounter()
	}
}

func (a *Apy) AddEndpoint(e Endpoint) {
	a.Route.Endpoints[e.Path+e.Method] = &e
	a.App.Handle(e.Method, e.Path, func(c *gin.Context) {
		if err := c.MustGet("error").(bool); err == true {
			return
		}

		res, err := a.Fetch(c)

		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"body": res,
			})
		}
	})
}

func (a *Apy) Fetch(c *gin.Context) (string, error) {
	method, path := c.Request.Method, c.Request.URL.Path
	url := a.GetUrl(path)

	endpoint := a.Route.Endpoints[path+method]

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

func (a *Apy) GetUrl(path string) string {
	return a.Route.Path + path
}

func (a *Apy) Run() {
	a.App.GET("/auth", func(c *gin.Context) {
		token, err := auth.CreateJwtToken()
		if err != nil {
			return
		}
		c.JSON(200, token)
	})
	a.App.Run()
}

func (e *Endpoint) SetCache(body string) {
	e.Cache = body
}

func (e *Endpoint) ResetRequest() {
	e.Requests = 0
	e.LastTime = int(time.Now().Unix())
}

func (e *Endpoint) IncReqCounter() {
	e.Requests++
}

func (a *Apy) EnableMetrics() {
	a.App.GET("/metrics", func(c *gin.Context) {
		metrics := make(map[string]Metric)
		for _, endpoint := range a.Route.Endpoints {
			metrics[endpoint.Name] = Metric{Path: endpoint.Path, Method: endpoint.Method, Requests: endpoint.Requests}
		}
		c.JSON(200, metrics)
	})
}

func (a *Apy) Init(path string) {
	route := Route{Name: "Api", Path: path, Endpoints: make(map[string]*Endpoint)}
	a.App = gin.Default()
	a.Route = route
	a.App.Use(MiddleWare(a.Route.Endpoints))
}
