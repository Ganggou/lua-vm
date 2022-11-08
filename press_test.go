package luavm

import (
	"sync"
	"testing"
	"time"

	"github.com/Ganggou/utils"
	lua "github.com/yuin/gopher-lua"
)

func multiCall(vm *LuaVm, times int) (successCount int) {
	var wg sync.WaitGroup
	successChan := make(chan int)
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			defer wg.Done()
			expectedRes := utils.RandStringRunes(6)
			res, err := vm.CallFunc("TestFunc", 1, lua.LString(expectedRes))
			if err != nil {
				return
			}
			if len(res) == 1 && lua.LVAsString(res[0]) == expectedRes {
				successChan <- 1
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(successChan)
	}()

	for n := range successChan {
		if n > 0 {
			successCount += 1
		}
	}
	return
}

func TestPress(t *testing.T) {
	vm := GetLuaVm()
	source := `
	function TestFunc(str)
		return str
	end
	`
	if err := vm.DoString(source); err != nil {
		t.Fatal("Expected to DoString successfully, got error: ", err)
	}
	tryTimes := 200000
	if successCount := multiCall(vm, tryTimes); successCount != tryTimes {
		t.Fatalf("Expected to get %d successful result, got %d", tryTimes, successCount)
	}
	if vm.GetThreadCount() <= ThreadCapacity {
		t.Fatal("Expected to new threads successfully")
	}
	time.Sleep(time.Second)
	if vm.GetThreadCount() > ThreadCapacity {
		t.Fatal("Expected to recover threads successfully")
	}
}
