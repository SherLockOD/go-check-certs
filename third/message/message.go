package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"git.ifengidc.com/likuo/go-check-certs/config"
	"git.ifengidc.com/likuo/go-check-certs/third/message/setting"
	"io/ioutil"
	"net/http"
	netURL "net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	client *Service
)

// Init : init
func Init() {
	client = NewMessageClient("v1", config.MessageAppID, config.MessageAppKey)
	err := client.InitConnection()
	if err != nil {
		panic(err)
	}
}

// Wechat : message wechat
func Wechat(uid, title, content, url string) {
	uuid := time.Now().UnixNano()
	config.Logger.Info("prepare to send wechat work", zap.String("uid", uid), zap.String("title", title), zap.String("content", content), zap.String("url", url), zap.Int64("uuid", uuid))
	res, err := client.PostWechat(uid, title, content, url)
	if err != nil {
		config.Logger.Error("PostWechat err", zap.Int64("uuid", uuid), zap.Error(err))
		return
	}
	if res.Code != 200 {
		config.Logger.Error("PostWechat failed", zap.Int64("uuid", uuid), zap.Any("response", res))
		fmt.Println(uid, title, content, url)
		return
	}
	config.Logger.Info("PostWechat succ", zap.Int64("uuid", uuid))
}

// Service : for message client
type Service struct {
	Version      string
	AppID        string
	AppKey       string
	AppJWTKey    string
	Debug        bool
	DebugAddress string
}

// response : base response
type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// NewMessageClient : init new message client
func NewMessageClient(version, appID, appKey string) *Service {
	service := Service{
		Version: version,
		AppID:   appID,
		AppKey:  appKey,
	}
	return &service
}

// InitConnection : init message connection
func (s *Service) InitConnection() error {
	setting.ModelDebug(s.Debug, s.DebugAddress)
	if s.Version == "v1" {
		err := s.GetJWTKey()
		// update AppJWTKey per 12 hours
		if err == nil {
			go func() {
				for {
					time.Sleep(12 * time.Hour)
					s.GetJWTKey()
				}
			}()
		}
		return err
	}

	return errors.New("init message connection error")
}

// JWTKeyResponse : message jwt key response for GetJWTKey
type JWTKeyResponse response

// GetJWTKey : get app jwt key
// check if AppID and AppKey are ok
func (s *Service) GetJWTKey() error {
	client := &http.Client{}
	request, err := http.NewRequest("GET", setting.MessageJWTURL, nil)
	if err != nil {
		return err
	}
	request.Header.Add("AppID", s.AppID)
	request.Header.Add("AppKey", s.AppKey)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	jwtKeyResponse := JWTKeyResponse{}
	err = json.Unmarshal(res, &jwtKeyResponse)
	if err != nil {
		return err
	}
	if jwtKeyResponse.Code != 200 {
		return errors.New(jwtKeyResponse.Msg)
	}
	s.AppJWTKey = jwtKeyResponse.Data
	return nil
}

// LimitResponse : message limit response for GetLimit
type LimitResponse struct {
	Code int               `json:"code"`
	Data LimitResponseData `json:"data"`
	Msg  string            `json:"msg"`
}

// LimitResponseData : for LimitResponse Data
type LimitResponseData struct {
	AppID      string `json:"appid"`
	AppIDName  string `json:"appidname"`
	Type       string `json:"type"`
	Owners     string `json:"owners"`
	RatedLimit uint   `json:"ratedlimit"`
	Used       uint   `json:"used"`
	Day        string `json:"day"`
}

// GetLimit : get app today limit
func (s *Service) GetLimit() (LimitResponse, error) {
	limitResponse := LimitResponse{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", setting.MessageLimitURL, nil)
	if err != nil {
		return limitResponse, err
	}
	request.Header.Add("AppJWTKey", s.AppJWTKey)
	response, err := client.Do(request)
	if err != nil {
		return limitResponse, err
	}
	defer response.Body.Close()
	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return limitResponse, err
	}
	err = json.Unmarshal(res, &limitResponse)
	if err != nil {
		return limitResponse, err
	}
	return limitResponse, nil
}

// PostWechatResponse : post wechat message response for PostWechat
type PostWechatResponse response

// PostWechat : post wechat message
func (s *Service) PostWechat(account, title, content, url string) (PostWechatResponse, error) {
	postWechatResponse := PostWechatResponse{}
	b := netURL.Values{"account": {account}, "title": {title}, "content": {content}, "url": {url}}

	res, err := s.post(setting.MessagePostWechatURL, b.Encode())
	if err != nil {
		return postWechatResponse, err
	}
	err = json.Unmarshal(res, &postWechatResponse)
	return postWechatResponse, err
}

// PostMailResponse : post mail message response for PostMail
type PostMailResponse response

// PostMail : post mail message
func (s *Service) PostMail(to, cc, subject, content string) (PostMailResponse, error) {
	postMailResponse := PostMailResponse{}
	b := netURL.Values{"to": {to}, "cc": {cc}, "subject": {subject}, "content": {content}}

	res, err := s.post(setting.MessagePostMailURL, b.Encode())
	if err != nil {
		return postMailResponse, err
	}
	err = json.Unmarshal(res, &postMailResponse)
	return postMailResponse, err
}

// PostSMSResponse : post sms message response for PostSMS
type PostSMSResponse response

// PostSMS : post sms message
func (s *Service) PostSMS(mobile, content string) (PostSMSResponse, error) {
	postSMSResponse := PostSMSResponse{}
	b := netURL.Values{"mobile": {mobile}, "content": {content}}

	res, err := s.post(setting.MessagePostSMSURL, b.Encode())
	if err != nil {
		return postSMSResponse, err
	}
	err = json.Unmarshal(res, &postSMSResponse)
	return postSMSResponse, err
}

// PostIVRResponse : post ivr message response for PostIVR
type PostIVRResponse struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data []PostIVRResponseData `json:"data"`
}

// PostIVRResponseData : for PostIVRResponse Data
type PostIVRResponseData struct {
	Mobile  string `json:"mobile"`
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	CallID  string `json:"callid"`
}

// PostIVR : post ivr message
func (s *Service) PostIVR(mobile, ttsCode string, body []byte) (PostIVRResponse, error) {
	postIVRResponse := PostIVRResponse{}
	reader := bytes.NewReader(body)
	urlEncoded := fmt.Sprintf(setting.MessagePostIVRURL, mobile, ttsCode)
	req, err := http.NewRequest("POST", urlEncoded, reader)
	if err != nil {
		return postIVRResponse, err
	}
	req.Header.Add("AppJWTKey", s.AppJWTKey)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return postIVRResponse, err
	}
	defer res.Body.Close()
	resIVR, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return postIVRResponse, err
	}
	err = json.Unmarshal(resIVR, &postIVRResponse)
	if err != nil {
		return postIVRResponse, err
	}
	return postIVRResponse, err
}

// GetIVRQueryResponse : get ivr message query response for PostIVR
type GetIVRQueryResponse struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data GetIVRQueryResponseData `json:"data"`
}

// GetIVRQueryResponseData : for GetIVRQueryResponse Data
type GetIVRQueryResponseData struct {
	CallID           string `json:"callId"`
	StartDate        string `json:"startDate"`
	StateDesc        string `json:"stateDesc"`
	Duration         int    `json:"duration"`
	CallerShowNumber string `json:"callerShowNumber"`
	GmtCreate        string `json:"gmtCreate"`
	State            string `json:"state"`
	EndDate          string `json:"endDate"`
	Callee           string `json:"callee"`
}

// GetIVRQuery : get ivr message query call detail
func (s *Service) GetIVRQuery(callID string) (GetIVRQueryResponse, error) {
	getIVRQueryResponse := GetIVRQueryResponse{}
	urlEncoded := fmt.Sprintf(setting.MessageGetIVRQueryURL, callID)
	req, err := http.NewRequest("GET", urlEncoded, nil)
	if err != nil {
		return getIVRQueryResponse, err
	}
	req.Header.Add("AppJWTKey", s.AppJWTKey)
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return getIVRQueryResponse, err
	}
	defer res.Body.Close()
	resIVR, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return getIVRQueryResponse, err
	}
	err = json.Unmarshal(resIVR, &getIVRQueryResponse)
	if err != nil {
		return getIVRQueryResponse, err
	}
	return getIVRQueryResponse, err
}

// post : for post request
func (s *Service) post(url, data string) ([]byte, error) {
	body := strings.NewReader(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("AppJWTKey", s.AppJWTKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("post url failed: %s code: %d", url, res.StatusCode)
		return nil, err
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil
}
