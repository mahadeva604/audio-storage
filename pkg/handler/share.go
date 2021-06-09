package handler

import (
	"github.com/gin-gonic/gin"
	storage "github.com/mahadeva604/audio-storage"
	"net/http"
	"strconv"
)

// @Summary Share AAC file
// @Security ApiKeyAuth
// @Tags share
// @Description share aac file
// @ID share-file
// @Accept json
// @Produce json
// @Param input body storage.ShareInput true "share to"
// @Param id path int true "audio id"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/share/{id} [post]
func (h *Handler) shareAudio(c *gin.Context) {
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

	var input storage.ShareInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	err = h.services.ShareAudio(userId, audioId, input.ShareTo)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

// @Summary Unshare AAC file
// @Security ApiKeyAuth
// @Tags share
// @Description unshare aac file
// @ID unshare-file
// @Accept json
// @Produce json
// @Param input body storage.ShareInput true "unshare from"
// @Param id path int true "audio id"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/share/{id} [delete]
func (h *Handler) unshareAudio(c *gin.Context) {
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

	var input storage.ShareInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	err = h.services.UnshareAudio(userId, audioId, input.ShareTo)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

// @Summary Get share list
// @Security ApiKeyAuth
// @Tags share
// @Description get share list
// @ID get-share-list
// @Accept  json
// @Produce  json
// @Param offset query integer true "offset" minimum(0)
// @Param limit query integer true "limit"  minimum(1)
// @Success 200 {object} storage.ShareListJson
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/shares [get]
func (h *Handler) getSharedAudio(c *gin.Context) {
	var input storage.ShareListParam

	if err := c.BindQuery(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	result, err := h.services.GetSharedList(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
