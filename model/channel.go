package model

import "gorm.io/gorm"

type ChannelType int

const (
	CHANNELTYPE_OUT ChannelType = 0 //上游
	CHANNELTYPE_IN  ChannelType = 1 //下游
)

type Channel struct {
	gorm.Model
	//ID        int64       `gorm:"primaryKey,column:id" json:"id"`
	ChannelId int64       `gorm:"uniqueIndex" json:"channel_id"`
	PlatId    int64       `json:"plat_id"`
	Type      ChannelType `json:"type"`
}

func (Channel) TableName() string {
	return "channels"
}
