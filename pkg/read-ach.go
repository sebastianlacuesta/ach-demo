package transactions

import (
	"fmt"
	"log"
	"os"

	"github.com/moov-io/ach"
)

func ReadACH() {
	fCredit, err := os.Open("transactions.ach")
	if err != nil {
		log.Fatalln(err)
	}
	rCredit := ach.NewReader(fCredit)
	achFileCredit, err := rCredit.Read()
	if err != nil {
		log.Fatalf("reading file: %v\n", err)
	}
	if err := achFileCredit.Validate(); err != nil {
		log.Fatalf("validating file: %v\n", err)
	}
	fmt.Printf("File Name: %s\n\n", fCredit.Name())
	fmt.Printf("Total Credit Amount: %d\n", achFileCredit.Control.TotalCreditEntryDollarAmountInFile)
	for _, b := range achFileCredit.Batches {
		fmt.Printf("SEC Code: %s\n\n", b.GetHeader().StandardEntryClassCode)
	}
}
