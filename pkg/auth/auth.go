package auth

import (
	logs "brainwars/pkg/logger"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// New instantiates the *Authenticator.
func New() (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
// This function takes an OAuth2 token, extracts the id_token, and verifies that it's valid and intended for your app.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token") // id_token contains identity info (e.g., who the user is). thats the whole purpose of oicd (to verify the user identity)
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func HandleLogin(ctx context.Context, c *gin.Context) (state string, err error) {
	l := logs.GetLoggerctx(ctx)
	state, err = generateRandomState()
	if err != nil {
		l.Sugar().Errorf("generate random state failed", err)
		return "", err
	}
	session := sessions.Default(c)
	session.Set("state", state)
	err = session.Save()
	if err != nil {
		l.Sugar().Errorf("session save failed", err)
		return "", err
	}
	return state, nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil

}

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
