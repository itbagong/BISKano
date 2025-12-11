package karamodel

type OpStatus string

const (
	TrxOK         OpStatus = "OK"
	TrxNeedReview OpStatus = "Need review"
	TrxCancelled  OpStatus = "Cancelled"
)

type OpCode string

const (
	Checkin  OpCode = "Checkin"
	Checkout OpCode = "Checkout"
	Transit  OpCode = "Transit"
	None     OpCode = "None"
)
