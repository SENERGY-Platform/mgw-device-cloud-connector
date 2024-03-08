package cloud_hdl

import (
	"encoding/json"
	"os"
	"path"
)

const dataFile = "data.json"

type data struct {
	HubID          string            `json:"hub_id"`
	DefaultHubName string            `json:"-"`
	DeviceIDMap    map[string]string `json:"device_id_map"` // localID:ID
}

func readData(p string) (data, error) {
	f, err := os.Open(path.Join(p, dataFile))
	if err != nil {
		return data{}, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var hi data
	if err = d.Decode(&hi); err != nil {
		return data{}, err
	}
	return hi, nil
}

func writeData(p string, hi data) error {
	f, err := os.OpenFile(path.Join(p, dataFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	return e.Encode(hi)
}
