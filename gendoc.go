// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build ignore

// gendoc creates the matrix, mat64 and cmat128 package doc comments.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode/utf8"
)

var docs = template.Must(template.New("docs").Funcs(funcs).Parse(`{{define "common"}}// Generated by running
//  go generate github.com/gonum/matrix
// DO NOT EDIT.

// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package {{.Name}} provides {{.Provides}}
//
// Overview
//
// This section provides a quick overview of the {{.Name}} package. The following
// sections provide more in depth commentary.
//
{{.Overview}}
//{{end}}
{{define "interfaces"}}// The Matrix Interfaces
//
// The Matrix interface is the common link between the concrete types. The Matrix
// interface is defined by three functions: Dims, which returns the dimensions
// of the Matrix, At, which returns the element in the specified location, and
// T for returning a Transpose (discussed later). All of the concrete types can
// perform these behaviors and so implement the interface. Methods and functions
// are designed to use this interface, so in particular the method
//  func (m *Dense) Mul(a, b Matrix)
// constructs a *Dense from the result of a multiplication with any Matrix types,
// not just *Dense. Where more restrictive requirements must be met, there are also the
// Symmetric and Triangular interfaces. For example, in
//  func (s *SymDense) AddSym(a, b Symmetric)
// the Symmetric interface guarantees a symmetric result.
//
// Transposes
//
// The T method is used for transposition. For example, c.Mul(a.T(), b) computes
// c = a^T * b. The {{if .ExamplePackage}}{{.ExamplePackage}}{{else}}{{.Name}}{{end}} types implement this method using an implicit transpose —
// see the Transpose type for more details. Note that some operations have a
// transpose as part of their definition, as in *SymDense.SymOuterK.
//{{end}}
{{define "factorization"}}// Matrix Factorization
//
// Matrix factorizations, such as the LU decomposition, typically have their own
// specific data storage, and so are each implemented as a specific type. The
// factorization can be computed through a call to Factorize
//  var lu {{if .ExamplePackage}}{{.ExamplePackage}}{{else}}{{.Name}}{{end}}.LU
//  lu.Factorize(a)
// The elements of the factorization can be extracted through methods on the
// appropriate type, i.e. *TriDense.LFromLU and *TriDense.UFromLU. Alternatively,
// they can be used directly, as in *Dense.SolveLU. Some factorizations can be
// updated directly, without needing to update the original matrix and refactorize,
// as in *LU.RankOne.
//{{end}}
{{define "blas"}}// BLAS and LAPACK
//
// BLAS and LAPACK are the standard APIs for linear algebra routines. Many
// operations in {{if .Description}}{{.Description}}{{else}}{{.Name}}{{end}} are implemented using calls to the wrapper functions
// in gonum/blas/{{.BLAS|alts}} and gonum/lapack/{{.LAPACK|alts}}. By default, {{.BLAS|join "/"}} and
// {{.LAPACK|join "/"}} call the native Go implementations of the routines. Alternatively,
// it is possible to use C-based implementations of the APIs through the respective
// cgo packages and "Use" functions. The Go implementation of LAPACK makes calls
// through {{.BLAS|join "/"}}, so if a cgo BLAS implementation is registered, the {{.LAPACK|join "/"}}
// calls will be partially executed in Go and partially executed in C.
//{{end}}
{{define "switching"}}// Type Switching
//
// The Matrix abstraction enables efficiency as well as interoperability. Go's
// type reflection capabilities are used to choose the most efficient routine
// given the specific concrete types. For example, in
//  c.Mul(a, b)
// if a and b both implement RawMatrixer, that is, they can be represented as a
// {{.BLAS|alts}}.General, {{.BLAS|alts}}.Gemm (general matrix multiplication) is called, while
// instead if b is a RawSymmetricer {{.BLAS|alts}}.Symm is used (general-symmetric
// multiplication), and if b is a *Vector {{.BLAS|alts}}.Gemv is used.
//
// There are many possible type combinations and special cases. No specific guarantees
// are made about the performance of any method, and in particular, note that an
// abstract matrix type may be copied into a concrete type of the corresponding
// value. If there are specific special cases that are needed, please submit a
// pull-request or file an issue.
//{{end}}
{{define "invariants"}}// Invariants
//
// Matrix input arguments to functions are never directly modified. If an operation
// changes Matrix data, the mutated matrix will be the receiver of a function.
//
// For convenience, a matrix may be used as both a receiver and as an input, e.g.
//  a.Pow(a, 6)
//  v.SolveVec(a.T(), v)
// though in many cases this will cause an allocation (see Element Aliasing).
// An exception to this rule is Copy, which does not allow a.Copy(a.T()).
//{{end}}
{{define "aliasing"}}// Element Aliasing
//
// Most methods in {{if .Description}}{{.Description}}{{else}}{{.Name}}{{end}} modify receiver data. It is forbidden for the modified
// data region of the receiver to overlap the used data area of the input
// arguments. The exception to this rule is when the method receiver is equal to one
// of the input arguments, as in the a.Pow(a, 6) call above, or its implicit transpose.
//
// This prohibition is to help avoid subtle mistakes when the method needs to read
// from and write to the same data region. There are ways to make mistakes using the
// {{.Name}} API, and {{.Name}} functions will detect and complain about those.
// There are many ways to make mistakes by excursion from the {{.Name}} API via
// interaction with raw matrix values.
//
// If you need to read the rest of this section to understand the behavior of
// your program, you are being clever. Don't be clever. If you must be clever,
// {{.BLAS|join "/"}} and {{.LAPACK|join "/"}} may be used to call the behavior directly.
//
// {{if .Description}}{{.Description|sentence}}{{else}}{{.Name}}{{end}} will use the following rules to detect overlap between the receiver and one
// of the inputs:
//  - the input implements one of the Raw methods, and
//  - the Raw type matches that of the receiver, and
//  - the address ranges of the backing data slices overlap, and
//  - the strides differ or there is an overlap in the used data elements.
// If such an overlap is detected, the method will panic.
//
// The following cases will not panic:
//  - the data slices do not overlap,
//  - there is pointer identity between the receiver and input values after
//    the value has been untransposed if necessary.
//
// {{if .Description}}{{.Description|sentence}}{{else}}{{.Name}}{{end}} will not attempt to detect element overlap if the input does not implement a
// Raw method, or if the Raw method differs from that of the receiver except when a
// conversion has occurred through a {{.Name}} API function. Method behavior is undefined
// if there is undetected overlap.
//{{end}}`))

type Package struct {
	path string

	Name           string
	Provides       string
	Description    string
	ExamplePackage string
	Overview       string

	BLAS   []string
	LAPACK []string

	template string
}

var pkgs = []Package{
	{
		path: ".",

		Name:        "matrix",
		Description: "the matrix packages",
		Provides: `common error handling mechanisms for matrix operations
// in mat64 and cmat128.`,
		ExamplePackage: "mat64",

		Overview: `// matrix provides:
//  - Error type definitions
//  - Error recovery mechanisms
//  - Common constants used by mat64 and cmat128
//
// Errors
//
// The mat64 and cmat128 matrix packages share a common set of errors
// provided by matrix via the matrix.Error type.
//
// Errors are either returned directly or used as the parameter of a panic
// depending on the class of error encountered. Returned errors indicate
// that a call was not able to complete successfully while panics generally
// indicate a programmer or unrecoverable error.
//
// Examples of each type are found in the mat64 Solve methods, which find
// x such that A*x = b.
//
// An error value is returned from the function or method when the operation
// can meaningfully fail. The Solve operation cannot complete if A is
// singular. However, determining the singularity of A is most easily
// discovered during the Solve procedure itself and is a valid result from
// the operation, so in this case an error is returned.
//
// A function will panic when the input parameters are inappropriate for
// the function. In Solve, for example, the number of rows of each input
// matrix must be equal because of the rules of matrix multiplication.
// Similarly, for solving A*x = b, a non-zero receiver must have the same
// number of rows as A has columns and must have the same number of columns
// as b. In all cases where a function will panic, conditions that would
// lead to a panic can easily be checked prior to a call.
//
// Error Recovery
//
// When a matrix.Error is the parameter of a panic, the panic can be
// recovered by a Maybe function, which will then return the error.
// Panics that are not of type matrix.Error are re-panicked by the
// Maybe functions.`,
		BLAS:   []string{"blas64", "cblas128"},
		LAPACK: []string{"lapack64", "clapack128"},

		template: `{{template "common" .}}
{{template "invariants" .}}
{{template "aliasing" .}}
package {{.Name}}

// TODO(kortschak) Update docs to indicate the second special case; we
// will check for Vector/Dense overlap because vector extraction from
// a matrix is directly supported by the mat64 API via RowView and ColView.
`,
	},
	{
		path: "mat64",

		Name: "mat64",
		Provides: `implementations of float64 matrix structures and
// linear algebra operations on them.`,

		Overview: `// mat64 provides:
//  - Interfaces for Matrix classes (Matrix, Symmetric, Triangular)
//  - Concrete implementations (Dense, SymDense, TriDense)
//  - Methods and functions for using matrix data (Add, Trace, SymRankOne)
//  - Types for constructing and using matrix factorizations (QR, LU)
//
// A matrix may be constructed through the corresponding New function. If no
// backing array is provided the matrix will be initialized to all zeros.
//  // Allocate a zeroed matrix of size 3×5
//  zero := mat64.NewDense(3, 5, nil)
// If a backing data slice is provided, the matrix will have those elements.
// Matrices are all stored in row-major format.
//  // Generate a 6×6 matrix of random values.
//  data := make([]float64, 36)
//  for i := range data {
//  	data[i] = rand.NormFloat64()
//  }
//  a := mat64.NewDense(6, 6, data)
//
// Operations involving matrix data are implemented as functions when the values
// of the matrix remain unchanged
//  tr := mat64.Trace(a)
// and are implemented as methods when the operation modifies the receiver.
//  zero.Copy(a)
//
// Receivers must be the correct size for the matrix operations, otherwise the
// operation will panic. As a special case for convenience, a zero-sized matrix
// will be modified to have the correct size, allocating data if necessary.
//  var c mat64.Dense // construct a new zero-sized matrix
//  c.Mul(a, a)       // c is automatically adjusted to be 6×6`,

		BLAS:   []string{"blas64"},
		LAPACK: []string{"lapack64"},

		template: `{{template "common" .}}
{{template "interfaces" .}}
{{template "factorization" .}}
{{template "blas" .}}
{{template "switching" .}}
{{template "invariants" .}}
{{template "aliasing" .}}
// BUG(kortschak) Currently only RawMatrixer aliasing detection is supported.
//
package {{.Name}}

// TODO(kortschak) Update docs to indicate the second special case; we
// will check for Vector/Dense overlap because vector extraction from
// a matrix is directly supported by the mat64 API via RowView and ColView.
`,
	},
	{
		path: "cmat128",

		Name: "cmat128",
		Provides: `implementations of complex128 matrix structures and
// linear algebra operations on them.`,

		Overview: `// cmat128 provides:
//  - Interfaces for a complex Matrix`,

		BLAS:   []string{"cblas128"},
		LAPACK: []string{"clapack128"},

		template: `{{template "common" . }}
{{template "blas" .}}
{{template "switching" .}}
{{template "invariants" .}}
{{template "aliasing" .}}
package {{.Name}}

// TODO(kortschak) Update docs to indicate the second special case; we
// will check for Vector/Dense overlap because vector extraction from
// a matrix is directly supported by the mat64 API via RowView and ColView.
`,
	},
}

var funcs = template.FuncMap{
	"sentence": sentence,
	"alts":     alts,
	"join":     join,
}

// sentence converts a string to sentence case where the string is the prefix of the sentence.
func sentence(s string) string {
	if len(s) == 0 {
		return ""
	}
	_, size := utf8.DecodeRune([]byte(s))
	return strings.ToUpper(s[:size]) + s[size:]
}

// alts renders a []string as a glob alternatives list.
func alts(s []string) string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0]
	default:
		return fmt.Sprintf("{%s}", strings.Join(s, ","))
	}
}

// join is strings.Join with the parameter order changed.
func join(sep string, s []string) string {
	return strings.Join(s, sep)
}

func main() {
	for _, pkg := range pkgs {
		t, err := template.Must(docs.Clone()).Parse(pkg.template)
		if err != nil {
			log.Fatalf("failed to parse template: %v", err)
		}
		file := filepath.Join(pkg.path, "doc.go")
		f, err := os.Create(file)
		if err != nil {
			log.Fatalf("failed to create %q: %v", file, err)
		}
		err = t.Execute(f, pkg)
		if err != nil {
			log.Fatalf("failed to execute template: %v", err)
		}
		f.Close()
	}
}
