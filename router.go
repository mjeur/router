package router

import (
	"log"
	"net/http"
)

type Router struct {
	mux http.ServeMux
	log *log.Logger
}

func New(log *log.Logger) *Router {
	r := &Router{
		mux: *http.NewServeMux(),
		log: log,
	}
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func Handle[S any, U any](r *Router, pattern string, method string, handlers ...HandlerFunc[S, U]) {
	r.mux.HandleFunc(method+" "+pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := CreateContext[S, U](w, req)
		ctx.Log = r.log
		defer ctx.cancelFunc()

		//ctxHandlers := make([]HandlerFunc[S, U], 0)
		ctx.handlers = handlers
		ctx.index = -1

		var reqData S

		if err := ctx.Decode(req, &reqData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respData, err := ctx.Next(&reqData)
		if err != nil {
			if err.Error() == "unauthorized" {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if respData != nil {
			if err := JSON(w, respData); err != nil {
				r.log.Println("Error writing response:", err)
			}
		} else {
			http.NotFound(w, req)
		}
	})
}

func POST[S any, U any](r *Router, pattern string, handlers ...HandlerFunc[S, U]) {
	Handle(r, pattern, http.MethodPost, handlers...)
}

func GET[S any, U any](r *Router, pattern string, handlers ...HandlerFunc[S, U]) {
	Handle(r, pattern, http.MethodGet, handlers...)
}

type Group struct {
	prefix   string
	router   *Router
	handlers []HandlerFunc[any, any]
}

func (r *Router) Group(prefix string, handlers ...HandlerFunc[any, any]) *Group {
	return &Group{
		prefix:   prefix,
		router:   r,
		handlers: handlers,
	}
}
