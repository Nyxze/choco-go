package choco

import "encoding/base64"

const (
	ContentTypeAppJSON   = "application/json"
	ContentTypeAppXML    = "application/xml"
	ContentTypeTextPlain = "text/plain"
)

type AuthSchema string

const (
	AuthSchemeBasic  AuthSchema = "Basic"
	AuthSchemeBearer AuthSchema = "Bearer"
	AuthSchemeDigest AuthSchema = "Digest"
	AuthSchemeOAuth  AuthSchema = "OAuth"
	AuthSchemeJWT    AuthSchema = "JWT"
)

const (
	HeaderAuthorization   = "Authorization"
	HeaderAccept          = "Accept"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderLocation        = "Location"
	HeaderRetryAfter      = "Retry-After"
	HeaderRetryAfterMS    = "Retry-After-Ms"
	HeaderUserAgent       = "User-Agent"
	HeaderWWWAuthenticate = "WWW-Authenticate"
)

// SetHeader sets a header key to the provided value.
func (r *Request) SetHeader(key, value string) {
	r.req.Header.Set(key, value)
}

// AddHeader appends a value to an existing header key.
func (r *Request) AddHeader(key, value string) {
	r.req.Header.Add(key, value)
}

// DelHeader removes the specified header key.
func (r *Request) DelHeader(key string) {
	r.req.Header.Del(key)
}

// SetAuthorization sets the Authorization header using a defined AuthSchema and credentials/token.
//
// Example:
//
//	r.SetAuthorization(AuthSchemeBearer, "abc123")  -> Authorization: Bearer abc123
func (r *Request) SetAuthorization(scheme AuthSchema, token string) {
	r.SetHeader("Authorization", string(scheme)+" "+token)
}

// SetBasicAuth sets the Authorization header using HTTP Basic Auth with username and password.
func (r *Request) SetBasicAuth(username, password string) {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	r.SetAuthorization(AuthSchemeBasic, credentials)
}

// SetContentType sets the Content-Type header.
func (r *Request) SetContentType(value string) {
	r.SetHeader(HeaderContentType, value)
}

// SetAccept sets the Accept header.
func (r *Request) SetAccept(value string) {
	r.SetHeader(HeaderAccept, value)
}
