package scmlogic

import (
	"fmt"
	"io"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
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

type assetAcquisitionPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.AssetAcquisitionJournal
	trxType string

	inventTrxs []*scmmodel.InventTrx
	items      *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewAssetAcquisitionPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.AssetAcquisitionJournal, scmmodel.AssetItemTransfer] {
	p := new(assetAcquisitionPosting)
	p.ctx = ctx
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*scmmodel.AssetAcquisitionJournal, scmmodel.AssetItemTransfer](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *assetAcquisitionPosting) GetAccount() string {
	return p.header.TransferName
}

func (p *assetAcquisitionPosting) Header() (*scmmodel.AssetAcquisitionJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.AssetAcquisitionJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.AssetAcquisitionJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(jt.TrxType)
	if p.trxType == "" {
		p.trxType = "Asset Acquisition"
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

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *assetAcquisitionPosting) Lines() ([]scmmodel.AssetItemTransfer, error) {
	p.header.ItemTranfers = lo.Map(p.header.ItemTranfers, func(line scmmodel.AssetItemTransfer, index int) scmmodel.AssetItemTransfer {
		line.Dimension = tenantcorelogic.TernaryDimension(p.header.Dimension, line.Dimension)
		line.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.TransferFrom, &line.InventDim)

		if line.Text == "" {
			line.Text = p.header.Text
		}

		item, _ := p.items.Get(line.ItemID)
		line.Item = *item

		return line
	})

	return p.header.ItemTranfers, nil
}

func (p *assetAcquisitionPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.AssetAcquisitionJournal, lines []scmmodel.AssetItemTransfer) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	if err := p.Validate(); err != nil {
		return &preview, trxs, 0, err
	}

	inventTrxs := []orm.DataModel{}
	totalAmt := float64(0)
	for _, line := range p.header.ItemTranfers {
		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.TrxDate = p.header.TrxDate
		inventTrx.Item = line.Item
		inventTrx.SKU = line.SKU
		inventTrx.Text = line.Text

		inventTrx.Qty = lo.Ternary(line.Qty > 0, (-1 * line.Qty), line.Qty)
		inventTrx.TrxQty = line.Qty
		inventTrx.TrxUnitID = line.UnitID
		inventTrx.AmountPhysical = line.UnitCost * inventTrx.Qty
		inventTrx.Status = scmmodel.ItemConfirmed

		inventTrx.SourceType = scmmodel.ModuleInventory
		inventTrx.SourceJournalID = p.header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.trxType)

		inventTrx.InventDim = line.InventDim
		inventTrx.Dimension = line.Dimension

		p.inventTrxs = append(p.inventTrxs, inventTrx)
		inventTrxs = append(inventTrxs, inventTrx)
		totalAmt += inventTrx.AmountPhysical
	}

	trxs[inventTrxs[0].TableName()] = inventTrxs

	// TODO: set preview

	return &preview, trxs, totalAmt, nil
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *assetAcquisitionPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.AssetAcquisitionJournal, lines []scmmodel.AssetItemTransfer) []ficomodel.JournalLine {
	return assetLineToFicoLines(p.header.ItemTranfers)
}

// Post proses me-reserve
func (p *assetAcquisitionPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.AssetAcquisitionJournal, lines []scmmodel.AssetItemTransfer, trxs map[string][]orm.DataModel) (string, error) {
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

	res, err = ficologic.PostModelSave(db, header, "AssetAcquisitionVoucherNo", trxs)
	if err != nil {
		return res, err
	}

	// if _, err := NewInventBalanceCalc(db).Sync(p.inventTrxs); err != nil {
	// 	return res, fmt.Errorf("update balance: %s", err.Error())
	// }

	_, err = NewItemBalanceHub(db).Sync(nil, ItemBalanceOpt{
		CompanyID:       header.CompanyID,
		ConsiderSKU:     true,
		DisableGrouping: true,
		ItemIDs: lo.Map(p.inventTrxs, func(t *scmmodel.InventTrx, index int) string {
			return t.Item.ID
		}),
	})
	if err != nil {
		return res, fmt.Errorf("update balance: %s", err.Error())
	}

	/*
		Based on mba Evilda explanation:
		a. jika dari menu SCM untuk "Does Fixed Asset number is exist" di centang,
		maka data dari menu SCM akan mereplace data yg ada di menu Fixed Asset (sesuai Asset ID yg dipilih).
		Data apa saja yg akan ke-update :
			- Asset Name = dari nama Item
			- Asset Acquisition Date (Baik di tab-Reference maupun di tab-Depreciation) = dari tanggal posting date

		b. jika dari menu SCM untuk "Does Fixed Asset number is exist" tidak di centang,
		maka data dari menu SCM akan insert data baru menu Fixed Asset.
		Data apa saja yg akan ke-insert :
			- Asset ID = dari Asset ID yg muncul dari list Fixed Asset Number
			- Asset Name = dari nama Item
			- Asset Acquisition Date (Baik di tab-Reference maupun di tab-Depreciation) = dari tanggal posting date

		Convert to logic:
		- Update / Insert All asset (create new Asset if id retrieved from FixedAssetNumberList.ID): /v1/bagong/asset/acquire
		- Update FixedAssetNumberList.IsUsed with unchecked DoesFixedAssetNumberIsExist: /v1/fico/fixedassetnumberlist/use
	*/

	// update data in tenantcore and bagong asset
	itemM := lo.SliceToMap(header.ItemTranfers, func(item scmmodel.AssetItemTransfer) (string, tenantcoremodel.Item) {
		return item.Item.ID, item.Item
	})

	type AssetUpdateParam struct {
		ID      string
		GroupID string
		Name    string
	}

	assetParams := lo.Map(header.AssetRegisters, func(asset scmmodel.AssetRegister, index int) AssetUpdateParam {
		return AssetUpdateParam{ID: asset.AssetID, Name: itemM[asset.ItemID].Name, GroupID: asset.AssetGroup}
	})

	if len(assetParams) > 0 {
		payload := struct {
			AcquisitionDate time.Time
			Assets          []AssetUpdateParam
		}{
			AcquisitionDate: time.Now(),
			Assets:          assetParams,
		}

		e := scmconfig.Config.EventHub().Publish(
			"/v1/bagong/asset/acquire",
			&payload,
			nil,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": header.CompanyID, sebar.CtxJWTReferenceID: p.opt.UserID}},
		)
		fmt.Printf("journal id: %s | asset/acquire e: %s\n", header.ID, e)
	}

	// update IsUsed in FixedAssetNumberList
	fanIDs := lo.FilterMap(header.AssetRegisters, func(asset scmmodel.AssetRegister, index int) (string, bool) {
		return asset.AssetID, asset.DoesFixedAssetNumberIsExist == false // UI: tidak dicentang
	})

	if len(fanIDs) > 0 {
		payload := struct {
			IDs    []string
			IsUsed bool
		}{
			IDs:    fanIDs,
			IsUsed: true,
		}

		e := scmconfig.Config.EventHub().Publish(
			"/v1/fico/fixedassetnumberlist/use",
			&payload,
			nil,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": header.CompanyID, sebar.CtxJWTReferenceID: p.opt.UserID}},
		)
		fmt.Printf("journal id: %s | fixedassetnumberlist/use e: %s\n", header.ID, e)
	}

	return res, err
}

func (p *assetAcquisitionPosting) Validate() error {
	// AssetID dari AssetRegisters tidak boleh ada yang sama
	if duplicateAssetIDs := findDuplicateAssetIDs(p.header.AssetRegisters); len(duplicateAssetIDs) > 0 {
		return fmt.Errorf("duplicate AssetID found: %s", strings.Join(duplicateAssetIDs, ", "))
	}

	return nil
}

func assetLineToFicoLines(lines []scmmodel.AssetItemTransfer) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.AssetItemTransfer, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.UnitCost * line.Qty
		return *jl
	})
}

func findDuplicateAssetIDs(assets []scmmodel.AssetRegister) []string {
	countMap := make(map[string]int)

	// Menghitung berapa kali setiap AssetID muncul
	for _, asset := range assets {
		countMap[asset.AssetID]++
	}

	// Menyimpan nilai AssetID yang memiliki duplikat
	var duplicateAssetIDs []string
	for assetID, count := range countMap {
		if count > 1 {
			duplicateAssetIDs = append(duplicateAssetIDs, assetID)
		}
	}

	return duplicateAssetIDs
}

func (p *assetAcquisitionPosting) Approved() error {
	return nil
}

func (p *assetAcquisitionPosting) Rejected() error {
	return nil
}

func (p *assetAcquisitionPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				TrxType:                  "Asset Acquisition",
				Menu:                     "Asset Acquisition",
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

func (p *assetAcquisitionPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
