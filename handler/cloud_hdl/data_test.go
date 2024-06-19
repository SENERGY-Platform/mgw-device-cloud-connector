package cloud_hdl

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func TestWriteReadData(t *testing.T) {
	a := data{
		NetworkID: "test",
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
		a.NetworkID = "test 2"
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
		if b.NetworkID != "test 2" {
			t.Error("data file not updated")
		}
	})
}
