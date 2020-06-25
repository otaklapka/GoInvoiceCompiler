package main

import (
	"fmt"
	"github.com/signintech/gopdf"
	qrcode "github.com/skip2/go-qrcode"
)

const rightPageSide = 297.64
const marginHr = 20
const marginVr = 40
const defaultFontSize = 12
const headingFontSize = 16
const a4Width = 595.28
const a4Height = 841.89
const gray = .4
const black = 0

type Invoice struct {
	pdf    *gopdf.GoPdf
	config *Config
}

/*
	A4 = 210mm X 297mm
	margin top 20mm
	margin left, right 10mm
*/
func (config *Config) NewInvoice() (error, *Invoice) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4}) //595.28, 841.89 = A4
	pdf.AddPage()
	err := pdf.AddTTFFont("Roboto", "./assets/Roboto-Regular.ttf")
	err = pdf.AddTTFFont("RobotoBold", "./assets/Roboto-Bold.ttf")
	if err != nil {
		return err, nil
	}

	return nil, &Invoice{
		pdf:    pdf,
		config: config,
	}
}

func (invoice *Invoice) compilePdf() error {
	invoiceSerial := invoice.config.Serial()

	// header
	invoice.pdf.SetFont("RobotoBold", "", headingFontSize)
	invoice.pdf.SetX(rightPageSide + marginHr/2)
	invoice.pdf.SetY(marginVr)
	invoice.pdf.Cell(nil, fmt.Sprintf("Faktura %s", invoiceSerial))
	invoice.pdf.Br(headingFontSize * 2)

	// customer
	invoice.printEntity("ODBĚRATEL", invoice.config.Customer, rightPageSide+marginHr/2)

	invoice.pdf.SetY(marginVr)
	invoice.pdf.Br(headingFontSize * 2)

	// acc entity
	invoice.printEntity("DODAVATEL", invoice.config.AccountingEntity, marginHr)

	invoice.pdf.SetY(marginVr + headingFontSize*2 + defaultFontSize*10)
	invoice.pdf.SetX(marginHr)

	// invoice info
	invoice.printSideBySideText(marginHr, (a4Width-marginHr*3)/2, defaultFontSize,
		"Datum vystavení:", invoice.config.IssuedDate)

	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetX(marginHr)

	invoice.printSideBySideText(marginHr, (a4Width-marginHr*3)/2, defaultFontSize,
		"Datum splatnosti:", invoice.config.DueDate)

	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetX(marginHr)

	invoice.printSideBySideText(marginHr, (a4Width-marginHr*3)/2, defaultFontSize,
		"Bankovní účet:", invoice.config.BankAccount)

	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetX(marginHr)

	invoice.printSideBySideText(marginHr, (a4Width-marginHr*3)/2, defaultFontSize,
		"Variabilní symbol:", invoice.config.GetVariableSymbol())

	invoice.pdf.Br(defaultFontSize * 2)

	// invoice items
	invoice.pdf.SetGrayFill(gray)
	invoice.printItemRow("POPIS", "MNOŽSTVÍ", "CENA/MJ", "CELKEM")
	invoice.pdf.SetGrayFill(black)
	for _, item := range invoice.config.Items {
		invoice.printItemRow(
			item.Description,
			fmt.Sprintf("%.2f", item.Quantity),
			fmt.Sprintf("%.2f Kč", item.UnitPrice),
			fmt.Sprintf("%.2f Kč", item.Total()))
	}

	invoice.pdf.Br(headingFontSize * 2)
	invoice.pdf.SetX(rightPageSide + marginHr/2)

	// total
	invoice.pdf.SetLineWidth(2)
	invoice.pdf.SetLineType("solid")
	invoice.pdf.Line((a4Width/2)+marginHr, invoice.pdf.GetY(), a4Width-marginHr, invoice.pdf.GetY())
	invoice.pdf.Br(defaultFontSize)

	invoice.pdf.SetFont("RobotoBold", "", headingFontSize)
	invoice.pdf.SetX(rightPageSide + marginHr/2)
	invoice.pdf.SetGrayFill(black)
	invoice.pdf.CellWithOption(&gopdf.Rect{
		W: (a4Width - marginHr*3) / 2,
		H: headingFontSize,
	}, fmt.Sprintf("%.2f Kč", invoice.config.Total()), gopdf.CellOption{Align: gopdf.Right})

	invoice.pdf.Br(headingFontSize * 2)
	invoice.pdf.SetX(marginHr)

	// QRCode
	qrErr, qrString := invoice.config.QRString()
	if qrErr == nil {
		var png []byte
		png, err := qrcode.Encode(qrString, qrcode.Medium, 256)
		imgHolder, err := gopdf.ImageHolderByBytes(png)
		if err != nil {
			return err
		}
		invoice.pdf.ImageByHolder(imgHolder, 0, invoice.pdf.GetY(), nil)

		invoice.pdf.SetFont("Roboto", "", defaultFontSize)
		invoice.pdf.SetX(marginHr)
		invoice.pdf.SetY(invoice.pdf.GetY())

		invoice.pdf.SetGrayFill(gray)
		invoice.pdf.Cell(nil, "QR Platba")
	}

	// print
	return invoice.pdf.WritePdf(fmt.Sprintf("faktura_%s.pdf", invoiceSerial))
}

func (invoice *Invoice) printItemRow(description, quantity, unitPrice, total string) {
	invoice.pdf.SetX(marginHr)

	invoice.pdf.Cell(&gopdf.Rect{
		W: (a4Width - 2*marginHr) / 2,
		H: defaultFontSize,
	}, description)

	invoice.pdf.CellWithOption(&gopdf.Rect{
		W: (a4Width - 2*marginHr) / 10,
		H: defaultFontSize,
	}, quantity, gopdf.CellOption{Align: gopdf.Right})

	invoice.pdf.CellWithOption(&gopdf.Rect{
		W: ((a4Width - 2*marginHr) / 10) * 2,
		H: defaultFontSize,
	}, unitPrice, gopdf.CellOption{Align: gopdf.Right})

	invoice.pdf.CellWithOption(&gopdf.Rect{
		W: ((a4Width - 2*marginHr) / 10) * 2,
		H: defaultFontSize,
	}, total, gopdf.CellOption{Align: gopdf.Right})

	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetLineWidth(gray)
	invoice.pdf.SetLineType("solid")
	invoice.pdf.Line(marginHr, invoice.pdf.GetY(), a4Width-marginHr, invoice.pdf.GetY())
	invoice.pdf.Br(4)
}

func (invoice *Invoice) printSideBySideText(x0, width, height float64, leftText, rightText string) {
	invoice.pdf.SetGrayFill(gray)
	invoice.pdf.Cell(nil, leftText)
	invoice.pdf.SetGrayFill(black)
	invoice.pdf.SetX(x0)
	invoice.pdf.CellWithOption(&gopdf.Rect{
		W: width,
		H: height,
	}, rightText, gopdf.CellOption{Align: gopdf.Right})
}

func (invoice *Invoice) printEntity(title string, entity Entity, x0 float64) {
	invoice.pdf.SetGrayFill(gray)
	invoice.pdf.SetFont("Roboto", "", defaultFontSize)
	invoice.pdf.SetX(x0)
	invoice.pdf.Cell(nil, title)
	invoice.pdf.Br(defaultFontSize * 2)

	invoice.pdf.SetFont("RobotoBold", "", defaultFontSize)
	invoice.pdf.SetGrayFill(black)
	invoice.pdf.SetX(x0)
	invoice.pdf.Cell(nil, entity.Name)
	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetX(x0)
	invoice.pdf.SetFont("Roboto", "", defaultFontSize)
	invoice.pdf.SetGrayFill(gray)
	invoice.pdf.Cell(nil, entity.Address)
	invoice.pdf.Br(defaultFontSize)
	invoice.pdf.SetX(x0)
	invoice.pdf.Cell(nil, entity.Zip+" ")
	invoice.pdf.Cell(nil, entity.City)
	invoice.pdf.Br(defaultFontSize * 2)
	invoice.pdf.SetX(x0)

	if entity.Id != "" {
		invoice.printSideBySideText(x0, (a4Width-marginHr*3)/2, defaultFontSize,
			"IČ:", entity.Id)

		invoice.pdf.Br(defaultFontSize)
		invoice.pdf.SetX(x0)
	}
	if entity.VatId != "" {
		invoice.printSideBySideText(x0, (a4Width-marginHr*3)/2, defaultFontSize,
			"DIČ:", entity.VatId)

		invoice.pdf.Br(defaultFontSize)
		invoice.pdf.SetX(x0)
	}

	invoice.pdf.SetGrayFill(black)
	if entity.IsVatPayer {
		invoice.pdf.Cell(nil, "Je plátcem DPH")
	} else {
		invoice.pdf.Cell(nil, "Není plátcem DPH")
	}
}
