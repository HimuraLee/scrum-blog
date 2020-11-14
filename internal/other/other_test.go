package other

import (
	"testing"
)

func TestPathExist(t *testing.T) {
	exit, err := PathExist("../others")
	t.Log(exit, err)
	exit, err = PathExist("../others_fault")
	t.Log(exit, err)
}