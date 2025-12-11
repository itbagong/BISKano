package bagonglogic

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ResponseHttp struct {
	Header http.Header
	Body   interface{}
}

func NewResponseHttp() *ResponseHttp {
	kr := new(ResponseHttp)
	kr.Header = make(http.Header)

	return kr
}

func GetURLQueryParams(ctx *kaos.Context) map[string]string {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return map[string]string{}
	}

	res := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			res[key] = values[0]
		}
	}

	return res
}

func GetFilterDimension(dimIface []interface{}) []*dbflex.Filter {
	and := []*dbflex.Filter{}

	dim := tenantcoremodel.Dimension{}
	if err := serde.Serde(dimIface, &dim); err != nil {
		return and
	}

	for _, item := range dim {
		if item.Value != "" {
			and = append(and, dbflex.ElemMatch("Dimension", dbflex.Eq("Key", item.Key), dbflex.Eq("Value", item.Value)))
		}
	}

	return and
}

var mimeByExt = []map[string]string{
	{"ext": ".aac", "mimeType": "audio/aac"},
	{"ext": ".abw", "mimeType": "application/x-abiword"},
	{"ext": ".arc", "mimeType": "application/x-freearc"},
	{"ext": ".avi", "mimeType": "video/x-msvideo"},
	{"ext": ".azw", "mimeType": "application/vnd.amazon.ebook"},
	{"ext": ".bin", "mimeType": "application/octet-stream"},
	{"ext": ".bmp", "mimeType": "image/bmp"},
	{"ext": ".bz", "mimeType": "application/x-bzip"},
	{"ext": ".bz2", "mimeType": "application/x-bzip2"},
	{"ext": ".csh", "mimeType": "application/x-csh"},
	{"ext": ".css", "mimeType": "text/css"},
	{"ext": ".csv", "mimeType": "text/csv"},
	{"ext": ".doc", "mimeType": "application/msword"},
	{"ext": ".docx", "mimeType": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	{"ext": ".eot", "mimeType": "application/vnd.ms-fontobject"},
	{"ext": ".epub", "mimeType": "application/epub+zip"},
	{"ext": ".gz", "mimeType": "application/gzip"},
	{"ext": ".gif", "mimeType": "image/gif"},
	{"ext": ".htm", "mimeType": "text/html"},
	{"ext": ".html", "mimeType": "text/html"},
	{"ext": ".ico", "mimeType": "image/vnd.microsoft.icon"},
	{"ext": ".ics", "mimeType": "text/calendar"},
	{"ext": ".jar", "mimeType": "application/java-archive"},
	{"ext": ".jpeg", "mimeType": "image/jpeg"},
	{"ext": ".js", "mimeType": "text/javascript"},
	{"ext": ".json", "mimeType": "application/json"},
	{"ext": ".jsonld", "mimeType": "application/ld+json"},
	{"ext": ".mid", "mimeType": "audio/midi"},
	{"ext": ".mjs", "mimeType": "text/javascript"},
	{"ext": ".mp3", "mimeType": "audio/mpeg"},
	{"ext": ".mpeg", "mimeType": "video/mpeg"},
	{"ext": ".mpkg", "mimeType": "application/vnd.apple.installer+xml"},
	{"ext": ".odp", "mimeType": "application/vnd.oasis.opendocument.presentation"},
	{"ext": ".ods", "mimeType": "application/vnd.oasis.opendocument.spreadsheet"},
	{"ext": ".odt", "mimeType": "application/vnd.oasis.opendocument.text"},
	{"ext": ".oga", "mimeType": "audio/ogg"},
	{"ext": ".ogv", "mimeType": "video/ogg"},
	{"ext": ".ogx", "mimeType": "application/ogg"},
	{"ext": ".opus", "mimeType": "audio/opus"},
	{"ext": ".otf", "mimeType": "font/otf"},
	{"ext": ".png", "mimeType": "image/png"},
	{"ext": ".pdf", "mimeType": "application/pdf"},
	{"ext": ".php", "mimeType": "application/php"},
	{"ext": ".ppt", "mimeType": "application/vnd.ms-powerpoint"},
	{"ext": ".pptx", "mimeType": "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	{"ext": ".rar", "mimeType": "application/x-rar-compressed"},
	{"ext": ".rtf", "mimeType": "application/rtf"},
	{"ext": ".sh", "mimeType": "application/x-sh"},
	{"ext": ".svg", "mimeType": "image/svg+xml"},
	{"ext": ".swf", "mimeType": "application/x-shockwave-flash"},
	{"ext": ".tar", "mimeType": "application/x-tar"},
	{"ext": ".tif", "mimeType": "image/tiff"},
	{"ext": ".tiff", "mimeType": "image/tiff"},
	{"ext": ".ts", "mimeType": "video/mp2t"},
	{"ext": ".ttf", "mimeType": "font/ttf"},
	{"ext": ".txt", "mimeType": "text/plain"},
	{"ext": ".vsd", "mimeType": "application/vnd.visio"},
	{"ext": ".wav", "mimeType": "audio/wav"},
	{"ext": ".weba", "mimeType": "audio/webm"},
	{"ext": ".webm", "mimeType": "video/webm"},
	{"ext": ".webp", "mimeType": "image/webp"},
	{"ext": ".woff", "mimeType": "font/woff"},
	{"ext": ".woff2", "mimeType": "font/woff2"},
	{"ext": ".xhtml", "mimeType": "application/xhtml+xml"},
	{"ext": ".xls", "mimeType": "application/vnd.ms-excel"},
	{"ext": ".xlsx", "mimeType": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	{"ext": ".xml", "mimeType": "application/xml "},
	{"ext": ".xul", "mimeType": "application/vnd.mozilla.xul+xml"},
	{"ext": ".zip", "mimeType": "application/zip"},
	{"ext": ".3gp", "mimeType": "video/3gpp"},
	{"ext": ".3g2", "mimeType": "video/3gpp2"},
	{"ext": ".7z", "mimeType": "application/x-7z-compressed"},
}

// GetMimeByExtention definition
func GetMimeByExtention(ext string) string {
	for _, each := range mimeByExt {
		if each["ext"] == ext {
			return each["mimeType"]
		}
	}

	return ""
}

func ExcelHeaderForIndex(i int) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := ""

	if i >= len(alphabet) {
		r += string(alphabet[i/len(alphabet)-1])
	}

	return r + string(alphabet[i%len(alphabet)])
}

// CreateExcel build excel from givven data
func CreateExcelWithStyleBackgroundColor(headers, types []string, datas []codekit.M, sheetNames ...string) ([]byte, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("error no header data")
	}

	excel := excelize.NewFile()

	i := 1
	if len(types) > 0 {
		for idx, head := range types {
			cell := ExcelHeaderForIndex(idx) + "" + strconv.Itoa(i)
			excel.SetCellValue("Sheet1", cell, head)
		}
		i++
	}

	for idx, head := range headers {
		cell := ExcelHeaderForIndex(idx) + "" + strconv.Itoa(i)
		excel.SetCellValue("Sheet1", cell, head)
	}
	i++

	for _, data := range datas {
		for j, head := range headers {
			cols := ExcelHeaderForIndex(j)
			val := data.Get(head)
			valColor := data.GetString(head + " COLOR")
			valAxis := fmt.Sprintf("%v%d", cols, i)
			excel.SetCellValue("Sheet1", valAxis, val)

			if valColor != "" {
				style, err := excel.NewStyle(`{"fill":{"type":"pattern","color":["` + valColor + `"],"pattern":1}}`)
				if err != nil {
					fmt.Println(err)
				}
				excel.SetCellStyle("Sheet1", valAxis, valAxis, style)
			}
		}
		i++
	}

	if len(sheetNames) > 0 {
		excel.SetSheetName("Sheet1", sheetNames[0])
	}

	b, err := excel.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func BuildDownloadExcel(headers, types []string, datas []codekit.M, sheetNames ...string) ([]byte, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("error no header data")
	}

	excel := excelize.NewFile()

	i := 1
	if len(types) > 0 {
		for idx, head := range types {
			cell := ExcelHeaderForIndex(idx) + "" + strconv.Itoa(i)
			excel.SetCellValue("Sheet1", cell, head)
		}
		i++
	}

	for idx, head := range headers {
		cell := ExcelHeaderForIndex(idx) + "" + strconv.Itoa(i)
		excel.SetCellValue("Sheet1", cell, head)
	}
	i++

	for _, data := range datas {
		for j, head := range headers {
			cols := ExcelHeaderForIndex(j)
			val := data.Get(head)
			valAxis := fmt.Sprintf("%v%d", cols, i)
			excel.SetCellValue("Sheet1", valAxis, val)

			// if valColor != "" {
			// 	style, err := excel.NewStyle(`{"fill":{"type":"pattern","color":["` + valColor + `"],"pattern":1}}`)
			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}
			// 	excel.SetCellStyle("Sheet1", valAxis, valAxis, style)
			// }
		}
		i++
	}

	if len(sheetNames) > 0 {
		excel.SetSheetName("Sheet1", sheetNames[0])
	}

	b, err := excel.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// ExcelFromBase64 read excel from Base64
func ExcelFromBase64(input string) ([]byte, error) {
	dec, err := base64.StdEncoding.DecodeString(input[strings.IndexByte(input, ',')+1:])
	if err != nil {
		return nil, fmt.Errorf("cannot decode excel file %s", err.Error())
	}

	return dec, nil
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func MonthsBetween(start, end time.Time) int {
	// Hitung perbedaan tahun dan bulan
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())

	// Total bulan
	totalMonths := years*12 + months

	// Sesuaikan dengan hari untuk memperhitungkan bulan yang tidak penuh
	if end.Day() < start.Day() {
		totalMonths--
	}

	return totalMonths
}

// convert first character to uppercase
func title(word string) string {
	words := strings.Split(word, " ")
	for i := range words {
		words[i] = cases.Title(language.Indonesian).String(words[i])
	}

	return strings.Join(words, " ")
}
