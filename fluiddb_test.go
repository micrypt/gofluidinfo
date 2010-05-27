package main

import (
"fmt"
"./fluiddb"
//"http"
"io/ioutil"
"log"
)

func main() {

url := "/users/username"

myclient := fluiddb.NewClient("test","test")

r, err := myclient.Get(url)

var b []byte;
if err == nil {
    b, err = ioutil.ReadAll(r.Body);
    r.Body.Close();
}

if err != nil {
    log.Stderr(err)
} else {
    fmt.Println(string(b));
}



}

