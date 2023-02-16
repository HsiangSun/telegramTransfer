package tool

import "regexp"

//判断当前信息是否是订单号
func IsOrder(text string) bool {
	reg := regexp.MustCompile(`^2023\d{16,19}$`)
	result := reg.FindAllStringSubmatch(text, -1)

	//没有匹配上，不是单号信息 不处理
	if result == nil {
		return false
	}
	return true
}
