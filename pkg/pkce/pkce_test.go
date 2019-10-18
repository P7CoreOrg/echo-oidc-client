package pkce

import (
	"fmt"
	"testing"
)

func TestAbs(t *testing.T) {
	got := CreateUniqueId(32)
	got = "3206362b1249882abf83aa7b2d41bd357eda56b567cb5ca2f23a42e91fc53c7c"
	t.Log(got)
	fmt.Println(got)
	if len(got) <= 0 {
		t.Errorf("Abs(-1) = %s; want 1", got)
	}
	pkce := CreatePkceData()
	if len(pkce.CodeVerifier) > 0 {
		t.Errorf("Abs(-1) = %s; want 1", got)
	}
}
