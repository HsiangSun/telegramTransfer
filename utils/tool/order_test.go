package tool

import (
	"fmt"
	"telgramTransfer/telegram"
	"testing"
)

func TestIsOrder(t *testing.T) {

	text := "20230211210641015651342放量啦"

	order := telegram.IsOrder(text)

	fmt.Println(order)

}
