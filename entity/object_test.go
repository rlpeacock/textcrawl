package entity

import (
	"testing"
)

func TestDeserializeAttrib(t *testing.T) {
	_, e := DeserializeAttrib("")
	if e == nil {
		t.Error(`DeserializeAttrib("") should have returned an error.`)
	}
	_, e = DeserializeAttrib("1")
	if e == nil {
		t.Error(`DeserializeAttrib("1") should have returned an error.`)
	}
	_, e = DeserializeAttrib("1:")
	if e == nil {
		t.Error(`DeserializeAttrib("1:") should have returned an error.`)
	}
	_, e = DeserializeAttrib("a:1")
	if e == nil {
		t.Error(`DeserializeAttrib("a:1") should have returned an error.`)
	}
	_, e = DeserializeAttrib("1:1:1")
	if e == nil {
		t.Error(`DeserializeAttrib("1:1:1") should have returned an error.`)
	}
	_, e = DeserializeAttrib("1:1.1")
	if e == nil {
		t.Errorf(`DeserializeAttrib("1:1.1") should have returned an error. Got: %s`, e)
	}
	_, e = DeserializeAttrib("1:1")
	if e != nil {
		t.Errorf(`DeserializeAttrib("1:1") should not have returned an error. Got: %s`, e)
	}
	a, e := DeserializeAttrib("1:2")
	if a.Cur != 2 {
		t.Errorf(`DeserializeAttrib("1:2") should have got a Cur value of 2 but got %d`, a.Cur)
	}
	a, e = DeserializeAttrib("1:2")
	if a.Real != 1 {
		t.Errorf(`DeserializeAttrib("1:2") should have got a Real value of 1 but got %d`, a.Cur)
	}
}

func TestDeserializeAttribList(t *testing.T) {
	a := Attrib{}
	e := DeserializeAttribList("", &a)
	if e == nil {
		t.Errorf(`DeserializeAttribList("", &a) should have returned an error`)
	}
	e = DeserializeAttribList("1:2,", &a)
	if e == nil {
		t.Errorf(`DeserializeAttribList("1:2,", &a) should have returned an error`)
	}
	e = DeserializeAttribList("1:2", &a)
	if e != nil {
		t.Errorf(`DeserializeAttribList("1:2", &a) returned an error: %s`, e)
	}
	if a.Real != 1 {
		t.Errorf(`DeserializeAttribList("1:2", &a) a should have Real value 1 but got %d`, a.Real)
	}
	if a.Cur != 2 {
		t.Errorf(`DeserializeAttribList("1:2", &a) a should have Cur value 2 but got %d`, a.Cur)
	}
	b, c := Attrib{}, Attrib{}
	e = DeserializeAttribList("1:2,3:4", &a, &b, &c)
	if e == nil {
		t.Errorf(`DeserializeAttribList("1:2,3:4", &a, &b, &c) should have returned an error`)
	}
	e = DeserializeAttribList("1:2,3:4,5:6", &a, &b)
	if e == nil {
		t.Errorf(`DeserializeAttribList("1:2,3:4,5:6", &a, &b) should have returned an error`)
	}
	e = DeserializeAttribList("1:2,3:4,5:6", &a, &b, &c)
	if e != nil {
		t.Errorf(`DeserializeAttribList("1:2,3:4,5:6", &a, &b, &c) returned an error: %s`, e)
	}
	if a.Cur != 2 || b.Cur != 4 || c.Cur != 6 {
		t.Errorf(`DeserializeAttribList("1:2,3:4,5:6", &a, &b, &c) returned a bad cur value. Expected, 2,4,6 but got %d,%d,%d`, a.Cur, b.Cur, c.Cur)
	}
	if a.Real != 1 || b.Real != 3 || c.Real != 5 {
		t.Errorf(`DeserializeAttribList("1:2,3:4,5:6", &a, &b, &c) returned a bad real value. Expected, 1,3,5 but got %d,%d,%d`, a.Real, b.Real, c.Real)
	}
}
