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

func TestSerializeAttrib(t *testing.T) {
	a := Attrib{
		Real:1,
		Cur:2,
	}
	s := SerializeAttrib(a)
	if s != "1:2" {
		t.Errorf("attrib should have serlialized to 1:2 but got '%s'", s)
	}
}

func TestSerializeAttribList(t *testing.T) {
	a := []Attrib{
		{1,2},
		{3,4},
		{5,6},
	}
	s := SerializeAttribList(a...)
	if s != "1:2,3:4,5:6" {
		t.Errorf("attrib list should have serialized to 1:2,3:4,5:6 but got '%s'", s)
	}	
}

func TestLoadObjects(t *testing.T) {
	db, err := openDB("1")
	if err != nil {
		t.Errorf("Unable to open DB: %v", err)
	}
	things := LoadThings(db)
	t1 := things["T1"]
	if t1 == nil {
		t.Errorf("failed to find T1 in DB")
	}
	if t1.Id != "T1" {
		t.Errorf("expected ID T1 but got '%s'", t1.Id)
	}
	if t1.Title != "tin knife" {
		t.Errorf("expected T1 title to be 'tin knife' but was '%s'", t1.Title)
	}
	if t1.Weight.Real != 1 || t1.Weight.Cur != 1 {
		t.Errorf("T1 weight should be 1:1 but was %d:%d", t1.Weight.Real, t1.Weight.Cur)
	}
	if t1.Size.Real != 2 || t1.Size.Cur != 2 {
		t.Errorf("T1 size should be 2:2 but was %d:%d", t1.Size.Real, t1.Size.Cur)
	}
	if t1.Durability.Real != 3 || t1.Durability.Cur != 3 {
		t.Errorf("T1 durability should be 3:3 but was %d:%d", t1.Durability.Real, t1.Durability.Cur)
	}
	t2 := things["T2"]
	if t2.Id != "T2" {
		t.Errorf("expected ID T2 but got '%s'", t2.Id)		
	}
	if t2.Title != "a man" {
		t.Errorf("expected T2 title to be 'a man' but was '%s'", t1.Title)		
	}
}
