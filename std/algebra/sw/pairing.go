/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sw

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/fields"
)

// PairingContext contains useful info about the pairing
type PairingContext struct {
	AteLoop     uint64 // stores the ate loop
	Extension   fields.Extension
	BTwistCoeff fields.E2
}

// LineEvaluation represents a sparse Fp12 Elmt (result of the line evaluation)
type LineEvaluation struct {
	R0, R1, R2 fields.E2
}

// mlStep the i-th ml step contains (f,q) where q=[i]Q, f=f_{i,Q}(P) where (f_{i,Q})=i(Q)-([i]Q)-(i-1)O
type mlStep struct {
	f fields.E12
	q G2Affine
}

// computeLineCoef computes the coefficients of the line passing through Q, R of equation
// x*LineCoeff.R0 +  y*LineCoeff.R1 + LineCoeff.R2
func computeLineCoef(api frontend.API, Q, R G2Affine, ext fields.Extension) LineEvaluation {

	var res LineEvaluation
	res.R0.Sub(api, Q.Y, R.Y)
	res.R1.Sub(api, R.X, Q.X)
	var tmp fields.E2
	res.R2.Mul(api, Q.X, R.Y, ext)
	tmp.Mul(api, R.X, Q.Y, ext)
	res.R2.Sub(api, res.R2, tmp)
	return res
}

// MillerLoop computes the miller loop
func MillerLoop(api frontend.API, P G1Affine, Q G2Affine, res *fields.E12, pairingInfo PairingContext) *fields.E12 {

	var ateLoopBin [64]uint
	var ateLoopBigInt big.Int
	ateLoopBigInt.SetUint64(pairingInfo.AteLoop)
	for i := 0; i < 64; i++ {
		ateLoopBin[i] = ateLoopBigInt.Bit(i)
	}

	res.SetOne(api)
	var l LineEvaluation

	var QCur, QNext G2Affine
	QCur = Q

	fsquareStep := func() {
		QNext.Double(api, &QCur, pairingInfo.Extension).Neg(api, &QNext)
		l = computeLineCoef(api, QCur, QNext, pairingInfo.Extension)
		l.R0.MulByFp(api, l.R0, P.X)
		l.R1.MulByFp(api, l.R1, P.Y)
		res.MulBy034(api, l.R1, l.R0, l.R2, pairingInfo.Extension)
		QCur.Neg(api, &QNext)
	}

	squareStep := func() {
		res.Mul(api, *res, *res, pairingInfo.Extension)
		QNext.Double(api, &QCur, pairingInfo.Extension).Neg(api, &QNext)
		l = computeLineCoef(api, QCur, QNext, pairingInfo.Extension)
		l.R0.MulByFp(api, l.R0, P.X)
		l.R1.MulByFp(api, l.R1, P.Y)
		res.MulBy034(api, l.R1, l.R0, l.R2, pairingInfo.Extension)
		QCur.Neg(api, &QNext)
	}

	mulStep := func() {
		QNext.Neg(api, &QNext).AddAssign(api, &Q, pairingInfo.Extension)
		l = computeLineCoef(api, QCur, Q, pairingInfo.Extension)
		l.R0.MulByFp(api, l.R0, P.X)
		l.R1.MulByFp(api, l.R1, P.Y)
		res.MulBy034(api, l.R1, l.R0, l.R2, pairingInfo.Extension)
		QCur = QNext
	}

	nSquare := func(n int) {
		for i := 0; i < n; i++ {
			squareStep()
		}
	}

	var ml33 mlStep
	fsquareStep()
	nSquare(4)
	mulStep()
	ml33.f = *res
	ml33.q = QCur
	nSquare(7)

	res.Mul(api, *res, ml33.f, pairingInfo.Extension)
	l = computeLineCoef(api, QCur, ml33.q, pairingInfo.Extension)
	l.R0.MulByFp(api, l.R0, P.X)
	l.R1.MulByFp(api, l.R1, P.Y)
	res.MulBy034(api, l.R1, l.R0, l.R2, pairingInfo.Extension)
	QCur.AddAssign(api, &ml33.q, pairingInfo.Extension)

	nSquare(4)
	mulStep()
	squareStep()
	mulStep()

	// remaining 46 bits
	nSquare(46)
	mulStep()

	return res
}
