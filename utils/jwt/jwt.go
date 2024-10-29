package jwtUtils

import (
	"errors"
	"fmt"
	"squad-maker/utils/env"
	"strconv"
	"time"

	pbAuth "squad-maker/generated/auth"

	"github.com/ezaurum/jwt-go/v4"
)

type (
	AuthClaims struct {
		jwt.StandardClaims
		SessionId int64           `json:"sid,omitempty"`
		UserType  pbAuth.UserType `json:"uty,omitempty"`
	}
)

var (
	jwtSecret    []byte
	jwtExpiresIn int32
)

func init() {
	var err error
	jwtSecret, err = env.GetSecretKey("JWT_SECRET")
	if err != nil {
		fmt.Printf("jwt init error (secret): %v\n", err)
	}
	jwtExpiresIn, err = env.GetInt32("JWT_EXPIRES_IN")

	if err != nil {
		fmt.Printf("jwt init error (expires_in): %v\n", err)
	}
}

func GetAuthTokenAndClaimsIfValid(bearer string) (*jwt.Token, *AuthClaims) {
	token := GetTokenIfValid(bearer, &AuthClaims{})
	if token != nil {
		claims, ok := token.Claims.(*AuthClaims)
		if ok {
			return token, claims
		}
	}
	return nil, nil
}

func GetExpiresAt() (*jwt.Time, int32) {
	return jwt.NewTime(float64(time.Now().Unix()) + float64(jwtExpiresIn)), jwtExpiresIn
}

func GenerateToken(claims jwt.Claims) (token string, err error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err = t.SignedString(jwtSecret)
	return token, err
}

func GenerateAuthToken(userId int64, userType pbAuth.UserType, sessionId int64) (token string, expiresIn int32, err error) {
	claims := &AuthClaims{}
	claims.ExpiresAt, expiresIn = GetExpiresAt()
	claims.Subject = strconv.FormatInt(userId, 10)
	claims.SessionId = sessionId
	claims.UserType = userType

	token, err = GenerateToken(*claims)
	return token, expiresIn, err
}

func GetTokenIfValid(tokenStr string, claimsType jwt.Claims) *jwt.Token {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claimsType,
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				fmt.Println("unexpected token signing method")
				return nil, errors.New("unexpected token signing method")
			}

			return jwtSecret, nil
		},
	)
	if err == nil && token.Valid {
		return token
	}
	fmt.Printf("received invalid token %v\n", err)
	return nil
}
