package restacular

import (
	"encoding/json"
	"net/http"
)

// The private implementation of that one
type ResponseWriter struct {
	http.ResponseWriter
}

// A typical error message, maybe add some details?
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *ApiError) Error() string {
	return err.Message
}

// TODO: add logging options

func (writer *ResponseWriter) Respond(code int, obj interface{}) {
	content, err := json.Marshal(obj)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(code)
	writer.Write(content)
}

func (writer *ResponseWriter) Error(code int, message string) {
	writer.Respond(code, &ApiError{code, message})
}
