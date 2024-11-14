package meander

import (
	"errors"
	"strings"
)

type Cost int8

type CostRange struct {
	From Cost
	To   Cost
}

const (
	_ Cost = iota
	Cost1
	Cost2
	Cost3
	Cost4
	Cost5
)

var costStrings = map[string]Cost{
	"$":     Cost1,
	"$$":    Cost2,
	"$$$":   Cost3,
	"$$$$":  Cost4,
	"$$$$$": Cost5,
}

func ParseCost(s string) Cost {
	return costStrings[s]
}

func ParseCostRange(s string) (CostRange, error) {
	var r CostRange
	sl := strings.Split(s, "...")
	if len(sl) != 2 {
		return r, errors.New("invalid cost range")
	}
	r.From = ParseCost(sl[0])
	r.To = ParseCost(sl[1])
	return r, nil
}

func (c CostRange) String() string {
	return c.From.String() + "..." + c.To.String()
}

func (l Cost) String() string {
	for s, v := range costStrings {
		if l == v {
			return s
		}
	}
	return "invalid"
}
