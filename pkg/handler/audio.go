package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"net/http"

	"strconv"
)

const (
	MaxUploadSize = 10 << 20
)

// @Summary Get audio list
// @Security ApiKeyAuth
// @Tags audio
// @Description get audio list
// @ID get-all-audio
// @Accept  json
// @Produce  json
// @Param offset query integer true "offset" minimum(0)
// @Param limit query integer true "limit"  minimum(1)
// @Param order_type query string true "order type" Enums(owner,alphabet)
// @Success 200 {object} storage.AudioListJson
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/audio/ [get]
func (h *Handler) getAllAudio(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var input storage.AudioListParam
	if err := c.BindQuery(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid query")
		return
	}

	result, err := h.services.GetAudioList(userId, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Upload AAC file
// @Security ApiKeyAuth
// @Tags audio
// @Description upload aac file
// @ID upload-file
// @Accept multipart/form-data
// @Produce  json
// @Param file formData file true "Body with aac file"
// @Success 200 {object} idResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/audio/ [post]
func (h *Handler) uploadAudio(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	fileId := uuid.New()

	err = h.services.StoreFile(fileId, file)

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	audioId, err := h.services.UploadFile(userId, fileId.String())
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, idResponse{
		ID: audioId,
	})
}

// @Summary Add description to AAC file
// @Security ApiKeyAuth
// @Tags audio
// @Description add description
// @ID add-description
// @Accept  json
// @Produce  json
// @Param input body storage.UpdateAudio true "aac description"
// @Param id path int true "audio id"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/audio/{id} [put]
func (h *Handler) addDescription(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	audioId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid audio id param")
		return
	}

	var input storage.UpdateAudio
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := input.Validate(); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.AddDescription(userId, audioId, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

// @Summary Download AAC file
// @Security ApiKeyAuth
// @Tags audio
// @Description download aac file
// @ID download-file
// @Accept  json
// @Produce  application/octet-stream
// @Param id path int true "audio id"
// @Success 200 "Success Download"
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/audio/{id} [get]
func (h *Handler) downloadAudio(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	audioId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid audio id param")
		return
	}

	audio, err := h.services.DownloadFile(userId, audioId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	fileId, err := uuid.Parse(audio.FilePath)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	file, fileSize, err := h.services.GetFile(fileId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	c.DataFromReader(http.StatusOK, fileSize, "application/octet-stream", file, map[string]string{"Content-Disposition": audio.Title + storage.FileExt})
}
