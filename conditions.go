package gorev

import (
	"fmt"
	"strings"
)

type Comparison string

type Condition struct {
	And   []Condition
	Or    []Condition
	Xor   []Condition
	Key   string
	Value interface{}
	ValueNotEqual bool
}

func (c *Condition) Describe() (n string){
	if c.Key != "" {
		n += c.Key
		if c.Value != nil {
			n += fmt.Sprintf("='%v'", c.Value)
		}
		return
	}
	n += " ("
	first := true
	for _, cond := range c.And {
		if !first {
			n += " &&"
		} else {
			first = false
		}
		n += " " + cond.Describe()
	}
	for _, cond := range c.Xor {
		if !first {
			n += " ⊕"
		} else {
			first = false
		}
		n += " " + cond.Describe()
	}
	for _, cond := range c.Or {
		if !first {
			n += " ||"
		} else {
			first = false
		}
		n += " " + cond.Describe()
	}

	n += " )"
	return
}


func ValidateParamConditions(params map[string]interface{}, condition Condition) error {
	if condition.Key != "" {
		v, ok := params[condition.Key]
		if !condition.ValueNotEqual {
			if !ok || v == nil {
				return fmt.Errorf("missing param: %s", condition.Key)
			}
			if condition.Value != nil && condition.Value != v {
				return fmt.Errorf("param '%s' did not match expected '%v' (%v)", condition.Key, condition.Value, v)
			}
			switch t := v.(type) {
			case string:
				if t == "" {
					return fmt.Errorf("empty/missing param '%s'", condition.Key)
				}
			}
		} else {
			if condition.Value == nil && ok {
				return fmt.Errorf("parameter '%s' was specified when not expected", condition.Key)
			} else if condition.Value != nil && ok {
				if condition.Value == v {
					return fmt.Errorf("parameter '%s' value '%v' matched '%v' when inequality expected", condition.Key, v, condition.Value)
				}
			}
		}

	}

	// And
	for _, c := range condition.And {
		if err := ValidateParamConditions(params, c); err != nil {
			return err
		}
	}


	// Xor
	var found []string
	var missing []string
	var errs []string
	for _, c := range condition.Xor {
		if err := ValidateParamConditions(params, c); err != nil {
			errs = append(errs, err.Error())
			missing = append(missing, c.Describe())
		} else {
			found = append(found, c.Describe())
		}
		if len(found) > 1 {
			return fmt.Errorf("2+ XOR param(s): %s", strings.Join(found, ", "))
		}
	}
	if len(condition.Xor) > 0 && len(found) < 1 {
		return fmt.Errorf("invalid/missing XOR param(s): %s\n\tREASON:\n\t%s\n", strings.Join(missing, ", "), strings.Join(errs, "\n\t"))
	}

	// Or
	missing = []string{}
	for _, c := range condition.Or {
		if err := ValidateParamConditions(params, c); err == nil {
			return nil
		} else {
			errs = append(errs, err.Error())
		}
		missing = append(missing, c.Describe())
	}
	if len(condition.Or) > 1 {
		return fmt.Errorf("invalid/missing OR param(s): %s\n\tREASON:\n\t%s\n", strings.Join(missing, ", "), strings.Join(errs, "\n\t"))
	}
	return nil
}
