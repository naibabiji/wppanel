package handlers

import (
	"net/http"
	"strconv"

	"github.com/naibabiji/wp-panel/database"
	"github.com/naibabiji/wp-panel/executor"
	"github.com/naibabiji/wp-panel/models"

	"github.com/gin-gonic/gin"
)

type SecurityHandler struct{}

func (h *SecurityHandler) GetSettings(c *gin.Context) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, skey, svalue, description, updated_at FROM security_settings")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败"))
		return
	}
	defer rows.Close()

	var settings []models.SecuritySetting
	for rows.Next() {
		var s models.SecuritySetting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Description, &s.UpdatedAt); err != nil {
			continue
		}
		settings = append(settings, s)
	}
	if settings == nil {
		settings = []models.SecuritySetting{}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(settings))
}

func (h *SecurityHandler) UpdateSettings(c *gin.Context) {
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误"))
		return
	}

	db := database.GetDB()

	for key, val := range raw {
		if key == "panel_title" || key == "last_whitelist_update" {
			continue
		}
		var strVal string
		switch v := val.(type) {
		case string:
			strVal = v
		case float64:
			strVal = strconv.Itoa(int(v))
		case bool:
			if v {
				strVal = "true"
			} else {
				strVal = "false"
			}
		default:
			continue
		}
		db.Exec("UPDATE security_settings SET svalue = ?, updated_at = CURRENT_TIMESTAMP WHERE skey = ?", strVal, key)
	}

	go executor.ApplyFail2banSettings()
	go applyRateLimit()

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": "安全设置已更新"}))
}

func (h *SecurityHandler) RefreshWhitelist(c *gin.Context) {
	go refreshOfficialWhitelist()

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": "白名单刷新任务已提交"}))
}

func refreshOfficialWhitelist() {
	executor.GlobalQueue.Enqueue(executor.TaskRefreshWhitelist, nil)
}

func applyRateLimit() {
	enabled, rpm, burst := executor.GetRateLimitSettings()
	executor.EnsureRateLimit(enabled, rpm, burst)
}
