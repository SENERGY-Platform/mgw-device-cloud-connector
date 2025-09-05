package sqlite

import (
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/persistent_msg_relay_hdl"
	"path"
	"reflect"
	"testing"
	"time"
)

var largePayload = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque vehicula maximus nunc a pellentesque. Mauris venenatis velit nec neque aliquet ullamcorper. Vivamus in nisi nulla. Maecenas imperdiet felis sit amet augue cursus luctus. Ut gravida urna ac tempor hendrerit. Maecenas interdum purus eu arcu pretium, et efficitur erat scelerisque. Phasellus id arcu vel urna ullamcorper maximus. Donec id augue eget turpis lacinia faucibus et ac turpis. Proin posuere ullamcorper dolor, sed faucibus justo ultricies fringilla. Nam egestas turpis vel risus eleifend venenatis. Nam faucibus lacus orci. Sed sodales, nunc ut condimentum eleifend, diam magna dictum velit, non scelerisque neque lectus quis diam. Ut rutrum eu lorem nec dictum. Duis id pharetra risus, in porttitor eros. Nunc laoreet pulvinar neque, id rhoncus metus vehicula et. Donec vitae malesuada lacus. Morbi at dignissim ex. Phasellus ante velit, congue ut dolor et, efficitur mattis enim. Praesent consequat risus non libero dignissim, et vehicula leo accumsan. Etiam et ornare lectus. In viverra nisl sit amet eros gravida ultricies. Aliquam auctor convallis aliquam. Quisque porta volutpat felis, sit amet semper nisl dignissim at. Vivamus at dui nunc. Donec porttitor mollis diam ac gravida. Nulla vel justo maximus augue ornare pellentesque eu at nibh. Phasellus tincidunt in mi et maximus. Phasellus tempus lorem vitae diam vehicula, mollis sagittis sapien auctor. Vivamus porta lacus in massa viverra, nec sodales felis porta. Vestibulum magna nisi, blandit sed bibendum vel, suscipit vel ante. Donec diam turpis, cursus non pellentesque id, rutrum at enim. Vestibulum ligula turpis, consectetur id nisl vel, euismod porta augue. Nam vehicula libero massa, vestibulum faucibus urna fermentum nec. Aenean a velit sit amet dolor condimentum auctor. Vestibulum vitae ultrices lorem. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Cras tincidunt massa diam, id mollis sapien lobortis nec. Mauris tincidunt elementum sem in rhoncus. Cras quis facilisis mi. Sed non porttitor erat. Aliquam erat volutpat. Donec felis dolor, convallis id risus et, vehicula elementum magna. Quisque convallis sit amet urna et dapibus. Curabitur non dignissim odio. Nulla ac purus sit amet justo imperdiet sollicitudin. Etiam luctus sem vitae fringilla euismod. Mauris eleifend, lacus pellentesque rutrum auctor, enim quam facilisis lacus, non lobortis elit justo ac turpis. Nam ornare scelerisque sem. Nullam tristique dolor at libero pharetra suscipit. Pellentesque quis pretium diam, at luctus eros. Fusce tristique tortor ut quam congue, a aliquet neque viverra. Mauris maximus nisl vitae urna euismod molestie. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nam vehicula pretium eleifend. Vivamus et ligula vel dolor tristique elementum. Donec vitae velit et mi molestie ullamcorper. Nulla facilisi. Donec nisl velit, mollis in maximus eu, pretium pulvinar eros. Fusce tellus magna, vestibulum eget suscipit ac, molestie et nisi. Donec faucibus nibh tortor, eget elementum augue porta non. Aliquam malesuada elit at odio ullamcorper, eu consectetur ipsum gravida. Nam sollicitudin sed diam tristique mattis. Sed auctor vitae neque ut auctor. Morbi eget metus quis augue volutpat fermentum. Pellentesque elementum justo non ante scelerisque elementum. Donec nec felis felis. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam bibendum eleifend ex, non varius dui consequat in. Vivamus dolor erat, auctor ullamcorper mi sit amet, tristique euismod urna. Nullam sed viverra velit. Fusce a velit justo. Nam at dapibus urna. Phasellus ipsum quam, ullamcorper at auctor sed, luctus et neque. Quisque imperdiet ante eu enim lobortis cursus. Fusce commodo efficitur dui vel commodo. Pellentesque sodales mauris augue, in facilisis metus tincidunt ac. Maecenas commodo eros nisl, nec pretium ligula tincidunt vel. Praesent nec risus luctus leo iaculis commodo eget non odio. Aenean erat nisi, dignissim at sollicitudin vitae, condimentum a felis. Nunc et est felis. In et facilisis leo. Quisque ullamcorper rhoncus venenatis. Praesent eu eros vitae erat feugiat sodales. Curabitur hendrerit eleifend leo, sit amet rhoncus tortor semper a. Nulla egestas pellentesque sem, a accumsan orci interdum id. Mauris faucibus et mauris ac placerat. Suspendisse eu ultrices nibh, et porttitor nulla. Vestibulum massa nulla, porttitor sit amet auctor quis, suscipit sed est. In eu metus arcu. Maecenas in consequat tortor, vel facilisis eros. Quisque vel ultrices lorem. Phasellus vestibulum sem ex, nec finibus erat pellentesque a. Ut facilisis neque ac leo accumsan ullamcorper vel in urna. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Ut porttitor gravida augue ut pharetra. Nullam quis sem laoreet, porta tellus id, fringilla massa. Donec convallis laoreet luctus. Morbi dapibus malesuada justo quis tincidunt. Etiam id cursus mauris, vitae molestie sapien. Morbi molestie dapibus lacus, vel pretium eros suscipit quis. Aliquam ac tristique ex. Sed sed lorem non massa mattis tincidunt. Quisque quis viverra leo. Nullam condimentum convallis velit et iaculis. Praesent viverra rhoncus semper. Suspendisse laoreet metus dui, nec euismod nibh condimentum quis. Quisque id tempus turpis. Suspendisse imperdiet ultricies nisi, id faucibus erat molestie sed. Donec ipsum metus, iaculis at bibendum sed, blandit vel massa. Nam libero velit, scelerisque vel nisi nec, laoreet venenatis tortor. Aenean venenatis elit odio, ullamcorper tempor enim convallis non. Nam eleifend leo quis est volutpat, id egestas sem dignissim. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Suspendisse eros tortor, pellentesque dictum turpis a, ultrices aliquam risus. Duis ac rutrum metus. Etiam vulputate quis turpis sed posuere. Nullam iaculis arcu at elit pellentesque tristique. Maecenas bibendum libero dui, id sollicitudin ante suscipit et. Sed a arcu tempor, euismod justo in, vestibulum sem. Nunc quam mauris, pharetra ut vehicula sed, sodales consequat libero. Vivamus elementum id nunc nec bibendum. Nullam ullamcorper maximus purus, ut fringilla justo feugiat a. Nam accumsan pretium mauris, id ullamcorper erat sodales non. Fusce volutpat luctus sodales. In commodo elementum mi, fringilla facilisis eros iaculis et. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vestibulum iaculis lobortis risus, vitae tempus nunc sagittis ut. Donec porta porta nulla, non blandit augue porttitor nec. Cras pellentesque metus magna, quis tempus diam pulvinar quis. Curabitur placerat, ante sed commodo pellentesque, nisl dolor scelerisque tortor, vitae mattis turpis lorem non magna. Quisque quis lectus maximus augue sollicitudin euismod. Nullam cursus egestas dolor. Duis vitae massa vehicula, pretium leo ut, molestie nisi. Nulla facilisis a velit faucibus bibendum. Quisque sollicitudin venenatis ante ut tempus. Donec varius, nunc at rhoncus lobortis, enim ante hendrerit metus, condimentum convallis nibh sapien nec mauris. In metus arcu, faucibus vel semper quis, efficitur sit amet urna. In mattis odio eget massa tristique laoreet eget ut nunc. Aliquam et nunc augue. Donec at enim ex. Nam a hendrerit ipsum. Praesent viverra diam vitae congue placerat. Fusce lacinia congue felis sed pulvinar. Donec efficitur erat sit amet sapien euismod, in volutpat libero laoreet. Etiam faucibus orci vitae augue imperdiet, ac ornare velit tempus. Suspendisse potenti. Aliquam vitae leo et tellus placerat tempor. Sed commodo posuere rutrum. Pellentesque lobortis laoreet finibus. Praesent nunc odio, ultrices ac vestibulum quis, suscipit non sem. Etiam at libero ut ante rutrum eleifend. Mauris mollis tincidunt lorem, at semper tortor fringilla at. Nam mattis, nulla eget cursus ultrices, lorem nibh ultrices purus, nec tempus ipsum ligula fermentum lorem. Fusce tortor nunc, condimentum sed rhoncus vitae, ornare a erat. Cras placerat, diam non rhoncus pretium, leo turpis tincidunt tellus, sit amet gravida magna orci sit amet mauris. Mauris tincidunt dui pulvinar, egestas leo hendrerit, egestas odio. Cras lobortis placerat eros et convallis. In hac habitasse platea dictumst. Sed aliquam malesuada nisi, non pulvinar tortor convallis eget. Cras pretium rutrum laoreet. Fusce tempus magna mauris, eget porttitor lectus dapibus sed. In hac habitasse platea dictumst. Integer congue sem tortor, sit amet molestie eros viverra vel. In convallis elit libero, nec commodo nibh venenatis nec. Aenean ultrices ultricies sapien, sit amet luctus nibh gravida a. Duis dui libero, varius eu lorem sed, consequat vehicula odio. Praesent lobortis, augue eget blandit mollis, mi orci porttitor ex, at congue dui metus ac nisl. In nec libero et enim pharetra consequat vitae quis nunc. Donec id dui nec magna vestibulum euismod. Proin ut congue ex. Quisque metus justo, scelerisque nec ornare sit amet, consectetur et sapien. Curabitur pharetra ante sed nisl hendrerit, id blandit dui vulputate. Vestibulum tincidunt nibh hendrerit dui lobortis, et elementum tellus molestie. Curabitur condimentum malesuada hendrerit. In pulvinar fermentum justo, vel rutrum quam iaculis quis. Ut placerat vulputate risus, id convallis purus mattis sed. Sed interdum varius nisi sit amet scelerisque. Nunc ut dui non nisl sodales fringilla. Integer metus purus, hendrerit at elit eget, condimentum hendrerit erat. Nam posuere commodo massa, eget tempor ante condimentum vel. Curabitur venenatis metus metus, eget porta arcu lacinia eget. Aenean non vulputate dolor. Sed in tincidunt est, at vehicula dui. Sed volutpat ornare hendrerit. Praesent ligula odio, interdum et arcu hendrerit, maximus fermentum sapien. Nunc sed ante iaculis, tincidunt arcu ut, venenatis lectus. Suspendisse ullamcorper ultrices hendrerit. Donec ut malesuada odio. Phasellus erat sem, hendrerit non suscipit vitae, suscipit malesuada velit. Mauris non.")

func TestHandler_CreateAndReadMessages(t *testing.T) {
	a := []persistent_msg_relay_hdl.StorageMessage{
		{
			ID:           "A",
			DayHour:      1,
			Number:       1,
			MsgTopic:     "test",
			MsgPayload:   []byte("test"),
			MsgTimestamp: time.Now().Truncate(0),
		},
		{
			ID:           "B",
			DayHour:      1,
			Number:       2,
			MsgTopic:     "test",
			MsgPayload:   []byte("test"),
			MsgTimestamp: time.Now().Truncate(0),
		},
	}
	m := persistent_msg_relay_hdl.StorageMessage{
		ID:           "C",
		DayHour:      2,
		Number:       1,
		MsgTopic:     "test",
		MsgPayload:   []byte("test"),
		MsgTimestamp: time.Now().Truncate(0),
	}
	t.Run("normal", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("empty", func(t *testing.T) {
			messages, err := h.ReadMessages(t.Context(), 10)
			if err != nil {
				t.Error(err)
			}
			if len(messages) != 0 {
				t.Errorf("expected 0, got %d", len(messages))
			}
		})
		t.Run("create multiple", func(t *testing.T) {
			err = h.CreateMessages(t.Context(), a)
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("create single", func(t *testing.T) {
			err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{m})
			if err != nil {
				t.Error(err)
			}
			a = append(a, m)
		})
		t.Run("read 10", func(t *testing.T) {
			b, err := h.ReadMessages(t.Context(), 10)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(a, b) {
				t.Errorf("expected %v, got %v", a, b)
			}
		})
		t.Run("read 1", func(t *testing.T) {
			b, err := h.ReadMessages(t.Context(), 1)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(a[0], b[0]) {
				t.Errorf("expected %v, got %v", a[0], b[0])
			}
		})
	})
	t.Run("unordered", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Run("create", func(t *testing.T) {
			err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
				a[2],
				a[1],
				a[0],
			})
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("read", func(t *testing.T) {
			b, err := h.ReadMessages(t.Context(), 10)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(a, b) {
				t.Errorf("expected %v, got %v", a, b)
			}
		})
	})
	t.Run("error", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), pageSize)
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Run("size limit", func(t *testing.T) {
			err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
				{
					ID:           "A",
					DayHour:      1,
					Number:       1,
					MsgTopic:     "test",
					MsgPayload:   largePayload,
					MsgTimestamp: time.Now(),
				},
			})
			if err == nil {
				t.Error("expected error")
			}
			if !errors.Is(err, &persistent_msg_relay_hdl.FullErr{}) {
				t.Errorf("expected %T, got %T", &persistent_msg_relay_hdl.FullErr{}, err)
			}
		})
		t.Run("unique", func(t *testing.T) {
			err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
				{
					ID:           "A",
					DayHour:      1,
					Number:       1,
					MsgTopic:     "test",
					MsgPayload:   []byte("test"),
					MsgTimestamp: time.Now(),
				},
				{
					ID:           "A",
					DayHour:      2,
					Number:       2,
					MsgTopic:     "test",
					MsgPayload:   []byte("test"),
					MsgTimestamp: time.Now(),
				},
			})
			if err == nil {
				t.Error("expected error")
			}
			err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
				{
					ID:           "A",
					DayHour:      1,
					Number:       1,
					MsgTopic:     "test",
					MsgPayload:   []byte("test"),
					MsgTimestamp: time.Now(),
				},
				{
					ID:           "b",
					DayHour:      1,
					Number:       1,
					MsgTopic:     "test",
					MsgPayload:   []byte("test"),
					MsgTimestamp: time.Now(),
				},
			})
			if err == nil {
				t.Error("expected error")
			}
		})
	})
}

func TestHandler_LastPosition(t *testing.T) {
	t.Run("multiple rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			{
				ID:           "B",
				DayHour:      1,
				Number:       2,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		dayHour, number, err := h.LastPosition(t.Context())
		if err != nil {
			t.Error(err)
		}
		if dayHour != 1 {
			t.Errorf("expected 1, got %d", dayHour)
		}
		if number != 2 {
			t.Errorf("expected 2, got %d", number)
		}
	})
	t.Run("single row", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		dayHour, number, err := h.LastPosition(t.Context())
		if err != nil {
			t.Error(err)
		}
		if dayHour != 1 {
			t.Errorf("expected 1, got %d", dayHour)
		}
		if number != 1 {
			t.Errorf("expected 1, got %d", number)
		}
	})
	t.Run("no rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = h.LastPosition(t.Context())
		if err == nil {
			t.Error("expected error")
		}
		if !errors.Is(err, persistent_msg_relay_hdl.NoResultsErr) {
			t.Errorf("expected %T, got %T", persistent_msg_relay_hdl.NoResultsErr, err)
		}
	})
}

func TestHandler_NoEntries(t *testing.T) {
	t.Run("multiple rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			{
				ID:           "B",
				DayHour:      1,
				Number:       2,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		ok, err := h.NoEntries(t.Context())
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("expected false")
		}
	})
	t.Run("single row", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		ok, err := h.NoEntries(t.Context())
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("expected false")
		}
	})
	t.Run("no rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		ok, err := h.NoEntries(t.Context())
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Error("expected true")
		}
	})
}

func TestHandler_DeleteMessages(t *testing.T) {
	m := persistent_msg_relay_hdl.StorageMessage{
		ID:           "C",
		DayHour:      2,
		Number:       1,
		MsgTopic:     "test",
		MsgPayload:   []byte("test"),
		MsgTimestamp: time.Now().Truncate(0),
	}
	t.Run("multiple rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			{
				ID:           "B",
				DayHour:      1,
				Number:       2,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			m,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteMessages(t.Context(), []string{"A", "B"})
		if err != nil {
			t.Error(err)
		}
		a := []persistent_msg_relay_hdl.StorageMessage{
			m,
		}
		b, err := h.ReadMessages(t.Context(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("expected %v, got %v", a, b)
		}
	})
	t.Run("single row", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			m,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteMessages(t.Context(), []string{"A"})
		if err != nil {
			t.Error(err)
		}
		a := []persistent_msg_relay_hdl.StorageMessage{
			m,
		}
		b, err := h.ReadMessages(t.Context(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("expected %v, got %v", a, b)
		}
	})
	t.Run("no rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteMessages(t.Context(), []string{"A", "B"})
		if err != nil {
			t.Error(err)
		}
	})
}

func TestHandler_DeleteFirstNMessages(t *testing.T) {
	t.Run("number of rows larger than limit", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		m := persistent_msg_relay_hdl.StorageMessage{
			ID:           "C",
			DayHour:      2,
			Number:       1,
			MsgTopic:     "test",
			MsgPayload:   []byte("test"),
			MsgTimestamp: time.Now().Truncate(0),
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			{
				ID:           "B",
				DayHour:      1,
				Number:       2,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			m,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteFirstNMessages(t.Context(), 2)
		if err != nil {
			t.Error(err)
		}
		a := []persistent_msg_relay_hdl.StorageMessage{
			m,
		}
		b, err := h.ReadMessages(t.Context(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("expected %v, got %v", a, b)
		}
	})
	t.Run("number of rows equal to limit", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
			{
				ID:           "B",
				DayHour:      1,
				Number:       2,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteFirstNMessages(t.Context(), 2)
		if err != nil {
			t.Error(err)
		}
		messages, err := h.ReadMessages(t.Context(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(messages) > 0 {
			t.Errorf("expected no messages")
		}
	})
	t.Run("number of rows less than limit", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.CreateMessages(t.Context(), []persistent_msg_relay_hdl.StorageMessage{
			{
				ID:           "A",
				DayHour:      1,
				Number:       1,
				MsgTopic:     "test",
				MsgPayload:   []byte("test"),
				MsgTimestamp: time.Now().Truncate(0),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteFirstNMessages(t.Context(), 2)
		if err != nil {
			t.Error(err)
		}
		messages, err := h.ReadMessages(t.Context(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(messages) > 0 {
			t.Errorf("expected no messages")
		}
	})
	t.Run("no rows", func(t *testing.T) {
		h, err := New(path.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		err = h.Init(t.Context(), 1048576)
		if err != nil {
			t.Fatal(err)
		}
		err = h.DeleteFirstNMessages(t.Context(), 2)
		if err != nil {
			t.Error(err)
		}
	})
}
