package main

import (
	"fmt"
	"reflect"
	"strconv"
)

// (ConnectingRule) OR (Rule)
type Evaluator int

const (
	EQUAL Evaluator = iota
	NOTEQUAL
	GREATER
	LESSER
	GREATERANDEQUAL
	LESSERANDEQUAL
)

type Rule struct {
	ID           string
	VariableName string
	Evaluator    Evaluator
	ValueString  string
	NotOperation bool
}

type LogicalOperation int

const (
	ANDOPERATION LogicalOperation = iota
	OROPERATION
)

type ConnectingRule struct {
	ID               string
	LogicalOperation LogicalOperation
	RuleA            *Rule
	RuleB            *ConnectingRule
}

type UserAttribute struct {
	ID               string
	Country          string
	KycVerified      bool
	MonthlyVolume    int64 //Defined int here so that float overflow is not a miss
	AccountAgeInDays int64
	IsBlocked        bool
	AppVersion       string
}

// Constructors
func NewRule(ID string, VariableName string, Evaluator Evaluator, ValueString string, NotOperation bool) *Rule {
	return &Rule{
		ID:           ID,
		VariableName: VariableName,
		Evaluator:    Evaluator,
		ValueString:  ValueString,
		NotOperation: NotOperation,
	}
}
func NewUserAttribute(ID string, Country string, KycVerified bool, MonthlyVolume int64, AccountAgeInDays int64, IsBlocked bool, AppVersion string) *UserAttribute {
	return &UserAttribute{
		ID:               ID,
		Country:          Country,
		KycVerified:      KycVerified,
		MonthlyVolume:    MonthlyVolume,
		AccountAgeInDays: AccountAgeInDays,
		IsBlocked:        IsBlocked,
		AppVersion:       AppVersion,
	}
}
func NewConnectingRule(ID string, LogicalOperation LogicalOperation, RuleA *Rule, RuleB *ConnectingRule) *ConnectingRule {
	return &ConnectingRule{
		ID:               ID,
		LogicalOperation: LogicalOperation,
		RuleA:            RuleA,
		RuleB:            RuleB,
	}
}

// DB Storage In memory
type DB struct {
	UserAttributes map[string]*UserAttribute
	Rules          map[string]*Rule
	ConnectingRule map[string]*ConnectingRule
}

func NewDB() *DB {
	return &DB{
		UserAttributes: make(map[string]*UserAttribute),
		Rules:          make(map[string]*Rule),
		ConnectingRule: make(map[string]*ConnectingRule),
	}
}
func (d *DB) NewUserAttribute(ID string, Country string, KycVerified bool, MonthlyVolume int64, AccountAgeInDays int64, IsBlocked bool, AppVersion string) *UserAttribute {
	user := NewUserAttribute(ID, Country, KycVerified, MonthlyVolume, AccountAgeInDays, IsBlocked, AppVersion)
	d.UserAttributes[ID] = user
	return user
}
func (d *DB) NewRuleAddition(Rule *Rule) {
	d.Rules[Rule.ID] = Rule
}

func (d *DB) NewConnectingRuleAddition(ID string, RuleA *Rule, RuleB *ConnectingRule, LogicalOperation LogicalOperation) *ConnectingRule {
	cr := NewConnectingRule(ID, LogicalOperation, RuleA, RuleB)
	d.ConnectingRule[ID] = cr
	return cr
}

// Rule Evaluator Service
type RuleEvaluator struct {
	DB *DB
}

func NewRuleEvaluator(DB *DB) *RuleEvaluator {
	return &RuleEvaluator{
		DB: DB,
	}
}

func (r *RuleEvaluator) Evaluate(userAttribute *UserAttribute, connectingRule *ConnectingRule) bool {
	var rule1Analysis bool = false
	var rule2Analysis bool = false
	if connectingRule.RuleA != nil {
		rule1Analysis = r.EvaluateIndividualRule(userAttribute, connectingRule.RuleA)
	}
	if connectingRule.RuleA != nil {
		rule2Analysis = r.Evaluate(userAttribute, connectingRule.RuleB) // Connecting Rule Needs Recursion over here.
	}

	var resultAnalysis bool

	switch connectingRule.LogicalOperation {
	case ANDOPERATION:
		resultAnalysis = rule1Analysis && rule2Analysis
	case OROPERATION:
		resultAnalysis = rule1Analysis || rule2Analysis
	default:
		resultAnalysis = false
	}

	return resultAnalysis
}

func (r *RuleEvaluator) EvaluateIndividualRule(userAttribute *UserAttribute, rule *Rule) bool {
	result := false
	switch rule.Evaluator {
	case EQUAL:
		equalEval := &EqualEvaluatorStrategy{}
		if equalEval.Validate(userAttribute, rule) {
			result = equalEval.Calculate(userAttribute, rule)
			if rule.NotOperation {
				result = !result
			}
		}
	case GREATER:
		greaterEval := &GreaterEvaluatorStrategy{}
		if greaterEval.Validate(userAttribute, rule) {
			result = greaterEval.Calculate(userAttribute, rule)
			if rule.NotOperation {
				result = !result
			}
		}
	}

	return result
}

type EvaluatorStrategy interface {
	Validate(userAttribute *UserAttribute, rule *Rule) bool
	Calculate(userAttribute *UserAttribute, rule *Rule) bool
}

type EqualEvaluatorStrategy struct{}

func (eq *EqualEvaluatorStrategy) Validate(userAttribute *UserAttribute, rule *Rule) bool {
	result := false
	t := reflect.ValueOf(userAttribute)
	r := rule.VariableName
	value := t.FieldByName(r)
	if value.CanInt() || value.String() != "" {
		result = true
	}
	return result
}
func (eq *EqualEvaluatorStrategy) Calculate(userAttribute *UserAttribute, rule *Rule) bool {
	result := false
	t := reflect.ValueOf(userAttribute)
	r := rule.VariableName
	value := t.FieldByName(r)
	comparisonValue := rule.ValueString
	fmt.Println(comparisonValue)
	if value.CanInt() {
		val, _ := strconv.Atoi(comparisonValue)
		result = value.Elem().Int() == int64(val)
	} else {
		result = value.Elem().String() == comparisonValue
	}
	return result
}

type GreaterEvaluatorStrategy struct{}

func (eq *GreaterEvaluatorStrategy) Validate(userAttribute *UserAttribute, rule *Rule) bool {
	result := false
	t := reflect.ValueOf(userAttribute)
	r := rule.VariableName
	value := t.FieldByName(r)
	if value.CanInt() {
		result = true
	}
	return result
}
func (eq *GreaterEvaluatorStrategy) Calculate(userAttribute *UserAttribute, rule *Rule) bool {
	result := false
	t := reflect.ValueOf(userAttribute)
	r := rule.VariableName
	value := t.FieldByName(r)
	comparisonValue := rule.ValueString
	fmt.Println(comparisonValue)
	if value.CanInt() {
		val, _ := strconv.Atoi(comparisonValue)
		result = value.Int() > int64(val)
	}
	return result
}

func main() {
	fmt.Println("Starting of the program!!!")

	db := NewDB()
	user1 := db.NewUserAttribute("123", "IN", true, 250000, 180, false, "2.5.0")

	r1 := NewRule("i1", "IsBlocked", EQUAL, "true", true)
	db.NewRuleAddition(r1)
	r2 := NewRule("i2", "AccountAgeInDays", GREATER, "30", false)
	db.NewRuleAddition(r2)

	cr1 := db.NewConnectingRuleAddition("cr1", r1, nil, ANDOPERATION)

	ruleEvaluator := NewRuleEvaluator(db)

	result := ruleEvaluator.Evaluate(user1, cr1)

	fmt.Println("Result: ", result)
}

// NOT(isBlocked = true) AND (accountAgeInDays > 30)
