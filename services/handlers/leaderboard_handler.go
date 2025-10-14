package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type LeaderboardHandler struct {
	userSvc UserServiceInterface
	jwtSvc  JWTServiceInterface
}

func NewLeaderboardHandler(userSvc UserServiceInterface, jwtSvc JWTServiceInterface) *LeaderboardHandler {
	return &LeaderboardHandler{
		userSvc: userSvc,
		jwtSvc:  jwtSvc,
	}
}

// @Summary Get Weekly Leaderboard
// @Description Get weekly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/weekly [get]
func (h *LeaderboardHandler) GetWeeklyLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := h.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := h.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := h.userSvc.GetWeeklyLeaderboard(limit, userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}

// @Summary Get Monthly Leaderboard
// @Description Get monthly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/monthly [get]
func (h *LeaderboardHandler) GetMonthlyLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := h.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := h.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := h.userSvc.GetMonthlyLeaderboard(limit, userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}

// @Summary Get All Time Leaderboard
// @Description Get all-time leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/all-time [get]
func (h *LeaderboardHandler) GetAllTimeLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := h.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := h.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := h.userSvc.GetAllTimeLeaderboard(limit, userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}
