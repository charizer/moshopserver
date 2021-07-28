package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"moshopserver/models"
	"moshopserver/services"
	"moshopserver/utils"
)

type AuthController struct {
	beego.Controller
}

type AuthLoginBody struct {
	Code     string               `json:"code"`
	UserInfo services.ResUserInfo `json:"userInfo"`
}

func (this *AuthController) Auth_LoginByWeixin() {

	var alb AuthLoginBody
	body := this.Ctx.Input.RequestBody

	err := json.Unmarshal(body, &alb)
	//fmt.Print(alb)
	clientIP := this.Ctx.Input.IP()

	/*userInfo := services.Login(alb.Code, alb.UserInfo)
	if userInfo == nil {

	}*/

	o := orm.NewOrm()

	var user models.NideshopUser
	usertable := new(models.NideshopUser)
	err = o.QueryTable(usertable).Filter("mobile", alb.UserInfo.Mobile).One(&user)
	if err == orm.ErrNoRows {
		newuser := models.NideshopUser{Username: alb.UserInfo.Mobile, Password: "", RegisterTime: utils.GetTimestamp(),
			RegisterIp: clientIP, Mobile: alb.UserInfo.Mobile, WeixinOpenid: "", Avatar: "", Gender: 1,
			Nickname: alb.UserInfo.Mobile}
		o.Insert(&newuser)
		err = o.QueryTable(usertable).Filter("mobile", alb.UserInfo.Mobile).One(&user)
		if err == orm.ErrNoRows {
			fmt.Println("no user")
		}
	}

	userinfo := make(map[string]interface{})
	userinfo["id"] = user.Id
	userinfo["username"] = user.Username
	userinfo["nickname"] = user.Nickname
	userinfo["gender"] = user.Gender
	userinfo["avatar"] = user.Avatar
	userinfo["birthday"] = user.Birthday
	userinfo["mobile"] = user.Mobile
	user.LastLoginIp = clientIP
	user.LastLoginTime = utils.GetTimestamp()

	if _, err := o.Update(&user); err == nil {

	}

	sessionKey := services.Create(utils.Int2String(user.Id))
	//fmt.Println("sessionkey==" + sessionKey)

	rtnInfo := make(map[string]interface{})
	rtnInfo["token"] = sessionKey
	rtnInfo["userInfo"] = userinfo

	utils.ReturnHTTPSuccess(&this.Controller, rtnInfo)
	this.ServeJSON()

}
