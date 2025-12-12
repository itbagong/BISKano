package tenantcorelogic

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/tenantcore/tenantcoreconfig"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/ariefdarmawan/kasset"
	"github.com/sebarcode/codekit"
)

// PDF Global Constant
const (
	PDFTemplatePath = "templates/pdf"
	PDFPageSize     = wkhtmltopdf.PageSizeA4
	PDFDPI          = 600
)

type PDFEngine struct {
	fs kasset.AssetFS
}

func NewPDFEngine(fs kasset.AssetFS) *PDFEngine {
	p := new(PDFEngine)
	p.fs = fs
	return p
}

type PDFGenParam struct {
	PageSize string
	DPI      uint
}

type PDFByteResponse struct {
	PDFByte []byte
}

type PDFFromUrlRequest struct {
	URL         string
	PDFGenParam *PDFGenParam // optional
}

func (o *PDFEngine) FromUrl(ctx *kaos.Context, req *PDFFromUrlRequest) (*PDFByteResponse, error) {
	pdfg, e := o.generator(req.PDFGenParam)
	if e != nil {
		return nil, e
	}

	pdfg.AddPage(wkhtmltopdf.NewPage(req.URL))
	return o.create(pdfg)
}

type PDFFromHtmlRequest struct {
	HTMLByte    []byte // takes precedence
	HTMLPath    string
	PDFGenParam *PDFGenParam // optional
}

func (o *PDFEngine) FromHtml(ctx *kaos.Context, req *PDFFromHtmlRequest) (*PDFByteResponse, error) {
	pdfg, e := o.generator(req.PDFGenParam)
	if e != nil {
		return nil, e
	}

	if req.HTMLByte == nil {
		htmlfile, e := os.ReadFile(req.HTMLPath)
		if e != nil {
			return nil, e
		}

		req.HTMLByte = htmlfile
	}

	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(req.HTMLByte)))

	return o.create(pdfg)
}

type PDFFromTemplateRequest struct {
	TemplateName string
	Data         codekit.M
	PDFGenParam  *PDFGenParam // optional
}

func (o *PDFEngine) FromTemplate(ctx *kaos.Context, req *PDFFromTemplateRequest) (*PDFByteResponse, error) {
	fmt.Println("Tenantcore FromTemplate | begins generating PDF")
	pdfg, e := o.generator(req.PDFGenParam)
	if e != nil {
		fmt.Println("FromTemplate | o.generator error:", e)
		return nil, e
	}

	if tenantcoreconfig.Config.PdfTemplatePath == "" {
		fmt.Println("FromTemplate | config 'pdf_template_path' has not been set")
		return nil, fmt.Errorf("config 'pdf_template_path' has not been set")
	}

	if o.fs == nil || reflect.ValueOf(o.fs).IsNil() {
		fmt.Println("FromTemplate | unable to use asset fs, please check storage configuration")
		return nil, fmt.Errorf("unable to use asset fs, please check storage configuration")
	}

	fmt.Println("Tenantcore FromTemplate | o.fs.Read")
	templByte, e := o.fs.Read(fmt.Sprintf("%s/%s.html", tenantcoreconfig.Config.PdfTemplatePath, req.TemplateName))
	if e != nil {
		fmt.Println("FromTemplate | o.fs.Read error:", e)
		return nil, e
	}

	funcMap := template.FuncMap{
		"lenArr":   func(i []interface{}) int { return len(i) },
		"getWidth": func(i int) int { return (100 / i) },
		"getTextRight": func(list []interface{}, idx int) bool {
			for i, c := range list {
				if i == idx {
					if strings.Contains(c.(string), ":R") {
						return true
					}
				}
			}
			return false
		},
		"replace":  strings.ReplaceAll,
		"contains": strings.Contains,
	}

	fmt.Println("Tenantcore FromTemplate | template.New.Parse")
	tmpl, e := template.New("pdf_template").Funcs(funcMap).Parse(string(templByte))
	if e != nil {
		fmt.Println("FromTemplate | template.New:", e)
		return nil, e
	}

	fmt.Println("Tenantcore FromTemplate | tmpl.Execute")
	w := bytes.NewBufferString("")
	if e := tmpl.Execute(w, req.Data); e != nil {
		fmt.Println("FromTemplate | tmpl.Execute:", e)
		return nil, e
	}

	fmt.Println("Tenantcore FromTemplate | pdfg.AddPage")
	pdfg.AddPage(wkhtmltopdf.NewPageReader(w))

	fmt.Println("Tenantcore FromTemplate | o.create")
	return o.create(pdfg)
}

func (o *PDFEngine) generator(pdfGenParam *PDFGenParam) (*wkhtmltopdf.PDFGenerator, error) {
	if tenantcoreconfig.Config.PdfExePath == "" {
		return nil, fmt.Errorf("config 'pdf_exe_path' has not been set")
	}

	wkhtmltopdf.SetPath(tenantcoreconfig.Config.PdfExePath)

	pdfg, e := wkhtmltopdf.NewPDFGenerator()
	if e != nil {
		return nil, e
	}

	if pdfGenParam == nil {
		pdfGenParam = &PDFGenParam{
			PageSize: PDFPageSize,
			DPI:      PDFDPI,
		}
	}

	pdfg.Dpi.Set(pdfGenParam.DPI)
	pdfg.PageSize.Set(pdfGenParam.PageSize)

	return pdfg, nil
}

func (o *PDFEngine) create(pdfg *wkhtmltopdf.PDFGenerator) (*PDFByteResponse, error) {
	fmt.Println("Tenantcore create | pdfg.Create")
	if e := pdfg.Create(); e != nil {
		fmt.Println("Tenantcore create | pdfg.Create error:", e)
		return nil, e
	}

	return &PDFByteResponse{PDFByte: pdfg.Bytes()}, nil
}
