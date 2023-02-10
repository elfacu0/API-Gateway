package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Endpoint struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	EnableCache bool   `json:"cache"`
	EnableAuth  bool   `json:"auth"`
	RateLimit   int    `json:"rate-limit"`
	Requests    int    `json:"requests"`
}

type Route struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Apy struct {
	App   *gin.Engine
	Route Route
}

// func MiddleWare() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		c.Set("example", "12345")

// 		c.Next()
// 	}
// }

func (a *Apy) AddEndpoint(e Endpoint) *Apy {
	a.Route.Endpoints = append(a.Route.Endpoints, e)
	return a
}

func fetch(url string, method string, c *gin.Context) (string, error) {
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
		fmt.Println(err)
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (a *Apy) getUrl(e *Endpoint) string {
	return a.Route.Path + e.Path
}

func (a *Apy) Run() {
	// a.App.Use(MiddleWare())
	for i, endpoint := range a.Route.Endpoints {
		a.App.Handle(endpoint.Method, endpoint.Path, func(c *gin.Context) {

			if endpoint.EnableAuth {

			}

			if endpoint.EnableCache {

			}

			if endpoint.RateLimit > 0 && a.Route.Endpoints[i].Requests >= endpoint.RateLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many Request",
				})
				return
			}

			res, err := fetch(a.getUrl(&endpoint), endpoint.Method, c)

			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"message": res,
				})

				a.Route.Endpoints[i].incReqCounter()
			}

		})
	}
	a.App.Run()
}

func (e *Endpoint) incReqCounter() {
	e.Requests++
}

func main() {
	route := Route{Name: "Api", Path: "https://reqres.in"}
	apy := Apy{App: gin.Default(), Route: route}

	apy.AddEndpoint(Endpoint{Name: "getUsers", Path: "/api/users", Method: http.MethodGet, EnableCache: true})
	apy.AddEndpoint(Endpoint{Name: "createUsers", Path: "/api/users", Method: http.MethodPost, RateLimit: 2})

	apy.Run()

	// api1 := apy.AddRoute("www.apy.com")
	// api1.AddEndpoint("/users",{cache = false, })
	//
}
