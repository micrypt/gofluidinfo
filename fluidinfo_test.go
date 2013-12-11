//Example call using the fluidinfo package

package main

import (
	"fmt"
	"github.com/micrypt/gofluidinfo/fluidinfo"
	"io/ioutil"
	"log"
)

func main() {

	// Call to url pattern "/users/username"
	url := "/users/test"

	myclient := fluidinfo.NewClient("test", "test")

	r, err := myclient.Get(url)

	var b []byte
	if err == nil {
		b, err = ioutil.ReadAll(r.Body)
		r.Body.Close()
	}

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(b))
	}

}
