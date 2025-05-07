package middleware

import (
	"net/http"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Chain struct {
	Middlewares []Middleware
}

func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		Middlewares: middlewares,
	}
}

func (mc *Chain) Use(handler Middleware) *Chain {
	return &Chain{
		Middlewares: append(mc.Middlewares, handler),
	}
}

func (mc *Chain) Handle(next http.HandlerFunc) http.HandlerFunc {
	handler := next

	for i := (len(mc.Middlewares) - 1); i >= 0; i-- {
		handler = mc.Middlewares[i](handler)
	}

	return handler.ServeHTTP
}
