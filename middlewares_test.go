package restacular

import (
	"net/http"
	"testing"
)

func createTaggedMiddleware(tag string) Middleware {
	return func(h http.Handler) http.Handler {
		return HandlerFunc(func(w ResponseWriter, r *http.Request) {
			w.Write([]byte(tag))
			h.ServeHTTP(w, r)
		})
	}
}

func TestMiddlewareRightOrder(t *testing.T) {
	router := NewRouter("https://www.glass.com/")
	m1 := createTaggedMiddleware("m1\n")
	m2 := createTaggedMiddleware("m2\n")
	m3 := createTaggedMiddleware("m3\n")

	chain := GoThrough(m1, m2, m3).Then(fakeView)
	router.GET("/timeline", chain)

	w := doRequest("GET", "/timeline", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}
}
