package telegram

import (
	tb "gopkg.in/telebot.v3"
	"sync"
	"telgramTransfer/model"
	"telgramTransfer/utils/config"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
	"telgramTransfer/utils/tool"
)

func OnTextMessage(c tb.Context) error {

	//这句话不转发
	if c.Text() == "/start" {
		return nil
	}

	admins := config.BootC.Admins
	if c.Text() == "绑定" {

		if !tool.Contains(admins, c.Sender().Username) {
			return c.Reply("官人,请不要调戏奴家~")
		}

		groupId := c.Chat().ID

		group := model.Group{Gid: groupId}
		err := orm.Gdb.Model(model.Group{}).Create(&group).Error
		if err != nil {
			log.Sugar.Errorf("insert group have error:%s", err.Error())
			return err
		}
		return c.Reply("绑定成功~")
	}

	//只有私聊才有用
	if c.Message().Private() && tool.Contains(admins, c.Sender().Username) {

		var groups []model.Group

		//查询所有的已知群组
		err := orm.Gdb.Model(model.Group{}).Find(&groups).Error
		if err != nil {
			log.Sugar.Errorf("查询所有的群组失败：%s", err.Error())
		}

		//协程转发消息
		var wg sync.WaitGroup
		wg.Add(len(groups))
		for _, group := range groups {
			group := group
			go func() {

				defer wg.Done()

				to := tb.Chat{ID: group.Gid}
				c.Bot().Forward(&to, c.Message())

			}()
		}
		wg.Wait()
	}

	return nil
}
