package main

import (
	oidc "echo-oidc-client/pkg/p7coreorg/go-oidc"
	"echo-oidc-client/pkg/pkce"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
)

func init() {
	viper.SetConfigFile("config/appsettings.json")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	options := &oidc.ProviderOptions{
		Authority: "https://accounts.google.com",
		//		Authority:            "https://localhost:6001",
		AuthorityIssuerMatch: false,
	}
	provider, err := oidc.NewProvider(ctx, options)
	//provider, err := oidc.NewProvider(ctx, "https://localhost:6001")
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://127.0.0.1:1323/auth/google/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	PkceState := &pkce.AuthorizeState{}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/login", func(c echo.Context) error {
		PkceState = pkce.CreateAuthorizePkceState()

		url := config.AuthCodeURL(PkceState.State,
			oauth2.SetAuthURLParam("nonce", PkceState.Nonce),
			oauth2.SetAuthURLParam("code_challenge", PkceState.Pkce.CodeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", pkce.Sha256),
		)

		fmt.Println(fmt.Sprintf("code_url:%s", url))
		return c.Redirect(http.StatusFound, url)
	})
	e.GET("/auth/google/callback", func(c echo.Context) error {
		r := c.Request()

		if r.URL.Query().Get("state") != PkceState.State {
			return c.String(http.StatusBadRequest, "state did not match!")
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"),
			oauth2.SetAuthURLParam("code_verifier", PkceState.Pkce.CodeVerifier))
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			return c.String(http.StatusInternalServerError, "No id_token field in oauth2 token.")
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to verify ID Token: "+err.Error())
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, string(data))
	})

	e.Logger.Fatal(e.Start(":1323"))
}
