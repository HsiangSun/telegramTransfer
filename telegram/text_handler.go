package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"telgramTransfer/model"
	"telgramTransfer/utils/config"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
)

func OnTextMessage(c tb.Context) error {

	//这句话不转发
	if c.Text() == "/start" {
		return nil
	}
	text := c.Text()

	if strings.HasPrefix(text, "绑定通道") {
		return bindingChannel(c, text)
	}

	if text == "绑定商户" {
		return bindingMerch(c, text)
	}

	//区分上下游
	groupId := c.Chat().ID
	var currentChanel model.Channel

	err := orm.Gdb.Model(model.Channel{}).First(&currentChanel, "channel_id = ?", groupId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("当前群尚未绑定:%d", groupId)
			return err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return err
	}
	//判断当前群是否是下游群
	if currentChanel.Type == model.CHANNELTYPE_IN {
		return chanelInTxtHandler(c)
	} else if currentChanel.Type == model.CHANNELTYPE_OUT {
		return chanelOutTxtHandler(c)
	}
	return nil
}

// 下游群文本消息处理
func chanelInTxtHandler(c tb.Context) error {
	text := c.Text()
	//查单
	//if IsOrder(text) {
	//	return checkOrder(c, text)
	//}
	if strings.HasPrefix(text, "查单") {
		split := strings.Split(text, " ")
		if len(split) != 2 {
			return nil
		}
		return checkOrder(c, split[1])
	}

	//不是华泰 = 新晨
	if config.Apic.Url != "https://bsh00oo.wxyhome.com" {
		//加急
		if text == "加急" {
			//判断是否是回复某个信息
			if c.Update().Message.ReplyTo != nil {

				//判断回复的是图片还是视频
				var uniqueID = ""
				if c.Update().Message.ReplyTo.Photo != nil {
					uniqueID = c.Update().Message.ReplyTo.Photo.File.UniqueID
				} else if c.Update().Message.ReplyTo.Video != nil {
					uniqueID = c.Update().Message.ReplyTo.Video.File.UniqueID
				}

				//加急
				var outChanPhoto model.Photo
				//拿到原始上游群信息
				err := orm.Gdb.Model(model.Photo{}).Where("unique_id = ?", uniqueID).First(&outChanPhoto).Error
				if err != nil {
					log.Sugar.Errorf("查询上游群图片信息失败:%s", err.Error())
					return err
				}

				//不是订单信息
				if outChanPhoto.ID == 0 {
					return nil
				}

				var outChanMsg tb.Message

				jerr := json.Unmarshal([]byte(outChanPhoto.OutMessage), &outChanMsg)
				if jerr != nil {
					log.Sugar.Errorf("json反序列化上游群消息响应失败:%s", err.Error())
				}

				//转发到上游群
				to := tb.Chat{ID: outChanMsg.Chat.ID}

				sendOpts := tb.SendOptions{}
				sendOpts.ReplyTo = &outChanMsg

				//_, err = c.Bot().Forward(&to, c.Message(), tb.ForceReply)
				_, err = c.Bot().Send(&to, "这笔单子加急!麻烦尽快处理一下!", &sendOpts)
				if err != nil {
					log.Sugar.Errorf("Forward error:%s", err.Error())
					return err
				}

				return c.Reply("已加急~ 请耐心等待回复")

			}
		}
	} else {
		//华泰任何信息都转发
		if c.Update().Message.ReplyTo != nil {

			//判断回复的是图片还是视频
			var uniqueID = ""
			if c.Update().Message.ReplyTo.Photo != nil {
				uniqueID = c.Update().Message.ReplyTo.Photo.File.UniqueID
			} else if c.Update().Message.ReplyTo.Video != nil {
				uniqueID = c.Update().Message.ReplyTo.Video.File.UniqueID
			}

			//加急
			var outChanPhoto model.Photo
			//拿到原始上游群信息
			err := orm.Gdb.Model(model.Photo{}).Where("unique_id = ?", uniqueID).First(&outChanPhoto).Error
			if err != nil {
				log.Sugar.Errorf("查询上游群图片信息失败:%s", err.Error())
				return err
			}

			//不是订单信息
			if outChanPhoto.ID == 0 {
				return nil
			}

			var outChanMsg tb.Message

			jerr := json.Unmarshal([]byte(outChanPhoto.OutMessage), &outChanMsg)
			if jerr != nil {
				log.Sugar.Errorf("json反序列化上游群消息响应失败:%s", err.Error())
			}

			//转发到上游群
			to := tb.Chat{ID: outChanMsg.Chat.ID}

			sendOpts := tb.SendOptions{}
			sendOpts.ReplyTo = &outChanMsg

			//_, err = c.Bot().Forward(&to, c.Message(), tb.ForceReply)
			_, err = c.Bot().Send(&to, c.Message().Text, &sendOpts)
			if err != nil {
				log.Sugar.Errorf("Forward error:%s", err.Error())
				return err
			}

			return c.Reply("请耐心等待回复")

		}
	}

	return nil
}

// 上游群文本消息处理
func chanelOutTxtHandler(c tb.Context) error {
	return OutChanelResponse(c)
}

// 转发从下游群到上游群
//func forward(c tb.Context, text string) error {
//
//	//fmt.Println("Order:", text)
//
//	//匹配订单号
//	reg := regexp.MustCompile(`^2023\d{16,19}$`)
//	result := reg.FindAllStringSubmatch(text, -1)
//
//	//没有匹配上，不是单号信息 不处理
//	if result == nil {
//		return nil
//	}
//
//	groupId := c.Chat().ID
//
//	var dbChannel model.Channel
//
//	err := orm.Gdb.Model(model.Channel{}).First(&dbChannel, "channel_id = ?", groupId).Error
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			log.Sugar.Errorf("当前群尚未绑定:%d", groupId)
//			return err
//		}
//		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
//		return err
//	}
//	//判断当前群是否是下游群
//	if dbChannel.Type != model.CHANNELTYPE_IN {
//		//log.Sugar.Infof("在上游群发现单号:%d", groupId)
//		return nil
//	}
//
//	//查询上下游关系
//	channelTagName := dbChannel.Name
//
//	var dbPeerChannel model.Channel
//	err = orm.Gdb.Model(model.Channel{}).First(&dbPeerChannel, "name = ? AND type = ?", channelTagName, model.CHANNELTYPE_OUT).Error
//
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			log.Sugar.Errorf("无法找到对应的上游群:%d", groupId)
//			return err
//		}
//		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
//		return err
//	}
//
//	//开始转发操作
//	to := tb.Chat{ID: dbPeerChannel.ChannelId}
//	_, err = c.Bot().Forward(&to, c.Message())
//	if err != nil {
//		log.Sugar.Errorf("Forward error:%s", err.Error())
//		return err
//	}
//	return nil
//}

var PlatStatusRes = []string{
	"未支付⭕",
	"支付成功✔",
	"支付失败❌",
}

var NoticeStatusRes = []string{
	"未通知⭕",
	"通知成功✔",
	"通知未成功❌",
}

// 查单
func checkOrder(c tb.Context, text string) error {
	//auth
	rsp, err := GetPlatByOrderId(text)
	if err != nil {
		if strings.Contains(err.Error(), "订单不存在") {
			return c.Reply("请输入正确的订单号哦~")
		}
		log.Sugar.Errorf("查询订单其他错误:%s", err.Error())
		return err
	}

	message := fmt.Sprintf("订单:%s\n结果:%s,%s", text, PlatStatusRes[rsp.Data.PlatStatus-1], NoticeStatusRes[rsp.Data.NoticeStatus-1])

	return c.Reply(message)
}

// 绑定商户
func bindingMerch(c tb.Context, text string) error {

	isAdmin := IsAdmin(c.Sender().Username)

	if !isAdmin {
		return c.Reply("官人,请不要调戏奴家~")
	}
	//bindings := strings.Split(text, " ")
	//if len(bindings) != 2 {
	//	log.Sugar.Infof("无法完成绑定，原始信息：%s,来自:%s", text, c.Sender().FirstName+c.Sender().LastName)
	//	return nil
	//}

	newChanel := model.Channel{
		ChannelId: c.Chat().ID,
		PlatId:    0, //下游没有平台id
		Type:      model.CHANNELTYPE_IN,
	}

	err := orm.Gdb.Model(model.Channel{}).Create(&newChanel).Error
	if err != nil {
		log.Sugar.Errorf("INSERT data to chanels has error:%s", err.Error())
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return c.Reply(fmt.Sprintf("%s失败!已经绑定【%d】过了", text, c.Chat().ID))
		}
		return err
	}
	return c.Reply(fmt.Sprintf("绑定商户成功!成功绑定【%d】", c.Chat().ID))
}

// 绑定通道
func bindingChannel(c tb.Context, text string) error {

	isAdmin := IsAdmin(c.Sender().Username)

	if !isAdmin {
		return c.Reply("官人,请不要调戏奴家~")
	}

	bindings := strings.Split(text, " ")
	if len(bindings) != 2 {
		log.Sugar.Infof("无法完成绑定，原始信息：%s,来自:%s", text, c.Sender().FirstName+c.Sender().LastName)
		return nil
	}

	channelName := bindings[1]

	platId, err2 := strconv.ParseInt(channelName, 10, 64)
	if err2 != nil {
		log.Sugar.Errorf("绑定失败,输入的platId不是数字:%s", err2.Error())
		return c.Reply("绑定失败,请输入数字")
	}

	newChanel := model.Channel{
		ChannelId: c.Chat().ID,
		PlatId:    platId,
		Type:      model.CHANNELTYPE_OUT,
	}

	err := orm.Gdb.Model(model.Channel{}).Create(&newChanel).Error
	if err != nil {
		log.Sugar.Errorf("INSERT data to chanels has error:%s", err.Error())
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return c.Reply(fmt.Sprintf("%s失败!已经绑定【%d】过了", text, c.Chat().ID))
		}
		return err
	}
	return c.Reply(fmt.Sprintf("%s成功!成功绑定【%d】", text, c.Chat().ID))
}
