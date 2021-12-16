// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package manager

import (
	"encoding/json"
	"fmt"
	"github.com/onosproject/chronos-exporter/pkg/collector"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var log = logging.GetLogger("manager")

// Manager single point of entry for the topology system.
type Manager struct {
	Config collector.AetherModel
}

func NewManager(configData []byte) *Manager {
	modelConfig, err := collector.LoadModel(configData)
	if err != nil {
		log.Fatal("Error unmarshalling configuration %v", err)
	}
	log.Infof("Config model loaded, with %d sites", len(modelConfig.Sites))
	return &Manager{
		Config: *modelConfig,
	}
}

func (mgr *Manager) Run() {
	mgr.Config.Collect()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/config", mgr.handleConfig)
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatal(err)
	}
}

func (mgr *Manager) handleConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	switch req.Method {
	case "OPTIONS":
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")
		w.Header().Set("Vary", "Access-Control-Request-Headers")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.WriteHeader(http.StatusOK)
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		configJson, err := json.MarshalIndent(mgr.Config, " ", " ")
		if err != nil {
			fmt.Fprintf(w, "error marshalling model %s", err)
		}
		fmt.Fprintf(w, "%s", configJson)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}
