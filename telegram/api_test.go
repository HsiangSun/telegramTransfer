package telegram

import (
	"fmt"
	"strings"
	"testing"
)

func TestForwardTag(t *testing.T) {
	tag := "哈哈:我爱你呀,这是我的证据:1.aa2.bbb"

	index := strings.IndexRune(tag, ':')

	name := tag[:index]

	res := tag[index+1:]

	fmt.Println(res)
	fmt.Println(name)

}
