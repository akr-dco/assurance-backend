package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateTrxSamAdhoc(c *gin.Context) {
	// ================= REQUEST DATA =================
	description := c.PostForm("description")
	deviceID := c.PostForm("device_id")
	captureFileType := c.PostForm("capture_file_type")

	latStr := c.PostForm("latitude")
	lonStr := c.PostForm("longitude")

	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")
	storageType := c.GetString("storage_type")

	// ================= VALIDATION =================
	if description == "" || deviceID == "" {
		utils.JSONError(c, http.StatusBadRequest, "description and device_id are required")
		return
	}

	// ================= PARSE LAT / LNG =================
	var latitude *float64
	var longitude *float64

	if latStr != "" && lonStr != "" {
		lat, err1 := strconv.ParseFloat(latStr, 64)
		lon, err2 := strconv.ParseFloat(lonStr, 64)

		if err1 != nil || err2 != nil {
			utils.JSONError(c, http.StatusBadRequest, "Invalid latitude or longitude")
			return
		}

		latitude = &lat
		longitude = &lon
	}

	// ================= BEGIN TRANSACTION =================
	tx := config.DB.Begin()

	trx := models.TrxSamAdhoc{
		Description:     description,
		DeviceID:        deviceID,
		CompanyID:       userCompanyID,
		CaptureFileType: captureFileType,
		Latitude:        latitude,
		Longitude:       longitude,
		CreatedBy:       username,
		UpdatedBy:       username,
	}

	// ================= UPLOAD CAPTURE FILE =================
	fileHeader, err := c.FormFile("capture_file")
	if err == nil { // file ada (optional)
		file, _ := fileHeader.Open()
		defer file.Close()

		var objectKey string

		switch storageType {
		case "e2":
			fileKey := GenerateE2ObjectKey(c, "Trx-Sam-Adhoc", fileHeader.Filename)
			objectKey, err = UploadFileToE2(
				c,
				file,
				fileKey,
				fileHeader.Header.Get("Content-Type"),
				"SamAdhoc/"+userCompanyID,
				nil,
			)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, err.Error())
				return
			}

		case "local":
			objectKey, err = utils.UploadFileToLocal(
				c,
				fileHeader,
				"uploads",
				filepath.Join("companies", userCompanyID, "SamAdhoc"),
			)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, err.Error())
				return
			}

		default:
			tx.Rollback()
			utils.JSONError(c, http.StatusBadRequest, "Invalid storage type")
			return
		}

		trx.CaptureFile = objectKey
	}

	// ================= CREATE MASTER =================
	if err := tx.Create(&trx).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()

	// ================= RESPONSE =================
	//config.DB.
	//	Preload("Company").
	//	Preload("Device").
	//	First(&trx, trx.ID)

	utils.JSONSuccess(c, "TrxSamAdhoc created successfully", trx)
}
