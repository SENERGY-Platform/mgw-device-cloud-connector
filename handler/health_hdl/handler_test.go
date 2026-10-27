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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
)

const (
	healthyModules = `{"deployments":[],"total_enabled_deployments":3}`
	healthyCore    = `{"state":1,"version":"v0.32.0","services":[{"name":"a","ok":true}],"update":null}`
)

type fakeEndpoint struct {
	mu     sync.Mutex
	status int
	body   string
	paths  []string
}

func (f *fakeEndpoint) set(status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.body = status, body
}

func (f *fakeEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	status, body := f.status, f.body
	f.paths = append(f.paths, r.URL.RequestURI())
	f.mu.Unlock()
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

type publishRecorder struct {
	mu     sync.Mutex
	topics []string
	data   [][]byte
	err    error
}

func (p *publishRecorder) publish(topic string, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.topics = append(p.topics, topic)
	p.data = append(p.data, data)
	return nil
}

func (p *publishRecorder) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.data)
}

func (p *publishRecorder) lastReport(t *testing.T) Report {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.data) == 0 {
		t.Fatal("nothing published")
	}
	var r Report
	if err := json.Unmarshal(p.data[len(p.data)-1], &r); err != nil {
		t.Fatal(err)
	}
	return r
}

type fixture struct {
	mm, cm *fakeEndpoint
	pub    *publishRecorder
	hdl    *Handler
	clock  time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	f := &fixture{
		mm:    &fakeEndpoint{status: 200, body: healthyModules},
		cm:    &fakeEndpoint{status: 200, body: healthyCore},
		pub:   &publishRecorder{},
		clock: time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC),
	}
	mmSrv := httptest.NewServer(f.mm)
	cmSrv := httptest.NewServer(f.cm)
	t.Cleanup(mmSrv.Close)
	t.Cleanup(cmSrv.Close)
	hdl, err := New(http.DefaultClient, mmSrv.URL+"/module-manager/", cmSrv.URL+"/core-manager", f.pub.publish, time.Minute, 5*time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	f.hdl = hdl
	f.hdl.now = func() time.Time { return f.clock }
	return f
}

func (f *fixture) tick(d time.Duration) {
	f.clock = f.clock.Add(d)
	f.hdl.check(context.Background())
}

func TestReadModules(t *testing.T) {
	big := `{"deployments":[],"pad":"` + strings.Repeat("x", maxBodySize) + `"}`
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr string
		wantRaw bool
	}{
		{name: "ok", status: 200, body: healthyModules, wantRaw: true},
		{name: "unavailable", status: 503, body: "busy", wantErr: "unexpected status 503"},
		{name: "engine error", status: 500, body: "{}", wantErr: "unexpected status 500"},
		{name: "not found", status: 404, body: "", wantErr: "unexpected status 404"},
		{name: "malformed", status: 200, body: `{"deployments":`, wantErr: "invalid response: "},
		{name: "oversized", status: 200, body: big, wantErr: "response too large"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			f.mm.set(c.status, c.body)
			raw, info, errTxt := f.hdl.readModules(context.Background())
			if !strings.HasPrefix(errTxt, c.wantErr) || (c.wantErr == "" && errTxt != "") {
				t.Errorf("expected error %q, got %q", c.wantErr, errTxt)
			}
			if c.wantRaw != (raw != nil) || c.wantRaw != (info != nil) {
				t.Errorf("unexpected raw=%v info=%v", raw != nil, info != nil)
			}
			if c.wantRaw && string(raw) != c.body {
				t.Errorf("raw body must be forwarded verbatim, got %s", raw)
			}
			f.mm.mu.Lock()
			defer f.mm.mu.Unlock()
			if len(f.mm.paths) != 1 || f.mm.paths[0] != "/module-manager/health/deployments?auxiliary_deployments=true" {
				t.Errorf("unexpected request path %q", f.mm.paths)
			}
		})
	}
}

func TestReadCore(t *testing.T) {
	big := `{"state":1,"pad":"` + strings.Repeat("x", maxBodySize) + `"}`
	cases := []struct {
		name            string
		status          int
		body            string
		wantErr         string
		wantUnavailable bool
		wantRaw         bool
	}{
		{name: "ok", status: 200, body: healthyCore, wantRaw: true},
		{name: "old installer 404", status: 404, body: "not found", wantErr: "core health not available (status 404)", wantUnavailable: true},
		{name: "old installer 403", status: 403, body: "forbidden", wantErr: "core health not available (status 403)", wantUnavailable: true},
		{name: "server error", status: 500, body: "{}", wantErr: "unexpected status 500"},
		{name: "unavailable", status: 503, body: "", wantErr: "unexpected status 503"},
		{name: "malformed", status: 200, body: `[1,`, wantErr: "invalid response: "},
		{name: "oversized", status: 200, body: big, wantErr: "response too large"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			f.cm.set(c.status, c.body)
			raw, info, errTxt, unavailable := f.hdl.readCore(context.Background())
			if !strings.HasPrefix(errTxt, c.wantErr) || (c.wantErr == "" && errTxt != "") {
				t.Errorf("expected error %q, got %q", c.wantErr, errTxt)
			}
			if unavailable != c.wantUnavailable {
				t.Errorf("expected unavailable=%v", c.wantUnavailable)
			}
			if c.wantRaw != (raw != nil) || c.wantRaw != (info != nil) {
				t.Errorf("unexpected raw=%v info=%v", raw != nil, info != nil)
			}
			f.cm.mu.Lock()
			defer f.cm.mu.Unlock()
			if len(f.cm.paths) != 1 || f.cm.paths[0] != "/core-manager/health" {
				t.Errorf("unexpected request path %q", f.cm.paths)
			}
		})
	}
}

func TestReadUnreachable(t *testing.T) {
	f := newFixture(t)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	f.hdl.mmBaseUrl = closed.URL
	f.hdl.cmBaseUrl = closed.URL
	if _, _, errTxt := f.hdl.readModules(context.Background()); errTxt == "" {
		t.Error("expected modules error")
	}
	if _, _, errTxt, unavailable := f.hdl.readCore(context.Background()); errTxt == "" || unavailable {
		t.Errorf("expected core error, got %q unavailable=%v", errTxt, unavailable)
	}
}

func TestCheckReport(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		f := newFixture(t)
		f.hdl.check(context.Background())
		if f.pub.count() != 1 || f.pub.topics[0] != "health" {
			t.Fatalf("unexpected publishes: %v", f.pub.topics)
		}
		r := f.pub.lastReport(t)
		if r.Version != 1 || r.State != 1 || r.Time != "2026-10-27T14:00:00Z" || len(r.Reasons) != 0 || r.CoreError != "" || r.ModulesError != "" {
			t.Errorf("unexpected report %+v", r)
		}
		if string(r.Core) != healthyCore || string(r.Modules) != healthyModules {
			t.Errorf("bodies not forwarded: %s %s", r.Core, r.Modules)
		}
		if !strings.Contains(string(f.pub.data[0]), `"reasons":[]`) {
			t.Errorf("reasons must be an empty array: %s", f.pub.data[0])
		}
	})
	t.Run("old core", func(t *testing.T) {
		f := newFixture(t)
		f.cm.set(404, "nope")
		f.hdl.check(context.Background())
		r := f.pub.lastReport(t)
		if r.State != 1 || r.CoreError != "core health not available (status 404)" || (r.Core != nil && string(r.Core) != "null") {
			t.Errorf("unexpected report %+v", r)
		}
		if strings.Contains(string(f.pub.data[0]), `"core":{`) {
			t.Errorf("core must be null: %s", f.pub.data[0])
		}
	})
	t.Run("modules unavailable", func(t *testing.T) {
		f := newFixture(t)
		f.mm.set(503, "")
		f.hdl.check(context.Background())
		r := f.pub.lastReport(t)
		if r.State != 0 || r.ModulesError != "unexpected status 503" || (r.Modules != nil && string(r.Modules) != "null") {
			t.Errorf("unexpected report %+v", r)
		}
	})
	t.Run("both failing", func(t *testing.T) {
		f := newFixture(t)
		f.mm.set(500, "")
		f.cm.set(500, "")
		f.hdl.check(context.Background())
		r := f.pub.lastReport(t)
		if r.State != 0 || r.ModulesError == "" || r.CoreError == "" {
			t.Errorf("unexpected report %+v", r)
		}
	})
	t.Run("unhealthy module", func(t *testing.T) {
		f := newFixture(t)
		f.mm.set(200, `{"deployments":[{"module_id":"github.com/x/y","state":2,"restarts":4}],"total_enabled_deployments":1}`)
		f.hdl.check(context.Background())
		r := f.pub.lastReport(t)
		if r.State != 2 || !reflect.DeepEqual(r.Reasons, []string{"module github.com/x/y: unhealthy, 4 restarts"}) {
			t.Errorf("unexpected report %+v", r)
		}
	})
	t.Run("size cap", func(t *testing.T) {
		f := newFixture(t)
		f.mm.set(200, `{"deployments":[{"module_id":"m"}],"pad":"`+strings.Repeat("x", 70*1024)+`"}`)
		f.hdl.check(context.Background())
		if n := len(f.pub.data[0]); n > 60*1024 {
			t.Errorf("payload too large: %d", n)
		}
		r := f.pub.lastReport(t)
		if r.ModulesError != "omitted: report too large" || r.State != 2 || len(r.Reasons) != 1 || string(r.Core) != healthyCore {
			t.Errorf("unexpected report %+v", r)
		}
	})
}

func TestPublishDecision(t *testing.T) {
	f := newFixture(t)
	unhealthy := `{"deployments":[{"module_id":"m"}]}`

	f.tick(0)
	if f.pub.count() != 1 {
		t.Fatalf("first check must publish, got %d", f.pub.count())
	}
	f.tick(time.Minute)
	f.tick(time.Minute)
	if f.pub.count() != 1 {
		t.Errorf("unchanged report within heartbeat must not publish, got %d", f.pub.count())
	}
	f.tick(3 * time.Minute) // 5m since last publish
	if f.pub.count() != 2 {
		t.Errorf("heartbeat must publish, got %d", f.pub.count())
	}
	f.tick(4*time.Minute + 59*time.Second)
	if f.pub.count() != 2 {
		t.Errorf("heartbeat not yet due, got %d", f.pub.count())
	}
	f.mm.set(200, unhealthy)
	f.tick(time.Second)
	if f.pub.count() != 3 || f.pub.lastReport(t).State != 2 {
		t.Errorf("state change must publish, got %d", f.pub.count())
	}
	f.mm.set(200, `{"deployments":[{"module_id":"m"},{"module_id":"n"}]}`)
	f.tick(time.Minute)
	if f.pub.count() != 4 || len(f.pub.lastReport(t).Reasons) != 2 {
		t.Errorf("changed reasons must publish, got %d", f.pub.count())
	}
	f.mm.set(200, `{"deployments":[{"module_id":"m","restarts":7},{"module_id":"n","last_error":"boom"}]}`)
	f.tick(time.Minute)
	if f.pub.count() != 4 {
		t.Errorf("changed restart count or last error alone must not publish, got %d", f.pub.count())
	}
}

func TestNewRejectsInvalidIntervals(t *testing.T) {
	for _, d := range [][3]time.Duration{{0, time.Minute, time.Minute}, {time.Minute, -1, time.Minute}, {time.Minute, time.Minute, 0}} {
		if _, err := New(http.DefaultClient, "http://mm", "http://cm", nil, d[0], d[1], d[2]); err == nil {
			t.Errorf("expected an error for intervals %v", d)
		}
	}
}

func TestPublishErrorRetry(t *testing.T) {
	f := newFixture(t)
	warnings := 0
	f.hdl.warnf = func(string, ...any) { warnings++ }
	f.pub.err = errors.New("not connected")
	f.tick(0)
	if f.pub.count() != 0 || f.hdl.last != nil {
		t.Fatal("failed publish must not be recorded")
	}
	if f.hdl.warnedState == nil || *f.hdl.warnedState != 1 {
		t.Error("expected warning recorded for state 1")
	}
	f.tick(time.Minute)
	if f.pub.count() != 0 || f.hdl.warnedState == nil {
		t.Error("still failing")
	}
	if warnings != 1 {
		t.Errorf("expected one warning while offline, got %d", warnings)
	}
	f.pub.err = nil
	f.tick(time.Minute)
	if f.pub.count() != 1 {
		t.Fatalf("next tick must retry, got %d", f.pub.count())
	}
	if f.hdl.warnedState != nil {
		t.Error("warning state must reset after success")
	}
	f.tick(time.Minute)
	if f.pub.count() != 1 {
		t.Errorf("retry succeeded, nothing more to publish, got %d", f.pub.count())
	}
	// Failure after a recorded publish: changed state is retried until it succeeds.
	f.mm.set(200, `{"deployments":[{"module_id":"m"}]}`)
	f.pub.err = errors.New("not connected")
	f.tick(time.Minute)
	f.tick(time.Minute)
	if warnings != 2 {
		t.Errorf("a state change must warn once more, got %d", warnings)
	}
	f.pub.err = nil
	f.tick(time.Minute)
	if f.pub.count() != 2 || f.pub.lastReport(t).State != 2 {
		t.Errorf("unhealthy report must be delivered after retry, got %d", f.pub.count())
	}
}

func TestStartStop(t *testing.T) {
	f := newFixture(t)
	f.hdl.checkInterval = 5 * time.Millisecond
	f.hdl.heartbeatInterval = 0
	f.hdl.now = time.Now
	f.hdl.Start()
	deadline := time.Now().Add(5 * time.Second)
	for f.pub.count() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	f.hdl.Stop()
	n := f.pub.count()
	if n < 2 {
		t.Fatalf("expected periodic publishes, got %d", n)
	}
	time.Sleep(30 * time.Millisecond)
	if f.pub.count() != n {
		t.Error("publishing continued after Stop")
	}
}
