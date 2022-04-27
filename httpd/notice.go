package httpd

import (
	"git.ifengidc.com/likuo/go-check-certs/model"
	"git.ifengidc.com/likuo/go-check-certs/third/message"
	"strings"
)

// noticeToUser : send expires info to user when the domain cert will expire by wxwork notice
func noticeToUser(cm model.CertModel, ci model.CertInfo) bool {
	go message.Wechat(
		strings.Join(cm.User, "|"),
		"HTTPS证书过期提醒",
		"检测域名: "+cm.Host+"\n主题名称: "+ci.CommonName+"\n过期时间: "+ci.NotAfter.Format("2006-01-02 15:04:05")+"\n是否CA: "+swapBoolToString(ci.IsCA),
		"https://"+cm.Host)
	return true
}
