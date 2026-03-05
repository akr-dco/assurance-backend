package controllers

import (
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func CreateCompany(c *gin.Context) {
	var company models.MstrCompany

	company.CompanyName = c.PostForm("company_name")
	company.CompanyID = c.PostForm("company_id")
	company.StorageType = c.DefaultPostForm("storage_type", "local")
	company.ReportType = c.DefaultPostForm("report_type", "local")

	// E2 fields (boleh kosong kalau local)
	company.E2Endpoint = c.PostForm("e2_endpoint")
	company.E2Region = c.PostForm("e2_region")
	company.E2BucketName = c.PostForm("e2_bucket_name")
	company.E2AccessKey = c.PostForm("e2_access_key")
	company.E2SecretKey = c.PostForm("e2_secret_key")

	// ==== UPLOAD LOGO ====
	fileHeader, err := c.FormFile("image_url")
	if err == nil {

		switch company.StorageType {

		case "e2":
			file, err := fileHeader.Open()
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}
			defer file.Close()

			override := &E2Config{
				Endpoint:   company.E2Endpoint,
				Region:     company.E2Region,
				BucketName: company.E2BucketName,
				AccessKey:  company.E2AccessKey,
				SecretKey:  company.E2SecretKey,
			}

			fileKey := GenerateE2ObjectKey(c, "Company-Logo", fileHeader.Filename)

			objectKey, err := UploadFileToE2(
				c,
				file,
				fileKey,
				fileHeader.Header.Get("Content-Type"),
				"Assurance/"+company.CompanyID,
				override,
			)
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}

			company.ImageUrl = objectKey

		default: // local
			path, err := utils.UploadFileToLocal(
				c,
				fileHeader,
				"uploads",
				filepath.Join("companies", company.CompanyID, "Company-Logo"),
			)
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}

			company.ImageUrl = path
		}
	}

	username := c.GetString("username")
	company.CreatedBy = username
	company.UpdatedBy = username

	if err := config.DB.Create(&company).Error; err != nil {
		utils.JSONError(c, 500, "Failed create company")
		return
	}

	utils.JSONSuccess(c, "Company created successfully", company)
}

func UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")

	var company models.MstrCompany
	if err := config.DB.WithContext(c.Request.Context()).
		First(&company, companyID).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Company not found")
		return
	}

	// ===== BASIC FIELDS =====
	if v := c.PostForm("company_name"); v != "" {
		company.CompanyName = v
	}

	if v := c.PostForm("company_id"); v != "" {
		company.CompanyID = v
	}

	if v := c.PostForm("is_active"); v != "" {
		company.IsActive = (v == "true" || v == "1")
	}

	// storage type (default ikut nilai lama / local)
	company.StorageType = c.DefaultPostForm("storage_type", company.StorageType)
	if company.StorageType == "" {
		company.StorageType = "local"
	}

	// report type (default ikut nilai lama / local)
	company.ReportType = c.DefaultPostForm("report_type", company.ReportType)
	if company.ReportType == "" {
		company.ReportType = "local"
	}

	// ===== E2 FIELDS =====
	if v := c.PostForm("e2_endpoint"); v != "" {
		company.E2Endpoint = v
	}
	if v := c.PostForm("e2_region"); v != "" {
		company.E2Region = v
	}
	if v := c.PostForm("e2_bucket_name"); v != "" {
		company.E2BucketName = v
	}
	if v := c.PostForm("e2_access_key"); v != "" {
		company.E2AccessKey = v
	}
	if v := c.PostForm("e2_secret_key"); v != "" {
		company.E2SecretKey = v
	}

	// ===== UPLOAD LOGO (jika ada file baru) =====
	fileHeader, err := c.FormFile("image_url")
	if err == nil {

		switch company.StorageType {

		case "e2":
			file, err := fileHeader.Open()
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}
			defer file.Close()

			override := &E2Config{
				Endpoint:   company.E2Endpoint,
				Region:     company.E2Region,
				BucketName: company.E2BucketName,
				AccessKey:  company.E2AccessKey,
				SecretKey:  company.E2SecretKey,
			}

			fileKey := GenerateE2ObjectKey(
				c,
				"Company-Logo",
				fileHeader.Filename,
			)

			objectKey, err := UploadFileToE2(
				c,
				file,
				fileKey,
				fileHeader.Header.Get("Content-Type"),
				"Assurance/"+company.CompanyID,
				override,
			)
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}

			company.ImageUrl = objectKey

		default: // local
			path, err := utils.UploadFileToLocal(
				c,
				fileHeader,
				"uploads",
				filepath.Join("Companies", company.CompanyID, "Company-Logo"),
			)
			if err != nil {
				utils.JSONError(c, 500, err.Error())
				return
			}

			company.ImageUrl = path
		}
	}

	// ===== AUDIT =====
	username := c.GetString("username")
	company.UpdatedBy = username

	if err := config.DB.WithContext(c.Request.Context()).
		Save(&company).Error; err != nil {
		utils.JSONError(c, 500, "Failed update company")
		return
	}

	utils.JSONSuccess(c, "Company updated successfully", company)
}

func DeleteCompany(c *gin.Context) {
	CompanyID := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrCompany{}).
		Where("id = ?", CompanyID).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Cek relasi di tabel lain, misalnya mstr_user
	/*
		var count int64
		config.DB.Model(&models.MstrCompany{}).Where("id = ?", CompanyID).Count(&count)
		if count > 0 {
			utils.JSONError(c, http.StatusBadRequest, "Company masih digunakan di tabel lain")
			return
		}
	*/

	if err := config.DB.Delete(&models.MstrCompany{}, CompanyID).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Company deleted", nil)
}

func GetFilteredCompanies(c *gin.Context) {
	createdBy := c.Query("created_by")
	updatedBy := c.Query("updated_by")
	companyName := c.Query("company_name")

	var companies []models.MstrCompany
	query := config.DB.Model(&models.MstrCompany{})

	role, _ := c.Get("role")
	userCompanyID, _ := c.Get("company_id")
	if role != "super-admin" {
		query = query.Where("company_id = ?", userCompanyID)
	}

	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if updatedBy != "" {
		query = query.Where("updated_by = ?", updatedBy)
	}
	if companyName != "" {
		query = query.Where("company_name ILIKE ?", "%"+companyName+"%")
	}

	if err := query.Order("id DESC").Find(&companies).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Filtered companies", companies)
}
