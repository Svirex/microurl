package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims - данные для JWT
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// JWTKey - тип для записи в контекст запроса uid пользователя
type JWTKey string

func (api *API) cookieAuth(next http.Handler) http.Handler {
	fn := func(response http.ResponseWriter, request *http.Request) {
		jwtKey, err := request.Cookie("jwt")
		if errors.Is(err, http.ErrNoCookie) {
			if strings.Contains(request.URL.Path, "api/user/urls") {
				api.logger.Error("not found auth cookei for api/user/urls")
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			api.generateCookieAndHandleNext(next, response, request)
			return
		}
		uid, err := getUserID(api.secretKey, jwtKey.Value)
		if err != nil { // если токен не проходит проверку на подлинность
			api.logger.Errorln("cookie auth middleware, jwt not valid", err)
			api.generateCookieAndHandleNext(next, response, request)
			return
		}
		if uid == "" {
			api.logger.Infoln("cookie auth middleware, uid is empty")
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.SetCookie(response, jwtKey)
		ctx := context.WithValue(request.Context(), JWTKey("uid"), uid)
		next.ServeHTTP(response, request.WithContext(ctx))

	}
	return http.HandlerFunc(fn)
}

func (api *API) generateCookieAndHandleNext(next http.Handler, response http.ResponseWriter, request *http.Request) {
	uid, err := generateUserIDAndJWTAndSetCookie(api.secretKey, response)
	if err != nil {
		api.logger.Errorln("cookie auth middleware, couldn't generate jwt, cookie exist: ", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := context.WithValue(request.Context(), JWTKey("uid"), uid)
	next.ServeHTTP(response, request.WithContext(ctx))
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
