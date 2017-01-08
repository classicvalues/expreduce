package expreduce

import "fmt"
import "math/big"
import "bytes"

// Floating point numbers represented by big.Float
type Flt struct {
	Val *big.Float
}

func (f *Flt) Eval(es *EvalState) Ex {
	return f
}

func (f *Flt) StringForm(form string) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%.6g", f.Val))
	if bytes.IndexRune(buffer.Bytes(), '.') == -1 {
		buffer.WriteString(".")
	}
	return buffer.String()
}

func (this *Flt) String() string {
	return this.StringForm("InputForm")
}

func (this *Flt) IsEqual(other Ex, cl *CASLogger) string {
	otherConv, ok := other.(*Flt)
	if !ok {
		otherInteger, ok := other.(*Integer)
		if ok {
			otherAsFlt := big.NewFloat(0)
			otherAsFlt.SetInt(otherInteger.Val)
			if otherAsFlt.Cmp(this.Val) == 0 {
				return "EQUAL_TRUE"
			}
		}
		return "EQUAL_UNK"
	}
	if this.Val.Cmp(otherConv.Val) != 0 {
		return "EQUAL_FALSE"
	}
	return "EQUAL_TRUE"
}

func (this *Flt) DeepCopy() Ex {
	tmp := big.NewFloat(0)
	tmp.Copy(this.Val)
	return &Flt{tmp}
}

func IntegerToFlt(i *Integer) (*Flt, bool) {
	newfloat := big.NewFloat(0)
	newfloat.SetInt(i.Val)
	return &Flt{newfloat}, true
}

func RationalToFlt(r *Rational) (*Flt, bool) {
	newfloat := big.NewFloat(0)
	newfloat.SetRat(r.AsBigRat())
	return &Flt{newfloat}, true
}

// Integer numbers represented by big.Int
type Integer struct {
	Val *big.Int
}

func (f *Integer) Eval(es *EvalState) Ex {
	return f
}

func (f *Integer) StringForm(form string) string {
	return fmt.Sprintf("%d", f.Val)
}

func (this *Integer) String() string {
	return this.StringForm("InputForm")
}

func (this *Integer) IsEqual(other Ex, cl *CASLogger) string {
	otherConv, ok := other.(*Integer)
	if !ok {
		otherFlt, ok := other.(*Flt)
		if ok {
			thisAsFlt := big.NewFloat(0)
			thisAsFlt.SetInt(this.Val)
			if thisAsFlt.Cmp(otherFlt.Val) == 0 {
				return "EQUAL_TRUE"
			}
		}
		return "EQUAL_UNK"
	}
	if this.Val.Cmp(otherConv.Val) != 0 {
		return "EQUAL_FALSE"
	}
	return "EQUAL_TRUE"
}

func (this *Integer) DeepCopy() Ex {
	tmp := big.NewInt(0)
	tmp.Set(this.Val)
	return &Integer{tmp}
}

type Rational struct {
	Num *big.Int
	Den *big.Int
}

func (this *Rational) Eval(es *EvalState) Ex {
	if this.Num.Cmp(big.NewInt(0)) == 0 && this.Den.Cmp(big.NewInt(0)) == 0 {
		return &Symbol{"Indeterminate"}
	}
	if this.Den.Cmp(big.NewInt(0)) == 0 {
		return &Symbol{"ComplexInfinity"}
	}
	if this.Num.Cmp(big.NewInt(0)) == 0 {
		return &Integer{big.NewInt(0)}
	}
	negNum := this.Num.Cmp(big.NewInt(0)) == -1
	negDen := this.Den.Cmp(big.NewInt(0)) == -1
	negateRes := negNum != negDen
	absNum := big.NewInt(0)
	absNum.Abs(this.Num)
	absDen := big.NewInt(0)
	absDen.Abs(this.Den)

	gcd := big.NewInt(0)
	gcd.GCD(nil, nil, absNum, absDen)
	absNum.Div(absNum, gcd)
	absDen.Div(absDen, gcd)

	if absDen.Cmp(big.NewInt(1)) == 0 {
		if !negateRes {
			return &Integer{absNum}
		} else {
			return &Integer{absNum.Neg(absNum)}
		}
	}

	if !negateRes {
		this.Num.Set(absNum)
		this.Den.Set(absDen)
		return this
	} else {
		this.Num.Set(absNum.Neg(absNum))
		this.Den.Set(absDen)
		return this
	}
	return this
}

func (this *Rational) StringForm(form string) string {
	return fmt.Sprintf("%d/%d", this.Num, this.Den)
}

func (this *Rational) String() string {
	return this.StringForm("InputForm")
}

func (this *Rational) IsEqual(other Ex, cl *CASLogger) string {
	otherConv, otherIsRational := other.(*Rational)
	if !otherIsRational {
		return "EQUAL_FALSE"
	}
	// Assume rational already simplified
	if (this.Num.Cmp(otherConv.Num) != 0) || (this.Den.Cmp(otherConv.Den) != 0) {
		return "EQUAL_FALSE"
	}
	return "EQUAL_TRUE"
}

func (this *Rational) DeepCopy() Ex {
	tmpn := big.NewInt(0)
	tmpn.Set(this.Num)
	tmpd := big.NewInt(0)
	tmpd.Set(this.Den)
	return &Rational{tmpn, tmpd}
}

func (this *Rational) AsBigRat() *big.Rat {
	return big.NewRat(this.Num.Int64(), this.Den.Int64())
}

type String struct {
	Val string
}

func (this *String) Eval(es *EvalState) Ex {
	return this
}

func (this *String) StringForm(form string) string {
	if form == "OutputForm" {
		return fmt.Sprintf("%v", this.Val)
	}
	return fmt.Sprintf("\"%v\"", this.Val)
}

func (this *String) String() string {
	return this.StringForm("InputForm")
}

func (this *String) IsEqual(other Ex, cl *CASLogger) string {
	otherConv, ok := other.(*String)
	if !ok {
		return "EQUAL_FALSE"
	}
	if this.Val != otherConv.Val {
		return "EQUAL_FALSE"
	}
	return "EQUAL_TRUE"
}

func (this *String) DeepCopy() Ex {
	thiscopy := *this
	return &thiscopy
}

func GetConstantsDefinitions() (defs []Definition) {
	defs = append(defs, Definition{
		Name: "Rational",
		legacyEvalFn: func(this *Expression, es *EvalState) Ex {
			if len(this.Parts) != 3 {
				return this
			}
			nAsInt, nIsInt := this.Parts[1].(*Integer)
			dAsInt, dIsInt := this.Parts[2].(*Integer)
			if nIsInt && dIsInt {
				return (&Rational{nAsInt.Val, dAsInt.Val}).Eval(es)
			}
			return this
		},
		Tests: []TestInstruction{
			&StringTest{"10/7", "Rational[10, 7]"},
			&StringTest{"5/3", "Rational[10, 6]"},
			&StringTest{"Rational[x, 10]", "Rational[x, 10]"},
			&StringTest{"10", "Rational[100, 10]"},
			&StringTest{"-10", "Rational[-100, 10]"},
			&StringTest{"10", "Rational[-100, -10]"},
			&StringTest{"-5/3", "Rational[-10, 6]"},
			&StringTest{"5/3", "Rational[-10, -6]"},
			&StringTest{"0", "Rational[0, 5]"},
			&StringTest{"Rational[0, n]", "Rational[0, n]"},
			&StringTest{"Indeterminate", "Rational[0, 0]"},
			&StringTest{"ComplexInfinity", "Rational[1, 0]"},
			&StringTest{"ComplexInfinity", "Rational[-1, 0]"},
			&StringTest{"ComplexInfinity", "Rational[-1, -0]"},
			&StringTest{"Indeterminate", "Rational[-0, -0]"},
			&StringTest{"Indeterminate", "Rational[-0, 0]"},

			// Rational matching and replacement
			&SameTest{"2/3", "test = Rational[2, 3]"},
			&SameTest{"True", "MatchQ[test, 2/3]"},
			&SameTest{"True", "MatchQ[test, Rational[a_Integer, b_Integer]]"},
			&SameTest{"{2, 3}", "2/3 /. Rational[a_Integer, b_Integer] -> {a, b}"},
			&SameTest{"2/3", "2/3 /. a_Integer/b_Integer -> {a, b}"},
			&SameTest{"buzz[bar]", "foo[bar, 1/2] /. foo[base_, 1/2] -> buzz[base]"},
			&SameTest{"buzz[bar]", "foo[bar, 1/2] /. foo[base_, Rational[1, 2]] -> buzz[base]"},
			&SameTest{"buzz[bar]", "foo[bar, Rational[1, 2]] /. foo[base_, 1/2] -> buzz[base]"},
			&SameTest{"buzz[bar]", "foo[bar, Rational[1, 2]] /. foo[base_, Rational[1, 2]] -> buzz[base]"},
			&SameTest{"True", "MatchQ[1/2, Rational[1, 2]]"},
			&SameTest{"True", "MatchQ[Rational[1, 2], 1/2]"},
			&SameTest{"False", "Hold[Rational[1, 2]] === Hold[1/2]"},
		},
	})
	defs = append(defs, Definition{
		Name: "NumberQ",
		legacyEvalFn: func(this *Expression, es *EvalState) Ex {
			if len(this.Parts) != 2 {
				return this
			}
			_, ok := this.Parts[1].(*Integer)
			if ok {
				return &Symbol{"True"}
			}
			_, ok = this.Parts[1].(*Flt)
			if ok {
				return &Symbol{"True"}
			}
			_, ok = this.Parts[1].(*Rational)
			if ok {
				return &Symbol{"True"}
			}
			return &Symbol{"False"}
		},
		Tests: []TestInstruction{
			&SameTest{"True", "NumberQ[2]"},
			&SameTest{"True", "NumberQ[2.2]"},
			&SameTest{"True", "NumberQ[Rational[5, 2]]"},
			&SameTest{"False", "NumberQ[Infinity]"},
			&SameTest{"False", "NumberQ[Sqrt[2]]"},
			&SameTest{"False", "NumberQ[randomvar]"},
			&SameTest{"False", "NumberQ[\"hello\"]"},
		},
	})
	defs = append(defs, Definition{
		Name: "String",
		Tests: []TestInstruction{
			&SameTest{"\"Hello\"", "\"Hello\""},
			&SameTest{"True", "\"Hello\" == \"Hello\""},
			&SameTest{"False", "\"Hello\" == \"Hello world\""},
		},
	})
	return
}
