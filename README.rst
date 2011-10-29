============
gofluidinfo
============

gofluidinfo is an light wrapper for the Fluidinfo API in Go. 

(Based on the Go Language docs + library source and go-twitter)

Installation
============
Make files coming soon. 
For now include in your project directory and import "./fluidinfo"

Quick Start
===========

- Create a url to run the request against.

url := "/users/username"

- Create a client by passing in username and password string (empty strings for unauthenticated calls)/

myclient := fluidinfo.NewClient("test","test")

- Call the desired HTTP method. Returns http.Response and os.Error 

r, err := myclient.Get(url)

or

r, err := myclient.Post(url, data)

Matching client methods for the request methods supported by fluidinfo (GET, POST, PUT, DELETE & HEAD) are included.

Also included is a UrlEncode method which converts a string map "map[string]string" into a query string in the format "?param1=value1&param2=value2..."

Documentation
=============

doc/ - godoc generated files, coming soon

