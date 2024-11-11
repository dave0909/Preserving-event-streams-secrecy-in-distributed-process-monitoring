package processStateManager

type WorkflowViolation struct {
	//The event id that generated the violation
	//GeneratedByEvent string
	//The case id that generated the violation
	GeneratedByCase string
	//The timestamp of the violation
	Timestamp string
	//The sequence of events that led to the violation
	//ErroneousSequence []string
}

type ComplianceCheckingViolation struct {
	//ViolatedConstraint
	ViolatedConstraint string
	//Involved case
	InvolvedCase string
	//timestamp of the violation
	Timestamp string
}

//type ComplianceCheckingViolation struct {
//	//ViolatedConstraint
//	ViolatedConstraint string
//	//Involved cases
//	InvolvedCases []string
//}
