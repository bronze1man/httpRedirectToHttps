package httpRedirectToHttps

import (
	"net/http"
)

type Handler struct {
	in http.Handler
}

func NewHandler(in http.Handler) http.Handler {
	return &Handler{in: in}
}

func (this *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.TLS == nil {
		newUrl := `https://` + req.Host + req.RequestURI
		http.Redirect(w, req, newUrl, http.StatusTemporaryRedirect)
		return
	}
	this.in.ServeHTTP(w, req)
}
