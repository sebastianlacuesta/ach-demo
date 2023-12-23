package main

import (
	p "github.com/sebastianlacuesta/ach-demo/pkg"
)

func main() {
	p.SendTransactions()
	p.ChargeBackTransactions()
	p.ReadACH()
}
