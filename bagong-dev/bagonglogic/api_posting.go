package bagonglogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongconfig"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
)

type PostingProfileHandler struct {
}

type FixByTypeRequest struct {
	Worker      int
	TenantID    string
	CompanyID   string
	JournalType string
	JournalIDs  []string
}

func (obj *PostingProfileHandler) PostRecalculate(ctx *kaos.Context, payload *FixByTypeRequest) (interface{}, error) {
	res := ResponseTrayekPosting{}
	if ctx == nil {
		return res, errors.New("missing: ctx")
	}

	// h := sebar.GetTenantDBFromContext(ctx)
	// if h == nil {
	// 	return res, errors.New("missing: connection")
	// }

	// coID := "DEMO00"

	// userID := "admin_admin"

	// call API Posting
	// postReq := ficologic.EventPostRequest{
	// 	CompanyID: coID,
	// 	UserID:    userID,
	// }

	// mapPost := []ficologic.PostRequest{}

	// custJournal := []ficomodel.CustomerJournal{}
	// if e := h.GetsByFilter(new(ficomodel.CustomerJournal), dbflex.Eq("Status", "DRAFT"), &custJournal); e != nil {
	// 	return nil, e
	// }

	evHub, _ := ctx.DefaultEvent()

	post := ficologic.EnvelopePost{}
	_ = evHub.Publish("/v1/fico/recalc/fix-by-type-ev", payload, &post, nil)

	return payload, nil
}

type UpdateAttachmentTags struct {
	JournalType string
	JournalID   string
	Tags        []string
	NewTags     []string
}

type AttachmentTagDetail struct {
	Module    string
	JournalID string
	LineID    string
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payload PostRequest) (interface{}, error) {
	res := ResponseTrayekPosting{}
	if ctx == nil {
		return res, errors.New("missing: ctx")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	// coID = "DEMO00"
	if coID == "DEMO" || coID == "" {
		return res, errors.New("missing: Company, please relogin")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	// userID = "admin_admin"
	if userID == "" {
		return res, errors.New("missing: User, please relogin")
	}

	reqOpt := ficologic.PostingHubCreateOpt{
		Db:        h,
		UserID:    userID,
		CompanyID: coID,
		ModuleID:  string(payload.JournalType),
		JournalID: payload.JournalID}

	switch payload.JournalType {
	case PostTypeSiteEntryTrayek:
		ritase, mapID, customerJournalID, vendorJournalID, err := SiteEntryTrayekJournal(ctx, reqOpt)
		if err != nil {
			// rollback journal
			payloadRollback := SiteEntryRollbackRequest{
				JournalType:       payload.JournalType,
				JournalID:         ritase.ID,
				VendorJournalID:   vendorJournalID,
				CustomerJournalID: customerJournalID,
				MapJournalID:      mapID,
			}
			SiteEntryRollbackJournal(h, payloadRollback)
			return res, err
		}
		res, err := SiteEntryTrayekPosting(ctx, *ritase, customerJournalID, vendorJournalID, payload, reqOpt)
		if err != nil {
			// rollback journal
			payloadRollback := SiteEntryRollbackRequest{
				JournalType:       payload.JournalType,
				JournalID:         ritase.ID,
				VendorJournalID:   vendorJournalID,
				CustomerJournalID: customerJournalID,
				MapJournalID:      mapID,
			}
			SiteEntryRollbackJournal(h, payloadRollback)
			return res, err
		}
		return res, nil
	case PostTypeSiteEntryBTS:
		bts := new(bagongmodel.SiteEntryBTSDetail)
		if e := h.GetByID(bts, payload.JournalID); e != nil {
			return res, fmt.Errorf("error when get site entry bts: %s", e.Error())
		}

		mapID := map[string]bool{}
		for _, d := range bts.Expense {
			if d.JournalID != "" {
				mapID[d.JournalID] = true
			}
		}

		vendorJournal, err := SiteEntryNonTrayek(ctx, reqOpt)
		if err != nil {
			// rollback journal
			for _, v := range vendorJournal {
				payloadRollback := SiteEntryRollbackRequest{
					JournalType:     PostType(reqOpt.ModuleID),
					JournalID:       reqOpt.JournalID,
					VendorJournalID: v.ID,
					MapJournalID:    mapID,
				}
				SiteEntryRollbackJournal(h, payloadRollback)
			}
			return res, err
		}

		for _, v := range vendorJournal {
			payload.Text = v.Text
			payload.JournalID = v.ID
			payload.JournalType = PostType(ficomodel.SubledgerVendor)
			res, err := SiteEntryPosting(ctx, payload, reqOpt)
			if err != nil {
				// rollback journal
				for _, vr := range vendorJournal {
					payloadRollback := SiteEntryRollbackRequest{
						JournalType:     PostType(reqOpt.ModuleID),
						JournalID:       reqOpt.JournalID,
						VendorJournalID: vr.ID,
						MapJournalID:    mapID,
					}
					SiteEntryRollbackJournal(h, payloadRollback)
				}
				return res, err
			}
		}

		return res, nil
	case PostTypeSiteEntryMining:
		mining := new(bagongmodel.SiteEntryMiningDetail)
		if e := h.GetByID(mining, payload.JournalID); e != nil {
			return res, fmt.Errorf("error when get site entry mining: %s", e.Error())
		}

		mapID := map[string]bool{}
		for _, d := range mining.Expense {
			if d.JournalID != "" {
				mapID[d.JournalID] = true
			}
		}

		vendorJournal, err := SiteEntryNonTrayek(ctx, reqOpt)
		if err != nil {
			// rollback journal
			for _, v := range vendorJournal {
				payloadRollback := SiteEntryRollbackRequest{
					JournalType:     PostType(reqOpt.ModuleID),
					JournalID:       reqOpt.JournalID,
					VendorJournalID: v.ID,
					MapJournalID:    mapID,
				}
				SiteEntryRollbackJournal(h, payloadRollback)
			}
			return res, err
		}

		for _, v := range vendorJournal {
			payload.Text = v.Text
			payload.JournalID = v.ID
			payload.JournalType = PostType(ficomodel.SubledgerVendor)
			res, err := SiteEntryPosting(ctx, payload, reqOpt)
			if err != nil {
				// rollback journal
				for _, vr := range vendorJournal {
					payloadRollback := SiteEntryRollbackRequest{
						JournalType:     PostType(reqOpt.ModuleID),
						JournalID:       reqOpt.JournalID,
						VendorJournalID: vr.ID,
						MapJournalID:    mapID,
					}
					SiteEntryRollbackJournal(h, payloadRollback)
				}
				return res, err
			}
		}
		return res, nil
	case PostTypeSiteEntryTourism:
		tourism := new(bagongmodel.SiteEntryTourismDetail)
		if e := h.GetByID(tourism, payload.JournalID); e != nil {
			return res, fmt.Errorf("error when get site entry tourism: %s", e.Error())
		}

		mapID := map[string]bool{}
		for _, d := range tourism.OtherExpense {
			if d.JournalID != "" {
				mapID[d.JournalID] = true
			}
		}

		for _, d := range tourism.OperationalExpense {
			if d.JournalID != "" {
				mapID[d.JournalID] = true
			}
		}

		vendorJournal, err := SiteEntryNonTrayek(ctx, reqOpt)
		if err != nil {
			// rollback journal
			for _, v := range vendorJournal {
				payloadRollback := SiteEntryRollbackRequest{
					JournalType:     PostType(reqOpt.ModuleID),
					JournalID:       reqOpt.JournalID,
					VendorJournalID: v.ID,
					MapJournalID:    mapID,
				}
				SiteEntryRollbackJournal(h, payloadRollback)
			}
			return res, err
		}

		for _, v := range vendorJournal {
			payload.Text = v.Text
			payload.JournalID = v.ID
			payload.JournalType = PostType(ficomodel.SubledgerVendor)
			res, err := SiteEntryPosting(ctx, payload, reqOpt)
			if err != nil {
				// rollback journal
				for _, vr := range vendorJournal {
					payloadRollback := SiteEntryRollbackRequest{
						JournalType:     PostType(reqOpt.ModuleID),
						JournalID:       reqOpt.JournalID,
						VendorJournalID: vr.ID,
						MapJournalID:    mapID,
					}
					SiteEntryRollbackJournal(h, payloadRollback)
				}
				return res, err
			}
		}

		return res, nil
	case PostTypeSiteEntryNonAsset:
		vendorJournal, mapID, err := SiteEntryTrayekNonAssetJournal(ctx, reqOpt)
		if err != nil {
			// rollback journal
			for _, v := range vendorJournal {
				payloadRollback := SiteEntryRollbackRequest{
					JournalType:     PostType(reqOpt.ModuleID),
					JournalID:       reqOpt.JournalID,
					VendorJournalID: v.ID,
					MapJournalID:    mapID,
				}
				SiteEntryRollbackJournal(h, payloadRollback)
			}
			return res, err
		}

		for _, v := range vendorJournal {
			payload.Text = v.Text
			payload.JournalID = v.ID
			payload.JournalType = PostType(ficomodel.SubledgerVendor)
			res, err := SiteEntryPosting(ctx, payload, reqOpt)
			if err != nil {
				// rollback journal
				for _, vr := range vendorJournal {
					payloadRollback := SiteEntryRollbackRequest{
						JournalType:     PostType(reqOpt.ModuleID),
						JournalID:       reqOpt.JournalID,
						VendorJournalID: vr.ID,
						MapJournalID:    mapID,
					}
					SiteEntryRollbackJournal(h, payloadRollback)
				}
				return res, err
			}
		}

		return res, nil
	case PostTypeAssetMovement:
		ev, _ := ctx.DefaultEvent()
		var postinger ficologic.JournalPosting
		postinger = NewAssetMovementPosting(h, ev, payload.JournalID, userID, coID)
		res, err := obj.PostSingle(ctx, &payload, postinger)
		if err != nil {
			fmt.Errorf("Error posting single: %s", err.Error())
		}

		return res, nil
	default:
		return res, fmt.Errorf("invalid module: %s", payload.JournalType)
	}
}

func (obj *PostingProfileHandler) PostSingle(ctx *kaos.Context, request *PostRequest, postingEngine ficologic.JournalPosting) (*tenantcoremodel.PreviewReport, error) {
	var res *tenantcoremodel.PreviewReport
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}
	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}
	if pr, err := PostJournal(postingEngine, userID, string(request.Op), request.Text); err != nil {
		return nil, err
	} else {
		res = pr
	}
	return res, nil
}

func SiteEntryTrayekNonAssetJournal(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) ([]ficomodel.VendorJournal, map[string]bool, error) {
	h := opt.Db
	vendorJournal := []ficomodel.VendorJournal{}
	totalVendorAmount := 0.0

	mapID := map[string]bool{}
	updateAttachmentTags := []UpdateAttachmentTags{}
	nonAsset := new(bagongmodel.SiteEntryNonAsset)
	if e := h.GetByID(nonAsset, opt.JournalID); e != nil {
		return vendorJournal, mapID, fmt.Errorf("nonAsset not found: %s", opt.JournalID)
	}

	for _, d := range nonAsset.ExpenseDetail {
		if d.JournalID != "" {
			mapID[d.JournalID] = true
		}
	}

	siteEntry := new(bagongmodel.SiteEntry)
	e := h.GetByID(siteEntry, nonAsset.ID)
	if e != nil {
		return vendorJournal, mapID, fmt.Errorf("siteEntry detail not found: %s", opt.JournalID)
	}

	trxDate := siteEntry.TrxDate.Format("2006-01-02")
	resTrxDate, _ := time.Parse("2006-01-02", trxDate)

	getJTVendor, err := GetDetailLedgerJournalPost(h, JournalPostReq{SiteID: siteEntry.SiteID, Type: "Expense"})
	if err != nil {
		return vendorJournal, mapID, fmt.Errorf("getJTVendor not found: %s", opt.JournalID)
	}

	jtVendor := new(ficomodel.VendorJournalType)
	bodyVendor, _ := json.Marshal(getJTVendor)
	json.Unmarshal(bodyVendor, &jtVendor)

	vendor := new(tenantcoremodel.Vendor)
	h.GetByFilter(vendor, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", siteEntry.SiteID), dbflex.Eq("GroupID", "vendorvirtual")))

	// get site
	site := new(bagongmodel.Site)
	e = h.GetByID(site, siteEntry.SiteID)
	if e != nil {
		return vendorJournal, mapID, fmt.Errorf("siteEntry not found: %s", opt.JournalID)
	}

	journalName := fmt.Sprintf("NON ASSET | %s | %s", site.Name, trxDate)
	if opt.ModuleID == string(PostTypeSiteEntryTrayek) {
		journalName = fmt.Sprintf("NonAsset | %s | %s | %s | %s", siteEntry.Purpose, site.Name, trxDate, opt.JournalID)
	}

	// mapping expense
	if len(nonAsset.ExpenseDetail) == 0 {
		return vendorJournal, mapID, fmt.Errorf("nonAsset expense empty: %s", opt.JournalID)
	}

	// mapping expense
	vendorJournalLines := []ficomodel.JournalLine{}
	for _, c := range nonAsset.ExpenseDetail {
		if c.JournalID == "" {
			tmp := ficomodel.JournalLine{
				LineNo:           c.LineNo,
				Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ExpenseTypeID},
				OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
				OffsetTransRefID: "",
				TagObjectID1:     ficomodel.SubledgerAccount{},
				TagObjectID2:     ficomodel.SubledgerAccount{},
				CurrencyID:       "IDR",
				LedgerDirection:  "",
				TrxType:          "Site Entry Non Asset Expense",
				PaymentType:      "",
				Qty:              c.Value,
				UnitID:           c.UnitID,
				PriceEach:        c.Amount,
				Amount:           c.TotalAmount,
				ApproveAmount:    0,
				Text:             c.Name + " | Notes: " + c.Notes,
				Critical:         false,
				Taxable:          false,
				TaxCodes:         []string{},
				Locked:           false,
				ChequeGiroID:     "",
				References:       []tenantcoremodel.ReferenceItem{},
				Dimension:        site.Dimension,
			}
			vendorJournalLines = append(vendorJournalLines, tmp)
		}
	}

	// get mapping pic with expense
	pics := ficologic.GetPostingSetupPICByJournalLine(h, jtVendor.PostingProfileID, site.Dimension, vendorJournalLines)
	if len(pics) < 1 {
		return nil, mapID, fmt.Errorf("error : not found GetPostingSetupPICByJournalLine: %s", opt.JournalID)
	}

	// get smallest priority, exclude other priority and set pic with journalLine
	postingPic := []ficomodel.PostingProfilePIC{}
	priority := 0
	for i, c := range pics {
		if i == 0 {
			priority = c.Priority
		}
		if c.Priority < priority {
			priority = c.Priority
		}
	}
	for _, c := range pics {
		if c.Priority == priority {
			postingPic = append(postingPic, *c)
		}
	}

	mapPic := map[string][]ficomodel.JournalLine{}
	for _, e := range vendorJournalLines {
		for _, c := range postingPic {
			if toolkit.HasMember(c.Account.AccountIDs, e.Account.AccountID) ||
				toolkit.IsNilOrEmpty(c.Account.AccountIDs) {
				mapPic[c.Name] = append(mapPic[c.Name], e)
				break
			}
		}
	}

	expJournal := map[string]string{}
	for idx, p := range mapPic {
		if len(vendorJournalLines) > 0 {
			totalVendorAmount = 0.0
			for _, c := range p {
				totalVendorAmount += c.Amount
			}

			vendorJournalTmp := ficomodel.VendorJournal{
				JournalTypeID:    jtVendor.ID,
				TransactionType:  "Site Entry Expense",
				VendorID:         vendor.ID,
				TrxDate:          resTrxDate,
				ExpectedDate:     &time.Time{},
				DeliveryDate:     &time.Time{},
				Text:             journalName + " | " + idx,
				CurrencyID:       "IDR",
				Status:           ficomodel.JournalStatusDraft,
				References:       []tenantcoremodel.ReferenceItem{},
				ChecklistTemp:    []tenantcoremodel.ChecklistItem{},
				Lines:            p,
				InvoiceID:        "",
				SubtotalAmount:   totalVendorAmount,
				TaxAmount:        0,
				ChargeAmount:     0,
				DiscountAmount:   0,
				Taxes:            []ficomodel.Charge{},
				Charges:          []ficomodel.Charge{},
				TotalAmount:      totalVendorAmount,
				ReportingAmount:  0,
				PaymentTermID:    "",
				PostingProfileID: jtVendor.PostingProfileID,
				CompanyID:        opt.CompanyID,
				Dimension:        site.Dimension,
				Created:          time.Now(),
				LastUpdate:       time.Now(),
			}

			tenantcorelogic.MWPreAssignSequenceNo("VendorJournalSiteEntry", false, "_id")(ctx, &vendorJournalTmp)
			if vendorJournalTmp.ID == "" {
				return nil, mapID, fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
			if e := h.GetByID(new(ficomodel.VendorJournal), vendorJournalTmp.ID); e == nil {
				return nil, mapID, fmt.Errorf("error duplicate sequence key: %s ", vendorJournalTmp.ID)
			}
			vendorJournalTmp.ID = "AP/" + site.Alias + vendorJournalTmp.ID

			if e := h.Save(&vendorJournalTmp); e != nil {
				return vendorJournal, mapID, fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
			// set expense with journalID
			for _, exp := range p {
				expJournal[exp.Account.AccountID] = vendorJournalTmp.ID
			}

			vendorJournal = append(vendorJournal, vendorJournalTmp)
		}
	}

	// update SiteEntry with JournalID
	exp := []bagongmodel.SiteExpense{}
	for _, c := range nonAsset.ExpenseDetail {
		tmp := c
		if c.JournalID == "" {
			if v, ok := expJournal[c.ExpenseTypeID]; ok {
				tmp.JournalID = v

				// set attachment
				vTag := "SE_NON_ASSET_EXPENSE_" + opt.JournalID + "_" + c.ID
				vNewTag := string(ficomodel.SubledgerVendor) + "_" + v + "_" + strconv.Itoa(c.LineNo)
				updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
					JournalType: "",
					JournalID:   "",
					Tags:        []string{vTag},
					NewTags:     []string{vNewTag},
				})
			}
		}
		exp = append(exp, tmp)
	}
	nonAsset.ExpenseDetail = exp

	if e := h.Save(nonAsset); e != nil {
		return vendorJournal, mapID, fmt.Errorf("error : save nonAsset: %s", opt.JournalID)
	}

	// update attachment tag
	if len(updateAttachmentTags) > 0 {
		evHub, _ := ctx.DefaultEvent()

		vRes := ""
		err := evHub.Publish("/v1/asset/update-tag-by-journal", updateAttachmentTags, &vRes, nil)
		if err != nil {
			return nil, mapID, fmt.Errorf("error : update attachment tags: %s", err.Error())
		}
	}

	return vendorJournal, mapID, nil
}

func Truncate(t time.Time) time.Time {
	return t.Truncate(23 * time.Hour)
}

func SiteEntryTrayekJournal(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) (*bagongmodel.SiteEntryTrayekRitase, map[string]bool, string, string, error) {
	h := opt.Db

	mapID := map[string]bool{}
	updateAttachmentTags := []UpdateAttachmentTags{}
	ritase := new(bagongmodel.SiteEntryTrayekRitase)
	if e := h.GetByID(ritase, opt.JournalID); e != nil {
		return nil, mapID, "", "", fmt.Errorf("ritase not found: %s", ritase.ID)
	}

	// get journal id that is not empty
	// for rollback purpose, so it will only reset data with empty journal id
	for _, c := range ritase.FixExpense {
		if c.JournalID != "" {
			mapID[c.JournalID] = true
		}
	}
	for _, c := range ritase.OtherExpense {
		if c.JournalID != "" {
			mapID[c.JournalID] = true
		}
	}
	for _, c := range ritase.OtherIncome {
		if c.JournalID != "" {
			mapID[c.JournalID] = true
		}
	}

	detailRitase := new(bagongmodel.SiteEntryTrayekDetail)
	e := h.GetByID(detailRitase, ritase.SiteEntryAssetID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("ritase detail not found: %s", ritase.ID)
	}

	siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	e = h.GetByID(siteEntryAsset, detailRitase.ID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("siteEntryAsset detail not found: %s", detailRitase.ID)
	}

	siteEntry := new(bagongmodel.SiteEntry)
	e = h.GetByID(siteEntry, siteEntryAsset.SiteEntryID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("siteEntry detail not found: %s", siteEntryAsset.SiteEntryID)
	}

	trxDate := siteEntry.TrxDate.Format("2006-01-02")
	resTrxDate, _ := time.Parse("2006-01-02", trxDate)

	// get site
	site := new(bagongmodel.Site)
	e = h.GetByID(site, siteEntry.SiteID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("site not found: %s", siteEntryAsset.SiteEntryID)
	}

	// get customer asset
	asset := new(bagongmodel.Asset)
	e = h.GetByID(asset, siteEntryAsset.AssetID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("asset not found: %s", siteEntryAsset.AssetID)
	}

	// get trayek
	trayek := new(bagongmodel.Trayek)
	e = h.GetByID(trayek, asset.DetailUnit.TrayekID)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("trayek not found: %s - %s", siteEntryAsset.AssetID, asset.DetailUnit.TrayekID)
	}
	driverId := tenantcorelogic.TernaryString(detailRitase.DriverID, detailRitase.DriverID2)
	// get driver
	driver := new(tenantcoremodel.Employee)
	e = h.GetByID(driver, driverId)
	if e != nil {
		return ritase, mapID, "", "", fmt.Errorf("driver not found: %s", driverId)
	}

	userInfo := bagongmodel.UserInfo{}
	if len(asset.UserInfo) > 0 {
		userInfo = asset.UserInfo[0]
		for _, c := range asset.UserInfo {
			if c.AssetDateTo.After(userInfo.AssetDateTo) {
				userInfo = c
			}
		}
	}

	if userInfo.CustomerID == "" {
		return ritase, mapID, "", "", fmt.Errorf("get user info from asset not found: %s", opt.JournalID)
	}

	siteID := detailRitase.Dimension.Get("Site")

	// get journal type
	getJTCustomer, err := GetDetailLedgerJournalPost(h, JournalPostReq{SiteID: siteID, Type: "Revenue"})
	if err != nil {
		return ritase, mapID, "", "", fmt.Errorf("getJTCustomer not found: %s", siteID)
	}

	jtCustomer := new(ficomodel.CustomerJournalType)
	bodyCustomer, _ := json.Marshal(getJTCustomer)
	json.Unmarshal(bodyCustomer, &jtCustomer)

	getJTVendor, err := GetDetailLedgerJournalPost(h, JournalPostReq{SiteID: siteID, Type: "Expense"})
	if err != nil {
		return ritase, mapID, "", "", fmt.Errorf("getJTVendor not found: %s", siteID)
	}

	jtVendor := new(ficomodel.VendorJournalType)
	bodyVendor, _ := json.Marshal(getJTVendor)
	json.Unmarshal(bodyVendor, &jtVendor)

	vendor := new(tenantcoremodel.Vendor)
	h.GetByFilter(vendor, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", siteID), dbflex.Eq("GroupID", "vendorvirtual")))

	journalName := fmt.Sprintf("Trayek %s | %s | %s | %s | %s | %s", ritase.TrayekName, ritase.RevenueType, trxDate, asset.DetailUnit.PoliceNum, driver.Name, ritase.ID)
	if ritase.RevenueType == "Premi" {
		journalName = fmt.Sprintf("Trayek %s | %s | %s | %s | %s | %s | %s", ritase.TrayekName, ritase.RevenueType, trxDate, asset.DetailUnit.PoliceNum, driver.Name, "Ritase "+fmt.Sprint(ritase.RitasePremi[0])+" & "+fmt.Sprint(ritase.RitasePremi[1]), ritase.ID)
	}

	// mapping income
	customerJournalLines := []ficomodel.JournalLine{}
	incomePassenger := ficomodel.JournalLine{}
	incomeDeposit := ficomodel.JournalLine{}
	totalCustomerAmount := 0.0
	totalVendorAmount := 0.0

	for _, c := range ritase.OtherIncome {
		if c.JournalID == "" {
			tmp := ficomodel.JournalLine{
				LineNo:           c.LineNo,
				Account:          ficomodel.SubledgerAccount{AccountType: jtCustomer.DefaultOffset.AccountType, AccountID: jtCustomer.DefaultOffset.AccountID},
				OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerCustomer, AccountID: userInfo.CustomerID},
				OffsetTransRefID: "",
				TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
				TagObjectID2:     ficomodel.SubledgerAccount{},
				CurrencyID:       "IDR",
				LedgerDirection:  "",
				TrxType:          "Site Entry Revenue",
				PaymentType:      "",
				Qty:              1,
				UnitID:           "Each",
				PriceEach:        c.Amount,
				Amount:           c.Amount,
				ApproveAmount:    0,
				Text:             c.Name,
				Critical:         false,
				Taxable:          false,
				TaxCodes:         []string{},
				Locked:           false,
				ChequeGiroID:     "",
				References:       []tenantcoremodel.ReferenceItem{},
				Dimension:        detailRitase.Dimension,
			}
			customerJournalLines = append(customerJournalLines, tmp)
			totalCustomerAmount += c.Amount
		}
	}

	if ritase.RevenueType == "Premi" {
		if ritase.PassengerIncome.JournalID == "" {
			// set income passenger
			incomePassenger = ficomodel.JournalLine{
				LineNo:           0,
				Account:          ficomodel.SubledgerAccount{AccountType: jtCustomer.DefaultOffset.AccountType, AccountID: jtCustomer.DefaultOffset.AccountID},
				OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerCustomer, AccountID: userInfo.CustomerID},
				OffsetTransRefID: "",
				TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
				TagObjectID2:     ficomodel.SubledgerAccount{},
				CurrencyID:       "IDR",
				LedgerDirection:  "",
				TrxType:          "Site Entry Revenue",
				PaymentType:      "",
				Qty:              1,
				UnitID:           "Each",
				PriceEach:        ritase.RitaseSummary.TotalRitaseIncome,
				Amount:           ritase.RitaseSummary.TotalRitaseIncome,
				ApproveAmount:    0,
				Text:             journalName,
				Critical:         false,
				Taxable:          false,
				TaxCodes:         []string{},
				Locked:           false,
				ChequeGiroID:     "",
				References:       []tenantcoremodel.ReferenceItem{},
				Dimension:        detailRitase.Dimension,
			}
			customerJournalLines = append(customerJournalLines, incomePassenger)
			totalCustomerAmount += ritase.RitaseSummary.TotalRitaseIncome
		}

	} else {
		var targetSetoran float64
		if ritase.CategoryDeposit == "Flat" {
			targetSetoran = ritase.ConfigDeposit.TargetFlat
		} else if ritase.CategoryDeposit == "Non Flat" {
			targetSetoran = ritase.ConfigDeposit.TargetNonFlat
		} else {
			targetSetoran = ritase.ConfigDeposit.TargetRent
		}
		if ritase.DepositIncome.JournalID == "" {
			incomeDeposit = ficomodel.JournalLine{
				LineNo:           0,
				Account:          ficomodel.SubledgerAccount{AccountType: jtCustomer.DefaultOffset.AccountType, AccountID: jtCustomer.DefaultOffset.AccountID},
				OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerCustomer, AccountID: userInfo.CustomerID},
				OffsetTransRefID: "",
				TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
				TagObjectID2:     ficomodel.SubledgerAccount{},
				CurrencyID:       "IDR",
				LedgerDirection:  "",
				TrxType:          "Site Entry Revenue",
				PaymentType:      "",
				Qty:              1,
				UnitID:           "Each",
				PriceEach:        targetSetoran,
				Amount:           targetSetoran,
				ApproveAmount:    0,
				Text:             journalName + " | " + ritase.CategoryDeposit,
				Critical:         false,
				Taxable:          false,
				TaxCodes:         []string{},
				Locked:           false,
				ChequeGiroID:     "",
				References:       []tenantcoremodel.ReferenceItem{},
				Dimension:        detailRitase.Dimension,
			}
			customerJournalLines = append(customerJournalLines, incomeDeposit)
			totalCustomerAmount += targetSetoran
		}
	}

	customerJournal := ficomodel.CustomerJournal{
		JournalTypeID:    jtCustomer.ID,
		TransactionType:  "Site Entry Income",
		CustomerID:       userInfo.CustomerID,
		TrxDate:          resTrxDate,
		DefaultOffset:    ficomodel.SubledgerAccount{AccountType: jtCustomer.DefaultOffset.AccountType, AccountID: jtCustomer.DefaultOffset.AccountID},
		CashPayment:      false,
		InclusiveTax:     false,
		CashBankID:       "",
		Text:             journalName,
		CurrencyID:       "IDR",
		Status:           "DRAFT",
		References:       []tenantcoremodel.ReferenceItem{},
		ChecklistTemp:    []tenantcoremodel.ChecklistItem{},
		Lines:            customerJournalLines,
		InvoiceID:        "",
		SubtotalAmount:   totalCustomerAmount,
		TaxAmount:        0,
		ChargeAmount:     0,
		DiscountAmount:   0,
		TaxCodes:         []string{},
		Taxes:            []ficomodel.Charge{},
		Charges:          []ficomodel.Charge{},
		TotalAmount:      totalCustomerAmount,
		ReportingAmount:  0,
		PaymentTermID:    "",
		PostingProfileID: jtCustomer.PostingProfileID,
		AddressAndTax:    ficomodel.AddressAndTax{},
		Errors:           "",
		CompanyID:        opt.CompanyID,
		Dimension:        detailRitase.Dimension,
		Created:          time.Now(),
		LastUpdate:       time.Now(),
	}

	if len(customerJournalLines) > 0 {
		tenantcorelogic.MWPreAssignCustomSequenceNo("CustomerJournalInvoiceTrayek")(ctx, &customerJournal)
		if customerJournal.ID == "" {
			return nil, mapID, "", "", fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
		}
		if e := h.GetByID(new(ficomodel.CustomerJournal), customerJournal.ID); e == nil {
			return nil, mapID, "", "", fmt.Errorf("error duplicate sequence key: %s ", customerJournal.ID)
		}
		if e := h.Save(&customerJournal); e != nil {
			return ritase, mapID, "", "", fmt.Errorf("error : save customerJournal: %s", opt.JournalID)
		}
	}

	oi := []bagongmodel.SiteIncome{}
	for _, c := range ritase.OtherIncome {
		if c.JournalID == "" {
			tmp := c
			tmp.JournalID = customerJournal.ID
			oi = append(oi, tmp)

			// set attachment
			vTag := "SE_TRAYEK_RITASE_OTHER_INCOME_" + opt.JournalID + "_" + c.ID
			vNewTag := string(ficomodel.SubledgerCustomer) + "_" + customerJournal.ID + "_" + strconv.Itoa(c.LineNo)
			updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
				JournalType: "",
				JournalID:   "",
				Tags:        []string{vTag},
				NewTags:     []string{vNewTag},
			})
		}
	}
	ritase.OtherIncome = oi

	vendorJournal := ficomodel.VendorJournal{}
	if ritase.RevenueType == "Premi" {
		// update journalID passengerIncome
		if ritase.PassengerIncome.JournalID == "" {
			ritase.PassengerIncome = bagongmodel.SiteIncome{
				Name:      incomePassenger.Text,
				JournalID: customerJournal.ID,
				Amount:    incomePassenger.Amount,
			}
		}

		// mapping expense
		vendorJournalLines := []ficomodel.JournalLine{}
		for _, c := range ritase.FixExpense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          "Site Entry Expense",
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           "Each",
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailRitase.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
				totalVendorAmount += c.Amount
			}
		}

		for _, c := range ritase.OtherExpense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          "Site Entry Expense",
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           "Each",
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailRitase.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
				totalVendorAmount += c.TotalAmount
			}
		}

		// bonus
		tmp := ficomodel.JournalLine{
			LineNo:           len(vendorJournalLines) + 1,
			Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: trayek.ExpenseTypeID},
			OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
			OffsetTransRefID: "",
			TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
			TagObjectID2:     ficomodel.SubledgerAccount{},
			CurrencyID:       "IDR",
			LedgerDirection:  "",
			TrxType:          "Site Entry Expense",
			PaymentType:      "",
			Qty:              1,
			UnitID:           "Each",
			PriceEach:        ritase.RitaseSummary.TotalBonus,
			Amount:           ritase.RitaseSummary.TotalBonus,
			ApproveAmount:    0,
			Text:             "BONUS",
			Critical:         false,
			Taxable:          false,
			TaxCodes:         []string{},
			Locked:           false,
			ChequeGiroID:     "",
			References:       []tenantcoremodel.ReferenceItem{},
			Dimension:        detailRitase.Dimension,
		}
		vendorJournalLines = append(vendorJournalLines, tmp)
		totalVendorAmount += ritase.RitaseSummary.TotalBonus

		vendorJournal = ficomodel.VendorJournal{
			JournalTypeID:    jtVendor.ID,
			TransactionType:  "Site Entry Expense",
			VendorID:         vendor.ID,
			TrxDate:          resTrxDate,
			ExpectedDate:     &time.Time{},
			DeliveryDate:     &time.Time{},
			Text:             journalName,
			CurrencyID:       "IDR",
			Status:           ficomodel.JournalStatusDraft,
			References:       []tenantcoremodel.ReferenceItem{},
			ChecklistTemp:    []tenantcoremodel.ChecklistItem{},
			Lines:            vendorJournalLines,
			InvoiceID:        "",
			SubtotalAmount:   totalVendorAmount,
			TaxAmount:        0,
			ChargeAmount:     0,
			DiscountAmount:   0,
			Taxes:            []ficomodel.Charge{},
			Charges:          []ficomodel.Charge{},
			TotalAmount:      totalVendorAmount,
			ReportingAmount:  0,
			PaymentTermID:    "",
			PostingProfileID: jtVendor.PostingProfileID,
			CompanyID:        opt.CompanyID,
			Dimension:        detailRitase.Dimension,
			Created:          time.Now(),
			LastUpdate:       time.Now(),
		}

		if len(vendorJournalLines) > 0 {
			tenantcorelogic.MWPreAssignCustomSequenceNo("VendorJournalSiteEntryTrayek")(ctx, &vendorJournal)
			if vendorJournal.ID == "" {
				return nil, mapID, "", "", fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
			if e := h.GetByID(new(ficomodel.VendorJournal), vendorJournal.ID); e == nil {
				return nil, mapID, "", "", fmt.Errorf("error duplicate sequence key: %s ", vendorJournal.ID)
			}
			if e := h.Save(&vendorJournal); e != nil {
				return nil, mapID, "", "", fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
		}

		// update journalID
		fe := []bagongmodel.SiteExpense{}
		for _, c := range ritase.FixExpense {
			tmp := c
			if c.JournalID == "" {
				tmp.JournalID = vendorJournal.ID

				// set attachment
				vTag := "SE_TRAYEK_RITASE_FIX_EXPENSE_" + opt.JournalID + "_" + c.ID
				vNewTag := string(ficomodel.SubledgerVendor) + "_" + vendorJournal.ID + "_" + strconv.Itoa(c.LineNo)
				updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
					JournalType: "",
					JournalID:   "",
					Tags:        []string{vTag},
					NewTags:     []string{vNewTag},
				})
			}
			fe = append(fe, tmp)
		}
		ritase.FixExpense = fe

		oe := []bagongmodel.SiteExpense{}
		for _, c := range ritase.OtherExpense {
			tmp := c
			if c.JournalID == "" {
				tmp.JournalID = vendorJournal.ID

				// set attachment
				vTag := "SE_TRAYEK_RITASE_OTHER_EXPENSE_" + opt.JournalID + "_" + c.ID
				vNewTag := string(ficomodel.SubledgerVendor) + "_" + vendorJournal.ID + "_" + strconv.Itoa(c.LineNo)
				updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
					JournalType: "",
					JournalID:   "",
					Tags:        []string{vTag},
					NewTags:     []string{vNewTag},
				})
			}
			oe = append(oe, tmp)
		}
		ritase.OtherExpense = oe
	} else {
		if ritase.DepositIncome.JournalID == "" {
			ritase.DepositIncome = bagongmodel.SiteIncome{
				Name:      incomeDeposit.Text,
				JournalID: customerJournal.ID,
				Amount:    incomeDeposit.Amount,
			}
		}
	}

	// update attachment tag
	// update tag tab attachment
	vTag := "SE_TRAYEK_" + ritase.SiteEntryAssetID
	vNewTag := string(ficomodel.SubledgerVendor) + "_" + string(ficomodel.SubledgerVendor) + " " + vendorJournal.ID
	updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
		JournalType: "SE_TRAYEK",
		JournalID:   ritase.SiteEntryAssetID,
		Tags:        []string{vTag},
		NewTags:     []string{vNewTag},
	})
	if len(updateAttachmentTags) > 0 {
		evHub, _ := ctx.DefaultEvent()

		vRes := ""
		err := evHub.Publish("/v1/asset/update-tag-by-journal", updateAttachmentTags, &vRes, nil)
		if err != nil {
			return nil, mapID, "", "", fmt.Errorf("error : update attachment tags: %s", err.Error())
		}
	}

	if e := h.Save(ritase); e != nil {
		return ritase, mapID, "", "", fmt.Errorf("error : save ritase: %s", ritase.ID)
	}

	if e := SumTrayek(h, detailRitase); e != nil {
		return ritase, mapID, "", "", fmt.Errorf("calculate summary Trayek error: %s", detailRitase.ID)
	}

	return ritase, mapID, customerJournal.ID, vendorJournal.ID, nil
}

func SiteEntryNonTrayek(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) ([]ficomodel.VendorJournal, error) {
	h := opt.Db
	vendorJournal := ficomodel.VendorJournal{}

	var siteEntryDetailId, siteId, driverId, journalName string
	dimension := tenantcoremodel.Dimension{}
	totalVendorAmount := 0.0

	updateAttachmentTags := []UpdateAttachmentTags{}
	detailBTS := new(bagongmodel.SiteEntryBTSDetail)
	detailTourism := new(bagongmodel.SiteEntryTourismDetail)
	detailMining := new(bagongmodel.SiteEntryMiningDetail)

	if opt.ModuleID == string(PostTypeSiteEntryBTS) {
		if e := h.GetByID(detailBTS, opt.JournalID); e != nil {
			return nil, fmt.Errorf("detailBTS not found: %s", opt.JournalID)
		}
		siteEntryDetailId = detailBTS.ID
		dimension = detailBTS.Dimension
	} else if opt.ModuleID == string(PostTypeSiteEntryMining) {
		if e := h.GetByID(detailMining, opt.JournalID); e != nil {
			return nil, fmt.Errorf("detailMining not found: %s", opt.JournalID)
		}
		siteEntryDetailId = detailMining.ID
		dimension = detailMining.Dimension
		driverId = tenantcorelogic.TernaryString(detailMining.DriverID, detailMining.DriverID2)
	} else if opt.ModuleID == string(PostTypeSiteEntryTourism) {
		if e := h.GetByID(detailTourism, opt.JournalID); e != nil {
			return nil, fmt.Errorf("detailTourism not found: %s", opt.JournalID)
		}
		siteEntryDetailId = detailTourism.ID
		dimension = detailTourism.Dimension
		driverId = tenantcorelogic.TernaryString(detailTourism.DriverID, detailTourism.DriverID2)
	}

	siteId = dimension.Get("Site")

	getJTVendor, err := GetDetailLedgerJournalPost(h, JournalPostReq{SiteID: siteId, Type: "Expense"})
	if err != nil {
		return nil, fmt.Errorf("getJTVendor not found: %s", opt.JournalID)
	}

	jtVendor := new(ficomodel.VendorJournalType)
	bodyVendor, _ := json.Marshal(getJTVendor)
	json.Unmarshal(bodyVendor, &jtVendor)

	vendor := new(tenantcoremodel.Vendor)
	h.GetByFilter(vendor, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", siteId), dbflex.Eq("GroupID", "vendorvirtual")))

	siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	e := h.GetByID(siteEntryAsset, siteEntryDetailId)
	if e != nil {
		return nil, fmt.Errorf("siteEntryAsset detail not found: %s", opt.JournalID)
	}

	siteEntry := new(bagongmodel.SiteEntry)
	e = h.GetByID(siteEntry, siteEntryAsset.SiteEntryID)
	if e != nil {
		return nil, fmt.Errorf("siteEntry not found: %s", opt.JournalID)
	}

	trxDate := siteEntry.TrxDate.Format("2006-01-02")
	resTrxDate, _ := time.Parse("2006-01-02", trxDate)

	// get customer asset
	asset := new(bagongmodel.Asset)
	e = h.GetByID(asset, siteEntryAsset.AssetID)
	if e != nil {
		return nil, fmt.Errorf("asset not found: %s", opt.JournalID)
	}

	// get site
	site := new(bagongmodel.Site)
	e = h.GetByID(site, siteEntry.SiteID)
	if e != nil {
		return nil, fmt.Errorf("siteEntry not found: %s", opt.JournalID)
	}

	driver := new(tenantcoremodel.Employee)
	if opt.ModuleID == string(PostTypeSiteEntryBTS) {
		// journalName = fmt.Sprintf("%s %s | %s | %s | %s", siteEntry.Purpose, site.Name, trxDate, asset.DetailUnit.PoliceNum, opt.JournalID)
	} else {
		e = h.GetByID(driver, driverId)
		if e != nil {
			return nil, fmt.Errorf("driver not found: %s", driverId)
		}
		// journalName = fmt.Sprintf("%s %s | %s | %s | %s | %s", siteEntry.Purpose, site.Name, trxDate, asset.DetailUnit.PoliceNum, driver.Name, opt.JournalID)
	}

	journalName = fmt.Sprintf("ASSET | %s | %s | %s", site.Name, trxDate, asset.DetailUnit.PoliceNum)

	// mapping expense
	vendorJournalLines := []ficomodel.JournalLine{}
	if opt.ModuleID == string(PostTypeSiteEntryBTS) {
		if len(detailBTS.Expense) == 0 {
			return nil, fmt.Errorf("detailBTS expense empty: %s", opt.JournalID)
		}
		for _, c := range detailBTS.Expense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ExpenseTypeID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          opt.ModuleID,
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           "Each",
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailBTS.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
			}
		}
	} else if opt.ModuleID == string(PostTypeSiteEntryMining) {
		if len(detailMining.Expense) == 0 {
			return nil, fmt.Errorf("detailMining expense empty: %s", opt.JournalID)
		}
		for _, c := range detailMining.Expense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ExpenseTypeID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          opt.ModuleID,
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           c.UnitID,
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailMining.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
			}
		}
	} else if opt.ModuleID == string(PostTypeSiteEntryTourism) {
		if len(detailTourism.OperationalExpense) == 0 && len(detailTourism.OtherExpense) == 0 {
			return nil, fmt.Errorf("detailTourism expense empty: %s", opt.JournalID)
		}
		for _, c := range detailTourism.OperationalExpense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ExpenseTypeID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          opt.ModuleID,
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           "Each",
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailTourism.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
			}
		}
		for _, c := range detailTourism.OtherExpense {
			if c.JournalID == "" {
				tmp := ficomodel.JournalLine{
					LineNo:           c.LineNo,
					Account:          ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: c.ExpenseTypeID},
					OffsetAccount:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerVendor, AccountID: vendor.ID},
					OffsetTransRefID: "",
					TagObjectID1:     ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAsset, AccountID: siteEntryAsset.AssetID},
					TagObjectID2:     ficomodel.SubledgerAccount{},
					CurrencyID:       "IDR",
					LedgerDirection:  "",
					TrxType:          opt.ModuleID,
					PaymentType:      "",
					Qty:              c.Value,
					UnitID:           "Each",
					PriceEach:        c.Amount,
					Amount:           c.TotalAmount,
					ApproveAmount:    0,
					Text:             c.Name + " | Notes: " + c.Notes,
					Critical:         false,
					Taxable:          false,
					TaxCodes:         []string{},
					Locked:           false,
					ChequeGiroID:     "",
					References:       []tenantcoremodel.ReferenceItem{},
					Dimension:        detailTourism.Dimension,
				}
				vendorJournalLines = append(vendorJournalLines, tmp)
			}
		}
	}

	// get mapping pic with expense
	pics := ficologic.GetPostingSetupPICByJournalLine(h, jtVendor.PostingProfileID, dimension, vendorJournalLines)
	if len(pics) < 1 {
		return nil, fmt.Errorf("error : not found GetPostingSetupPICByJournalLine: %s", opt.JournalID)
	}

	// get smallest priority, exclude other priority and set pic with journalLine
	postingPic := []ficomodel.PostingProfilePIC{}
	priority := 0
	for i, c := range pics {
		if i == 0 {
			priority = c.Priority
		}
		if c.Priority < priority {
			priority = c.Priority
		}
	}
	for _, c := range pics {
		if c.Priority == priority {
			postingPic = append(postingPic, *c)
		}
	}

	mapPic := map[string][]ficomodel.JournalLine{}
	for _, e := range vendorJournalLines {
		for _, c := range postingPic {
			if toolkit.HasMember(c.Account.AccountIDs, e.Account.AccountID) ||
				toolkit.IsNilOrEmpty(c.Account.AccountIDs) {
				mapPic[c.Name] = append(mapPic[c.Name], e)
				break
			}
		}
	}

	// set journal lines
	expJournal := map[string]string{}
	expJournalUpload := map[string]string{}
	vendorJournals := []ficomodel.VendorJournal{}
	for idx, p := range mapPic {
		if len(vendorJournalLines) > 0 {
			totalVendorAmount = 0.0
			for _, c := range p {
				totalVendorAmount += c.Amount
			}
			vendorJournal = ficomodel.VendorJournal{
				JournalTypeID:    jtVendor.ID,
				TransactionType:  "Site Entry Expense",
				VendorID:         vendor.ID,
				TrxDate:          resTrxDate,
				ExpectedDate:     &time.Time{},
				DeliveryDate:     &time.Time{},
				Text:             journalName + " | " + idx,
				CurrencyID:       "IDR",
				Status:           ficomodel.JournalStatusDraft,
				References:       []tenantcoremodel.ReferenceItem{},
				ChecklistTemp:    []tenantcoremodel.ChecklistItem{},
				Lines:            p,
				InvoiceID:        "",
				SubtotalAmount:   totalVendorAmount,
				TaxAmount:        0,
				ChargeAmount:     0,
				DiscountAmount:   0,
				Taxes:            []ficomodel.Charge{},
				Charges:          []ficomodel.Charge{},
				TotalAmount:      totalVendorAmount,
				ReportingAmount:  0,
				PaymentTermID:    "",
				PostingProfileID: jtVendor.PostingProfileID,
				CompanyID:        opt.CompanyID,
				Dimension:        dimension,
				Created:          time.Now(),
				LastUpdate:       time.Now(),
			}

			tenantcorelogic.MWPreAssignSequenceNo("VendorJournalSiteEntry", false, "_id")(ctx, &vendorJournal)
			if vendorJournal.ID == "" {
				return nil, fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
			if e := h.GetByID(new(ficomodel.VendorJournal), vendorJournal.ID); e == nil {
				return nil, fmt.Errorf("error duplicate sequence key: %s ", vendorJournal.ID)
			}
			vendorJournal.ID = "AP/" + site.Alias + vendorJournal.ID

			if e := h.Save(&vendorJournal); e != nil {
				return nil, fmt.Errorf("error : save vendorJournal: %s", opt.JournalID)
			}
			// set expense with journalID
			for _, exp := range p {
				expJournal[exp.Account.AccountID] = vendorJournal.ID
			}
			vendorJournals = append(vendorJournals, vendorJournal)
		}
	}
	// update SiteEntry with JournalID
	exp := []bagongmodel.SiteExpense{}
	if opt.ModuleID == string(PostTypeSiteEntryBTS) {
		for _, c := range detailBTS.Expense {
			tmp := c
			if c.JournalID == "" {
				if v, ok := expJournal[c.ExpenseTypeID]; ok {
					tmp.JournalID = v
					if _, ok := expJournalUpload[v]; !ok {
						expJournalUpload[v] = v
					}

					// set attachment
					vTag := "SE_BTS_EXPENSE_" + opt.JournalID + "_" + c.ID
					vNewTag := []string{string(ficomodel.SubledgerVendor) + "_" + v + "_" + strconv.Itoa(c.LineNo)}
					vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+v)
					updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
						JournalType: "",
						JournalID:   "",
						Tags:        []string{vTag},
						NewTags:     vNewTag,
					})
				}
			}
			exp = append(exp, tmp)
		}
		detailBTS.Expense = exp
		if e := h.Save(detailBTS); e != nil {
			return nil, fmt.Errorf("error : save detailBTS: %s", opt.JournalID)
		}

		// update tag tab attachment
		vNewTag := []string{}
		if len(expJournalUpload) > 0 {
			for _, valjur := range expJournalUpload {
				vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+valjur)
			}
		}

		if len(vNewTag) > 0 {
			updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
				JournalType: "SE_BTS",
				JournalID:   opt.JournalID,
				Tags:        []string{},
				NewTags:     vNewTag,
			})
		}
	} else if opt.ModuleID == string(PostTypeSiteEntryMining) {
		for _, c := range detailMining.Expense {
			tmp := c
			if c.JournalID == "" {
				if v, ok := expJournal[c.ExpenseTypeID]; ok {
					tmp.JournalID = v
					if _, ok := expJournalUpload[v]; !ok {
						expJournalUpload[v] = v
					}

					// set attachment
					vTag := "SE_MINING_EXPENSE_" + opt.JournalID + "_" + c.ID
					vNewTag := []string{string(ficomodel.SubledgerVendor) + "_" + v + "_" + strconv.Itoa(c.LineNo)}
					vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+v)
					updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
						JournalType: "",
						JournalID:   "",
						Tags:        []string{vTag},
						NewTags:     vNewTag,
					})
				}
			}
			exp = append(exp, tmp)
		}
		detailMining.Expense = exp
		if e := h.Save(detailMining); e != nil {
			return nil, fmt.Errorf("error : save detailMining: %s", opt.JournalID)
		}

		// update tag tab attachment
		vNewTag := []string{}
		if len(expJournalUpload) > 0 {
			for _, valjur := range expJournalUpload {
				vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+valjur)
			}
		}

		if len(vNewTag) > 0 {
			updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
				JournalType: "SE_MINING",
				JournalID:   opt.JournalID,
				Tags:        []string{},
				NewTags:     vNewTag,
			})
		}
	} else if opt.ModuleID == string(PostTypeSiteEntryTourism) {
		expOp := []bagongmodel.SiteExpense{}
		expOt := []bagongmodel.SiteExpense{}
		for _, c := range detailTourism.OperationalExpense {
			tmp := c
			if c.JournalID == "" {
				if v, ok := expJournal[c.ExpenseTypeID]; ok {
					tmp.JournalID = v
					if _, ok := expJournalUpload[v]; !ok {
						expJournalUpload[v] = v
					}

					// set attachment
					vTag := "SE_TOURISM_OPERATIONAL_EXPENSE_" + opt.JournalID + "_" + c.ID
					vNewTag := []string{string(ficomodel.SubledgerVendor) + "_" + v + "_" + strconv.Itoa(c.LineNo)}
					vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+v)
					updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
						JournalType: "",
						JournalID:   "",
						Tags:        []string{vTag},
						NewTags:     vNewTag,
					})
				}
			}
			expOp = append(expOp, tmp)
		}

		for _, c := range detailTourism.OtherExpense {
			tmp := c
			if c.JournalID == "" {
				if v, ok := expJournal[c.ExpenseTypeID]; ok {
					tmp.JournalID = v
					if _, ok := expJournalUpload[v]; !ok {
						expJournalUpload[v] = v
					}

					// set attachment
					vTag := "SE_TOURISM_OTHER_EXPENSE_" + opt.JournalID + "_" + c.ID
					vNewTag := []string{string(ficomodel.SubledgerVendor) + "_" + v + "_" + strconv.Itoa(c.LineNo)}
					vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+v)
					updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
						JournalType: "",
						JournalID:   "",
						Tags:        []string{vTag},
						NewTags:     vNewTag,
					})
				}
			}
			expOt = append(expOt, tmp)
		}
		detailTourism.OperationalExpense = expOp
		detailTourism.OtherExpense = expOt
		if e := h.Save(detailTourism); e != nil {
			return nil, fmt.Errorf("error : save detailTourism: %s", opt.JournalID)
		}

		// update tag tab attachment
		vNewTag := []string{}
		if len(expJournalUpload) > 0 {
			for _, valjur := range expJournalUpload {
				vNewTag = append(vNewTag, string(ficomodel.SubledgerVendor)+"_"+string(ficomodel.SubledgerVendor)+" "+valjur)
			}
		}

		if len(vNewTag) > 0 {
			updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
				JournalType: "SE_TOURISM",
				JournalID:   opt.JournalID,
				Tags:        []string{},
				NewTags:     vNewTag,
			})
		}
	}

	// update attachment tag
	if len(updateAttachmentTags) > 0 {
		evHub, _ := ctx.DefaultEvent()

		vRes := ""
		err := evHub.Publish("/v1/asset/update-tag-by-journal", updateAttachmentTags, &vRes, nil)
		if err != nil {
			return nil, fmt.Errorf("error : update attachment tags: %s", err.Error())
		}

	}

	return vendorJournals, nil
}

type ResponseTrayekPosting struct {
	CustomerJournal []*tenantcoremodel.PreviewReport
	LedgerJournal   []*tenantcoremodel.PreviewReport
}

func SiteEntryPosting(ctx *kaos.Context, payload PostRequest, reqOpt ficologic.PostingHubCreateOpt) (ficologic.EnvelopePost, error) {
	post := ficologic.EnvelopePost{}
	// call API Posting
	postReq := ficologic.EventPostRequest{
		CompanyID: reqOpt.CompanyID,
		UserID:    reqOpt.UserID,
		PostRequest: []ficologic.PostRequest{
			{
				JournalType: tenantcoremodel.TrxModule(payload.JournalType),
				JournalID:   payload.JournalID,
				Op:          ficologic.PostOp(payload.Op),
				Text:        payload.Text,
			},
		},
	}

	evHub, _ := ctx.DefaultEvent()

	err := evHub.Publish("/v1/fico/postingprofile/post", postReq, &post, nil)
	if err != nil {
		return post, fmt.Errorf("PostingProfileHandler: %s | %s | %s ", tenantcoremodel.TrxModule(payload.JournalType), payload.JournalID, err)
	}

	return post, nil
}

func SiteEntryTrayekPosting(ctx *kaos.Context, ritase bagongmodel.SiteEntryTrayekRitase, customerJournalID, vendorJournalID string, payload PostRequest, reqOpt ficologic.PostingHubCreateOpt) (ficologic.EnvelopePost, error) {
	post := ficologic.EnvelopePost{}
	// call API Posting
	postReq := ficologic.EventPostRequest{
		CompanyID: reqOpt.CompanyID,
		UserID:    reqOpt.UserID,
	}

	if customerJournalID != "" {
		postReq.PostRequest = append(postReq.PostRequest, ficologic.PostRequest{
			JournalType: ficomodel.SubledgerCustomer,
			JournalID:   customerJournalID,
			Op:          ficologic.PostOp(payload.Op),
			Text:        ritase.TrayekName + " " + payload.Text,
		})
	}

	if ritase.RevenueType == "Premi" {
		if vendorJournalID != "" {
			postReq.PostRequest = append(postReq.PostRequest, ficologic.PostRequest{
				JournalType: ficomodel.SubledgerVendor,
				JournalID:   vendorJournalID,
				Op:          ficologic.PostOp(payload.Op),
				Text:        ritase.TrayekName + " " + payload.Text,
			})
		}
	}

	evHub, _ := ctx.DefaultEvent()

	err := evHub.Publish("/v1/fico/postingprofile/post", postReq, &post, nil)
	if err != nil {
		return post, fmt.Errorf("PostingProfileHandler: %s | %s | %s ", tenantcoremodel.TrxModule(payload.JournalType), payload.JournalID, err)
	}

	return post, nil
}

func GetDetailLedgerJournalPost(hub *datahub.Hub, payload JournalPostReq) (interface{}, error) {
	if payload.SiteID == "" {
		return 0, errors.New("missing: invalid request, please check your payload")
	}

	mapSiteJournal := new(tenantcoremodel.SiteEntryJournalType)
	e := hub.GetByFilter(mapSiteJournal, dbflex.And(dbflex.Eq("SiteID", payload.SiteID), dbflex.Eq("Type", payload.Type)))
	if e != nil {
		return 0, e
	}

	if payload.Type == "Revenue" {
		siteJournal := new(ficomodel.CustomerJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	} else if payload.Type == "Expense" {
		siteJournal := new(ficomodel.VendorJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	} else {
		siteJournal := new(ficomodel.LedgerJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	}
}

func SiteEntryRollbackJournal(hub *datahub.Hub, payload SiteEntryRollbackRequest) error {
	if payload.JournalType == PostTypeSiteEntryTrayek {
		trayekRitase := new(bagongmodel.SiteEntryTrayekRitase)
		e := hub.GetByID(trayekRitase, payload.JournalID)
		if e != nil {
			return fmt.Errorf("error : trayekRitase not found: %s", payload.JournalID)
		}

		for i := range trayekRitase.FixExpense {
			if _, ok := payload.MapJournalID[trayekRitase.FixExpense[i].JournalID]; !ok {
				trayekRitase.FixExpense[i].JournalID = ""
			}
		}
		for i := range trayekRitase.OtherExpense {
			if _, ok := payload.MapJournalID[trayekRitase.OtherExpense[i].JournalID]; !ok {
				trayekRitase.OtherExpense[i].JournalID = ""
			}
		}
		for i := range trayekRitase.OtherIncome {
			if _, ok := payload.MapJournalID[trayekRitase.OtherIncome[i].JournalID]; !ok {
				trayekRitase.OtherIncome[i].JournalID = ""
			}
		}

		trayekRitase.PassengerIncome.JournalID = ""
		if e := hub.Save(trayekRitase); e != nil {
			return fmt.Errorf("error : save trayekRitase: %s", payload.JournalID)
		}

		customerJournal := new(ficomodel.CustomerJournal)
		e = hub.GetByID(customerJournal, payload.CustomerJournalID)
		if e != nil {
			return errors.New(fmt.Sprintf("customerJournal not found: %s", payload.JournalID))
		}
		if e := hub.Delete(customerJournal); e != nil {
			return fmt.Errorf("error when delete customerJournal: %s", e)
		}
	} else if payload.JournalType == PostTypeSiteEntryBTS {
		detail := new(bagongmodel.SiteEntryBTSDetail)
		e := hub.GetByID(detail, payload.JournalID)
		if e != nil {
			return fmt.Errorf("site entry detail not found: %s", payload.JournalID)
		}
		for i := range detail.Expense {
			if _, ok := payload.MapJournalID[detail.Expense[i].JournalID]; !ok {
				detail.Expense[i].JournalID = ""
			}
		}
		if e := hub.Save(detail); e != nil {
			return fmt.Errorf("error : save detail: %s", payload.JournalID)
		}
	} else if payload.JournalType == PostTypeSiteEntryMining {
		detail := new(bagongmodel.SiteEntryMiningDetail)
		e := hub.GetByID(detail, payload.JournalID)
		if e != nil {
			return fmt.Errorf("site entry detail not found: %s", payload.JournalID)
		}
		for i := range detail.Expense {
			if _, ok := payload.MapJournalID[detail.Expense[i].JournalID]; !ok {
				detail.Expense[i].JournalID = ""
			}
		}
		if e := hub.Save(detail); e != nil {
			return fmt.Errorf("error : save detail: %s", payload.JournalID)
		}
	} else if payload.JournalType == PostTypeSiteEntryTourism {
		detail := new(bagongmodel.SiteEntryTourismDetail)
		e := hub.GetByID(detail, payload.JournalID)
		if e != nil {
			return fmt.Errorf("site entry detail not found: %s", payload.JournalID)
		}

		for i := range detail.OperationalExpense {
			if _, ok := payload.MapJournalID[detail.OperationalExpense[i].JournalID]; !ok {
				detail.OperationalExpense[i].JournalID = ""
			}
		}

		for i := range detail.OtherExpense {
			if _, ok := payload.MapJournalID[detail.OtherExpense[i].JournalID]; !ok {
				detail.OtherExpense[i].JournalID = ""
			}
		}

		if e := hub.Save(detail); e != nil {
			return fmt.Errorf("error : save detail: %s", payload.JournalID)
		}
	} else if payload.JournalType == PostTypeSiteEntryNonAsset {
		nonAsset := new(bagongmodel.SiteEntryNonAsset)
		e := hub.GetByID(nonAsset, payload.JournalID)
		if e != nil {
			return fmt.Errorf("nonAsset not found: %s", payload.JournalID)
		}

		for i := range nonAsset.ExpenseDetail {
			if _, ok := payload.MapJournalID[nonAsset.ExpenseDetail[i].JournalID]; !ok {
				nonAsset.ExpenseDetail[i].JournalID = ""
			}
		}

		if e := hub.Save(nonAsset); e != nil {
			return fmt.Errorf("error : save trayekRitase: %s", payload.JournalID)
		}
	}

	vendorJournal := new(ficomodel.VendorJournal)
	e := hub.GetByID(vendorJournal, payload.VendorJournalID)
	if e != nil {
		return errors.New(fmt.Sprintf("vendorJournal not found: %s", payload.JournalID))
	}
	if e := hub.Delete(vendorJournal); e != nil {
		return fmt.Errorf("error when delete vendorJournal: %s", e)
	}
	return nil
}

type PostOp string
type PostType string

const (
	PostOpPreview PostOp = "Preview"
	PostOpSubmit  PostOp = "Submit"
	PostOpApprove PostOp = "Approve"
	PostOpReject  PostOp = "Reject"
	PostOpPost    PostOp = "Post"

	PostTypeSiteEntryTrayek   PostType = "SITEENTRY_TRAYEK"
	PostTypeSiteEntryNonAsset PostType = "SITEENTRY_NONASSET"
	PostTypeSiteEntryBTS      PostType = "SITEENTRY_BTS"
	PostTypeSiteEntryTourism  PostType = "SITEENTRY_TOURISM"
	PostTypeSiteEntryMining   PostType = "SITEENTRY_MINING"
	PostTypeAssetMovement     PostType = "ASSETMOVEMENT"
)

type PostRequest struct {
	JournalType PostType
	JournalID   string
	Op          PostOp
	Text        string
}

type SiteEntryRollbackRequest struct {
	JournalType       PostType
	JournalID         string
	VendorJournalID   string
	CustomerJournalID string
	// for saving journal id when there is an error
	// for reset journal id purpose
	MapJournalID map[string]bool
}

type MapSourceDataToURLRequest struct {
	SourceType string
	SourceID   string
}

type MapSourceDataToURLResponse struct {
	Menu string
	URL  string
}

func (pph *PostingProfileHandler) MapSourceDataToUrl(ctx *kaos.Context, req *MapSourceDataToURLRequest) (*MapSourceDataToURLResponse, error) {
	res := &MapSourceDataToURLResponse{
		Menu: req.SourceType,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: db")
	}

	url := strings.Replace(bagongconfig.Config.AddrWebTenant, "/app", "/bagong", 1) // http://localhost:37000/app -> http://localhost:37000/bagong
	urlPath := bagongmodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(bagongmodel.ModuleAssetmovement) {
		soORM := sebar.NewMapRecordWithORM(h, new(bagongmodel.AssetMovement))
		j, _ := soORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string("Asset Movement")
			urlPath = bagongmodel.SourceTypeURLMap[string(bagongmodel.ModuleAssetmovement)]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
	return res, nil
}
