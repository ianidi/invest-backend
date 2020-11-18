package operator

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/ianidi/exchange-server/internal/db"
)

// NewsGetByID
// @Summary
// @Description NewsGetByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-News-Get-By-ID
// @Param   id	path		int		true		"News ID"
// @Success 200 {object} News
// @Failure 400 {object} Error
// @Router /operator/news/{id} [get]
func NewsGetByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		NewsID int `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var news News

	if err := db.Get(&news, "SELECT * FROM News WHERE NewsID=$1", query.NewsID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_NEWS_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": news,
	})
}

// NewsGet
// @Summary
// @Description NewsGet
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-News-Get
// @Param   Offset			query		int		false		"Offset"
// @Param   Limit				query		int		false		"Limit"
// @Success 200 {object} News
// @Failure 400 {object} Error
// @Router /operator/news [get]
func NewsGet(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		Offset int `form:"offset"`
		Limit  int `form:"limit"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	if query.Limit == 0 {
		query.Limit = 100
	}

	var news []*News

	if err := db.Select(&news, "SELECT * FROM News OFFSET $1 LIMIT $2", query.Offset, query.Limit); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_NEWS_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	c.JSON(200, gin.H{
		"status": true,
		"result": news,
	})
}

// NewsAdd добавить новый пропуск
// @Summary добавить новый пропуск
// @Description NewsAdd
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-News-Add
// @Param   MemberID				query		int				true		"ID пользователя, которому выдается пропуск"
// @Param   CoworkingID			query		int				true		"ID коворкинга, к которому относится пропуск"
// @Param   DateStart				query		string		true		"Дата начала действия пропуска в формате YYYY-MM-DD"
// @Param   DateEnd					query		string		true		"Дата окончания действия пропуска в формате YYYY-MM-DD"
// @Param   Active					query		bool			true		"Состояние (активность)"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/news/add [post]
func NewsAdd(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		MemberID    int    `json:"MemberID" binding:"required"`
		CoworkingID int    `json:"CoworkingID" binding:"required"`
		DateStart   string `json:"DateStart" binding:"required"`
		DateEnd     string `json:"DateEnd" binding:"required"`
		Active      bool   `json:"Active"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	//Дата начала действия пропуска не менее текущей. Не более 100 лет в будущем. Формат даты YYYY-MM-DD
	if err := validation.Validate(query.DateStart,
		validation.Length(10, 10),
		validation.Date("2006-01-02").Min(time.Now()).Max(time.Now().AddDate(100, 0, 0)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	//Дата окончания действия пропуска не менее текущей. Не более 100 лет в будущем. Формат даты YYYY-MM-DD
	if err := validation.Validate(query.DateEnd,
		validation.Length(10, 10),
		validation.Date("2006-01-02").Min(time.Now()).Max(time.Now().AddDate(100, 0, 0)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	DateStart, err := time.Parse("2006-01-02", query.DateStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATE_START",
		})
		return
	}

	DateEnd, err := time.Parse("2006-01-02", query.DateEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATE_END",
		})
		return
	}

	if DateStart.After(DateEnd) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATES",
		})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO News (MemberID, CoworkingID, IssuerID, Status, DateStart, DateEnd, Active) VALUES ($1, $2, $3, $4, $5, $6, $7)", query.MemberID, query.CoworkingID, "pending", query.DateStart, query.DateEnd, query.Active)
	tx.MustExec("INSERT INTO Access (MemberID, CoworkingID, Active) VALUES ($1, $2, $3)", query.MemberID, query.CoworkingID, true)
	tx.Commit()

	var RecordID int

	// Получить ID созданной записи
	if err := db.Get(&RecordID, "SELECT NewsID FROM News WHERE MemberID=$1 AND CoworkingID=$2 AND IssuerID=$3 AND Status=$4 AND DateStart=$5 AND DateStart=$6 AND Active=$7 ORDER BY NewsID DESC", query.MemberID, query.CoworkingID, "pending", query.DateStart, query.DateEnd, query.Active); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"status":   true,
		"RecordID": RecordID,
	})
}

// NewsUpdateByID обновить информацию о пропуске по ID пропуска
// @Summary обновить информацию о пропуске по ID пропуска
// @Description NewsUpdateByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-News-Update-By-ID
// @Param   NewsID					query		int				true		"ID пропуска"
// @Param   DateStart				query		string		true		"Дата начала действия пропуска в формате YYYY-MM-DD"
// @Param   DateEnd					query		string		true		"Дата окончания действия пропуска в формате YYYY-MM-DD"
// @Param   Active					query		bool			true		"Состояние (активность)"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/news/update [post]
func NewsUpdateByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		NewsID    int    `json:"NewsID" binding:"required"`
		DateStart string `json:"DateStart" binding:"required"`
		DateEnd   string `json:"DateEnd" binding:"required"`
		Active    bool   `json:"Active"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	//Дата начала действия пропуска не менее текущей. Не более 100 лет в будущем. Формат даты YYYY-MM-DD
	if err := validation.Validate(query.DateStart,
		validation.Length(10, 10),
		validation.Date("2006-01-02").Min(time.Now()).Max(time.Now().AddDate(100, 0, 0)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	//Дата окончания действия пропуска не менее текущей. Не более 100 лет в будущем. Формат даты YYYY-MM-DD
	if err := validation.Validate(query.DateEnd,
		validation.Length(10, 10),
		validation.Date("2006-01-02").Min(time.Now()).Max(time.Now().AddDate(100, 0, 0)),
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	DateStart, err := time.Parse("2006-01-02", query.DateStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATE_START",
		})
		return
	}

	DateEnd, err := time.Parse("2006-01-02", query.DateEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATE_END",
		})
		return
	}

	if DateStart.After(DateEnd) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "INVALID_DATES",
		})
		return
	}

	// Существует ли пропуск, который передан в параметре NewsID
	var news News
	if err := db.Get(&news, "SELECT * FROM News WHERE NewsID=$1", query.NewsID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_NEWS_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	tx := db.MustBegin()
	tx.MustExec("UPDATE News SET DateStart=$1, DateEnd=$2, Status=$3, Active=$4 WHERE NewsID=$5", query.DateStart, query.DateEnd, "updated", query.Active, query.NewsID)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}

// NewsDeleteByID удалить пропуск по ID пропуска
// @Summary удалить пропуск по ID пропуска
// @Description NewsDeleteByID
// @Tags Operator
// @Accept  json
// @Produce  json
// @ID Operator-News-Delete-By-ID
// @Param   NewsID					query		int				true		"ID пропуска"
// @Success 200 {object} Success
// @Failure 400 {object} Error
// @Router /operator/news/delete [delete]
func NewsDeleteByID(c *gin.Context) {
	db := db.GetDB()

	var query struct {
		NewsID int `json:"NewsID" binding:"required"`
	}

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	var news News

	if err := db.Get(&news, "SELECT * FROM News WHERE NewsID=$1", query.NewsID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "NO_NEWS_RECORD",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
		}
		return
	}

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM News WHERE NewsID=$1", query.NewsID)
	tx.Commit()

	c.JSON(200, gin.H{
		"status": true,
	})
}
