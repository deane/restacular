package restacular

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

// Idea taken from Alice (https://github.com/justinas/alice)
type Chain struct {
	middlewares []Middleware
}

func GoThrough(middlewares ...Middleware) Chain {
	chain := Chain{middlewares}
	return chain
}

func (chain Chain) Then(handler HandlerFunc) HandlerFunc {
	final := handler

	// We execute middlewares in the reverse order of the array
	for i := len(chain.middlewares) - 1; i >= 0; i-- {
		if f, ok := chain.middlewares[i](final).(HandlerFunc); ok {
			final = f
		} else {
          panic("Wrong response writer in Middleware")
        }
	}

	return final
}
