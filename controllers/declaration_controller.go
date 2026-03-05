package controllers

import (
	"errors"
	"go-api/config"
	"go-api/models"
	"go-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateDeclaration(c *gin.Context) {

	var declaration models.MstrDeclaration
	if err := c.ShouldBindJSON(&declaration); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	CompanyID := c.GetString("company_id")
	username := c.GetString("username")
	declaration.CompanyID = CompanyID
	declaration.CreatedBy = username
	declaration.UpdatedBy = username

	if err := config.DB.Create(&declaration).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Declaration created", declaration)
}

func UpdateDeclarationByID(c *gin.Context) {
	declarationID := c.Param("id")

	var declaration models.MstrDeclaration
	if err := config.DB.
		Where("id = ?", declarationID).
		First(&declaration).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Device not found")
		return
	}

	if err := c.ShouldBindJSON(&declaration); err != nil {
		utils.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update field
	CompanyID := c.GetString("company_id")
	username := c.GetString("username")
	declaration.CompanyID = CompanyID
	declaration.UpdatedBy = username

	if err := config.DB.Save(&declaration).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Declaration updated", declaration)
}

func DeleteDeclarationByID(c *gin.Context) {
	id := c.Param("id")

	deletedBy := c.GetString("username")

	// Set DeletedBy
	if err := config.DB.Model(&models.MstrDeclaration{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := config.DB.Delete(&models.MstrDeclaration{}, id).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSONSuccess(c, "Declaration deleted", nil)
}

func GetFilteredDeclaration(c *gin.Context) {

	var declaration models.MstrDeclaration
	query := config.DB.Model(&models.MstrDeclaration{})

	role := c.GetString("role")
	companyID := c.GetString("company_id")

	if role != "super-admin" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.
		Order("id DESC").
		First(&declaration).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(c, http.StatusNotFound, "Data not found")
			return
		}

		utils.JSONError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONSuccess(c, "Last declaration", declaration)
}
