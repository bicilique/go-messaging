package http

import (
	"go-messaging/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService service.AdminServiceInterface
}

type CreateAdminRequest struct {
	TelegramUserID int64  `json:"telegram_user_id" binding:"required"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
}

type UserActionRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	AdminID string `json:"admin_id" binding:"required"`
}

func NewAdminHandler(adminService service.AdminServiceInterface) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// POST /api/admin/create
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.adminService.CreateAdmin(c.Request.Context(), req.TelegramUserID, req.Username, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create admin",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin created successfully",
		"admin": gin.H{
			"telegram_user_id": req.TelegramUserID,
			"username":         req.Username,
			"first_name":       req.FirstName,
			"last_name":        req.LastName,
		},
	})
}

// GET /api/admin/users/pending
func (h *AdminHandler) GetPendingUsers(c *gin.Context) {
	users, err := h.adminService.GetPendingUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get pending users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GET /api/admin/users/approved
func (h *AdminHandler) GetApprovedUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}

	users, err := h.adminService.GetApprovedUsers(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get approved users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// POST /api/admin/users/:userID/approve
func (h *AdminHandler) ApproveUser(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	adminIDParam := c.Query("admin_id")
	if adminIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "admin_id query parameter is required",
		})
		return
	}

	adminID, err := uuid.Parse(adminIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID format",
		})
		return
	}

	err = h.adminService.ApproveUser(c.Request.Context(), userID, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to approve user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User approved successfully",
		"user_id": userID,
	})
}

// POST /api/admin/users/:userID/reject
func (h *AdminHandler) RejectUser(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	adminIDParam := c.Query("admin_id")
	if adminIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "admin_id query parameter is required",
		})
		return
	}

	adminID, err := uuid.Parse(adminIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID format",
		})
		return
	}

	err = h.adminService.RejectUser(c.Request.Context(), userID, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reject user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User rejected successfully",
		"user_id": userID,
	})
}

// POST /api/admin/users/:userID/disable
func (h *AdminHandler) DisableUser(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	adminIDParam := c.Query("admin_id")
	if adminIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "admin_id query parameter is required",
		})
		return
	}

	adminID, err := uuid.Parse(adminIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID format",
		})
		return
	}

	err = h.adminService.DisableUser(c.Request.Context(), userID, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to disable user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User disabled successfully",
		"user_id": userID,
	})
}

// POST /api/admin/users/:userID/enable
func (h *AdminHandler) EnableUser(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	adminIDParam := c.Query("admin_id")
	if adminIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "admin_id query parameter is required",
		})
		return
	}

	adminID, err := uuid.Parse(adminIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid admin ID format",
		})
		return
	}

	err = h.adminService.EnableUser(c.Request.Context(), userID, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enable user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User enabled successfully",
		"user_id": userID,
	})
}

// GET /api/admin/stats
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	stats, err := h.adminService.GetUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// POST /api/admin/cleanup
func (h *AdminHandler) CleanupPendingUsers(c *gin.Context) {
	count, err := h.adminService.CleanupPendingUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup pending users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Cleanup completed successfully",
		"deleted_count": count,
	})
}
