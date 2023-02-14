package main

import (
	"gateway/pkg/apy"
)

func main() {
	app := apy.Apy{}
	app.Init()

	app.EnableNewEndpoints()
	app.EnableMetrics()
	app.LoadEnpoints()

	app.Run()
}
