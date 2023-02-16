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
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
	"telgramTransfer/utils/tool"
	"time"
)

func OnTextMessage(c tb.Context) error {

	//这句话不转发
	if c.Text() == "/start" {
		return nil
	}
	text := c.Text()
	if strings.HasPrefix(text, "绑定商户") || strings.HasPrefix(text, "绑定通道") {
		return binding(c, text)
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
		outChan, err := FindPeerChanel(currentChanel.Name, model.CHANNELTYPE_OUT)
		if err != nil {
			log.Sugar.Errorf("无法找到匹配的群绑定关系：channel_id:%d,tagName:%s", currentChanel.ChannelId, currentChanel.Name)
			return c.Reply("群绑定关系错误,请联系管理员")
		}
		return chanelInTxtHandler(c, currentChanel, *outChan)
	} else if currentChanel.Type == model.CHANNELTYPE_OUT {
		inChan, err := FindPeerChanel(currentChanel.Name, model.CHANNELTYPE_IN)
		if err != nil {
			log.Sugar.Errorf("无法找到匹配的群绑定关系：channel_id:%d,tagName:%s", currentChanel.ChannelId, currentChanel.Name)
			return c.Reply("群绑定关系错误,请联系管理员")
		}
		return chanelOutTxtHandler(c, *inChan, currentChanel)
	}
	return nil
}

// 下游群文本消息处理
func chanelInTxtHandler(c tb.Context, inChanel model.Channel, outChanel model.Channel) error {
	text := c.Text()
	//查单
	if tool.IsOrder(text) {
		return checkOrder(c, text)
	}

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

			//转发到上游群
			to := tb.Chat{ID: outChanel.ChannelId}

			var outChanMsg tb.Message

			jerr := json.Unmarshal([]byte(outChanPhoto.OutMessage), &outChanMsg)
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

			return c.Reply("已加急~ 请耐心等待回复")

		}
	}

	return nil
}

// 上游群文本消息处理
func chanelOutTxtHandler(c tb.Context, inChanel model.Channel, outChanel model.Channel) error {
	return OutChanelResponse(c, inChanel)
}

// 转发从下游群到上游群
func forward(c tb.Context, text string) error {

	//fmt.Println("Order:", text)

	//匹配订单号
	reg := regexp.MustCompile(`^2023\d{16,19}$`)
	result := reg.FindAllStringSubmatch(text, -1)

	//没有匹配上，不是单号信息 不处理
	if result == nil {
		return nil
	}

	groupId := c.Chat().ID

	var dbChannel model.Channel

	err := orm.Gdb.Model(model.Channel{}).First(&dbChannel, "channel_id = ?", groupId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("当前群尚未绑定:%d", groupId)
			return err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return err
	}
	//判断当前群是否是下游群
	if dbChannel.Type != model.CHANNELTYPE_IN {
		//log.Sugar.Infof("在上游群发现单号:%d", groupId)
		return nil
	}

	//查询上下游关系
	channelTagName := dbChannel.Name

	var dbPeerChannel model.Channel
	err = orm.Gdb.Model(model.Channel{}).First(&dbPeerChannel, "name = ? AND type = ?", channelTagName, model.CHANNELTYPE_OUT).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("无法找到对应的上游群:%d", groupId)
			return err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return err
	}

	//开始转发操作
	to := tb.Chat{ID: dbPeerChannel.ChannelId}
	_, err = c.Bot().Forward(&to, c.Message())
	if err != nil {
		log.Sugar.Errorf("Forward error:%s", err.Error())
		return err
	}
	return nil
}

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

	//bindings := strings.Split(text, " ")
	//if len(bindings) != 2 {
	//	log.Sugar.Infof("无法完成查单，原始信息：%s,来自:%s", text, c.Sender().FirstName+c.Sender().LastName)
	//	return nil
	//}

	//auth
	auth := crypt.TokenGenerate()

	client := http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://fourpay-intest.ncjimmy.com/botapi/Xc/Order?id=%s", text), nil)
	req.Header = http.Header{
		"Authorization": {auth},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Sugar.Errorf("查询订单信息失败:%s", err.Error())
		return nil
	}

	type ApiRsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PlatStatus   int `json:"plat_status"`
			NoticeStatus int `json:"notice_status"`
		} `json:"data"`
	}

	rspBytes, _ := io.ReadAll(resp.Body)

	//fmt.Println("RESPONSE:" + string(rspBytes))

	var rsp ApiRsp

	json.Unmarshal(rspBytes, &rsp)

	if rsp.Code != 0 {
		log.Sugar.Errorf("查询订单信息响应错误:%s", rsp.Msg)
		return c.Reply(rsp.Msg)
		//return nil
	}

	message := fmt.Sprintf("订单:%s\n结果:%s,%s", text, PlatStatusRes[rsp.Data.PlatStatus-1], NoticeStatusRes[rsp.Data.NoticeStatus-1])

	//orderNo := bindings[1]

	//c.Send("订单:%s查询中......", orderNo)
	//
	//time.Sleep(2000)

	return c.Reply(message)
}

// 绑定通道|商户
func binding(c tb.Context, text string) error {
	bindings := strings.Split(text, " ")
	if len(bindings) != 2 {
		log.Sugar.Infof("无法完成绑定，原始信息：%s,来自:%s", text, c.Sender().FirstName+c.Sender().LastName)
		return nil
	}

	channelName := bindings[1]

	var bindingType model.ChannelType
	if bindings[0] == "绑定商户" {
		bindingType = model.CHANNELTYPE_IN
	} else {
		bindingType = model.CHANNELTYPE_OUT
	}

	newChanel := model.Channel{
		ChannelId: c.Chat().ID,
		Name:      channelName,
		Type:      bindingType,
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
