package cloud_hdl

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/models/go/models"
	"slices"
	"strings"
)

func newCloudDevice(device model.Device, rID, attrOrigin string) models.Device {
	var attributes []models.Attribute
	for _, attribute := range device.Attributes {
		attributes = append(attributes, models.Attribute{
			Key:    attribute.Key,
			Value:  attribute.Value,
			Origin: attrOrigin,
		})
	}
	return models.Device{
		Id:           rID,
		LocalId:      device.ID,
		Name:         device.Name,
		Attributes:   attributes,
		DeviceTypeId: device.Type,
	}
}

func notEqual(cDevice models.Device, lDevice model.Device, attrOrigin string) bool {
	var cAttrPairs []string
	for _, attr := range cDevice.Attributes {
		if attr.Origin == attrOrigin {
			cAttrPairs = append(cAttrPairs, attr.Key+attr.Value)
		}
	}
	slices.Sort(cAttrPairs)
	var lAttrPairs []string
	for _, attr := range lDevice.Attributes {
		lAttrPairs = append(lAttrPairs, attr.Key+attr.Value)
	}
	slices.Sort(lAttrPairs)
	return genHash(cDevice.DeviceTypeId, cDevice.Name, strings.Join(cAttrPairs, "")) != genHash(lDevice.Type, lDevice.Name, strings.Join(lAttrPairs, ""))
}

func genHash(str ...string) string {
	hash := sha1.New()
	for _, s := range str {
		hash.Write([]byte(s))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
