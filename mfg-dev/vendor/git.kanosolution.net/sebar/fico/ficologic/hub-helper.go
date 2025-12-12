package ficologic

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

func ValidatePost(db *datahub.Hub, hdr orm.DataModel) error {
	var (
		hdrID      string
		hdrCoID    string
		hdrTrxDate time.Time
		hdrStatus  ficomodel.JournalStatus
	)

	rf := reflector.From(hdr)
	rf.GetTo("ID", &hdrID)
	rf.GetTo("CompanyID", &hdrCoID)
	rf.GetTo("Status", &hdrStatus)
	rf.GetTo("TrxDate", &hdrTrxDate)

	if hdrStatus != "READY" && hdrStatus != "DRAFT" && hdrStatus != "POSTED" {
		return fmt.Errorf("journal is not yet approved: %s", hdrID)
	}
	if GetPeriodStatus(db, hdrCoID, "", hdrTrxDate, "Finance") != ficomodel.PeriodOpen {
		return fmt.Errorf("period %s is not open", hdrTrxDate.Format("02-Jan-2006"))
	}

	return nil
}

func PostModelSave[H orm.DataModel](db *datahub.Hub, header H, voucherNoName string, trxs map[string][]orm.DataModel) (string, error) {
	var (
		companyID string
		trxDate   time.Time
		dimension tenantcoremodel.Dimension
	)
	// check status
	err := ValidatePost(db, header)
	if err != nil {
		return "", fmt.Errorf("validate posting: %s", err.Error())
	}

	// headerInfo
	rf := reflector.From(header)
	if err = rf.GetTo("CompanyID", &companyID); err != nil {
		return "", fmt.Errorf("company id: %s", err.Error())
	}
	if err = rf.GetTo("TrxDate", &trxDate); err != nil {
		return "", fmt.Errorf("trx date: %s", err.Error())
	}
	if err = rf.GetTo("Dimension", &dimension); err != nil {
		return "", fmt.Errorf("trx date: %s", err.Error())
	}

	// get voucher no
	resp := new(tenantcorelogic.NumSeqClaimRespond)
	if e := ficoconfig.Config.EventHub.Publish(ficoconfig.Config.SeqNumTopic,
		&tenantcorelogic.NumSeqMapping{Mapping: voucherNoName, CompanyID: companyID, Date: trxDate}, resp, nil); e != nil {
		return "", fmt.Errorf("voucher no: %s", e.Error())
	}
	if resp.Number == "" {
		return "", fmt.Errorf("invalid voucher number setup")
	}
	voucherNo := resp.Number

	// save
	balErrors := []error{}
	for name, records := range trxs {
		for _, record := range records {
			if record != nil {
				rf := reflector.From(record)
				if err := rf.Set("VoucherNo", voucherNo).Flush(); err != nil {
					return "", fmt.Errorf("set voucher no to %s: %s", name, err.Error())
				}
				if err := db.Save(record); err != nil {
					return "", fmt.Errorf("save %s: %s", name, err.Error())
				}
			}
		}

		// sync balances
		var balErr error
		switch name {
		case new(ficomodel.LedgerTransaction).TableName():
			ids := lo.Uniq(lo.Map(FromDataModels(records, new(ficomodel.LedgerTransaction)),
				func(record *ficomodel.LedgerTransaction, index int) string {
					return record.Account.ID
				}))
			_, balErr = NewLedgerBalanceHub(db).Sync(nil,
				LedgerBalanceOpt{
					CompanyID:  companyID,
					AccountIDs: ids,
				})

		case new(ficomodel.CashTransaction).TableName():
			ids := lo.Uniq(lo.Map(FromDataModels(records, new(ficomodel.CashTransaction)),
				func(record *ficomodel.CashTransaction, index int) string {
					return record.CashBank.ID
				}))
			_, balErr = NewCashBalanceHub(db).Sync(nil,
				CashBalanceOpt{
					CompanyID:  companyID,
					AccountIDs: ids,
				})

		case new(ficomodel.CustomerTransaction).TableName():
			ids := lo.Uniq(lo.Map(FromDataModels(records, new(ficomodel.CustomerTransaction)),
				func(record *ficomodel.CustomerTransaction, index int) string {
					return record.Customer.ID
				}))
			_, balErr = NewCustomerBalanceHub(db).Sync(nil,
				CustomerBalanceOpt{
					CompanyID:  companyID,
					AccountIDs: ids,
				})

		case new(ficomodel.VendorTransaction).TableName():
			ids := lo.Uniq(lo.Map(FromDataModels(records, new(ficomodel.VendorTransaction)),
				func(record *ficomodel.VendorTransaction, index int) string {
					return record.Vendor.ID
				}))
			_, balErr = NewVendorBalanceHub(db).Sync(nil,
				VendorBalanceOpt{
					CompanyID:  companyID,
					AccountIDs: ids,
				})
		}

		if balErr != nil {
			balErrors = append(balErrors, balErr)
		}
	}

	if len(balErrors) > 0 {
		return voucherNo, fmt.Errorf("balance sync error: %s", strings.Join(
			lo.Map(balErrors, func(e error, index int) string {
				return e.Error()
			}), ". "))
	}

	return voucherNo, nil
}

func ToDataModels[T orm.DataModel](records []T) []orm.DataModel {
	return lo.Map(records, func(t T, index int) orm.DataModel {
		return t
	})
}

func FromDataModels[T orm.DataModel](models []orm.DataModel, model T) []T {
	return lo.Map(models, func(t orm.DataModel, index int) T {
		return t.(T)
	})
}

func GetVoucherNo(p *tenantcoremodel.PreviewReport) string {
	return p.Header.GetString("VoucherNo")
}
