package telegram

import (
	"errors"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"telgramTransfer/model"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
)

func OnTextMessage(c tb.Context) error {
	text := c.Text()
	//这句话不转发
	if c.Text() == "/start" {
		return nil
	}

	if strings.HasPrefix(text, "绑定") {
		return bindingChannel(c, text)
	}

	//私聊
	if c.Message().Private() {

		//判断是不是管理员
		isAdmin := IsAdmin(c.Sender().Username)
		if !isAdmin {
			return c.Send("官人请不要调戏奴家~")
		}

		//绑定关系查看
		if c.Text() == "/binding" {
			data, err := getBindings(c)
			if err != nil {
				log.Sugar.Errorf("查看绑定错误:%s", err.Error())
			}
			return c.Send(data)
		}

		if strings.HasPrefix(c.Text(), "/delete") {
			return deleteBindings(c)
		}

		if strings.HasPrefix(c.Text(), "/help") {
			return showHelp(c)
		}

		return forward(c)
	}

	return nil
}

//显示帮助信息
func showHelp(c tb.Context) error {

	helpMsg := `
	1.绑定: 绑定 <空格> <分组名>
	2.查看所有绑定: /binding
	3.删除绑定: /delete <空格> id
	4.广播消息: <分组名>:信息内容
`

	return c.Send(helpMsg)

}

func getBindings(c tb.Context) (string, error) {
	var dbChannels []model.Channel

	err := orm.Gdb.Model(model.Channel{}).Find(&dbChannels).Error
	if err != nil {
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return "", err
	}

	resHead := "ID|群名|绑定名|\n"

	sb := strings.Builder{}

	for _, dbChannel := range dbChannels {
		sb.WriteString(fmt.Sprintf("%d|%s|%s\n", dbChannel.ID, dbChannel.GroupName, dbChannel.Name))
	}

	resConten := sb.String()

	return resHead + resConten, nil

}

func deleteBindings(c tb.Context) error {
	splits := strings.Split(c.Text(), " ")
	if len(splits) != 2 {
		return c.Send("指令有误:/delete <空格> id")
	}

	idStr := splits[1]

	idInt, perr := strconv.ParseInt(idStr, 10, 64)
	if perr != nil {
		log.Sugar.Errorf("删除绑定的id不是数字")
		return perr
	}

	err := orm.Gdb.Model(model.Channel{}).Unscoped().Delete(&model.Channel{}, idInt).Error
	if err != nil {
		log.Sugar.Errorf("从Chanel表中删除数据错误:%s", err.Error())
		return c.Send("删除有误,请联系管理员")
	}

	return c.Send("删除成功！")

}

// 转发从下游群到上游群
func forward(c tb.Context) error {

	//获取tag先
	name, content := GetChannelNameAndContent(c.Text())

	groupId := c.Chat().ID

	var dbChannels []model.Channel

	err := orm.Gdb.Model(model.Channel{}).Find(&dbChannels, "name = ?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Errorf("当前群尚未绑定:%d", groupId)
			return err
		}
		log.Sugar.Errorf("从Chanel表中查询数据错误:%s", err.Error())
		return err
	}

	var what interface{}

	//文字消息还是图片消息
	if c.Message().Photo != nil {
		photo := c.Message().Photo
		photo.Caption = content
		what = photo
	} else {
		what = content
	}

	for _, dbChannel := range dbChannels {
		//开始转发操作
		to := tb.Chat{ID: dbChannel.ChannelId}
		_, err = c.Bot().Send(&to, what)
		if err != nil {
			log.Sugar.Errorf("Forward error:%s", err.Error())
		}
	}

	return nil

}

// 绑定通道
func bindingChannel(c tb.Context, text string) error {

	isAdmin := IsAdmin(c.Sender().Username)

	if !isAdmin {
		return c.Reply("官人,请不要调戏奴家~")
	}

	bindings := strings.Split(text, " ")
	if len(bindings) != 2 {
		log.Sugar.Infof("无法完成绑定，原始信息：%s,来自:%s", text, c.Sender().FirstName+c.Sender().LastName)
		return nil
	}

	channelName := bindings[1]

	newChanel := model.Channel{
		ChannelId: c.Chat().ID,
		Name:      channelName,
		GroupName: c.Chat().Title,
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
