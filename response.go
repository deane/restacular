package restacular

import (
	"encoding/json"
	"net/http"
)

// Cramming the params in the response feels a bit weird
// but it allows to keep the http.Handler interface
// while staying easy to use
type Param struct {
	Name  string
	Value string
}

type Params []Param

func (params Params) Get(name string) string {
	for _, param := range params {
		if param.Name == name {
			return param.Value
		}
	}
	return ""
}

type ResponseWriter interface {
	http.ResponseWriter
	Respond(int, interface{})
	Error(int, string)
}

type responseWriter struct {
	http.ResponseWriter
	params Params
	env    map[string]interface{}
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

func (writer *responseWriter) Respond(code int, obj interface{}) {
	content, err := json.Marshal(obj)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(code)
	writer.Write(content)
}

func (writer *responseWriter) Error(code int, message string) {
	writer.Respond(code, &ApiError{code, message})
}

func newResponse(writer http.ResponseWriter, params Params) ResponseWriter {
	return &responseWriter{
		ResponseWriter: writer,
		env:            make(map[string]interface{}),
		params:         params,
	}
}
