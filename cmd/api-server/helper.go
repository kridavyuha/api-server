package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	ErrCouldNotParseBody = errors.New("could not parse request body")
	ErrCouldNotReadBody  = errors.New("could not read request body")
)

type httpResp struct {
	Status  int         `json:"status"`
	IsError bool        `json:"is_error"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func getBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ErrCouldNotReadBody
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return ErrCouldNotParseBody
	}
	return nil
}

func sendResponse(rw http.ResponseWriter, resp httpResp) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(resp.Status)
	out, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"status": 500, "is_error": true, "error": "could not marshal response"}`))
		return
	}
	rw.Write(out)
}
