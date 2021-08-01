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

type OrderController struct {
	beego.Controller
}

//It may need to be refactored.
func GetOrderPageData(rawData []models.NideshopOrder, page int, size int) utils.PageData {

	count := len(rawData)
	totalpages := (count + size - 1) / size
	var pagedata []models.NideshopOrder

	for idx := (page - 1) * size; idx < page*size && idx < count; idx++ {
		pagedata = append(pagedata, rawData[idx])
	}

	return utils.PageData{NumsPerPage: size, CurrentPage: page, Count: int64(count), TotalPages: int64(totalpages), Data: pagedata}
}

type OrderListRtnJson struct {
	models.NideshopOrder
	GoodsList       []models.NideshopOrderGoods `json:"goodList"`
	GoodsCount      int                         `json:"goodsCount"`
	OrderStatusText string                      `json:"order_status_text"`
	HandOption      models.OrderHandleOption    `json:"handleOption"`
	OrderDate       string 						`json:"order_date"`
}

func (this *OrderController) Order_List() {
	page := utils.String2Int(this.GetString("page"))
	if page == -1{
		page = 1
	}
	size := utils.String2Int(this.GetString("size"))
	if size == -1 {
		size = 10
	}
	status := utils.String2Int(this.GetString("status"))
	userId, _ := getUserIdFromJwt(this.Ctx)
	o := orm.NewOrm()
	ordertable := new(models.NideshopOrder)
	var orders []models.NideshopOrder
	pageCount := int64(0)
	var err error
	switch status {
	case models.ORDER_ALL:
		pageCount, err = o.QueryTable(ordertable).Filter("user_id", userId).Exclude("order_status", models.ORDER_DELETE).OrderBy("-add_time").Offset((page-1)*size).Limit(size).All(&orders)
	case models.ORDER_FINISH:
		statusIds := []int{models.ORDER_CANCEL, models.ORDER_SUCC}
		pageCount, err = o.QueryTable(ordertable).Filter("user_id", userId).Filter("order_status__in",statusIds).OrderBy("-add_time").Offset((page-1)*size).Limit(size).All(&orders)
	default:
		pageCount, err = o.QueryTable(ordertable).Filter("user_id", userId).Filter("order_status", status).OrderBy("-add_time").Offset((page-1)*size).Limit(size).All(&orders)
	}
	if err != nil {
		fmt.Println("get order list err", err.Error())
	}
	count , err := o.QueryTable(ordertable).Filter("user_id", userId).Count()
	if err != nil {
		fmt.Println("count order err", err.Error())
	}
	orderIds := make([]int,0,pageCount)
	for _, order := range orders{
		orderIds = append(orderIds, order.Id)
	}
	if len(orderIds) <= 0 {
		pageData := utils.PageData {
			NumsPerPage : size,
			CurrentPage: page,
			Count: count,
			TotalPages: 0,
			Data: []OrderListRtnJson{},
		}
		utils.ReturnHTTPSuccess(&this.Controller, pageData)
		this.ServeJSON()
	}
	//firstpagedorders := GetOrderPageData(orders, 1, 10)

	var rtnorderlist []OrderListRtnJson
	ordergoodstable := new(models.NideshopOrderGoods)
	var ordergoods []models.NideshopOrderGoods
	o.QueryTable(ordergoodstable).Filter("order_id__in", orderIds).All(&ordergoods)
	/*for _, val := range orders {
		qsordergoods.Filter("order_id", val.Id).All(&ordergoods)
		var goodscount int
		for _, val := range ordergoods {
			goodscount += val.Number
		}
		orderstatustext := models.GetOrderStatusText(val.Id)
		orderhandoption := models.GetOrderHandleOption(val.Id)
		orderlistrtn := OrderListRtnJson{val, ordergoods, goodscount, orderstatustext, orderhandoption}
		rtnorderlist = append(rtnorderlist, orderlistrtn)
	}*/
	hash := getOrderGoods(ordergoods)
	for _, val := range orders {
		orderDate := utils.FormatTimestamp(val.AddTime, "2006-01-02 15:04:05")
		orderstatustext := models.GetOrderStatusText(val.Id)
		orderhandoption := models.GetOrderHandleOption(val.Id)
		orderlistrtn := OrderListRtnJson{val, hash[val.Id], len(hash[val.Id]), orderstatustext, orderhandoption, orderDate}
		rtnorderlist = append(rtnorderlist, orderlistrtn)
	}
	totalPages := (count + int64(size) - 1) / int64(size)
	//orders.Data = rtnorderlist
	pageData := utils.PageData {
		NumsPerPage : size,
		CurrentPage: page,
		Count: count,
		TotalPages: totalPages,
		Data:rtnorderlist,
	}
	utils.ReturnHTTPSuccess(&this.Controller, pageData)
	this.ServeJSON()
}

func getOrderGoods(goods []models.NideshopOrderGoods) map[int][]models.NideshopOrderGoods{
	hash := make(map[int][]models.NideshopOrderGoods)
	for _, good := range goods {
		if _, ok := hash[good.OrderId]; ok {
			hash[good.OrderId] = append(hash[good.OrderId], good)
		}else{
			hash[good.OrderId] = []models.NideshopOrderGoods{}
			hash[good.OrderId] = append(hash[good.OrderId], good)
		}
	}
	return hash
}

type OrderInfo struct {
	models.NideshopOrder
	/*ProvinceName        string                  `json:"province_name"`
	CityName            string                  `json:"city_name"`
	DistrictName        string                  `json:"district_name"`*/
	FullRegion          string                  `json:"full_region"`
	Express             services.ExpressRtnInfo `json:"express"`
	OrderStatusText     string                  `json:"order_status_text"`
	FormatAddTime       string                  `json:"add_time"`
	FormatFinalPlayTime string                  `json:"final_pay_time"`
}

type OrderDetailRtnJson struct {
	OrderInfo    OrderInfo                   `json:"orderInfo"`
	OrderGoods   []models.NideshopOrderGoods `json:"orderGoods"`
	HandleOption models.OrderHandleOption    `json:"handleOption"`
}

func (this *OrderController) Order_Detail() {

	orderId := this.GetString("orderId")
	intorderId := utils.String2Int(orderId)
	userId, _ := getUserIdFromJwt(this.Ctx)
	o := orm.NewOrm()
	ordertable := new(models.NideshopOrder)
	var order models.NideshopOrder
	err := o.QueryTable(ordertable).Filter("id", intorderId).Filter("user_id", userId).One(&order)

	if err == orm.ErrNoRows {
		this.Abort("订单不存在")
	}

	orderinfo:= OrderInfo{NideshopOrder: order}
	/*orderinfo.ProvinceName = models.GetRegionName(order.Province)
	orderinfo.CityName = models.GetRegionName(order.City)
	orderinfo.DistrictName = models.GetRegionName(order.District)
	orderinfo.FullRegion = orderinfo.ProvinceName + orderinfo.CityName + orderinfo.DistrictName*/

	/*lastestexpressinfo := models.GetLatestOrderExpress(intorderId)
	orderinfo.Express = lastestexpressinfo*/

	ordergoodstable := new(models.NideshopOrderGoods)
	var ordergoods []models.NideshopOrderGoods

	o.QueryTable(ordergoodstable).Filter("order_id", intorderId).All(&ordergoods)

	orderinfo.OrderStatusText = models.GetOrderStatusText(intorderId)
	orderinfo.FormatAddTime = utils.FormatTimestamp(orderinfo.AddTime, "2006-01-02 15:04:05")
	orderinfo.FormatFinalPlayTime = utils.FormatTimestamp(1234, "04:05")

	if orderinfo.OrderStatus == 0 {
		//todo 订单超时逻辑
	}

	handleoption := models.GetOrderHandleOption(intorderId)
	utils.ReturnHTTPSuccess(&this.Controller, OrderDetailRtnJson{
		OrderInfo:    orderinfo,
		OrderGoods:   ordergoods,
		HandleOption: handleoption,
	})
	this.ServeJSON()
}

type SubmitOrderBody struct {
	AddressId int `json:"addressId"`
	GoodsId   int `json:"goodsId"`
	Number    int `json:"number"`
	ProductId int `json:"productId"`
}

type UpdateOrderStatusBody struct {
	OrderId int `json:"orderId"`
	Status  int `json:"status"`
}

func (this *OrderController) submitQuick(address models.NideshopAddress, intGoodsId, number,productId int){
	userId, _ := getUserIdFromJwt(this.Ctx)
	o := orm.NewOrm()
	producttable := new(models.NideshopProduct)
	var product models.NideshopProduct
	err := o.QueryTable(producttable).Filter("goods_id", intGoodsId).Filter("id", productId).One(&product)
	if err == orm.ErrNoRows || product.GoodsNumber < number {
		this.CustomAbort(400, "库存不足")
	}
	var goodone models.NideshopGoods
	good := new(models.NideshopGoods)
	o.QueryTable(good).Filter("id", intGoodsId).One(&goodone)
	var freightPrice float64 = 0
	var goodstotalprice float64 = 0
	goodstotalprice += float64(number) * product.RetailPrice
	var couponprice float64
	ordertotalprice := goodstotalprice + freightPrice - couponprice
	actualprice := ordertotalprice - 0
	currenttime := utils.GetTimestamp()
	postscript := ""
	orderinfo := models.NideshopOrder{
		OrderSn:      models.GenerateOrderNumber(),
		UserId:       userId,
		Consignee:    address.Name,
		Mobile:       address.Mobile,
		Province:     address.ProvinceId,
		City:         address.CityId,
		District:     address.DistrictId,
		Address:      address.Address,
		FreightPrice: 0,
		Postscript:   postscript,
		CouponId:     0,
		CouponPrice:  couponprice,
		AddTime:      currenttime,
		GoodsPrice:   goodstotalprice,
		OrderPrice:   ordertotalprice,
		ActualPrice:  actualprice,
		ProvinceName: address.ProvinceName,
		CityName: address.Address,
		DistrictName: address.DistrictName,
		CallbackStatus: "true",
	}

	orderid, err := o.Insert(&orderinfo)
	if err != nil {
		this.Abort("订单提交失败")
	}
	orderinfo.Id = int(orderid)

	ordergood := models.NideshopOrderGoods{
		OrderId:                   int(orderid),
		GoodsId:                   goodone.Id,
		GoodsSn:                   goodone.GoodsSn,
		ProductId:                 productId,
		GoodsName:                 goodone.Name,
		ListPicUrl:                goodone.ListPicUrl,
		MarketPrice:           	   product.RetailPrice,
		RetailPrice:               product.RetailPrice,
		Number:                    number,
		GoodsSpecifitionNameValue: "",
		GoodsSpecifitionIds:       "",
	}
	o.Insert(&ordergood)
	utils.ReturnHTTPSuccess(&this.Controller, orderinfo)
	this.ServeJSON()
}

func (this *OrderController) Order_Submit() {
	var sob SubmitOrderBody
	body := this.Ctx.Input.RequestBody
	json.Unmarshal(body, &sob)

	//addressId := this.GetString("addressId")
	//couponId := this.GetString("couponId")
	postscript := ""//this.GetString("postscript")
	intaddressId := sob.AddressId//utils.String2Int(addressId)
	//intcouponId := utils.String2Int(couponId)
	intGoodsId := sob.GoodsId
	number := sob.Number
	productId := sob.ProductId
	o := orm.NewOrm()
	addresstable := new(models.NideshopAddress)
	var address models.NideshopAddress

	err := o.QueryTable(addresstable).Filter("id", intaddressId).One(&address)
	if err == orm.ErrNoRows {
		this.Abort("请选择收获地址")
	}
	userId, _ := getUserIdFromJwt(this.Ctx)
	carttable := new(models.NideshopCart)
	var carts []models.NideshopCart
	if intGoodsId != 0 {
		this.submitQuick(address, intGoodsId, number, productId)
		return
	}else{
		_, err = o.QueryTable(carttable).Filter("user_id", userId).Filter("session_id", 1).Filter("checked", 1).All(&carts)
		if err == orm.ErrNoRows {
			this.Abort("请选择商品")
		}
	}

	var freightPrice float64 = 0
	var goodstotalprice float64 = 0

	for _, val := range carts {
		goodstotalprice += float64(val.Number) * val.RetailPrice
	}

	var couponprice float64
	ordertotalprice := goodstotalprice + freightPrice - couponprice
	actualprice := ordertotalprice - 0
	currenttime := utils.GetTimestamp()

	orderinfo := models.NideshopOrder{
		OrderSn:      models.GenerateOrderNumber(),
		UserId:       userId,
		Consignee:    address.Name,
		Mobile:       address.Mobile,
		Province:     address.ProvinceId,
		City:         address.CityId,
		District:     address.DistrictId,
		Address:      address.Address,
		FreightPrice: 0,
		Postscript:   postscript,
		CouponId:     0,
		CouponPrice:  couponprice,
		AddTime:      currenttime,
		GoodsPrice:   goodstotalprice,
		OrderPrice:   ordertotalprice,
		ActualPrice:  actualprice,
		ProvinceName: address.ProvinceName,
		CityName: address.Address,
		DistrictName: address.DistrictName,
		CallbackStatus: "true",
	}

	orderid, err := o.Insert(&orderinfo)
	if err != nil {
		this.Abort("订单提交失败")
	}
	orderinfo.Id = int(orderid)

	for _, item := range carts {
		ordergood := models.NideshopOrderGoods{
			OrderId:                   int(orderid),
			GoodsId:                   item.GoodsId,
			GoodsSn:                   item.GoodsSn,
			ProductId:                 item.ProductId,
			GoodsName:                 item.GoodsName,
			ListPicUrl:                item.ListPicUrl,
			MarketPrice:               item.MarketPrice,
			RetailPrice:               item.RetailPrice,
			Number:                    item.Number,
			GoodsSpecifitionNameValue: item.GoodsSpecifitionNameValue,
			GoodsSpecifitionIds:       item.GoodsSpecifitionIds,
		}
		o.Insert(&ordergood)
	}
	models.ClearBuyGoods(userId)

	utils.ReturnHTTPSuccess(&this.Controller, orderinfo)
	this.ServeJSON()

}

func (this *OrderController) Order_Express() {
	orderId := this.GetString("orderId")
	intorderId := utils.String2Int(orderId)

	if orderId == "" {
		this.Abort("订单不存在")
	}

	latestexpressinfo := models.GetLatestOrderExpress(intorderId)

	utils.ReturnHTTPSuccess(&this.Controller, latestexpressinfo)
	this.ServeJSON()
}

func (this *OrderController) Order_UpdateStatus(){
	userId, _ := getUserIdFromJwt(this.Ctx)
	var uos UpdateOrderStatusBody
	body := this.Ctx.Input.RequestBody
	json.Unmarshal(body, &uos)
	o := orm.NewOrm()
	ordertable := new(models.NideshopOrder)
	var order models.NideshopOrder
	err := o.QueryTable(ordertable).Filter("id", uos.OrderId).Filter("user_id", userId).One(&order)
	if err != nil {
		if err == orm.ErrNoRows {
			this.Abort("订单不存在")
		}else {
			this.Abort("查询订单失败")
		}
	}
	_, err = o.QueryTable(ordertable).Filter("id", uos.OrderId).Filter("user_id", userId).Update(orm.Params{
		"order_status": uos.Status,
	})
	if err != nil {
		this.Abort("更改订单状态失败")
	}
	utils.ReturnHTTPSuccess(&this.Controller, nil)
	this.ServeJSON()
}
