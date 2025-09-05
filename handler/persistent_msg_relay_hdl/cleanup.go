package persistent_msg_relay_hdl

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"os"
	"path"
)

func (h *Handler) cleanup(ctx context.Context) error {
	file, err := os.Open(path.Join(h.workSpacePath, cleanupFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer func() {
		file.Close()
		if e := os.Remove(path.Join(h.workSpacePath, cleanupFile)); e != nil {
			util.Logger.Errorf("%s remove cleanup file: %s", logPrefix, e)
		}
	}()
	var msgIDs []string
	err = json.NewDecoder(file).Decode(&msgIDs)
	if err != nil {
		return err
	}
	if len(msgIDs) > 0 {
		err = h.storageHdl.DeleteMessages(ctx, msgIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) createCleanupFile(msgIDs []string) error {
	file, err := os.Create(path.Join(h.workSpacePath, cleanupFile))
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(msgIDs)
}
