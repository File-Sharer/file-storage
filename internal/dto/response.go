package dto

import (
	"encoding/json"
	"net/http"
)

type BasicResponse struct {
	Ok      bool   `json:"ok"`
	Details string `json:"details"`
}

type UploadResponse struct {
	Ok       bool   `json:"ok"`
	URL      string `json:"url"`
	FileSize int64  `json:"file_size"`
}

func Respond(w http.ResponseWriter, code int, resp interface{}) {
	body, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(body)
}
