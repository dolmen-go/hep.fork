package rootio

import (
	"bytes"
	B "encoding/binary"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestFileDirectory(t *testing.T) {
	f, err := Open("testdata/small.root")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer f.Close()

	for _, table := range []struct {
		test     string
		value    string
		expected string
	}{
		{"Name", f.Name(), "test-small.root"}, // name when created
		{"Title", f.Title(), "small event file"},
		{"Class", f.Class(), "TFile"},
	} {
		if table.value != table.expected {
			t.Fatalf("%v: expected [%v] got [%v]", table.test, table.expected, table.value)
		}
	}

	for _, table := range []struct {
		name     string
		expected bool
	}{
		{"tree", true},
		{"tree;0", false},
		{"tree;1", true},
		{"tree;9999", true},
		{"tree_nope", false},
		{"tree_nope;0", false},
		{"tree_nope;1", false},
		{"tree_nope;9999", false},
	} {
		_, ok := f.Get(table.name)
		if ok != table.expected {
			t.Fatalf("%s: expected key to exist=%v (got=%v)", table.name, table.expected, ok)
		}
	}

	for _, table := range []struct {
		name     string
		expected string
	}{
		{"tree", "TTree"},
		{"tree;1", "TTree"},
	} {
		k, ok := f.Get(table.name)
		if !ok {
			t.Fatalf("%s: expected key to exist! (got %v)", table.name, ok)
		}

		if k.Class() != table.expected {
			t.Fatalf("%s: expected key with class=%s (got=%v)", table.name, table.expected, k.Class())
		}
	}

	for _, table := range []struct {
		name     string
		expected string
	}{
		{"tree", "tree"},
		{"tree;1", "tree"},
	} {
		k, ok := f.Get(table.name)
		if !ok {
			t.Fatalf("%s: expected key to exist! (got %v)", table.name, ok)
		}

		if k.Name() != table.expected {
			t.Fatalf("%s: expected key with name=%s (got=%v)", table.name, table.expected, k.Name())
		}
	}

	for _, table := range []struct {
		name     string
		expected string
	}{
		{"tree", "my tree title"},
		{"tree;1", "my tree title"},
	} {
		k, ok := f.Get(table.name)
		if !ok {
			t.Fatalf("%s: expected key to exist! (got %v)", table.name, ok)
		}

		if k.Title() != table.expected {
			t.Fatalf("%s: expected key with title=%s (got=%v)", table.name, table.expected, k.Title())
		}
	}
}

// FIXME: this should be done in tree_test
func DontTestFileReader(t *testing.T) {
	f, err := Open("testdata/small.root")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer f.Close()

	//f.Map()
	//return

	getkey := func(n string) Object {
		for i := range f.root.keys {
			k := &f.root.keys[i]
			if k.name == n {
				return k
			}
		}
		t.Fatalf("could not find key [%s]", n)
		return nil
	}

	for _, table := range []struct {
		n string
		t reflect.Type
	}{
		{"Int32", reflect.TypeOf(int32(0))},
		{"Int64", reflect.TypeOf(int64(0))},
		{"UInt32", reflect.TypeOf(uint32(0))},
		{"UInt64", reflect.TypeOf(uint64(0))},
		{"Float32", reflect.TypeOf(float32(0))},
		{"Float64", reflect.TypeOf(float64(0))},

		{"ArrayInt32", reflect.TypeOf([10]int32{})},
		{"ArrayInt64", reflect.TypeOf([10]int64{})},
		{"ArrayUInt32", reflect.TypeOf([10]uint32{})},
		{"ArrayUInt64", reflect.TypeOf([10]uint64{})},
		{"ArrayFloat32", reflect.TypeOf([10]float32{})},
		{"ArrayFloat64", reflect.TypeOf([10]float64{})},
	} {
		obj := getkey(table.n)

		k := obj.(*Key)
		basket := obj.(*Basket)
		data, err := k.Bytes()
		if err != nil {
			t.Fatalf(err.Error())
		}
		buf := bytes.NewBuffer(data)

		if buf.Len() == 0 {
			t.Fatalf("invalid key size")
		}

		if true {
			fd, _ := os.Create("testdata/read_" + table.n + ".bytes")
			io.Copy(fd, buf)
			fd.Close()
			// buf has been consumed...
			buf = bytes.NewBuffer(data)
		}

		for i := 0; i < int(basket.Nevbuf); i++ {
			data := reflect.New(table.t)
			err := B.Read(buf, E, data.Interface())
			if err != nil {
				t.Fatalf("could not read entry [%d]: %v\n", i, err)
			}
			switch table.t.Kind() {
			case reflect.Array:
				for jj := 0; jj < table.t.Len(); jj++ {
					vref := reflect.ValueOf(i).Convert(table.t.Elem())
					vchk := reflect.ValueOf(data.Elem().Index(jj).Interface()).Convert(table.t.Elem())
					if !reflect.DeepEqual(vref, vchk) {
						t.Fatalf("%s: expected data[%d]=%v (got=%v)\n", table.n, jj, vref.Interface(), vchk.Interface())
					}
				}
			default:
				vref := reflect.ValueOf(i).Convert(table.t)
				vchk := reflect.ValueOf(data.Elem().Interface()).Convert(table.t)
				if !reflect.DeepEqual(vref, vchk) {
					t.Fatalf("%s: expected data=%v (got=%v)\n", table.n, vref.Interface(), vchk.Interface())
				}
			}
		}
	}

}
