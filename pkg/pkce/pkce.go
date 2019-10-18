package pkce

import (
	"crypto"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/hex"
	"hash"
	"strings"
)

type CodeChallengeMethod string

const (
	Plain  CodeChallengeMethod = "plain"
	Sha256                     = "S256"
)

type Pkce struct {
	CodeVerifier  string
	CodeChallenge string
}
type AuthorizeState struct {
	Pkce  Pkce
	Nonce string
	State string
}

func CreatePkceData() Pkce {
	codeVerifier := CreateUniqueId(32)
	var digest hash.Hash
	digest = crypto.SHA256.New()
	digest.Write([]byte(codeVerifier))
	h := digest.Sum(nil)
	codeChallenge := Base64UrlEncode(h)

	pkce := &Pkce{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
	}
	return *pkce
}

func createNonce() string {
	return CreateUniqueId(16)
}
func createState() string {
	return CreateUniqueId(16)
}
func Base64UrlEncode(arg []byte) string {
	sEnc := b64.StdEncoding.EncodeToString(arg)
	parts := strings.Split(sEnc, "=")
	sEnc = parts[0]
	r := strings.NewReplacer("+", "-", "/", "_")
	sEnc = r.Replace(sEnc)
	return sEnc
}

func CreateRandomKey(length int) []byte {
	token := make([]byte, length)
	rand.Read(token)
	return token
}
func CreateRandomKeyString(length int) string {
	randKey := CreateRandomKey(length)
	sEnc := b64.StdEncoding.EncodeToString(randKey)
	return sEnc
}

func CreateUniqueId(length int) string {
	bytes := CreateRandomKey(length)
	sEnc := hex.EncodeToString(bytes)
	return sEnc
}

func CreateAuthorizePkceState() *AuthorizeState {
	pkce := CreatePkceData()
	state := &AuthorizeState{
		Pkce:  pkce,
		Nonce: createNonce(),
		State: createState(),
	}
	return state
}
