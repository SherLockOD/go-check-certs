package httpd

import (
	"git.ifengidc.com/likuo/go-check-certs/config"
	"go.uber.org/zap"
	"net/http"
)

func (s *Service) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			config.Logger.Error("http ParseFrom error", zap.Error(err))
			http.Error(w, http.StatusText(400), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}
