package models

import (
	"testing"

	"github.com/koyif/metrics/pkg/pool"
)

func TestPoolWithResetableStruct(t *testing.T) {
	p := pool.New(func() *ResetableStruct {
		return &ResetableStruct{
			s: make([]int, 0, 10),
			m: make(map[string]string),
		}
	})

	obj := p.Get()
	obj.i = 42
	obj.str = "test"
	obj.s = append(obj.s, 1, 2, 3)
	obj.m["key"] = "value"

	p.Put(obj)

	obj2 := p.Get()
	if obj2.i != 0 {
		t.Errorf("Expected i to be 0, got %d", obj2.i)
	}
	if obj2.str != "" {
		t.Errorf("Expected str to be empty, got %s", obj2.str)
	}
	if len(obj2.s) != 0 {
		t.Errorf("Expected slice to be empty, got length %d", len(obj2.s))
	}
	if len(obj2.m) != 0 {
		t.Errorf("Expected map to be empty, got length %d", len(obj2.m))
	}
}

func TestPoolWithComplexStruct(t *testing.T) {
	p := pool.New(func() *ComplexStruct {
		return &ComplexStruct{
			slice:     make([]string, 0, 20),
			intSlice:  make([]int, 0, 20),
			mapData:   make(map[string]int),
			mapString: make(map[int]string),
		}
	})

	obj := p.Get()
	obj.intVal = 100
	obj.stringVal = "hello"
	obj.boolVal = true
	obj.floatVal = 3.14
	obj.slice = append(obj.slice, "a", "b", "c")
	obj.mapData["test"] = 999

	p.Put(obj)

	obj2 := p.Get()
	if obj2.intVal != 0 {
		t.Errorf("Expected intVal to be 0, got %d", obj2.intVal)
	}
	if obj2.stringVal != "" {
		t.Errorf("Expected stringVal to be empty, got %s", obj2.stringVal)
	}
	if obj2.boolVal != false {
		t.Errorf("Expected boolVal to be false, got %v", obj2.boolVal)
	}
	if obj2.floatVal != 0 {
		t.Errorf("Expected floatVal to be 0, got %f", obj2.floatVal)
	}
	if len(obj2.slice) != 0 {
		t.Errorf("Expected slice to be empty, got length %d", len(obj2.slice))
	}
	if len(obj2.mapData) != 0 {
		t.Errorf("Expected mapData to be empty, got length %d", len(obj2.mapData))
	}
}

func TestPoolWithNestedResetableStruct(t *testing.T) {
	p := pool.New(func() *ResetableStruct {
		return &ResetableStruct{
			s: make([]int, 0, 10),
			m: make(map[string]string),
			child: &ResetableStruct{
				s: make([]int, 0, 5),
				m: make(map[string]string),
			},
		}
	})

	obj := p.Get()
	obj.i = 10
	obj.str = "parent"
	obj.child.i = 20
	obj.child.str = "child"
	obj.child.s = append(obj.child.s, 1, 2, 3)

	p.Put(obj)

	obj2 := p.Get()
	if obj2.i != 0 || obj2.str != "" {
		t.Error("Parent fields not reset")
	}
	if obj2.child == nil {
		t.Fatal("Child should not be nil")
	}
	if obj2.child.i != 0 || obj2.child.str != "" {
		t.Error("Child fields not reset")
	}
	if len(obj2.child.s) != 0 {
		t.Error("Child slice not reset")
	}
}

func BenchmarkPoolGetPut(b *testing.B) {
	p := pool.New(func() *ResetableStruct {
		return &ResetableStruct{
			s: make([]int, 0, 100),
			m: make(map[string]string),
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := p.Get()
		obj.i = i
		obj.s = append(obj.s, 1, 2, 3)
		obj.m["key"] = "value"
		p.Put(obj)
	}
}

func BenchmarkWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := &ResetableStruct{
			s: make([]int, 0, 100),
			m: make(map[string]string),
		}
		obj.i = i
		obj.s = append(obj.s, 1, 2, 3)
		obj.m["key"] = "value"
	}
}