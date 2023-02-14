---
title: API Reference

language_tabs: # must be one of https://github.com/rouge-ruby/rouge/wiki/List-of-supported-languages-and-lexers
  - shell

includes:
  - errors

search: true

code_clipboard: true

meta:
  - name: Docs
    content: Api-gateway documentaiton
---

# Introduction

This is a simple API Gateway written in Go, that enables you to handle incoming HTTP request, provide routing and authorization, monitor traffic, cache responses and enforce rate limit for each endpoint

See repo [Here](https://github.com/elfacu0/API-Gateway).

# Endpoints

## Add new endpoint


```shell
curl "https://api-0fx2.onrender.com/add" \
  -d "name=api&url=https://xkcd.com/3/info.0.json&method=GET"  \
  -X POST 
```

> The above command returns JSON structured like this:

```json
  {
    "Message": "Endpoint Created Successfully",
    "Path": "/ff863bc5792dbe25e098"
  }
```

This endpoint creates new custom Endpoints.

### HTTP Request

`POST https://api-0fx2.onrender.com/add`

### Request Body

Parameter | Default | Description
--------- | ------- | -----------
name | "" | Name of the endpoint.
url | "" | Url of the desired endpoint.
method | "" | Method of request. (GET-POST-DELETE).
rate-limit | "0"(disabled) | Maximum number of requests per 15 minutes.
enable-cache | ""(disabled) | If enabled the response will be cached
enable-auth | ""(disabled) | If enabled each request will need to have a bearer token


<aside class="success">
Remember — a happy kitten is an authenticated kitten!
</aside>

## Get Metrics

```shell
curl "https://api-0fx2.onrender.com/metrics"
```

> The above command returns JSON structured like this:

```json
{
   "Add Endpoint":{
      "url":"",
      "path":"/add",
      "method":"POST",
      "requests":0
   },
   "Auth":{
      "url":"",
      "path":"/auth",
      "method":"GET",
      "requests":0
   },
   "Metrics":{
      "url":"",
      "path":"/metrics",
      "method":"GET",
      "requests":3
   }
}
```

Retrieve the metrics of all current endpoints   

<aside class="warning">This won't show any setting like rate limit</aside>

### HTTP Request

`GET https://api-0fx2.onrender.com/metrics`


# Authentication

> To get Auth token, use this code:

```shell
curl "https://api-0fx2.onrender.com/auth"
```

Any endpoint with enable-auth set to true will need to use a bearer token, you can get it doing a GET Request to /auth

Set header to

`Bearer <token>`

<aside class="notice">
You must replace <code>token</code> with the returned value
</aside>

