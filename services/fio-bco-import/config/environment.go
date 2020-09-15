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

package config

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func loadConfFromEnv() Configuration {
	logLevel := strings.ToUpper(getEnvString("FIO_BCO_LOG_LEVEL", "DEBUG"))
	encryptionKey := getEnvString("FIO_BCO_ENCRYPTION_KEY", "")
	rootStorage := getEnvString("FIO_BCO_STORAGE", "/data")
	tenant := getEnvString("FIO_BCO_TENANT", "")
	fioGateway := getEnvString("FIO_BCO_FIO_GATEWAY", "https://www.fio.cz/ib_api/rest")
	ledgerGateway := getEnvString("FIO_BCO_LEDGER_GATEWAY", "https://127.0.0.1:4401")
	vaultGateway := getEnvString("FIO_BCO_VAULT_GATEWAY", "https://127.0.0.1:4400")
	syncRate := getEnvDuration("FIO_BCO_SYNC_RATE", 22*time.Second)
	lakeHostname := getEnvString("FIO_BCO_LAKE_HOSTNAME", "")
	metricsOutput := getEnvFilename("FIO_BCO_METRICS_OUTPUT", "/tmp")
	metricsRefreshRate := getEnvDuration("FIO_BCO_METRICS_REFRESHRATE", time.Second)

	if tenant == "" || lakeHostname == "" || rootStorage == "" || encryptionKey == "" {
		log.Error().Msg("missing required parameter to run")
		panic("missing required parameter to run")
	}

	keyData, err := ioutil.ReadFile(encryptionKey)
	if err != nil {
		log.Error().Msgf("unable to load encryption key from %s", encryptionKey)
		panic(fmt.Sprintf("unable to load encryption key from %s", encryptionKey))
	}

	key, err := hex.DecodeString(string(keyData))
	if err != nil {
		log.Error().Msgf("invalid encryption key %+v at %s", err, encryptionKey)
		panic(fmt.Sprintf("invalid encryption key %+v at %s", err, encryptionKey))
	}

	return Configuration{
		Tenant:             tenant,
		RootStorage:        rootStorage + "/t_" + tenant + "/import/fio",
		EncryptionKey:      []byte(key),
		FioGateway:         fioGateway,
		SyncRate:           syncRate,
		LedgerGateway:      ledgerGateway,
		VaultGateway:       vaultGateway,
		LakeHostname:       lakeHostname,
		LogLevel:           logLevel,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
	}
}

func getEnvFilename(key string, fallback string) string {
	var value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	value = filepath.Clean(value)
	if os.MkdirAll(value, os.ModePerm) != nil {
		return fallback
	}
	return value
}

func getEnvString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Error().Msgf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Error().Msgf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}
