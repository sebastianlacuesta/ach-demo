# ach-demo

TODO: Explain

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
