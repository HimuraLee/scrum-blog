package md5view

import (
	"testing"
)

func TestEditConfigJS(t *testing.T) {
	file := "../../vueVisitor/docs/.vuepress/config.js"
	key, value := "title", "绯村真之"
	t.Log(EditConfigJS(file, key, value))
}

func TestYarnBuild(t *testing.T) {
	t.Log(YarnBuild())
}