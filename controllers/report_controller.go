package controllers

import (
	"fmt"
	"go-api/config"
	"go-api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// //////////////////////////////////////////////////
// ASSURANCE FAILURE REPORT
func GetAssuranceFailure(c *gin.Context) {
	assuranceName := c.Query("assurance_name")
	chainingName := c.Query("chaining_name")
	deviceName := c.Query("device_name")
	createdBy := c.Query("created_by")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var reports []AssuranceFailureReport

	query := config.DB.
		Table("trx_inspection_answer tia").
		Select(`
			mc.name_chaining AS chaining_name,
			md.device_name,
			mi.name_inspection AS assurance_name,
			mid.name_coordinate AS sam_name,
			mi.company_id,
			q.text AS question_text,
			tia.answer_text AS answer_user,
			os.is_correct AS is_failure,
			sub.answer_file ,
			tid.capture_file AS evidence_type_file,
			tid.capture_url AS evidence_capture_url,
			tid.description AS evidence_description,
			tia.created_by,
			ti.created_at AS submitted_at
		`).
		Joins("JOIN trx_inspection_detail tid ON tid.id = tia.id_trx_inspection_detail").
		Joins("JOIN trx_inspection ti ON ti.id = tid.id_trx_inspection").
		Joins("JOIN mstr_inspection mi ON mi.id = ti.id_inspection").
		Joins("JOIN mstr_inspection_detail mid ON mid.id = tid.id_coordinate").
		Joins("JOIN mstr_inspection_question q ON q.id = tia.question_id").
		Joins("JOIN mstr_chainings mc ON mc.id = ti.chaining_id").
		Joins("JOIN mstr_device md ON md.device_id = ti.device_id").
		Joins(`
			LEFT JOIN mstr_inspection_question_option os 
			ON os.inspection_question_id = q.id
			AND os.is_correct = true
			AND LOWER(TRIM(os.text)) = LOWER(TRIM(tia.answer_text))
		`).
		Joins("LEFT JOIN mstr_company mc2 ON mc2.company_id = mi.company_id").
		Joins(`
			LEFT JOIN (
				SELECT id_trx_inspection_detail, MAX(answer_file) AS answer_file
				FROM trx_inspection_answer
				WHERE answer_file <> ''
				GROUP BY id_trx_inspection_detail
			) sub ON sub.id_trx_inspection_detail = tia.id_trx_inspection_detail
		`).
		Where("os.id IS NOT NULL")

	/* =====================
	   ROLE FILTER
	===================== */
	role := c.GetString("role")
	companyID := c.GetString("company_id")

	if role != "super-admin" {
		query = query.Where("mi.company_id = ?", companyID)
	}

	/* =====================
	   TEXT FILTER
	===================== */
	if assuranceName != "" {
		query = query.Where("mi.name_inspection ILIKE ?", "%"+assuranceName+"%")
	}

	if chainingName != "" {
		query = query.Where("mc.name_chaining ILIKE ?", "%"+chainingName+"%")
	}

	if deviceName != "" {
		query = query.Where("md.device_name ILIKE ?", "%"+deviceName+"%")
	}

	if createdBy != "" {
		query = query.Where("tia.created_by = ?", createdBy)
	}

	/* =====================
	   DATE FILTER
	===================== */
	if startDate != "" {
		query = query.Where("ti.created_at >= ?", startDate+" 00:00:00")
	}

	if endDate != "" {
		query = query.Where("ti.created_at <= ?", endDate+" 23:59:59")
	}

	/* =====================
	   EXECUTE
	===================== */
	if err := query.
		Order("ti.created_at DESC").
		Scan(&reports).Error; err != nil {

		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Assurance Failure Report", reports)
}

type AssuranceFailureReport struct {
	ChainingName  string `json:"chaining_name"`
	DeviceName    string `json:"device_name"`
	AssuranceName string `json:"assurance_name"`
	SamName       string `json:"sam_name"`
	CompanyID     string `json:"company_id"`
	QuestionText  string `json:"question_text"`
	AnswerUser    string `json:"answer_user"`
	AnswerFile    string `json:"answer_file"`
	IsFailure     bool   `json:"is_failure"`

	EvidenceTypeFile    string `json:"evidence_type_file"`
	EvidenceCaptureURL  string `json:"evidence_capture_url"`
	EvidenceDescription string `json:"evidence_description"`

	CreatedBy   string    `json:"created_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// //////////////////////////////////////////////////
// ASSURANCE DOWNLOAD JSON REPORT
func GetAssuranceDownloadJson(c *gin.Context) {
	assuranceName := c.Query("assurance_name")
	chainingName := c.Query("chaining_name")
	deviceName := c.Query("device_name")
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var reports []AssuranceDownloadJson

	query := config.DB.
		Table("trx_inspection ti").
		Select(`
            ti.trn_assurance_id,
            ti.name_inspection AS assurance_name,
            mc.name_chaining as chaining_name,
            md.device_name,
            ti.company_id,
            ti.created_by,
            ti.created_at,
			ti.raw_payload
        `).
		Joins("JOIN mstr_device md ON md.device_id = ti.device_id").
		Joins("JOIN mstr_chainings mc ON mc.id = ti.chaining_id").
		Where("ti.trn_assurance_id IS NOT NULL")

	// ROLE FILTER
	role := c.GetString("role")
	companyID := c.GetString("company_id")

	if role != "super-admin" {
		query = query.Where("ti.company_id = ?", companyID)
	}

	// ASSURANCE NAME FILTER
	if assuranceName != "" {
		query = query.Where("ti.name_inspection ILIKE ?", "%"+assuranceName+"%")
	}

	// CHAINING NAME FILTER
	if chainingName != "" {
		query = query.Where("mc.name_chaining ILIKE ?", "%"+chainingName+"%")
	}

	// DEVICE NAME FILTER
	if deviceName != "" {
		query = query.Where("md.device_name ILIKE ?", "%"+deviceName+"%")
	}

	// USER FILTER
	if createdBy != "" {
		query = query.Where("ti.created_by = ?", createdBy)
	}

	if updatedBy != "" {
		query = query.Where("ti.updated_by = ?", updatedBy)
	}

	// DATE RANGE FILTER
	if startDate != "" {
		query = query.Where("ti.created_at >= ?", startDate+" 00:00:00")
	}

	if endDate != "" {
		query = query.Where("ti.created_at <= ?", endDate+" 23:59:59")
	}

	// EXECUTE
	if err := query.
		Order("ti.created_at DESC").
		Scan(&reports).Error; err != nil {

		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Assurance Downloadn Json Report", reports)
}

type AssuranceDownloadJson struct {
	TrnAssuranceID string         `json:"trn_assurance_id"`
	AssuranceName  string         `json:"assurance_name"`
	ChainingName   string         `json:"chaining_name"`
	DeviceName     string         `json:"device_name"`
	CompanyID      string         `json:"company_id"`
	CreatedBy      string         `json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	RawPayload     datatypes.JSON `json:"raw_payload"`
}

func AdminSummaryReport(c *gin.Context) {
	companyID := c.GetString("company_id")
	db := config.DB

	// === Ambil filter ===
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	//companyID := c.Query("company_id")

	// === Base WHERE ===
	where := " WHERE 1=1 "
	args := []interface{}{}
	i := 1

	if startDate != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", i)
		args = append(args, startDate)
		i++
	}

	if endDate != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", i)
		args = append(args, endDate)
		i++
	}

	role := c.GetString("role")
	if role != "super-admin" {
		if companyID != "" {
			where += fmt.Sprintf(" AND company_id = $%d", i)
			args = append(args, companyID)
			i++
		}
	}

	var res AdminSummaryResponse

	// === Assurance ===
	queryAssurance := `
		SELECT COUNT(*)
		FROM mstr_inspection
	` + where + ` AND deleted_at IS NULL`

	db.Raw(queryAssurance, args...).Scan(&res.AssuranceCount)

	// === Questionnaire ===
	queryQuestionnaire := `
		SELECT COUNT(*)
		FROM questionnaires
	` + where + ` AND deleted_at IS NULL AND is_active = true`

	db.Raw(queryQuestionnaire, args...).Scan(&res.QuestionnaireCount)

	// === Device ===
	queryDevice := `
		SELECT COUNT(*)
		FROM mstr_device
	` + where + ` AND deleted_at IS NULL AND is_active = true`

	db.Raw(queryDevice, args...).Scan(&res.DeviceCount)

	// === User ===
	queryUser := `
		SELECT COUNT(*)
		FROM mstr_user
	` + where + ` AND deleted_at IS NULL AND is_active = true`

	db.Raw(queryUser, args...).Scan(&res.UserCount)

	// === Chaining ===
	queryChaining := `
		SELECT COUNT(*)
		FROM mstr_chainings
	` + where + ` AND deleted_at IS NULL AND is_active = true`

	db.Raw(queryChaining, args...).Scan(&res.ChainingCount)

	// === Group ===
	queryGroup := `
		SELECT COUNT(*)
		FROM mstr_group
	` + where + ` AND deleted_at IS NULL`

	db.Raw(queryGroup, args...).Scan(&res.GroupCount)

	// === Response Gin ===
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin Summary Report",
		"data":    res,
	})
}

type AdminSummaryResponse struct {
	AssuranceCount     int `json:"assurance_count"`
	QuestionnaireCount int `json:"questionnaire_count"`
	DeviceCount        int `json:"device_count"`
	UserCount          int `json:"user_count"`
	ChainingCount      int `json:"chaining_count"`
	GroupCount         int `json:"group_count"`
}
