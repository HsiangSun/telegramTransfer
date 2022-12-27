package model

type Group struct {
	ID  int64 `gorm:"primaryKey,column:id" json:"id"`
	Gid int64 `gorm:"uniqueIndex" json:"gid"`
}

func (Group) TableName() string {
	return "groups"
}
