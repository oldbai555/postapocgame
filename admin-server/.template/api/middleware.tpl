// Code scaffolded by goctl. Safe to edit.
// goctl {{.version}}

package middleware

import (
	"net/http"
	"postapocgame/admin-server/internal/svc"
)

type {{.name}} struct {
	svcCtx *svc.ServiceContext
}

func New{{.name}}(svcCtx *svc.ServiceContext) *{{.name}} {
	return &{{.name}}{
		svcCtx: svcCtx,
	}
}

func (m *{{.name}})Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
