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
  conn, err := net.Dial("tcp", "", addr)
  if err != nil {
    return nil, err
  }

  err = req.Write(conn)
  if err != nil {
    conn.Close()
    return nil, err
  }

  reader := bufio.NewReader(conn)
  resp, err = http.ReadResponse(reader, req.Method)
  if err != nil {
    conn.Close()
    return nil, err
  }

  r := io.Reader(reader)
  if v := resp.GetHeader("Content-Length"); v != "" {
    n, err := strconv.Atoi64(v)
    if err != nil {
      return nil, &badStringError{"invalid Content-Length", v}
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

func authGet(url, user, pwd string) (r *http.Response, err os.Error) {
  var req http.Request

  req.Header = map[string]string{"Authorization": "Basic " +
    encodedUsernameAndPassword(user, pwd)}
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
  req.Header = map[string]string{
    "Content-Type":         bodyType,
    "Transfer-Encoding":    "chunked",
    "User-Agent":           agent,
    "X-FluidDB-Client":     client,
    "X-FluidDB-Client-URL": clientURL,
    "X-FluidDB-Version":    version,
    "Authorization": "Basic " + encodedUsernameAndPassword(user, pwd),
  }

  req.URL, err = http.ParseURL(url)
  if err != nil {
    return nil, err
  }

  return send(&req)
}

// Do an authenticated Get if we've called Authenticated, otherwise
// just Get it without authentication
func httpGet(url, user, pass string) (*http.Response, string, os.Error) {
  var r *http.Response
  var full string = ""
  var err os.Error

  if user != "" && pass != "" {
    r, err = authGet(url, user, pass)
  } else {
    r, full, err = http.Get(url)
  }

  return r, full, err
}

// Do an authenticated Post if we've called Authenticated, otherwise
// just Post it without authentication
func httpPost(url, user, pass, client, clientURL, version, agent,
              data string) (*http.Response, os.Error) {
  var r *http.Response
  var err os.Error

  body := bytes.NewBufferString(data)
  bodyType := "application/x-www-form-urlencoded"

  if user != "" && pass != "" {
    r, err = authPost(url, user, pass, client, clientURL,
      version, agent, bodyType, body)
  } else {
    r, err = http.Post(url, bodyType, body)
  }

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
	// Here's hoping there's a Go port to the Windows platform sometime in the future.
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

func (self *Client) Call(method, url, data string) (*http.Response, os.Error) {

    method = strings.ToUpper(method)

    url = self.URL + url

    var resp *http.Response
    var err os.Error

    switch method {
        case "GET":
                    resp, _ , err = httpGet(url, self.Username, self.Password)
        case "POST":
                    resp, err = httpPost(url, self.Username, self.Password, self.Client, self.ClientURL, self.Version, self.Agent, data)
        default:        
                return nil, &badStringError{ ERROR, "Incorrect method"}

    }
    
    return resp, err

}
