package message_hdl

import "testing"

func Test_parseTopic(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		subTopic := "foo/bar"
		if !parseTopic(subTopic, "foo/bar") {
			t.Error("expected true")
		}
		t.Run("single level wildcard end", func(t *testing.T) {
			subTopic = "foo/bar/+"
			var arg string
			if !parseTopic(subTopic, "foo/bar/test", &arg) {
				t.Error("expected true")
			}
			if arg != "test" {
				t.Error("expected test, got", arg)
			}
		})
		t.Run("single level wildcard middle", func(t *testing.T) {
			subTopic = "foo/+/bar"
			var arg string
			if !parseTopic(subTopic, "foo/test/bar", &arg) {
				t.Error("expected true")
			}
			if arg != "test" {
				t.Error("expected test, got", arg)
			}
		})
		t.Run("single level wildcard beginning", func(t *testing.T) {
			subTopic = "+/foo/bar"
			var arg string
			if !parseTopic(subTopic, "test/foo/bar", &arg) {
				t.Error("expected true")
			}
			if arg != "test" {
				t.Error("expected test, got", arg)
			}
		})
		t.Run("single level wildcard consecutive", func(t *testing.T) {
			subTopic = "+/+"
			var arg string
			var arg2 string
			if !parseTopic(subTopic, "foo/bar", &arg, &arg2) {
				t.Error("expected true")
			}
			if arg != "foo" {
				t.Error("expected foo, got", arg)
			}
			if arg2 != "bar" {
				t.Error("expected bar, got", arg)
			}
		})
		t.Run("single level wildcard random", func(t *testing.T) {
			subTopic = "+/+/bar/+/foo/+"
			var arg string
			var arg2 string
			var arg3 string
			var arg4 string
			if !parseTopic(subTopic, "a/b/bar/c/foo/d", &arg, &arg2, &arg3, &arg4) {
				t.Error("expected true")
			}
			if arg != "a" {
				t.Error("expected a, got", arg)
			}
			if arg2 != "b" {
				t.Error("expected b, got", arg)
			}
			if arg3 != "c" {
				t.Error("expected c, got", arg)
			}
			if arg4 != "d" {
				t.Error("expected d, got", arg)
			}
		})
		t.Run("multi level wildcard", func(t *testing.T) {
			subTopic = "foo/bar/#"
			var arg string
			if !parseTopic(subTopic, "foo/bar/a", &arg) {
				t.Error("expected true")
			}
			if arg != "a" {
				t.Error("expected a, got", arg)
			}
		})
		t.Run("multi level wildcard", func(t *testing.T) {
			subTopic = "foo/bar/#"
			var arg string
			if !parseTopic(subTopic, "foo/bar/a/b", &arg) {
				t.Error("expected true")
			}
			if arg != "a/b" {
				t.Error("expected a/b, got", arg)
			}
		})
	})
	t.Run("no match", func(t *testing.T) {
		subTopic := "foo/bar"
		if parseTopic(subTopic, "bar/foo") {
			t.Error("expected false")
		}
		t.Run("single level wildcard", func(t *testing.T) {
			subTopic = "foo/bar/+"
			var arg string
			if parseTopic(subTopic, "foo/bar", &arg) {
				t.Error("expected false")
			}
		})
		t.Run("single level wildcard", func(t *testing.T) {
			subTopic = "+/+/bar/+/foo/+"
			var arg string
			var arg2 string
			var arg3 string
			var arg4 string
			if parseTopic(subTopic, "bar/foo", &arg, &arg2, &arg3, &arg4) {
				t.Error("expected false")
			}
		})
		t.Run("multi level wildcard", func(t *testing.T) {
			subTopic = "foo/bar/#"
			if parseTopic(subTopic, "foo/bar") {
				t.Error("expected false")
			}
		})
		t.Run("multi level wildcard", func(t *testing.T) {
			subTopic = "foo/bar/#"
			if parseTopic(subTopic, "bar/foo") {
				t.Error("expected false")
			}
		})
	})
}
