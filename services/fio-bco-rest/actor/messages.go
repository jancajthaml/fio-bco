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

package actor

const (
	// ReqTokens fio message request code for "Get Tokens"
	ReqTokens = "GT"
	// RespTokens fio message response code for "Get Tokens"
	RespTokens = "TG"
	// ReqCreateToken fio message request code for "New Token"
	ReqCreateToken = "NT"
	// RespCreateToken fio message response code for "New Token"
	RespCreateToken = "TN"
	// ReqDeleteToken fio message request code for "Delete Token"
	ReqDeleteToken = "DT"
	// RespDeleteToken fio message response code for "Delete Token"
	RespDeleteToken = "TD"
	// FatalError fio message response code for "Error"
	FatalError = "EE"
)

// CreateTokenMessage is message for creation of new token
func CreateTokenMessage(token Token) string {
	return ReqCreateToken + " " + token.Value
}

// DeleteTokenMessage is message for deletion of new token
func DeleteTokenMessage() string {
	return ReqDeleteToken
}
