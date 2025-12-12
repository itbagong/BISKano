package ficologic

import (
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

// JournalPosting
type JournalPosting interface {
	ExtractHeader() error
	ExtractLines() error
	Validate() error
	Calculate() error
	PostingProfile() *ficomodel.PostingProfile
	Status() string
	Submit() (*ficomodel.PostingApproval, error)
	Approve(string, string) (string, error)
	Post() error
	Preview() *tenantcoremodel.PreviewReport
	Transactions(name string) []orm.DataModel
}

// BalanceCalc
type BalanceCalc[T, O any] interface {
	Get(opt O) ([]T, error)
	Update(bal T) (T, error)
	Sync(sources []T) ([]T, error)
	MakeSnapshot(companyID string, balanceDate time.Time) error
}

// TrxSplit
type TrxSplit[T, O any] interface {
	SetOpts(Opts O) TrxSplit[T, O]
	Split(qty float64, newStatus string) ([]T, []T, error)
}
