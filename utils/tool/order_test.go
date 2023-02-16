package tool

import (
	"fmt"
	"testing"
)

func TestIsOrder(t *testing.T) {

	text := "20230211210641015651342放量啦"

	order := IsOrder(text)

	fmt.Println(order)

}
