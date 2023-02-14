# ApyApp

## Description
This is a simple API Gateway written in Go, that enables you to handle incoming HTTP request, provide routing and authorization, monitor traffic, cache responses and enforce rate limit for each endpoint

## Live preview
You can test this project by making requests to https://api-0fx2.onrender.com/

## Docs
https://elfacu0.github.io/API-Gateway/#introduction

## Run
Build and run the docker-compose command to start Redis.
```
sudo docker compose up
```
Execute the main file
```
go run main.go
```

Or build it and then run the executable.

```
go build main.go
./main
```