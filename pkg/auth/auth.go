package auth

import (
	"context"
	"fmt"
	logs "brainwars/pkg/logger"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	EmailID     string
	ExpiryDate  time.Time
	InitiatedAt time.Time
}

func CreateJWTToken(input string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"emailID": input,
			"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), //7 days max expiry date
			"iat":     time.Now().Unix(),
		})
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWTToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	l := logs.GetLoggerctx(ctx)

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		l.Sugar().Errorf("parse jwt token failed", err)
		return nil, err
	}

	if !token.Valid {
		err := fmt.Errorf("invalid token")
		l.Sugar().Errorf("invalid jwt token", err)
		return nil, err
	}

	return token, nil
}

func ExtractClaims(token *jwt.Token) (*JWTClaims, error) {

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		expDate := claims["exp"].(float64)
		expseconds := int64(expDate)
		expnanoseconds := int64((expDate - float64(expseconds)) * 1e9)

		iatDate := claims["iat"].(float64)
		iatseconds := int64(iatDate)
		iatnanoseconds := int64((iatDate - float64(iatseconds)) * 1e9)

		// Create time.Time object

		return &JWTClaims{
			EmailID:     claims["emailID"].(string),
			ExpiryDate:  time.Unix(expseconds, expnanoseconds),
			InitiatedAt: time.Unix(iatseconds, iatnanoseconds),
		}, nil

	}
	return nil, fmt.Errorf("err extracting jwt token claims")
}
