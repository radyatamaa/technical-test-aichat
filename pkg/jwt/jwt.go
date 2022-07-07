package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwts "github.com/golang-jwt/jwt/v4"
)

type (
	Payload map[string]interface{}

	JWT interface {
		// Ctx Which shallowly clones current object and sets the context for next operation.
		Ctx(ctx context.Context) JWT

		// SetAdapter Set a cache adapter for authentication.
		SetAdapter(adapter Adapter) JWT

		// MiddlewareRPCAuth Implemented basic JWT permission authentication.
		MiddlewareRPCAuth(ctx context.Context, token string) (context.Context, error)

		// Middleware Implemented basic JWT permission authentication.
		Middleware(r *http.Request) (*http.Request, error)

		// GenerateToken Generates and returns a new token object with payload.
		GenerateToken(payload Payload, issuer string, expiredTime int) (*Token, error)

		// RetreadToken Retreads and returns a new token object depend on old token.
		// By default, the token expired error doesn't ignore.
		// You can ignore expired error by setting the `ignoreExpired` parameter.
		RetreadToken(token string, expiredTime int, ignoreExpired ...bool) (*Token, error)

		// RefreshToken Generates and returns a new token object from.
		RefreshToken(r *http.Request, expiredTime int) (*Token, error)

		// DestroyToken Destroy the cache of a token.
		DestroyToken(r *http.Request) error

		// DestroyIdentity Destroy the identification mark.
		DestroyIdentity(issuer, identity interface{}) error

		// GetToken Get token from request.
		// By default, the token expired error doesn't ignored.
		// You can ignore expired error by setting the `ignoreExpired` parameter.
		GetToken(r *http.Request, expiredTime int, ignoreExpired ...bool) (*Token, error)

		// GetPayload Retrieve payload from request.
		// By default, the token expired error doesn't ignore.
		// You can ignore expired error by setting the `ignoreExpired` parameter.
		GetPayload(r *http.Request, ignoreExpired ...bool) (payload Payload, err error)

		// GetIdentity Retrieve identity from request.
		// By default, the token expired error doesn't ignore.
		// You can ignore expired error by setting the `ignoreExpired` parameter.
		GetIdentity(r *http.Request, ignoreExpired ...bool) (interface{}, error)
	}
)

type Options struct {

	// Define the token seek locations within requests.
	// Support header, form, cookie and query parameter.
	// Support to seek multiple locations, Separate multiple seek locations with commas.
	Locations string

	// Define the signing method for generate token.
	// Support multiple signing method such as HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384 and ES512
	SignMethod string

	// Define the secret cacheKey of HMAC.
	// Only support secret value.
	// The secret cacheKey is required, when the signing method is one of HS256, HS384 or HS512.
	SecretKey string

	// Define the public cacheKey of RSA or ECDSA.
	// Support file path or value.
	// The public cacheKey is required, when the signing method is one of RS256, RS384, RS512, ES256, ES384 and ES512.
	PublicKey string

	// Define the private cacheKey of RSA or ECDSA.
	// Support file path or value.
	// The private cacheKey is required, when the signing method is one of RS256, RS384, RS512, ES256, ES384 and ES512.
	PrivateKey string

	// Define the identity cacheKey of the claims.
	// After opening the identification identifier and cache interface, the system will
	// construct a unique authorization identifier for each token. If the same user is
	// authorized to log in elsewhere, the previous token will no longer be valid.
	IdentityKey string
}

type jwt struct {
	signMethod      string
	tokenCtxKey     string
	tokenSeeks      [][2]string
	rsaPublicKey    *rsa.PublicKey
	rsaPrivateKey   *rsa.PrivateKey
	ecdsaPublicKey  *ecdsa.PublicKey
	ecdsaPrivateKey *ecdsa.PrivateKey
	secretKey       []byte
	ctx             context.Context
	identityKey     string
	adapter         Adapter
}

type Token struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
	RefreshAt time.Time `json:"refresh_at"`
}

const (
	jwtAudience    = "aud"
	jwtId          = "jti"
	jwtIssueAt     = "iat"
	jwtExpired     = "exp"
	jwtIssuer      = "iss"
	jwtNotBefore   = "nbf"
	jwtSubject     = "sub"
	noDetailReason = "no detail reason"
)

const (
	HS256 = "HS256"
	HS512 = "HS512"
	HS384 = "HS384"
	RS256 = "RS256"
	RS384 = "RS384"
	RS512 = "RS512"
	ES256 = "ES256"
	ES384 = "ES384"
	ES512 = "ES512"

	tokenSeekFromHeader  = "header"
	tokenSeekFromQuery   = "query"
	tokenSeekFromCookie  = "cookie"
	tokenSeekFromForm    = "form"
	tokenSeekFieldHeader = "Authorization"
	authorizationBearer  = "Bearer"

	defaultSignMethod     = HS256
	defaultExpirationTime = time.Hour
	defaultPayloadCtxKey  = "JWT_PAYLOAD"
	defaultTokenCtxKey    = "JWT_TOKEN"
	defaultIdentityKey    = "jwt:%s:identity:%s"
)

func NewJwt(opt *Options) (JWT, error) {
	j := new(jwt)

	if err := j.init(opt); err != nil {
		return nil, err
	}

	return j, nil
}

func (j *jwt) init(opt *Options) (err error) {
	j.setLocations(opt.Locations)
	j.setIdentityKey(opt.IdentityKey)

	if err = j.setSigningMethod(opt.SignMethod); err != nil {
		return
	}

	if j.isHMAC() {
		if err = j.setSecretKey(opt.SecretKey); err != nil {
			return
		}
	} else {
		if err = j.setPublicKey(opt.PublicKey); err != nil {
			return
		}

		if err = j.setPrivateKey(opt.PrivateKey); err != nil {
			return
		}
	}

	return
}

// Ctx Which shallowly clones current object and sets the context for next operation.
func (j *jwt) Ctx(ctx context.Context) JWT {
	newJwt := j.clone()
	newJwt.ctx = ctx
	return newJwt
}

// SetAdapter Set a cache adapter for authentication.
func (j *jwt) SetAdapter(adapter Adapter) JWT {
	j.adapter = adapter
	return j
}

// MiddlewareRPCAuth Implemented basic JWT permission authentication.
func (j *jwt) MiddlewareRPCAuth(ctx context.Context, token string) (context.Context, error) {
	payload, err := j.parseTokenRPC(token)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, defaultPayloadCtxKey, payload)
	ctx = context.WithValue(ctx, defaultTokenCtxKey, token)

	return ctx, nil
}

// Middleware Implemented basic JWT permission authentication.
func (j *jwt) Middleware(r *http.Request) (*http.Request, error) {
	payload, token, err := j.parseRequest(r)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, defaultPayloadCtxKey, payload)
	ctx = context.WithValue(ctx, defaultTokenCtxKey, token)

	return r.WithContext(ctx), nil
}

// GenerateToken Generates and returns a new token object with payload.
func (j *jwt) GenerateToken(payload Payload, issuer string, expiredTime int) (*Token, error) {
	if j.identityKey != "" {
		if _, ok := payload[j.identityKey]; !ok {
			return nil, errMissingIdentity
		}
	}

	var (
		claims    = make(jwts.MapClaims)
		now       = time.Now()
		expiredAt = now.Add(j.setExpiredTime(expiredTime))
		id        = strconv.FormatInt(now.UnixNano(), 10)
	)

	claims[jwtId] = id
	claims[jwtIssuer] = issuer
	claims[jwtIssueAt] = now.Unix()
	claims[jwtExpired] = expiredAt.Unix()
	for k, v := range payload {
		switch k {
		case jwtAudience, jwtExpired, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
			// ignore the standard claims
		default:
			claims[k] = v
		}
	}

	token, err := j.signToken(claims)
	if err != nil {
		return nil, err
	}

	if j.identityKey != "" {
		if err = j.saveIdentity(issuer, payload[j.identityKey], id, j.setExpiredTime(expiredTime)); err != nil {
			return nil, err
		}
	}

	return &Token{
		Token:     token,
		ExpiredAt: expiredAt,
	}, nil
}

// RefreshToken Generates and returns a new token object from.
func (j *jwt) RefreshToken(r *http.Request, expiredTime int) (*Token, error) {
	return j.RetreadToken(j.seekToken(r), expiredTime,true)
}

// RetreadToken Retreads and returns a new token object depend on old token.
// By default, the token expired error doesn't ignore.
// You can ignore expired error by setting the `ignoreExpired` parameter.
func (j *jwt) RetreadToken(token string, expiredTime int, ignoreExpired ...bool) (*Token, error) {
	if token == "" {
		return nil, errMissingToken
	}

	var (
		err       error
		claims    jwts.MapClaims
		newClaims jwts.MapClaims
		now       = time.Now()
	)

	claims, err = j.parseToken(token, ignoreExpired...)
	if err != nil {
		return nil, err
	}


	newClaims = make(jwts.MapClaims)
	for k, v := range claims {
		newClaims[k] = v
	}

	expiredAt := now.Add(j.setExpiredTime(expiredTime))

	newClaims[jwtIssueAt] = now.Unix()
	newClaims[jwtExpired] = expiredAt.Unix()

	token, err = j.signToken(newClaims)
	if err != nil {
		return nil, err
	}

	object := &Token{Token: token, ExpiredAt: expiredAt}

	if j.identityKey == "" {
		return object, nil
	}

	if _, ok := claims[j.identityKey]; !ok {
		return nil, errMissingIdentity
	}

	if err = j.verifyIdentity(claims[jwtIssuer], claims[j.identityKey], claims[jwtId], false); err != nil {
		return nil, err
	}

	if err = j.saveIdentity(claims[jwtIssuer], claims[j.identityKey], claims[jwtId], j.setExpiredTime(expiredTime)); err != nil {
		return nil, err
	}

	return object, nil
}

// DestroyToken Destroy the cache of a token.
func (j *jwt) DestroyToken(r *http.Request) error {
	if j.identityKey == "" {
		return nil
	}

	claims, err := j.parseToken(j.seekToken(r), true)
	if err != nil {
		return err
	}

	if _, ok := claims[j.identityKey]; !ok {
		return errMissingIdentity
	}

	if err = j.verifyIdentity(claims[jwtIssuer].(string), claims[j.identityKey], claims[jwtId], true); err != nil {
		return err
	}

	return j.removeIdentity(claims[jwtIssuer].(string), claims[j.identityKey])
}

// DestroyIdentity Destroy the identification mark.
func (j *jwt) DestroyIdentity(issuer, identity interface{}) error {
	return j.removeIdentity(issuer, identity)
}

// GetToken Retrieve token from request.
// By default, the token expired error doesn't ignored.
// You can ignore expired error by setting the `ignoreExpired` parameter.
func (j *jwt) GetToken(r *http.Request, expiredTime int, ignoreExpired ...bool) (*Token, error) {
	var token string

	if v := r.Context().Value(defaultTokenCtxKey); v != nil {
		token = v.(string)
	} else if token = j.seekToken(r); token == "" {
		return nil, errMissingToken
	}

	claims, err := j.parseToken(token, ignoreExpired...)
	if err != nil {
		return nil, err
	}

	expiredAt := time.Unix(int64(claims[jwtExpired].(float64)), 0)

	return &Token{
		Token:     token,
		ExpiredAt: expiredAt,
	}, nil
}

// GetPayload Retrieve payload from request.
// By default, the token expired error doesn't ignore.
// You can ignore expired error by setting the `ignoreExpired` parameter.
func (j *jwt) GetPayload(r *http.Request, ignoreExpired ...bool) (payload Payload, err error) {
	if v := r.Context().Value(defaultPayloadCtxKey); v != nil {
		payload = v.(Payload)
	} else {
		payload, _, err = j.parseRequest(r, ignoreExpired...)
	}

	return
}

// GetIdentity Retrieve identity from request.
// By default, the token expired error doesn't ignore.
// You can ignore expired error by setting the `ignoreExpired` parameter.
func (j *jwt) GetIdentity(r *http.Request, ignoreExpired ...bool) (interface{}, error) {
	if j.identityKey == "" {
		return nil, errMissingIdentity
	}

	payload, err := j.GetPayload(r, ignoreExpired...)
	if err != nil {
		return nil, err
	}

	identity, ok := payload[j.identityKey]
	if !ok {
		return nil, errMissingIdentity
	}

	return identity, nil
}

// Parses and returns the payload and token from requests.
func (j *jwt) parseTokenRPC(token string, ignoreExpired ...bool) (payload Payload, err error) {
	claims, err := j.parseToken(token, ignoreExpired...)
	if err != nil {
		return
	}

	if j.identityKey != "" {
		if err = j.verifyIdentity(claims[jwtIssuer].(string), claims[j.identityKey], claims[jwtId], false); err != nil {
			return
		}
	}

	payload = make(Payload)
	for k, v := range claims {
		switch k {
		case jwtAudience, jwtExpired, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
			// ignore the standard claims
		default:
			payload[k] = v
		}
	}

	return
}

// Parses and returns the payload and token from requests.
func (j *jwt) parseRequest(r *http.Request, ignoreExpired ...bool) (payload Payload, token string, err error) {
	if token = j.seekToken(r); token == "" {
		err = errMissingToken
		return
	}

	claims, err := j.parseToken(token, ignoreExpired...)
	if err != nil {
		return
	}

	if j.identityKey != "" {
		if err = j.verifyIdentity(claims[jwtIssuer].(string), claims[j.identityKey], claims[jwtId], false); err != nil {
			return
		}
	}

	payload = make(Payload)
	for k, v := range claims {
		switch k {
		case jwtAudience, jwtExpired, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
			// ignore the standard claims
		default:
			payload[k] = v
		}
	}

	return
}

// Seeks and returns token from request.
// 1.from header    Authorization: Bearer ${token}
// 2.from query     ${url}?${cacheKey}=${token}
// 3.from cookie    ${cacheKey}=${token}
// 4.from form      ${cacheKey}=${token}
func (j *jwt) seekToken(r *http.Request) (token string) {
	for _, item := range j.tokenSeeks {
		if len(token) > 0 {
			break
		}
		switch item[0] {
		case tokenSeekFromHeader:
			token = j.seekTokenFromHeader(r, item[1])
		case tokenSeekFromQuery:
			token = j.seekTokenFromQuery(r, item[1])
		case tokenSeekFromCookie:
			token = j.seekTokenFromCookie(r, item[1])
		case tokenSeekFromForm:
			token = j.seekTokenFromForm(r, item[1])
		}
	}

	return
}

// Seeks and returns JWT token from the headers of request.
func (j *jwt) seekTokenFromHeader(r *http.Request, key string) string {
	parts := strings.SplitN(r.Header.Get(key), " ", 2)
	if len(parts) != 2 || parts[0] != authorizationBearer {
		return ""
	}

	return parts[1]
}

// Seeks and returns JWT token from the query params of request.
func (j *jwt) seekTokenFromQuery(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// Seeks and returns JWT token from the cookies of request.
func (j *jwt) seekTokenFromCookie(r *http.Request, key string) string {
	cookie, _ := r.Cookie(key)
	return cookie.String()
}

// Seeks and returns JWT token from the post forms of request.
func (j *jwt) seekTokenFromForm(r *http.Request, key string) string {
	return r.Form.Get(key)
}

// Parses and returns a claims map from the token.
// By default, The token expiration errors will not be ignored.
// The claims are nil when the token expiration errors not be ignored.
func (j *jwt) parseToken(token string, ignoreExpired ...bool) (jwts.MapClaims, error) {
	jt, err := jwts.Parse(token, func(t *jwts.Token) (key interface{}, err error) {
		if jwts.GetSigningMethod(j.signMethod) != t.Method {
			err = errSigningMethodNotMatch
			return
		}

		switch {
		case j.isHMAC():
			key = j.secretKey
		case j.isRSA():
			key = j.rsaPublicKey
		case j.isECDSA():
			key = j.ecdsaPublicKey
		}

		return
	})
	if err != nil {
		switch e := err.(type) {
		case *jwts.ValidationError:
			switch e.Errors {
			case jwts.ValidationErrorExpired:
				if len(ignoreExpired) > 0 && ignoreExpired[0] {
					// ignore token expired error
				} else {
					return nil, errExpiredToken
				}
			default:
				return nil, errInvalidToken
			}
		default:
			return nil, errInvalidToken
		}
	}

	if jt == nil || !jt.Valid {
		return nil, errInvalidToken
	}

	claims := jt.Claims.(jwts.MapClaims)

	if _, ok := claims[jwtId]; !ok {
		return nil, errInvalidToken
	}

	if _, ok := claims[jwtIssueAt]; !ok {
		return nil, errInvalidToken
	}

	if _, ok := claims[jwtExpired]; !ok {
		return nil, errInvalidToken
	}

	return claims, nil
}

// Signings and returns a token depend on the claims.
func (j *jwt) signToken(claims jwts.MapClaims) (token string, err error) {
	jt := jwts.New(jwts.GetSigningMethod(j.signMethod))
	jt.Claims = claims

	switch {
	case j.isHMAC():
		token, err = jt.SignedString(j.secretKey)
	case j.isRSA():
		token, err = jt.SignedString(j.rsaPrivateKey)
	case j.isECDSA():
		token, err = jt.SignedString(j.ecdsaPrivateKey)
	}
	if err != nil {
		return
	}

	return
}

// Check whether the signing method is HMAC.
func (j *jwt) isHMAC() bool {
	switch j.signMethod {
	case HS256, HS384, HS512:
		return true
	}
	return false
}

// Check whether the signing method is RSA.
func (j *jwt) isRSA() bool {
	switch j.signMethod {
	case RS256, RS384, RS512:
		return true
	}
	return false
}

// Check whether the signing method is ECDSA.
func (j *jwt) isECDSA() bool {
	switch j.signMethod {
	case RS256, RS384, RS512:
		return true
	}
	return false
}

// Set signing method.
// Support multiple signing method such as HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384 and ES512
func (j *jwt) setSigningMethod(signingMethod string) error {
	switch signingMethod {
	case HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512:
		j.signMethod = signingMethod
	case "":
		j.signMethod = defaultSignMethod
	default:
		return errInvalidSigningMethod
	}

	return nil
}

// SetTokenLookup Set the token search location.
func (j *jwt) setLocations(tokenLookup string) {
	j.tokenSeeks = make([][2]string, 0)

	for _, method := range strings.Split(tokenLookup, ",") {
		parts := strings.Split(strings.TrimSpace(method), ":")
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case tokenSeekFromHeader, tokenSeekFromQuery, tokenSeekFromCookie, tokenSeekFromForm:
			j.tokenSeeks = append(j.tokenSeeks, [2]string{k, v})
		}
	}

	if len(j.tokenSeeks) == 0 {
		j.tokenSeeks = append(j.tokenSeeks, [2]string{tokenSeekFromHeader, tokenSeekFieldHeader})
	}
}

// Set the identity of the token.
// After opening the identification identifier and cache interface, the system will
// construct a unique authorization identifier for each token. If the same user is
// authorized to log in elsewhere, the previous token will no longer be valid.
func (j *jwt) setIdentityKey(identityKey string) {
	j.identityKey = identityKey
}

// Set expiration time.
// If only set the expiration time,
// The refresh time will automatically be set to half of the expiration time.
func (j *jwt) setExpiredTime(expirationTime int) time.Duration {
	if expirationTime > 0 {
		return time.Duration(expirationTime) * time.Second
	} else {
		return defaultExpirationTime
	}
}

// Set secret cacheKey.
func (j *jwt) setSecretKey(secretKey string) (err error) {
	if secretKey == "" {
		return errInvalidSecretKey
	}

	j.secretKey = StringToBytes(secretKey)

	return
}

// Set public cacheKey.
// Allow setting of public cacheKey file or public cacheKey.
func (j *jwt) setPublicKey(publicKey string) (err error) {
	if publicKey == "" {
		return errInvalidPublicKey
	}

	var (
		fileInfo os.FileInfo
		key      []byte
	)

	if fileInfo, err = os.Stat(publicKey); err != nil {
		key = StringToBytes(publicKey)
	} else {
		if fileInfo.Size() == 0 {
			return errInvalidPublicKey
		}

		if key, err = ioutil.ReadFile(publicKey); err != nil {
			return
		}
	}

	if j.isRSA() {
		if j.rsaPublicKey, err = jwts.ParseRSAPublicKeyFromPEM(key); err != nil {
			return
		}
	}

	if j.isECDSA() {
		if j.ecdsaPublicKey, err = jwts.ParseECPublicKeyFromPEM(key); err != nil {
			return
		}
	}

	return
}

// Set private cacheKey.
// Allow setting of private cacheKey file or private cacheKey.
func (j *jwt) setPrivateKey(privateKey string) (err error) {
	if privateKey == "" {
		return errInvalidPrivateKey
	}

	var (
		fileInfo os.FileInfo
		key      []byte
	)

	if fileInfo, err = os.Stat(privateKey); err != nil {
		key = StringToBytes(privateKey)
	} else {
		if fileInfo.Size() == 0 {
			return errInvalidPrivateKey
		}

		if key, err = ioutil.ReadFile(privateKey); err != nil {
			return
		}
	}

	if j.isRSA() {
		if j.rsaPrivateKey, err = jwts.ParseRSAPrivateKeyFromPEM(key); err != nil {
			return
		}
	}

	if j.isECDSA() {
		if j.ecdsaPrivateKey, err = jwts.ParseECPrivateKeyFromPEM(key); err != nil {
			return
		}
	}

	return
}

// returns context object
func (j *jwt) getCtx() context.Context {
	if j.ctx == nil {
		return context.Background()
	}
	return j.ctx
}

// returns a shallow copy of current object.
func (j *jwt) clone() *jwt {
	return &jwt{
		signMethod:      j.signMethod,
		tokenCtxKey:     j.tokenCtxKey,
		tokenSeeks:      j.tokenSeeks,
		rsaPublicKey:    j.rsaPublicKey,
		rsaPrivateKey:   j.rsaPrivateKey,
		ecdsaPublicKey:  j.ecdsaPublicKey,
		ecdsaPrivateKey: j.ecdsaPrivateKey,
		secretKey:       j.secretKey,
		adapter:         j.adapter,
		identityKey:     j.identityKey,
		ctx:             j.ctx,
	}
}

// build a cache key by identity.
func (j *jwt) cacheKey(issuer, identity interface{}) string {
	return fmt.Sprintf(defaultIdentityKey, String(issuer), String(identity))
}

// save identification mark.
func (j *jwt) saveIdentity(issuer, identity, jid interface{}, expiredTime time.Duration) error {
	if j.adapter == nil {
		return nil
	}

	return j.adapter.Put(j.getCtx(), j.cacheKey(issuer, identity), String(jid), expiredTime)
}

// verify identification mark.
func (j *jwt) verifyIdentity(issuer, identity, jid interface{}, ignoreMissed bool) error {
	if j.adapter == nil {
		return nil
	}

	v, err := j.adapter.Get(j.getCtx(), j.cacheKey(issuer, identity))
	if err != nil {
		return err
	}

	oldJid := String(v)

	if oldJid == "" {
		if ignoreMissed {
			return nil
		} else {
			return errInvalidToken
		}
	}

	if String(jid) != oldJid {
		return errAuthElsewhere
	}

	return nil
}

// remove identification mark.
func (j *jwt) removeIdentity(issuer, identity interface{}) error {
	if j.adapter == nil {
		return nil
	}
	return j.adapter.Delete(j.getCtx(), j.cacheKey(issuer, identity))
}
