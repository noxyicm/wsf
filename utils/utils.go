package utils

import (
	"fmt"
	"runtime"
	"strconv"
)

const (
	// MaxTreeLevel defines maximum tree depth
	MaxTreeLevel = 127

	// Yes is a string representation of true
	Yes = "1"

	// No is a string representation of false
	No = "0"
)

// FetchIndexes parses input name and splits it into separate indexes list
func FetchIndexes(s string) []string {
	var (
		pos  int
		ch   string
		keys = make([]string, 1)
	)

	for _, c := range s {
		ch = string(c)
		switch ch {
		case " ":
			continue
		case "[":
			pos = 1
			continue
		case "]":
			if pos == 1 {
				keys = append(keys, "")
			}
			pos = 2
		default:
			if pos == 1 || pos == 2 {
				keys = append(keys, "")
			}

			keys[len(keys)-1] += ch
			pos = 0
		}
	}

	return keys
}

// DataTree represents a data tree
type DataTree map[string]interface{}

// Push pushes value into data tree
func (d DataTree) Push(k string, v []string) {
	keys := FetchIndexes(k)
	if len(keys) <= MaxTreeLevel {
		d.Mount(keys, v)
	}
}

// Mount mounts data tree recursively
func (d DataTree) Mount(i []string, v []string) {
	if len(i) == 1 {
		d[i[0]] = v[0]
		return
	}

	if len(i) == 2 && i[1] == "" {
		d[i[0]] = v
		return
	}

	if p, ok := d[i[0]]; ok {
		p.(DataTree).Mount(i[1:], v)
		return
	}

	d[i[0]] = make(DataTree)
	d[i[0]].(DataTree).Mount(i[1:], v)
}

// Less determines wherethere a is less than b
func Less(a, b interface{}) bool {
	switch a.(type) {
	case int:
		switch b.(type) {
		case int:
			return a.(int) < b.(int)

		case int8:
			return a.(int) < int(b.(int8))

		case int16:
			return a.(int) < int(b.(int16))

		case int32:
			return a.(int) < int(b.(int32))

		case int64:
			return a.(int) < int(b.(int64))

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return a.(int) < bi
		}

	case int8:
		switch b.(type) {
		case int:
			return int(a.(int8)) < b.(int)

		case int8:
			return a.(int8) < b.(int8)

		case int16:
			return int(a.(int8)) < int(b.(int16))

		case int32:
			return int(a.(int8)) < int(b.(int32))

		case int64:
			return int(a.(int8)) < int(b.(int64))

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return int(a.(int8)) < bi
		}

	case int16:
		switch b.(type) {
		case int:
			return int(a.(int16)) < b.(int)

		case int8:
			return int(a.(int16)) < int(b.(int8))

		case int16:
			return a.(int16) < b.(int16)

		case int32:
			return int(a.(int16)) < int(b.(int32))

		case int64:
			return int(a.(int16)) < int(b.(int64))

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return int(a.(int16)) < bi
		}

	case int32:
		switch b.(type) {
		case int:
			return int(a.(int32)) < b.(int)

		case int8:
			return int(a.(int32)) < int(b.(int8))

		case int16:
			return int(a.(int32)) < int(b.(int16))

		case int32:
			return a.(int32) < b.(int32)

		case int64:
			return int(a.(int32)) < int(b.(int64))

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return int(a.(int32)) < bi
		}

	case int64:
		ai := int(a.(int64))
		switch b.(type) {
		case int:
			return ai < b.(int)

		case int8:
			return ai < int(b.(int8))

		case int16:
			return ai < int(b.(int16))

		case int32:
			return ai < int(b.(int32))

		case int64:
			return a.(int64) < b.(int64)

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return int(a.(int64)) < bi
		}

	case string:
		ai, _ := strconv.Atoi(a.(string))
		switch b.(type) {
		case int:
			return ai < b.(int)

		case int8:
			return ai < int(b.(int8))

		case int16:
			return ai < int(b.(int16))

		case int32:
			return ai < int(b.(int32))

		case int64:

			return ai < int(b.(int64))

		case string:
			bi, _ := strconv.Atoi(b.(string))
			return ai < bi
		}
	}

	return false
}

// ReverseSlice reverses the order of slice of integers
func ReverseSlice(slice []int) []int {
	s := make([]int, len(slice))
	copy(s, slice)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

// ReverseSliceS reverses the order of slice of strings
func ReverseSliceS(slice []string) []string {
	s := make([]string, len(slice))
	copy(s, slice)
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

// ReverseMapIS reverses the order of map of strings
func ReverseMapIS(m map[int]string) map[int]string {
	s := make(map[int]string)
	keys := MapISKeys(m)
	keys = ReverseSlice(keys)
	for _, k := range keys {
		s[k] = m[k]
	}

	return s
}

// ReverseMapSS reverses the order of map of strings
func ReverseMapSS(m map[string]string) map[string]string {
	s := make(map[string]string)
	keys := MapSSKeys(m)
	keys = ReverseSliceS(keys)
	for _, k := range keys {
		s[k] = m[k]
	}

	return s
}

// MapSKeys returns slice that contains keys of the map
func MapSKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// MapISKeys returns slice that contains keys of the map
func MapISKeys(m map[int]string) []int {
	keys := make([]int, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// MapSSKeys returns slice that contains keys of the map
func MapSSKeys(m map[string]string) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// MapSMerge merges two maps with string keys
func MapSMerge(c interface{}, b interface{}) map[string]interface{} {
	a := make(map[string]interface{})

	bV, bOk := b.(map[string]interface{})
	cV, cOk := c.(map[string]interface{})
	if cOk && bOk {
		for key, value := range cV {
			a[key] = value
		}

		for key, value := range bV {
			if _, ok := a[key]; ok {
				if v, ok := a[key].(map[string]interface{}); ok {
					a[key] = MapSMerge(v, value)
				} else {
					a[key] = value
				}
			} else {
				a[key] = value
			}
		}
	} else if cOk {
		for key, value := range cV {
			a[key] = value
		}

		a[strconv.Itoa(len(a))] = b
	} else if bOk {
		a[strconv.Itoa(len(a))] = c

		for key, value := range bV {
			a[key] = value
		}
	}

	return a
}

// MapSSMerge merges two maps with string keys and string values
func MapSSMerge(c interface{}, b interface{}) map[string]string {
	a := make(map[string]string)

	bV, bOk := b.(map[string]string)
	cV, cOk := c.(map[string]string)
	if cOk && bOk {
		for key, value := range cV {
			a[key] = value
		}

		for key, value := range bV {
			a[key] = value
		}
	} else if cOk {
		for key, value := range cV {
			a[key] = value
		}

		switch b.(type) {
		case map[string]interface{}:
			for bKey, bValue := range b.(map[string]interface{}) {
				switch bValue.(type) {
				case string:
					a[bKey] = bValue.(string)

				case int, int8, int16, int32, int64:
					a[bKey] = fmt.Sprintf("%s", bValue)

				case float32, float64:
					a[bKey] = fmt.Sprintf("%s", bValue)
				}
			}

		case map[int]interface{}:
			for bKey, bValue := range b.(map[int]interface{}) {
				switch bValue.(type) {
				case string:
					a[strconv.Itoa(bKey)] = bValue.(string)

				case int, int8, int16, int32, int64:
					a[strconv.Itoa(bKey)] = fmt.Sprintf("%s", bValue)

				case float32, float64:
					a[strconv.Itoa(bKey)] = fmt.Sprintf("%s", bValue)
				}
			}

		case map[string]int:
			for bKey, bValue := range b.(map[string]int) {
				a[bKey] = strconv.Itoa(bValue)
			}

		case map[int]int:
			for bKey, bValue := range b.(map[int]int) {
				a[strconv.Itoa(bKey)] = strconv.Itoa(bValue)
			}

		case string:
			a["0"] = b.(string)

		case int, int8, int16, int32, int64:
			a["0"] = fmt.Sprintf("%s", b)

		case float32, float64:
			a["0"] = fmt.Sprintf("%s", b)
		}
	} else if bOk {
		a = bV
	}

	return a
}

// InSSlice returns true if slice contains provided string
func InSSlice(s string, sl []string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}

	return false
}

// InISlice returns true if slice contains provided int
func InISlice(s int, sl []int) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}

	return false
}

// InI64Slice returns true if slice contains provided int64
func InI64Slice(s int64, sl []int64) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}

	return false
}

// ShiftISlice pops first element of slice and returns it
func ShiftISlice(sl *[]int) (int, bool) {
	s := *sl
	if len(s) > 0 {
		rt := s[0]
		*sl = s[1:]
		return rt, true
	}

	return 0, false
}

// ShiftSSlice pops first element of slice and returns it
func ShiftSSlice(sl *[]string) (string, bool) {
	s := *sl
	if len(s) > 0 {
		rt := s[0]
		*sl = s[1:]
		return rt, true
	}

	return "", false
}

// InSMap returns true if map contains provided value
func InSMap(s interface{}, m map[string]interface{}) bool {
	for _, v := range m {
		if v == s {
			return true
		}
	}

	return false
}

// MapSKeyExists returns true if map contains provided string key
func MapSKeyExists(key string, m map[string]interface{}) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// MapSSKeyExists returns true if map contains provided string key
func MapSSKeyExists(key string, m map[string]string) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// MapSBKeyExists returns true if map contains provided string key
func MapSBKeyExists(key string, m map[string]bool) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// MapISKeyExists returns true if map contains provided int key
func MapISKeyExists(key int, m map[int]string) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

// IKey returns key for value in slice of integers
func IKey(value int, sl []int) (int, bool) {
	for k, v := range sl {
		if v == value {
			return k, true
		}
	}

	return 0, false
}

// SKey returns key for value in slice of strings
func SKey(value string, sl []string) (int, bool) {
	for k, v := range sl {
		if v == value {
			return k, true
		}
	}

	return 0, false
}

// EqualISlice tells whether a and b contain the same elements
func EqualISlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// EqualSSlice tells whether a and b contain the same elements
func EqualSSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// EqualBSlice tells whether a and b contain the same elements
func EqualBSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// IntersectSSlice computes slice intersections
func IntersectSSlice(a, b []string) ([]string, bool) {
	s := []string{}
	has := false
	for _, aV := range a {
		for _, bV := range b {
			if aV == bV {
				s = append(s, aV)
				has = true
			}
		}
	}

	return s, has
}

// Addslashes quote string with slashes
func Addslashes(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

// DebugBacktrace is
func DebugBacktrace() {
	var pcs [32]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[0:n])
	more := true
	str := "Backtrace:\n\t"
	var frame runtime.Frame
	for more {
		frame, more = frames.Next()
		str = str + fmt.Sprint(frame.Function+" in "+frame.File+" on line "+strconv.Itoa(frame.Line)) + "\n\t"
	}
	fmt.Println(str)
}
