package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/fatih/structs"
	"moshopserver/models"
	"moshopserver/utils"
)

type AddressController struct {
	beego.Controller
}

type AddressListRtnJson struct {
	models.NideshopAddress
	ProviceName  string `json:"provice_name"`
	CityName     string `json:"city_name"`
	DistrictName string `json:"district_name"`
	FullRegion   string `json:"full_region"`
}

func (this *AddressController) Address_List() {
	userId, err := getUserIdFromJwt(this.Ctx)
	if err != nil {
		this.CustomAbort(401, "token失效")
	}
	o := orm.NewOrm()
	addresstable := new(models.NideshopAddress)
	var addresses []models.NideshopAddress

	_, err = o.QueryTable(addresstable).Filter("user_id", userId).OrderBy("-is_default").All(&addresses)
	if err != nil {
		fmt.Println(err.Error())
	}
	/*rtnaddress := make([]AddressListRtnJson, 0)

	for _, val := range addresses {

		provicename := models.GetRegionName(val.ProvinceId)
		cityname := models.GetRegionName(val.CityId)
		distinctname := models.GetRegionName(val.DistrictId)
		rtnaddress = append(rtnaddress, AddressListRtnJson{
			NideshopAddress: val,
			ProviceName:     provicename,
			CityName:        cityname,
			DistrictName:    distinctname,
			FullRegion:      provicename + cityname + distinctname,
		})

	}*/

	utils.ReturnHTTPSuccess(&this.Controller, addresses)
	this.ServeJSON()

}
func (this *AddressController) Address_Detail() {
	id := this.GetString("id")

	intid := utils.String2Int(id)

	o := orm.NewOrm()
	addresstable := new(models.NideshopAddress)
	var address models.NideshopAddress
	userId, _ := getUserIdFromJwt(this.Ctx)
	/*err := o.QueryTable(addresstable).Filter("id", intid).Filter("user_id", userId).One(&address)
	if err != nil {
		fmt.Printf("address id:%d err:%s", intid, err.Error())
	}*/
	var err error
	var addresses []models.NideshopAddress
	if id == "" {
		_, err = o.QueryTable(addresstable).Filter("user_id", userId).OrderBy("-is_default").All(&addresses)
		if err != nil && err != orm.ErrNoRows {
			fmt.Println("check order get addr err:%s", err.Error())
		}
		if len(addresses) > 0 {
			address = addresses[0]
		}
	} else {
		err = o.QueryTable(addresstable).Filter("id", intid).Filter("user_id", userId).One(&address)
	}
	if err != orm.ErrNoRows {

	}
	/*var val AddressListRtnJson

	if err != orm.ErrNoRows {

		provicename := models.GetRegionName(address.ProvinceId)
		cityname := models.GetRegionName(address.CityId)
		distinctname := models.GetRegionName(address.DistrictId)
		val = AddressListRtnJson{
			NideshopAddress: address,
			ProviceName:     provicename,
			CityName:        cityname,
			DistrictName:    distinctname,
			FullRegion:      provicename + cityname + distinctname,
		}
	}*/
	utils.ReturnHTTPSuccess(&this.Controller, address)
	this.ServeJSON()
}

type AddressSaveBody struct {
	Address    string `json:"address"`
	CityId     int    `json:"city_id"`
	DistrictId int    `json:"district_id"`
	IsDefault  bool   `json:"is_default"`
	Mobile     string `json:"mobile"`
	Name       string `json:"name"`
	ProvinceId int    `json:"province_id"`
	AddressId  int    `json:"id"`
	ProvinceName string `json:"province_name"`
	CityName     string `json:"city_name"`
	DistrictName string `json:"district_name"`
}

func (this *AddressController) Address_Save() {

	var asb AddressSaveBody
	body := this.Ctx.Input.RequestBody
	json.Unmarshal(body, &asb)

	address := asb.Address
	name := asb.Name
	mobile := asb.Mobile
	provinceid := asb.ProvinceId
	cityid := asb.CityId
	distinctid := asb.DistrictId
	isdefault := asb.IsDefault
	addressid := asb.AddressId

	userid, err := getUserIdFromJwt(this.Ctx)
	if err != nil {
		this.CustomAbort(401, "token失效")
	}
	/*var intisdefault int
	if isdefault {
		intisdefault = 1
	} else {
		intisdefault = 0
	}*/

	intcityid := cityid
	intprovinceid := provinceid
	intdistinctid := distinctid

	addressdata := models.NideshopAddress{
		Address:    address,
		CityId:     intcityid,
		DistrictId: intdistinctid,
		ProvinceId: intprovinceid,
		Name:       name,
		Mobile:     mobile,
		UserId:     userid,
		IsDefault:  isdefault,
		ProvinceName: asb.ProvinceName,
		CityName:        asb.CityName,
		DistrictName:    asb.DistrictName,
	}
	o := orm.NewOrm()
	addresstable := new(models.NideshopAddress)


	if addressid == 0 {
		id, err := o.Insert(&addressdata)
		if err == nil {
			addressid = int(id)
		}
		fmt.Println("insert address")
	} else {
		/*o.QueryTable(addresstable).Filter("id", intid).Filter("user_id", userid).Update(orm.Params{
			"is_default": 0,
		})*/
		addressdata.Id = addressid
		_, err := o.QueryTable(addresstable).Filter("id", addressid).Filter("user_id", userid).Update(structs.Map(addressdata))
		if err != nil {
			fmt.Println("update address err:", err.Error())
		}else {
			fmt.Println("update address:", addressid)
		}
	}
	if isdefault {
		_, err := o.Raw("UPDATE nideshop_address SET is_default = false where user_id = ? and id <> ? ", userid, addressid).Exec()
		if err != nil {
			fmt.Println("update err:", err.Error())
			//res.RowsAffected()
			//fmt.Println("mysql row affected nums: ", num)
		}
	}
	var addressinfo models.NideshopAddress
	o.QueryTable(addresstable).Filter("id", addressid).One(&addressinfo)

	utils.ReturnHTTPSuccess(&this.Controller, addressinfo)
	this.ServeJSON()

}

func (this *AddressController) Address_Delete() {

	addressid := this.GetString("id")
	intaddressid := utils.String2Int(addressid)
	userid, err := getUserIdFromJwt(this.Ctx)
	if err != nil {
		this.CustomAbort(401, "token失效")
	}
	o := orm.NewOrm()
	addresstable := new(models.NideshopAddress)
	fmt.Println("delete address id:", intaddressid)
	o.QueryTable(addresstable).Filter("id", intaddressid).Filter("user_id", userid).Delete()
	utils.ReturnHTTPSuccess(&this.Controller, nil)
	this.ServeJSON()
	return

}
