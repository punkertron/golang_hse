//go:build !solution

package testequal

import (
	"fmt"
	"maps"
	"slices"
)

func areEqual(expected, actual any) bool {
	switch v := expected.(type) {
	case []int:
		a, ok := actual.([]int)
		return ok && (v == nil) == (a == nil) && slices.Equal(v, a)
	case []byte:
		a, ok := actual.([]byte)
		return ok && (v == nil) == (a == nil) && slices.Equal(v, a)
	case map[string]string:
		a, ok := actual.(map[string]string)
		return ok && (v == nil) == (a == nil) && maps.Equal(v, a)
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, string:
		return expected == actual
	default:
		return false
	}
}

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !areEqual(expected, actual) {
		message := ""
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}

		t.Errorf("not equal:\nexpected: %v\nactual. : %v\nmessage : %v", expected, actual, message)

		return false
	}

	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if areEqual(expected, actual) {
		message := ""
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}

		t.Errorf("are equal:\nShould not be: %v\nmessage : %v", expected, message)

		return false
	}

	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if !areEqual(expected, actual) {
		message := ""
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}

		t.Errorf("not equal:\nexpected: %v\nactual. : %v\nmessage : %v", expected, actual, message)
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if areEqual(expected, actual) {
		message := ""
		if len(msgAndArgs) > 0 {
			message = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}

		t.Errorf("are equal:\nShould not be: %v\nmessage : %v", expected, message)
		t.FailNow()
	}
}
