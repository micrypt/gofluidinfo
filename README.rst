============
gofluiddb
============

GoFluidDB is an light wrapper for the FluidDB library in Go. 

(Based on the Go Lang docs + library source and go-twitter)

Installation
============

Quick Start
===========

1) Create a url to run the request against
2) Create a client and run the Call method 

(I am considering adding HTTP Method specific calls as Go discourages default method parameters so it's hard to do something like default=None on otherwise empty fields)

url := "/users/esteve"

myclient := fluiddb.NewClient("test","test")

r, err := myclient.Call("get",url, "")

Documentation
=============

doc/ - godoc generated files, coming soon

