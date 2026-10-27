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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDeriveHealth(t *testing.T) {
	now := time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC)
	timeout := time.Hour
	okCore := func() *coreInfo {
		return &coreInfo{State: 1, Services: []coreService{{Name: "a", Ok: true}}}
	}
	cases := []struct {
		name          string
		modules       *modulesInfo
		modulesFailed bool
		core          *coreInfo
		coreFailed    bool
		state         int
		reasons       []string
	}{
		{name: "healthy", modules: &modulesInfo{}, core: okCore(), state: 1, reasons: []string{}},
		{name: "old core ignored", modules: &modulesInfo{}, state: 1, reasons: []string{}},
		{name: "old core with unhealthy module", modules: &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "m"}}}, state: 2, reasons: []string{"module m: unhealthy"}},
		{
			name:    "module with restarts and error",
			modules: &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "github.com/x/y", Restarts: 4, LastError: "boom"}}},
			core:    okCore(),
			state:   2,
			reasons: []string{"module github.com/x/y: unhealthy, 4 restarts, last error: boom"},
		},
		{
			name:    "core state unhealthy with service",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 2, Services: []coreService{{Name: "b", Ok: false}, {Name: "a", Ok: true}}},
			state:   2,
			reasons: []string{"core service b not ok"},
		},
		{
			name:    "core state unhealthy without service",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 2},
			state:   2,
			reasons: []string{"core unhealthy"},
		},
		{
			name:    "failed update",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{FromVersion: "v0.31.1", ToVersion: "v0.32.0", Result: "failed", Step: "updating binaries"}},
			state:   2,
			reasons: []string{"update v0.31.1 -> v0.32.0 failed at updating binaries"},
		},
		{
			name:    "failed update without step",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{FromVersion: "a", ToVersion: "b", Result: "failed"}},
			state:   2,
			reasons: []string{"update a -> b failed"},
		},
		{
			name:    "successful update",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{Result: "success", Started: "2020-01-01T00:00:00Z"}},
			state:   1,
			reasons: []string{},
		},
		{
			name:    "running update within timeout hides stopped services",
			modules: &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "m"}}},
			core:    &coreInfo{State: 2, Services: []coreService{{Name: "a"}}, Update: &updateInfo{Result: "running", Started: now.Add(-10 * time.Minute).Format(time.RFC3339)}},
			state:   0,
			reasons: []string{},
		},
		{
			name:    "update reason before core and module reasons",
			modules: &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "a"}}},
			core:    &coreInfo{State: 2, Services: []coreService{{Name: "s"}}, Update: &updateInfo{FromVersion: "x", ToVersion: "y", Result: "failed"}},
			state:   2,
			reasons: []string{"update x -> y failed", "core service s not ok", "module a: unhealthy"},
		},
		{
			name:    "running update within timeout",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{Result: "running", Started: now.Add(-59 * time.Minute).Format(time.RFC3339)}},
			state:   0,
			reasons: []string{},
		},
		{
			name:    "running update at timeout boundary",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{Result: "running", Started: now.Add(-time.Hour).Format(time.RFC3339)}},
			state:   0,
			reasons: []string{},
		},
		{
			name:    "stuck update",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{FromVersion: "a", ToVersion: "b", Result: "running", Started: "2026-10-27T12:00:00Z"}},
			state:   2,
			reasons: []string{"update a -> b running since 2026-10-27T12:00:00Z"},
		},
		{
			name:    "running update with unparseable start",
			modules: &modulesInfo{},
			core:    &coreInfo{State: 1, Update: &updateInfo{Result: "running", Started: "yesterday"}},
			state:   0,
			reasons: []string{},
		},
		{name: "modules failed", modulesFailed: true, core: okCore(), state: 0, reasons: []string{}},
		{name: "core failed", modules: &modulesInfo{}, coreFailed: true, state: 0, reasons: []string{}},
		{name: "both failed", modulesFailed: true, coreFailed: true, state: 0, reasons: []string{}},
		{name: "core undetermined", modules: &modulesInfo{}, core: &coreInfo{State: 0}, state: 0, reasons: []string{}},
		{
			name:          "modules failed, core unhealthy",
			modulesFailed: true,
			core:          &coreInfo{State: 1, Services: []coreService{{Name: "a"}}},
			state:         2,
			reasons:       []string{"core service a not ok"},
		},
		{
			name:       "core failed, module unhealthy",
			modules:    &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "m", Restarts: 1}}},
			coreFailed: true,
			state:      2,
			reasons:    []string{"module m: unhealthy, 1 restarts"},
		},
		{
			name:    "reasons sorted",
			modules: &modulesInfo{Deployments: []deploymentInfo{{ModuleId: "z"}, {ModuleId: "a"}}},
			core:    okCore(),
			state:   2,
			reasons: []string{"module a: unhealthy", "module z: unhealthy"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state, reasons, _ := deriveHealth(c.modules, c.modulesFailed, c.core, c.coreFailed, now, timeout)
			if state != c.state {
				t.Errorf("expected state %d, got %d", c.state, state)
			}
			if !reflect.DeepEqual(reasons, c.reasons) {
				t.Errorf("expected reasons %q, got %q", c.reasons, reasons)
			}
		})
	}
}

func TestDeriveHealthReasonLimits(t *testing.T) {
	var deployments []deploymentInfo
	for i := 0; i < 30; i++ {
		deployments = append(deployments, deploymentInfo{ModuleId: fmt.Sprintf("m%02d", i), LastError: strings.Repeat("ä", 500)})
	}
	state, reasons, _ := deriveHealth(&modulesInfo{Deployments: deployments}, false, nil, false, time.Now(), time.Hour)
	if state != 2 {
		t.Fatal(state)
	}
	if len(reasons) != 20 {
		t.Fatalf("expected 20 reasons, got %d", len(reasons))
	}
	if !strings.HasPrefix(reasons[0], "module m00") || !strings.HasPrefix(reasons[19], "module m19") {
		t.Errorf("unexpected selection: %q ... %q", reasons[0][:12], reasons[19][:12])
	}
	for _, r := range reasons {
		if n := len([]rune(r)); n != 200 {
			t.Errorf("expected 200 characters, got %d", n)
		}
	}
}

func TestMarshalReportSizeCap(t *testing.T) {
	big := json.RawMessage(`"` + strings.Repeat("x", 61*1024) + `"`)
	small := json.RawMessage(`{"state":1}`)
	t.Run("within limit", func(t *testing.T) {
		data, err := marshalReport(Report{Reasons: []string{}, Core: small, Modules: small})
		if err != nil {
			t.Fatal(err)
		}
		var r Report
		if err = json.Unmarshal(data, &r); err != nil {
			t.Fatal(err)
		}
		if r.Core == nil || r.Modules == nil || r.ModulesError != "" || r.CoreError != "" {
			t.Errorf("unexpected report %s", data)
		}
	})
	t.Run("modules dropped first", func(t *testing.T) {
		data, err := marshalReport(Report{Reasons: []string{"r"}, Core: small, Modules: big, State: 2})
		if err != nil {
			t.Fatal(err)
		}
		var r Report
		if err = json.Unmarshal(data, &r); err != nil {
			t.Fatal(err)
		}
		if r.Modules != nil && string(r.Modules) != "null" {
			t.Errorf("modules not dropped: %s", r.Modules)
		}
		if r.ModulesError != omittedTooLarge {
			t.Errorf("unexpected modules_error %q", r.ModulesError)
		}
		if string(r.Core) != string(small) || r.CoreError != "" {
			t.Errorf("core must stay: %s %q", r.Core, r.CoreError)
		}
		if len(r.Reasons) != 1 {
			t.Errorf("reasons must stay: %q", r.Reasons)
		}
	})
	t.Run("core dropped too", func(t *testing.T) {
		data, err := marshalReport(Report{Reasons: []string{}, Core: big, Modules: big})
		if err != nil {
			t.Fatal(err)
		}
		if len(data) > maxReportSize {
			t.Errorf("report still too large: %d", len(data))
		}
		var r Report
		if err = json.Unmarshal(data, &r); err != nil {
			t.Fatal(err)
		}
		if r.ModulesError != omittedTooLarge || r.CoreError != omittedTooLarge {
			t.Errorf("unexpected errors %q %q", r.ModulesError, r.CoreError)
		}
	})
	t.Run("only core too large", func(t *testing.T) {
		data, err := marshalReport(Report{Reasons: []string{}, Core: big, Modules: small})
		if err != nil {
			t.Fatal(err)
		}
		var r Report
		if err = json.Unmarshal(data, &r); err != nil {
			t.Fatal(err)
		}
		if r.ModulesError != omittedTooLarge || r.CoreError != omittedTooLarge {
			t.Errorf("unexpected errors %q %q", r.ModulesError, r.CoreError)
		}
	})
}

func TestDeriveHealthUpdateReasonSurvivesCap(t *testing.T) {
	var deployments []deploymentInfo
	for i := 0; i < 30; i++ {
		deployments = append(deployments, deploymentInfo{ModuleId: fmt.Sprintf("m%02d", i)})
	}
	core := &coreInfo{State: 1, Update: &updateInfo{FromVersion: "a", ToVersion: "b", Result: "failed", Step: "starting components"}}
	_, reasons, _ := deriveHealth(&modulesInfo{Deployments: deployments}, false, core, false, time.Now(), time.Hour)
	if len(reasons) != 20 || reasons[0] != "update a -> b failed at starting components" {
		t.Errorf("expected the update reason first within the cap, got %d reasons starting with %q", len(reasons), reasons[0])
	}
}

func TestDeriveHealthKeysIgnoreVolatileDetails(t *testing.T) {
	_, reasons1, keys1 := deriveHealth(&modulesInfo{Deployments: []deploymentInfo{{ModuleId: "m", Restarts: 4, LastError: "x"}}}, false, nil, false, time.Now(), time.Hour)
	_, reasons2, keys2 := deriveHealth(&modulesInfo{Deployments: []deploymentInfo{{ModuleId: "m", Restarts: 9, LastError: "y"}}}, false, nil, false, time.Now(), time.Hour)
	if reflect.DeepEqual(reasons1, reasons2) {
		t.Fatalf("reasons must carry the details: %q", reasons1)
	}
	if !reflect.DeepEqual(keys1, keys2) || !reflect.DeepEqual(keys1, []string{"module m: unhealthy"}) {
		t.Errorf("keys must not depend on restarts or last error: %q vs %q", keys1, keys2)
	}
}
