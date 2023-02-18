package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"io"
	"net/http"
	"regexp"
	"strings"
	"telgramTransfer/crypt"
	"telgramTransfer/model"
	"telgramTransfer/utils/config"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
	"time"
)

type ApiRspCode struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ApiRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		PlatStatus   int    `json:"plat_status"`
		NoticeStatus int    `json:"notice_status"`
		ChannelId    int64  `json:"channel_id"`
		PlatOrderId  string `json:"plat_order_id"`
	} `json:"data"`
}

//type ApiRspData struct {
//	Data struct {
//		PlatStatus   int   `json:"plat_status"`
//		NoticeStatus int   `json:"notice_status"`
//		ChannelId    int64 `json:"channel_id"`
//	} `json:"data"`
//}

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

//从消息中获取订单号，如果全是返回第一个否则为空
func GetOrderFromText(text string) string {
	orders := strings.Split(text, "\n")

	var isAllAreOrder = true
	for _, order := range orders {
		//检测每一行是否是单号

		match := config.OrderC.Match
		reg := regexp.MustCompile(match)
		result := reg.FindAllStringSubmatch(order, -1)

		//没有匹配上，不是单号信息 不处理
		if result == nil {
			isAllAreOrder = false
			break
		}
	}

	//不是全部都是订单的话就不转发
	if isAllAreOrder {
		return orders[0]
	}
	return ""
}

//判断当前信息是否是订单号
func IsOrder(text string) bool {
	match := config.OrderC.Match
	reg := regexp.MustCompile(match)
	result := reg.FindAllStringSubmatch(text, -1)

	//没有匹配上，不是单号信息 不处理
	if result == nil {
		return false
	}
	return true
}

//通过platId查询channel信息
func FindChanelByPlatId(platId int64) (*model.Channel, error) {
	var dbPeerChannel model.Channel
	err := orm.Gdb.Model(model.Channel{}).First(&dbPeerChannel, "plat_id = ?", platId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("无法找到对应的上游群:%d", platId)
			return nil, err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return nil, err
	}
	return &dbPeerChannel, nil
}

// 通过订单号查询 平台Id 与订单信息
func GetPlatByOrderId(orderId string) (*ApiRsp, error) {
	auth := crypt.TokenGenerate()

	client := http.Client{Timeout: 5 * time.Second}

	var apiUrl = config.Apic.Url

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/botapi/Xc/Order?id=%s", apiUrl, orderId), nil)
	req.Header = http.Header{
		"Authorization": {auth},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Sugar.Errorf("查询订单信息失败:%s", err.Error())
		return nil, err
	}

	rspBytes, _ := io.ReadAll(resp.Body)

	//fmt.Println("RESPONSE:" + string(rspBytes))

	var rsp ApiRspCode

	jerr := json.Unmarshal(rspBytes, &rsp)
	if jerr != nil {
		log.Sugar.Errorf("查询订单响应错误:%s,api msg:%s", string(rspBytes), string(rspBytes))
		return nil, jerr
	}

	if rsp.Code != 0 {
		return nil, errors.New(rsp.Msg)
	}

	var rspData ApiRsp

	jerr = json.Unmarshal(rspBytes, &rspData)
	if jerr != nil {
		log.Sugar.Errorf("理论上不会出现错误:%s,api msg:%s", string(rspBytes), string(rspBytes))
		return nil, jerr
	}

	return &rspData, nil
}

// 通过当前群查询对应的节点群
//func FindPeerChanel(tagName string, channelType model.ChannelType) (*model.Channel, error) {
//	var dbPeerChannel model.Channel
//	err := orm.Gdb.Model(model.Channel{}).First(&dbPeerChannel, "name = ? AND type = ?", tagName, channelType).Error
//
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			log.Sugar.Errorf("无法找到对应的上游群:%s", tagName)
//			return nil, err
//		}
//		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
//		return nil, err
//	}
//	return &dbPeerChannel, nil
//}

// 上游回复消息给下游
func OutChanelResponse(c tb.Context) error {
	//只能上游给下游回复 ??
	if c.Message().ReplyTo != nil {
		var uniqueID string
		if c.Message().ReplyTo.Photo != nil { //回复的是图片信息
			uniqueID = c.Message().ReplyTo.Photo.UniqueID
		} else if c.Message().ReplyTo.Video != nil {
			uniqueID = c.Message().ReplyTo.Video.UniqueID
		}

		var dbPhoto model.Photo
		err := orm.Gdb.Model(model.Photo{}).Where("unique_id = ?", uniqueID).First(&dbPhoto).Error
		if err != nil {
			log.Sugar.Errorf("sql query from photo has error:%s", err.Error())
			return nil
		}

		// 图片是否来自下游群
		if dbPhoto.ID > 0 {

			//反序列化原始消息
			originMessageStr := dbPhoto.InMessage

			var message tb.Message

			err = json.Unmarshal([]byte(originMessageStr), &message)
			if err != nil {
				log.Sugar.Errorf("反序列化原始消息错误:%s", err.Error())
				return err
			}

			//转发到下游群
			to := tb.Chat{ID: message.Chat.ID}

			sendOpts := tb.SendOptions{}
			sendOpts.ReplyTo = &message

			//回复给下游的消息
			var replayMsg interface{}
			//上游回复是否带图片
			if c.Update().Message.Photo != nil {
				photo := c.Message().Photo
				photo.Caption = c.Text()
				replayMsg = photo
			} else if c.Update().Message.Video != nil {
				video := c.Message().Video
				video.Caption = c.Text()
				replayMsg = video
			} else {
				replayMsg = c.Message().Text
			}

			_, err = c.Bot().Send(&to, replayMsg, &sendOpts)
			if err != nil {
				log.Sugar.Errorf("Forward error:%s", err.Error())
				return err
			}

		}
	}

	return nil
}
