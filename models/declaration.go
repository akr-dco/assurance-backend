package models

import (
	"time"

	"gorm.io/gorm"
)

type MstrDeclaration struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Declaration string         `json:"declaration" gorm:"type:text;comment:"`
	CompanyID   string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy   string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (MstrDeclaration) TableName() string {
	return "mstr_declaration"
}
