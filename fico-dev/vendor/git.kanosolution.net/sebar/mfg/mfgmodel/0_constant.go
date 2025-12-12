package mfgmodel

type InventTrxType string

const (
	JournalWorkRequest                InventTrxType = "Work Request"
	JournalWorkOrderPlan              InventTrxType = "Work Order"
	JournalWorkOrderReportConsumption InventTrxType = "Work Order Report Consumption"
	JournalWorkOrderReportResource    InventTrxType = "Work Order Report Resource"
	JournalWorkOrderReportOutput      InventTrxType = "Work Order Report Output"
)

var SourceTypeURLMap = map[string]string{
	string(JournalWorkRequest):   "mfg/WorkRequestor",
	string(JournalWorkOrderPlan): "mfg/WorkOrderPlan",
}
