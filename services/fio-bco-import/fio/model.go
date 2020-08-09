// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fio

import (
	"math"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/fio-bco-import/model"
)

// FioImportEnvelope represents fio gateway import statement entity
type FioImportEnvelope struct {
	Statement accountStatement `json:"accountStatement"`
}

type accountStatement struct {
	Info            accountInfo     `json:"info"`
	TransactionList transactionList `json:"transactionList"`
}

type accountInfo struct {
	AccountID      string  `json:"accountId"`
	BankID         string  `json:"bankId"`
	Currency       string  `json:"currency"`
	IBAN           string  `json:"iban"`
	BIC            string  `json:"bic"`
	OpeningBalance float64 `json:"openingBalance"`
	ClosingBalance float64 `json:"closingBalance"`
	IDFrom         int     `json:"idFrom"`
	IDTo           int     `json:"idTo"`
	IDLastDownload int     `json:"idLastDownload"`
}

type transactionList struct {
	Transactions []fioTransaction `json:"transaction"`
}

type fioTransaction struct {
	Column0  *stringNode `json:"column0"`
	Column1  *floatNode  `json:"column1"`
	Column2  *stringNode `json:"column2"`
	Column3  *stringNode `json:"column3"`
	Column4  *stringNode `json:"column4"`
	Column5  *stringNode `json:"column5"`
	Column6  *stringNode `json:"column6"`
	Column7  *stringNode `json:"column7"`
	Column8  *stringNode `json:"column8"`
	Column9  *stringNode `json:"column9"`
	Column10 *stringNode `json:"column10"`
	Column12 *stringNode `json:"column12"`
	Column14 *stringNode `json:"column14"`
	Column16 *stringNode `json:"column16"`
	Column17 *intNode    `json:"column17"`
	Column18 *stringNode `json:"column18"`
	Column22 *intNode    `json:"column22"`
	Column25 *stringNode `json:"column25"`
	Column26 *stringNode `json:"column26"`
}

type stringNode struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type dateNode struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type intNode struct {
	Value int64  `json:"value"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

type floatNode struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
	ID    int     `json:"id"`
}

// GetTransactions return generator of fio transactions over given envelope
func (envelope *FioImportEnvelope) GetTransactions(tenant string) <-chan model.Transaction {
	chnl := make(chan model.Transaction)
	if envelope == nil {
		close(chnl)
		return chnl
	}

	var previousIdTransaction = ""
	var buffer = make([]model.Transfer, 0)

	go func() {

		now := time.Now()

		var credit string
		var debit string
		var valueDate time.Time

		for _, transfer := range envelope.Statement.TransactionList.Transactions {
			if transfer.Column22 == nil || transfer.Column1 == nil {
				continue
			}

			if transfer.Column1.Value > 0 {
				credit = envelope.Statement.Info.IBAN
				if transfer.Column2 == nil {
					debit = envelope.Statement.Info.BIC
				} else {
					if transfer.Column3 != nil {
						debit = model.NormalizeAccountNumber(transfer.Column2.Value, transfer.Column3.Value, envelope.Statement.Info.BankID)
					} else {
						debit = model.NormalizeAccountNumber(transfer.Column2.Value, "", envelope.Statement.Info.BankID)
					}
				}
			} else {
				if transfer.Column2 == nil {
					credit = envelope.Statement.Info.BIC
				} else {
					if transfer.Column3 != nil {
						credit = model.NormalizeAccountNumber(transfer.Column2.Value, transfer.Column3.Value, envelope.Statement.Info.BankID)
					} else {
						credit = model.NormalizeAccountNumber(transfer.Column2.Value, "", envelope.Statement.Info.BankID)
					}
				}
				debit = envelope.Statement.Info.IBAN
			}

			if transfer.Column0 == nil {
				valueDate = now
			} else if date, err := time.Parse("2006-01-02-0700", transfer.Column0.Value); err == nil {
				valueDate = date.UTC()
			} else {
				valueDate = now
			}

			buffer = append(buffer, model.Transfer{
				IDTransfer: transfer.Column22.Value,
				Credit: model.AccountPair{
					Tenant: tenant,
					Name:   credit,
				},
				Debit: model.AccountPair{
					Tenant: tenant,
					Name:   debit,
				},
				ValueDate: valueDate.Format("2006-01-02T15:04:05Z0700"),
				Amount:    math.Abs(transfer.Column1.Value),
				Currency:  envelope.Statement.Info.Currency, // FIXME not true in all cases
			})

			idTransaction := envelope.Statement.Info.IBAN + strconv.FormatInt(transfer.Column17.Value, 10)

			if previousIdTransaction == "" {
				previousIdTransaction = idTransaction
			} else if previousIdTransaction != idTransaction {
				previousIdTransaction = idTransaction
				transfers := make([]model.Transfer, len(buffer))
				copy(transfers, buffer)
				buffer = make([]model.Transfer, 0)
				chnl <- model.Transaction{
					IDTransaction: idTransaction,
					Transfers:     transfers,
				}
			}

		}

		if len(buffer) > 0 {
			transfers := make([]model.Transfer, len(buffer))
			copy(transfers, buffer)
			buffer = make([]model.Transfer, 0)
			chnl <- model.Transaction{
				IDTransaction: previousIdTransaction,
				Transfers:     transfers,
			}
		}

		close(chnl)
	}()

	return chnl
}

// GetAccounts return generator of fio accounts over given envelope
func (envelope *FioImportEnvelope) GetAccounts() <-chan model.Account {
	chnl := make(chan model.Account)
	if envelope == nil {
		close(chnl)
		return chnl
	}

	var visited = make(map[string]interface{})

	go func() {

		var set = make(map[string]fioTransaction)

		for _, transfer := range envelope.Statement.TransactionList.Transactions {
			if transfer.Column2 == nil {
				// INFO fee and taxes and maybe card payments
				set[envelope.Statement.Info.BIC] = transfer
			} else {
				set[transfer.Column2.Value] = transfer
			}
		}

		var normalizedAccount string
		var accountFormat string

		for account, transfer := range set {
			if transfer.Column3 != nil {
				normalizedAccount = model.NormalizeAccountNumber(account, transfer.Column3.Value, envelope.Statement.Info.BankID)
			} else {
				normalizedAccount = model.NormalizeAccountNumber(account, "", envelope.Statement.Info.BankID)
			}

			if normalizedAccount != account {
				accountFormat = "IBAN"
			} else {
				accountFormat = "FIO_UNKNOWN"
			}

			if _, ok := visited[normalizedAccount]; !ok {
				chnl <- model.Account{
					Name:           normalizedAccount,
					Format:         accountFormat,
					Currency:       envelope.Statement.Info.Currency, // FIXME not true in all cases
					IsBalanceCheck: false,
				}
				visited[normalizedAccount] = nil
			}
		}

		if _, ok := visited[envelope.Statement.Info.IBAN]; !ok {
			chnl <- model.Account{
				Name:           envelope.Statement.Info.IBAN,
				Format:         "IBAN",
				Currency:       envelope.Statement.Info.Currency,
				IsBalanceCheck: false,
			}
			visited[envelope.Statement.Info.IBAN] = nil
		}

		close(chnl)
	}()

	return chnl
}
