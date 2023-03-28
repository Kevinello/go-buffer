// Code generated by "enumer -transform snake_upper -type Method -output method_enum.go"; DO NOT EDIT.

package compress

import (
	"fmt"
	"strings"
)

const (
	_MethodName_0      = "NONE"
	_MethodLowerName_0 = "none"
	_MethodName_1      = "LZ4"
	_MethodLowerName_1 = "lz4"
	_MethodName_2      = "ZSTD"
	_MethodLowerName_2 = "zstd"
)

var (
	_MethodIndex_0 = [...]uint8{0, 4}
	_MethodIndex_1 = [...]uint8{0, 3}
	_MethodIndex_2 = [...]uint8{0, 4}
)

func (i Method) String() string {
	switch {
	case i == 2:
		return _MethodName_0
	case i == 130:
		return _MethodName_1
	case i == 144:
		return _MethodName_2
	default:
		return fmt.Sprintf("Method(%d)", i)
	}
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _MethodNoOp() {
	var x [1]struct{}
	_ = x[None-(2)]
	_ = x[LZ4-(130)]
	_ = x[ZSTD-(144)]
}

var _MethodValues = []Method{None, LZ4, ZSTD}

var _MethodNameToValueMap = map[string]Method{
	_MethodName_0[0:4]:      None,
	_MethodLowerName_0[0:4]: None,
	_MethodName_1[0:3]:      LZ4,
	_MethodLowerName_1[0:3]: LZ4,
	_MethodName_2[0:4]:      ZSTD,
	_MethodLowerName_2[0:4]: ZSTD,
}

var _MethodNames = []string{
	_MethodName_0[0:4],
	_MethodName_1[0:3],
	_MethodName_2[0:4],
}

// MethodString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func MethodString(s string) (Method, error) {
	if val, ok := _MethodNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _MethodNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Method values", s)
}

// MethodValues returns all values of the enum
func MethodValues() []Method {
	return _MethodValues
}

// MethodStrings returns a slice of all String values of the enum
func MethodStrings() []string {
	strs := make([]string, len(_MethodNames))
	copy(strs, _MethodNames)
	return strs
}

// IsAMethod returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Method) IsAMethod() bool {
	for _, v := range _MethodValues {
		if i == v {
			return true
		}
	}
	return false
}