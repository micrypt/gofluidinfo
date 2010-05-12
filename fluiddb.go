package fluidb

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"http"
	"io"
	"json"
	"os"
	"net"
//	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_CLIENT         = "gofluiddb"
	DEFAULT_CLIENT_URL     = "http://github.com/micrypt/gofluiddb"
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
)

//var FLUIDDB_PATH, _ = http.ParseURL("http://fluiddb.fluidinfo.com")
var FLUIDDB_PATH, _ = http.ParseURL("http://sandbox.fluidinfo.com")
//var SANDBOX_PATH, _ = http.ParseURL("http://sandbox.fluidinfo.com")


type Object struct {
	Id   string
	Body   []string
	Tags map[string]byte
}

type Response struct {
	Header string
    Object Object
}


type Api struct {
	clientConn *http.ClientConn
	url        *http.URL
	FluidDBStream     chan Response
	authData   string
	postData   string
    method     string
	stale      bool
}

func (self *Api) Close() {
	self.stale = true
	tcpConn, _ := self.clientConn.Close()
	if tcpConn != nil {
		tcpConn.Close()
	}
}

func (self *Api) connect(HttpMethod string) (*http.Response, os.Error) {
	tcpConn, err := net.Dial("tcp", "", self.url.Host+":80")
	if err != nil {
		return nil, err
	}
	self.clientConn = http.NewClientConn(tcpConn, nil)

	var req http.Request
	req.URL = self.url
	req.Method = HttpMethod
	req.Header = map[string]string{}
	req.Header["Authorization"] = "Basic " + self.authData

	if self.postData != "" {
		req.Method = "POST"
		req.Body = nopCloser{bytes.NewBufferString(self.postData)}
		req.ContentLength = int64(len(self.postData))
		req.Header["Content-Type"] = "application/x-www-form-urlencoded"
	}

	err = self.clientConn.Write(&req)
	if err != nil {
		return nil, err
	}

	resp, err := self.clientConn.Read()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (self *Api) readFluidDBStream(resp *http.Response) {
	var reader *bufio.Reader
	reader = bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			//we've been closed
			if self.stale {
				return
			}

			//otherwise, reconnect
			resp, err := self.connect(self.method)
			if err != nil {
				println(err.String())
				time.Sleep(RETRY_TIMEOUT)
				continue
			}

			if resp.StatusCode != 200 {
				continue
			}

			reader = bufio.NewReader(resp.Body)
			continue
		}
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		var Response Response
		json.Unmarshal(line, &Response)

		self.FluidDBStream <- Response
	}
}


type Client struct {
	Username string
	Password string
	FluidDBStream   chan Response
	conn     *Api
	connLock *sync.Mutex
}

func NewClient(username, password string) *Client {
	return &Client{username, password, make(chan Response), nil, new(sync.Mutex)}
}

func encodedAuth(user, pwd string) string {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(user + ":" + pwd))
	encoder.Close()
	return buf.String()
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() os.Error { return nil }

func (c *Client) connect(url *http.URL, HttpMethod, body string) (err os.Error) {
	if c.Username == "" || c.Password == "" {
		return os.NewError("The username or password is invalid")
	}

	c.connLock.Lock()
	var resp *http.Response
	//initialize the new FluidDBStream
	var sc Api

	sc.authData = encodedAuth(c.Username, c.Password)
    sc.method = HttpMethod

    if len(body) != 0 {
	    sc.postData = body
    }

	sc.url = url
	resp, err = sc.connect(sc.method)
	if err != nil {
		goto Return
	}

	if resp.StatusCode != 200 {
		err = os.NewError("FluidDB HTTP Error" + resp.Status)
		goto Return
	}

	//close the current connection
	if c.conn != nil {
		c.conn.Close()
	}

	c.conn = &sc
	sc.FluidDBStream = c.FluidDBStream
	go sc.readFluidDBStream(resp)

Return:
	c.connLock.Unlock()
	return
}

// GET function
func (c *Client) HttpGet(url string) os.Error {

    new_url, _ := http.ParseURL( FLUIDDB_PATH.String() + url);
	return c.connect(new_url, "GET", "")
}

// POST function
func (c *Client) HttpPost(url, body string) os.Error {

//	var body bytes.Buffer
//	body.WriteString("*****")

    new_url, _ := http.ParseURL(FLUIDDB_PATH.String() + url);
	return c.connect(new_url, "POST", body)
}

// PUT function
func (c *Client) HttpPut(url, body string) os.Error {

    new_url, _ := http.ParseURL(FLUIDDB_PATH.String() + url);
	return c.connect(new_url, "PUT", body)
}

// Close the client
func (c *Client) Close() {
	//has it already been closed?
	if c.conn.stale {
		return
	}
	c.conn.Close()
}
