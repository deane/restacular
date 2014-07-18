package restacular

import (
	"encoding/json"
	"net/http"
	"testing"
)

type User struct {
	Name     string
	Location string
}

func userView(resp ResponseWriter, req *http.Request) {
	resp.Respond(200, &User{"Bob Marley", "ISS"})
}

func errorView(resp ResponseWriter, req *http.Request) {
	resp.Error(401, "Unauthorized")
}

func TestGettingJSONResponse(t *testing.T) {
	router := NewRouter("https://www.glass.com/")
	router.GET("/timeline", userView)

	w := doRequest("GET", "/timeline", router, t)
	if w.Code != 200 {
		t.Errorf("Received http code %d instead of 200", w.Code)
	}

	content, _ := json.Marshal(&User{"Bob Marley", "ISS"})
	body := w.Body.String()

	if body != string(content) {
		t.Errorf("JSON gotten from the view was different from the expected one: %s - %s", body, content)
	}
}

func TestGettingJSONError(t *testing.T) {
	router := NewRouter("https://www.glass.com/")
	router.GET("/timeline", errorView)

	w := doRequest("GET", "/timeline", router, t)
	if w.Code != 401 {
		t.Errorf("Received http code %d instead of 401", w.Code)
	}

	content, _ := json.Marshal(&ApiError{401, "Unauthorized"})
	body := w.Body.String()

	if body != string(content) {
		t.Errorf("JSON gotten from the view was different from the expected one: %s - %s", body, content)
	}
}
