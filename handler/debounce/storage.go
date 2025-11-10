/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package debounce

import (
	"context"
	"database/sql"
	"encoding/json"
	"maps"
	"slices"
	"sync"

	_ "modernc.org/sqlite"
)

type storageProvider interface {
	storeState(state WorkerState) error
	getAllStates() ([]WorkerState, error)
}

type sqliteStorageProvider struct {
	db *sql.DB
}

const queryMigration = `
CREATE TABLE IF NOT EXISTS states
(
    id    TEXT PRIMARY KEY,
    data TEXT NOT NULL
);
PRAGMA optimize =0x10002;
`

const queryInsert = "REPLACE INTO states (id, data) VALUES (?, ?);"
const querySelectAll = "SELECT data from states;"

func newSqliteStorageProvider(filePath string) (storageProvider, error) {
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(queryMigration)
	if err != nil {
		return nil, err
	}
	return &sqliteStorageProvider{
		db: db,
	}, nil
}

func (s *sqliteStorageProvider) storeState(state WorkerState) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(queryInsert, state.Id, string(b))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *sqliteStorageProvider) getAllStates() ([]WorkerState, error) {
	rows, err := s.db.Query(querySelectAll)
	if err != nil {
		return nil, err
	}
	states := []WorkerState{}
	for rows.Next() {
		var stateStr string
		err = rows.Scan(&stateStr)
		if err != nil {
			return nil, err
		}
		var state WorkerState
		err = json.Unmarshal([]byte(stateStr), &state)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

type inMemoryStorageProvider struct {
	db  map[string]WorkerState
	mux sync.RWMutex
}

func newInMemoryStorageProvider() (storageProvider, error) {
	return &inMemoryStorageProvider{
		db:  map[string]WorkerState{},
		mux: sync.RWMutex{},
	}, nil
}

func (s *inMemoryStorageProvider) storeState(state WorkerState) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.db[state.Id] = state
	return nil
}

func (s *inMemoryStorageProvider) getAllStates() ([]WorkerState, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return slices.Collect(maps.Values(s.db)), nil
}
