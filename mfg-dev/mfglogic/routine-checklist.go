package mfglogic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type RoutineChecklistEngine struct{}

type RoutineChecklistGetChecklistRequest struct {
	RoutineDetailID string
}

type RoutineChecklistResponse struct {
	RoutineDetailID             string
	RoutineChecklist            mfgmodel.RoutineChecklist
	RoutineChecklistCategories  []RoutineChecklistCategoryResponse
	RoutineChecklistAttachments []mfgmodel.RoutineChecklistAttachment
}

type RoutineChecklistCategoryResponse struct {
	NoText                  string // "A", "B", ...
	CategoryID              string
	CategoryName            string
	RoutineChecklistDetails []RoutineChecklistDetailResponse
}

type RoutineChecklistDetailResponse struct {
	mfgmodel.RoutineChecklistDetail
	No int // 1, 2, ...
}

func (o *RoutineChecklistEngine) GetChecklist(ctx *kaos.Context, payload *RoutineChecklistGetChecklistRequest) (*RoutineChecklistResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	// get data employee base on login id
	emp := new(tenantcoremodel.Employee)
	if e := h.GetByID(emp, userID); e != nil && e != io.EOF {
		return nil, fmt.Errorf("employee not found: %s", payload)
	}

	rc := new(mfgmodel.RoutineChecklist)
	h.GetByAttr(rc, "_id", payload.RoutineDetailID) // 1-to-1 with RoutineDetail

	// set name, departmnet and work location base on employee
	if rc.Name == "" {
		rc.Name = emp.ID
	}
	if rc.Department == "" {
		rc.Department = emp.Dimension.Get("CC")
	}
	if rc.WorkLocation == "" {
		rc.WorkLocation = emp.Dimension.Get("Site")
	}

	if rc.ID == "" {
		// get empty data with checklist data from master
		rd := new(mfgmodel.RoutineDetail)
		if e := h.GetByID(rd, payload.RoutineDetailID); e != nil {
			return nil, e
		}

		rts := []mfgmodel.RoutineTemplate{}
		if e := h.GetsByFilter(new(mfgmodel.RoutineTemplate), dbflex.And(
			dbflex.Eq("AssetType", rd.AssetType),
			dbflex.Eq("DriveType", rd.DriveType),
		), &rts); e != nil {
			return nil, e
		}

		if len(rts) == 0 {
			return nil, fmt.Errorf("no routine template found, AssetType: %s", rd.AssetType)
		}

		/*
			Point of Interest:
			1. Routine Template ini per 1 Category ID, jadi dalam 1 P2H pasti memakai lebih dari 1 template
			2. Template dalam 1 Category yang sama, hanya akan dipakai 1 template saja (proses penentuannya ada dibawah)

			Proses penentuan template yang akan digunakan:
			loop map[CategoryID][]templates per category id nya:
			- cari template yang CustomerID, Dimension.Site, Dimension.PC nya sesuai, kalo nemu langsung break itu yang dipake
			- cari template yang CustomerID="", Dimension.Site, Dimension.PC nya sesuai, kalo nemu langsung break itu yang dipake
			- cari template yang CustomerID="", Dimension.Site="", Dimension.PC nya sesuai, kalo nemu langsung break itu yang dipake
			- cari template yang CustomerID="", Dimension.Site="", Dimension.PC="", kalo nemu langsung break itu yang dipake
			- kalo masi ga nemu, ga usah dipake semua template di category itu
		*/

		templatePerCtgM := lo.GroupBy(rts, func(d mfgmodel.RoutineTemplate) string {
			return d.CategoryID
		})

		templateUsed := map[string]mfgmodel.RoutineTemplate{}
		for ctgID, tpls := range templatePerCtgM {
			tpl, found := lo.Find(tpls, func(d mfgmodel.RoutineTemplate) bool {
				return d.CustomerID == rd.CustomerID &&
					d.Dimension.Get("Site") == rd.Dimension.Get("Site") &&
					d.Dimension.Get("PC") == rd.Dimension.Get("PC")
			})
			if found {
				templateUsed[ctgID] = tpl
				continue
			}

			tpl, found = lo.Find(tpls, func(d mfgmodel.RoutineTemplate) bool {
				return d.CustomerID == "" &&
					d.Dimension.Get("Site") == rd.Dimension.Get("Site") &&
					d.Dimension.Get("PC") == rd.Dimension.Get("PC")
			})
			if found {
				templateUsed[ctgID] = tpl
				continue
			}

			tpl, found = lo.Find(tpls, func(d mfgmodel.RoutineTemplate) bool {
				return d.CustomerID == "" &&
					d.Dimension.Get("Site") == rd.Dimension.Get("Site") &&
					d.Dimension.Get("PC") == ""
			})
			if found {
				templateUsed[ctgID] = tpl
				continue
			}

			tpl, found = lo.Find(tpls, func(d mfgmodel.RoutineTemplate) bool {
				return d.CustomerID == "" &&
					d.Dimension.Get("Site") == "" &&
					d.Dimension.Get("PC") == rd.Dimension.Get("PC")
			})
			if found {
				templateUsed[ctgID] = tpl
				continue
			}

			tpl, found = lo.Find(tpls, func(d mfgmodel.RoutineTemplate) bool {
				return d.CustomerID == "" &&
					d.Dimension.Get("Site") == "" &&
					d.Dimension.Get("PC") == ""
			})
			if found {
				templateUsed[ctgID] = tpl
				continue
			}
		}

		catM := map[string][]mfgmodel.RoutineChecklistDetail{}
		templateIDUseds := []string{}
		for ctgID, tpl := range templateUsed {
			for _, item := range tpl.Items {
				catM[ctgID] = append(catM[ctgID], mfgmodel.RoutineChecklistDetail{
					RoutineChecklistID: payload.RoutineDetailID,
					CategoryID:         ctgID,
					Name:               item.ItemName,
					Code:               item.DangerousCode,
				})
			}
			templateIDUseds = append(templateIDUseds, tpl.ID)
		}
		rc.RoutineTemplateIDs = templateIDUseds

		// for _, rt := range rts {
		// 	for _, rtItem := range rt.Items {
		// 		catM[rt.CategoryID] = append(catM[rt.CategoryID], mfgmodel.RoutineChecklistDetail{
		// 			RoutineChecklistID: payload.RoutineDetailID,
		// 			CategoryID:         rt.CategoryID,
		// 			Name:               rtItem.ItemName,
		// 			Code:               rtItem.DangerousCode,
		// 		})
		// 	}
		// }

		mdORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))
		catNoM := map[string]string{}
		catIDs := lo.Keys(catM)
		sort.Strings(catIDs)

		for i, catID := range catIDs {
			catNoM[catID] = string('A' + i)
		}

		catRes := []RoutineChecklistCategoryResponse{}
		for catID, dts := range catM {
			md, _ := mdORM.Get(catID)

			cdets := lo.Map(dts, func(dt mfgmodel.RoutineChecklistDetail, i int) RoutineChecklistDetailResponse {
				return RoutineChecklistDetailResponse{RoutineChecklistDetail: dt, No: (i + 1)}
			})

			sort.SliceStable(cdets, func(i, j int) bool {
				return cdets[i].No < cdets[j].No
			})

			catRes = append(catRes, RoutineChecklistCategoryResponse{
				NoText:                  catNoM[catID],
				CategoryID:              catID,
				CategoryName:            md.Name,
				RoutineChecklistDetails: cdets,
			})
		}

		res := RoutineChecklistResponse{
			RoutineDetailID:             payload.RoutineDetailID,
			RoutineChecklist:            *rc,
			RoutineChecklistCategories:  catRes,
			RoutineChecklistAttachments: []mfgmodel.RoutineChecklistAttachment{},
		}

		return &res, nil
	}

	cds := []mfgmodel.RoutineChecklistDetail{}
	if e := h.GetsByFilter(new(mfgmodel.RoutineChecklistDetail), dbflex.Eq("RoutineChecklistID", rc.ID), &cds); e != nil {
		return nil, e
	}

	catRes := ToChecklistCategoryResponse(h, cds)

	atts := []mfgmodel.RoutineChecklistAttachment{}
	if e := h.GetsByFilter(new(mfgmodel.RoutineChecklistAttachment), dbflex.Eq("RoutineChecklistID", rc.ID), &atts); e != nil {
		return nil, e
	}

	res := RoutineChecklistResponse{
		RoutineDetailID:             payload.RoutineDetailID,
		RoutineChecklist:            *rc,
		RoutineChecklistCategories:  catRes,
		RoutineChecklistAttachments: atts,
	}

	return &res, nil
}

func (o *RoutineChecklistEngine) SaveChecklist(ctx *kaos.Context, payload *RoutineChecklistResponse) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	catIDs := []string{}
	for cati, cat := range payload.RoutineChecklistCategories {
		catIDs = append(catIDs, cat.CategoryID)
		for dti, dte := range cat.RoutineChecklistDetails {
			dt := dte.RoutineChecklistDetail
			dt.RoutineChecklistID = payload.RoutineDetailID
			dt.CategoryID = cat.CategoryID
			if e := h.Save(&dt); e != nil {
				return nil, e
			}
			cat.RoutineChecklistDetails[dti].RoutineChecklistDetail = dt
		}
		payload.RoutineChecklistCategories[cati] = cat
	}

	for atti, att := range payload.RoutineChecklistAttachments {
		if e := h.Save(&att); e != nil {
			return nil, e
		}
		payload.RoutineChecklistAttachments[atti] = att
	}

	payload.RoutineChecklist.ID = payload.RoutineDetailID
	payload.RoutineChecklist.CategoryIDs = catIDs
	if e := h.Save(&payload.RoutineChecklist); e != nil {
		return nil, e
	}

	return payload, nil
}

type RoutineChecklistDownloadAsPdfRequest struct {
	RoutineDetailID string
}

func (o *RoutineChecklistEngine) DownloadAsPdf(ctx *kaos.Context, _ *interface{}) ([]byte, error) {
	RoutineDetailID := GetURLQueryParams(ctx)["RoutineDetailID"]

	w, wOK := ctx.Data().Get("http_writer", nil).(http.ResponseWriter)

	if !wOK {
		return nil, errors.New("not a http compliant writer")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	rd := new(mfgmodel.RoutineDetail)
	if e := h.GetByID(rd, RoutineDetailID); e != nil {
		return nil, fmt.Errorf("Please Save before Download")
	}

	routine := new(mfgmodel.Routine)
	if e := h.GetByID(routine, rd.RoutineID); e != nil {
		return nil, fmt.Errorf("Please Save before Download")
	}

	rc := new(mfgmodel.RoutineChecklist)
	if e := h.GetByID(rc, RoutineDetailID); e != nil { // 1-to-1 with RoutineDetail
		return nil, fmt.Errorf("Please Save before Download")
	}

	cds := []mfgmodel.RoutineChecklistDetail{}
	if e := h.GetsByFilter(new(mfgmodel.RoutineChecklistDetail), dbflex.Eq("RoutineChecklistID", rc.ID), &cds); e != nil {
		return nil, e
	}

	catRes := ToChecklistCategoryResponse(h, cds)

	dimORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.DimensionMaster))
	empORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))
	site, _ := dimORM.Get(routine.SiteID)
	emp, _ := empORM.Get(rc.Name)

	pdfData := codekit.M{
		"Name":                emp.Name,
		"Department":          rc.Department,
		"ExecutionDate":       routine.ExecutionDate.Format("Monday, 02 January 2006"),
		"Site":                lo.Ternary(site.Label != "", site.Label, routine.SiteID),
		"Routine":             routine,
		"RoutineDetail":       rd,
		"RoutineChecklist":    rc,
		"ChecklistCategories": catRes,
	}

	fmt.Println("DownloadAsPdf | before TemplateToPDF")
	content, e := TemplateToPDF(&tenantcorelogic.PDFFromTemplateRequest{
		TemplateName: "p2h",
		Data:         pdfData,
	})
	if e != nil {
		fmt.Println("DownloadAsPdf | TemplateToPDF error:", e)
		return nil, e
	}
	fmt.Println("DownloadAsPdf | after TemplateToPDF")

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(content)

	ctx.Data().Set("kaos_command_1", "stop")
	return content, nil
}

func ToChecklistCategoryResponse(h *datahub.Hub, cds []mfgmodel.RoutineChecklistDetail) []RoutineChecklistCategoryResponse {
	catM := map[string][]mfgmodel.RoutineChecklistDetail{}
	for _, cd := range cds {
		catM[cd.CategoryID] = append(catM[cd.CategoryID], cd)
	}

	mdORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))
	catNoM := map[string]string{}
	catIDs := lo.Keys(catM)
	sort.Strings(catIDs)

	for i, catID := range catIDs {
		catNoM[catID] = string('A' + i)
	}

	catRes := []RoutineChecklistCategoryResponse{}
	for catID, dts := range catM {
		md, _ := mdORM.Get(catID)

		cdets := lo.Map(dts, func(dt mfgmodel.RoutineChecklistDetail, i int) RoutineChecklistDetailResponse {
			return RoutineChecklistDetailResponse{RoutineChecklistDetail: dt, No: (i + 1)}
		})

		sort.SliceStable(cdets, func(i, j int) bool {
			return cdets[i].No < cdets[j].No
		})

		catRes = append(catRes, RoutineChecklistCategoryResponse{
			NoText:                  catNoM[catID],
			CategoryID:              catID,
			CategoryName:            md.Name,
			RoutineChecklistDetails: cdets,
		})
	}

	sort.SliceStable(catRes, func(i, j int) bool {
		return catRes[i].NoText < catRes[j].NoText
	})

	return catRes
}
