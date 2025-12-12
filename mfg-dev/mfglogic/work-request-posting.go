package mfglogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkRequestPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkRequest
	trxType string
}

func NewWorkRequestPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkRequest, scmmodel.InventReceiveIssueLine] {
	workRequest := new(WorkRequestPosting)
	workRequest.ctx = ctx
	workRequest.opt = &opt

	pvd := ficologic.PostingProvider[*mfgmodel.WorkRequest, scmmodel.InventReceiveIssueLine](workRequest)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *WorkRequestPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkRequest, lines []scmmodel.InventReceiveIssueLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}

func (p *WorkRequestPosting) Header() (*mfgmodel.WorkRequest, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkRequest), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	line := p.buildDummyLine()
	j.Lines = []scmmodel.InventReceiveIssueLine{line}

	p.trxType = string(j.TrxType)
	if p.trxType == "" {
		p.trxType = string(scmmodel.InventReceive)
	}

	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile")
	}

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.header = j
	return j, pp, nil
}

func (p *WorkRequestPosting) Lines() ([]scmmodel.InventReceiveIssueLine, error) {
	return p.header.Lines, nil
}

func (p *WorkRequestPosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkRequest, lines []scmmodel.InventReceiveIssueLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	trxs := map[string][]orm.DataModel{}
	return p.GetPreview(opt, header, lines), trxs, 0, err
}

func (p *WorkRequestPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkRequest, lines []scmmodel.InventReceiveIssueLine) *tenantcoremodel.PreviewReport {
	preview := tenantcoremodel.PreviewReport{}

	requestor, err := datahub.GetByID(opt.Db, new(tenantcoremodel.Employee), header.Name)
	if err == nil {
		requestor = new(tenantcoremodel.Employee)
	}

	cc, err := datahub.GetByID(opt.Db, new(tenantcoremodel.DimensionMaster), header.Dimension.Get("CC"))
	if err != nil {
		cc = new(tenantcoremodel.DimensionMaster)
	}

	eq, err := datahub.GetByID(opt.Db, new(tenantcoremodel.Asset), header.EquipmentNo)
	if err != nil {
		eq = new(tenantcoremodel.Asset)
	}

	signatureRequestor := tenantcoremodel.Signature{
		ID:        requestor.Name,
		Header:    "Pembuat",
		Footer:    requestor.Name,
		Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	}
	preview.Signature = append(preview.Signature, signatureRequestor)

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(header.SourceType), header.ID)
	preview.Signature = append(preview.Signature, signature...)

	// preview header and footer
	preview.Header = codekit.M{
		"Data": [][]string{
			{"WR No:", header.ID, "", "WR Date:", FormatDate(header.TrxDate)},
			{"Site:", header.Dimension.Get("Site"), "", "Requestor:", requestor.Name},
			{"Source Type:", string(header.SourceType), "", "Departement:", cc.Label},
			{"Source ID:", header.SourceID, "", "WR Type:", header.WorkRequestType},
			{"", "", "", "", ""},
			{"Detail Information :", "", "", "", ""},
			{"Asset:", eq.Name, "", "KM:", FormatFloatDecimal2(header.Kilometers)},
			{"Start Breakdown:", FormatDate(header.StartDownTime), "", "Target Finish Time:", FormatDate(header.TargetFinishTime)},
			{"Description :", "", "", "", ""},
			{"", header.Description, "", ""},
		},
		"Footer": [][]string{},
	}

	preview.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"WR No:", header.ID},
			{"WR Date:", FormatDate(header.TrxDate)},
			{"Site:", header.Dimension.Get("Site")},
			{"Requestor:", requestor.Name},
			{"Source Type:", string(header.SourceType)},
			{"Departement:", cc.Label},
			{"Source ID:", header.SourceID},
			{"WR Type:", header.WorkRequestType},
			{"", ""},
			{"Detail Information :", ""},
			{"Asset:", eq.Name},
			{"KM:", FormatFloatDecimal2(header.Kilometers)},
			{"Start Breakdown:", FormatDate(header.StartDownTime)},
			{"Target Finish Time:", FormatDate(header.TargetFinishTime)},
			{"Description :", ""},
			{header.Description, ""},
		},
		Footer: [][]string{},
	}

	return &preview
}

func (p *WorkRequestPosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkRequest, lines []scmmodel.InventReceiveIssueLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	db, _ = p.opt.Db.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	header.Status = ficomodel.JournalStatusPosted
	err = db.Save(header)
	if err != nil {
		return res, err
	}

	/*
		TODO: create WO otomatis dari WR, kecuali ada tambahan dari mba Evilda kemarin
		intinya mba evi:
		- WR: ada tambahan expedisi, kalo expedisi, ga perlu jadi WO [PENDING Mba Evi]
	*/

	woID, _ := tenantcorelogic.GenerateIDFromNumSeq(p.ctx, "WorkOrder") // pakai cara ini karena bila pakai nats api, waktu masuk MWPreAssignSequenceNo(), coID nya selalu ga dapet
	// TODO: akan selalu contoh: "WO X04240070" jadi X karena ketika save, journal type belum dikasih dari sini

	payload := mfgmodel.WorkOrderPlan{
		ID:      woID,
		TrxDate: *header.TrxDate,
		TrxType: mfgmodel.JournalWorkOrderPlan,
		// Source:  mfgmodel.FromWorkRequest,
		Status: ficomodel.JournalStatusDraft,

		// WOName:              header.Name,
		RequestorName: header.Name,
		// RequestorDepartment:   header.Department,
		WorkRequestID:         header.ID,
		CompanyID:             p.opt.CompanyID,
		Asset:                 header.EquipmentNo,
		ExpectedCompletedDate: *header.TargetFinishTime,
		WRDate:                *header.TrxDate,
		// SourceType:          string(header.SourceType),
		// SourceID:            header.SourceID,
		// JournalTypeID: header.JournalTypeID,
		// PostingProfileID: header.PostingProfileID,
		// EquipmentNo:      header.EquipmentNo,
		StartDownTime:   header.StartDownTime,
		WRDescription:   header.Description,
		Kilometers:      header.Kilometers,
		Dimension:       header.Dimension,
		WorkRequestType: header.WorkRequestType,
		IsFirsttimeSave: false,
	}

	if header.SourceType == mfgmodel.WRSourceTypeSalesOrder {
		// payload.WOType = "PRODUKSI"
	}

	e := Config.EventHub.Publish(
		"/v1/mfg/workorderplan/insert",
		&payload,
		&payload,
		&kaos.PublishOpts{Headers: codekit.M{"CompanyID": p.opt.CompanyID, sebar.CtxJWTReferenceID: p.opt.UserID}},
	)
	fmt.Printf("journal id: %s | workorder/insert e: %s | result: %s\n", header.ID, e, codekit.JsonStringIndent(payload, "\t"))

	return res, err
}

func (p *WorkRequestPosting) Approved() error {
	return nil
}

func (p *WorkRequestPosting) Rejected() error {
	return nil
}

func (p *WorkRequestPosting) GetAccount() string {
	return ""
}

func (p *WorkRequestPosting) buildDummyLine() scmmodel.InventReceiveIssueLine {
	line := scmmodel.InventReceiveIssueLine{}
	item := new(tenantcoremodel.Item)
	p.opt.Db.Get(item)
	line.Item = *item
	line.ItemID = item.ID
	line.Qty = 10

	return line
}

func (p *WorkRequestPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", p.opt.UserID),
	))
	if err != nil && err != io.EOF {
		return fmt.Errorf("error when get email employee : %s", err.Error())
	}

	for _, app := range pa.Approvals {
		if app.Line == pa.CurrentStage {
			notification := ficomodel.Notification{
				UserSubmitter:            p.opt.UserID,
				UserSubmitterEmail:       employee.Email,
				JournalID:                p.header.ID,
				JournalType:              p.header.JournalTypeID,
				PostingProfileApprovalID: pa.ID,
				TrxDate:                  *p.header.TrxDate,
				TrxType:                  string(p.header.TrxType),
				Text:                     p.header.Text,
				UserTo:                   app.UserID,
				Menu:                     string(p.header.TrxType),
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
			}

			employee := new(tenantcoremodel.Employee)
			err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
				dbflex.Eq("_id", app.UserID),
			))
			if err != nil && err != io.EOF {
				return fmt.Errorf("error when get employee : %s", err.Error())
			}

			notification.UserToEmail = employee.Email

			err = p.opt.Db.Save(&notification)
			if err != nil {
				return fmt.Errorf("error when save notification : %s", err.Error())
			}
		}
	}

	return nil
}

func (p *WorkRequestPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
	// get latest notification
	latestNotif := new(ficomodel.Notification)
	err := p.opt.Db.GetByParm(latestNotif, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("JournalID", p.header.ID),
			dbflex.Eq("UserTo", p.opt.UserID),
		),
	).SetSort("-Created"))
	if err != nil {
		return fmt.Errorf("error when get notification : %s", err.Error())
	}

	if op == ficologic.PostOpApprove {
		latestNotif.Status = string(ficomodel.JournalStatusApproved)
	} else {
		latestNotif.Status = string(ficomodel.JournalStatusRejected)
	}

	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification : %s", err.Error())
	}

	// create notification for submitter user
	latestNotif.ID = primitive.NewObjectID().Hex()
	latestNotif.UserTo = latestNotif.UserSubmitter
	latestNotif.UserToEmail = latestNotif.UserSubmitterEmail
	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification for submitter user : %s", err.Error())
	}

	// get user approval stage
	userApprovals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.UserID == p.opt.UserID
	})

	approvals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.Line == userApprovals[0].Line && a.Status != "PENDING"
	})

	// check if need to send notification
	if len(approvals) >= pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line != pa.CurrentStage {
		for _, app := range pa.Approvals {
			if app.Line == pa.CurrentStage {
				latestNotif.ID = primitive.NewObjectID().Hex()
				latestNotif.UserTo = app.UserID
				latestNotif.Status = app.Status

				employee := new(tenantcoremodel.Employee)
				err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
					dbflex.Eq("_id", app.UserID),
				))
				if err != nil && err != io.EOF {
					return fmt.Errorf("error when get employee : %s", err.Error())
				}

				latestNotif.UserToEmail = employee.Email

				err = p.opt.Db.Save(latestNotif)
				if err != nil {
					return fmt.Errorf("error when save notification for next approval : %s", err.Error())
				}
			}
		}
	}

	return nil
}
