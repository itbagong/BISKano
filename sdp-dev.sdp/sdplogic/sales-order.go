package sdplogic

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/kasset"
	"github.com/leekchan/accounting"
	"github.com/sebarcode/codekit"
	"github.com/signintech/gopdf"
)

type SalesOrderEngine struct {
	UC UnitCalendarEngine
}
type typetime time.Time

type PayloadActionSO struct {
	ID string `json:"_id"`
}

func (o *SalesOrderEngine) upsert(h *datahub.Hub, model *sdpmodel.SalesOrder) error {
	if e := h.GetByID(new(sdpmodel.SalesOrder), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Sales Order : " + e.Error())
		}
	} else {
		if e := h.Save(model); e != nil {
			return errors.New("error update Sales Order : " + e.Error())
		}
	}

	return nil
}

func (o *SalesOrderEngine) DuplicateSO(ctx *kaos.Context, payload *PayloadActionSO) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" {
		return nil, errors.New("missing: payload")
	}

	SO := new(sdpmodel.SalesOrder)
	e := h.GetByID(SO, payload.ID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing sales order in by ID: %s", payload.ID))
	}

	SO.ID = ""
	SO.Status = ficomodel.JournalStatusDraft

	e = o.upsert(h, SO)
	if e != nil {
		return nil, e
	}
	return SO, nil
}

func (o *SalesOrderEngine) ActionSO(ctx *kaos.Context, payload *PayloadActionSO) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" {
		return nil, errors.New("missing: payload")
	}

	//get quotation in
	SO := new(sdpmodel.SalesQuotation)
	e := h.GetByID(SO, payload.ID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing quotation in by ID: %s", payload.ID))
	}
	//get list of item by list of ids
	items := []sdpmodel.SalesOrder{}
	e = h.GetsByFilter(new(sdpmodel.SalesOrder), dbflex.Eq("Year", time.Now().Year()), &items)
	if e != nil {
		ctx.Log().Errorf("Failed populate data order: %s", e.Error())
	}

	so := new(sdpmodel.SalesOrder)

	so.SalesQuotationRefNo = SO.QuotationNo
	so.SalesOpportunityRefNo = SO.OpportunityNo
	so.CustomerID = SO.Customer
	so.SalesPriceBookID = SO.SalesPriceBook
	// so.SalesOrderDate = time.Now()
	// so.Lines = SO.Lines
	so.Dimension = SO.Dimension
	//editor
	so.LetterHeadAsset = SO.LetterHeadAsset
	so.FooterAsset = SO.FooterAsset
	so.LetterHeadFirst = SO.LetterHeadFirst
	so.FooterLastPage = SO.FooterLastPage
	so.Editor = SO.Editor
	so.TaxCodes = SO.TaxCodes
	so.TransactionType = SO.TransactionType
	// so.JournalType = SO.JournalType
	// so.CompanyID = SO.CompanyID
	so.SalesOrderDate = time.Now()
	TotalAmount := float64(0)
	for _, line := range SO.Lines {
		TotalAmount = TotalAmount + float64(line.Amount)
		if line.Item != "" && int(line.Qty) > 0 && line.Asset == "" {
			for i := 0; i < int(line.Qty); i++ {
				so.Lines = append(so.Lines, struct {
					Asset          string
					Item           string
					Description    string
					Shift          int64
					Qty            uint
					UoM            string
					ContractPeriod int
					UnitPrice      int64
					Amount         int64
					DiscountType   string
					Discount       int
					Taxable        bool
					TaxCodes       []string
					Spesifications []string
					StartDate      time.Time
					EndDate        time.Time
					Checklists     tenantcoremodel.Checklists
					References     tenantcoremodel.References
				}{
					Asset:          line.Asset,
					Item:           line.Item,
					Description:    line.Description,
					Spesifications: line.Spesifications,
					Shift:          line.Shift,
					Qty:            uint(1),
					UoM:            line.UoM,
					ContractPeriod: line.ContractPeriod,
					UnitPrice:      line.UnitPrice,
					Amount:         line.Amount / int64(line.Qty),
					// DiscountType   : line.DiscountType,
					Discount:   line.Discount,
					Taxable:    line.Taxable,
					TaxCodes:   line.TaxCodes,
					StartDate:  time.Now(),
					EndDate:    time.Now(),
					Checklists: tenantcoremodel.Checklists{},
					References: tenantcoremodel.References{},
				})
			}
		} else {
			so.Lines = append(so.Lines, struct {
				Asset          string
				Item           string
				Description    string
				Shift          int64
				Qty            uint
				UoM            string
				ContractPeriod int
				UnitPrice      int64
				Amount         int64
				DiscountType   string
				Discount       int
				// Account        sdpmodel.SubledgerAccount
				Taxable        bool
				TaxCodes       []string
				Spesifications []string
				StartDate      time.Time
				EndDate        time.Time
				Checklists     tenantcoremodel.Checklists
				References     tenantcoremodel.References
			}{
				Asset:          line.Asset,
				Item:           line.Item,
				Description:    line.Description,
				Spesifications: line.Spesifications,
				Shift:          line.Shift,
				Qty:            line.Qty,
				UoM:            line.UoM,
				ContractPeriod: line.ContractPeriod,
				UnitPrice:      line.UnitPrice,
				Amount:         line.Amount,
				// DiscountType   : line.DiscountType,
				Discount:   line.Discount,
				Taxable:    line.Taxable,
				TaxCodes:   line.TaxCodes,
				StartDate:  time.Now(),
				EndDate:    time.Now(),
				Checklists: tenantcoremodel.Checklists{},
				References: tenantcoremodel.References{},
			})
		}
	}

	so.TotalAmount = TotalAmount

	// if so.ID == "" {
	// 	t := time.Now()
	// 	sTime := t.Format("01/2006")
	// 	so.Year = t.Year()
	// 	so.SalesOrderNo = "SO/" + sTime + "/" + strconv.Itoa(len(items))
	// }

	customer := new(tenantcoremodel.Customer)
	err := h.GetByID(customer, so.CustomerID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v", err))
	}

	cs := customer.CustomerAlias
	if cs == "" {
		return nil, errors.New("Customer alias not null")
	}

	last := []sdpmodel.SalesOrder{}
	err = h.Gets(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC))).SetSort("-No").SetTake(1), &last)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Last sales order", err))
	}

	if len(last) > 0 {
		so.No = last[0].No + 1
		so.SalesOrderNo = fmt.Sprintf("SO/%04d/BDM-HO-%s/%02d/%d", so.No, cs, (int(time.Now().Month())), time.Now().Year())
	} else {
		so.No = 1
		so.SalesOrderNo = fmt.Sprintf("SO/%04d/BDM-HO-%s/%02d/%d", so.No, cs, (int(time.Now().Month())), time.Now().Year())
	}

	// Kemungkinan masih perlu di lengkapi data apa saja yag di ambil dari Quotation
	e = o.upsert(h, so)
	if e != nil {
		return nil, e
	}

	return so, nil
}

type SaveSalesOrder struct {
	sdpmodel.SalesOrder
	UploadLetterHeadAsset kasset.AssetDataBase64
	UploadFooterAsset     kasset.AssetDataBase64
}

func (o *SalesOrderEngine) Insert(ctx *kaos.Context, payload *SaveSalesOrder) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.CustomerID == "" {
		return nil, errors.New("missing: payload")
	}

	// check Event hub
	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, errors.New("Event not null")
	}

	so := payload.SalesOrder
	so.TrxDate = so.SalesOrderDate
	so.Text = ""

	writeAsset := func(attachReq kasset.AssetDataBase64) (kasset.Asset, error) {
		spliter := strings.Split(attachReq.Content, ",")
		if len(spliter) > 1 {
			attachReq.Content = spliter[1]
		}
		bs, e := base64.StdEncoding.DecodeString(attachReq.Content)
		if e != nil {
			return kasset.Asset{}, fmt.Errorf("fail to decode content. %s", e.Error())
		}

		reply := kasset.Asset{}
		err = event.Publish("/v1/asset/write", &kasset.AssetData{
			Asset:   attachReq.Asset,
			Content: bs,
		}, &reply, nil)
		if err != nil {
			return kasset.Asset{}, err
		}

		return reply, nil
	}

	if payload.UploadFooterAsset.Content != "" {
		reply, err := writeAsset(payload.UploadFooterAsset)
		if err != nil {
			return nil, err
		}
		so.LetterHeadAsset = reply.ID
	}

	if payload.UploadFooterAsset.Content != "" {
		reply, err := writeAsset(payload.UploadFooterAsset)
		if err != nil {
			return nil, err
		}

		so.FooterAsset = reply.ID
	}

	customer := new(tenantcoremodel.Customer)
	err = h.GetByID(customer, so.CustomerID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v", err))
	}

	cs := customer.CustomerAlias
	if cs == "" {
		return nil, errors.New("Customer alias not null")
	}

	last := []sdpmodel.SalesOrder{}
	err = h.Gets(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC))).SetSort("-No").SetTake(1), &last)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Last sales order", err))
	}

	if len(last) > 0 {
		so.No = last[0].No + 1
		so.SalesOrderNo = fmt.Sprintf("SO/%04d/BDM-HO-%s/%02d/%d", so.No, cs, (int(time.Now().Month())), time.Now().Year())
	} else {
		so.No = 1
		so.SalesOrderNo = fmt.Sprintf("SO/%04d/BDM-HO-%s/%02d/%d", so.No, cs, (int(time.Now().Month())), time.Now().Year())
	}

	reply := []map[string]any{}
	err = event.Publish("/v1/tenant/warehouse/find", &map[string]any{
		"where": map[string]any{
			"Op": "$and",
			"Items": []map[string]any{
				{
					"Op":    "$eq",
					"Field": "Dimension.Key",
					"Value": "Site",
				},
				{
					"Op":    "$eq",
					"Field": "Dimension.Value",
					"Value": "SITE032",
				},
			},
		},
	}, &reply, nil)
	if err != nil {
		return nil, err
	}

	// NOTE: disable because sales order can't set warehouse id by Site on save
	// so.WarehouseID = reply[0]["_id"].(string)

	err = o.upsert(h, &so)
	if err != nil {
		return nil, err
	}

	return so, nil
}

func (o *SalesOrderEngine) Update(ctx *kaos.Context, payload *SaveSalesOrder) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.CustomerID == "" {
		return nil, errors.New("missing: payload")
	}

	// check Event hub
	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, errors.New("Event not null")
	}

	so := payload.SalesOrder
	so.TrxDate = so.SalesOrderDate
	so.Text = ""

	writeAsset := func(attachReq kasset.AssetDataBase64) (kasset.Asset, error) {
		spliter := strings.Split(attachReq.Content, ",")
		if len(spliter) > 1 {
			attachReq.Content = spliter[1]
		}
		bs, e := base64.StdEncoding.DecodeString(attachReq.Content)
		if e != nil {
			return kasset.Asset{}, fmt.Errorf("fail to decode content. %s", e.Error())
		}

		reply := kasset.Asset{}
		err = event.Publish("/v1/asset/write", &kasset.AssetData{
			Asset:   attachReq.Asset,
			Content: bs,
		}, &reply, nil)
		if err != nil {
			return kasset.Asset{}, err
		}

		return reply, nil
	}

	if payload.UploadLetterHeadAsset.Content != "" {
		reply, err := writeAsset(payload.UploadLetterHeadAsset)
		if err != nil {
			return nil, err
		}
		so.LetterHeadAsset = reply.ID
	}

	if payload.UploadFooterAsset.Content != "" {
		reply, err := writeAsset(payload.UploadFooterAsset)
		if err != nil {
			return nil, err
		}

		so.FooterAsset = reply.ID
	}

	reply := []map[string]any{}
	err = event.Publish("/v1/tenant/warehouse/find", &map[string]any{
		"where": map[string]any{
			"Op": "$and",
			"Items": []map[string]any{
				{
					"Op":    "$eq",
					"Field": "Dimension.Key",
					"Value": "Site",
				},
				{
					"Op":    "$eq",
					"Field": "Dimension.Value",
					"Value": "SITE032",
				},
			},
		},
	}, &reply, nil)
	if err != nil {
		return nil, err
	}

	// NOTE: disable because sales order can't set warehouse id by Site on save
	// so.WarehouseID = reply[0]["_id"].(string)

	err = o.upsert(h, &so)
	if err != nil {
		return nil, err
	}

	return so, nil
}

func (o *SalesOrderEngine) PrintPDF(ctx *kaos.Context, payload *PayloadPrintPDF) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" {
		return nil, errors.New("Missing: Payload")
	}

	SO := sdpmodel.SalesOrder{}
	err := h.GetByID(&SO, payload.ID)
	if err != nil {
		return nil, err
	}

	ev, e := ctx.DefaultEvent()
	if e != nil {
		return "", e
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	reply := map[string]any{}
	ev.Publish("/v1/bagong/customer/get", []string{SO.CustomerID}, &reply, &kaos.PublishOpts{Headers: codekit.M{sebar.CtxJWTReferenceID: userID}})

	id, err := printpdfso(ctx, payload, &SO, reply)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func printpdfso(ctx *kaos.Context, payload *PayloadPrintPDF, SO *sdpmodel.SalesOrder, customer map[string]any) (string, error) {

	// check Event hub
	event, err := ctx.DefaultEvent()
	if err != nil {
		return "", err
	}

	if event == nil {
		return "", errors.New("Event not null")
	}

	var paperSize = &gopdf.Rect{
		W: 596,
		H: 935,
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *paperSize})
	// 		var defa TtfOption
	// defa.UseKerning = false
	// defa.Style = Regular
	// defa.OnGlyphNotFoundSubstitute = DefaultOnGlyphNotFoundSubstitute
	// return defa

	err = pdf.AddTTFFontWithOption("times-new-roman", "data/template/times-new-roman.ttf", gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Regular,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return "", err
	}

	err = pdf.AddTTFFontWithOption("times-new-roman", "data/template/times-new-roman.ttf", gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Underline,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return "", err
	}

	err = pdf.AddTTFFontWithOption("times-new-roman", "data/template/times-new-roman-italic.ttf", gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Italic,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return "", err
	}

	err = pdf.AddTTFFontWithOption("times-new-roman", "data/template/times-new-roman-bold.ttf", gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Bold,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return "", err
	}

	// err = pdf.AddTTFFontWithOption("times-new-roman", "data/template/times-new-roman-bold-italic.ttf", gopdf.TtfOption{
	// 	UseKerning:                true,
	// 	Style:                     gopdf.Bold,
	// 	OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	// })
	// if err != nil {
	// 	return err
	// }

	const fontSize = 11
	var Padding = struct {
		Top    float64
		Left   float64
		Right  float64
		Bottom float64
	}{
		Top:    5 * 2.83465, // mm * 2,83465 (point)
		Left:   10 * 2.83465,
		Right:  10 * 2.83465,
		Bottom: 10 * 2.83465,
	}

	const spacing = 2

	err = pdf.SetFont("times-new-roman", "", fontSize)
	if err != nil {
		return "", err
	}

	// pdf.SetMargins(2, 2, 2, 2)

	var wg sync.WaitGroup
	var headerData, footerData []byte

	wg.Add(1)
	go func() {
		defer wg.Done()

		if SO.LetterHeadAsset == "" {
			return
		}

		urlasset := payload.PublicURL + "/asset/view?id=" + SO.LetterHeadAsset
		resp, err := http.Get(urlasset)
		if err != nil {
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			err = errors.New("id asset not found")
			return
		}

		headerData, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		return
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if SO.FooterAsset == "" {
			return
		}

		urlasset := payload.PublicURL + "/asset/view?id=" + SO.FooterAsset
		resp, err := http.Get(urlasset)
		if err != nil {
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			err = errors.New("id asset not found")
			return
		}

		footerData, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		return
	}()

	wg.Wait()

	if err != nil {
		return "", err
	}

	// 72dpi
	pdf.AddPage()

	var line float64 = Padding.Top // 100 is height of header and 10 is new line

	if SO.LetterHeadFirst {
		if SO.LetterHeadAsset == "" || len(headerData) <= 0 {
			return "", errors.New("Asset not found")
		}

		im, _, err := image.DecodeConfig(bytes.NewReader(headerData))
		if err != nil {
			return "", err
		}

		imgheader, err := gopdf.ImageHolderByBytes(headerData)
		if err != nil {
			return "", err
		}

		width := float64(im.Width)
		err = pdf.ImageByHolder(imgheader, Padding.Left, Padding.Top, &gopdf.Rect{
			W: (width - (width * 67.5 / 100)) - (Padding.Left + Padding.Right), // 67.5 is percentage
			H: float64(im.Height) - (float64(im.Height) * 67.5 / 100),
		})

		if err != nil {
			err = errors.New(err.Error() + ". insert PDF")
			return "", err
		}

		line = 106.7 + Padding.Top
	}

	// No SO
	pdf.SetXY(Padding.Left, line)
	line += fontSize + spacing
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Nomer: "+SO.SalesOrderNo)
	if err != nil {
		return "", err
	}

	// // Perihal
	// pdf.SetXY(Padding.Left, line)
	// line += fontSize + spacing
	// if line > paperSize.H {
	// 	pdf.AddPage()
	// line = 0
	// }
	// err = pdf.Cell(nil, "Perihal: "+SO.Name)
	// if err != nil {
	// 	return "", err
	// }

	// Date
	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
	line += (fontSize + spacing) * 2
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Malang, "+SO.SalesOrderDate.Format("15-01-2006"))
	if err != nil {
		return "", err
	}

	// Kepada Yth
	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
	line += fontSize + spacing
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Kepada Yth:")
	if err != nil {
		return "", err
	}

	if customer["Name"] != nil && customer["Name"].(string) != "" {
		// Customer Name
		pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
		line += fontSize + spacing
		if line > paperSize.H {
			pdf.AddPage()
			line = 0
		}
		err = pdf.Cell(nil, customer["Name"].(string))
		if err != nil {
			return "", err
		}
	}

	if customer["Detail"] != nil {
		if customer["Detail"].(map[string]any)["PersonalContact"] != nil && customer["Detail"].(map[string]any)["PersonalContact"].(string) != "" {
			// Customer Personal Contact
			pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
			line += fontSize + spacing
			if line > paperSize.H {
				pdf.AddPage()
				line = 0
			}
			err = pdf.Cell(nil, customer["Detail"].(map[string]any)["PersonalContact"].(string))
			if err != nil {
				return "", err
			}
		}

		if customer["Detail"].(map[string]any)["Address"] != nil && customer["Detail"].(map[string]any)["Address"].(string) != "" {
			// Customer Address
			address := customer["Detail"].(map[string]any)["Address"].(string)

			err := wraptext(func(textwrap string) error {
				pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
				line += fontSize + spacing
				if line > paperSize.H {
					pdf.AddPage()
					line = 0
				}
				err = pdf.Cell(nil, textwrap)
				if err != nil {
					return err
				}

				return nil
			}, address, &pdf, 200/fontSize, nil)

			if err != nil {
				return "", err
			}
		}
	}

	// // detail mail
	// pdf.SetXY(Padding.Left, line)
	// line += fontSize + spacing
	// if line > paperSize.H {
	// 	pdf.AddPage()
	// line = 0
	// }
	// err = pdf.Cell(nil, "Dengan Hormat")
	// if err != nil {
	// 	return "", err
	// }

	// pdf.SetXY(Padding.Left, line)
	// line += fontSize + spacing
	// if line > paperSize.H {
	// 	pdf.AddPage()
	// line = 0
	// }
	// err = pdf.Cell(nil, "Bersama dengan ini kami sampaikan "+SO.QuotationName+" dengan rincian sebagai berikut:")
	// if err != nil {
	// 	return "", err
	// }

	Items := []string{}
	Assets := []string{}
	for _, line := range SO.Lines {
		Items = append(Items, line.Item)
		Assets = append(Assets, line.Asset)
	}

	if len(Assets) > 0 {
		AssetTenants := []map[string]any{}
		err = event.Publish("/v1/tenant/asset/find", map[string]any{
			"Select": []string{
				"_id",
				"Name",
			},
			"Where": map[string]any{
				"Op":    "$in",
				"field": "_id",
				"value": Assets,
			},
		}, &AssetTenants, nil)
		if err != nil {
			return "", err
		}

		for _, AssetTenant := range AssetTenants {
			for index, line := range SO.Lines {
				if AssetTenant["_id"] == line.Asset {
					line.Asset = AssetTenant["Name"].(string)
					SO.Lines[index] = line
				}
			}
		}
	}

	if len(Items) > 0 {
		ItemTenants := []map[string]any{}
		err = event.Publish("/v1/tenant/item/find", map[string]any{
			"Select": []string{
				"_id",
				"Name",
			},
			"Where": map[string]any{
				"Op":    "$in",
				"field": "_id",
				"value": Items,
			},
		}, &ItemTenants, nil)
		if err != nil {
			return "", err
		}

		for _, ItemTenant := range ItemTenants {
			for index, line := range SO.Lines {
				if ItemTenant["_id"] == line.Item {
					line.Item = ItemTenant["Name"].(string)
					SO.Lines[index] = line
				}
			}
		}
	}

	err = tablepdfso(sdpmodel.SalesOrderLinePreviewGrid{
		Qty:  25,
		Item: 129,
	}, SO.Lines, &line, fontSize, paperSize, &pdf, Padding, struct {
		Top  float64
		Left float64
	}{
		Top:  2,
		Left: 2,
	})
	if err != nil {
		return "", err
	}

	line += 4

	ac := accounting.Accounting{Symbol: "Rp ", Precision: 0, Thousand: ".", Decimal: ","}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-160, line)
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}

	err = pdf.Cell(nil, "Subtotal Amount")
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-70, line)
	line += (fontSize) * 2
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}

	err = pdf.Cell(nil, ac.FormatMoney(SO.SubTotalAmount))
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-160, line)
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Tax ")
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-70, line)
	line += (fontSize) * 2
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, ac.FormatMoney(SO.TaxAmount))
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-160, line)
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Discount ")
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-70, line)
	line += (fontSize) * 2
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, ac.FormatMoney(SO.DiscountAmount))
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-160, line+2)
	err = pdf.Cell(nil, "Total Amount ")
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-70, line+2)
	err = pdf.Cell(nil, ac.FormatMoney(SO.TotalAmount))
	if err != nil {
		return "", err
	}

	err = pdf.Rectangle(Padding.Left, line, (paperSize.W - Padding.Right), (float64(fontSize) + line + 4), "D", 0, 0)
	if err != nil {
		return "", err
	}
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	line += (fontSize + spacing + 2) * 2

	err = htmlpdf(SO.Editor, &pdf, &line, fontSize, spacing, paperSize, Padding)
	if err != nil {
		return "", err
	}

	if SO.FooterLastPage {
		if SO.FooterAsset == "" || len(footerData) <= 0 {
			return "", errors.New("asset not found")
		}

		im, _, err := image.DecodeConfig(bytes.NewReader(footerData))
		if err != nil {
			return "", err
		}

		imgheader, err := gopdf.ImageHolderByBytes(footerData)
		if err != nil {
			return "", err
		}

		width := (float64(im.Width) - (float64(im.Width) * 67.5 / 100)) - (Padding.Left + Padding.Right)
		height := float64(im.Height) - (float64(im.Height) * 67.5 / 100)
		if line+height > paperSize.H {
			line = 0
			pdf.AddPage()
			line = 0
		}

		err = pdf.ImageByHolder(imgheader, Padding.Left, line, &gopdf.Rect{
			W: width, // 67.5 is percentage
			H: height,
		})

		if err != nil {
			err = errors.New(err.Error() + ". insert PDF")
			return "", err
		}
	}

	datapdf := bytes.Buffer{}
	_, err = pdf.WriteTo(&datapdf)
	if err != nil {
		return "", err
	}

	type Assetcontent struct {
		Asset   *kasset.Asset
		Content []byte
	}

	reply := map[string]any{}
	err = event.Publish("/v1/asset/write", Assetcontent{
		Asset: &kasset.Asset{
			OriginalFileName: "PDF-" + SO.ID,
			Kind:             "Sales Quotation",
			RefID:            SO.ID,
			Data: map[string]any{
				"pdf": true,
			},
		},
		Content: datapdf.Bytes(),
	}, &reply, nil)
	if err != nil {
		return "", err
	}

	return reply["_id"].(string), nil
}

func tablepdfso(model any, SOlines []sdpmodel.SalesOrderLine, line *float64, fontSize int, paper *gopdf.Rect, pdf *gopdf.GoPdf, globalpadding struct {
	Top    float64
	Left   float64
	Right  float64
	Bottom float64
}, Padding struct {
	Top  float64
	Left float64
}) error {
	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if t.Kind() == reflect.Struct {
		widthRect := ((paper.W - (globalpadding.Right + globalpadding.Left)) / float64(t.NumField()))
		XCache := globalpadding.Left
		marginY := 2
		for i := 0; i < t.NumField(); i++ {
			name := t.Field(i).Name
			pdf.SetStrokeColor(0, 0, 0)
			pdf.SetLineWidth(1)
			pdf.SetXY(XCache+Padding.Left, *line+Padding.Top)
			err := pdf.Cell(nil, name)
			if err != nil {
				return err
			}

			if v.Field(i).Interface().(float64) > 0 {
				err = pdf.Rectangle(XCache, *line, (XCache + v.Field(i).Interface().(float64)), (float64(marginY) + float64(fontSize) + *line), "D", 0, 0)
				if err != nil {
					return err
				}
				XCache += v.Field(i).Interface().(float64)
			} else {
				err = pdf.Rectangle(XCache, *line, (XCache + widthRect), (float64(marginY) + float64(fontSize) + *line), "D", 0, 0)
				if err != nil {
					return err
				}
				XCache += widthRect
			}
		}

		*line += float64(marginY) + float64(fontSize)
		if *line > paper.H {
			pdf.AddPage()
			*line = 0
		}

		for _, SOline := range SOlines {
			XCache := globalpadding.Left
			maxLine := 0
			for i := 0; i < t.NumField(); i++ {
				name := t.Field(i).Name
				value := ""
				switch name {
				case "Item":
					if SOline.Item == "" {
						value = SOline.Asset
					} else {
						value = SOline.Item
					}
					break

				case "Description":
					value = SOline.Description
					break

				case "Qty":
					value = strconv.Itoa(int(SOline.Qty))
					break

				case "Uom":
					value = SOline.UoM
					break

				case "ContractPeriod":
					value = strconv.Itoa(SOline.ContractPeriod)
					break

				case "UnitPrice":
					ac := accounting.Accounting{Symbol: "Rp", Precision: 0, Thousand: ".", Decimal: ","}
					value = ac.FormatMoney(SOline.UnitPrice)
					// value = strconv.Itoa(int(SOline.UnitPrice))
					break

				case "Amount":
					ac := accounting.Accounting{Symbol: "Rp", Precision: 0, Thousand: ".", Decimal: ","}
					value = ac.FormatMoney(SOline.Amount)
					// value = strconv.Itoa(int(SOline.Amount))
					break

				default:
					break
				}

				if v.Field(i).Interface().(float64) > 0 {
					addline := 0
					err := wraptext(func(textwrap string) error {
						// fmt.Println((*line+2)*float64(addline*fontSize), textwrap)
						liner := float64(addline * fontSize)
						if liner == 0 {
							pdf.SetXY(XCache+Padding.Left, (*line + Padding.Top))
						} else {
							pdf.SetXY(XCache+Padding.Left, (*line+Padding.Top)*float64(addline*fontSize))
						}
						err := pdf.Cell(nil, textwrap)
						if err != nil {
							return err
						}
						addline++
						return nil
					}, value, pdf, int(v.Field(i).Interface().(float64)/float64(fontSize)), nil)
					if err != nil {
						return err
					}

					if addline > maxLine {
						maxLine = addline
					}

					XCache += v.Field(i).Interface().(float64)
				} else {
					addline := 0
					err := wraptext(func(textwrap string) error {
						// fmt.Println((*line+2)*float64(addline*fontSize), textwrap)
						liner := float64(addline * fontSize)
						if liner == 0 {
							pdf.SetXY(XCache+Padding.Left, (*line + Padding.Top))
						} else {
							pdf.SetXY(XCache+Padding.Left, (*line+Padding.Top)*float64(addline*fontSize))
						}
						err := pdf.Cell(nil, textwrap)
						if err != nil {
							return err
						}
						addline++
						return nil
					}, value, pdf, int(widthRect/float64(fontSize)), nil)
					if err != nil {
						return err
					}

					if addline > maxLine {
						maxLine = addline
					}

					XCache += widthRect
				}
			}

			XCache = globalpadding.Left
			for i := 0; i < t.NumField(); i++ {
				pdf.SetStrokeColor(0, 0, 0)
				pdf.SetLineWidth(1)

				// points := []gopdf.Point{}
				// points = append(points, gopdf.Point{X: XCache, Y: *line})
				// // points = append(points, gopdf.Point{X: (XCache + widthRect), Y: *line})
				// points = append(points, gopdf.Point{X: (XCache + widthRect), Y: (float64(marginY) + float64(fontSize*maxLine) + *line)})
				// points = append(points, gopdf.Point{X: XCache, Y: (float64(marginY) + float64(fontSize*maxLine) + *line)})
				// pdf.Polygon(points, "D")

				if v.Field(i).Interface().(float64) > 0 {
					err := pdf.Rectangle(XCache, *line, (XCache + v.Field(i).Interface().(float64)), (float64(marginY) + float64(fontSize*maxLine) + *line), "D", 0, 0)
					if err != nil {
						return err
					}
					XCache += v.Field(i).Interface().(float64)

				} else {
					err := pdf.Rectangle(XCache, *line, (XCache + widthRect), (float64(marginY) + float64(fontSize*maxLine) + *line), "D", 0, 0)
					if err != nil {
						return err
					}
					XCache += widthRect
				}
			}

			*line += float64(marginY) + float64(maxLine*fontSize) + 1
			if *line > paper.H {
				pdf.AddPage()
				*line = 0
			}
		}
	}

	return nil
}

// func (o *SalesOrderEngine) InsertData(ctx *kaos.Context, id string) (interface{}, error) {
// 	h := sebar.GetTenantDBFromContext(ctx)
// 	if h == nil {
// 		return nil, errors.New("missing: connection")
// 	}

// 	f, err := excelize.OpenFile("import.xlsx")
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer func() {
// 		// Close the spreadsheet.
// 		if err := f.Close(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()
// 	// // Get value from cell by given worksheet name and cell reference.
// 	// cell, err := f.GetCellValue("Sheet1", "B2")
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println(cell)
// 	// Get all the rows in the Sheet1.
// 	rows, err := f.GetRows("FIX BGT")
// 	if err != nil {
// 		return nil, err
// 	}

// 	skiplines := 3

// 	salesorders := []sdpmodel.SalesOrder{}
// 	unitcalendars := []sdpmodel.UnitCalendar{}
// 	for indexrow, row := range rows {
// 		if indexrow < skiplines || len(row) < 28 {
// 			continue
// 		}

// 		skip := false
// 		for _, column := range row {
// 			if column == "#N/A" {
// 				skip = true
// 				break
// 			}
// 		}

// 		if skip {
// 			continue
// 		}

// 		if !containstructs(salesorders, row[10], row[7]) {
// 			names := []string{}
// 			Lines := []struct {
// 				Asset          string
// 				Item           string
// 				Description    string
// 				Shift          int64
// 				Qty            uint
// 				UoM            string
// 				ContractPeriod int
// 				UnitPrice      int64
// 				Amount         int64
// 				DiscountType   string
// 				Discount       int
// 				Taxable        bool
// 				TaxCodes       []string
// 				Spesifications []string
// 				StartDate      time.Time
// 				EndDate        time.Time
// 			}{}

// 			var subtotalall float64 = 0.0
// 			var totalamountall float64 = 0.0

// 			for indexcheckrow, checkrow := range rows {
// 				if indexcheckrow < skiplines || len(checkrow) <= 27 {
// 					continue
// 				}

// 				skip := false
// 				for _, column := range checkrow {
// 					if column == "#N/A" {
// 						skip = true
// 						break
// 					}
// 				}

// 				if skip {
// 					continue
// 				}

// 				if checkrow[10] == row[10] && row[7] == checkrow[7] && !containstrings(names, checkrow[26]) {
// 					names = append(names, checkrow[26])
// 				}

// 				if row[10] == checkrow[10] && row[7] == checkrow[7] {
// 					qty, err := strconv.Atoi(strings.ReplaceAll(checkrow[25], ",", ""))
// 					if err != nil {
// 						return nil, err
// 					}

// 					akhirjalan, err := time.Parse("2006-01-02", checkrow[19])
// 					if err != nil {
// 						return nil, err
// 					}

// 					contractperiod, err := strconv.Atoi(strings.ReplaceAll(checkrow[23], ",", ""))
// 					if err != nil {
// 						return nil, err
// 					}

// 					unitprice, err := strconv.Atoi(strings.ReplaceAll(checkrow[24], ",", ""))
// 					if err != nil {
// 						return nil, err
// 					}

// 					awaljalan, err := time.Parse("2006-01-02", checkrow[18])
// 					if err != nil {
// 						return nil, err
// 					}

// 					totalamount, err := strconv.ParseFloat(strings.ReplaceAll(checkrow[17], ",", ""), 64)
// 					if err != nil {
// 						return nil, err
// 					}

// 					totalamountall += totalamount

// 					subtotalamount, err := strconv.ParseFloat(strings.ReplaceAll(checkrow[16], ",", ""), 64)
// 					if err != nil {
// 						return nil, err
// 					}

// 					subtotalall += subtotalamount

// 					Lines = append(Lines, struct {
// 						Asset          string
// 						Item           string
// 						Description    string
// 						Shift          int64
// 						Qty            uint
// 						UoM            string
// 						ContractPeriod int
// 						UnitPrice      int64
// 						Amount         int64
// 						DiscountType   string
// 						Discount       int
// 						Taxable        bool
// 						TaxCodes       []string
// 						Spesifications []string
// 						StartDate      time.Time
// 						EndDate        time.Time
// 					}{
// 						Asset:          checkrow[14],
// 						Description:    checkrow[12],
// 						Qty:            uint(qty),
// 						UoM:            checkrow[27],
// 						ContractPeriod: contractperiod,
// 						UnitPrice:      int64(unitprice),
// 						Amount:         int64(subtotalamount),
// 						DiscountType:   "fixed",
// 						Discount:       0,
// 						Shift:          0,
// 						StartDate:      awaljalan,
// 						EndDate:        akhirjalan,
// 					})
// 				}
// 			}

// 			awaljalan, err := time.Parse("2006-01-02", row[18])
// 			if err != nil {
// 				return nil, err
// 			}

// 			salesorders = append(salesorders, sdpmodel.SalesOrder{
// 				Name: strings.Join(names, " - "),
// 				// WarehouseID: "",
// 				Charges:       []sdpmodel.SalesCharge{},
// 				Lines:         Lines,
// 				Editor:        "",
// 				BreakdownCost: []sdpmodel.SalesOrderBreakdownCost{},
// 				ManPower:      []sdpmodel.SalesOrderManPower{},
// 				// SalesOpportunityRefNo: "",
// 				// SaleSOuotationRefNo:   "",
// 				SalesOrderNo:   row[10],
// 				SpkNo:          row[11],
// 				SalesOrderDate: awaljalan,
// 				// SalesPriceBookID:      "",
// 				TaxCodes: []string{},
// 				// JournalTypeID:         "",
// 				// PostingProfileID:      "",
// 				// CompanyID:             "",
// 				// Notes:                 "",
// 				CustomerID:     row[2],
// 				TotalAmount:    totalamountall,
// 				SubTotalAmount: subtotalall,
// 				DiscountAmount: 0,
// 				TaxAmount:      0,
// 				Dimension: []tenantcoremodel.DimensionItem{
// 					{
// 						Key:   "Site",
// 						Value: row[7],
// 					},
// 					{
// 						Key:   "PC",
// 						Value: row[8],
// 					},
// 					{
// 						Key:   "CC",
// 						Value: row[9],
// 					},
// 				},
// 				Status:     "DRAFT",
// 				Year:       awaljalan.Year(),
// 				Created:    time.Time{},
// 				LastUpdate: time.Time{},
// 			})
// 		}

// 		// unit calendar
// 		if !containUCstructs(unitcalendars, row[10], row[7]) {
// 			lines := []struct {
// 				Index        uint32
// 				AssetUnitID  string
// 				IsItem       bool
// 				StartDate    time.Time
// 				EndDate      time.Time
// 				Uom          string
// 				Duration     uint32
// 				Qty          uint32
// 				Descriptions string
// 			}{}

// 			indexlines := uint32(0)
// 			for indexcheckrow, checkrow := range rows {
// 				if indexcheckrow < skiplines || len(checkrow) < 28 {
// 					continue
// 				}

// 				skip := false
// 				for _, column := range checkrow {
// 					if column == "#N/A" {
// 						skip = true
// 						break
// 					}
// 				}

// 				if skip {
// 					continue
// 				}

// 				if row[10] == checkrow[10] && row[6] == checkrow[6] {
// 					akhirjalan, err := time.Parse("2006-01-02", checkrow[19])
// 					if err != nil {
// 						return nil, err
// 					}

// 					awaljalan, err := time.Parse("2006-01-02", checkrow[18])
// 					if err != nil {
// 						return nil, err
// 					}

// 					contractperiod, err := strconv.Atoi(strings.ReplaceAll(checkrow[23], ",", ""))
// 					if err != nil {
// 						return nil, err
// 					}

// 					qty, err := strconv.Atoi(strings.ReplaceAll(checkrow[25], ",", ""))
// 					if err != nil {
// 						return nil, err
// 					}

// 					lines = append(lines, struct {
// 						Index        uint32
// 						AssetUnitID  string
// 						IsItem       bool
// 						StartDate    time.Time
// 						EndDate      time.Time
// 						Uom          string
// 						Duration     uint32
// 						Qty          uint32
// 						Descriptions string
// 					}{
// 						Index:        indexlines,
// 						AssetUnitID:  checkrow[14],
// 						IsItem:       false,
// 						StartDate:    awaljalan,
// 						EndDate:      akhirjalan,
// 						Uom:          checkrow[27],
// 						Duration:     uint32(contractperiod),
// 						Qty:          uint32(qty),
// 						Descriptions: "",
// 					})
// 					indexlines++
// 				}
// 			}

// 			unitcalendars = append(unitcalendars, sdpmodel.UnitCalendar{
// 				SORefNo:   row[10],
// 				Lines:     lines,
// 				Customer:  row[2],
// 				ProjectID: row[5],
// 				Remark:    "",
// 				Dimension: []tenantcoremodel.DimensionItem{
// 					{
// 						Key:   "Site",
// 						Value: row[7],
// 					},
// 					{
// 						Key:   "PC",
// 						Value: row[8],
// 					},
// 					{
// 						Key:   "CC",
// 						Value: row[9],
// 					},
// 				},
// 				Created:    time.Time{},
// 				LastUpdate: time.Time{},
// 			})
// 		}

// 		// for indexcol, colCell := range row {
// 		// 	if !containstructs(salesorders, colCell) && indexcol == 10 {
// 		// 		SalesOrderNo = append(SalesOrderNo, colCell)
// 		// 	}
// 		// }
// 	}

// 	var wg sync.WaitGroup
// 	var mtx sync.Mutex

// 	for index, salesorder := range salesorders {
// 		wg.Add(1)
// 		go func(index int, salesorder sdpmodel.SalesOrder) {
// 			mtx.Lock()
// 			defer mtx.Unlock()
// 			defer wg.Done()
// 			errs := h.Save(&salesorder)
// 			if errs != nil {
// 				err = errs
// 				return
// 			}

// 			salesorders[index] = salesorder
// 			return
// 		}(index, salesorder)
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, unitcalendar := range unitcalendars {
// 		wg.Add(1)
// 		go func(unitcalendar sdpmodel.UnitCalendar) {
// 			mtx.Lock()
// 			defer mtx.Unlock()
// 			defer wg.Done()
// 			salesorder := sdpmodel.SalesOrder{}
// 			for _, so := range salesorders {
// 				if unitcalendar.SORefNo == so.SalesOrderNo {
// 					salesorder = so
// 					break
// 				}
// 			}

// 			unitcalendar.SORefNo = salesorder.ID

// 			o.UC.Insert(ctx, &unitcalendar)
// 			return
// 		}(unitcalendar)
// 	}
// 	wg.Wait()

// 	return salesorders, nil
// }

// func containstructs(s []sdpmodel.SalesOrder, SO string, site string) bool {
// 	for _, a := range s {
// 		if a.SalesOrderNo == SO && a.Dimension.Get("Site") == site {
// 			return true
// 		}
// 	}
// 	return false
// }

// func containstrings(s []string, e string) bool {
// 	for _, a := range s {
// 		if a == e {
// 			return true
// 		}
// 	}
// 	return false
// }

// func containUCstructs(s []sdpmodel.UnitCalendar, SO string, site string) bool {
// 	for _, a := range s {
// 		if a.SORefNo == SO && a.Dimension.Get("Site") == site {
// 			return true
// 		}
// 	}
// 	return false
// }
