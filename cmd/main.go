package main

import "fmt"

type Data map[string]any

type Node interface {
	Evaluate(Data) (bool, error)
}

type AndNode struct {
	Children []Node
}

type OrNode struct {
	Children []Node
}

func (n AndNode) Evaluate(d Data) (bool, error) {
	var ans bool
	for i, childNode := range n.Children {
		r, _ := childNode.Evaluate(d)
		if i == 0 {
			ans = r
		} else {
			ans = ans && r
		}
	}
	return ans, nil
}

func (n OrNode) Evaluate(d Data) (bool, error) {
	var ans bool
	for i, childNode := range n.Children {
		r, _ := childNode.Evaluate(d)
		if i == 0 {
			ans = r
		} else {
			ans = ans || r
		}
	}
	return ans, nil
}

type Operator string

const (
	OPEQ   Operator = "equal"
	OPNEQ  Operator = "notequal"
	OPGR   Operator = "greater"
	OPGREQ Operator = "greaterandequal"
)

type ConditionNode struct {
	key      string
	val      any
	operator Operator
}

func (n ConditionNode) Evaluate(d Data) (bool, error) {
	aval, ok := d[n.key]
	if !ok {
		return false, nil
	}

	switch v := aval.(type) {
	case string:
		return StringOperation(v, n.val, n.operator), nil
	case int:
		targInt, _ := n.val.(int)
		return IntOperation(float64(v), targInt, n.operator), nil
	case float64:
		return IntOperation(v, n.val, n.operator), nil
	case bool:
		return BoolOperation(v, n.val, n.operator), nil
	}

	return false, nil
}

func IntOperation(val float64, comp any, op Operator) bool {
	targ, _ := comp.(float64)
	switch op {
	case OPEQ:
		return val == targ
	case OPNEQ:
		return val != targ
	case OPGR:
		return val > targ
	case OPGREQ:
		return val >= targ
	}
	return false
}

func StringOperation(val string, comp any, op Operator) bool {
	targ, _ := comp.(string)
	switch op {
	case OPEQ:
		return val == targ
	case OPNEQ:
		return val != targ
	}
	return false
}

func BoolOperation(val bool, comp any, op Operator) bool {
	targ, _ := comp.(bool)
	switch op {
	case OPEQ:
		return val == targ
	case OPNEQ:
		return val != targ
	}
	return false
}

func main() {
	fmt.Println("Welcome to program!")

	data := Data{
		"userId":           "123",
		"country":          "IN",
		"kycVerified":      true,
		"monthlyVolume":    250000,
		"accountAgeInDays": 180,
		"isBlocked":        false,
		"appVersion":       "2.5.0",
	}

	rule1 := AndNode{
		Children: []Node{
			ConditionNode{
				key:      "isBlocked",
				val:      false,
				operator: OPEQ,
			},
			ConditionNode{
				key:      "accountAgeInDays",
				val:      30.0,
				operator: OPGR,
			},
		},
	}

	result, _ := rule1.Evaluate(data)

	fmt.Println("Rule 1 result: ", result)
}

// NOT(isBlocked = true) AND (accountAgeInDays > 30)
