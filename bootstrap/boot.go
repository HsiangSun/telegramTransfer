package bootstrap

import (
	"os"
	"os/signal"
	"syscall"
	"telgramTransfer/telegram"
	"telgramTransfer/utils/config"
	"telgramTransfer/utils/log"
	"telgramTransfer/utils/orm"
	"telgramTransfer/utils/tool"
)

func init() {
	config.InitConfig()
	log.InitLog()
	orm.InitDb()
	//暂时没用
	tool.InitChannelMap()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				//log.Printf("server bot err:%+v \n", err)
				log.Sugar.Error("****************Bot have error will restart*********************")
			}
		}()
		telegram.BootRun()
	}()

	//kill program
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

}
