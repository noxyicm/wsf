package utils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
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
func (d DataTree) Push(k string, v interface{}) {
	keys := FetchIndexes(k)
	if len(keys) <= MaxTreeLevel {
		d.Mount(keys, v)
	}
}

// Get retreives a value from tree
func (d DataTree) Get(k string) interface{} {
	keys := FetchIndexes(k)
	if len(keys) <= MaxTreeLevel {
		return d.Unmount(keys)
	}

	return nil
}

// Has returns true if key exists in a tree
func (d DataTree) Has(k string) bool {
	keys := FetchIndexes(k)
	if len(keys) > 0 {
		if len(keys) == 1 {
			if _, ok := d[keys[0]]; ok {
				return true
			} else {
				return false
			}
		}

		if len(keys) == 2 && keys[1] == "" {
			if _, ok := d[keys[0]]; ok {
				return true
			} else {
				return false
			}
		}

		if p, ok := d[keys[0]]; ok {
			return p.(DataTree).Has(strings.Join(keys[1:], ""))
		}
	}

	return false
}

// Mount mounts data tree recursively
func (d DataTree) Mount(i []string, v interface{}) {
	if len(i) == 1 {
		if i[0] == "" {
			d[strconv.Itoa(len(d))] = v
		} else {
			d[i[0]] = v
		}
		return
	}

	if p, ok := d[i[0]]; ok {
		p.(DataTree).Mount(i[1:], v)
		return
	}

	d[i[0]] = make(DataTree)
	d[i[0]].(DataTree).Mount(i[1:], v)
}

// Unmount retrives data from tree recursively
func (d DataTree) Unmount(i []string) interface{} {
	if len(i) == 1 {
		return d[i[0]]
	}

	if p, ok := d[i[0]]; ok {
		return p.(DataTree).Unmount(i[1:])
	}

	return nil
}

func (d DataTree) Keys(k string) []string {
	keys := FetchIndexes(k)
	if len(keys) <= MaxTreeLevel {
		outkeys := make([]string, 0)

		if k == "" {
			for k := range d {
				outkeys = append(outkeys, k)
			}
		} else {
			for k := range d.Unmount(keys).(DataTree) {
				outkeys = append(outkeys, k)
			}
		}

		return outkeys
	}

	return make([]string, 0)
}

// AsMapSS converts all posible values of DataTree into strings and returns map
func (d DataTree) AsMapSS() (map[string]string, bool) {
	m := make(map[string]string)
	for k, v := range d {
		if vs, ok := v.(string); ok {
			m[k] = vs
		}
	}

	return m, len(m) > 0
}

// MapFromDataTree recursievly converts DataTree to map
func MapFromDataTree(d DataTree) map[string]interface{} {
	m := make(map[string]interface{})
	for key := range d {
		switch v := d.Get(key).(type) {
		case DataTree:
			m[key] = MapFromDataTree(v)

		default:
			m[key] = v
		}
	}

	return m
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

// MapIIKeys returns slice that contains keys of the map
func MapIIKeys(m map[int]int) []int {
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
func MapSMerge(a interface{}, b interface{}) map[string]interface{} {
	c := make(map[string]interface{})
	z := make(map[string]interface{})
	x := make(map[string]interface{})

	switch v := a.(type) {
	case map[string]string:
		for key, val := range v {
			z[key] = val
		}

	case map[string]int:
		for key, val := range v {
			z[key] = val
		}

	case []interface{}:
		for key, val := range v {
			z[strconv.Itoa(key)] = val
		}

	case []string:
		for key, val := range v {
			z[strconv.Itoa(key)] = val
		}

	case []int:
		for key, val := range v {
			z[strconv.Itoa(key)] = val
		}

	case map[string]interface{}:
		z = v

	default:
		z[strconv.Itoa(len(z))] = v
	}

	switch v := b.(type) {
	case map[string]string:
		for key, val := range v {
			x[key] = val
		}

	case map[string]int:
		for key, val := range v {
			x[key] = val
		}

	case []interface{}:
		for key, val := range v {
			x[strconv.Itoa(key)] = val
		}

	case []string:
		for key, val := range v {
			x[strconv.Itoa(key)] = val
		}

	case []int:
		for key, val := range v {
			x[strconv.Itoa(key)] = val
		}

	case map[string]interface{}:
		x = v

	default:
		x[strconv.Itoa(len(z))] = v
	}

	for key, value := range z {
		c[key] = value
	}

	for key, value := range x {
		if _, ok := c[key]; ok {
			if v, ok := c[key].(map[string]interface{}); ok {
				c[key] = MapSMerge(v, value)
			} else {
				c[key] = value
			}
		} else {
			c[key] = value
		}
	}

	return c
}

// MapSCopy creates a copy of a map
func MapSCopy(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = MapSCopy(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
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

// I64Key returns key for value in slice of 64bits integers
func I64Key(value int64, sl []int64) (int, bool) {
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

// UniqueISlice filters duplicatest from slice
func UniqueISlice(a []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range a {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

// UniqueSSlice filters duplicatest from slice
func UniqueSSlice(a []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range a {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
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
