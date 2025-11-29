package models

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}

// generate:reset
type ComplexStruct struct {
	intVal    int
	int64Val  int64
	uintVal   uint
	floatVal  float64
	boolVal   bool
	stringVal string

	intPtr    *int
	stringPtr *string
	boolPtr   *bool

	slice     []string
	intSlice  []int
	mapData   map[string]int
	mapString map[int]string

	nested  NestedStruct
	nestedP *NestedStruct
}

type NestedStruct struct {
	value int
}
