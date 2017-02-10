package py

import "testing"

func TestDictGetAndSetItem(t *testing.T) {
	d, err := NewDict()
	if err != nil {
		t.Fatal(err)
	}
	k, err := NewString("a")
	if err != nil {
		t.Fatal(err)
	}

	v, err := d.GetItem(k.Object)
	if err != nil {
		t.Error(err)
	} else if v.PyObject != nil {
		t.Error("expected nil")
	}

	vn, err := NewInt(1)
	if err != nil {
		t.Fatal(err)
	}
	err = d.SetItem(k.Object, vn.Object)
	if err != nil {
		t.Fatal(err)
	}

	v, err = d.GetItem(k.Object)
	if err != nil {
		t.Fatal(err)
	}
	n, err := v.GoInt()
	if err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}
