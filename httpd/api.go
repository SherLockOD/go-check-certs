package httpd

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type RequestBody struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
}

type Response struct {
	Code uint        `json:"code"`
	Msg  interface{} `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

var (
	error4000Response = genResponseStr(Response{Code: 4000, Msg: "Invalid arguments"})
	error4001Response = genResponseStr(Response{Code: 4001, Msg: "get domain expire time error"})
	//error4002Response = genResponseStr(Response{Code: 4002, Msg: "get root path failed"})
	//error4003Response = genResponseStr(Response{Code: 4003, Msg: "access token decode error!"})
	//error4004Response = genResponseStr(Response{Code: 4004, Msg: "access token app id error!"})
	//error4005Response = genResponseStr(Response{Code: 4005, Msg: "access token sign error!"})
	//error4006Response = genResponseStr(Response{Code: 4006, Msg: "access token expire!"})
	//error4007Response = genResponseStr(Response{Code: 4007, Msg: "access token empty error!"})
	//error4008Response = genResponseStr(Response{Code: 4008, Msg: "access token element less!"})
	//error4009Response = genResponseStr(Response{Code: 4009, Msg: "未查到ID对应的实例"})

	error5000Response = genResponseStr(Response{Code: 5000, Msg: "Database error"})
	error5001Response = genResponseStr(Response{Code: 5001, Msg: "Duplicate key"})
	error5002Response = genResponseStr(Response{Code: 5002, Msg: "Host not found"})
)

func (s *Service) Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, _ = w.Write([]byte("Welcome to Certificate check Tools!"))
	return
}

func genResponseStr(data interface{}) []byte {
	result, _ := json.Marshal(data)
	return result
}
