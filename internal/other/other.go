package other

import (
	"blog/internal/rate"
	"os"
)

var LoginLimiter = rate.NewLimiter(20, 5)


func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}