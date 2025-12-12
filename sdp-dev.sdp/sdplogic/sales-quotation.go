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
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/kasset"
	"github.com/leekchan/accounting"
	"github.com/sebarcode/codekit"
	"github.com/signintech/gopdf"
	"golang.org/x/net/html"
)

type SalesQuotationEngine struct{}

func (o *SalesQuotationEngine) upsert(h *datahub.Hub, model *sdpmodel.SalesQuotation) error {
	if e := h.GetByID(new(sdpmodel.SalesQuotation), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Sales Order : " + e.Error())
		}
	} else {
		if e := h.Update(model); e != nil {
			return errors.New("error update Sales Order : " + e.Error())
		}
	}

	return nil
}

type PayloadActionSQ struct {
	ID string `json:"_id"`
}

func (o *SalesQuotationEngine) ActionSQ(ctx *kaos.Context, payload *PayloadActionSQ) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	if payload.ID == "" {
		return nil, errors.New("missing: payload")
	}

	//get quotation in
	Opportunity := new(sdpmodel.SalesOpportunity)
	e := h.GetByID(Opportunity, payload.ID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Error Opportunity: %v Missing opportunity in by ID: %s", e, payload.ID))
	}

	customer := new(tenantcoremodel.Customer)
	err = h.GetByID(customer, Opportunity.Customer)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error Customer: %v", err))
	}

	// customersplit := strings.Split(strings.ReplaceAll(customer.Name, ", PT", ""), " ")
	// cs := ""
	// for _, custspl := range customersplit {
	// 	splits := strings.Split(custspl, "")
	// 	if len(splits) > 0 {
	// 		cs += splits[0]
	// 	}
	// }
	cs := customer.CustomerAlias
	if cs == "" {
		return nil, errors.New("Customer alias not null")
	}

	bagongcustomer := map[string]any{}
	err = event.Publish("/v1/bagong/customer/get", []string{Opportunity.Customer}, &bagongcustomer, nil)
	if err != nil {
		return kasset.Asset{}, err
	}

	taxcodes := []any{bagongcustomer["Detail"].(map[string]any)["Tax1"], bagongcustomer["Detail"].(map[string]any)["Tax2"]}

	OpporQuota := []sdpmodel.SalesQuotation{}
	e = h.Gets(new(sdpmodel.SalesQuotation), dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("Customer", Opportunity.Customer), dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)), dbflex.Eq("OpportunityNo", payload.ID))).SetSort("-Rev").SetTake(1), &OpporQuota)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Missing opportunity in by ID: %s", e, payload.ID))
	}

	last := []sdpmodel.SalesQuotation{}
	e = h.Gets(new(sdpmodel.SalesQuotation), dbflex.NewQueryParam().SetWhere(dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC))).SetSort("-No").SetTake(1), &last)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Last sales quotation", e))
	}

	sq := new(sdpmodel.SalesQuotation)
	if len(OpporQuota) > 0 {
		sq.No = OpporQuota[0].No
		sq.Rev = OpporQuota[0].Rev + 1
		sq.QuotationNo = fmt.Sprintf("SQ/%04d/BDM-HO-%s/%02d/%d-rev%d", sq.No, cs, int(time.Now().Month()), time.Now().Year(), sq.Rev)
	} else if len(last) > 0 {
		sq.No = last[0].No + 1
		sq.QuotationNo = fmt.Sprintf("SQ/%04d/BDM-HO-%s/%02d/%d", sq.No, cs, int(time.Now().Month()), time.Now().Year())
	} else {
		sq.No = 1
		sq.QuotationNo = fmt.Sprintf("SQ/%04d/BDM-HO-%s/%02d/%d", sq.No, cs, int(time.Now().Month()), time.Now().Year())
	}

	sq.OpportunityNo = Opportunity.ID
	sq.QuotationDate = time.Now()
	sq.Customer = Opportunity.Customer
	sq.TransactionType = Opportunity.TransactionType
	sq.Dimension = Opportunity.Dimension

	taxable := false
	if taxcodes != nil {
		taxable = true
		for _, taxcode := range taxcodes {
			sq.TaxCodes = append(sq.TaxCodes, taxcode.(string))
			// fmt.Printf("%v", item.(map[string]interface{})["title"])
		}
	}

	for _, line := range Opportunity.Lines {
		sq.Lines = append(sq.Lines, struct {
			Asset          string
			Item           string
			Description    string
			Shift          int64
			Qty            uint
			UoM            string
			ContractPeriod int
			UnitPrice      int64
			Amount         int64
			DiscountType   sdpmodel.DiscountTypeSQ
			Discount       int
			Taxable        bool
			TaxCodes       []string
			Spesifications []string
		}{
			Asset:          line.Asset,
			Item:           line.Item,
			Description:    line.Description,
			UoM:            line.Uom,
			ContractPeriod: line.ContractPeriod,
			Qty:            line.Qty,
			Spesifications: line.Spesifications,
			Taxable:        taxable,
			TaxCodes:       sq.TaxCodes,
		})
	}

	e = o.upsert(h, sq)
	if e != nil {
		return nil, e
	}

	return sq, nil
}

type SaveSalesQuotation struct {
	sdpmodel.SalesQuotation
	UploadLetterHeadAsset kasset.AssetDataBase64
	UploadFooterAsset     kasset.AssetDataBase64
}

func (o *SalesQuotationEngine) Insert(ctx *kaos.Context, payload *SaveSalesQuotation) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Customer == "" {
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

	sq := sdpmodel.SalesQuotation{
		LetterHeadFirst:      payload.LetterHeadFirst,
		FooterLastPage:       payload.FooterLastPage,
		Editor:               payload.Editor,
		No:                   0,
		Rev:                  0,
		OpportunityNo:        payload.OpportunityNo,
		QuotationNo:          payload.OpportunityNo,
		QuotationDate:        time.Now(),
		QuotationName:        payload.QuotationName,
		SalesPriceBook:       payload.SalesPriceBook,
		Customer:             payload.Customer,
		TotalAmount:          payload.TotalAmount,
		SubTotalAmount:       payload.SubTotalAmount,
		DiscountAmount:       payload.DiscountAmount,
		TaxAmount:            payload.TaxAmount,
		TaxCodes:             payload.TaxCodes,
		Lines:                payload.Lines,
		Dimension:            payload.Dimension,
		JournalType:          payload.JournalType,
		TransactionType:      payload.TransactionType,
		PostingProfileID:     payload.PostingProfileID,
		CompanyID:            payload.CompanyID,
		HeaderDiscountValue:  payload.HeaderDiscountValue,
		HeaderDiscountAmount: payload.HeaderDiscountAmount,
		HeaderDiscountType:   payload.HeaderDiscountType,
		TrxDate:              time.Now(),
		Text:                 "",
	}

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
		sq.LetterHeadAsset = reply.ID
	}

	if payload.UploadFooterAsset.Content != "" {
		reply, err := writeAsset(payload.UploadFooterAsset)
		if err != nil {
			return nil, err
		}

		sq.FooterAsset = reply.ID
	}

	customer := new(tenantcoremodel.Customer)
	err = h.GetByID(customer, sq.Customer)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v", err))
	}

	// customersplit := strings.Split(strings.ReplaceAll(customer.Name, ", PT", ""), " ")
	// cs := ""
	// for _, custspl := range customersplit {
	// 	cs += strings.Split(custspl, "")[0]
	// }

	cs := customer.CustomerAlias
	if cs == "" {
		return nil, errors.New("Customer alias not null")
	}

	last := []sdpmodel.SalesQuotation{}
	err = h.Gets(new(sdpmodel.SalesQuotation), dbflex.NewQueryParam().SetWhere(dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC))).SetSort("-No").SetTake(1), &last)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Last sales quotation", err))
	}

	if len(last) > 0 {
		sq.No = last[0].No + 1
		sq.QuotationNo = fmt.Sprintf("SQ/%04d/BDM-HO-%s/%02d/%d", sq.No, cs, (int(time.Now().Month())), time.Now().Year())
	} else {
		sq.No = 1
		sq.QuotationNo = fmt.Sprintf("SQ/%04d/BDM-HO-%s/%02d/%d", sq.No, cs, (int(time.Now().Month())), time.Now().Year())
	}

	err = o.upsert(h, &sq)
	if err != nil {
		return nil, err
	}

	return sq, nil
}

func (o *SalesQuotationEngine) Update(ctx *kaos.Context, payload *SaveSalesQuotation) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Customer == "" {
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

	sq := sdpmodel.SalesQuotation{
		ID:                   payload.ID,
		LetterHeadFirst:      payload.LetterHeadFirst,
		FooterLastPage:       payload.FooterLastPage,
		Editor:               payload.Editor,
		OpportunityNo:        payload.OpportunityNo,
		QuotationName:        payload.QuotationName,
		SalesPriceBook:       payload.SalesPriceBook,
		Customer:             payload.Customer,
		TotalAmount:          payload.TotalAmount,
		SubTotalAmount:       payload.SubTotalAmount,
		DiscountAmount:       payload.DiscountAmount,
		TaxAmount:            payload.TaxAmount,
		TaxCodes:             payload.TaxCodes,
		Lines:                payload.Lines,
		Dimension:            payload.Dimension,
		JournalType:          payload.JournalType,
		TransactionType:      payload.TransactionType,
		PostingProfileID:     payload.PostingProfileID,
		No:                   payload.No,
		QuotationNo:          payload.QuotationNo,
		QuotationDate:        payload.QuotationDate,
		Rev:                  payload.Rev,
		Status:               payload.Status,
		LetterHeadAsset:      payload.LetterHeadAsset,
		FooterAsset:          payload.FooterAsset,
		CompanyID:            payload.CompanyID,
		HeaderDiscountValue:  payload.HeaderDiscountValue,
		HeaderDiscountAmount: payload.HeaderDiscountAmount,
		HeaderDiscountType:   payload.HeaderDiscountType,
		TrxDate:              time.Now(),
		Text:                 "",
	}

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
		sq.LetterHeadAsset = reply.ID
	}

	if payload.UploadFooterAsset.Content != "" {
		reply, err := writeAsset(payload.UploadFooterAsset)
		if err != nil {
			return nil, err
		}

		sq.FooterAsset = reply.ID
	}

	err = o.upsert(h, &sq)
	if err != nil {
		return nil, err
	}

	return sq, nil
}

type PayloadPrintPDF struct {
	ID        string `json:"_id"`
	PublicURL string `json:"public_url"`
}

func (o *SalesQuotationEngine) PrintPDF(ctx *kaos.Context, payload *PayloadPrintPDF) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" {
		return nil, errors.New("Missing: Payload")
	}

	SQ := sdpmodel.SalesQuotation{}
	err := h.GetByID(&SQ, payload.ID)
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
	ev.Publish("/v1/bagong/customer/get", []string{SQ.Customer}, &reply, &kaos.PublishOpts{Headers: codekit.M{sebar.CtxJWTReferenceID: userID}})

	id, err := printpdf(ctx, payload, &SQ, reply)
	if err != nil {
		return nil, err
	}

	return id, nil

	// msgTemplate := ksmsg.SendTemplateRequest{
	// 	TemplateName: SQ.QuotationName,
	// 	LanguageID:   "ID-id",
	// 	Message: &kmsg.Message{
	// 		Kind:   SQ.QuotationName,
	// 		Method: "SMTP",
	// 		To:     user.Email,
	// 	},
	// 	Data: codekit.M{
	// 		"DisplayName":       user.DisplayName,
	// 		"ResetPasswordLink": linkActivation,
	// 		"Token":             token.Token,
	// 	},
	// }

	// ev, e := ctx.DefaultEvent()
	// if e != nil {
	// 	return "", e
	// }
	// resp := 0
	// e = ev.Publish("/v1/msg/send-template", &msgTemplate, &resp, nil)
	// if e != nil {
	// 	return "NOK", e
	// }

	// SQ.SendEmail = true

	// err = o.upsert(h, &SQ)
	// if err != nil {
	// 	return nil, err
	// }

	return SQ, nil
}

type PayloadSendEmail struct {
	ID        string `json:"_id"`
	PublicURL string `json:"public_url"`
}

func (o *SalesQuotationEngine) SendEmail(ctx *kaos.Context, payload *PayloadSendEmail) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" {
		return nil, errors.New("Missing: Payload")
	}

	SQ := sdpmodel.SalesQuotation{}
	err := h.GetByID(&SQ, payload.ID)
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
	ev.Publish("/v1/bagong/customer/get", []string{SQ.Customer}, &reply, &kaos.PublishOpts{Headers: codekit.M{sebar.CtxJWTReferenceID: userID}})

	payloadpdf := PayloadPrintPDF{
		ID:        payload.ID,
		PublicURL: payload.PublicURL,
	}

	id, err := printpdf(ctx, &payloadpdf, &SQ, reply)
	if err != nil {
		return nil, err
	}

	return id, nil

	// msgTemplate := ksmsg.SendTemplateRequest{
	// 	TemplateName: SQ.QuotationName,
	// 	LanguageID:   "ID-id",
	// 	Message: &kmsg.Message{
	// 		Kind:   SQ.QuotationName,
	// 		Method: "SMTP",
	// 		To:     user.Email,
	// 	},
	// 	Data: codekit.M{
	// 		"DisplayName":       user.DisplayName,
	// 		"ResetPasswordLink": linkActivation,
	// 		"Token":             token.Token,
	// 	},
	// }

	// ev, e := ctx.DefaultEvent()
	// if e != nil {
	// 	return "", e
	// }
	// resp := 0
	// e = ev.Publish("/v1/msg/send-template", &msgTemplate, &resp, nil)
	// if e != nil {
	// 	return "NOK", e
	// }

	// SQ.SendEmail = true

	// err = o.upsert(h, &SQ)
	// if err != nil {
	// 	return nil, err
	// }

}

func printpdf(ctx *kaos.Context, payload *PayloadPrintPDF, SQ *sdpmodel.SalesQuotation, customer map[string]any) (string, error) {

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

		if SQ.LetterHeadAsset == "" {
			return
		}

		urlasset := payload.PublicURL + "/asset/view?id=" + SQ.LetterHeadAsset
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

		if SQ.FooterAsset == "" {
			return
		}

		urlasset := payload.PublicURL + "/asset/view?id=" + SQ.FooterAsset
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

	if SQ.LetterHeadFirst {
		if SQ.LetterHeadAsset == "" || len(headerData) <= 0 {
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

		// fmt.Println(float64(im.Width))
		// width := (float64(im.Width) - (float64(im.Width) * 67.5 / 100)) - (Padding.Left + Padding.Right)
		// if float64(im.Width) < paperSize.W {
		// 	width = float64(im.Width)
		// }
		// height := float64(im.Height) - (float64(im.Height) * 67.5 / 100)
		// if float64(im.Height) < paperSize.H {
		// 	height = float64(im.Height)
		// }
		width := (paperSize.W - (Padding.Left + Padding.Right))
		height := (width / float64(im.Width)) * float64(im.Height)

		err = pdf.ImageByHolder(imgheader, Padding.Left, Padding.Top, &gopdf.Rect{
			W: width, // 67.5 is percentage
			H: height,
		})

		if err != nil {
			err = errors.New(err.Error() + ". insert PDF")
			return "", err
		}

		line = 106.7 + Padding.Top
	}

	// No SQ
	pdf.SetXY(Padding.Left, line)
	line += fontSize + spacing
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Nomer: "+SQ.QuotationNo)
	if err != nil {
		return "", err
	}

	// Perihal
	pdf.SetXY(Padding.Left, line)
	line += fontSize + spacing
	if line > paperSize.H {
		pdf.AddPage()
		line = 0
	}
	err = pdf.Cell(nil, "Perihal: "+SQ.QuotationName)
	if err != nil {
		return "", err
	}

	// Date
	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-100, line)
	line += (fontSize + spacing) * 2
	if line > paperSize.H {

		pdf.AddPage()
		line = 0

	}
	err = pdf.Cell(nil, "Malang, "+SQ.QuotationDate.Format("15-01-2006"))
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
					line = Padding.Bottom
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

	// detail mail
	pdf.SetXY(Padding.Left, line)
	line += fontSize + spacing
	if line > paperSize.H {

		pdf.AddPage()
		line = 0

	}
	err = pdf.Cell(nil, "Dengan Hormat")
	if err != nil {
		return "", err
	}

	pdf.SetXY(Padding.Left, line)
	line += fontSize + spacing
	if line > paperSize.H {

		pdf.AddPage()
		line = 0

	}
	err = pdf.Cell(nil, "Bersama dengan ini kami sampaikan "+SQ.QuotationName+" dengan rincian sebagai berikut:")
	if err != nil {
		return "", err
	}

	Items := []string{}
	Assets := []string{}
	for _, line := range SQ.Lines {
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
			for index, line := range SQ.Lines {
				if AssetTenant["_id"] == line.Asset {
					line.Asset = AssetTenant["Name"].(string)
					SQ.Lines[index] = line
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
			for index, line := range SQ.Lines {
				if ItemTenant["_id"] == line.Item {
					line.Item = ItemTenant["Name"].(string)
					SQ.Lines[index] = line
				}
			}
		}
	}

	err = tablepdf(sdpmodel.SalesQuotationLinePreviewGrid{
		Qty:  25,
		Item: 129,
	}, SQ.Lines, &line, fontSize, paperSize, &pdf, Padding, struct {
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

	err = pdf.Cell(nil, ac.FormatMoney(SQ.SubTotalAmount))
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
	err = pdf.Cell(nil, ac.FormatMoney(SQ.TaxAmount))
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
	err = pdf.Cell(nil, ac.FormatMoney(SQ.DiscountAmount))
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-160, line+2)
	err = pdf.Cell(nil, "Total Amount ")
	if err != nil {
		return "", err
	}

	pdf.SetXY((paperSize.W-(Padding.Left+Padding.Right))-70, line+2)
	err = pdf.Cell(nil, ac.FormatMoney(SQ.TotalAmount))
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

	err = htmlpdf(SQ.Editor, &pdf, &line, fontSize, spacing, paperSize, Padding)
	if err != nil {
		return "", err
	}

	if SQ.FooterLastPage {
		if SQ.FooterAsset == "" || len(footerData) <= 0 {
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

		// width := (float64(im.Width) - (float64(im.Width) * 67.5 / 100)) - (Padding.Left + Padding.Right)
		// if float64(im.Width) < paperSize.W {
		// 	width = float64(im.Width)
		// }
		width := (paperSize.W - (Padding.Left + Padding.Right))
		height := (width / float64(im.Width)) * float64(im.Height)
		// height := float64(im.Height) - (float64(im.Height) * 67.5 / 100)

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
			OriginalFileName: "PDF-" + SQ.ID,
			Kind:             "Sales Quotation",
			RefID:            SQ.ID,
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

func htmlpdf(text string, pdf *gopdf.GoPdf, line *float64, fontSize int, spacing int, paper *gopdf.Rect, padding struct {
	Top    float64
	Left   float64
	Right  float64
	Bottom float64
}) error {
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	root, err := html.Parse(strings.NewReader(text))
	if err != nil {
		return err
	}

	x := padding.Left
	err = replaceHTML(root, pdf, line, &x, fontSize, spacing, paper, padding, nil)
	if err != nil {
		return err
	}

	return nil
}

func tablepdf(model any, SQlines []struct {
	Asset          string
	Item           string
	Description    string
	Shift          int64
	Qty            uint
	UoM            string
	ContractPeriod int
	UnitPrice      int64
	Amount         int64
	DiscountType   sdpmodel.DiscountTypeSQ
	Discount       int
	Taxable        bool
	TaxCodes       []string
	Spesifications []string
}, line *float64, fontSize int, paper *gopdf.Rect, pdf *gopdf.GoPdf, globalpadding struct {
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
			*line = Padding.Top

		}

		for _, sqline := range SQlines {
			XCache := globalpadding.Left
			maxLine := 0
			for i := 0; i < t.NumField(); i++ {
				name := t.Field(i).Name
				value := ""
				switch name {
				case "Item":
					if sqline.Item == "" {
						value = sqline.Asset
					} else {
						value = sqline.Item
					}
					break

				case "Description":
					value = sqline.Description
					break

				case "Qty":
					value = strconv.Itoa(int(sqline.Qty))
					break

				case "Uom":
					value = sqline.UoM
					break

				case "ContractPeriod":
					value = strconv.Itoa(sqline.ContractPeriod)
					break

				case "UnitPrice":
					ac := accounting.Accounting{Symbol: "Rp ", Precision: 0, Thousand: ".", Decimal: ","}
					value = ac.FormatMoney(sqline.UnitPrice)
					// value = strconv.Itoa(int(sqline.UnitPrice))
					break

				case "Amount":
					ac := accounting.Accounting{Symbol: "Rp ", Precision: 0, Thousand: ".", Decimal: ","}
					value = ac.FormatMoney(sqline.Amount)
					// value = strconv.Itoa(int(sqline.Amount))
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
				*line = Padding.Top

			}
		}

	}

	return nil
}

func wraptext(fn func(textwrap string) error, text string, pdf *gopdf.GoPdf, textsize int, startx *float64) error {
	textsplits := strings.Split(text, " ")
	cText := ""

	x := 0.0
	if startx != nil {
		x = *startx
	}
	for index, split := range textsplits {
		indexx := float64(index)
		size, _ := pdf.MeasureTextWidth(cText + split)
		// fmt.Println(cText + split, size)
		if size+x > float64(textsize) || indexx == float64(len(textsplits)-1) {
			if indexx == float64(len(textsplits)-1) {
				cText += split + " "
			}

			if len(cText) > 0 {
				err := fn(cText[:len(cText)-len(" ")])
				if err != nil {
					return err
				}
				x = 0.0
			}
			cText = split + " "
		} else {
			cText += split + " "
		}
	}

	return nil
}

func replaceHTML(n *html.Node, pdf *gopdf.GoPdf, line *float64, x *float64, fontSize int, spacing int, paper *gopdf.Rect, padding struct {
	Top    float64
	Left   float64
	Right  float64
	Bottom float64
}, cache map[string]any) error {

	tag := ""
	if n.Type == html.ElementNode {
		tag = n.Data
	}

	text := ""
	if n.Type == html.TextNode {
		text = n.Data
	}

	newline := false
	switch tag {
	case "ul":
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			err := replaceHTML(child, pdf, line, x, fontSize, spacing, paper, struct {
				Top    float64
				Left   float64
				Right  float64
				Bottom float64
			}{
				Top:    padding.Top,
				Left:   padding.Left + 10,
				Right:  padding.Right,
				Bottom: padding.Bottom,
			}, map[string]any{
				"head": "ul",
			})
			if err != nil {
				return err
			}
		}
		return nil

	case "ol":
		index := 0
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			err := replaceHTML(child, pdf, line, x, fontSize, spacing, paper, struct {
				Top    float64
				Left   float64
				Right  float64
				Bottom float64
			}{
				Top:    padding.Top,
				Left:   padding.Left + 10,
				Right:  padding.Right,
				Bottom: padding.Bottom,
			}, map[string]any{
				"index": index,
			})
			if err != nil {
				return err
			}
			index++
		}
		return nil

	case "li":
		val, ok := cache["index"]
		if ok {
			text = strconv.Itoa(val.(int)+1) + ". " + text
		}

		val, ok = cache["head"]
		if ok {
			text = "• " + text
		}

		newline = true
		break

	case "p":
		newline = true
		break

	case "strong":
		err := pdf.SetFont("times-new-roman", "B", fontSize)
		if err != nil {
			return err
		}
		break

	case "b":
		err := pdf.SetFont("times-new-roman", "B", fontSize)
		if err != nil {
			return err
		}
		break

	case "italic":
		err := pdf.SetFont("times-new-roman", "I", fontSize)
		if err != nil {
			return err
		}
		break

	case "i":
		err := pdf.SetFont("times-new-roman", "I", fontSize)
		if err != nil {
			return err
		}
		break

	case "em":
		err := pdf.SetFont("times-new-roman", "I", fontSize)
		if err != nil {
			return err
		}
		break

	case "u":
		err := pdf.SetFont("times-new-roman", "U", fontSize)
		if err != nil {
			return err
		}
		break

	case "h1":
		err := pdf.SetFont("times-new-roman", "B", 34)
		if err != nil {
			return err
		}
		newline = true
		break

	case "h2":
		err := pdf.SetFont("times-new-roman", "B", 30)
		if err != nil {
			return err
		}
		newline = true
		break

	case "h3":
		err := pdf.SetFont("times-new-roman", "B", 24)
		if err != nil {
			return err
		}
		newline = true
		break

	case "h4":
		err := pdf.SetFont("times-new-roman", "B", 20)
		if err != nil {
			return err
		}
		newline = true
		break

	case "h5":
		err := pdf.SetFont("times-new-roman", "B", 18)
		if err != nil {
			return err
		}
		newline = true
		break

	case "h6":
		err := pdf.SetFont("times-new-roman", "B", 16)
		if err != nil {
			return err
		}
		newline = true
		break
	}

	if newline {
		*x = padding.Left
		*line += float64(fontSize + spacing)
		if *line > paper.H-padding.Bottom {
			pdf.AddPage()
			*line = padding.Top
		}
	}

	if text != "" {
		widthRect := (paper.W - padding.Right)
		split := 0
		err := wraptext(func(textwrap string) error {
			if split > 0 {
				*x = padding.Left
				*line += float64(fontSize + spacing)
				if *line > paper.H-padding.Bottom {
					pdf.AddPage()
					*line = padding.Top
				}
			}
			pdf.SetXY(*x, *line)
			err := pdf.Cell(nil, textwrap)
			if err != nil {
				return err
			}

			xwidth, _ := pdf.MeasureTextWidth(textwrap)
			*x += xwidth
			if paper.W < *x {
				*x = padding.Left
			}

			split++
			return nil
		}, text, pdf, int(widthRect), x)
		if err != nil {
			return err
		}

		err = pdf.SetFont("times-new-roman", "", fontSize)
		if err != nil {
			return err
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		err := replaceHTML(child, pdf, line, x, fontSize, spacing, paper, padding, cache)
		if err != nil {
			return err
		}
	}

	return nil
}

// type taghtml struct {
// 	tag  []string
// 	text string
// }

// func extractP(n *html.Node, pdf *gopdf.GoPdf, tags *[]string) []taghtml {
// 	taghtmls := []taghtml{}
// 	if n.Type == html.ElementNode {
// 		tags = append(tags, n.Data)
// 		tagcaches = append(tagcaches, n.Data)
// 	}

// 	if n.Type == html.TextNode {
// 		if len(tagcaches) > 0 {
// 			tags = append(tags, taghtml{
// 				tag:  tag,
// 				text: n.Data,
// 			})
// 		}
// 	}

// 	for child := n.FirstChild; child != nil; child = child.NextSibling {
// 		childtags := extractP(child, pdf)
// 		for _, childtag := range childtags {
// 			tags = append(tags, childtag)
// 		}
// 	}

// 	return tags
// }

// func (o *SalesQuotationEngine) Test(ctx *kaos.Context, payload *string) (interface{}, error) {

// 	fmt.Println(*payload)
// 	// reply := new(string)
// 	// err = event.Publish("test", "oke", reply, nil)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return *payload, nil
// }

// func (o *SalesQuotationEngine) Tests(ctx *kaos.Context, payload *any) (interface{}, error) {
// 	event, err := ctx.DefaultEvent()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if event == nil {
// 		return nil, errors.New("Event not null")
// 	}

// 	reply := new(string)
// 	err = event.Publish("/v1/sdp/salesquotation/test", "oke", reply, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return reply, nil
// }
