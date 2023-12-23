package transactions

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/moov-io/ach"
)

var now = time.Now

type ACHData struct {
	TransactionsDate       string
	ReferenceCode          string
	Destination            string
	DestinationName        string
	Origin                 string
	OriginName             string
	StandardEntryClassCode string
	BatchNumber            int
}

type Originator struct {
	CompanyName        string
	CompanyDescription string
	Identification     string
	BatchNumber        int
}

type Transaction interface {
	BuildACHEntry() *ach.EntryDetail
}

type BaseTransaction struct {
	DepositoryAccountNumber string
	ReceivingCompany        string
	OriginalTraceNumber     string
	Amount                  int
}

func (b BaseTransaction) BuildACHEntry() *ach.EntryDetail {
	entry := ach.NewEntryDetail()
	entry.DFIAccountNumber = b.DepositoryAccountNumber
	entry.Amount = b.Amount
	entry.SetOriginalTraceNumber(b.OriginalTraceNumber)
	entry.SetReceivingCompany(b.ReceivingCompany)
	return entry
}

type CreditTransaction struct {
	BaseTransaction
}

func (c CreditTransaction) BuildACHEntry() *ach.EntryDetail {
	entry := c.BaseTransaction.BuildACHEntry()
	entry.TransactionCode = ach.CheckingCredit
	return entry
}

type DebitTransaction struct {
	BaseTransaction
}

func (d DebitTransaction) BuildACHEntry() *ach.EntryDetail {
	entry := d.BaseTransaction.BuildACHEntry()
	entry.TransactionCode = ach.CheckingDebit
	return entry
}

type ChargebackTransaction struct {
	BaseTransaction
	ReturnCode                    string
	OriginalTrace                 string
	AddendaInformation            string
	OriginalDepositoryInstitution string
}

func (c ChargebackTransaction) BuildACHEntry() *ach.EntryDetail {
	entry := c.BaseTransaction.BuildACHEntry()
	entry.Category = ach.CategoryReturn
	entry.TransactionCode = ach.GLCredit
	entry.AddendaRecordIndicator = 1
	addenda := ach.NewAddenda99()
	addenda.ReturnCode = c.ReturnCode
	addenda.OriginalTrace = c.OriginalTrace
	addenda.AddendaInformation = c.AddendaInformation
	addenda.OriginalDFI = c.OriginalDepositoryInstitution
	entry.Addenda99 = addenda
	return entry
}

func (t BaseTransaction) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Depository AccountNumber: %s\n", t.DepositoryAccountNumber)
	fmt.Fprintf(&b, "Receiving Company: %s\n", t.ReceivingCompany)
	fmt.Fprintf(&b, "Original Trace Number: %s\n", t.OriginalTraceNumber)
	fmt.Fprintf(&b, "Amount: %d\n", t.Amount)
	return b.String()
}

func (c CreditTransaction) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Type: %s\n", "CREDIT")
	fmt.Fprintf(&b, "%s", c.BaseTransaction)
	return b.String()
}

func (c DebitTransaction) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Type: %s\n", "DEBIT")
	fmt.Fprintf(&b, "%s", c.BaseTransaction)
	return b.String()
}

func (c ChargebackTransaction) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Type: %s\n", "CHARGEBACK")
	fmt.Fprintf(&b, "%s", c.BaseTransaction)
	fmt.Fprintf(&b, "Return Code: %s\n", c.ReturnCode)
	fmt.Fprintf(&b, "Original Trace: %s\n", c.OriginalTrace)
	fmt.Fprintf(&b, "Addenda Information: %s\n", c.AddendaInformation)
	fmt.Fprintf(&b, "Original Depository Institution: %s\n", c.OriginalDepositoryInstitution)
	return b.String()
}

func BuildACH(meta ACHData, originator Originator, transactions []Transaction) (string, error) {
	ts := now()

	file := ach.NewFile()

	fh := ach.NewFileHeader()
	fh.ImmediateDestination = meta.Destination
	fh.ImmediateOrigin = meta.Origin
	fh.FileCreationDate = ts.Format("060102")
	fh.FileCreationTime = ts.Format("1504")
	fh.ImmediateDestinationName = meta.DestinationName
	fh.ImmediateOriginName = meta.OriginName
	fh.ReferenceCode = meta.ReferenceCode

	bh := ach.NewBatchHeader()
	bh.ServiceClassCode = ach.MixedDebitsAndCredits
	bh.CompanyName = originator.CompanyName
	bh.CompanyIdentification = fh.ImmediateOrigin
	bh.StandardEntryClassCode = meta.StandardEntryClassCode
	bh.CompanyEntryDescription = originator.CompanyName
	bh.CompanyDescriptiveDate = meta.TransactionsDate
	bh.EffectiveEntryDate = meta.TransactionsDate
	bh.ODFIIdentification = originator.Identification
	bh.BatchNumber = meta.BatchNumber

	batch := ach.NewBatchPPD(bh)

	for i, t := range transactions {
		entry := t.BuildACHEntry()
		entry.SetRDFI(meta.Destination)
		entry.SetTraceNumber(bh.ODFIIdentification, i)
		batch.AddEntry(entry)
	}

	if err := batch.Create(); err != nil {
		return "", fmt.Errorf("unexpected error building batch: %w", err)
	}

	file.SetHeader(fh)
	file.AddBatch(batch)

	if err := file.Create(); err != nil {
		return "", fmt.Errorf("unexpected error building file: %w", err)
	}

	var buf bytes.Buffer
	achWriter := ach.NewWriter(&buf)
	if err := achWriter.Write(file); err != nil {
		return "", fmt.Errorf("could not write ach: %w", err)
	}
	return buf.String(), nil
}

func SendTransactions() {
	meta := ACHData{
		Destination:            "123456780",
		DestinationName:        "DEST BANK",
		Origin:                 "123456789",
		OriginName:             "ORIG BANK",
		TransactionsDate:       "231229",
		ReferenceCode:          "1",
		StandardEntryClassCode: ach.PPD,
		BatchNumber:            4964830,
	}
	originator := Originator{
		CompanyName:        "COMPANYONE",
		CompanyDescription: "VNDR PAY",
		Identification:     "123456780",
	}
	transactions := []Transaction{
		CreditTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111111",
				ReceivingCompany:        "CompOne",
				OriginalTraceNumber:     "8058467",
				Amount:                  234430,
			},
		},
		DebitTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111112",
				ReceivingCompany:        "CompTwo",
				OriginalTraceNumber:     "8058468",
				Amount:                  100000,
			},
		},
		CreditTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111113",
				ReceivingCompany:        "CompThree",
				OriginalTraceNumber:     "8058469",
				Amount:                  200000,
			},
		},
		CreditTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111114",
				ReceivingCompany:        "CompFour",
				OriginalTraceNumber:     "8058470",
				Amount:                  500000,
			},
		},
		CreditTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111115",
				ReceivingCompany:        "CompFive",
				OriginalTraceNumber:     "8058471",
				Amount:                  800000,
			},
		},
		DebitTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111116",
				ReceivingCompany:        "CompSix",
				OriginalTraceNumber:     "8058472",
				Amount:                  107000,
			},
		},
		CreditTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111117",
				ReceivingCompany:        "CompSeven",
				OriginalTraceNumber:     "8058473",
				Amount:                  103400,
			},
		},
	}
	fmt.Println("Transactions:")
	for _, t := range transactions {
		fmt.Println(t)
	}
	var ach string
	var err error
	if ach, err = BuildACH(meta, originator, transactions); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(ach)
	dumpACH(ach, "transactions.ach")
}

func ChargeBackTransactions() {
	meta := ACHData{
		Destination:            "123456780",
		DestinationName:        "DEST BANK",
		Origin:                 "123456789",
		OriginName:             "ORIG BANK",
		TransactionsDate:       "231229",
		ReferenceCode:          "1",
		StandardEntryClassCode: ach.PPD,
		BatchNumber:            4964830,
	}
	originator := Originator{
		CompanyName:        "COMPANYONE",
		CompanyDescription: "VNDR PAY",
		Identification:     "123456780",
	}
	transactions := []Transaction{
		ChargebackTransaction{
			BaseTransaction: BaseTransaction{
				DepositoryAccountNumber: "1111111117",
				ReceivingCompany:        "CompEight",
				OriginalTraceNumber:     "8058474",
				Amount:                  103400,
			},
			ReturnCode:                    "R10",
			OriginalTrace:                 "1111111111",
			AddendaInformation:            "Authorization Revoked",
			OriginalDepositoryInstitution: "123456780",
		},
	}
	fmt.Println("Chargebacks:")
	for _, t := range transactions {
		fmt.Println(t)
	}
	var ach string
	var err error
	if ach, err = BuildACH(meta, originator, transactions); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(ach)
	dumpACH(ach, "chargebacks.ach")
}

func dumpACH(ach, path string) error {
	f, err := os.Create(path)

	if err != nil {
		log.Printf("Could not write file %s\n", path)
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(ach); err != nil {
		log.Printf("Error at writing file %s\n", path)
		return err
	}
	return nil
}
