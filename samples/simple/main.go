package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hiwjd/quick"
	"github.com/labstack/echo/v4"
)

func main() {
	app := quick.New(quick.Config{})
	app.RegisterModules(
		quick.ModuleFunc(func(ac quick.Context) {
			timer := time.NewTicker(time.Second)
			count := 0
			go func() {
				for range timer.C {
					count++
				}
			}()
			ac.RegisterShutdown(func() {
				timer.Stop()
			})
			ac.GET("/now", func(c echo.Context) error {
				return c.String(200, fmt.Sprintf("Count: %d", count))
			})
		}),
	)
	close := app.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	close()
}
