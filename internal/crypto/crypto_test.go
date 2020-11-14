package crypto

import (
	"github.com/thoas/go-funk"
	"testing"
)

func Test_Hmc(t *testing.T) {
	set := make([]rune, 10)
	for i := 0; i < 10; i++ {
		set[i] = rune('0' + i)
	}
	rnd := funk.RandomString(5, set)
	vreal := Hmc(rnd, "v.c.o.d.e")
	t.Log(vreal)
}

func Test_Passwd2Md5(t *testing.T) {
	passwd, salt := "123456", "abc"
	t.Log(Passwd2Md5(passwd, salt))
}