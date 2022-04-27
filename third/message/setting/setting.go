package setting

var (
	// MessageAddress : message address
	MessageAddress = "https://message.ifengidc.com"
	// MessageJWTURL : message jwt url
	MessageJWTURL = MessageAddress + "/api/v1/jwt"
	// MessageLimitURL : message limit url
	MessageLimitURL = MessageAddress + "/api/v1/limit"
	// MessagePostWechatURL : message post wechat url
	MessagePostWechatURL = MessageAddress + "/api/v1/wechat"
	// MessagePostMailURL : message post mail url
	MessagePostMailURL = MessageAddress + "/api/v1/mail"
	// MessagePostSMSURL : message post sms url
	MessagePostSMSURL = MessageAddress + "/api/v1/sms"
	// MessagePostIVRURL : message post ivr url
	MessagePostIVRURL = MessageAddress + "/api/v1/ivr?mobile=%s&ttscode=%s"
	// MessageGetIVRQueryURL : message get ivr query call detail url
	MessageGetIVRQueryURL = MessageAddress + "/api/v1/ivr?callid=%s"
)

// ModelDebug : if true, change setting value
func ModelDebug(debug bool, debugAddress string) {
	if debug {
		MessageAddress = "http://localhost:9990"
		if debugAddress != "" {
			MessageAddress = debugAddress
		}
		MessageJWTURL = MessageAddress + "/api/v1/jwt"
		MessageLimitURL = MessageAddress + "/api/v1/limit"
		MessagePostWechatURL = MessageAddress + "/api/v1/wechat"
		MessagePostMailURL = MessageAddress + "/api/v1/mail"
		MessagePostSMSURL = MessageAddress + "/api/v1/sms"
		MessagePostIVRURL = MessageAddress + "/api/v1/ivr?mobile=%s&ttscode=%s"
		MessageGetIVRQueryURL = MessageAddress + "/api/v1/ivr?callid=%s"
	}
}
