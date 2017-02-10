package py

import "testing"

func mustInt(t *testing.T, n int) Int {
	pn, err := NewInt(n)
	if err != nil {
		t.Fatal(err)
	}
	return pn
}

func mustString(t *testing.T, s string) String {
	ps, err := NewString(s)
	if err != nil {
		t.Fatal(err)
	}
	return ps
}

func TestImportModule(t *testing.T) {
	m, err := ImportModule("this_does_not_exist")
	if err == nil {
		t.Error("expected error, got nil")
	}

	m, err = ImportModule("hello")
	if err != nil {
		t.Fatal(err)
	}
	v, err := m.GetAttrString("n")
	if err != nil {
		t.Fatal(err)
	}
	n, err := v.GoInt()
	if err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestObjectCall(t *testing.T) {
	m, err := ImportModule("hello")
	if err != nil {
		t.Fatal(err)
	}

	fn, err := m.GetAttrString("fn")
	if err != nil {
		t.Fatal(err)
	}

	args := []Object{}
	for _, s := range []string{"foo", "bar", "baz"} {
		ps, err := NewString(s)
		if err != nil {
			t.Fatal(err)
		}
		args = append(args, ps.Object)
	}
	result, err := fn.Call(args...)
	if err != nil {
		t.Fatal(err)
	}
	s, err := result.GoString()
	if err != nil {
		t.Error(err)
	} else if s != "foo bar baz" {
		t.Errorf(`expected "foo bar baz", got %q`, s)
	}
}

func TestObjectRefs(t *testing.T) {
	o := mustString(t, "foo").Object
	defer o.DecRef()

	if rc := int(o.PyObject.ob_refcnt); rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	o.IncRef()
	if rc := int(o.PyObject.ob_refcnt); rc != 2 {
		t.Errorf("expected 2, got %d", rc)
	}
	o.DecRef()
	if rc := int(o.PyObject.ob_refcnt); rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
}

func TestObjectGoInt(t *testing.T) {
	o := mustInt(t, 123).Object
	defer o.DecRef()
	n, err := o.GoInt()
	if err != nil {
		t.Error(err)
	} else if n != 123 {
		t.Errorf("expected 123, got %d", n)
	}

	o = mustString(t, "foo").Object
	defer o.DecRef()
	_, err = o.GoInt()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestObjectGoString(t *testing.T) {
	ps := mustString(t, "foo")
	defer ps.DecRef()
	s, err := ps.Object.GoString()
	if err != nil {
		t.Error(err)
	} else if s != "foo" {
		t.Errorf(`expected "foo", got %q`, s)
	}

	pn := mustInt(t, 123)
	defer pn.DecRef()
	_, err = pn.Object.GoString()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestObjectInt(t *testing.T) {
	pn1 := mustInt(t, 123)
	defer pn1.DecRef()
	pn2, err := pn1.Object.Int()
	if err != nil {
		t.Error(err)
	} else if pn1 != pn2 {
		t.Errorf("pn1 != pn2 (expected %v, got %v)", pn1, pn2)
	}

	o := mustString(t, "foo").Object
	defer o.DecRef()
	_, err = o.Int()
	if err == nil {
		t.Error("expected error, got nil")
	}
}
