package cloud_device_hdl

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func TestWriteReadData(t *testing.T) {
	a := data{
		HubID: "test",
		DeviceIDMap: map[string]string{
			"foo": "bar",
		},
	}
	tmpDir := t.TempDir()
	t.Run("read non existent data file", func(t *testing.T) {
		_, err := readData(tmpDir)
		if err == nil {
			t.Error("error ignored")
		}
	})
	t.Run("create data file", func(t *testing.T) {
		err := writeData(tmpDir, a)
		if err != nil {
			t.Error(err)
		}
		_, err = os.Stat(path.Join(tmpDir, dataFile))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("read data file", func(t *testing.T) {
		b, err := readData(tmpDir)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%+v != %+v", a, b)
		}
	})
	t.Run("update data file", func(t *testing.T) {
		a.HubID = "test 2"
		err := writeData(tmpDir, a)
		if err != nil {
			t.Error(err)
		}
		_, err = os.Stat(path.Join(tmpDir, dataFile))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("read data file after update", func(t *testing.T) {
		b, err := readData(tmpDir)
		if err != nil {
			t.Error(err)
		}
		if b.HubID != "test 2" {
			t.Error("data file not updated")
		}
	})
}
