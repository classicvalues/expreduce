package cas

func (this *Expression) EvalIf(es *EvalState) Ex {
	if len(this.Parts) != 4 {
		return this
	}

	var isequal string = this.Parts[1].Eval(es).IsEqual(&Symbol{"True"}, &es.CASLogger)
	if isequal == "EQUAL_UNK" {
		return this
	} else if isequal == "EQUAL_TRUE" {
		return this.Parts[2].Eval(es)
	} else if isequal == "EQUAL_FALSE" {
		return this.Parts[3].Eval(es)
	}

	return &Expression{[]Ex{&Symbol{"Error"}, &String{"Unexpected equality return value."}}}
}

func (this *Expression) EvalWhile(es *EvalState) Ex {
	if len(this.Parts) != 3 {
		return this
	}
	isequal := this.Parts[1].Eval(es).IsEqual(&Symbol{"True"}, &es.CASLogger)
	cont := isequal == "EQUAL_TRUE"
	for cont {

		isequal = this.Parts[1].Eval(es).IsEqual(&Symbol{"True"}, &es.CASLogger)
		cont = isequal == "EQUAL_TRUE"
	}

	if isequal == "EQUAL_UNK" {
		return &Expression{[]Ex{&Symbol{"Error"}, &String{"Encountered EQUAL_UNK when evaluating test for the While."}}}
	} else if isequal == "EQUAL_TRUE" {
		return &Symbol{"Null"}
	} else if isequal == "EQUAL_FALSE" {
		return &Symbol{"Null"}
	}

	return &Expression{[]Ex{&Symbol{"Error"}, &String{"Unexpected equality return value."}}}
}
