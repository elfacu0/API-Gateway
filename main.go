package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	Name      string              `json:"name"`
	Path      string              `json:"path"`
	Endpoints map[string]Endpoint `json:"endpoints"`
}

type Apy struct {
	App   *gin.Engine
	Route Route
}

func CreateJwtToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": time.After(15 * time.Minute),
	})
	tokenString, err := token.SignedString([]byte("netdata.io"))
	return tokenString, err
}

func ValidJwtToken(tokenString string) error {
	if tokenString == "" {
		return errors.New("There is no token")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("AllYourBase"), nil
	})

	if token.Valid {
		return nil
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		return errors.New("Expired Token")
	}
	return errors.New("Couldn't handle this token")
}

// func MiddleWare() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		c.Set("example", "12345")

// 		c.Next()
// 	}
// }

func (a *Apy) AddEndpoint(e Endpoint) {
	a.Route.Endpoints[e.Name] = e
}

func Fetch(url string, method string, c *gin.Context) (string, error) {
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

func (a *Apy) GetUrl(e *Endpoint) string {
	return a.Route.Path + e.Path
}

func (a *Apy) Run() {
	// a.App.Use(MiddleWare())
	for i, endpoint := range a.Route.Endpoints {

		a.App.Handle(endpoint.Method, endpoint.Path, func(c *gin.Context) {
			if endpoint.EnableAuth {
				token := c.Request.Header.Get("Authorization")
				err := ValidJwtToken(token)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": err.Error(),
					})
					return
				}
			}

			if endpoint.EnableCache {

			}

			if endpoint.RateLimit > 0 && a.Route.Endpoints[i].Requests >= endpoint.RateLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many Request",
				})
				return
			}

			res, err := Fetch(a.GetUrl(&endpoint), endpoint.Method, c)

			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"message": res,
				})

				// a.Route.Endpoints[i].IncReqCounter()
			}

		})
	}
	a.App.Run()
}

func (e *Endpoint) IncReqCounter() {
	e.Requests++
}

func main() {
	route := Route{Name: "Api", Path: "https://reqres.in"}
	apy := Apy{App: gin.Default(), Route: route}

	apy.AddEndpoint(Endpoint{Name: "getUsers", Path: "/api/users", Method: http.MethodGet, EnableAuth: false})
	apy.AddEndpoint(Endpoint{Name: "createUsers", Path: "/api/users", Method: http.MethodPost, RateLimit: 2, EnableAuth: true})

	CreateJwtToken()

	apy.Run()

}
