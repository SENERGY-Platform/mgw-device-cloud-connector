package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	_ "modernc.org/sqlite"
	"strings"
	"sync"
	"time"
)

const logPrefix = "[sqlite-hdl]"
const pageSize = 4096

type Handler struct {
	db       *sql.DB
	rwMu     sync.RWMutex
	doneChan chan struct{}
}

func New(filePath string) (*Handler, error) {
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, err
	}
	return &Handler{
		db:       db,
		doneChan: make(chan struct{}),
	}, nil
}

func (h *Handler) Close() error {
	<-h.doneChan
	return h.db.Close()
}

func (h *Handler) Init(ctx context.Context, size int64) error {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	_, err := h.db.ExecContext(ctx, messagesTable)
	if err != nil {
		return err
	}
	_, err = h.db.ExecContext(ctx, fmt.Sprintf("PRAGMA max_page_count=%d", size/pageSize))
	if err != nil {
		return err
	}
	return h.optimize(ctx, true)
}

func (h *Handler) PeriodicOptimization(ctx context.Context, interval time.Duration) {
	go h.periodicOptimization(ctx, interval)
}

func (h *Handler) optimize(ctx context.Context, init bool) error {
	stmt := "PRAGMA optimize"
	if init {
		stmt += "=0x10002"
	} else {
		h.rwMu.Lock()
		defer h.rwMu.Unlock()
	}
	_, err := h.db.ExecContext(ctx, stmt+";")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) periodicOptimization(ctx context.Context, interval time.Duration) {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	loop := true
	for loop {
		select {
		case <-ctx.Done():
			loop = false
			break
		case <-timer.C:
			err := h.optimize(ctx, false)
			if err != nil {
				util.Logger.Errorf("%s periodic optimization: %s", logPrefix, err)
			}
			timer.Reset(interval)
		}
	}
	h.doneChan <- struct{}{}
}

func genQuestionMarks(numCol int) string {
	if numCol <= 0 {
		return ""
	}
	if numCol >= 2 {
		return strings.Repeat("?, ", numCol-1) + "?"
	}
	return "?"
}
