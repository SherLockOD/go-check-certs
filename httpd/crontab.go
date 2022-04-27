package httpd

import (
	"git.ifengidc.com/likuo/go-check-certs/config"
	"git.ifengidc.com/likuo/go-check-certs/model"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	concurrencyNum  = 8
	noticeTimeHours = 10
)

func Init() {
	cron()
}

func cron() {
	/*
			1. 从库中获取待检查信息 (库操作)
		    2. 判断是否过期
		    3. 根据是否过期，判断是否通知
		    4. 休眠，定时循环
	*/
	go func() {
		for {
			checkCertExpireTimeToDB()
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		for {
			// notice time : 10:00 - 11:00
			time.Sleep(time.Minute)
			if checkNoticeTime() {
				checkCertExpireTimeFromDB()
			}
			time.Sleep(time.Hour)
		}
	}()
}

func checkNoticeTime() bool {
	now := time.Now()
	noticeTime := time.Date(now.Year(), now.Month(), now.Day(), noticeTimeHours, 0, 0, 0, now.Location())
	diffTime := now.Sub(noticeTime).Hours()

	// now - 10
	if diffTime >= 0 && diffTime < 1 {
		return true
	}
	return false
}

func swapBoolToString(b bool) string {
	switch b {
	case true:
		return "是"
	case false:
		return "否"
	}
	return "否"
}

// checkCertExpireTimeFromDB : run crontab for checking domain cert expire time
func checkCertExpireTimeFromDB() {
	certModelList, exists, err := model.GetCertInfoListAll()
	if err != nil {
		config.Logger.Error("func model.GetCertInfoListAll err", zap.String("uid", "cron"), zap.Error(err))
		return
	}

	if !exists {
		config.Logger.Error("func model.GetCertInfoListAll err, cert host list not found", zap.String("uid", "cron"))
		return
	}

	for _, certModel := range certModelList {
		if len(certModel.Cert) == 0 {
			continue
		}
		for _, c := range certModel.Cert {
			var expireTime int64
			// CA 提前5个月提醒，企业证书提前1个月提醒
			switch c.IsCA {
			case true:
				expireTime = 5 * 30 * 24
			case false:
				expireTime = 30 * 24
			}
			if c.ExpireHours <= expireTime {
				noticeToUser(certModel, c)
			}
		}
	}
	config.Logger.Info("crontab func checkCertExpireTimeFromDB success", zap.String("uid", "cron"))
}

// checkCertExpireTimeToDB : run crontab for checking domain cert expire time
func checkCertExpireTimeToDB() {

	// 完成通知 channel
	doneChan := make(chan struct{})
	defer close(doneChan)

	// get host channel
	hostChan := getHostsFromDB(doneChan)

	// make resultChan
	resultChan := make(chan HostResult)

	// 定义并发数
	var wg sync.WaitGroup

	wg.Add(concurrencyNum)
	for i := 0; i < concurrencyNum; i++ {
		go func() {
			getCertInfoToResultChan(doneChan, hostChan, resultChan)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		if r.err != nil {
			config.Logger.Error("func checkCertExpireTime err", zap.String("uid", "cron"), zap.String("host", r.Host), zap.Error(r.err))
			continue
		}
		certModel, exists, err := model.GetCertInfoByHost(r.Host)
		if err != nil {
			config.Logger.Error("func GetCertInfoByHost err", zap.String("uid", "cron"), zap.String("host", r.Host), zap.Error(err))
			continue
		}

		if !exists {
			config.Logger.Error("func GetCertInfoByHost err, host not found", zap.String("uid", "cron"), zap.String("host", r.Host), zap.String("err", "host not found"))
			continue
		}

		certModel.Cert = r.Certs
		ok, err := model.UpdateCertInfo(certModel)
		if err != nil {
			config.Logger.Error("func UpdateCertInfo err", zap.String("uid", "cron"), zap.String("host", r.Host), zap.Error(err))
			continue
		}
		if !ok {
			config.Logger.Error("func UpdateCertInfo err, host not found", zap.String("uid", "cron"), zap.String("host", r.Host), zap.String("err", "host not found"))
			continue
		}
	}
	config.Logger.Info("crontab func checkCertExpireTime success", zap.String("uid", "cron"))
}

func getCertInfoToResultChan(done <-chan struct{}, hostChan <-chan string, resultChan chan<- HostResult) {
	for host := range hostChan {
		select {
		case resultChan <- GetDomainCertInfo(host):
		case <-done:
			return
		}
	}
}

func getHostsFromDB(done <-chan struct{}) <-chan string {
	hosts := make(chan string)
	go func() {
		defer close(hosts)
		certModelList, exists, err := model.GetCertInfoListAll()
		if err != nil {
			config.Logger.Error("func model.GetCertInfoListAll err", zap.String("uid", "cron"), zap.Error(err))
			return
		}

		if !exists {
			config.Logger.Error("func model.GetCertInfoListAll err, cert host list not found", zap.String("uid", "cron"))
			return
		}

		for _, certModel := range certModelList {
			select {
			case hosts <- certModel.Host:
			case <-done:
				return
			}
		}
	}()
	return hosts
}
