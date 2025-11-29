package pool

import (
	"testing"
)

type TestStruct struct {
	value  int
	text   string
	items  []int
	data   map[string]string
	nested *TestStruct
}

func (ts *TestStruct) Reset() {
	if ts == nil {
		return
	}
	ts.value = 0
	ts.text = ""
	ts.items = ts.items[:0]
	clear(ts.data)
	if ts.nested != nil {
		ts.nested.Reset()
	}
}

func TestPoolNew(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			items: make([]int, 0, 10),
			data:  make(map[string]string),
		}
	})

	if p == nil {
		t.Fatal("Expected pool to be created")
	}
}

func TestPoolGetReturnsNewObject(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			items: make([]int, 0, 10),
			data:  make(map[string]string),
		}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Expected Get to return an object")
	}
	if obj.items == nil {
		t.Error("Expected items slice to be initialized")
	}
	if obj.data == nil {
		t.Error("Expected data map to be initialized")
	}
}

func TestPoolPutResetsObject(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			items: make([]int, 0, 10),
			data:  make(map[string]string),
		}
	})

	obj := p.Get()
	obj.value = 42
	obj.text = "test"
	obj.items = append(obj.items, 1, 2, 3)
	obj.data["key"] = "value"

	p.Put(obj)

	obj2 := p.Get()
	if obj2.value != 0 {
		t.Errorf("Expected value to be reset to 0, got %d", obj2.value)
	}
	if obj2.text != "" {
		t.Errorf("Expected text to be reset to empty, got %s", obj2.text)
	}
	if len(obj2.items) != 0 {
		t.Errorf("Expected items to be empty, got length %d", len(obj2.items))
	}
	if len(obj2.data) != 0 {
		t.Errorf("Expected data to be empty, got length %d", len(obj2.data))
	}
}

func TestPoolReuseObjects(t *testing.T) {
	callCount := 0
	p := New(func() *TestStruct {
		callCount++
		return &TestStruct{
			items: make([]int, 0, 10),
			data:  make(map[string]string),
		}
	})

	obj1 := p.Get()
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d", callCount)
	}

	obj1.value = 100
	p.Put(obj1)

	obj2 := p.Get()
	if callCount != 1 {
		t.Errorf("Expected factory to not be called again, got %d calls", callCount)
	}

	if obj2.value != 0 {
		t.Error("Expected reused object to be reset")
	}
}

func TestPoolNestedReset(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			items: make([]int, 0, 10),
			data:  make(map[string]string),
			nested: &TestStruct{
				items: make([]int, 0, 5),
				data:  make(map[string]string),
			},
		}
	})

	obj := p.Get()
	obj.value = 10
	obj.nested.value = 20
	obj.nested.text = "nested"
	obj.nested.items = append(obj.nested.items, 1, 2)

	p.Put(obj)

	obj2 := p.Get()
	if obj2.nested == nil {
		t.Fatal("Expected nested to not be nil")
	}
	if obj2.nested.value != 0 {
		t.Errorf("Expected nested value to be reset, got %d", obj2.nested.value)
	}
	if obj2.nested.text != "" {
		t.Errorf("Expected nested text to be reset, got %s", obj2.nested.text)
	}
	if len(obj2.nested.items) != 0 {
		t.Errorf("Expected nested items to be empty, got length %d", len(obj2.nested.items))
	}
}

func TestPoolPreservesCapacity(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			items: make([]int, 0, 100),
			data:  make(map[string]string),
		}
	})

	obj := p.Get()
	originalCap := cap(obj.items)

	for i := 0; i < 50; i++ {
		obj.items = append(obj.items, i)
	}

	p.Put(obj)

	obj2 := p.Get()
	if cap(obj2.items) != originalCap {
		t.Errorf("Expected capacity to be preserved at %d, got %d", originalCap, cap(obj2.items))
	}
	if len(obj2.items) != 0 {
		t.Errorf("Expected length to be 0, got %d", len(obj2.items))
	}
}
