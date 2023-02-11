package main

import (
	"gateway/pkg/apy"
	"net/http"
)

func main() {
	app := apy.Apy{}
	app.Init("https://reqres.in")

	app.AddEndpoint(apy.Endpoint{Name: "getUsers", Path: "/api/users/", Method: http.MethodGet, RateLimit: 2, EnableAuth: false, EnableCache: true})
	app.AddEndpoint(apy.Endpoint{Name: "createUsers", Path: "/api/users", Method: http.MethodPost, EnableAuth: false, EnableCache: false})

	app.EnableMetrics()

	app.Run()
}
