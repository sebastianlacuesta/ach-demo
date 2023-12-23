# ach-demo

This demo calls 3 functions:
- `SendTransactions()`: Creates a sample ACH file with transactions to send
- `ChargeBackTransactions()`: Creates a sample ACH file with a chargeback
- `ReadACH()`: Reads an ACH file

To read this code start with `cmd/main.go` and then follow the functions at `pkg/transactions.go` and `pkg/read-ach.go`.

## How to build

```shell
make build
```

or

```shell
go build cmd/main.go
```

## How to run

In the base of the repository path:

```shell
bin/ach-demo
```

## Notes
- Example sends a mixed credits and debits entries.
- Only one batch is sent, can be organized into multiples for different companies or dates
- In this examples, only Prearranged Payment and Deposit transaction types are sent

## External library
For the manipulation of ACH files, https://github.com/moov-io/ach was used. It is Apache-2.0 licensed and doesn't call any third party service.
