package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type JWTKey string

func CookieAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(response http.ResponseWriter, request *http.Request) {
			jwtKey, err := request.Cookie("jwt")
			var uid string
			if errors.Is(err, http.ErrNoCookie) {
				uid, err = generateUserIDAndJWTAndSetCookie(secretKey, response)
				if err != nil {
					response.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				uid, err = getUserID(secretKey, jwtKey.Value)
				if err != nil { // если токен не проходит проверку на подлинность
					uid, err = generateUserIDAndJWTAndSetCookie(secretKey, response)
					if err != nil {
						response.WriteHeader(http.StatusBadRequest)
						return
					}
				}
				if uid == "" {
					response.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
			ctx := context.WithValue(request.Context(), JWTKey("uid"), uid)
			next.ServeHTTP(response, request.WithContext(ctx))

		}
		return http.HandlerFunc(fn)
	}
}

func generateUserID() (string, error) {
	v, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}
	return v.String(), nil
}

func buildJWTString(secretKey string, uid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: uid,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("build jwt token: %w", err)
	}
	return tokenString, nil
}

func getUserID(secretKey string, token string) (string, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("parse user id from jwt: %w", err)
	}
	return claims.UserID, nil
}

func generateUserIDAndJWTAndSetCookie(secretKey string, response http.ResponseWriter) (string, error) {
	uid, err := generateUserID()
	if err != nil {
		return "", fmt.Errorf("generateUserID: %w", err)
	}
	token, err := buildJWTString(secretKey, uid)
	if err != nil {
		return "", fmt.Errorf("buildJWTString: %w", err)
	}
	jwtCookie := &http.Cookie{
		Name:  "jwt",
		Value: token,
	}
	http.SetCookie(response, jwtCookie)
	return uid, nil
}
