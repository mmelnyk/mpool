package mpool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBasicLimitedPool_NoQueue(t *testing.T) {
	pool := &limitedPool[*MyType]{}

	pool.new = func() *MyType {
		return &MyType{}
	}

	v, _ := pool.Get()
	if v != nil {
		t.Error("Expected nil")
		t.FailNow()
	}

	pool.Put(&MyType{Value: 5})
	v, _ = pool.Get()
	if v != nil {
		t.Error("Expected nil")
		t.FailNow()
	}

	pool.destroy()
}

func TestBasicLimitedPool_OneItemQueue(t *testing.T) {
	pool := &limitedPool[*MyType]{}

	pool.new = func() *MyType {
		return &MyType{Value: 1}
	}

	pool.queue = make(chan *MyType, 1)
	pool.max = 1

	if v, _ := pool.Get(); v.Value != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	pool.new = func() *MyType {
		return &MyType{Value: 2}
	}
	pool.Put(&MyType{Value: 1})

	if v, _ := pool.Get(); v.Value != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if pool.current != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}
}

func TestBasicLimitedPool_Callbacks(t *testing.T) {
	pool := &limitedPool[*MyType]{}

	var (
		flagnewcalled     bool
		flagreleasecalled bool
		flagcheckcalled   bool
	)

	pool.new = func() *MyType {
		flagnewcalled = true
		return &MyType{Value: 1}
	}

	pool.check = func(v *MyType) bool {
		flagcheckcalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
		return true
	}

	pool.release = func(v *MyType) {
		flagreleasecalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
	}

	pool.queue = make(chan *MyType, 1)
	pool.max = 1

	if v, _ := pool.Get(); v.Value != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if pool.current != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if !flagnewcalled {
		t.Error("New callback was not called as expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}

	flagnewcalled = false

	pool.new = func() *MyType {
		flagnewcalled = true
		return &MyType{Value: 2}
	}
	pool.Put(&MyType{Value: 1})

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}

	if v, _ := pool.Get(); v.Value != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if !flagcheckcalled {
		t.Error("Check callback was NOT called as expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}

	flagcheckcalled = false
	pool.Put(&MyType{Value: 2}) // Should be kept
	pool.Put(&MyType{Value: 1}) // Should be released

	if pool.current != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if !flagreleasecalled {
		t.Error("Release callback was NOT called as expected")
		t.FailNow()
	}

	flagnewcalled = false
	flagcheckcalled = false
	flagreleasecalled = false

	pool.new = func() *MyType {
		flagnewcalled = true
		return &MyType{Value: 3}
	}

	pool.check = func(v *MyType) bool {
		flagcheckcalled = true
		if v.Value != 2 {
			t.Error("Expected 1")
			t.FailNow()
		}
		return false
	}

	pool.release = func(v *MyType) {
		flagreleasecalled = true
		if v.Value != 2 {
			t.Error("Expected 1")
			t.FailNow()
		}
	}

	pool.current = 1

	if v, _ := pool.Get(); v.Value != 3 {
		t.Error("Expected 3")
		t.FailNow()
	}

	if pool.current != 1 {
		t.Error("Expected 1", pool.current)
		t.FailNow()
	}

	if !flagnewcalled {
		t.Error("New callback was NOT called as expected")
		t.FailNow()
	}

	if !flagcheckcalled {
		t.Error("Check callback was NOT called as expected")
		t.FailNow()
	}

	if !flagreleasecalled {
		t.Error("Release callback was NOT called as expected")
		t.FailNow()
	}
}

func TestBasicLimitedPool_API(t *testing.T) {
	var (
		flagnewcalled     bool
		flagreleasecalled bool
		flagcheckcalled   bool
	)

	fnnew := func() *MyType {
		flagnewcalled = true
		return &MyType{Value: 1}
	}

	fncheck := func(v *MyType) bool {
		flagcheckcalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
		return true
	}

	fnrelease := func(v *MyType) {
		flagreleasecalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
	}

	pool, err := NewLimitedPool(1, 1, fnnew, fnrelease, fncheck)

	if err != nil {
		t.Error("Errror is not expected")
		t.FailNow()
	}

	raw := pool.(*limitedPool[*MyType])

	if !flagnewcalled {
		t.Error("New callback was not called as expected")
		t.FailNow()
	}

	flagnewcalled = false
	if v, _ := pool.Get(); v.Value != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if !flagcheckcalled {
		t.Error("Check callback was NOT called as expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}

	flagcheckcalled = false
	pool.Put(&MyType{Value: 2}) // Should be kept
	pool.Put(&MyType{Value: 1}) // Should be released

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if !flagreleasecalled {
		t.Error("Release callback was NOT called as expected")
		t.FailNow()
	}

	raw.check = func(v *MyType) bool {
		return true
	}
	raw.release = func(v *MyType) {}
	pool = nil
	runtime.GC()
}

func TestBasicLimitedPool_APIOtherType(t *testing.T) {
	var (
		flagnewcalled     bool
		flagreleasecalled bool
		flagcheckcalled   bool
	)

	fnnew := func() int {
		flagnewcalled = true
		return 1
	}

	fncheck := func(v int) bool {
		flagcheckcalled = true
		if v != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
		return true
	}

	fnrelease := func(v int) {
		flagreleasecalled = true
		if v != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
	}

	pool, err := NewLimitedPool(1, 1, fnnew, fnrelease, fncheck)

	raw := pool.(*limitedPool[int])

	if raw.current != 1 {
		t.Error("Expected 1 ")
		t.FailNow()
	}

	if err != nil {
		t.Error("Errror is not expected")
		t.FailNow()
	}

	if !flagnewcalled {
		t.Error("New callback was not called as expected")
		t.FailNow()
	}

	flagnewcalled = false
	if v, _ := pool.Get(); v != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if raw.current != 1 {
		t.Error("Expected 1 ")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if !flagcheckcalled {
		t.Error("Check callback was NOT called as expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}

	flagcheckcalled = false
	pool.Put(1) // Should be kept
	pool.Put(1) // Should be released

	if raw.current != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if !flagreleasecalled {
		t.Error("Release callback was NOT called as expected")
		t.FailNow()
	}

	if v, _ := pool.Get(); v != 1 {
		t.Error("Expected 1")
		t.FailNow()
	}

	var wd sync.WaitGroup
	wd.Add(1)
	go func() {
		if v, _ := pool.Get(); v != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
		wd.Done()
	}()
	time.Sleep(time.Second)
	pool.Put(1) // Should be passed to go routine
	wd.Wait()

	wd.Add(1)
	go func() {
		wd.Done()
		if _, b := pool.Get(); b {
			t.Error("Expected false")
			t.FailNow()
		}
	}()
	wd.Wait()

	raw.check = func(v int) bool {
		return true
	}
	raw.release = func(v int) {}
	pool = nil
	runtime.GC()
}

func TestBasicLimitedPool_APIError(t *testing.T) {
	var (
		flagnewcalled     bool
		flagreleasecalled bool
		flagcheckcalled   bool
	)

	fnnew := func() *MyType {
		flagnewcalled = true
		return &MyType{Value: 1}
	}

	fncheck := func(v *MyType) bool {
		flagcheckcalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
		return true
	}

	fnrelease := func(v *MyType) {
		flagreleasecalled = true
		if v.Value != 1 {
			t.Error("Expected 1")
			t.FailNow()
		}
	}

	_, err := NewLimitedPool(2, 1, fnnew, fnrelease, fncheck)

	if err == nil {
		t.Error("Errror is expected")
		t.FailNow()
	}

	_, err = NewLimitedPool(0, 1, nil, fnrelease, fncheck)

	if err == nil {
		t.Error("Errror is expected")
		t.FailNow()
	}

	if flagnewcalled {
		t.Error("New callback was called as NOT expected")
		t.FailNow()
	}

	if flagcheckcalled {
		t.Error("Check callback was called as NOT expected")
		t.FailNow()
	}

	if flagreleasecalled {
		t.Error("Release callback was called as NOT expected")
		t.FailNow()
	}
}
