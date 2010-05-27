============
gofluiddb
============

GoFluidDB is an light wrapper for the FluidDB API in Go. 

(Based on the Go Lang docs + library source and go-twitter)

Installation
============
Make files coming soon. 
For now include in your project directory and import "./fluiddb"

Quick Start
===========

1) Create a url to run the request against
2) Create a client and run the Call method 

url := "/users/username"

myclient := fluiddb.NewClient("test","test")

r, err myclient.Get(url)

or

r, err myclient.Post(url, data)

Matching client methods for the request methods supported by FluidDB (GET, POST, PUT, DELETE & HEAD) are included.

(Artifact method: I was considering adding HTTP Method specific calls as Go discourages default method parameters, so it's hard to do something like default=None on otherwise empty fields, that's done now)
r, err := myclient.Call("get",url, "") (A quick hack that's unnecessary at this point but I've left in anyway, just in case someone feels a compulsion to write bad code. :P)

Also included is a UrlEncode method which converts a string map "map[string]string" into a query string in the format "?param1=value1&param2=value2..."

Documentation
=============

doc/ - godoc generated files, coming soon

