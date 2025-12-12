package shemodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type STModule string
type STModuleAttachment string
type SHEStatus string
type SHERelatedSite string
type MCUTemplateType string
type MCUCategory string
type MCUVisitResult string
type SOPNatureOfChange string
type SOPDocumentType string
type IBPRRutinitas string
type IBPRCondition string
type IBPRLingkup string
type AuditCategory string
type PostType string
type PostOp string
type TransactionType string

const (
	MODULE_PICA          STModule = "PICA"
	MODULE_SAFETYCARD    STModule = "SAFETYCARD"
	MODULE_MEETING       STModule = "MEETING"
	MODULE_INSPECTION    STModule = "INSPECTION"
	MODULE_INVESTIGATION STModule = "INVESTIGATION"

	MODULE_ATTACHMENT_PICA       STModuleAttachment = "SHE_PICA"
	MODULE_ATTACHMENT_SAFETYCARD STModuleAttachment = "SHE_SAFETYCARD"
	MODULE_ATTACHMENT_MEETING    STModuleAttachment = "SHE_MEETING"
	MODULE_ATTACHMENT_INSPECTION STModuleAttachment = "SHE_INSPECTION"

	SHEStatusActive    SHEStatus = "Active"
	SHEStatusInActive  SHEStatus = "InActive"
	SHEStatusOpen      SHEStatus = "Open"
	SHEStatusCompleted SHEStatus = "Completed"
	SHEStatusApproval  SHEStatus = "Approval"
	SHEStatusDraft     SHEStatus = "DRAFT"
	SHEStatusSubmitted SHEStatus = "SUBMITTED"
	SHEStatusReady     SHEStatus = "READY"
	SHEStatusPosted    SHEStatus = "POSTED"
	SHEStatusApproved  SHEStatus = "APPROVED"
	SHEStatusRejected  SHEStatus = "REJECTED"

	PostOpPreview PostOp = "Preview"
	PostOpSubmit  PostOp = "Submit"
	PostOpApprove PostOp = "Approve"
	PostOpReject  PostOp = "Reject"
	PostOpPost    PostOp = "Post"

	PostTypeSidak           PostType = "SIDAK"
	PostTypeCoaching        PostType = "COACHING"
	PostTypeInduction       PostType = "INDUCTION"
	PostTypeCSMS            PostType = "CSMS"
	PostTypeJSA             PostType = "JSA"
	PostTypeSafetycard      PostType = "SAFETYCARD"
	PostTypePICA            PostType = "PICA"
	PostTypeMeeting         PostType = "MEETING"
	PostTypeLegalRegister   PostType = "LEGALREGISTER"
	PostTypeLegalCompliance PostType = "LEGALCOMPLIANCE"
	PostTypeInvestigation   PostType = "INVESTIGATION"
	PostTypeIBPR            PostType = "IBPR"
	PostTypeRSCA            PostType = "RSCA"
	PostTypeAudit           PostType = "Audit"
	PostTypeObservation     PostType = "Observation"
	PostTypeInspection      PostType = "Inspection"
	PostTypeP3K             PostType = "P3K"
	PostTypeMCU             PostType = "MCU"

	TransactionTypeSidak           TransactionType = "Sidak"
	TransactionTypeCoaching        TransactionType = "Coaching"
	TransactionTypeInduction       TransactionType = "Induction"
	TransactionTypeCSMS            TransactionType = "CSMSTransaction"
	TransactionTypeJSA             TransactionType = "JobSafetyAnalysis"
	TransactionTypeSafetyCard      TransactionType = "SafetyCard"
	TransactionTypePICA            TransactionType = "PICA"
	TransactionTypeMeeting         TransactionType = "Meeting"
	TransactionTypeLegalRegister   TransactionType = "LegalRegister"
	TransactionTypeLegalCompliance TransactionType = "LegalCompliance"
	TransactionTypeInvestigation   TransactionType = "Investigation"
	TransactionTypeIBPR            TransactionType = "IBPR"
	TransactionTypeRSCA            TransactionType = "RSCA"
	TransactionTypeAudit           TransactionType = "Audit"
	TransactionTypeObservation     TransactionType = "Observation"
	TransactionTypeInspection      TransactionType = "Inspection"
	TransactionTypeP3K             TransactionType = "P3K"
	TransactionTypeMCU             TransactionType = "MCU"

	RS_HO       SHERelatedSite = "HO"
	RS_ALL_SITE SHERelatedSite = "All Site"

	MCU_TYPE_LIST   MCUTemplateType = "List"
	MCU_TYPE_RANGE  MCUTemplateType = "Range"
	MCU_TYPE_STRING MCUTemplateType = "String"

	MCU_CATEGORY_CANDIDATE MCUCategory = "Candidate"
	MCU_CATEGORY_EMPLOYEE  MCUCategory = "Employee"

	MCU_VISIT_RESULT_FIT   MCUVisitResult = "Fit"
	MCU_VISIT_RESULT_UNFIT MCUVisitResult = "UnFit"

	NOC_PEMBUATAN SOPNatureOfChange = "Pembuatan"
	NOC_REVISI    SOPNatureOfChange = "Revisi"
	NOC_OBSOLETE  SOPNatureOfChange = "Obsolete"

	SOP_DOCUMENT_TYPE_SOP    SOPDocumentType = "SOP"
	SOP_DOCUMENT_TYPE_MANUAL SOPDocumentType = "Manual"
	SOP_DOCUMENT_TYPE_STD    SOPDocumentType = "STD"
	SOP_DOCUMENT_TYPE_INK    SOPDocumentType = "INK"
	SOP_DOCUMENT_TYPE_FORM   SOPDocumentType = "Form"

	IBPR_RUTINITAS_RUTIN     IBPRRutinitas = "Rutin"
	IBPR_RUTINITAS_NON_RUTIN IBPRRutinitas = "Non-Rutin"

	IBPR_CONDITION_NORMAL    IBPRCondition = "Normal"
	IBPR_CONDITION_ABNORMAL  IBPRCondition = "Abnormal"
	IBPR_CONDITION_EMERGENCY IBPRCondition = "Emergency"

	IBPR_LINGKUP_K3 IBPRLingkup = "K3"
	IBPR_LINGKUP_L  IBPRLingkup = "L"

	AUDIT_CATEGORY_SMK3   AuditCategory = "SMK3"
	AUDIT_CATEGORY_SMKPAU AuditCategory = "SMKPAU"
	AUDIT_CATEGORY_SMKP   AuditCategory = "SMKP"

	ModuleSidak tenantcoremodel.TrxModule = "SIDAK"
)

var SourceTypeURLMap = map[string]string{
	string(ModuleSidak): "she/sidak",
}
