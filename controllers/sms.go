package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"math/rand"
	"moshopserver/cache"
	"moshopserver/utils"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	//SendURL 发送
	SendURL = "http://api.smsbao.com/sms"

	// VoiceURL 查询
	VoiceURL = "http://www.smsbao.com/voice"

	// QueryURL 查询
	QueryURL = "http://www.smsbao.com/query"

	UserName = "chenzr"

	PassWord = "90d3bbecbd55db0e4089bc372ecd26a6"
)

// 0: 成功
// 30：密码错误
// 40：账号不存在
// 41：余额不足
// 42：帐号过期
// 43：IP地址限制
// 50：内容含有敏感词
// 51：手机号码不正确

type SmsController struct {
	beego.Controller
}
type SmsBody struct {
	Mobile string `json:"mobile"`
}

// Result  返回的结果
type Result struct {
	Code    int
	Message string
}

func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func (this *SmsController) Sms_SendSms() {
	var sb SmsBody
	body := this.Ctx.Input.RequestBody
	json.Unmarshal(body, &sb)
	if !VerifyMobileFormat(sb.Mobile){
		this.CustomAbort(http.StatusBadRequest, "手机号格式错误")
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	cache.MemCachePool.Set(sb.Mobile, rnd, 30)
	content := "【花样百货】尊敬的会员，欢迎使用花样百货，您的验证码为：" + vcode
	query := url.Values{}
	query.Add("u", UserName)
	query.Add("p", PassWord)
	query.Add("m", sb.Mobile)
	query.Add("c", content)
	Result, err := this.queryByURL(SendURL, query)
	if err != nil {
		fmt.Printf("error code:%d err:%s",Result.Code, err.Error())
		this.CustomAbort(Result.Code, err.Error())
	}else{
		utils.ReturnHTTPSuccess(&this.Controller, Result)
		this.ServeJSON()
	}
}

func (this *SmsController) queryByURL(url string, query url.Values) (Result, error) {
	req, err := http.PostForm(url, query)
	if err != nil {
		return Result{Code: req.StatusCode, Message: "请求失败"}, err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return Result{Code: req.StatusCode, Message: "读取失败"}, err
	}
	// TODO 还没有解析结果
	bodyString := string(body)
	if bodyString != "0" {
		errMsg := "短信发送错误码:" + bodyString
		return Result{Code: http.StatusInternalServerError, Message: errMsg}, errors.New(errMsg)
	}
	return Result{Code: req.StatusCode, Message: bodyString}, nil
}
