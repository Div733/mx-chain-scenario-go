package orderedjson

import "testing"

func TestJSONStringNilObjectRendersNull(t *testing.T) {
	if got := JSONString(nil); got != "null" {
		t.Fatalf("expected nil root to render as null, got %q", got)
	}
}

func TestJSONStringNilReceiversRenderNull(t *testing.T) {
	var (
		nilMap    *OJsonMap
		nilList   *OJsonList
		nilString *OJsonString
		nilBool   *OJsonBool
	)

	cases := []struct {
		name string
		obj  OJsonObject
	}{
		{name: "map", obj: nilMap},
		{name: "list", obj: nilList},
		{name: "string", obj: nilString},
		{name: "bool", obj: nilBool},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := JSONString(tc.obj); got != "null" {
				t.Fatalf("expected nil receiver to render as null, got %q", got)
			}
		})
	}
}

func TestJSONStringNilChildrenRenderNull(t *testing.T) {
	root := NewMap()
	root.Put("missing", nil)
	list := OJsonList{nil, &OJsonString{Value: "ok"}}
	root.Put("list", &list)

	got := JSONString(root)
	want := "{\n" +
		"    \"missing\": null,\n" +
		"    \"list\": [\n" +
		"        null,\n" +
		"        \"ok\"\n" +
		"    ]\n" +
		"}"

	if got != want {
		t.Fatalf("unexpected rendered JSON:\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
