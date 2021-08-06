package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/cakturk/go-netstat/netstat"
	"moshopserver/cache"
	_ "moshopserver/models"
	_ "moshopserver/routers"
	"moshopserver/services"
	_ "moshopserver/utils"
	"time"
)

func displaySocket() error {
	tabs, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
		fmt.Printf("tcp socket %+v\n", s)
		return s.LocalAddr.Port == 8089
	})
	if err != nil {
		return err
	}
	tabs, err = netstat.TCP6Socks(func(s *netstat.SockTabEntry) bool {
		fmt.Printf("tcp6 socket %+v\n", s)
		return s.LocalAddr.Port == 8089
	})
	if err != nil {
		return err
	}
	for _, e := range tabs {
		fmt.Printf("result %v\n", e)
	}
	return nil
}




func timerDisplaySocket(){
	d := time.Second * 10
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		<- t.C
		err := displaySocket()
		if err != nil {
			fmt.Println("display socket err:", err.Error())
		}
	}
}

func main() {
	cache.InitMemCache()
	beego.BConfig.WebConfig.AutoRender = false
	beego.BConfig.CopyRequestBody = true

	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.Listen.HTTPAddr = ""
	beego.BConfig.Listen.HTTPPort = 8089

	beego.InsertFilter("/api/*", beego.BeforeExec, services.FilterFunc, true, true)
	go timerDisplaySocket()
	beego.Run() // listen and serve on 0.0.0.0:8080

}
