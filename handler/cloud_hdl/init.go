package cloud_hdl

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	context_hdl "github.com/SENERGY-Platform/mgw-go-service-base/context-hdl"
	"github.com/SENERGY-Platform/models/go/models"
	"os"
	"path"
	"time"
)

func (h *Handler) Init(ctx context.Context, networkID, networkName string, delay time.Duration) (string, string, error) {
	if !path.IsAbs(h.wrkSpacePath) {
		return "", "", fmt.Errorf("workspace path must be absolute")
	}
	util.Logger.Info(logPrefix, " begin init")
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return "", "", err
	}
	d, err := readData(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	if networkID == "" {
		networkID = d.NetworkID
	}
	ch := context_hdl.New()
	defer ch.CancelAll()
	timer := time.NewTimer(time.Millisecond * 10)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	stop := false
	for !stop {
		select {
		case <-timer.C:
			util.Logger.Debugf("%s get user ID", logPrefix)
			h.userID, err = h.subjectProvider.GetUserID(ch.Add(context.WithCancel(ctx)))
			if err != nil {
				util.Logger.Errorf("%s get user ID: %s", logPrefix, err)
				timer.Reset(delay)
				break
			}
			if h.userID == "" {
				util.Logger.Errorf("%s get user ID: invalid", logPrefix)
				timer.Reset(delay)
				break
			}
			stop = true
			break
		case <-ctx.Done():
			return "", "", fmt.Errorf("get user ID: %s", ctx.Err())
		}
	}
	timer.Reset(time.Millisecond * 10)
	stop = false
	for !stop {
		select {
		case <-timer.C:
			if networkID != "" {
				util.Logger.Debugf("%s get network (%s)", logPrefix, networkID)
				if hb, err := h.cloudClient.GetHub(ch.Add(context.WithCancel(ctx)), networkID); err != nil {
					var nfe *cloud_client.NotFoundError
					if !errors.As(err, &nfe) {
						util.Logger.Errorf("%s get network (%s): %s", logPrefix, networkID, err)
						timer.Reset(delay)
						break
					}
					util.Logger.Warningf("%s get network (%s): %s", logPrefix, networkID, err)
					if networkID != d.NetworkID && d.NetworkID != "" {
						networkID = d.NetworkID
						timer.Reset(time.Millisecond * 10)
						break
					} else {
						networkID = ""
					}
				} else {
					if hb.OwnerId != h.userID {
						util.Logger.Warningf("%s get network (%s): invalid user ID", logPrefix, networkID)
						if networkID != d.NetworkID && d.NetworkID != "" {
							networkID = d.NetworkID
							timer.Reset(time.Millisecond * 10)
							break
						} else {
							networkID = ""
						}
					} else {
						stop = true
						break
					}
				}
			}
			if networkID == "" {
				util.Logger.Info(logPrefix, " create network")
				hID, err := h.cloudClient.CreateHub(ch.Add(context.WithCancel(ctx)), models.Hub{Name: networkName})
				if err != nil {
					util.Logger.Errorf("%s create network: %s", logPrefix, err)
					timer.Reset(delay)
					break
				}
				networkID = hID
				util.Logger.Infof("%s created network (%s)", logPrefix, networkID)
				stop = true
				break
			}
		case <-ctx.Done():
			return "", "", fmt.Errorf("init network: %s", ctx.Err())
		}
	}
	d.NetworkID = networkID
	h.data = d
	return d.NetworkID, h.userID, writeData(h.wrkSpacePath, h.data)
}
