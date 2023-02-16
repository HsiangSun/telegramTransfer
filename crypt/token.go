package crypt

import (
	"fmt"
	"strconv"
	"time"
)

func TokenGenerate() string {
	timeStep := fmt.Sprintf("%d", time.Now().Unix())

	//加密
	str, _ := EncryptByAes([]byte(timeStep))
	return str
}

func TokenVerify(token string) bool {
	timeStep, err := DecryptByAes(token)
	if err != nil {
		return false
	}

	timeStepInt, err := strconv.ParseInt(string(timeStep), 10, 64)
	if err != nil {
		return false
	}

	return time.Now().Unix()-timeStepInt < 3

}
