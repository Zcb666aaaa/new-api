package controller

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func GetTopUpInfo(c *gin.Context) {
	// 获取支付方式
	payMethods := operation_setting.PayMethods

	// 如果启用了 Stripe 支付，添加到支付方法列表
	if setting.StripeApiSecret != "" && setting.StripeWebhookSecret != "" && setting.StripePriceId != "" {
		// 检查是否已经包含 Stripe
		hasStripe := false
		for _, method := range payMethods {
			if method["type"] == "stripe" {
				hasStripe = true
				break
			}
		}

		if !hasStripe {
			stripeMethod := map[string]string{
				"name":      "Stripe",
				"type":      "stripe",
				"color":     "rgba(var(--semi-purple-5), 1)",
				"min_topup": strconv.Itoa(setting.StripeMinTopUp),
			}
			payMethods = append(payMethods, stripeMethod)
		}
	}

	data := gin.H{
		"enable_online_topup": operation_setting.PayAddress != "" && operation_setting.EpayId != "" && operation_setting.EpayKey != "",
		"enable_stripe_topup": setting.StripeApiSecret != "" && setting.StripeWebhookSecret != "" && setting.StripePriceId != "",
		"enable_creem_topup":  setting.CreemApiKey != "" && setting.CreemProducts != "[]",
		"creem_products":      setting.CreemProducts,
		"pay_methods":         payMethods,
		"min_topup":           operation_setting.MinTopUp,
		"stripe_min_topup":    setting.StripeMinTopUp,
		"amount_options":      operation_setting.GetPaymentSetting().AmountOptions,
		"discount":            operation_setting.GetPaymentSetting().AmountDiscount,
	}
	common.ApiSuccess(c, data)
}

type EpayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

type AmountRequest struct {
	Amount int64 `json:"amount"`
}

func GetEpayClient() *epay.Client {
	if operation_setting.PayAddress == "" || operation_setting.EpayId == "" || operation_setting.EpayKey == "" {
		return nil
	}
	withUrl, err := epay.NewClient(&epay.Config{
		PartnerID: operation_setting.EpayId,
		Key:       operation_setting.EpayKey,
	}, operation_setting.PayAddress)
	if err != nil {
		return nil
	}
	return withUrl
}

func getPayMoney(amount int64, group string) float64 {
	dAmount := decimal.NewFromInt(amount)
	// 充值金额以“展示类型”为准：
	// - USD/CNY: 前端传 amount 为金额单位；TOKENS: 前端传 tokens，需要换成 USD 金额
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		dAmount = dAmount.Div(dQuotaPerUnit)
	}

	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}

	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dPrice := decimal.NewFromFloat(operation_setting.Price)
	// apply optional preset discount by the original request amount (if configured), default 1.0
	discount := 1.0
	if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(amount)]; ok {
		if ds > 0 {
			discount = ds
		}
	}
	dDiscount := decimal.NewFromFloat(discount)

	payMoney := dAmount.Mul(dPrice).Mul(dTopupGroupRatio).Mul(dDiscount)

	return payMoney.InexactFloat64()
}

func getMinTopup() int64 {
	minTopup := operation_setting.MinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		minTopup = int(dMinTopup.Mul(dQuotaPerUnit).IntPart())
	}
	return int64(minTopup)
}

// epayMAPIData 易支付 mapi.php 返回的 data 字段
type epayMAPIData struct {
	TradeNo    string `json:"trade_no"`
	OutTradeNo string `json:"out_trade_no"`
	PayURL     string `json:"payurl"`
	QRCode     string `json:"qrcode"`
	URLScheme  string `json:"urlscheme"`
}

// epayMAPIResponse 易支付 mapi.php 返回结构
type epayMAPIResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data epayMAPIData `json:"data"`
}

// callEpayMAPI 调用易支付后端 mapi.php 接口，返回二维码或跳转链接
func callEpayMAPI(payAddress string, formParams map[string]string, clientIP string) (*epayMAPIResponse, error) {
	mapiURL := strings.TrimRight(payAddress, "/") + "/mapi.php"

	// clientip 必须在签名之前加入参数，否则 mapi.php 收到的参数和签名不一致导致验签失败
	// formParams 已包含 sign，需去掉旧 sign，加入 clientip 后重新签名
	if clientIP != "" {
		formParams["clientip"] = clientIP
	}
	// 去掉旧 sign/sign_type，重新用 go-epay 的签名算法计算
	delete(formParams, "sign")
	delete(formParams, "sign_type")
	// 复用 go-epay 的 GenerateParams 逻辑：过滤空值→ASCII排序→拼接→MD5加盐
	epayKey := operation_setting.EpayKey
	filtered := map[string]string{}
	for k, v := range formParams {
		if v != "" {
			filtered[k] = v
		}
	}
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	// ASCII 升序排序
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(filtered[k])
	}
	digest := md5.Sum([]byte(sb.String() + epayKey))
	formParams["sign"] = fmt.Sprintf("%x", digest)
	formParams["sign_type"] = "MD5"

	formValues := url.Values{}
	for k, v := range formParams {
		formValues.Set(k, v)
	}
	resp, err := http.PostForm(mapiURL, formValues)
	if err != nil {
		return nil, fmt.Errorf("mapi请求失败: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mapi读取响应失败: %w", err)
	}
	log.Printf("[易支付mapi] url=%s | response=%s", mapiURL, string(bodyBytes))
	var result epayMAPIResponse
	if err := common.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("mapi响应解析失败: %w", err)
	}
	if result.Code != 1 {
		return nil, fmt.Errorf("mapi返回失败: %s", result.Msg)
	}
	return &result, nil
}

func RequestEpay(c *gin.Context) {
	var req EpayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getMinTopup())})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	if !operation_setting.ContainsPayMethod(req.PaymentMethod) {
		c.JSON(200, gin.H{"message": "error", "data": "支付方式不存在"})
		return
	}

	callBackAddress := service.GetCallbackAddress()
	returnUrl, _ := url.Parse(system_setting.ServerAddress + "/console/log")
	notifyUrl, _ := url.Parse(callBackAddress + "/api/user/epay/notify")
	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", id, tradeNo)
	client := GetEpayClient()
	if client == nil {
		c.JSON(200, gin.H{"message": "error", "data": "当前管理员未配置支付信息"})
		return
	}

	// 通过 go-epay 库生成签名参数（Device 用 PC，仅借助其签名能力）
	_, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           req.PaymentMethod,
		ServiceTradeNo: tradeNo,
		Name:           fmt.Sprintf("TUC%d", req.Amount),
		Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
		Device:         epay.PC,
		NotifyUrl:      notifyUrl,
		ReturnUrl:      returnUrl,
	})
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "生成签名失败"})
		return
	}

	// 调用 mapi.php 后端接口，获取二维码或支付跳转链接
	clientIP := c.ClientIP()
	mapiResult, err := callEpayMAPI(operation_setting.PayAddress, params, clientIP)
	if err != nil {
		log.Printf("[易支付mapi] 失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}
	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: req.PaymentMethod,
		CreateTime:    time.Now().Unix(),
		Status:        "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	// 返回给前端：二维码链接 或 跳转链接，以及订单号（用于轮询）
	c.JSON(200, gin.H{
		"message":   "success",
		"trade_no":  tradeNo,
		"qrcode":    mapiResult.Data.QRCode,
		"payurl":    mapiResult.Data.PayURL,
		"urlscheme": mapiResult.Data.URLScheme,
	})
}

// QueryEpayOrder 前端轮询订单支付状态
func QueryEpayOrder(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.JSON(200, gin.H{"message": "error", "data": "缺少订单号"})
		return
	}

	// 验证订单归属当前用户
	userId := c.GetInt("id")
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil || topUp.UserId != userId {
		c.JSON(200, gin.H{"message": "error", "data": "订单不存在"})
		return
	}

	// 如果已经完成，直接返回
	if topUp.Status == "success" {
		c.JSON(200, gin.H{"message": "success", "data": "success"})
		return
	}

	// 向易支付查询订单状态
	// 使用 HMAC-MD5 签名保护查询请求，防止参数被篡改
	queryParams := map[string]string{
		"act":          "order",
		"pid":          operation_setting.EpayId,
		"out_trade_no": tradeNo,
	}
	// 过滤空值 → ASCII 排序 → 拼接 → MD5 加盐签名
	queryKeys := make([]string, 0, len(queryParams))
	for k, v := range queryParams {
		if v != "" {
			queryKeys = append(queryKeys, k)
		}
	}
	sort.Strings(queryKeys)
	var querySb strings.Builder
	for i, k := range queryKeys {
		if i > 0 {
			querySb.WriteByte('&')
		}
		querySb.WriteString(k)
		querySb.WriteByte('=')
		querySb.WriteString(queryParams[k])
	}
	queryDigest := md5.Sum([]byte(querySb.String() + operation_setting.EpayKey))
	querySign := fmt.Sprintf("%x", queryDigest)

	queryURL := fmt.Sprintf("%s/api.php?act=order&pid=%s&out_trade_no=%s&sign=%s&sign_type=MD5",
		strings.TrimRight(operation_setting.PayAddress, "/"),
		url.QueryEscape(operation_setting.EpayId),
		url.QueryEscape(tradeNo),
		url.QueryEscape(querySign),
	)
	resp, err := http.Get(queryURL) //nolint:gosec // URL 已经过签名保护
	if err != nil {
		log.Printf("[易支付查询] 请求失败: tradeNo=%s, url=%s, err=%v", tradeNo, queryURL, err)
		c.JSON(200, gin.H{"message": "error", "data": "查询失败"})
		return
	}
	defer resp.Body.Close()
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("[易支付查询] 读取响应失败: tradeNo=%s, err=%v", tradeNo, readErr)
		c.JSON(200, gin.H{"message": "error", "data": "读取响应失败"})
		return
	}
	log.Printf("[易支付查询] tradeNo=%s, statusCode=%d, body=%s", tradeNo, resp.StatusCode, string(bodyBytes))

	var result struct {
		Code   int    `json:"code"`
		Status int    `json:"status"`
		Money  string `json:"money"` // 易支付返回实付金额，用于安全验证
		Msg    string `json:"msg"`
	}
	if err := common.Unmarshal(bodyBytes, &result); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "响应解析失败"})
		return
	}

	if result.Code == 1 && result.Status == 1 {
		// 漏洞修复：验证实付金额与订单金额一致，防止低价充高额度
		paidMoney, parseErr := strconv.ParseFloat(result.Money, 64)
		if parseErr != nil {
			log.Printf("易支付轮询金额解析失败: %v, tradeNo=%s, body=%s", parseErr, tradeNo, string(bodyBytes))
			c.JSON(200, gin.H{"message": "error", "data": "支付金额解析失败"})
			return
		}
		dPaidMoney := decimal.NewFromFloat(paidMoney)
		dOrderMoney := decimal.NewFromFloat(topUp.Money)
		if !dPaidMoney.Equal(dOrderMoney) {
			log.Printf("易支付轮询金额不匹配: 实付=%f, 订单金额=%f, 订单号=%s", paidMoney, topUp.Money, tradeNo)
			c.JSON(200, gin.H{"message": "error", "data": "支付金额异常"})
			return
		}

		// 支付成功，触发入账（幂等，和 EpayNotify 逻辑一致）
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)
		// 重新读取，防止并发
		topUp = model.GetTopUpByTradeNo(tradeNo)
		if topUp != nil && topUp.Status == "pending" {
			topUp.Status = "success"
			topUp.CompleteTime = time.Now().Unix()
			_ = topUp.Update()
			dAmount := decimal.NewFromInt(int64(topUp.Amount))
			dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
			quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
			_ = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money))
		}
		c.JSON(200, gin.H{"message": "success", "data": "success"})
		return
	}

	c.JSON(200, gin.H{"message": "success", "data": "pending"})
}

// tradeNo lock
var orderLocks sync.Map
var createLock sync.Mutex

// LockOrder 尝试对给定订单号加锁
func LockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if !ok {
		createLock.Lock()
		defer createLock.Unlock()
		lock, ok = orderLocks.Load(tradeNo)
		if !ok {
			lock = new(sync.Mutex)
			orderLocks.Store(tradeNo, lock)
		}
	}
	lock.(*sync.Mutex).Lock()
}

// UnlockOrder 释放给定订单号的锁
func UnlockOrder(tradeNo string) {
	lock, ok := orderLocks.Load(tradeNo)
	if ok {
		lock.(*sync.Mutex).Unlock()
	}
}

func EpayNotify(c *gin.Context) {
	var params map[string]string

	if c.Request.Method == "POST" {
		// POST 请求：从 POST body 解析参数
		if err := c.Request.ParseForm(); err != nil {
			log.Println("易支付回调POST解析失败:", err)
			_, _ = c.Writer.Write([]byte("fail"))
			return
		}
		params = lo.Reduce(lo.Keys(c.Request.PostForm), func(r map[string]string, t string, i int) map[string]string {
			r[t] = c.Request.PostForm.Get(t)
			return r
		}, map[string]string{})
	} else {
		// GET 请求：从 URL Query 解析参数
		params = lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
			r[t] = c.Request.URL.Query().Get(t)
			return r
		}, map[string]string{})
	}

	if len(params) == 0 {
		log.Println("易支付回调参数为空")
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	client := GetEpayClient()
	if client == nil {
		log.Println("易支付回调失败 未找到配置信息")
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
		return
	}
	verifyInfo, err := client.Verify(params)
	if err == nil && verifyInfo.VerifyStatus {
		_, err := c.Writer.Write([]byte("success"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	} else {
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
		log.Println("易支付回调签名验证失败")
		return
	}

	if verifyInfo.TradeStatus == epay.StatusTradeSuccess {
		log.Println(verifyInfo)
		LockOrder(verifyInfo.ServiceTradeNo)
		defer UnlockOrder(verifyInfo.ServiceTradeNo)
		topUp := model.GetTopUpByTradeNo(verifyInfo.ServiceTradeNo)
		if topUp == nil {
			log.Printf("易支付回调未找到订单: %v", verifyInfo)
			return
		}
		
		// 安全验证：比对支付金额（精确匹配，不允许误差）
		paidMoney, err := strconv.ParseFloat(verifyInfo.Money, 64)
		if err != nil {
			log.Printf("易支付回调金额解析失败: %v, 订单: %v", err, verifyInfo)
			return
		}
		// 使用 decimal 进行精确比较，不允许任何误差
		dPaidMoney := decimal.NewFromFloat(paidMoney)
		dOrderMoney := decimal.NewFromFloat(topUp.Money)
		if !dPaidMoney.Equal(dOrderMoney) {
			log.Printf("易支付回调金额不匹配: 实付=%f, 订单金额=%f, 订单号=%s", paidMoney, topUp.Money, verifyInfo.ServiceTradeNo)
			return
		}
		
		if topUp.Status == "pending" {
			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf("易支付回调更新订单失败: %v", topUp)
				return
			}
			//user, _ := model.GetUserById(topUp.UserId, false)
			//user.Quota += topUp.Amount * 500000
			dAmount := decimal.NewFromInt(int64(topUp.Amount))
			dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
			quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
			err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
			if err != nil {
				log.Printf("易支付回调更新用户失败: %v", topUp)
				return
			}
			log.Printf("易支付回调更新用户成功 %v", topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money))
		}
	} else {
		log.Printf("易支付异常回调: %v", verifyInfo)
	}
}

func RequestAmount(c *gin.Context) {
	var req AmountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	if req.Amount < getMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

func GetUserTopUps(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")

	var (
		topups []*model.TopUp
		total  int64
		err    error
	)
	if keyword != "" {
		topups, total, err = model.SearchUserTopUps(userId, keyword, pageInfo)
	} else {
		topups, total, err = model.GetUserTopUps(userId, pageInfo)
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(topups)
	common.ApiSuccess(c, pageInfo)
}

// GetAllTopUps 管理员获取全平台充值记录
func GetAllTopUps(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")

	var (
		topups []*model.TopUp
		total  int64
		err    error
	)
	if keyword != "" {
		topups, total, err = model.SearchAllTopUps(keyword, pageInfo)
	} else {
		topups, total, err = model.GetAllTopUps(pageInfo)
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(topups)
	common.ApiSuccess(c, pageInfo)
}

type AdminCompleteTopupRequest struct {
	TradeNo string `json:"trade_no"`
}

// AdminCompleteTopUp 管理员补单接口
func AdminCompleteTopUp(c *gin.Context) {
	var req AdminCompleteTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	// 订单级互斥，防止并发补单
	LockOrder(req.TradeNo)
	defer UnlockOrder(req.TradeNo)

	if err := model.ManualCompleteTopUp(req.TradeNo); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}
