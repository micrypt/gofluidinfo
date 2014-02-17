//Example call using the fluidinfo package

package fluidinfo

import (
	"io/ioutil"
	"testing"
)

func TestUser(t *testing.T) {
	// URL pattern is "/users/username"
	url := "/users/test"
	myclient := NewClient("test", "test")
	r, err := myclient.Get(url)
	defer r.Body.Close()
	if err != nil {
		t.Errorf("TestUser failed %v", err)
	}
	var b []byte
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("TestUser failed %v", err)
	}
	t.Logf("Test user passed %s", b)
}
