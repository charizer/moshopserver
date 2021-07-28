package utils

import (
	"encoding/json"

	"github.com/astaxie/beego"
)

type HTTPData struct {
	ErrNo  int         `json:"errno"`
	ErrMsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type JsonReturn struct {
	Msg  string 	    `json:"msg"`
	Code int		    `json:"code"`
	Data interface{}	`json:"data"`		//Data字段需要设置为interface类型以便接收任意数据
	//json标签意义是定义此结构体解析为json或序列化输出json时value字段对应的key值,如不想此字段被解析可将标签设为`json:"-"`
}


func ApiJsonReturn(this *beego.Controller, msg string,code int,data interface{}) {
	var JsonReturn JsonReturn
	JsonReturn.Msg = msg
	JsonReturn.Code = code
	JsonReturn.Data = data
	this.Data["json"] = JsonReturn		//将结构体数组根据tag解析为json
	this.ServeJSON()					//对json进行序列化输出
}

func ReturnHTTPSuccess(this *beego.Controller, val interface{}) {

	rtndata := HTTPData{
		ErrNo:  0,
		ErrMsg: "",
		Data:   val,
	}

	data, err := json.Marshal(rtndata)
	if err != nil {
		this.Data["json"] = err
	} else {
		this.Data["json"] = json.RawMessage(string(data))
	}
}

func GetHTTPRtnJsonData(errno int, errmsg string) interface{} {

	rtndata := HTTPData{
		ErrNo:  errno,
		ErrMsg: errmsg,
		Data:   nil,
	}
	data, _ := json.Marshal(rtndata)

	return json.RawMessage(string(data))

}
