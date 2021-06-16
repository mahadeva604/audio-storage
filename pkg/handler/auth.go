package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	storage "github.com/mahadeva604/audio-storage"
	"net/http"
)

// @Summary SignUp
// @Tags auth
// @Description create account
// @ID create-account
// @Accept  json
// @Produce  json
// @Param user body storage.User true "account info"
// @Success 200 {integer} integer 1
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var user storage.User

	if err := c.BindJSON(&user); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	id, err := h.services.CreateUser(user)

	if errors.Is(err, storage.UserExists) {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type signInInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshTokensInput struct {
	RefreshTooken string `json:"refresh_token" binding:"required"`
}

// @Summary SignIn
// @Tags auth
// @Description login
// @ID login
// @Accept  json
// @Produce  json
// @Param input body signInInput true "credentials"
// @Success 200 {object} tokensResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var input signInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userId, token, err := h.services.Authorization.GenerateAccessToken(input.Username, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := h.services.Authorization.GenerateRefreshToken(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokensResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// @Summary Refresh tokens
// @Tags auth
// @Description Generate new refresh and access tokens
// @ID refresh
// @Accept  json
// @Produce  json
// @Param input body refreshTokensInput true "refresh token"
// @Success 200 {object} tokensResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/refresh [post]
func (h *Handler) refreshTokens(c *gin.Context) {
	var input refreshTokensInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userId, newRefreshToken, err := h.services.UpdateRefreshToken(input.RefreshTooken)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	newAccessToken, err := h.services.UpdateAccessToken(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"token":         newAccessToken,
		"refresh_token": newRefreshToken,
	})
}
