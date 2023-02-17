package model

import "gorm.io/gorm"

// 图片
type Photo struct {
	gorm.Model
	UniqueID   string `gorm:"unique_id" json:"unique_id"`
	InMessage  string `gorm:"type:text" json:"in_message"`
	OutMessage string `gorm:"type:text" json:"out_message"`
}
