package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"

	"live-polling-app/backend/internal/services"
	"live-polling-app/backend/internal/utils"
	appws "live-polling-app/backend/internal/websocket"
)

type VoteHandler struct {
	votes *services.VoteService
	hub   *appws.Hub
}

func NewVoteHandler(votes *services.VoteService, hub *appws.Hub) *VoteHandler {
	return &VoteHandler{votes: votes, hub: hub}
}

type castVoteRequest struct {
	OptionID string `json:"optionId" binding:"required"`
}

// Vote handles POST /api/polls/:id/vote. The voter identifier is a hash of
// either the authenticated user's id (if logged in) or their IP+User-Agent
// fingerprint (if anonymous) — this is what prevents duplicate votes
// without storing raw personal data.
func (h *VoteHandler) Vote(c *gin.Context) {
	var req castVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "optionId is required")
		return
	}

	var rawIdentifier string
	if userID := c.GetString("userID"); userID != "" {
		rawIdentifier = "user:" + userID
	} else {
		rawIdentifier = "anon:" + c.ClientIP() + ":" + c.Request.UserAgent()
	}
	voterHash := utils.HashIdentifier(rawIdentifier)

	result, err := h.votes.CastVote(c.Request.Context(), c.Param("id"), req.OptionID, voterHash)
	if err != nil {
		status, code := mapPollError(err)
		utils.Fail(c, status, code, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, result)
}

var upgrader = gorillaws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // CORS is enforced by the REST layer; sockets are read-only fan-out
}

// ServeWS handles GET /ws/polls/:id, upgrading the connection and handing
// it to the hub for that poll's room.
func (h *VoteHandler) ServeWS(c *gin.Context) {
	pollID := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	appws.ServeClient(h.hub, conn, pollID)
}
