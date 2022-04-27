package httpd

import (
	"net"
	"net/http"
	"strings"
	"time"

	"git.ifengidc.com/likuo/go-check-certs/config"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type Service struct {
	addr   string
	ln     net.Listener
	router *httprouter.Router
}

func New(listen string) (*Service, error) {
	return &Service{
		addr:   listen,
		router: httprouter.New(),
	}, nil
}

func (s *Service) Start() error {
	s.initHandler()
	server := http.Server{}
	server.Handler = s.router
	server.Handler = s.auth(server.Handler)
	server.Handler = s.accessLog(cors(server.Handler))

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.ln = ln

	// server
	go func() {
		if err := server.Serve(s.ln); err != nil {
			config.Logger.Error("httpd server error", zap.Error(err))
		}
	}()

	config.Logger.Info("http service started", zap.String("listen", s.addr))
	return nil
}

func (s *Service) Close() error {
	_ = s.ln.Close()
	return nil
}

func (s *Service) initHandler() {
	s.router.GET("/receive/cert", s.Index)
	s.router.GET("/receive/cert/check", s.GetCertExpireTime)
	s.router.POST("/receive/cert/check", s.CreateCertInfo)
	// s.router.PUT("/receive/cert/check", s.UpdateCertInfo)
	s.router.DELETE("/receive/cert/check", s.DeleteCertInfo)
	s.router.GET("/receive/cert/list", s.GetCertInfolist)
	s.router.GET("/receive/cert/user/list", s.GetCertInfoByUser)
}

func (s *Service) accessLog(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			inner.ServeHTTP(w, r)
			return
		}
		stime := time.Now().UnixNano() / 1e3
		inner.ServeHTTP(w, r)
		dur := time.Now().UnixNano()/1e3 - stime
		remoteIP := r.Header.Get("RemoteClentIP")
		if dur <= 1e3 {
			config.Logger.Info("http access", zap.String("method", r.Method), zap.String("uri", r.RequestURI), zap.Int64("time(us)", dur), zap.String("RemoteClentIP", remoteIP))
		} else {
			config.Logger.Info("http access", zap.String("method", r.Method), zap.String("uri", r.RequestURI), zap.Int64("time(ms)", dur/1e3), zap.String("RemoteClentIP", remoteIP))
		}
	})
}

func cors(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set(`Access-Control-Allow-Origin`, origin)
			w.Header().Set(`Access-Control-Allow-Methods`, strings.Join([]string{
				`DELETE`,
				`GET`,
				`OPTIONS`,
				`POST`,
				`PUT`,
			}, ", "))

			w.Header().Set(`Access-Control-Allow-Headers`, strings.Join([]string{
				`Accept`,
				`Accept-Encoding`,
				`Authorization`,
				`Content-Length`,
				`Content-Type`,
				`X-CSRF-Token`,
				`X-HTTP-Method-Override`,
				`Authtoken`,
				`X-Requested-With`,
				`NS`,
				`Resource`,
			}, ", "))
		}

		if r.Method == "OPTIONS" {
			return
		}

		inner.ServeHTTP(w, r)
	})
}
