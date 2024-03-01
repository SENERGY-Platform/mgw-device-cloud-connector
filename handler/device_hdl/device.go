package device_hdl

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"slices"
)

type device struct {
	ID       string
	MetaHash string
	AttrHash string
	dm_client.Device
}

func newDevice(id string, d dm_client.Device) device {
	var attrPairs []string
	for _, attr := range d.Attributes {
		attrPairs = append(attrPairs, attr.Key+attr.Value)
	}
	slices.Sort(attrPairs)
	return device{
		ID:       id,
		MetaHash: genHash(d.Type, d.State, d.Name, d.ModuleID),
		AttrHash: genHash(attrPairs...),
		Device:   dm_client.Device{},
	}
}

func genHash(str ...string) string {
	hash := sha1.New()
	for _, s := range str {
		hash.Write([]byte(s))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
