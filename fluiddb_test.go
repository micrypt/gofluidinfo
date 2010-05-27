//Example call using the fluiddb package

package main

import (
"fmt"
"io/ioutil"
"log"
"./fluiddb"
)

func main() {

// Call to url pattern "/users/username"
url := "/users/test"

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

