package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token struct {
	Token   string `json:"access_token"`
	Type    string `json:"token_type"`
	Expires int64  `json:"expires_in"`
}

func CreateJwtToken() (Token, error) {
	expiration := time.Now().Unix() + (int64(time.Minute.Seconds()) * 15)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": expiration,
	})
	tokenString, err := token.SignedString([]byte("KwasHere"))
	res := Token{Token: tokenString, Type: "Bearer", Expires: expiration}
	return res, err
}

func ValidJwtToken(tokenString string) error {
	if tokenString == "" {
		return errors.New("Access Token is necessary")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("KwasHere"), nil
	})

	fmt.Println(err)

	if token.Valid {
		return nil
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		return errors.New("Expired Token")
	}
	return errors.New("Couldn't handle this token")
}
