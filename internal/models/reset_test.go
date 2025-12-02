package models

import "testing"

func TestResetableStructReset(t *testing.T) {
	str := "test"
	rs := &ResetableStruct{
		i:    42,
		str:  "hello",
		strP: &str,
		s:    []int{1, 2, 3},
		m:    map[string]string{"key": "value"},
		child: &ResetableStruct{
			i:   100,
			str: "child",
		},
	}

	rs.Reset()

	if rs.i != 0 {
		t.Errorf("Expected i to be 0, got %d", rs.i)
	}
	if rs.str != "" {
		t.Errorf("Expected str to be empty, got %s", rs.str)
	}
	if rs.strP == nil {
		t.Error("Expected strP to not be nil")
	} else if *rs.strP != "" {
		t.Errorf("Expected *strP to be empty, got %s", *rs.strP)
	}
	if len(rs.s) != 0 {
		t.Errorf("Expected slice length to be 0, got %d", len(rs.s))
	}
	if cap(rs.s) == 0 {
		t.Error("Expected slice capacity to be preserved")
	}
	if len(rs.m) != 0 {
		t.Errorf("Expected map length to be 0, got %d", len(rs.m))
	}
	if rs.child == nil {
		t.Error("Expected child to not be nil")
	} else {
		if rs.child.i != 0 {
			t.Errorf("Expected child.i to be 0, got %d", rs.child.i)
		}
		if rs.child.str != "" {
			t.Errorf("Expected child.str to be empty, got %s", rs.child.str)
		}
	}
}

func TestComplexStructReset(t *testing.T) {
	intVal := 42
	strVal := "test"
	boolVal := true

	cs := &ComplexStruct{
		intVal:    10,
		int64Val:  20,
		uintVal:   30,
		floatVal:  40.5,
		boolVal:   true,
		stringVal: "hello",
		intPtr:    &intVal,
		stringPtr: &strVal,
		boolPtr:   &boolVal,
		slice:     []string{"a", "b"},
		intSlice:  []int{1, 2, 3},
		mapData:   map[string]int{"key": 1},
		mapString: map[int]string{1: "one"},
		nested:    NestedStruct{value: 99},
		nestedP:   &NestedStruct{value: 88},
	}

	cs.Reset()

	if cs.intVal != 0 {
		t.Errorf("Expected intVal to be 0, got %d", cs.intVal)
	}
	if cs.int64Val != 0 {
		t.Errorf("Expected int64Val to be 0, got %d", cs.int64Val)
	}
	if cs.uintVal != 0 {
		t.Errorf("Expected uintVal to be 0, got %d", cs.uintVal)
	}
	if cs.floatVal != 0 {
		t.Errorf("Expected floatVal to be 0, got %f", cs.floatVal)
	}
	if cs.boolVal != false {
		t.Errorf("Expected boolVal to be false, got %v", cs.boolVal)
	}
	if cs.stringVal != "" {
		t.Errorf("Expected stringVal to be empty, got %s", cs.stringVal)
	}
	if cs.intPtr == nil {
		t.Error("Expected intPtr to not be nil")
	} else if *cs.intPtr != 0 {
		t.Errorf("Expected *intPtr to be 0, got %d", *cs.intPtr)
	}
	if cs.stringPtr == nil {
		t.Error("Expected stringPtr to not be nil")
	} else if *cs.stringPtr != "" {
		t.Errorf("Expected *stringPtr to be empty, got %s", *cs.stringPtr)
	}
	if cs.boolPtr == nil {
		t.Error("Expected boolPtr to not be nil")
	} else if *cs.boolPtr != false {
		t.Errorf("Expected *boolPtr to be false, got %v", *cs.boolPtr)
	}
	if len(cs.slice) != 0 {
		t.Errorf("Expected slice length to be 0, got %d", len(cs.slice))
	}
	if cap(cs.slice) == 0 {
		t.Error("Expected slice capacity to be preserved")
	}
	if len(cs.intSlice) != 0 {
		t.Errorf("Expected intSlice length to be 0, got %d", len(cs.intSlice))
	}
	if len(cs.mapData) != 0 {
		t.Errorf("Expected mapData length to be 0, got %d", len(cs.mapData))
	}
	if len(cs.mapString) != 0 {
		t.Errorf("Expected mapString length to be 0, got %d", len(cs.mapString))
	}
}

func TestResetNilPointer(t *testing.T) {
	var rs *ResetableStruct
	rs.Reset()
}

func TestResetPreservesSliceCapacity(t *testing.T) {
	rs := &ResetableStruct{
		s: make([]int, 5, 10),
	}
	for i := range rs.s {
		rs.s[i] = i + 1
	}

	originalCap := cap(rs.s)

	rs.Reset()

	if len(rs.s) != 0 {
		t.Errorf("Expected slice length to be 0, got %d", len(rs.s))
	}
	if cap(rs.s) != originalCap {
		t.Errorf("Expected slice capacity to be %d, got %d", originalCap, cap(rs.s))
	}
}

func TestResetNestedStructWithReset(t *testing.T) {
	rs := &ResetableStruct{
		i:   10,
		str: "parent",
		child: &ResetableStruct{
			i:   20,
			str: "child",
			s:   []int{1, 2, 3},
			m:   map[string]string{"a": "b"},
		},
	}

	rs.Reset()

	if rs.i != 0 || rs.str != "" {
		t.Error("Parent fields not reset")
	}
	if rs.child == nil {
		t.Fatal("Child pointer should not be nil")
	}
	if rs.child.i != 0 || rs.child.str != "" {
		t.Error("Child fields not reset")
	}
	if len(rs.child.s) != 0 {
		t.Error("Child slice not reset")
	}
	if len(rs.child.m) != 0 {
		t.Error("Child map not reset")
	}
}
