package controllers

import (
	"encoding/json"
	"fmt"
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func CreateTRXInspection(c *gin.Context) {
	idInspectionStr := c.PostForm("id_inspection")
	idUserStr := c.PostForm("id_user")
	deviceID := c.PostForm("device_id")
	chainingIDStr := c.PostForm("chaining_id")
	nameInspection := c.PostForm("name_inspection")
	imageUrl := c.PostForm("image_url")
	detailJSON := c.PostForm("details")

	userCompanyID := c.GetString("company_id")
	username := c.GetString("username")
	storageType := c.GetString("storage_type")

	if idInspectionStr == "" || idUserStr == "" || detailJSON == "" {
		utils.JSONError(c, http.StatusBadRequest, "id_inspection, id_user, and details are required")
		return
	}

	deviceTimeStr := c.PostForm("device_time")
	deviceTimezone := c.PostForm("device_timezone")
	var parsedDeviceTime time.Time
	if deviceTimeStr != "" {
		var err error
		parsedDeviceTime, err = time.Parse(time.RFC3339, deviceTimeStr)
		if err != nil {
			utils.JSONError(c, http.StatusBadRequest, "Invalid device_time format. Use ISO 8601 format")
			return
		}

		parsedDeviceTime = parsedDeviceTime.UTC()

	}

	// ================= PAYLOAD STRUCT =================
	type AnswerPayload struct {
		QuestionID uint   `json:"question_id"`
		AnswerText string `json:"answer_text"`
		AnswerFile string `json:"answer_file"`
		Type       string `json:"type"`
	}

	type DetailPayload struct {
		IdCoordinate   uint            `json:"id_coordinate"`
		CaptureUrl     string          `json:"capture_url"`
		CaptureFile    string          `json:"capture_file"`
		Description    string          `json:"description"`
		DeviceTime     string          `json:"device_time"`
		DeviceTimezone string          `json:"device_timezone"`
		Answers        []AnswerPayload `json:"answers"`
	}

	var payloads []DetailPayload
	if err := json.Unmarshal([]byte(detailJSON), &payloads); err != nil {
		utils.JSONError(c, http.StatusBadRequest, "Invalid details JSON: "+err.Error())
		return
	}

	// ================= INIT =================
	idInspection := parseUint(idInspectionStr)
	idUser := parseUint(idUserStr)
	chainingID := parseUint(chainingIDStr)
	now := time.Now()

	tx := config.DB.Begin()

	// ================= CREATE INSPECTION =================
	trnAssuranceID := utils.GenerateTRNCode("assurance", deviceID, idUser)
	inspection := models.TrxInspection{
		IdInspection:   idInspection,
		NameInspection: nameInspection,
		ImageUrl:       imageUrl,
		IdUser:         idUser,
		DeviceID:       deviceID,
		CompanyID:      userCompanyID,
		ChainingID:     chainingID,
		TrnAssuranceID: trnAssuranceID,
		DeviceTime:     parsedDeviceTime,
		DeviceTimezone: deviceTimezone,
		CreatedBy:      username,
		UpdatedBy:      username,
	}

	////////CREATE MASTER/////////
	if err := tx.Create(&inspection).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// ================= AMBIL QUESTION SEKALI =================
	questionIDsMap := make(map[uint]bool)
	coordinateIDsMap := make(map[uint]bool)

	for _, d := range payloads {
		coordinateIDsMap[d.IdCoordinate] = true
		for _, a := range d.Answers {
			questionIDsMap[a.QuestionID] = true
		}
	}

	var questionIDs, coordinateIDs []uint
	for id := range questionIDsMap {
		questionIDs = append(questionIDs, id)
	}
	for id := range coordinateIDsMap {
		coordinateIDs = append(coordinateIDs, id)
	}

	var questions []models.MstrInspectionQuestion
	if err := tx.Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var sams []models.MstrInspectionDetail
	if err := tx.Where("id IN ?", coordinateIDs).Find(&sams).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	questionMap := make(map[uint]models.MstrInspectionQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	samMap := make(map[uint]models.MstrInspectionDetail)
	for _, s := range sams {
		samMap[s.Id] = s
	}

	// ================= RESPONSE STRUCT =================
	type AnswerResponse struct {
		QuestionID   uint   `json:"question_id"`
		QuestionText string `json:"question"`
		Type         string `json:"type"`
		AnswerText   string `json:"answer_text,omitempty"`
		AnswerFile   string `json:"answer_file,omitempty"`
		TrnAnswerID  string `json:"trn_answer_id"`
	}

	type DetailResponse struct {
		IdCoordinate        uint             `json:"sam_id"`
		TrnSamID            string           `json:"trn_sam_id"`
		SamName             string           `json:"sam_name"`
		EvidenceIsMandatory bool             `json:"evidence_is_mandatory"`
		SendNowIsMandatory  bool             `json:"send_now_is_mandatory"`
		SamContentText      string           `json:"sam_content_text"`
		SamTypeTriggerID    *uint            `json:"sam_type_trigger_id"`
		SamTypeTriggerName  string           `json:"sam_type_trigger_name"`
		CaptureUrl          string           `json:"evidence_capture"`
		CaptureFile         string           `json:"evidence_capture_type"`
		Description         string           `json:"evidence_description"`
		Answers             []AnswerResponse `json:"answers"`
	}

	var responseDetails []DetailResponse

	// ================= LOOP DETAIL =================
	for _, dp := range payloads {
		sam, ok := samMap[dp.IdCoordinate]
		if !ok {
			tx.Rollback()
			utils.JSONError(c, http.StatusBadRequest,
				fmt.Sprintf("SAM ID %d not found", dp.IdCoordinate))
			return
		}

		trnSamID := utils.GenerateTRNCode("assurance_sam", deviceID, idUser)

		// ================= PARSE LAT / LNG =================
		var latitude *float64
		var longitude *float64
		latStr := c.PostForm("latitude")
		lonStr := c.PostForm("longitude")

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

		var parsedDetailDeviceTime time.Time

		if dp.DeviceTime != "" {
			t, err := time.Parse(time.RFC3339, dp.DeviceTime)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusBadRequest,
					fmt.Sprintf("Invalid device_time format for coordinate %d", dp.IdCoordinate))
				return
			}

			parsedDetailDeviceTime = t.UTC()

		}

		detail := models.TrxInspectionDetail{
			IdTrxInspection: inspection.Id,
			IdCoordinate:    dp.IdCoordinate,
			TrnSamID:        trnSamID,
			CaptureFile:     dp.CaptureFile,
			Description:     dp.Description,
			Latitude:        latitude,
			Longitude:       longitude,
			CreatedBy:       username,
			UpdatedBy:       username,

			DeviceTime:     parsedDetailDeviceTime,
			DeviceTimezone: dp.DeviceTimezone,
		}

		// ===== Upload Evidence =====
		if dp.CaptureUrl != "" {
			fileHeader, err := c.FormFile(dp.CaptureUrl)
			if err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusBadRequest, err.Error())
				return
			}

			file, _ := fileHeader.Open()
			defer file.Close()

			var objectKey string
			switch storageType {
			case "e2":
				fileKey := GenerateE2ObjectKey(c, "Trn-Assurance", fileHeader.Filename)
				objectKey, err = UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+userCompanyID, nil)
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
					filepath.Join("companies", userCompanyID, "Trn-Assurance"),
				)
				if err != nil {
					tx.Rollback()
					utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
					return
				}
			default:
				utils.JSONError(c, 400, "Invalid storage type")
				return
			}

			detail.CaptureUrl = objectKey
		}

		/////CREATE DETAIL///////
		if err := tx.Create(&detail).Error; err != nil {
			tx.Rollback()
			utils.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}

		// ================= GET TYPE NAME =================
		var typeTriggerName string
		tx.Model(&models.MstrTypeTrigger{}).
			Where("id = ?", sam.TypeTriggerID).
			Pluck("type_name", &typeTriggerName)

		respDetail := DetailResponse{
			IdCoordinate:        dp.IdCoordinate,
			TrnSamID:            trnSamID,
			SamName:             sam.NameCoordinate,
			EvidenceIsMandatory: sam.RequiredCoordinate,
			SendNowIsMandatory:  sam.SendNow,
			SamContentText:      sam.TutorialCoordinate,
			SamTypeTriggerID:    sam.TypeTriggerID,
			SamTypeTriggerName:  typeTriggerName,
			CaptureUrl:          detail.CaptureUrl,
			CaptureFile:         detail.CaptureFile,
			Description:         detail.Description,
		}

		// ===== Answers =====
		var answers []models.TrxInspectionAnswer
		var respAnswers []AnswerResponse

		for _, a := range dp.Answers {
			q, ok := questionMap[a.QuestionID]
			if !ok {
				tx.Rollback()
				utils.JSONError(c, http.StatusBadRequest,
					fmt.Sprintf("Question ID %d not found", a.QuestionID))
				return
			}

			trnAnswerID := utils.GenerateTRNCode("assurance_answer", deviceID, idUser)
			answer := models.TrxInspectionAnswer{
				IdTrxInspectionDetail: detail.Id,
				QuestionID:            a.QuestionID,
				TrnAnswerID:           trnAnswerID,
				CreatedBy:             username,
				UpdatedBy:             username,
				CreatedAt:             now,
				UpdatedAt:             now,
			}

			respAnswer := AnswerResponse{
				QuestionID:   a.QuestionID,
				QuestionText: q.Text,
				Type:         a.Type,
				TrnAnswerID:  trnAnswerID,
			}

			if strings.ToLower(a.Type) == "image" {
				fileHeader, err := c.FormFile(a.AnswerFile)
				if err != nil {
					tx.Rollback()
					utils.JSONError(c, http.StatusBadRequest, err.Error())
					return
				}

				file, _ := fileHeader.Open()
				defer file.Close()

				var objectKey string
				switch storageType {
				case "e2":
					fileKey := GenerateE2ObjectKey(c, "Trn-Assurance", fileHeader.Filename)
					objectKey, err = UploadFileToE2(c, file, fileKey, fileHeader.Header.Get("Content-Type"), "Assurance/"+userCompanyID, nil)
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
						filepath.Join("companies", userCompanyID, "Trn-Assurance"),
					)
					if err != nil {
						tx.Rollback()
						utils.JSONError(c, http.StatusInternalServerError, "Failed upload: "+err.Error())
						return
					}
				default:
					utils.JSONError(c, 400, "Invalid storage type")
					return
				}

				answer.AnswerFile = objectKey
				respAnswer.AnswerFile = objectKey
			} else {
				answer.AnswerText = strings.TrimSpace(a.AnswerText)
				respAnswer.AnswerText = answer.AnswerText
			}

			answers = append(answers, answer)
			respAnswers = append(respAnswers, respAnswer)
		}

		/////////CREATE ANSWER//////////
		if len(answers) > 0 {
			if err := tx.Create(&answers).Error; err != nil {
				tx.Rollback()
				utils.JSONError(c, http.StatusInternalServerError, err.Error())
				return
			}
		}

		respDetail.Answers = respAnswers
		responseDetails = append(responseDetails, respDetail)
	}

	// ================= GET DEVICE & CHAINING NAME =================

	var deviceName string
	tx.Model(&models.MstrDevice{}).
		Where("device_id = ?", deviceID).
		Pluck("device_name", &deviceName)

	var chainingName string
	tx.Model(&models.MstrChaining{}).
		Where("id = ?", chainingID).
		Pluck("name_chaining", &chainingName)

	// ================= RAW PAYLOAD =================
	finalPayload := gin.H{
		"assurance_id":         idInspectionStr,
		"trn_assurance_id":     trnAssuranceID,
		"assurance_name":       nameInspection,
		"assurance_image_path": imageUrl,
		"user_id":              idUserStr,
		"username":             username,
		"device_id":            deviceID,
		"device_name":          deviceName,
		"company_id":           userCompanyID,
		"chaining_id":          chainingIDStr,
		"chaining_name":        chainingName,
		"details":              responseDetails,
	}

	finalJSON, err := json.Marshal(finalPayload)
	if err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Model(&inspection).
		Update("raw_payload", datatypes.JSON(finalJSON)).Error; err != nil {
		tx.Rollback()
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()
	utils.JSONSuccess(c, "Inspection created successfully", finalPayload)
}

func UpdateTRXInspectionByID(c *gin.Context) {
	id := c.Param("id")
	var inspection models.TrxInspection
	if err := config.DB.Preload("Details").First(&inspection, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "TRX Inspection not found")
		return
	}

	var input models.TrxInspection
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	inspection.NameInspection = input.NameInspection
	inspection.ImageUrl = input.ImageUrl
	inspection.IdUser = input.IdUser
	inspection.UpdatedAt = time.Now()
	inspection.UpdatedBy = input.UpdatedBy

	// Update detail: hapus dulu lalu simpan ulang (simple logic)
	config.DB.Where("id_trx_inspection = ?", inspection.Id).Delete(&models.TrxInspectionDetail{})
	for i := range input.Details {
		input.Details[i].IdTrxInspection = inspection.Id
		input.Details[i].CreatedAt = time.Now()
		input.Details[i].UpdatedAt = time.Now()
	}
	inspection.Details = input.Details

	if err := config.DB.Save(&inspection).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "TRX Inspection updated", inspection)
}

func DeleteTRXInspectionByID(c *gin.Context) {

	id := c.Param("id")

	if err := config.DB.Where("id_trx_inspection = ?", id).Delete(&models.TrxInspectionDetail{}).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.TrxInspection{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "TRX Inspection deleted", nil)
}

func GetFilteredTRXInspections(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	nameInspection := c.Query("name_inspection")
	idInspection := c.Query("id_inspection")
	idTrx := c.Query("id_trx")

	var inspections []models.TrxInspection
	query := config.DB.Preload("Details")

	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	if updatedBy != "" {
		query = query.Where("updated_by = ?", updatedBy)
	}

	if nameInspection != "" {
		query = query.Where("name_inspection ILIKE ?", "%"+nameInspection+"%")
	}

	if idInspection != "" {
		query = query.Where("id_inspection = ?", idInspection)
	}

	if idTrx != "" {
		query = query.Where("id = ?", idTrx)
	}

	query = query.Order("id DESC")

	if err := query.Find(&inspections).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered TRX Inspections", inspections)
}

func AssuranceDownloadJSON(c *gin.Context) {
	trnAssuranceID := c.Query("trn_assurance_id")
	if trnAssuranceID == "" {
		c.JSON(400, gin.H{"error": "trn_assurance_id required"})
		return
	}

	var jsonData string

	tx := config.DB.Raw(`
        SELECT raw_payload::text
        FROM trx_inspection
        WHERE trn_assurance_id = ?
    `, trnAssuranceID).Scan(&jsonData)

	if tx.Error != nil {
		c.JSON(500, gin.H{"error": tx.Error.Error()})
		return
	}

	if tx.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "data not found"})
		return
	}

	// filename
	filename := fmt.Sprintf("%s.json", trnAssuranceID)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")

	c.Data(200, "application/json", []byte(jsonData))
}
