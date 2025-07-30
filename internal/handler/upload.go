package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/File-Sharer/file-storage/internal/dto"
	"github.com/File-Sharer/file-storage/internal/model"
)

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		dto.Respond(w, http.StatusBadRequest, dto.BasicResponse{
			Ok: false,
			Details: errNoFile.Error(),
		})
		return
	}

	path := strings.TrimSpace(r.FormValue("path"))

	fileSize, url, err := h.services.Uploader.Upload(model.UploadData{
		Path: path,
		File: file,
		FileHeader: fileHeader,
	})
	if err != nil {
		dto.Respond(w, http.StatusInternalServerError, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	dto.Respond(w, http.StatusOK, dto.UploadResponse{
		Ok: true,
		URL: url,
		FileSize: fileSize,
	})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.Respond(w, http.StatusBadRequest, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	paths := []string{}
	if err := json.Unmarshal(body, &paths); err != nil {
		dto.Respond(w, http.StatusBadRequest, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	if err := h.services.Uploader.Delete(paths); err != nil {
		dto.Respond(w, http.StatusInternalServerError, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	dto.Respond(w, http.StatusOK, dto.BasicResponse{
		Ok: true,
		Details: "",
	})
}

type createFolderReq struct {
	Path string `json:"path"`
}

func (h *Handler) createFolder(w http.ResponseWriter, r *http.Request) {
	var input createFolderReq
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.Respond(w, http.StatusBadRequest, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	if err := h.services.Uploader.CreateFolder(input.Path); err != nil {
		dto.Respond(w, http.StatusInternalServerError, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	dto.Respond(w, http.StatusOK, dto.BasicResponse{
		Ok: true,
		Details: "",
	})
}

type getZippedFolderReq struct {
	Path string `json:"path"`
}

func (h *Handler) getZippedFolder(w http.ResponseWriter, r *http.Request) {
	var input getZippedFolderReq
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.Respond(w, http.StatusBadRequest, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	data, err := h.services.Uploader.GetZippedFolder(input.Path)
	if err != nil {
		dto.Respond(w, http.StatusInternalServerError, dto.BasicResponse{
			Ok: false,
			Details: err.Error(),
		})
		return
	}

	parts := strings.Split(input.Path, "/")
	w.Header().Add("Content-Type", "application/zip")
	w.Header().Add("filename", parts[len(parts)-1])
	io.Copy(w, bytes.NewReader(data))
}
