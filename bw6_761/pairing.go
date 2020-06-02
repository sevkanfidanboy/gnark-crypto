// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gurvy/internal/generators DO NOT EDIT

package bw6_761

// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func (curve *Curve) FinalExponentiation(z *PairingResult, _z ...*PairingResult) PairingResult {
	var result PairingResult
	result.Set(z)

	// if additional parameters are provided, multiply them into z
	for _, e := range _z {
		result.Mul(&result, e)
	}

	result.FinalExponentiation(&result)

	return result
}

// FinalExponentiation sets z to the final expo x**((p**6 - 1)/r), returns z
func (z *PairingResult) FinalExponentiation(x *PairingResult) *PairingResult {
	// easy part: x**((p**3 - 1)*(p+1))
	var result PairingResult
	result.Set(x)

	var t [1]PairingResult // temp memory

	t[0].FrobeniusCube(&result) // x**(p**3)
	result.Inverse(&result)     // x**(-1)
	t[0].Mul(&t[0], &result)    // x**(p**3-1)
	result.Frobenius(&t[0]).    // x**((p**3-1)*p)
					Mul(&result, &t[0]) // x**((p**3-1)*(p+1))

	// hard part (up to permutation)
	// performs the hard part of the final expo
	// Algorithm 1 of https://eprint.iacr.org/2016/130.pdf
	// The result is the same as p**4-p**2+1/r, but up to permutation (it's 3* (p**4 -p**2 +1 /r)), ok since r=1 mod 3)

	// TODO
	z.Set(&result)
	return z
}

// MillerLoop Miller loop
func (curve *Curve) MillerLoop(P G1Affine, Q G2Affine, result *PairingResult) *PairingResult {

	// init result
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return result
	}

	// the line goes through QCur and QNext
	var QCur, QNext, QNextNeg G2Jac
	var QNeg G2Affine

	// Stores -Q
	QNeg.Neg(&Q)

	// init QCur with Q
	Q.ToJacobian(&QCur)

	var lEval lineEvalRes

	// Miller loop
	for i := len(curve.loopCounter) - 2; i >= 0; i-- {
		QNext.Set(&QCur)
		QNext.Double()
		QNextNeg.Neg(&QNext)

		result.Square(result)

		// evaluates line though Qcur,2Qcur at P
		lineEvalJac(QCur, QNextNeg, &P, &lEval)
		lEval.mulAssign(result)

		if curve.loopCounter[i] == 1 {
			// evaluates line through 2Qcur, Q at P
			lineEvalAffine(QNext, Q, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&Q)

		} else if curve.loopCounter[i] == -1 {
			// evaluates line through 2Qcur, -Q at P
			lineEvalAffine(QNext, QNeg, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&QNeg)
		}
		QCur.Set(&QNext)
	}

	return result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
// The case in which Q=R=Infinity is not handled as this doesn't happen in the SNARK pairing
func lineEvalJac(Q, R G2Jac, P *G1Affine, result *lineEvalRes) {
	// converts Q and R to projective coords
	Q.ToProjFromJac()
	R.ToProjFromJac()

	// line eq: w^3*(QyRz-QzRy)x +  w^2*(QzRx - QxRz)y + w^5*(QxRy-QyRxz)
	// result.r1 = QyRz-QzRy
	// result.r0 = QzRx - QxRz
	// result.r2 = QxRy-QyRxz

	result.r1.Mul(&Q.Y, &R.Z)
	result.r0.Mul(&Q.Z, &R.X)
	result.r2.Mul(&Q.X, &R.Y)

	Q.Z.Mul(&Q.Z, &R.Y)
	Q.X.Mul(&Q.X, &R.Z)
	Q.Y.Mul(&Q.Y, &R.X)

	result.r1.Sub(&result.r1, &Q.Z)
	result.r0.Sub(&result.r0, &Q.X)
	result.r2.Sub(&result.r2, &Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.Mul(&result.r1, &P.X)
	result.r0.Mul(&result.r0, &P.Y)
	//result.r2.Mul(&result.r2, &P.Z)
}

// Same as above but R is in affine coords
func lineEvalAffine(Q G2Jac, R G2Affine, P *G1Affine, result *lineEvalRes) {

	// converts Q and R to projective coords
	Q.ToProjFromJac()

	// line eq: w^3*(QyRz-QzRy)x +  w^2*(QzRx - QxRz)y + w^5*(QxRy-QyRxz)
	// result.r1 = QyRz-QzRy
	// result.r0 = QzRx - QxRz
	// result.r2 = QxRy-QyRxz

	result.r1.Set(&Q.Y)
	result.r0.Mul(&Q.Z, &R.X)
	result.r2.Mul(&Q.X, &R.Y)

	Q.Z.Mul(&Q.Z, &R.Y)
	Q.Y.Mul(&Q.Y, &R.X)

	result.r1.Sub(&result.r1, &Q.Z)
	result.r0.Sub(&result.r0, &Q.X)
	result.r2.Sub(&result.r2, &Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.Mul(&result.r1, &P.X)
	result.r0.Mul(&result.r0, &P.Y)
	// result.r2.Mul(&result.r2, &P.Z)
}

type lineEvalRes struct {
	r0 G2CoordType // c0.b1
	r1 G2CoordType // c1.b1
	r2 G2CoordType // c1.b2
}

func (l *lineEvalRes) mulAssign(z *PairingResult) *PairingResult {

	return z
}

const tAbsVal uint64 = 9586122913090633729

// Expt set z to x^t in PairingResult and return z
// TODO make a ExptAssign method that assigns the result to self; then this method can assert fail if z != x
// TODO Expt is the only method that depends on tAbsVal.  The rest of the tower does not depend on this value.  Logically, Expt should be separated from the rest of the tower.
func (z *PairingResult) Expt(x *PairingResult) *PairingResult {
	// TODO what if x==0?
	// TODO make this match Element.Exp: x is a non-pointer?

	// tAbsVal in binary: 1000010100001000110000000000000000000000000000000000000000000001
	// drop the low 46 bits (all 0 except the least significant bit): 100001010000100011 = 136227
	// Shortest addition chains can be found at https://wwwhomes.uni-bielefeld.de/achim/addition_chain.html

	var result, x33 PairingResult

	// a shortest addition chain for 136227
	result.Set(x)             // 0                1
	result.Square(&result)    // 1( 0)            2
	result.Square(&result)    // 2( 1)            4
	result.Square(&result)    // 3( 2)            8
	result.Square(&result)    // 4( 3)           16
	result.Square(&result)    // 5( 4)           32
	result.Mul(&result, x)    // 6( 5, 0)        33
	x33.Set(&result)          // save x33 for step 14
	result.Square(&result)    // 7( 6)           66
	result.Square(&result)    // 8( 7)          132
	result.Square(&result)    // 9( 8)          264
	result.Square(&result)    // 10( 9)          528
	result.Square(&result)    // 11(10)         1056
	result.Square(&result)    // 12(11)         2112
	result.Square(&result)    // 13(12)         4224
	result.Mul(&result, &x33) // 14(13, 6)      4257
	result.Square(&result)    // 15(14)         8514
	result.Square(&result)    // 16(15)        17028
	result.Square(&result)    // 17(16)        34056
	result.Square(&result)    // 18(17)        68112
	result.Mul(&result, x)    // 19(18, 0)     68113
	result.Square(&result)    // 20(19)       136226
	result.Mul(&result, x)    // 21(20, 0)    136227

	// the remaining 46 bits
	for i := 0; i < 46; i++ {
		result.Square(&result)
	}
	result.Mul(&result, x)

	z.Set(&result)
	return z
}
