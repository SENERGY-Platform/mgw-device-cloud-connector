package cloud_device_hdl

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
	decoder := json.NewDecoder(f)
	var d data
	if err = decoder.Decode(&d); err != nil {
		return data{}, err
	}
	return d, nil
}

func writeData(p string, d data) error {
	f, err := os.OpenFile(path.Join(p, dataFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	return encoder.Encode(d)
}
