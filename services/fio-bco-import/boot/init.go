// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package boot

import (
	"context"
	"os"

	"github.com/jancajthaml-openbank/fio-bco-import/actor"
	"github.com/jancajthaml-openbank/fio-bco-import/config"
	"github.com/jancajthaml-openbank/fio-bco-import/daemon"
	"github.com/jancajthaml-openbank/fio-bco-import/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// Application encapsulate initialized application
type Application struct {
	cfg         config.Configuration
	interrupt   chan os.Signal
	metrics     daemon.Metrics
	fio         daemon.FioImport
	actorSystem daemon.ActorSystem
	cancel      context.CancelFunc
}

// Initialize application
func Initialize() Application {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	utils.SetupLogger(cfg.LogLevel)

	log.Infof(">>> Setup <<<")

	metrics := daemon.NewMetrics(ctx, cfg)

	storage := localfs.NewStorage(cfg.RootStorage)
	storage.SetEncryptionKey(cfg.EncryptionKey)

	actorSystem := daemon.NewActorSystem(ctx, cfg, &metrics, &storage)
	actorSystem.Support.RegisterOnRemoteMessage(actor.ProcessRemoteMessage(&actorSystem))
	actorSystem.Support.RegisterOnLocalMessage(actor.ProcessLocalMessage(&actorSystem))

	fio := daemon.NewFioImport(ctx, cfg, &storage, actor.ProcessLocalMessage(&actorSystem))

	return Application{
		cfg:         cfg,
		interrupt:   make(chan os.Signal, 1),
		metrics:     metrics,
		actorSystem: actorSystem,
		fio:         fio,
		cancel:      cancel,
	}
}
