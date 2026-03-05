package models

type AssuranceFailure struct {
	NameChaining        string `json:"name_chaining"`
	DeviceName          string `json:"device_name"`
	AssuranceName       string `json:"assurance_name"`
	SAMName             string `json:"sam_name"`
	CompanyID           string `json:"company_id"`
	QuestionText        string `json:"question_text"`
	AnswerUser          string `json:"answer_user"`
	IsFailure           bool   `json:"is_failure"`
	CaptureFile         string `json:"capture_file"`
	Evidence            bool   `json:"evidence"`
	EvidenceCaptureFile string `json:"evidence_capture_file"`
	EvidenceTypeFile    string `json:"evidence_type_file"`
	EvidenceDescription string `json:"evidence_description"`
	CreatedBy           string `json:"created_by"`
	SubmittedAt         string `json:"submitted_at"`
}
