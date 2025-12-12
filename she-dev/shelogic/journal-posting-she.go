package shelogic

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

type SHEPosting struct {
	header         orm.DataModel
	lines          []SHEPostingLine
	inventTrxs     []orm.DataModel
	preview        *tenantcoremodel.PreviewReport
	postingProfile *ficomodel.PostingProfile

	ctx           *kaos.Context
	db            *datahub.Hub
	ev            kaos.EventHub
	journalID     string
	userID        string
	companyID     string
	SourceType    string
	FieldDateName string
}

type SHEPostingLine struct {
	TransType string
}

func NewSHEPosting(ctx *kaos.Context, db *datahub.Hub, ev kaos.EventHub, journalID, userID, companyID, model string) (*SHEPosting, error) {
	p := new(SHEPosting)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID
	p.companyID = companyID
	p.SourceType = model
	p.ctx = ctx

	switch model {
	case string(shemodel.PostTypeCoaching):
		p.header = new(shemodel.Coaching)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeCoaching),
		})
		p.lines = newLines
		p.FieldDateName = "Date"
	case string(shemodel.PostTypeInduction):
		p.header = new(shemodel.Induction)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeInduction),
		})
		p.lines = newLines
		p.FieldDateName = "InductionDate"
	case string(shemodel.PostTypeCSMS):
		p.header = new(shemodel.Csms)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeCSMS),
		})
		p.lines = newLines
		p.FieldDateName = "CsmsDate"
	case string(shemodel.PostTypeJSA):
		p.header = new(shemodel.Jsa)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeJSA),
		})
		p.lines = newLines
		p.FieldDateName = "JsaDate"
	case string(shemodel.PostTypeSafetycard):
		p.header = new(shemodel.SafetyCard)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeSafetyCard),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypePICA):
		p.header = new(shemodel.Pica)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypePICA),
		})
		p.lines = newLines
		p.FieldDateName = "DueDate"
	case string(shemodel.PostTypeMeeting):
		p.header = new(shemodel.Meeting)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeMeeting),
		})
		p.lines = newLines
		p.FieldDateName = "MeetingDate"
	case string(shemodel.PostTypeLegalRegister):
		p.header = new(shemodel.LegalRegister)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeLegalRegister),
		})
		p.lines = newLines
		p.FieldDateName = "Date"
	case string(shemodel.PostTypeLegalCompliance):
		p.header = new(shemodel.LegalCompliance)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeLegalCompliance),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeInvestigation):
		p.header = new(shemodel.Investigasi)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeInvestigation),
		})
		p.lines = newLines
		p.FieldDateName = "AccidentDate"
	case string(shemodel.PostTypeIBPR):
		p.header = new(shemodel.IBPR)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeIBPR),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeRSCA):
		p.header = new(shemodel.RSCA)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeRSCA),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeAudit):
		p.header = new(shemodel.Audit)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeAudit),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeObservation):
		p.header = new(shemodel.Observasi)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeObservation),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeInspection):
		p.header = new(shemodel.Inspection)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeInspection),
		})
		p.lines = newLines
		p.FieldDateName = "LastUpdate"
	case string(shemodel.PostTypeSidak):
		p.header = new(shemodel.Sidak)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeSidak),
		})
		p.lines = newLines
		p.FieldDateName = "DateTime"
	case string(shemodel.PostTypeP3K):
		p.header = new(shemodel.P3k)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeP3K),
		})
		p.lines = newLines
		p.FieldDateName = "Date"
	case string(shemodel.PostTypeMCU):
		p.header = new(shemodel.MCUTransaction)

		newLines := []SHEPostingLine{}
		newLines = append(newLines, SHEPostingLine{
			TransType: string(shemodel.TransactionTypeMCU),
		})
		p.lines = newLines
		p.FieldDateName = "Date"
	default:
		return p, fmt.Errorf("invalid module: %s", model)
	}

	return p, nil
}

func (p *SHEPosting) ExtractHeader() error {
	var err error

	p.header, err = datahub.GetByID(p.db, p.header, p.journalID)
	if err != nil {
		return fmt.Errorf("invalid journal id")
	}

	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	jtID := ""
	if field := v.FieldByName("JournalTypeID"); field.IsValid() {
		if field.Kind() == reflect.String {
			jtID = field.String()
		} else {
			// Handle non-string type here, e.g., convert to string or log a warning
			jtID = fmt.Sprintf("%v", field.Interface()) // Converts the value to string if it's not a string
		}
	} else {
		// Handle the case where the field doesn't exist
		jtID = "Field not found"
	}

	jt, err := datahub.GetByID(p.db, new(ficomodel.SheJournalType), jtID)
	if err != nil {
		return fmt.Errorf("missing: journal type: %s: %s", jtID, err.Error())
	}

	p.postingProfile, err = datahub.GetByID(p.db, new(ficomodel.PostingProfile), jt.PostingProfileID)
	if err != nil {
		return fmt.Errorf("invalid: posting profile: %s", jt.PostingProfileID)
	}

	return nil
}

func (p *SHEPosting) ExtractLines() error {
	// p.lines = lo.Map(p.header.Lines, func(line shemodel.Sidak, index int) shemodel.Sidak {
	// 	return line
	// })
	return nil
}

func (p *SHEPosting) Validate() error {
	// if len(p.lines) == 0 {
	// 	return fmt.Errorf("missing: lines")
	// }

	// groupedLines := lo.GroupBy(p.lines, func(l scmmodel.InventReceiveIssueLine) string {
	// 	return fmt.Sprintf("%s|%s", l.ItemID, l.InventDim.InventDimID)
	// })

	// for _, gls := range groupedLines {
	// 	bal, _ := NewInventBalanceCalc(p.db).Get(&InventBalanceCalcOpts{
	// 		CompanyID: p.header.CompanyID,
	// 		ItemID:    []string{gls[0].ItemID},
	// 		InventDim: scmmodel.InventDimension{InventDimID: gls[0].InventDim.InventDimID},
	// 	})

	// 	qtyPlanned := lo.SumBy(bal, func(b *scmmodel.ItemBalance) float64 {
	// 		return b.QtyPlanned
	// 	})

	// 	qtyConfirmed := lo.SumBy(gls, func(g scmmodel.InventReceiveIssueLine) float64 {
	// 		return g.Qty
	// 	})

	// 	if moreThan(qtyConfirmed, qtyPlanned, true) {
	// 		return fmt.Errorf("over qty: %s: receive %.2f, planned %.2f", gls[0].ItemID, qtyConfirmed, qtyPlanned)
	// 	}
	// }

	return nil
}

func (p *SHEPosting) Calculate() error {
	// for index, line := range p.lines {
	// 	inventTrx, err := receiveIssueLineToTrx(p.db, p.header, line)
	// 	if err != nil {
	// 		return fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
	// 	}
	// 	inventTrx.CompanyID = p.header.CompanyID
	// 	inventTrx.Status = scmmodel.ItemConfirmed
	// 	inventTrx.TrxDate = p.header.TrxDate
	// 	p.inventTrxs = append(p.inventTrxs, inventTrx)
	// }
	return nil
}

func (p *SHEPosting) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *SHEPosting) Status() string {
	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if field := v.FieldByName("Status"); field.IsValid() {
		return string(field.String())
	}

	return ""
}

func (p *SHEPosting) markAsReady(pa *ficomodel.PostingApproval) error {
	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v.FieldByName("Status").SetString("READY")

	p.db.Save(p.header)

	err := p.updatePicaStatus(v, string(ficomodel.JournalStatusReady))
	if err != nil {
		return err
	}

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}

func (p *SHEPosting) Submit() (*ficomodel.PostingApproval, error) {
	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	vID := ""
	var vDate time.Time
	if field := v.FieldByName("ID"); field.IsValid() {
		vID = string(field.String())
	}
	if field := v.FieldByName(p.FieldDateName); field.IsValid() {
		vDate = field.Interface().(time.Time)
	}

	pa, isNew, err := ficologic.GetOrCreatePostingApproval(p.db, p.userID, p.companyID, p.SourceType, vID, *p.postingProfile, nil, true, true, receiveIssuelineToficoLines(p.lines, tenantcoremodel.TrxModule("SHE")), "", vID, vDate, 0)
	if p.postingProfile.NeedApproval {
		if err != nil {
			return nil, fmt.Errorf("create approval: %s", err.Error())
		}
		if !isNew {
			return nil, fmt.Errorf("duplicate: approval: %s, %s", ficomodel.SubledgerAccounting, vID)
		}
		v.FieldByName("Status").SetString("SUBMITTED")

		err := p.createPicaJournal(v)
		if err != nil {
			return nil, fmt.Errorf("error when create pica journal: %s", err.Error())
		}

		err = p.updatePicaStatus(v, string(ficomodel.JournalStatusSubmitted))
		if err != nil {
			return nil, err
		}

		p.db.Save(p.header)
	}

	return pa, nil
}

func (p *SHEPosting) Approve(op string, txt string) (string, error) {
	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	vID := ""
	if field := v.FieldByName("ID"); field.IsValid() {
		vID = string(field.String())
	}

	pa, err := ficologic.GetPostingApprovalBySource(p.db, p.companyID, p.SourceType, vID, true)
	if err != nil {
		return "", fmt.Errorf("posting approval: %s", err.Error())
	}

	if err = pa.UpdateApproval(p.db, p.userID, op, txt); err != nil {
		return pa.Status, fmt.Errorf("posting approval: %s", err.Error())
	}
	if err = p.db.Save(pa); err != nil {
		return pa.Status, fmt.Errorf("posting approval save: %s", err.Error())
	}
	switch pa.Status {
	case "REJECTED":
		v.FieldByName("Status").SetString(string(pa.Status))
		p.db.Save(p.header)

		err := p.updatePicaStatus(v, string(ficomodel.JournalStatusRejected))
		if err != nil {
			return pa.Status, err
		}

	case "APPROVED":
		if err = p.markAsReady(pa); err != nil {
			return pa.Status, err
		}
	}
	return pa.Status, nil
}

func (p *SHEPosting) Post() error {
	v := reflect.ValueOf(p.header)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v.FieldByName("Status").SetString(string(ficomodel.JournalStatusPosted))
	p.db.Update(p.header, "Status")

	// add to asset and unit calendar
	// err := PostAssets(p.db, p.journalID)
	// if err != nil {
	// 	return err
	// }

	err := p.updatePicaStatus(v, string(ficomodel.JournalStatusPosted))
	if err != nil {
		return err
	}

	return nil
}

func (p *SHEPosting) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *SHEPosting) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxs)
}

func receiveIssuelineToficoLines(lineTrxs []SHEPostingLine, sourceType tenantcoremodel.TrxModule) []ficomodel.JournalLine {
	return lo.Map(lineTrxs, func(line SHEPostingLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(sourceType, line.TransType)
		return *jl
	})
}

// updatePicaStatus update pica status in journal
func (p *SHEPosting) updatePicaStatus(v reflect.Value, status string) error {
	sourceNumber := v.FieldByName("SourceNumber").String()
	module := v.FieldByName("SourceModule").String()

	switch p.SourceType {
	case string(shemodel.PostTypePICA):
		switch module {
		case string(shemodel.MODULE_SAFETYCARD):
			card := new(shemodel.SafetyCard)
			err := p.db.GetByID(card, sourceNumber)
			if err != nil {
				return fmt.Errorf("error when get safety card: %s", err.Error())
			}

			card.Pica.Status = status
			p.db.Update(card, "Pica")
		case string(shemodel.MODULE_INVESTIGATION):
			invests := new(shemodel.Investigasi)
			err := p.db.GetByID(invests, sourceNumber)
			if err != nil {
				return fmt.Errorf("error when get investigation: %s", err.Error())
			}

			// update pica status
			for i, p := range invests.PICA {
				if p.ID == v.FieldByName("ID").String() {
					invests.PICA[i].Status = status
					break
				}
			}
		case string(shemodel.MODULE_MEETING):
			meeting := new(shemodel.Meeting)
			err := p.db.GetByID(meeting, sourceNumber)
			if err != nil {
				return fmt.Errorf("error when get meeting: %s", err.Error())
			}

			// update pica status
			for i, p := range meeting.Result {
				if p.UsePica {
					if p.Pica.ID == v.FieldByName("ID").String() {
						meeting.Result[i].Pica.Status = status
						break
					}
				}
			}

			// check if all pica posted
			isAllPicaComplete := true
			for _, p := range meeting.Result {
				if p.UsePica {
					if p.Pica.Status != status {
						isAllPicaComplete = false
						break
					}
				}
			}

			if isAllPicaComplete {
				meeting.PicaStatus = status
			}

			p.db.Update(meeting, "PicaStatus", "Result")
		}
	}

	return nil
}

// createPicaJournal create pica journal only when submit
func (p *SHEPosting) createPicaJournal(v reflect.Value) error {
	switch p.SourceType {
	case string(shemodel.PostTypeSafetycard):
		pica := v.FieldByName("Pica").Interface().(*shemodel.Pica)
		if pica != nil && pica.EmployeeID != "" {
			//generate number pica
			tenantcorelogic.MWPreAssignSequenceNo("Pica", false, "_id")(p.ctx, pica)

			//save to pica
			if e := p.db.GetByID(new(shemodel.Pica), pica.ID); e != nil {
				pica.JournalTypeID = "SHE-001"
				pica.PostingProfileID = "PP-SHE"
				pica.SourceNumber = p.journalID
				pica.SourceModule = shemodel.MODULE_SAFETYCARD
				pica.Status = string(ficomodel.JournalStatusDraft)
				if e := p.db.Insert(pica); e != nil {
					return errors.New("error insert Pica: " + e.Error())
				}
			} else {
				if e := p.db.Save(pica); e != nil {
					return errors.New("error update Pica: " + e.Error())
				}
			}
		}
	case string(shemodel.PostTypeInvestigation):
		picas := v.FieldByName("PICA").Interface().([]shemodel.PICA)
		for i, pic := range picas {
			pica := new(shemodel.Pica)
			pica.SourceNumber = p.journalID
			pica.SourceModule = shemodel.MODULE_INVESTIGATION
			pica.FindingDescription = pic.Cause
			pica.Comment = pic.Action
			pica.EmployeeID = pic.PIC
			pica.DueDate = pic.DueDate
			pica.JournalTypeID = "SHE-001"
			pica.PostingProfileID = "PP-SHE"
			pica.Dimension = v.FieldByName("Dimension").Interface().(tenantcoremodel.Dimension)
			pica.Status = string(shemodel.SHEStatusSubmitted)
			tenantcorelogic.MWPreAssignSequenceNo("Pica", false, "_id")(p.ctx, pica)

			picas[i].ID = pica.ID

			p.db.Save(pica)
		}

		invest := new(shemodel.Investigasi)
		err := p.db.GetByID(invest, p.journalID)
		if err != nil {
			return fmt.Errorf("error when get investigation: %s", err.Error())
		}

		invest.PICA = picas
		p.db.Save(invest)
	case string(shemodel.PostTypeMeeting):
		picas := v.FieldByName("Result").Interface().([]shemodel.MeetingResult)
		for _, val := range picas {
			if val.Pica != nil && val.Pica.EmployeeID != "" {
				//generate number pica
				tenantcorelogic.MWPreAssignSequenceNo("Pica", false, "_id")(p.ctx, &val.Pica)

				//save to pica
				if e := p.db.GetByID(new(shemodel.Pica), val.Pica.ID); e != nil {
					val.Pica.JournalTypeID = "SHE-001"
					val.Pica.PostingProfileID = "PP-SHE"
					val.Pica.SourceNumber = p.journalID
					val.Pica.SourceModule = shemodel.MODULE_MEETING
					val.Pica.Status = string(ficomodel.JournalStatusDraft)
					if e := p.db.Insert(val.Pica); e != nil {
						return errors.New("error insert Pica: " + e.Error())
					}
				} else {
					if e := p.db.Save(val.Pica); e != nil {
						return errors.New("error update Pica: " + e.Error())
					}
				}
			}
		}

		meeting := new(shemodel.Meeting)
		err := p.db.GetByID(meeting, p.journalID)
		if err != nil {
			return fmt.Errorf("error when get meeting: %s", err.Error())
		}

		meeting.Result = picas
		p.db.Save(meeting)
	}

	return nil
}
