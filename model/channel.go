package model

import "gorm.io/gorm"

type ChannelType int

const (
	CHANNELTYPE_OUT ChannelType = 0
	CHANNELTYPE_IN  ChannelType = 1
)

type Channel struct {
	gorm.Model
	//ID        int64       `gorm:"primaryKey,column:id" json:"id"`
	ChannelId int64       `gorm:"uniqueIndex" json:"channel_id"`
	Name      string      `json:"name"`
	Type      ChannelType `json:"type"`
}

func (Channel) TableName() string {
	return "channels"
}
