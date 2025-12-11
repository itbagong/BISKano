package scmlogic

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/golang-module/carbon/v2"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemRequestPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.ItemRequest
	trxType string

	inventTrxs []*scmmodel.InventTrx
	lines      []scmmodel.ItemRequestDetail
	items      *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewItemRequestPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.ItemRequest, scmmodel.ItemRequestDetail] {
	itemRequest := new(ItemRequestPosting)
	itemRequest.ctx = ctx
	itemRequest.opt = &opt
	itemRequest.items = sebar.NewMapRecordWithORM(itemRequest.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*scmmodel.ItemRequest, scmmodel.ItemRequestDetail](itemRequest)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *ItemRequestPosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *ItemRequestPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.ItemRequest, lines []scmmodel.ItemRequestDetail) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.ItemRequestDetail, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		jl.Account = ficomodel.SubledgerAccount{AccountType: scmmodel.ModuleInventory, AccountID: line.ItemID}
		return *jl
	})
}

func (p *ItemRequestPosting) Header() (*scmmodel.ItemRequest, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.ItemRequest), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: item request: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.ItemRequestJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(jt.TrxType)
	if p.trxType == "" {
		return nil, nil, fmt.Errorf("invalid transaction type")
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, jt.PostingProfileID)
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

func (p *ItemRequestPosting) Lines() ([]scmmodel.ItemRequestDetail, error) {
	itemDetails := []scmmodel.ItemRequestDetail{}
	e := p.opt.Db.GetsByFilter(new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", p.header.ID), &itemDetails)
	if e != nil {
		return itemDetails, e
	}

	itemIDs := make([]string, len(itemDetails))
	for i, l := range itemDetails {
		itemIDs[i] = l.ItemID
	}

	items := []tenantcoremodel.Item{}
	err := p.opt.Db.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return itemDetails, err
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, string) {
		return v.ID, string(v.ItemType)
	})

	e = sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		itemDetails = lo.Map(itemDetails, func(d scmmodel.ItemRequestDetail, index int) scmmodel.ItemRequestDetail {
			d.Dimension = p.header.Dimension
			d.ItemType = mapItem[d.ItemID]
			qtyFulfilled := lo.SumBy(d.DetailLines, func(detLine scmmodel.ItemRequestDetailLine) float64 {
				return detLine.QtyFulfilled
			})
			d.QtyFulfilled = qtyFulfilled

			tx.Update(&d, "Dimension", "ItemType", "QtyFulfilled")

			return d
		})

		return nil
	})

	if e != nil {
		return itemDetails, e
	}

	p.lines = itemDetails
	return itemDetails, nil
}

func (p *ItemRequestPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.ItemRequest, lines []scmmodel.ItemRequestDetail) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	trxs := map[string][]orm.DataModel{}

	if p.opt.Op == ficologic.PostOpSubmit {
		if e := p.validateOnSubmit(); e != nil {
			return nil, nil, 0, e
		}
	}

	preview := p.GetPreview(opt, header, lines)
	return preview, trxs, lo.SumBy(p.lines, func(l scmmodel.ItemRequestDetail) float64 {
		return l.QtyFulfilled
	}), nil
}

func (p *ItemRequestPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.ItemRequest, lines []scmmodel.ItemRequestDetail, trxs map[string][]orm.DataModel) (string, error) {
	var (
		// db  *datahub.Hub
		err error
		res string
	)

	err = p.validate()
	if err != nil {
		return res, err
	}

	err = sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		referenceID, err := new(ItemRequestDetailEngine).Fulfillment(p.ctx, tx, &FulfillmentRequest{
			ItemRequestID: header.ID,
			UserID:        p.opt.UserID,
		})

		errMsg := "Sukses"
		isSukses := true
		if err != nil {
			errMsg = err.Error()
			isSukses = false
		}

		//remove history log ir pr
		tx.DeleteByFilter(new(scmmodel.TemporaryIRPRAuditTrail), dbflex.Eq("IRCreatedDate", carbon.Now().SubDays(7)))

		//save audit trails for fullfilment
		fullfimentAudit := scmmodel.TemporaryIRPRAuditTrail{
			IRID:                   header.ID,
			ReferenceID:            referenceID,
			Error:                  errMsg,
			IRCreatedDate:          time.Now(),
			FullfilmentCreatedDate: time.Now(),
			IsSuccess:              isSukses,
		}

		err = tx.Save(&fullfimentAudit)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return res, err
	}

	if header.WOReff != "" {
		payload := struct {
			WOID   string
			Status string
		}{
			WOID:   header.WOReff,
			Status: "IN PROGRESS",
		}

		scmconfig.Config.EventHub().Publish("/v1/mfg/workorder/update-status", payload, nil, &kaos.PublishOpts{Headers: codekit.M{"CompanyID": header.CompanyID, sebar.CtxJWTReferenceID: p.opt.UserID}})
	}

	return res, nil
}

func (p *ItemRequestPosting) Approved() error {
	return nil
}

func (p *ItemRequestPosting) Rejected() error {
	return nil
}

func (p *ItemRequestPosting) validate() error {
	// fulfilment validation
	errsUnFulfilledItemNames := []string{}

	for _, line := range p.lines {
		if len(line.DetailLines) == 0 {
			item, _ := p.items.Get(line.ItemID)
			errsUnFulfilledItemNames = append(errsUnFulfilledItemNames, item.Name)
		}
	}

	if len(errsUnFulfilledItemNames) > 0 {
		return fmt.Errorf("please fill fulfilment and save, for item: %s", strings.Join(errsUnFulfilledItemNames, ", "))
	}

	return nil
}

func (p *ItemRequestPosting) validateOnSubmit() error {
	// validate Requestor when Site is not HO (SITE020)

	// kalo user yang login adalah HO, skip validation
	isUserLoggedInHO := false
	userLoggedIn := new(tenantcoremodel.Employee)
	p.opt.Db.GetByID(userLoggedIn, p.opt.UserID)
	if userLoggedIn.Dimension.Get("Site") == "SITE020" {
		isUserLoggedInHO = true
	}
	if lo.Contains(userLoggedIn.Sites, "SITE020") {
		isUserLoggedInHO = true
	}

	site := p.header.Dimension.Get("Site")
	if site != "" && site != "SITE020" && isUserLoggedInHO == false {
		if p.header.Requestor == "" {
			return fmt.Errorf("requestor is required")
		}

		// get available requestor list
		availableRequestors := []tenantcoremodel.Employee{}
		// p.opt.Db.GetsByFilter(new(tenantcoremodel.Employee), dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", site)), &availableRequestors)
		p.opt.Db.GetsByFilter(new(tenantcoremodel.Employee), dbflex.Eq("Sites", site), &availableRequestors) // fix: pake sites yang sebelah kanan, bisa multiple, bukan Site yang ada di Dimension

		// check if requestor is available
		isAllowed := lo.ContainsBy(availableRequestors, func(e tenantcoremodel.Employee) bool {
			return e.ID == p.header.Requestor
		})
		if !isAllowed {
			dmORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.DimensionMaster))
			siteObj, _ := dmORM.Get(site)
			_ = siteObj
			// return fmt.Errorf("requestor is not allowed in this site: %s", siteObj.Label)
		}
	}

	return nil
}

func (p *ItemRequestPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.ItemRequest, lines []scmmodel.ItemRequestDetail) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	empORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Employee))
	siteORM := sebar.NewMapRecordWithORM(opt.Db, new(bagongmodel.Site))

	requestor, _ := empORM.Get(header.Requestor)

	site, _ := siteORM.Get(header.Dimension.Get("Site"))

	signatureRequestor := tenantcoremodel.Signature{
		ID:        header.Requestor,
		Header:    "Pembuat",
		Footer:    requestor.Name,
		Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	}

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.ItemRequestType), header.ID)
	pv.Signature = append(pv.Signature, signatureRequestor)
	pv.Signature = append(pv.Signature, signature...)

	pv.Header = codekit.M{}.Set("Data", [][]string{
		{"No:", header.ID, "", "", "Requestor:", requestor.Name, ""},
		{"IR Name:", header.Name, "", "", "Priority:", header.Priority, ""},
		{"Site:", site.Name, "", "", "WO Reff:", header.WOReff, ""},
		{"Date:", FormatDate(&header.TrxDate), "", "", "", "", ""},
	}).Set("Footer", [][]string{
		{"Note:", "", "", ""},
		{header.Remarks, "", "", ""},
	})

	pv.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"No:", header.ID},
			{"Requestor:", requestor.Name, ""},
			{"IR Name:", header.Name},
			{"Priority:", header.Priority, ""},
			{"Site:", site.Name},
			{"WO Reff:", header.WOReff, ""},
			{"Date:", FormatDate(&header.TrxDate)},
		},
		Footer: [][]string{
			{"Note:"},
			{header.Remarks},
		},
	}

	sectionLine := tenantcoremodel.PreviewSection{
		HideTitle:   false,
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items: [][]string{
			{"No", "Part Number", "Part Description", "Qty", "Qty Fullfilled", "UoM", "Remarks"},
		},
	}

	skuIDs := make([]string, len(lines))
	itemIDs := make([]string, len(lines))
	for i, l := range p.lines {
		skuIDs[i] = l.SKU
		itemIDs[i] = l.ItemID
	}

	skus := []tenantcoremodel.ItemSpec{}
	err := p.opt.Db.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", skuIDs...),
	), &skus)
	if err != nil {
		return nil
	}

	mapSku := lo.Associate(skus, func(v tenantcoremodel.ItemSpec) (string, string) {
		return v.ID, v.SKU
	})

	items := []tenantcoremodel.Item{}
	err = p.opt.Db.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return nil
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, tenantcoremodel.Item) {
		return v.ID, v
	})

	mapAssignItem, err := AssignItem(p.opt.Db, itemIDs, skuIDs)
	if err != nil {
		return nil
	}

	lineCount := 1
	sectionLine.Items = append(sectionLine.Items, lo.Map(p.lines, func(line scmmodel.ItemRequestDetail, index int) []string {
		item := mapItem[line.ItemID]
		item.ID = mapAssignItem[line.ItemID+line.SKU]

		res := make([]string, 7)
		res[0] = strconv.Itoa(lineCount)
		res[1] = mapSku[line.SKU]
		res[2] = lo.Ternary(item.ID != "", item.ID, item.Name)
		res[3] = fmt.Sprintf("%.2f", line.QtyRequested)
		res[4] = fmt.Sprintf("%.2f", line.QtyFulfilled)
		res[5] = line.UoM
		res[6] = line.Remarks

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *ItemRequestPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				TrxDate:                  p.header.TrxDate,
				Text:                     p.header.Text,
				UserTo:                   app.UserID,
				TrxType:                  string(p.header.TrxType),
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

func (p *ItemRequestPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
