package sync

import (
	"fmt"
	"strings"
	"testing"
)

func TestSome(t *testing.T) {
	fmt.Println(strings.SplitN("a/b/c/d/e","/",3))
}
