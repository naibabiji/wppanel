package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/naibabiji/wp-panel/database"
	"github.com/naibabiji/wp-panel/executor"
	"github.com/naibabiji/wp-panel/models"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct{}

func (h *BackupHandler) List(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()
	rows, err := db.Query(`SELECT id, site_id, filename, file_size, db_name, auto, transport_status, transport_message, created_at
		FROM db_backups WHERE site_id = ? ORDER BY created_at DESC`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("查询失败"))
		return
	}
	defer rows.Close()

	var backups []models.DBBackup
	for rows.Next() {
		var b models.DBBackup
		var auto int
		if rows.Scan(&b.ID, &b.SiteID, &b.Filename, &b.FileSize, &b.DBName, &auto,
			&b.TransportStatus, &b.TransportMessage, &b.CreatedAt) != nil {
			continue
		}
		b.Auto = auto == 1
		backups = append(backups, b)
	}
	if backups == nil {
		backups = []models.DBBackup{}
	}
	c.JSON(http.StatusOK, models.SuccessResponse(backups))
}

func (h *BackupHandler) Create(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	site := getWebsiteByID(id)
	if site == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("网站不存在"))
		return
	}
	payload := &executor.CreateBackupPayload{Site: site, Auto: false}
	task := executor.GlobalQueue.Enqueue(executor.TaskCreateBackup, payload)
	result := <-task.ResultCh
	if result.Success {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": result.Message}))
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(result.Message))
	}
}

func (h *BackupHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	bid, _ := strconv.Atoi(c.Param("bid"))

	db := database.GetDB()
	var filename string
	err := db.QueryRow("SELECT filename FROM db_backups WHERE id = ? AND site_id = ?", bid, id).Scan(&filename)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("备份记录不存在"))
		return
	}
	executor.ExecuteDeleteBackup(id, filename)
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": "已删除"}))
}

func (h *BackupHandler) Download(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	bid, _ := strconv.Atoi(c.Param("bid"))

	site := getWebsiteByID(id)
	if site == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("网站不存在"))
		return
	}

	db := database.GetDB()
	var filename string
	err := db.QueryRow("SELECT filename FROM db_backups WHERE id = ? AND site_id = ?", bid, id).Scan(&filename)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("备份记录不存在"))
		return
	}

	backupDir := filepath.Join("/www/server/panel/backups", site.Domain)
	filePath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(filePath); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("备份文件不存在"))
		return
	}
	c.FileAttachment(filePath, filename)
}

func (h *BackupHandler) Restore(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	bid, _ := strconv.Atoi(c.Param("bid"))

	site := getWebsiteByID(id)
	if site == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("网站不存在"))
		return
	}

	db := database.GetDB()
	var filename string
	err := db.QueryRow("SELECT filename FROM db_backups WHERE id = ? AND site_id = ?", bid, id).Scan(&filename)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("备份记录不存在"))
		return
	}

	payload := &executor.RestoreBackupPayload{Site: site, Filename: filename}
	task := executor.GlobalQueue.Enqueue(executor.TaskRestoreBackup, payload)
	result := <-task.ResultCh
	if result.Success {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": result.Message}))
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(result.Message))
	}
}

func (h *BackupHandler) UploadRestore(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	site := getWebsiteByID(id)
	if site == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("网站不存在"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("请选择备份文件"))
		return
	}
	if filepath.Ext(file.Filename) != ".gz" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("仅支持 .sql.gz 格式"))
		return
	}

	safeName := filepath.Base(file.Filename)
	tmpPath := filepath.Join("/tmp", "wppanel_upload_"+safeName)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("上传失败"))
		return
	}
	defer os.Remove(tmpPath)

	payload := &executor.RestoreBackupPayload{Site: site, FilePath: tmpPath}
	task := executor.GlobalQueue.Enqueue(executor.TaskRestoreBackup, payload)
	result := <-task.ResultCh
	if result.Success {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": result.Message}))
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(result.Message))
	}
}

func (h *BackupHandler) GetSettings(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()
	var enabled, keepCount int
	err := db.QueryRow("SELECT enabled, keep_count FROM backup_settings WHERE site_id = ?", id).Scan(&enabled, &keepCount)
	if err != nil {
		c.JSON(http.StatusOK, models.SuccessResponse(models.BackupSettings{Enabled: false, KeepCount: 7}))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(models.BackupSettings{Enabled: enabled == 1, KeepCount: keepCount}))
}

func (h *BackupHandler) UpdateSettings(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.BackupSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误"))
		return
	}
	if req.KeepCount < 1 {
		req.KeepCount = 1
	}
	if req.KeepCount > 30 {
		req.KeepCount = 30
	}
	enabledVal := 0
	if req.Enabled {
		enabledVal = 1
	}
	db := database.GetDB()
	db.Exec(`INSERT INTO backup_settings (site_id, enabled, keep_count) VALUES (?, ?, ?)
		ON CONFLICT(site_id) DO UPDATE SET enabled = ?, keep_count = ?`,
		id, enabledVal, req.KeepCount, enabledVal, req.KeepCount)
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": "设置已保存"}))
}
