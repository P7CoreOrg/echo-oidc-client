package cli

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	oidc "echo-oidc-client/pkg/p7coreorg/go-oidc"
	"echo-oidc-client/pkg/pkce"

	"fmt"
	"time"

	"html/template"

	"echo-oidc-client/pkg/globals"

	"github.com/skratchdot/open-golang/open"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	//	clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientID     = "native.code"
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	port         = "1323"
	ipaddress    = "127.0.0.1"
	redirectPath = "/auth/google/callback"
	rootUrl      = "http://" + ipaddress + ":" + port
	redirectUrl  = rootUrl + redirectPath
	//  authority    = "https://accounts.google.com"
	//	authority = "https://localhost:6001"
	authority = "https://demo.identityserver.io"
	options   = &oidc.ProviderOptions{
		Authority:            authority,
		AuthorityIssuerMatch: false,
	}
	pkceState *pkce.AuthorizeState
)

const itemsTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h4>You have logged into AFX!</h4>
		<p>You can close this window, or we will redirect you to the <a href="https://docs.microsoft.com/cli/azure/">Azure CLI documents</a> in 10 seconds.</p>
		{{range .Items}}<div>{{ . }}</div>{{else}}<div><strong>no rows</strong></div>{{end}}
	</body>
</html>`

type PageItems struct {
	Title string
	Items []string
}

const successTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="refresh" content="10;url={{.RedirectUrl}}">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h4>You have logged into AFX!</h4>
		<p>You can close this window, or we will redirect you to the <a href="{{.RedirectUrl}}">Azure CLI documents</a> in 10 seconds.</p>
	</body>
</html>`

type PageSuccess struct {
	Title       string
	RedirectUrl string
}

type LoginResult struct {
	IdToken     string
	AccessToken string
	Error       error
}

func LoginSignOut() (err error) {
	err = globals.Delete("_afx_access_token")
	return
}

func LoginState() (err error) {
	err, rawAccessToken := globals.Get("_afx_access_token")
	if err != nil {
		return
	}
	fmt.Println(fmt.Sprintf("access_token: %s", string(*rawAccessToken)))
	return
}

func Login() *LoginResult {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	loginResult := &LoginResult{}

	callbackHandled := make(chan bool)
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, options)
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID: clientID,
		//		ClientSecret: clientSecret,
		Endpoint:    provider.Endpoint(),
		RedirectURL: redirectUrl,
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}

	pkceState = pkce.CreateAuthorizePkceState()

	url := config.AuthCodeURL(pkceState.State,
		oauth2.SetAuthURLParam("nonce", pkceState.Nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceState.Pkce.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", pkce.Sha256),
	)

	fmt.Println(fmt.Sprintf("code_url:%s", url))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			callbackHandled <- true
		}()
		t, err := template.New("webpage").Parse(itemsTemplate)
		check(err)

		data := PageItems{
			Title: "Fail",
			Items: []string{
				"You may return to your app now.",
			},
		}

		err = t.Execute(w, data)
		check(err)
	})
	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			callbackHandled <- true
		}()
		t, err := template.New("webpage").Parse(successTemplate)
		check(err)
		data := PageSuccess{
			Title:       "Success",
			RedirectUrl: "https://docs.microsoft.com/cli/azure/",
		}

		err = t.Execute(w, data)
		check(err)
	})
	handleFail := func(w http.ResponseWriter, r *http.Request, message string) {

		http.Redirect(w, r, "/fail?message="+message, http.StatusFound)
	}

	http.HandleFunc(redirectPath, func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Get("state") != pkceState.State {
			handleFail(w, r, "state did not match")
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"),
			oauth2.SetAuthURLParam("code_verifier", pkceState.Pkce.CodeVerifier))
		if err != nil {
			handleFail(w, r, "Failed to exchange token: "+err.Error())
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			handleFail(w, r, "No id_token field in oauth2 token.")
			return
		}
		loginResult.IdToken = rawIDToken

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			handleFail(w, r, "Failed to verify ID Token: "+err.Error())
			return
		}

		accessToken, ok := oauth2Token.Extra("access_token").(string)
		if !ok {
			handleFail(w, r, "No access_token field in oauth2 token.")
			return
		}
		rawAccessToken := []byte(accessToken)

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			handleFail(w, r, err.Error())
			return
		}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			handleFail(w, r, err.Error())
			return
		}

		fmt.Println(string(data))

		// last thing we do is write the access_token to persistent storage
		err = globals.Put("_afx_access_token", &rawAccessToken)
		if err != nil {
			handleFail(w, r, "Cannot store access_token.")
			return
		}

		http.Redirect(w, r, "/success", http.StatusFound)

	})

	srv := startHttpServer(ipaddress + ":" + port)
	defer func() {
		if err := srv.Shutdown(context.TODO()); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}()

	errBrowser := open.Run(url)
	if errBrowser != nil {
		panic(errBrowser) // failure/timeout shutting down the server gracefully
	}

	<-callbackHandled
	fmt.Println("Login Done")
	return loginResult
}

func startHttpServer(addr string) *http.Server {

	srv := &http.Server{Addr: addr}

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop. don't use
			// code with race conditions like these for production. see post
			// comments below on more discussion on how to handle this.
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()
	serving := make(chan bool)

	go func() {
		success := false
		for n := 0; n <= 5; n++ {
			time.Sleep(time.Second)

			log.Println("Checking if started...")
			resp, err := http.Get(rootUrl)
			if err != nil {
				log.Println("Failed:", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Println("Not OK:", resp.StatusCode)
				continue
			}
			success = true
			log.Println("SERVER UP AND RUNNING!")

			// Reached this point: server is up and running!
			break
		}
		serving <- success
	}()

	<-serving
	// returning reference so caller can call Shutdown()
	return srv
}
