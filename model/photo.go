package model

import "gorm.io/gorm"

// 图片
type Photo struct {
	gorm.Model
	UniqueID   string `gorm:"uniqueIndex" json:"unique_id"`
	InMessage  string `gorm:"type:text" json:"in_message"`
	OutMessage string `gorm:"type:text" json:"out_message"`
}
