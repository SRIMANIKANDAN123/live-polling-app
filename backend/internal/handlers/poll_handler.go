package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"live-polling-app/backend/internal/services"
	"live-polling-app/backend/internal/utils"
)

type PollHandler struct {
	polls *services.PollService
	votes *services.VoteService
}

func NewPollHandler(polls *services.PollService, votes *services.VoteService) *PollHandler {
	return &PollHandler{polls: polls, votes: votes}
}

type createPollRequest struct {
	Question        string     `json:"question" binding:"required"`
	Description     string     `json:"description"`
	Options         []string   `json:"options" binding:"required"`
	ExpiresAt       *time.Time `json:"expiresAt"`
	AnonymousVoting bool       `json:"anonymousVoting"`
}

func (h *PollHandler) Create(c *gin.Context) {
	var req createPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "question and at least 2 options are required")
		return
	}
	creatorID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
	if err != nil {
		utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid session")
		return
	}

	poll, err := h.polls.Create(c.Request.Context(), creatorID, services.CreatePollInput{
		Question:        req.Question,
		Description:     req.Description,
		Options:         req.Options,
		ExpiresAt:       req.ExpiresAt,
		AnonymousVoting: req.AnonymousVoting,
	})
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	utils.Success(c, http.StatusCreated, poll)
}

func (h *PollHandler) Get(c *gin.Context) {
	poll, err := h.polls.GetByPublicID(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.Fail(c, http.StatusNotFound, "POLL_NOT_FOUND", "Poll not found")
		return
	}
	utils.Success(c, http.StatusOK, poll)
}

func (h *PollHandler) Results(c *gin.Context) {
	results, err := h.votes.GetResults(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.Fail(c, http.StatusNotFound, "POLL_NOT_FOUND", "Poll not found")
		return
	}
	utils.Success(c, http.StatusOK, results)
}

func (h *PollHandler) ListMine(c *gin.Context) {
	creatorID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
	if err != nil {
		utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid session")
		return
	}
	polls, err := h.polls.ListMine(c.Request.Context(), creatorID)
	if err != nil {
		utils.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list polls")
		return
	}
	utils.Success(c, http.StatusOK, polls)
}

func (h *PollHandler) Close(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
	if err != nil {
		utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid session")
		return
	}
	if err := h.polls.Close(c.Request.Context(), c.Param("id"), ownerID); err != nil {
		status, code := mapPollError(err)
		utils.Fail(c, status, code, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"closed": true})
}

func (h *PollHandler) Delete(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
	if err != nil {
		utils.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid session")
		return
	}
	if err := h.polls.Delete(c.Request.Context(), c.Param("id"), ownerID); err != nil {
		status, code := mapPollError(err)
		utils.Fail(c, status, code, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func mapPollError(err error) (int, string) {
	switch {
	case errors.Is(err, services.ErrPollNotFound):
		return http.StatusNotFound, "POLL_NOT_FOUND"
	case errors.Is(err, services.ErrNotPollOwner):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, services.ErrPollClosed):
		return http.StatusConflict, "POLL_CLOSED"
	case errors.Is(err, services.ErrInvalidOption):
		return http.StatusBadRequest, "INVALID_OPTION"
	case errors.Is(err, services.ErrAlreadyVoted):
		return http.StatusConflict, "ALREADY_VOTED"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
