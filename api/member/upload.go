package member

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ianidi/exchange-server/graph/methods/constants"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/s3"
	"github.com/minio/minio-go/v6"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"gopkg.in/h2non/bimg.v1"
)

type File struct {
	Category string
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	RecordID int
	Position int
	File     []byte
}

func CloudUpload(filename string) bool {
	endpoint := "hb.bizmrg.com"
	accessKeyID := "pKQkDFULfrBEqRP9wXUXFT"
	secretAccessKey := "6s5AH9FmAptZTeGqzuyNMs2nvsUtc9BwfKcVdh7Q1UPY"
	useSSL := true
	bucketName := "exchange"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return false
	}

	// Upload the file
	filePath := "/var/server/static/" + filename
	contentType := "image/jpeg"

	// Upload the zip file with FPutObject
	_, err = minioClient.FPutObject(bucketName, filename, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return false
	}

	return true
}

// func Upload(c *gin.Context) {

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": true,
// 	})
// }

// Upload
// @Summary
// @Description Upload
// @Tags Operator
// @Accept  multipart/form-data
// @Produce  json
// @ID Operator-Upload
// @Param file formData file true "File"
// @Success 200 {query} string "UploadID"
// @Failure 400 {object} Error
// @Router /upload [post]
func Upload(c *gin.Context) {
	db := db.GetDB()

	var err error

	Created := time.Now().Unix()

	var query struct {
		Category string `header:"Category" binding:"required"`
		RecordID int    `header:"RecordID"`
	}

	//category := c.Request.Header.Get("category")
	if err := c.ShouldBindHeader(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error(), "type": "validation"})
		return
	}

	file := File{
		RecordID: query.RecordID,
	}

	sender, err := QueryMember(c)
	if err != nil {
		c.Abort()
		return
	}

	err = file.ValidateCategory(query.Category)
	if err != nil {
		c.Abort()
		return
	}

	form, _ := c.MultipartForm()
	files := form.File["files"]

	filename := cast.ToString(sender.MemberID) + "_" + uuid.New().String() // + ".jpg"

	for _, fileItem := range files {

		if file.Category == constants.CATEGORY_DOCUMENT {
			filename = filename + filepath.Ext(fileItem.Filename)
		}

		if file.Category == constants.CATEGORY_PHOTO {
			filename = filename + ".jpg"
		}

		err := c.SaveUploadedFile(fileItem, "/var/server/static/temp_"+filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if file.Category == constants.CATEGORY_DOCUMENT {

		file.File, err = ioutil.ReadFile("/var/server/static/temp_" + filename)

	}

	if file.Category == constants.CATEGORY_PHOTO {

		//https://github.com/h2non/bimg/issues/241
		bimg.VipsCacheSetMax(0)
		bimg.VipsCacheSetMaxMem(0)

		file.File, err = bimg.Read("/var/server/static/temp_" + filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}

		file.File, err = bimg.NewImage(file.File).Convert(bimg.JPEG)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}

		if bimg.NewImage(file.File).Type() != "jpeg" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  "CONVERSION_ERROR",
			})
			return
		}

		options := bimg.Options{
			Quality: 80,
		}

		file.File, err = bimg.NewImage(file.File).Process(options)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}

	}

	os.Remove("/var/server/static/temp_" + filename)

	_, err = s3.Upload(filename, file.File)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO Upload (MemberID, FileName, Category, Created) VALUES ($1, $2, $3, $4)", sender.MemberID, filename, file.Category, Created)
	tx.Commit()

	if file.Category == constants.CATEGORY_AVATAR {
		tx := db.MustBegin()
		tx.MustExec("UPDATE Member SET Image=$1 WHERE MemberID=$2", filename, sender.MemberID)
		tx.Commit()
	}

	if file.Category == constants.CATEGORY_PHOTO || file.Category == constants.CATEGORY_DOCUMENT {

		err = file.DeterminePosition()
		if err != nil {
			return
		}

		tx := db.MustBegin()
		tx.MustExec("INSERT INTO Media (InvestID, FileName, Category, Created, MemberID, Position) VALUES ($1, $2, $3, $4, $5, $6)", query.RecordID, filename, file.Category, Created, sender.MemberID, file.Position)
		tx.Commit()
	}

	var RecordID int

	if err := db.Get(&RecordID, "SELECT UploadID FROM Upload WHERE MemberID=$1 AND FileName=$2 AND Category=$3 AND Created=$4 ORDER BY UploadID DESC", sender.MemberID, filename, file.Category, Created); err != nil {
		if err != sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   true,
		"URL":      viper.GetString("s3_cdn_url") + filename,
		"RecordID": RecordID,
	})
}

func (file *File) ValidateCategory(Category string) error {

	if Category != constants.CATEGORY_AVATAR && Category != constants.CATEGORY_PHOTO && Category != constants.CATEGORY_DOCUMENT {
		return errors.New("INVALID_CATEGORY")
	}

	file.Category = Category

	return nil
}

func (file *File) DeterminePosition() error {
	db := db.GetDB()

	var Position int

	if err := db.Get(&Position, "SELECT Position FROM Media WHERE InvestID=$1 AND Category=$2 ORDER BY Position DESC", file.RecordID, file.Category); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	file.Position = Position + 1

	return nil
}

func UploadFile(fileName string, content string) error {
	decode, err := base64.StdEncoding.DecodeString(content)
	file, err := os.Create(fileName)
	defer file.Close()
	_, err = file.Write(decode)
	return err
}
