// Copyright ©2022 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rhist

import (
	"fmt"
	"reflect"

	"go-hep.org/x/hep/groot/rbase"
	"go-hep.org/x/hep/groot/rbytes"
	"go-hep.org/x/hep/groot/root"
	"go-hep.org/x/hep/groot/rtypes"
	"go-hep.org/x/hep/groot/rvers"
)

// Formula describes a ROOT TFormula.
type Formula struct {
	named rbase.Named

	clingParams  []float64 // parameter values
	allParamsSet bool      // flag to control if all parameters are set

	params      map[string]int32 // list of parameter names
	formula     string           // string representing the formula expression
	ndim        int32            // Dimension - needed for lambda expressions
	linearParts []root.Object    // vector of linear functions
	vectorized  bool             // whether we should use vectorized or regular variables
}

func newFormula() *Formula {
	return &Formula{
		named: *rbase.NewNamed("", ""),
	}
}

func (*Formula) RVersion() int16 {
	return rvers.Formula
}

func (*Formula) Class() string {
	return "TFormula"
}

// Name returns the name of the instance
func (f *Formula) Name() string {
	return f.named.Name()
}

// Title returns the title of the instance
func (f *Formula) Title() string {
	return f.named.Title()
}

// MarshalROOT implements rbytes.Marshaler
func (f *Formula) MarshalROOT(w *rbytes.WBuffer) (int, error) {
	if w.Err() != nil {
		return 0, w.Err()
	}

	pos := w.WriteVersion(f.RVersion())
	w.WriteObject(&f.named)
	w.WriteStdVectorF64(f.clingParams)
	w.WriteBool(f.allParamsSet)
	writeMapStringInt(w, f.params)
	w.WriteString(f.formula)
	w.WriteI32(f.ndim)
	writeStdVectorObjP(w, f.linearParts)
	w.WriteBool(f.vectorized)

	return w.SetByteCount(pos, f.Class())
}

func (f *Formula) UnmarshalROOT(r *rbytes.RBuffer) error {
	if r.Err() != nil {
		return r.Err()
	}

	beg := r.Pos()
	vers, pos, bcnt := r.ReadVersion(f.Class())
	if vers > rvers.Formula {
		panic(fmt.Errorf("rhist: invalid TFormula version=%d > %d", vers, rvers.Formula))
	}

	if vers < 12 || vers > 13 {
		// tested with v12 and v13
		panic(fmt.Errorf("rhist: too old TFormula version=%d < 12", vers))
	}

	r.ReadObject(&f.named)
	r.ReadStdVectorF64(&f.clingParams)
	f.allParamsSet = r.ReadBool()
	f.params = readMapStringInt(r)
	f.formula = r.ReadString()
	f.ndim = r.ReadI32()

	f.linearParts = readStdVectorObjP(r)
	f.vectorized = r.ReadBool()

	r.CheckByteCount(pos, bcnt, beg, f.Class())
	return r.Err()
}

func (f *Formula) String() string {
	return fmt.Sprintf("TFormula{%s}", f.formula)
}

func readMapStringInt(r *rbytes.RBuffer) map[string]int32 {
	if r.Err() != nil {
		return nil
	}
	const typename = "map<TString,int,TFormulaParamOrder>"
	beg := r.Pos()
	vers, pos, bcnt := r.ReadVersion(typename)
	if vers != rvers.StreamerInfo {
		r.SetErr(fmt.Errorf("rbytes: invalid %s version: got=%d, want=%d",
			typename, vers, rvers.StreamerInfo,
		))
		return nil
	}
	n := int(r.ReadI32())
	o := make(map[string]int32, n)
	for i := 0; i < n; i++ {
		k := r.ReadString()
		v := r.ReadI32()
		o[k] = v
	}
	r.CheckByteCount(pos, bcnt, beg, typename)
	return o
}

func readStdVectorObjP(r *rbytes.RBuffer) []root.Object {
	if r.Err() != nil {
		return nil
	}
	const typename = "vector<TObject*>"
	beg := r.Pos()
	vers, pos, bcnt := r.ReadVersion(typename)
	if vers != rvers.StreamerInfo {
		r.SetErr(fmt.Errorf("rbytes: invalid %s version: got=%d, want=%d",
			typename, vers, rvers.StreamerInfo,
		))
		return nil
	}
	n := int(r.ReadI32())
	o := make([]root.Object, n)
	for i := range o {
		o[i] = r.ReadObjectAny()
	}
	r.CheckByteCount(pos, bcnt, beg, typename)
	return o
}

func writeMapStringInt(w *rbytes.WBuffer, m map[string]int32) {
	if w.Err() != nil {
		return
	}
	const typename = "map<TString,int,TFormulaParamOrder>"
	pos := w.WriteVersion(rvers.StreamerInfo)
	w.WriteI32(int32(len(m)))
	// FIXME(sbinet): write in correct order?
	for k, v := range m {
		w.WriteString(k)
		w.WriteI32(v)
	}
	_, _ = w.SetByteCount(pos, typename)
}

func writeStdVectorObjP(w *rbytes.WBuffer, vs []root.Object) {
	if w.Err() != nil {
		return
	}
	const typename = "vector<TObject*>"
	pos := w.WriteVersion(rvers.StreamerInfo)
	w.WriteI32(int32(len(vs)))
	for i := range vs {
		w.WriteObjectAny(vs[i])
	}
	_, _ = w.SetByteCount(pos, typename)
}

func init() {
	{
		f := func() reflect.Value {
			o := newFormula()
			return reflect.ValueOf(o)
		}
		rtypes.Factory.Add("TFormula", f)
	}
}

var (
	_ root.Object        = (*Formula)(nil)
	_ root.Named         = (*Formula)(nil)
	_ rbytes.Marshaler   = (*Formula)(nil)
	_ rbytes.Unmarshaler = (*Formula)(nil)
)
