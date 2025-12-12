package main

import (
	"github.com/phpdave11/gofpdf"
)

func main() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	// pdf.SetFont("Arial", "B", 16)
	// pdf.Cell(40, 10, "Hello, World")
	//margin kiri 8
	//margin kanan 158
	maxWidth := 193.0

	//logo
	pdf.Rect(8, 10, 35, 18, "")

	pdf.MoveTo(43, 10)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(158, 9, "PT. BAGONG DEKAKA MAKMUR", "1", 0, "CM", false, 0, "")

	//subject
	// pdf.Rect(8, 28, 193, 5, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.MoveTo(43, 19)
	pdf.CellFormat(158, 9, "FORM PENGECEKKAN DAN PERAWATAN HARIAN ( P2H ) ELF", "1", 0, "CM", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.MoveTo(8, 28)
	pdf.CellFormat(193, 9, "PERNYATAAN DRIVER", "1", 0, "CM", false, 0, "")

	y := 37.0
	x := 8.0
	pdf.SetFont("Arial", "", 8)
	pdf.MoveTo(x, y)
	_, lineHt := pdf.GetFontSize()
	lineHt += 2
	width := 28.0
	pdf.CellFormat(width, lineHt, "NAMA", "L", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "T", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth / 2) + 4) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "T", 0, "LM", false, 0, "")

	x += width
	width = 28.0
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "HO/JOB SITE", "LT", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "", 0, "T", false, 0, "")

	x += width
	width = ((maxWidth) + 8) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "RT", 0, "LM", false, 0, "")

	y = y + lineHt
	x = 8.0
	pdf.MoveTo(x, y)
	width = 28.0
	pdf.CellFormat(width, lineHt, "HARI/TGL", "L", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth / 2) + 4) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "", 0, "LM", false, 0, "")

	x += width
	width = 28.0
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "NO UNIT", "L", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth) + 8) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "R", 0, "LM", false, 0, "")

	y = y + lineHt
	x = 8.0
	pdf.MoveTo(x, y)
	width = 28.0
	pdf.CellFormat(width, lineHt, "SHIFT", "L", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth / 2) + 4) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "", 0, "LM", false, 0, "")

	x += width
	width = 28.0
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "KM UNIT", "L", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth) + 8) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "R", 0, "LM", false, 0, "")

	y = y + lineHt
	x = 8.0
	pdf.MoveTo(x, y)
	width = 28.0
	pdf.CellFormat(width, lineHt, "LOKASI KERJA", "LB", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, ":", "B", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth / 2) + 4) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "B", 0, "LM", false, 0, "")

	x += width
	width = 28.0
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "LB", 0, "LM", false, 0, "")

	x += width
	width = 4
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "B", 0, "LM", false, 0, "")

	x += width
	width = ((maxWidth) + 8) - x
	pdf.MoveTo(x, y)
	pdf.CellFormat(width, lineHt, "", "BR", 0, "LM", false, 0, "")

	x = 8
	y += lineHt
	pdf.SetFont("Arial", "", 8)
	pdf.MoveTo(x, y)
	width = maxWidth - (maxWidth / 4)
	pdf.CellFormat(width, lineHt, "Keterangan :", "LT", 0, "LM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = (maxWidth - width)
	pdf.CellFormat(width, lineHt, "", "LTR", 0, "LM", false, 0, "")

	x = 8
	y += lineHt
	pdf.MoveTo(x, y)
	width = maxWidth - (maxWidth / 4)
	pdf.CellFormat(width, lineHt, "1. Beri tanda V pada kolom KONDISI", "L", 0, "LM", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	x += width
	pdf.MoveTo(x, y)
	width = (maxWidth - width)
	pdf.CellFormat(width, lineHt, "NOMOR FORM :", "LR", 0, "CM", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	x = 8
	y += lineHt
	pdf.MoveTo(x, y)
	width = maxWidth - (maxWidth / 4)
	pdf.CellFormat(width, lineHt, "    - Kondisi Baik / Normal", "L", 0, "LM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = (maxWidth - width)
	pdf.CellFormat(width, lineHt, "", "LR", 0, "LM", false, 0, "")

	x = 8
	y += lineHt
	pdf.MoveTo(x, y)
	width = maxWidth - (maxWidth / 4)
	pdf.CellFormat(width, lineHt, "    - Kondisi Rusak / Tidak Normal", "L", 0, "LM", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	x += width
	pdf.MoveTo(x, y)
	width = (maxWidth - width)
	pdf.CellFormat(width, lineHt, "BG/SHE/F - 028", "LR", 0, "CM", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	x = 8
	y += lineHt
	pdf.MoveTo(x, y)
	width = maxWidth - (maxWidth / 4)
	pdf.CellFormat(width, lineHt, "2. Tuliskan dikolom Catatan apabila ada informasi tambahan mengenai kerusakan yang terjadi", "LB", 0, "LM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = (maxWidth - width)
	pdf.CellFormat(width, lineHt, "", "LBR", 0, "LM", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	x = 8
	y += lineHt
	pdf.MoveTo(x, y)
	width = 10
	pdf.CellFormat(width, lineHt+17.5, "NO", "LB", 0, "CM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = 60
	pdf.CellFormat(width, lineHt+17.5, "ITEM YANG HARUS DI PERIKSA", "LB", 0, "CM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = 30
	pdf.CellFormat(width, lineHt+17.5, "KODE BAHAYA", "LB", 0, "CM", false, 0, "")

	x += width
	pdf.MoveTo(x, y)
	width = 44.7
	pdf.CellFormat(width, lineHt+2.5, "KONDISI", "1", 0, "CM", false, 0, "")

	pdf.MoveTo(x, y+7.5)
	width = 22.7
	pdf.CellFormat(width, lineHt+2.5, "BAIK/", "L", 0, "CM", false, 0, "")

	pdf.MoveTo(x, y+(7.5*2))
	width = 22.7
	pdf.CellFormat(width, lineHt+2.5, "NORMAL", "LB", 0, "CM", false, 0, "")

	pdf.MoveTo(x+width, y+7.5)
	width = 22
	pdf.CellFormat(width, lineHt+2.5, "RUSAK/", "L", 0, "CM", false, 0, "")

	pdf.MoveTo(x+width+0.7, y+(7.5*2))
	width = 22
	pdf.CellFormat(width, lineHt+2.5, "TIDAK NORMAL", "LB", 0, "CM", false, 0, "")

	x += ((width * 2) + 0.7)
	pdf.MoveTo(x, y)
	width = 48.3
	pdf.CellFormat(width, lineHt+17.5, "CATATAN", "LBR", 0, "CM", false, 0, "")

	_ = maxWidth
	// //form kiri
	// fixHeight := 37 + lineHt
	// pdf.MoveTo(8, fixHeight)
	// pdf.CellFormat(width, lineHt, "HARI/TGL", "1", 0, "LM", false, 0, "")

	// fixHeight += lineHt
	// pdf.MoveTo(8, fixHeight)
	// pdf.CellFormat(width, lineHt, "SHIFT", "1", 0, "LM", false, 0, "")

	// fixHeight += lineHt
	// pdf.MoveTo(8, fixHeight)
	// pdf.CellFormat(width, lineHt, "LOKASI KERJA", "1", 0, "LM", false, 0, "")

	// fixHeight = 37 + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// fixHeight = 37 + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// //form kanan
	// xCursor = xCursor + width
	// fixHeight = 37
	// width = 28.0
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "HO/JOB SITE", "1", 0, "LM", false, 0, "")

	// fixHeight += lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "NO UNIT", "1", 0, "LM", false, 0, "")

	// fixHeight += lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "KM UNIT", "1", 0, "LM", false, 0, "")

	// xCursor = xCursor + width
	// width = 4
	// pdf.MoveTo(xCursor, 37)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// fixHeight = 37 + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, ":", "1", 0, "LM", false, 0, "")

	// xCursor = xCursor + width
	// width = (193 - xCursor) + 8
	// pdf.MoveTo(xCursor, 37)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// fixHeight = 37 + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// fixHeight = fixHeight + lineHt
	// pdf.MoveTo(xCursor, fixHeight)
	// pdf.CellFormat(width, lineHt, "", "1", 0, "LM", false, 0, "")

	// //Keterangan
	// fixHeight = fixHeight + (lineHt * 2)
	// pdf.MoveTo(8, fixHeight)
	// pdf.CellFormat(maxWidth, lineHt, "", "1", 0, "LM", false, 0, "")
	// //keterangan
	// pdf.Rect(8, 51, 134, 22, "")
	// //nomor form
	// pdf.Rect(142, 51, 59, 22, "")

	// //table header
	// //keterangan
	// pdf.Rect(8, 73, 134, 24, "")
	// //No
	// pdf.Rect(8, 73, 10, 24, "")
	// //Item yang harus di periksa
	// pdf.Rect(18, 73, 70, 24, "")
	// //Kode bahaya
	// pdf.Rect(88, 73, 17, 24, "")
	// //Kondisi
	// pdf.Rect(105, 73, 37, 12, "")
	// //Status baik
	// pdf.Rect(105, 85, 18, 12, "")
	// //Status rusak

	// //nomor form
	// pdf.Rect(142, 73, 59, 24, "")

	err := pdf.OutputFileAndClose("hello.pdf")

	_ = err
}
