/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package health_hdl

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	stateUndetermined = 0
	stateHealthy      = 1
	stateUnhealthy    = 2

	reportVersion = 1

	maxReasons      = 20
	maxReasonLen    = 200
	maxReportSize   = 60 * 1024
	omittedTooLarge = "omitted: report too large"

	updateFailed  = "failed"
	updateRunning = "running"
)

// modulesInfo is the subset of the module-manager deployments health response needed for the state.
type modulesInfo struct {
	Deployments []deploymentInfo `json:"deployments"`
}

type deploymentInfo struct {
	ModuleId  string `json:"module_id"`
	Restarts  int    `json:"restarts"`
	LastError string `json:"last_error"`
}

// coreInfo is the subset of the core-manager health response needed for the state.
type coreInfo struct {
	State    int           `json:"state"`
	Services []coreService `json:"services"`
	Update   *updateInfo   `json:"update"`
}

type coreService struct {
	Name string `json:"name"`
	Ok   bool   `json:"ok"`
}

type updateInfo struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	Started     string `json:"started"`
	Result      string `json:"result"`
	Step        string `json:"step"`
}

// Report is the payload published on the health topic.
type Report struct {
	Version      int             `json:"version"`
	Time         string          `json:"time"`
	State        int             `json:"state"`
	Reasons      []string        `json:"reasons"`
	Core         json.RawMessage `json:"core"`
	CoreError    string          `json:"core_error"`
	Modules      json.RawMessage `json:"modules"`
	ModulesError string          `json:"modules_error"`
}

// Reason priorities: the cap keeps update and core reasons, which name the cause, before per-module reasons.
const (
	reasonPrioUpdate = iota
	reasonPrioCore
	reasonPrioModule
)

// reason separates the text from its key; the key leaves out volatile details such as restart counts,
// so a crash-looping module does not count as a change on every check.
type reason struct {
	prio int
	key  string
	text string
}

// deriveHealth returns the state, the reasons for an unhealthy state and their keys for change detection.
// A failed source only yields state 0 if the other source shows no problem. A nil core with coreFailed
// false means the core-manager has no health route (old installer) and is ignored. An update that is
// still running within the timeout yields state 0, since stopped services are expected then.
func deriveHealth(modules *modulesInfo, modulesFailed bool, core *coreInfo, coreFailed bool, now time.Time, runningTimeout time.Duration) (int, []string, []string) {
	if core != nil && core.Update != nil && core.Update.Result == updateRunning && !updateTimedOut(core.Update, now, runningTimeout) {
		return stateUndetermined, []string{}, []string{}
	}
	var reasons []reason
	unhealthy := false
	undetermined := modulesFailed || coreFailed
	if modules != nil {
		for _, d := range modules.Deployments {
			unhealthy = true
			key := fmt.Sprintf("module %s: unhealthy", d.ModuleId)
			text := key
			if d.Restarts > 0 {
				text += fmt.Sprintf(", %d restarts", d.Restarts)
			}
			if d.LastError != "" {
				text += ", last error: " + d.LastError
			}
			reasons = append(reasons, reason{prio: reasonPrioModule, key: key, text: text})
		}
	}
	if core != nil {
		switch core.State {
		case stateUnhealthy:
			unhealthy = true
			if !slices.ContainsFunc(core.Services, func(s coreService) bool { return !s.Ok }) {
				reasons = append(reasons, newReason(reasonPrioCore, "core unhealthy"))
			}
		case stateUndetermined:
			undetermined = true
		}
		for _, s := range core.Services {
			if !s.Ok {
				unhealthy = true
				reasons = append(reasons, newReason(reasonPrioCore, fmt.Sprintf("core service %s not ok", s.Name)))
			}
		}
		if u := core.Update; u != nil {
			switch u.Result {
			case updateFailed:
				unhealthy = true
				text := fmt.Sprintf("update %s -> %s failed", u.FromVersion, u.ToVersion)
				if u.Step != "" {
					text += " at " + u.Step
				}
				reasons = append(reasons, newReason(reasonPrioUpdate, text))
			case updateRunning:
				unhealthy = true
				reasons = append(reasons, newReason(reasonPrioUpdate, fmt.Sprintf("update %s -> %s running since %s", u.FromVersion, u.ToVersion, u.Started)))
			}
		}
	}
	switch {
	case unhealthy:
		texts, keys := normalizeReasons(reasons)
		return stateUnhealthy, texts, keys
	case undetermined:
		return stateUndetermined, []string{}, []string{}
	default:
		return stateHealthy, []string{}, []string{}
	}
}

// updateTimedOut reports whether a running update started longer ago than the timeout; an unparseable start never times out.
func updateTimedOut(u *updateInfo, now time.Time, timeout time.Duration) bool {
	started, err := time.Parse(time.RFC3339, u.Started)
	return err == nil && now.Sub(started) > timeout
}

func newReason(prio int, text string) reason {
	return reason{prio: prio, key: text, text: text}
}

func normalizeReasons(reasons []reason) ([]string, []string) {
	slices.SortFunc(reasons, func(a, b reason) int {
		if a.prio != b.prio {
			return a.prio - b.prio
		}
		return strings.Compare(a.key, b.key)
	})
	if len(reasons) > maxReasons {
		reasons = reasons[:maxReasons]
	}
	texts := make([]string, len(reasons))
	keys := make([]string, len(reasons))
	for i, r := range reasons {
		texts[i] = truncate(r.text, maxReasonLen)
		keys[i] = truncate(r.key, maxReasonLen)
	}
	return texts, keys
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// marshalReport drops the raw module and then the raw core body if the report exceeds maxReportSize.
func marshalReport(r Report) ([]byte, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if len(data) > maxReportSize && r.Modules != nil {
		r.Modules = nil
		r.ModulesError = omittedTooLarge
		if data, err = json.Marshal(r); err != nil {
			return nil, err
		}
	}
	if len(data) > maxReportSize && r.Core != nil {
		r.Core = nil
		r.CoreError = omittedTooLarge
		if data, err = json.Marshal(r); err != nil {
			return nil, err
		}
	}
	return data, nil
}
