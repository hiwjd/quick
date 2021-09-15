package main

import (
	"os"
	"os/signal"

	"github.com/hiwjd/quick"
	"github.com/hiwjd/quick/contrib/admin"
	"github.com/hiwjd/quick/support/session"
)

func main() {
	app := quick.New(quick.Config{
		MysqlDSN: "root:@/quick?charset=utf8&parseTime=True&loc=Local",
	})
	app.Migrate(admin.Migrate)
	app.RegisterModules(
		quick.Provide("adminSessionStorage", session.NewRedisStorage("", app.Context().GetRedis())),
		quick.ModuleFunc(admin.AdminModule),
	)
	close := app.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	close()
}
