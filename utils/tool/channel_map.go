package tool

import (
	"errors"
	"sync"
	"telgramTransfer/model"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
)

var (
	ChannelMap sync.Map //key:int64 value:model.Channel
)

func InitChannelMap() {
	//加载所有的channel
	var channels []model.Channel
	err := orm.Gdb.Model(model.Channel{}).Find(&channels).Error
	if err != nil {
		log.Sugar.Errorf("InitChannelMap error:%s", err.Error())
		panic("InitChannelMap error")
	}

	for _, ch := range channels {
		ChannelMap.Store(ch.ChannelId, ch)
	}

	log.Sugar.Infof("Loaded all channel data")
}

func LoadChannelById(channelId int) (*model.Channel, error) {
	ch, ok := ChannelMap.Load(channelId)
	if ok {
		channel := ch.(model.Channel)
		return &channel, nil
	}
	return nil, errors.New("not found")
}
