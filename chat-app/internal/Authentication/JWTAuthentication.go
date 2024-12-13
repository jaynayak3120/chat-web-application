package jwtauth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("Jay-Golang-Developer")

func CreateToken(user_id string) (map[string]string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user_id,
		"exp":     time.Now().Add(time.Minute * 10).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["user_id"] = user_id
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	//fmt.Printf("Type of exp is: %T", time.Now().Add(time.Hour*24).Unix())
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	response := map[string]string{
		"access_token":  tokenString,
		"refresh_token": refreshTokenString,
		//"rt_exp_time":   rtClaims["exp"].(int64),
	}

	return response, nil
}

func RefreshAccessToken(refreshTokenString string) (map[string]string, error) {

	refreshToken, err := jwt.Parse(refreshTokenString, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !refreshToken.Valid {
		return nil, fmt.Errorf("refresh token is invalid")
	}

	rtClaims, ok := refreshToken.Claims.(jwt.MapClaims)
	if !ok || rtClaims["user_id"].(string) == "" {
		return nil, fmt.Errorf("invalid claims in refresh token")
	}

	tokenPair, err := CreateToken(rtClaims["user_id"].(string))
	if err != nil {
		return nil, err
	}

	tokenPair["user_id"] = rtClaims["user_id"].(string)

	return tokenPair, nil
}

func VerifyToken(r *http.Request) error {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" || len(tokenString) <= len("Bearer ") {
		return fmt.Errorf("missing authorization token")
	}

	tokenString = tokenString[len("Bearer "):]
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func VerifyToken_old(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}
