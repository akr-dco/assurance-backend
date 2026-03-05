package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TrxInspection struct {
	Id             uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for inspection transaction"`
	IdInspection   uint           `json:"id_inspection" gorm:"not null;comment:Foreign key to MstrInspection"`
	NameInspection string         `json:"name_inspection" gorm:"type:varchar(200);not null;comment:Name of the inspection"`
	ImageUrl       string         `json:"image_url" gorm:"type:varchar(500);comment:URL of the inspection image"`
	IdUser         uint           `json:"id_user" gorm:"not null;comment:Foreign key to MstrUser performing the inspection"`
	ChainingID     uint           `json:"chaining_id" gorm:"comment:chaining_id"`
	DeviceID       string         `json:"device_id" gorm:"type:varchar(50);comment:device identifier"`
	CompanyID      string         `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	RawPayload     datatypes.JSON `gorm:"type:jsonb"`
	TrnAssuranceID string         `json:"trn_assurance_id" gorm:"type:varchar(150);comment:Unique trn assurance"`

	ServerTime     time.Time `json:"server_time" gorm:"type:timestamptz;comment:Server Timestamp when transaction was created"`
	DeviceTime     time.Time `json:"device_time" gorm:"type:timestamptz;comment:Device Timestamp when transaction was created from device"`
	DeviceTimezone string    `json:"device_timezone" gorm:"type:varchar(50);comment:Device timezone identifier"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when transaction was created"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when transaction was last updated"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);not null;comment:User or system that created this record"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this record"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Details []TrxInspectionDetail `json:"details" gorm:"foreignKey:IdTrxInspection;constraint:OnDelete:CASCADE;comment:List of detailed coordinates and captured data for this transaction"`
}

func (m *TrxInspection) BeforeCreate(tx *gorm.DB) (err error) {
	m.ServerTime = time.Now().UTC()
	return
}

func (TrxInspection) TableName() string {
	return "trx_inspection"
}

type TrxInspectionDetail struct {
	Id              uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:Primary key for inspection detail record"`
	IdTrxInspection uint   `json:"id_trx_inspection" gorm:"not null;comment:Foreign key to TrxInspection"`
	IdCoordinate    uint   `json:"id_coordinate" gorm:"not null;comment:Foreign key to MstrInspectionDetail coordinate"`
	CaptureFile     string `json:"capture_file" gorm:"type:varchar(255);comment:Captured file name (Video, Audio, Photo)"`
	CaptureUrl      string `json:"capture_url" gorm:"type:varchar(500);comment:URL for the captured file"`
	Description     string `json:"description" gorm:"type:text;comment:Description or note for the captured data"`
	TrnSamID        string `json:"trn_sam_id" gorm:"type:varchar(150);comment:Unique trn sam"`

	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,7);comment:Latitude coordinate"`
	Longitude *float64 `json:"longitude" gorm:"type:decimal(10,7);comment:Longitude coordinate"`

	ServerTime     time.Time `json:"server_time" gorm:"type:timestamptz;comment:Server Timestamp when transaction was created"`
	DeviceTime     time.Time `json:"device_time" gorm:"type:timestamptz;comment:Device Timestamp when transaction was created from device"`
	DeviceTimezone string    `json:"device_timezone" gorm:"type:varchar(50);comment:Device timezone identifier"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when detail was created"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when detail was last updated"`
	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this detail record"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this detail record"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`

	Details []TrxInspectionAnswer `json:"details" gorm:"foreignKey:IdTrxInspectionDetail;constraint:OnDelete:CASCADE;comment:List of detailed coordinates and captured data for this transaction"`
}

func (m *TrxInspectionDetail) BeforeCreate(tx *gorm.DB) (err error) {
	m.ServerTime = time.Now().UTC()
	return
}

func (TrxInspectionDetail) TableName() string {
	return "trx_inspection_detail"
}

type TrxInspectionAnswer struct {
	ID                    uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	IdTrxInspectionDetail uint           `json:"id_trx_inspection_detail" gorm:"index;not null;comment:Foreign key to Question"`
	QuestionID            uint           `json:"question_id" gorm:"index;not null;comment:Foreign key to Question"`
	AnswerText            string         `json:"answer_text" gorm:"type:text;comment:User's answer in text format (Yes/No, A..E, essay)"`
	AnswerFile            string         `json:"answer_file" gorm:"type:varchar(255);comment:Relative path to uploaded image for image-type question"`
	TrnAnswerID           string         `json:"trn_answer_id" gorm:"type:varchar(150);comment:Unique trn answer"`
	CreatedBy             string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy             string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt             time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt             time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy             string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (TrxInspectionAnswer) TableName() string {
	return "trx_inspection_answer"
}

type TrxSamAdhoc struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Description string `json:"description" gorm:"type:varchar(255);comment:Description of the sam adhoc"`

	CaptureFile     string `json:"capture_file" gorm:"type:varchar(255);comment:Relative path to uploaded image for image-type question"`
	CaptureFileType string `json:"capture_file_type" gorm:"type:varchar(50);comment:Captured file name (Video, Audio, Photo)"`

	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,7);comment:Latitude coordinate"`
	Longitude *float64 `json:"longitude" gorm:"type:decimal(10,7);comment:Longitude coordinate"`

	CompanyID string `json:"company_id" gorm:"type:varchar(50);not null;comment:Foreign key to Company (MstrCompany.CompanyID)"`
	//Company   MstrCompany `gorm:"foreignKey:CompanyID;references:CompanyID;constraint:OnDelete:CASCADE"`

	DeviceID string `json:"device_id" gorm:"type:varchar(50);not null;comment:Foreign key to Device (MstrDevice.DeviceID)"`
	//Device   MstrDevice `gorm:"foreignKey:DeviceID;references:DeviceID;constraint:OnDelete:CASCADE"`

	CreatedBy string         `json:"created_by" gorm:"type:varchar(100);comment:User or system that created this answer record"`
	UpdatedBy string         `json:"updated_by" gorm:"type:varchar(100);comment:User or system that last updated this answer record"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:Timestamp when answer was created"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:Timestamp when answer was last updated"`
	DeletedBy string         `json:"deleted_by" gorm:"type:varchar(100);comment:User or system that deleted the record"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:Timestamp when the record was soft deleted"`
}

func (TrxSamAdhoc) TableName() string {
	return "trx_sam_adhoc"
}
