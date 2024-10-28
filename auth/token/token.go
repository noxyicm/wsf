package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"github.com/noxyicm/wsf/cache"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/crypt"
	"github.com/noxyicm/wsf/errors"
)

// Public constants
const (
	StorageName = "tokenCache"
	TokenKey    = "token"
	HashAlgo    = "sha256"

	StateChecked = 0
	StateInvalid
	StateExpired
	StateInvalidIssuer
	StateInvalidAudience
	StateInvalidApplicant
	StateEmpty
	StateInvalidApplicantHost
	StateInvalidHash

	TokenLifeTime int64 = 864000
)

// Public variables
var (
	ErrorExpired              = errors.New("Token expired")
	ErrorInvalid              = errors.New("Invalid token")
	ErrorInvalidIssuer        = errors.New("Invalid issuer")
	ErrorInvalidAudience      = errors.New("Invalid audience")
	ErrorInvalidApplicant     = errors.New("Invalid applicant")
	ErrorEmpty                = errors.New("Unauthorized request. Request has no token")
	ErrorInvalidApplicantHost = errors.New("Invalid applicant host")
	ErrorInvalidHash          = errors.New("Invalid request token")
	ErrorInvalidStorage       = errors.New("Storage is undefined")
)

// Token represents a token
type Token struct {
	params      map[string]interface{}
	response    response.Interface
	storage     cache.Interface
	token       string
	tokenID     string
	tokenSecret string

	lastError error
}

// SetStorage sets a token storage
func (t *Token) SetStorage(strg cache.Interface) {
	t.storage = strg
}

// Storage returns token storage
func (t *Token) Storage() cache.Interface {
	return t.storage
}

// SetParam sets a token parameter
func (t *Token) SetParam(key string, value interface{}) *Token {
	t.params[key] = value
	return t
}

// SetParams sets a token parameters
func (t *Token) SetParams(params map[string]interface{}) *Token {
	for key, value := range params {
		t.SetParam(key, value)
	}

	return t
}

// Param returns token parameter
func (t *Token) Param(key string) interface{} {
	if v, ok := t.params[key]; ok {
		return v
	}

	return nil
}

// SetToken sets a token
func (t *Token) SetToken(token string) *Token {
	t.token = token
	return t
}

// Token returns token
func (t *Token) Token() string {
	return t.token
}

// SetTokenID sets token id
func (t *Token) SetTokenID(tokenID string) error {
	t.tokenID = tokenID
	return nil

	parts, err := t.parseTokenID(tokenID)
	if err != nil {
		return err
	}

	if len(parts) > 0 {
		if v, ok := parts["audience"]; ok && v == t.token {
			t.tokenID = tokenID
		}
	}

	return nil
}

// CreateTokenID creates a new token id
func (t *Token) CreateTokenID() (string, error) {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 64)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	t.tokenID = string(b)
	return t.tokenID, nil
}

// TokenID returns token id
func (t *Token) TokenID() string {
	return t.tokenID
}

// SetTokenSecret sets a token secret
func (t *Token) SetTokenSecret(tokenSecret string) *Token {
	t.tokenSecret = tokenSecret
	return t
}

// TokenSecret returns token secret
func (t *Token) TokenSecret() string {
	return t.tokenSecret
}

// LastError returns last acuired error
func (t *Token) LastError() error {
	return t.lastError
}

// ToString packs a token to string
func (t *Token) ToString() string {
	str, err := t.pack()
	if err != nil {
		t.lastError = err
	}

	return string(str)
}

// Valid returns error if token is invalid
func (t *Token) Valid() error {
	if len(t.params) == 0 {
		t.Invalidate()
		return errors.New(ErrorInvalid.Error())
	}

	if v, ok := t.params["expire"]; !ok || int(v.(float64)) < int(time.Now().Unix()) {
		t.Invalidate()
		return errors.New(ErrorExpired.Error())
	}

	if v, ok := t.params["issuer"]; !ok || v.(string) != config.App.GetString("application.Domain") {
		t.Invalidate()
		return errors.New(ErrorInvalidIssuer.Error())
	}

	return nil
}

// Load loads token from storage
func (t *Token) Load() error {
	if t.storage == nil {
		return errors.New("Storage is undefined")
	}

	packed, _ := t.storage.Load(t.tokenID, false)
	if t.storage.Error() != nil {
		return t.storage.Error()
	}

	return t.unpack(packed)
}

// Save saves token to storage
func (t *Token) Save() error {
	if t.storage == nil {
		return errors.New("No storage")
	}

	packed, err := t.pack()
	if err != nil {
		return err
	}

	if t.storage.Save(packed, t.tokenID, []string{t.tokenID}, TokenLifeTime) {
		return nil
	}

	return t.storage.Error()
}

// Invalidate renders token invalid
func (t *Token) Invalidate() error {
	if t.storage == nil {
		return errors.New("No storage")
	}

	if t.storage.Remove(t.tokenID) {
		return nil
	}

	return t.storage.Error()
}

// Lifetime returns a token life time
func (t *Token) Lifetime() int {
	return int(TokenLifeTime)
}

func (t *Token) parseParameters(rsp response.Interface) map[string]interface{} {
	return make(map[string]interface{})
}

func (t *Token) parseTokenID(tokenID string) (map[string]string, error) {
	cph, err := crypt.NewCipher(nil)
	if err != nil {
		return nil, err
	}

	decrypted, err := cph.DecodeString(tokenID)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(decrypted[:len(decrypted)-14], ".")
	if len(parts) == 2 {
		return map[string]string{"applicant": parts[0], "audience": parts[1]}, nil
	}

	return nil, errors.New("Invalid token id")
}

func (t *Token) pack() ([]byte, error) {
	pt := packedToken{
		Token:       t.token,
		TokenID:     t.tokenID,
		TokenSecret: t.tokenSecret,
		Params:      t.params,
	}
	serialized, err := json.Marshal(pt)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pack token")
	}

	return serialized, nil
}

func (t *Token) unpack(hash []byte) error {
	pt := packedToken{}
	if err := json.Unmarshal(hash, &pt); err != nil {
		return errors.Wrap(err, "Unable to unpack token")
	}

	t.token = pt.Token
	t.tokenID = pt.TokenID
	t.tokenSecret = pt.TokenSecret
	t.params = pt.Params
	return nil
}

// CreateRequestToken creates a new request token
func CreateRequestToken(appID int, appSecret string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(strconv.Itoa(appID)))
	hashhmac := mac.Sum(nil)
	return base64.URLEncoding.EncodeToString(append([]byte(strconv.Itoa(appID)+":"), hashhmac...))
}

// NewToken creates a new token from request token or loads existsing
func NewToken(appID int, appSecret string, storage cache.Interface) (*Token, error) {
	t := new(Token)
	t.params = make(map[string]interface{})

	t.SetToken(strconv.Itoa(appID)).SetTokenSecret(appSecret)
	t.SetParams(map[string]interface{}{
		"issuer":   config.App.GetString("application.Domain"),
		"audience": appID,
		"expire":   time.Now().Unix() + TokenLifeTime,
	})

	if _, err := t.CreateTokenID(); err != nil {
		return nil, err
	}

	if storage == nil {
		return nil, errors.New(ErrorInvalidStorage.Error())
	}

	t.storage = storage
	return t, nil
}

// NewEmptyToken creates an empty token
func NewEmptyToken() *Token {
	return &Token{
		params: make(map[string]interface{}),
	}
}

type packedToken struct {
	Token       string
	TokenID     string
	TokenSecret string
	Params      map[string]interface{}
}
