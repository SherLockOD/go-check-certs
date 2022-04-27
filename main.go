package main

import (
	"git.ifengidc.com/likuo/go-check-certs/httpd"
	"git.ifengidc.com/likuo/go-check-certs/third/message"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

/*
基础功能:
1. 证书校验
2. 定时校验
3. 数据入库
4. 报警通知

扩展用户功能(api)：
* 支持查询域名是否证书过期 （SRE小助手查询、页面查询、配置monitor做简易的报警通知）
* 支持定时校验域名证书有效期 （定时查询时间暂时：每天 9点|通知也每天一次）
* 支持域名信息(域名、端口、过期通知时间(不给就用默认值30天)、webhook地址)的增删改查 | 考虑通过SRE小助手来操作
	查：
		证书 www.ifeng.com  // 查看域名证书有效期
		证书 likuo			// 查看域账号下域名列表，及对应证书有效期
	增：
		证书
		www.ifeng.com
		ucms.ifeng.com
	删:
		证书-
		www.ifeng.com
		ucms.ifeng.com

* ~~支持修改计划任务执行时间~~
* 支持添加通知用户（考虑用户是否与域名绑定，还是集体通知？还是用webhook？）
    * 消息通知：通知到个人，及时直观；但要确定用户名称，需要考虑用户与域名绑定。SRE 小助手可以只读uid，然后进行绑定。
    * ~~WEBHOOK：通知到群聊，实现简单；但需要用户自己创建群组和webhook地址，通知没法确定到人(除非自己单独的webhook)，需要自行查看。~~
* 接口鉴权
* ~~用户权限（只能管理员加，还是用户都可以加？）~~ 通过 sre 助手添加域名信息，无需特殊权限。
*/

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	service, err := httpd.New(":8888")
	if err != nil {
		panic(err)
	}
	err = service.Start()
	if err != nil {
		panic(err)
	}
	defer service.Close()

	go message.Init()
	go httpd.Init()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGHUP)
	<-terminate
}
