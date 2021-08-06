package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"moshopserver/cache"
	_ "moshopserver/models"
	"moshopserver/netstat"
	_ "moshopserver/routers"
	"moshopserver/services"
	_ "moshopserver/utils"
	"runtime"
	"time"
)

func displaySocket() error {
	tabs, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
		return s.LocalAddr.Port == 8089
	})
	if err != nil {
		return err
	}
	for _, e := range tabs {
		fmt.Printf("tcp result port:%d state:%s\n", e.LocalAddr, e.State.String())
	}
	tabs, err = netstat.TCP6Socks(func(s *netstat.SockTabEntry) bool {
		return s.LocalAddr.Port == 8089
	})
	if err != nil {
		return err
	}
	for _, e := range tabs {
		fmt.Printf("tcp6 result port:%d state:%s\n", e.LocalAddr, e.State.String())
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
	fmt.Println("os:", runtime.GOOS)
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
