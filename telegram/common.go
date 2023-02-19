package telegram

import (
	"strings"
	"telgramTransfer/utils/config"
)

//当前用户是否是系统管理员
func IsAdmin(username string) bool {
	var res = false
	admins := config.BootC.Admins
	for _, admin := range admins {
		if username == admin {
			res = true
			break
		}
	}
	return res
}

func GetChannelNameAndContent(text string) (channelName, content string) {

	index := strings.IndexRune(text, ':')

	if index == -1 {
		return "", ""
	}

	channelName = text[:index]
	content = text[index+1:]

	return
}
