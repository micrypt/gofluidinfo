//Copyright (c) 2010 Seyi Ogunyemi

//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in
//all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//THE SOFTWARE.


package fluiddb

import (
  "http"
  "encoding/base64"
  "io"
  "os"
  "strings"
  "net"
  "bufio"
  "strconv"
  "fmt"
  "bytes"
  "container/vector"
)

type readClose struct {
  io.Reader
  io.Closer
}

type badStringError struct {
  what string
  str  string
}

func (e *badStringError) String() string { return fmt.Sprintf("%s %q", e.what, e.str) }

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func send(req *http.Request) (resp *http.Response, err os.Error) {
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
    n, err := strconv.Atoi64(v[0])
    if err != nil {
      return nil, &badStringError{"invalid Content-Length", v[0]}
    }
    r = io.LimitReader(r, n)
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

// Naive utility url encode method.  Converts a string map into a query string
//
// Returns a string in the format "?param1=value1&param2=value2"
func UrlEncode(urlmap map[string]string) (string) {
    url := "?"
    var temp vector.StringVector
    var key, value string

    for key, value = range urlmap {
         temp.Push(key + "=" + value)
    }
    url += strings.Join(temp, "&")
    return url
}

func authGet(url, user, pwd string) (r *http.Response, err os.Error) {
  var req http.Request
  req.Method = "GET"
  req.Header = map[string][]string{"Authorization": {"Basic " +
    encodedUsernameAndPassword(user, pwd)} }
  if req.URL, err = http.ParseURL(url); err != nil {
    return
  }
  if r, err = send(&req); err != nil {
    return
  }
  return
}

// Post issues a POST to the specified URL.
//
// Caller should close r.Body when done reading it.
func authPost(url, user, pwd, client, clientURL, version, agent, bodyType string,
              body io.Reader) (r *http.Response, err os.Error) {
  var req http.Request
  req.Method = "POST"
  req.Body = body.(io.ReadCloser)
  req.Header = map[string][]string{
    "Content-Type":         {bodyType},
    "Transfer-Encoding":    {"chunked"},
    "User-Agent":           {agent},
    "X-FluidDB-Client":     {client},
    "X-FluidDB-Client-URL": {clientURL},
    "X-FluidDB-Version":    {version},
    "Authorization": {"Basic " + encodedUsernameAndPassword(user, pwd)},
  }

  req.URL, err = http.ParseURL(url)
  if err != nil {
    return nil, err
  }

  return send(&req)
}


// Put issues a PUT to the specified URL.
//
// Caller should close r.Body when done reading it.
func authPut(url, user, pwd, client, clientURL, version, agent, bodyType string,
              body io.Reader) (r *http.Response, err os.Error) {
  var req http.Request
  req.Method = "PUT"
  req.Body = body.(io.ReadCloser)
  if user != "" && pwd != "" {
      req.Header = map[string][]string{
        "Content-Type":         {bodyType},
        "Transfer-Encoding":    {"chunked"},
        "User-Agent":           {agent},
        "X-FluidDB-Client":     {client},
        "X-FluidDB-Client-URL": {clientURL},
        "X-FluidDB-Version":    {version},
        "Authorization": {"Basic " + encodedUsernameAndPassword(user, pwd)},
      }
  } else {
      req.Header = map[string][]string{
        "Content-Type":         {bodyType},
        "Transfer-Encoding":    {"chunked"},
        "User-Agent":           {agent},
        "X-FluidDB-Client":     {client},
        "X-FluidDB-Client-URL": {clientURL},
        "X-FluidDB-Version":    {version},
      }
  }

  req.URL, err = http.ParseURL(url)
  if err != nil {
    return nil, err
  }

  return send(&req)
}

// Delete issues a DELETE to the specified URL.
func authDelete(url, user, pwd string) (r *http.Response, err os.Error) {
  var req http.Request
  req.Method = "DELETE"
  if user != "" && pwd != "" {
      req.Header = map[string][]string{"Authorization": {"Basic " +
        encodedUsernameAndPassword(user, pwd)} }
  }
  if req.URL, err = http.ParseURL(url); err != nil {
    return
  }
  if r, err = send(&req); err != nil {
    return
  }
  return
}

// Head issues a HEAD to the specified URL.
func authHead(url, user, pwd string) (r *http.Response, err os.Error) {
  var req http.Request
  req.Method = "HEAD"
  if user != "" && pwd != "" {
      req.Header = map[string][]string{"Authorization": {"Basic " +
        encodedUsernameAndPassword(user, pwd)} }
  }
  if req.URL, err = http.ParseURL(url); err != nil {
    return
  }
  if r, err = send(&req); err != nil {
    return
  }
  return
}

// Do an authenticated Get if we've called Authenticated, otherwise
// just Get it without authentication
func httpGet(url, user, pwd string) (*http.Response, string, os.Error) {
  var r *http.Response
  var full string = ""
  var err os.Error

  if user != "" && pwd != "" {
    r, err = authGet(url, user, pwd)
  } else {
    r, err = http.Get(url)
  }

  return r, full, err
}

// Do an authenticated Post if we've called Authenticated, otherwise
// just Post it without authentication
func httpPost(url, user, pwd, client, clientURL, version, agent,
              data string) (*http.Response, os.Error) {
  var r *http.Response
  var err os.Error

  body := bytes.NewBufferString(data)
  bodyType := "application/json"

  if user != "" && pwd != "" {
    r, err = authPost(url, user, pwd, client, clientURL,
      version, agent, bodyType, body)
  } else {
    r, err = http.Post(url, bodyType, body)
  }

  return r, err
}

// Do an authenticated Put
func httpPut(url, user, pwd, client, clientURL, version, agent,
              data string) (*http.Response, os.Error) {
  var r *http.Response
  var err os.Error

  body := bytes.NewBufferString(data)
  bodyType := "application/json"

  r, err = authPut(url, user, pwd, client, clientURL,
      version, agent, bodyType, body)

  return r, err
}

// Do an authenticated Delete 
func httpDelete(url, user, pwd string) (*http.Response, os.Error) {
  var r *http.Response
  var err os.Error

  r, err = authDelete(url, user, pwd)

  return r, err
}

// Do an authenticated Head 
func httpHead(url, user, pwd string) (*http.Response, os.Error) {
  var r *http.Response
  var err os.Error

  r, err = authHead(url, user, pwd)

  return r, err
}

const (
	DEFAULT_CLIENT         = "GoFluidDB"
	DEFAULT_CLIENT_URL     = "http://github.com/micrypt/GoFluidDB"
	DEFAULT_CLIENT_VERSION = "0.1"
	DEFAULT_USER_AGENT     = "gofluiddb"
	ERROR                  = "GoFluidDB Error: "
	WARNING                = "GoFluidDB Warning: "
	DEFAULT_PORT           = 80
	SECURE_PORT            = 443
	UNIX_CREDENTIALS_FILE  = ".fluidDBcredentials"
	// Here's hoping there's a stable Go port to the Windows platform sometime in the future.
	// WINDOWS_CREDENTIALS_FILE    = "fluidDBcredentials.ini"
	RETRY_TIMEOUT          = 5e9 // unlikely the user will choose this
	PRIMITIVE_CONTENT_TYPE = "application/vnd.fluiddb.value+json"
	HEADER_ERROR           = "X-FluidDB-Error-Class"
	HEADER_REQUEST_ID      = "X-FluidDB-Request-Id"
    FLUIDDB_PATH    = "http://fluiddb.fluidinfo.com"
    SANDBOX_PATH    = "http://sandbox.fluidinfo.com"
)

type Client struct {
    Username    string
    Password    string
    URL         string
    Client      string
    ClientURL   string
    Version     string
    Agent       string
}

func NewClient(username, password string) *Client {
	return &Client{username, password, SANDBOX_PATH, DEFAULT_CLIENT, DEFAULT_CLIENT_URL, DEFAULT_CLIENT_VERSION, DEFAULT_USER_AGENT}
}

func (self *Client) SetActiveMode() {
    self.URL = FLUIDDB_PATH
}

func (self *Client) Get( url string) (*http.Response, os.Error) {

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    resp, _ , err = httpGet(url, self.Username, self.Password)
    
    return resp, err

}

func (self *Client) Post(url, data string) (*http.Response, os.Error) {

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    resp, err = httpPost(url, self.Username, self.Password, self.Client, self.ClientURL, self.Version, self.Agent, data)
    
    return resp, err

}

func (self *Client) Put(url, data string) (*http.Response, os.Error) {

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    resp, err = httpPut(url, self.Username, self.Password, self.Client, self.ClientURL, self.Version, self.Agent, data)
    
    return resp, err

}

func (self *Client) Delete( url string) (*http.Response, os.Error) {

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    resp, err = httpDelete(url, self.Username, self.Password)
    
    return resp, err

}

func (self *Client) Head( url string) (*http.Response, os.Error) {

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    resp, err = httpHead(url, self.Username, self.Password)
    
    return resp, err

}
