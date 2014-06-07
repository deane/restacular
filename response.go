package restacular

import (
	"encoding/json"
	"net/http"
)

const MIME_JSON = "application/json"

type ResponseWriter interface {
	http.ResponseWriter
	WriteJSON(int, interface{}) error
	WriteError(int, error)
}

type Response struct {
	http.ResponseWriter
	prettyPrint bool
}

func NewResponse(res http.ResponseWriter) *Response {
	//TODO(Dean): determine prettyPrint from config
	return &Response{res, true}
}

func (response *Response) WriteJSON(statusCode int, data interface{}) error {
	var (
		out []byte
		err error
	)

	if data == nil {
		// TODO(Dean): maybe make the nil response configurable?
		return nil
	}

	if response.prettyPrint {
		out, err = json.MarshalIndent(data, "", "  ")
	} else {
		out, err = json.Marshal(data)
	}
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}

	response.WriteHeader(statusCode)
	_, err = response.Write(out)
	return err
}

func (response *Response) WriteError(statusCode int, err error) {
	var out []byte

	errorMap := map[string]string{"error": err.Error()}
	if response.prettyPrint {
		out, _ = json.MarshalIndent(errorMap, "", "  ")
	} else {
		out, _ = json.Marshal(errorMap)
	}
	response.WriteHeader(statusCode)
	response.Write(out)
}
