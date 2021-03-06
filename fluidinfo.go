package fluidinfo

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type readClose struct {
	io.Reader
	io.Closer
}

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func send(req *http.Request) (resp *http.Response, err error) {
	addr := req.URL.Host
	if !hasPort(addr) {
		addr += ":http"
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	reader := bufio.NewReader(conn)
	resp, err = http.ReadResponse(reader, req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	r := io.Reader(reader)
	if v := resp.Header["Content-Length"]; v != nil {
		n, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, &badStringError{"invalid Content-Length", v[0]}
		}
		v := int64(n)
		r = io.LimitReader(r, v)
	}
	resp.Body = readClose{r, conn}
	return
}

func encodedUsernameAndPassword(user, pwd string) string {
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write([]byte(user + ":" + pwd))
	encoder.Close()
	return bb.String()
}

func authGet(url_, user, pwd string) (r *http.Response, err error) {
	var req http.Request
	req.Method = "GET"
	req.Header = map[string][]string{"Authorization": {"Basic " +
		encodedUsernameAndPassword(user, pwd)}}
	if req.URL, err = url.Parse(url_); err != nil {
		return
	}
	if r, err = send(&req); err != nil {
		return
	}
	return
}

// Post issues a POST to the specified URL.
// Caller should close r.Body when done reading it.
func authPost(url_, user, pwd, client, clientURL, version, agent, bodyType string,
	body io.Reader) (r *http.Response, err error) {
	var req http.Request
	req.Method = "POST"
	req.Body = body.(io.ReadCloser)
	req.Header = map[string][]string{
		"Content-Type":           {bodyType},
		"Transfer-Encoding":      {"chunked"},
		"User-Agent":             {agent},
		"X-Fluidinfo-Client":     {client},
		"X-Fluidinfo-Client-URL": {clientURL},
		"X-Fluidinfo-Version":    {version},
		"Authorization":          {"Basic " + encodedUsernameAndPassword(user, pwd)},
	}

	req.URL, err = url.Parse(url_)
	if err != nil {
		return nil, err
	}

	return send(&req)
}

// Put issues a PUT to the specified URL.
// Caller should close r.Body when done reading it.
func authPut(url_, user, pwd, client, clientURL, version, agent, bodyType string,
	body io.Reader) (r *http.Response, err error) {
	var req http.Request
	req.Method = "PUT"
	req.Body = body.(io.ReadCloser)
	if user != "" && pwd != "" {
		req.Header = map[string][]string{
			"Content-Type":           {bodyType},
			"Transfer-Encoding":      {"chunked"},
			"User-Agent":             {agent},
			"X-Fluidinfo-Client":     {client},
			"X-Fluidinfo-Client-URL": {clientURL},
			"X-Fluidinfo-Version":    {version},
			"Authorization":          {"Basic " + encodedUsernameAndPassword(user, pwd)},
		}
	} else {
		req.Header = map[string][]string{
			"Content-Type":           {bodyType},
			"Transfer-Encoding":      {"chunked"},
			"User-Agent":             {agent},
			"X-Fluidinfo-Client":     {client},
			"X-Fluidinfo-Client-URL": {clientURL},
			"X-Fluidinfo-Version":    {version},
		}
	}

	req.URL, err = url.Parse(url_)
	if err != nil {
		return nil, err
	}

	return send(&req)
}

// Delete issues a DELETE to the specified URL.
func authDelete(url_, user, pwd string) (r *http.Response, err error) {
	var req http.Request
	req.Method = "DELETE"
	if user != "" && pwd != "" {
		req.Header = map[string][]string{"Authorization": {"Basic " +
			encodedUsernameAndPassword(user, pwd)}}
	}
	if req.URL, err = url.Parse(url_); err != nil {
		return
	}
	if r, err = send(&req); err != nil {
		return
	}
	return
}

// Head issues a HEAD to the specified URL.
func authHead(url_, user, pwd string) (r *http.Response, err error) {
	var req http.Request
	req.Method = "HEAD"
	if user != "" && pwd != "" {
		req.Header = map[string][]string{"Authorization": {"Basic " +
			encodedUsernameAndPassword(user, pwd)}}
	}
	if req.URL, err = url.Parse(url_); err != nil {
		return
	}
	if r, err = send(&req); err != nil {
		return
	}
	return
}

// Do an authenticated Get if we've called Authenticated, otherwise
// just Get it without authentication.
func httpGet(url_, user, pwd string) (*http.Response, string, error) {
	var r *http.Response
	var full string = ""
	var err error
	if user != "" && pwd != "" {
		r, err = authGet(url_, user, pwd)
	} else {
		r, err = http.Get(url_)
	}
	return r, full, err
}

// Do an authenticated Post if we've called Authenticated, otherwise
// just Post it without authentication.
func httpPost(url_, user, pwd, client, clientURL, version, agent,
	data string) (*http.Response, error) {
	var r *http.Response
	var err error
	body := bytes.NewBufferString(data)
	bodyType := "application/json"
	if user != "" && pwd != "" {
		r, err = authPost(url_, user, pwd, client, clientURL,
			version, agent, bodyType, body)
	} else {
		r, err = http.Post(url_, bodyType, body)
	}
	return r, err
}

// Do an authenticated Put.
func httpPut(url_, user, pwd, client, clientURL, version, agent,
	data string) (*http.Response, error) {
	var r *http.Response
	var err error
	body := bytes.NewBufferString(data)
	bodyType := "application/json"
	r, err = authPut(url_, user, pwd, client, clientURL,
		version, agent, bodyType, body)
	return r, err
}

// Do an authenticated Delete.
func httpDelete(url_, user, pwd string) (*http.Response, error) {
	var r *http.Response
	var err error
	r, err = authDelete(url_, user, pwd)
	return r, err
}

// Do an authenticated Head.
func httpHead(url_, user, pwd string) (*http.Response, error) {
	var r *http.Response
	var err error
	r, err = authHead(url_, user, pwd)
	return r, err
}

const (
	DEFAULT_CLIENT           = "gofluidinfo"
	DEFAULT_CLIENT_URL       = "http://github.com/micrypt/gofluidinfo"
	DEFAULT_CLIENT_VERSION   = "0.1"
	DEFAULT_USER_AGENT       = "gofluidinfo"
	ERROR                    = "gofluidinfo Error: "
	WARNING                  = "gofluidinfo Warning: "
	DEFAULT_PORT             = 80
	SECURE_PORT              = 443
	UNIX_CREDENTIALS_FILE    = ".fluidinfocredentials"
	WINDOWS_CREDENTIALS_FILE = "fluidinfocredentials.ini"
	RETRY_TIMEOUT            = 5e9
	PRIMITIVE_CONTENT_TYPE   = "application/vnd.fluidinfo.value+json"
	HEADER_ERROR             = "X-Fluidinfo-Error-Class"
	HEADER_REQUEST_ID        = "X-Fluidinfo-Request-Id"
	FLUIDINFO_PATH           = "http://fluiddb.fluidinfo.com"
	SANDBOX_PATH             = "http://sandbox.fluidinfo.com"
)

type Client struct {
	Username  string
	Password  string
	URL       string
	Client    string
	ClientURL string
	Version   string
	Agent     string
}

func NewClient(username, password string) *Client {
	return &Client{username, password, SANDBOX_PATH, DEFAULT_CLIENT, DEFAULT_CLIENT_URL, DEFAULT_CLIENT_VERSION, DEFAULT_USER_AGENT}
}

func (self *Client) SetActiveMode() {
	self.URL = FLUIDINFO_PATH
}

func (self *Client) Get(url_ string) (*http.Response, error) {
	url_ = self.URL + url_
	var resp *http.Response
	var err error
	resp, _, err = httpGet(url_, self.Username, self.Password)
	return resp, err
}

func (self *Client) Post(url_, data string) (*http.Response, error) {
	url_ = self.URL + url_
	var resp *http.Response
	var err error
	resp, err = httpPost(url_, self.Username, self.Password, self.Client, self.ClientURL, self.Version, self.Agent, data)
	return resp, err
}

func (self *Client) Put(url_, data string) (*http.Response, error) {
	url_ = self.URL + url_
	var resp *http.Response
	var err error
	resp, err = httpPut(url_, self.Username, self.Password, self.Client, self.ClientURL, self.Version, self.Agent, data)
	return resp, err
}

func (self *Client) Delete(url_ string) (*http.Response, error) {
	url_ = self.URL + url_
	var resp *http.Response
	var err error
	resp, err = httpDelete(url_, self.Username, self.Password)
	return resp, err
}

func (self *Client) Head(url_ string) (*http.Response, error) {
	url_ = self.URL + url_
	var resp *http.Response
	var err error
	resp, err = httpHead(url_, self.Username, self.Password)
	return resp, err
}
