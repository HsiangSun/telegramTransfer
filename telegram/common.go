package telegram

import (
	"encoding/json"
	"errors"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"telgramTransfer/model"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
)

// 通过当前群查询对应的节点群
func FindPeerChanel(tagName string, channelType model.ChannelType) (*model.Channel, error) {
	var dbPeerChannel model.Channel
	err := orm.Gdb.Model(model.Channel{}).First(&dbPeerChannel, "name = ? AND type = ?", tagName, channelType).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("无法找到对应的上游群:%s", tagName)
			return nil, err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return nil, err
	}
	return &dbPeerChannel, nil
}

// 上游回复消息给下游
func OutChanelResponse(c tb.Context, inChanel model.Channel) error {
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
			to := tb.Chat{ID: inChanel.ChannelId}

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
