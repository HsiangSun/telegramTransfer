package telegram

import (
	"encoding/json"
	"errors"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"telgramTransfer/model"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
)

func OnImageMessage(c tb.Context) error {

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
		outChan, err := FindPeerChanel(currentChanel.Name, model.CHANNELTYPE_OUT)
		if err != nil {
			log.Sugar.Errorf("无法找到匹配的群绑定关系：channel_id:%d,tagName:%s", currentChanel.ChannelId, currentChanel.Name)
			return c.Reply("群绑定关系错误,请联系管理员")
		}
		return chanelInImgHandler(c, currentChanel, *outChan)
	} else if currentChanel.Type == model.CHANNELTYPE_OUT {
		inChan, err := FindPeerChanel(currentChanel.Name, model.CHANNELTYPE_IN)
		if err != nil {
			log.Sugar.Errorf("无法找到匹配的群绑定关系：channel_id:%d,tagName:%s", currentChanel.ChannelId, currentChanel.Name)
			return c.Reply("群绑定关系错误,请联系管理员")
		}
		return chanelOutImgHandler(c, *inChan, currentChanel)
	}
	return nil
}

//上游操作
func chanelOutImgHandler(c tb.Context, inChanel model.Channel, outChanel model.Channel) error {
	return OutChanelResponse(c, inChanel)
}

//下游操作
func chanelInImgHandler(c tb.Context, inChanel model.Channel, outChanel model.Channel) error {
	text := c.Text()

	orders := strings.Split(text, "\n")

	var isAllAreOrder = true

	for _, order := range orders {
		//检测每一行是否是单号
		reg := regexp.MustCompile(`^2023\d{16,19}$`)
		result := reg.FindAllStringSubmatch(order, -1)

		//没有匹配上，不是单号信息 不处理
		if result == nil {
			isAllAreOrder = false
			break
		}
	}

	//不是全部都是订单的话就不转发
	if !isAllAreOrder {
		return nil
	}

	//--------------是订单内容并且有图片---------------------

	//var newPhoto model.Photo

	//如果没有图片就不处理
	if c.Update().Message.Photo == nil {
		return nil
	}

	originMessage := c.Message()
	//fmt.Printf("originMessage:%+v \n", *originMessage)

	originMessageBytes, jerr := json.Marshal(*originMessage)
	if jerr != nil {
		log.Sugar.Errorf("josn 序列化原始消息错误:%s", jerr.Error())
		return nil
	}

	//保存原始信息到数据库

	uniqueID := c.Update().Message.Photo.File.UniqueID

	newPhoto := model.Photo{UniqueID: uniqueID, InMessage: string(originMessageBytes)}

	err := orm.Gdb.Model(model.Photo{}).Create(&newPhoto).Error
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {

			//加急
			var outChanPhoto model.Photo
			//拿到原始上游群信息
			err = orm.Gdb.Model(model.Photo{}).First(&outChanPhoto).Where("unique_id = ?", uniqueID).Error
			if err != nil {
				log.Sugar.Errorf("查询上游群图片信息失败:%s", err.Error())
				return err
			}

			//转发到上游群
			to := tb.Chat{ID: outChanel.ChannelId}

			var outChanMsg tb.Message

			jerr = json.Unmarshal([]byte(outChanPhoto.OutMessage), &outChanMsg)
			if jerr != nil {
				log.Sugar.Errorf("json反序列化上游群消息响应失败:%s", err.Error())
			}

			sendOpts := tb.SendOptions{}
			sendOpts.ReplyTo = &outChanMsg

			//_, err = c.Bot().Forward(&to, c.Message(), tb.ForceReply)
			_, err = c.Bot().Send(&to, "这笔单子加急!麻烦尽快处理一下!", &sendOpts)
			if err != nil {
				log.Sugar.Errorf("Forward error:%s", err.Error())
				return err
			}

			return c.Reply("已加急~ 请勿重复发送图片哦")
		}

		log.Sugar.Errorf("insert data to photo has error:%s", err.Error())

		return err
	}

	//开始转发操作
	to := tb.Chat{ID: outChanel.ChannelId}

	photo := c.Message().Photo
	photo.Caption = c.Text()

	outMessage, err := c.Bot().Send(&to, photo) //tb.Protected
	if err != nil {
		log.Sugar.Errorf("Forward error:%s", err.Error())
		return err
	}

	originOutMessageBytes, jerr := json.Marshal(*outMessage)
	if jerr != nil {
		log.Sugar.Errorf("josn 序列化原始响应消息错误:%s", jerr.Error())
		return nil
	}

	//保存上游信息
	err = orm.Gdb.Model(model.Photo{}).Where("id = ?", newPhoto.ID).Update("out_message", string(originOutMessageBytes)).Error
	if err != nil {
		log.Sugar.Errorf("update data to photo has error:%s", err.Error())
		return err
	}

	return c.Reply("正在为你处理中... 请稍等~")
}