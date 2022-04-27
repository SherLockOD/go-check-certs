package httpd

import (
	"crypto/tls"
	"encoding/json"
	"git.ifengidc.com/likuo/go-check-certs/config"
	"git.ifengidc.com/likuo/go-check-certs/model"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HostResult struct {
	Host  string           `json:"host"`
	Certs []model.CertInfo `json:"certs"`
	err   error
}

// GetCertExpireTime : get domain host cert expire time for checking
func (s *Service) GetCertExpireTime(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := r.Form.Get("uid")
	host := r.Form.Get("host")
	config.Logger.Info("new get domain cert expire time request", zap.String("uid", uid), zap.String("host", host))

	result := GetDomainCertInfo(host)
	if result.err != nil {
		config.Logger.Error("func GetDomainCertInfo err", zap.String("uid", uid), zap.String("host", host), zap.Error(result.err))
		w.Write(error4001Response)
		return
	}

	w.Write(genResponseStr(Response{
		Code: 200,
		Data: result,
		Msg:  "get domain cert expire time success",
	}))

}

// CreateCertInfo : create cert info for crontab
func (s *Service) CreateCertInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := RequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		config.Logger.Error("func CreateCertInfo decode json err", zap.String("uid", req.User), zap.Error(err))
		w.Write(error4000Response)
		return
	}

	c := model.CertModel{}
	c.Host = req.Host
	c.Port = req.Port
	c.User = append(c.User, req.User)

	config.Logger.Info("new create host cert info request", zap.String("uid", req.User), zap.Any("cert struct", &c))
	ok, err := model.CreateCertInfo(c)
	if err != nil {
		config.Logger.Error("func model.InsertCertInfo err", zap.String("uid", req.User), zap.Error(err))
		w.Write(error5000Response)
		return
	}

	if !ok {
		config.Logger.Error("func model.InsertCertInfo err, duplicate key", zap.String("uid", req.User))
		w.Write(error5001Response)
		return
	}

	w.Write(genResponseStr(Response{
		Code: 200,
		Msg:  "create cert info success",
	}))
}

// UpdateCertInfo : update cert info
func (s *Service) UpdateCertInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//req := RequestBody{}
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	config.Logger.Error("func CreateCertInfo decode json err", zap.String("uid", req.User), zap.Error(err))
	//	w.Write(error4000Response)
	//	return
	//}

}

// DeleteCertInfo : delete cert info
func (s *Service) DeleteCertInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := RequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		config.Logger.Error("func CreateCertInfo decode json err", zap.String("uid", req.User), zap.Error(err))
		w.Write(error4000Response)
		return
	}

	config.Logger.Info("new delete host cert info request", zap.String("user", req.User), zap.String("host", req.Host))
	cc, exists, err := model.GetCertInfoByUser(req.User, req.Host)
	if err != nil {
		config.Logger.Error("func model.GetCertInfoByUser err", zap.String("uid", req.User), zap.Error(err))
		w.Write(error5000Response)
		return
	}

	if !exists {
		config.Logger.Error("func model.GetCertInfoByUser err, cert host not found", zap.String("uid", req.User))
		w.Write(error5002Response)
		return
	}

	// user list 只有 user 自己时
	if len(cc.User) == 1 {
		ok, err := model.DeleteCertInfo(cc)
		if err != nil {
			config.Logger.Error("func model.DeleteCertInfo err", zap.String("uid", req.User), zap.Error(err))
			w.Write(error5000Response)
			return
		}

		if !ok {
			config.Logger.Error("func model.DeleteCertInfo err, cert host not found", zap.String("uid", req.User))
			w.Write(error5002Response)
			return
		}
	} else {
		ok, err := model.DeleteUserFromCertInfo(cc, req.User)
		if err != nil {
			config.Logger.Error("func model.DeleteUserFromCertInfo err", zap.String("uid", req.User), zap.Error(err))
			w.Write(error5000Response)
			return
		}

		if !ok {
			config.Logger.Error("func model.model.DeleteUserFromCertInfo err, cert host not found", zap.String("uid", req.User))
			w.Write(error5002Response)
			return
		}
	}

	w.Write(genResponseStr(Response{Code: 200, Data: req, Msg: "delete cert info success"}))
}

// GetCertInfolist : get all cert info list
func (s *Service) GetCertInfolist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := r.Form.Get("uid")
	config.Logger.Info("new get cert info list request", zap.String("uid", uid))

	certInfoList, ok, err := model.GetCertInfoListAll()
	if err != nil {
		config.Logger.Error("func model.GetCertInfolist err", zap.String("uid", uid), zap.Error(err))
		w.Write(error5000Response)
		return
	}

	if !ok {
		config.Logger.Error("func model.GetCertInfolist err, cert host list not found", zap.String("uid", uid))
		w.Write(error5002Response)
		return
	}

	w.Write(genResponseStr(Response{Code: 200, Data: certInfoList, Msg: "get cert info list success"}))
}

// GetCertInfoByUser : get cert info list for user
func (s *Service) GetCertInfoByUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := r.Form.Get("uid")
	config.Logger.Info("new get cert info by user request", zap.String("uid", uid))

	certInfoList, ok, err := model.GetCertInfoListByUser(uid)
	if err != nil {
		config.Logger.Error("func model.GetCertInfoListByUser err", zap.String("uid", uid), zap.Error(err))
		w.Write(error5000Response)
		return
	}

	if !ok {
		config.Logger.Error("func model.GetCertInfoListByUser err, user's cert host list not found", zap.String("uid", uid))
		w.Write(error5002Response)
		return
	}

	w.Write(genResponseStr(Response{Code: 200, Data: certInfoList, Msg: "get user cert info list success"}))
}

// GetCertInfoByHost : get cert info by host
func (s *Service) GetCertInfoByHost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := r.Form.Get("uid")
	host := r.Form.Get("host")
	config.Logger.Info("new get cert info by host request", zap.String("uid", uid), zap.String("host", host))

	certInfo, ok, err := model.GetCertInfoByHost(host)
	if err != nil {
		config.Logger.Error("func model.GetCertInfoByHost err", zap.String("uid", uid), zap.String("host", host), zap.Error(err))
		w.Write(error5000Response)
		return
	}

	if !ok {
		config.Logger.Error("func model.GetCertInfoByHost err, cert host not found", zap.String("uid", uid), zap.String("host", host))
		w.Write(error5002Response)
		return
	}

	w.Write(genResponseStr(Response{Code: 200, Data: certInfo, Msg: "get cert info success"}))
}

// GetDomainCertInfo : get domain origin cert info by http request
func GetDomainCertInfo(host string) (result HostResult) {
	port := "443"
	result = HostResult{
		Host:  host,
		Certs: []model.CertInfo{},
	}
	conn, err := tls.Dial("tcp", host+":"+port, nil)
	if err != nil {
		result.err = err
		return
	}
	defer conn.Close()

	timeNow := time.Now()
	checkedCerts := make(map[string]struct{})
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}

			// 过期时间
			expiresHours := int64(cert.NotAfter.Sub(timeNow).Hours())

			result.Certs = append(result.Certs, model.CertInfo{
				CommonName:  cert.Subject.CommonName,
				ExpireHours: expiresHours,
				IsCA:        cert.IsCA,
				NotBefore:   cert.NotBefore,
				NotAfter:    cert.NotAfter,
			})
		}
	}
	return
}
