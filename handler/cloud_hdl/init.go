package cloud_hdl

import (
	"context"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	context_hdl "github.com/SENERGY-Platform/mgw-go-service-base/context-hdl"
	"net/http"
	"time"
)

func (h *Handler) Init(ctx context.Context, interval time.Duration) (string, string, error) {
	util.Logger.Info(logPrefix, " begin init")
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
	loop := true
	for loop {
		select {
		case <-timer.C:
			util.Logger.Debugf("%s get network", logPrefix)
			networkInfo, err := h.certManagerClient.NetworkInfo(ch.Add(context.WithCancel(ctx)), true, "")
			if err != nil {
				util.Logger.Errorf("%s get network: %s", logPrefix, err)
				timer.Reset(interval)
				break
			}
			if networkInfo.CloudStatus.Code == http.StatusNotFound {
				util.Logger.Errorf("%s get network: cloud backend returned '%s'", logPrefix, http.StatusText(http.StatusNotFound))
				timer.Reset(interval)
				break
			}
			h.networkID = networkInfo.ID
			h.userID = networkInfo.UserID
			loop = false
			break
		case <-ctx.Done():
			return "", "", fmt.Errorf("get network: %s", ctx.Err())
		}
	}
	return h.networkID, h.userID, nil
}
