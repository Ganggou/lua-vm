package luavm

import (
	"testing"

	"github.com/Ganggou/utils"
	lua "github.com/yuin/gopher-lua"
)

func TestDoString(t *testing.T) {
	vm := GetLuaVm()
	funcName := utils.RandStringRunes(6)
	expectedRes := utils.RandStringRunes(6)
	source := `
	function ` + funcName + `(str)
		return str
	end
	`
	if err := vm.DoString(source); err != nil {
		t.Fatal("Expected to DoString successfully, got error: ", err)
	}
	if res, err := vm.CallFunc(funcName, 1, lua.LString(expectedRes)); err != nil {
		t.Fatal("Expected to CallFunc successfully, got error: ", err)
	} else {
		if len(res) != 1 {
			t.Fatal("Expected to get one result, got ", len(res))
		}
		if actualRes := lua.LVAsString(res[0]); actualRes != expectedRes {
			t.Fatalf("Expected to get result %s, got %s", expectedRes, actualRes)
		}
	}
}

func TestDoFile(t *testing.T) {
	vm := GetLuaVm()
	expectedRes := utils.RandStringRunes(6)
	if err := vm.DoFile("./test_file.lua"); err != nil {
		t.Fatal("Expected to DoFile successfully, got error: ", err)
	}
	if res, err := vm.CallFunc("TestFunc", 1, lua.LString(expectedRes)); err != nil {
		t.Fatal("Expected to CallFunc successfully, got error: ", err)
	} else {
		if len(res) != 1 {
			t.Fatal("Expected to get one result, got ", len(res))
		}
		if actualRes := lua.LVAsString(res[0]); actualRes != expectedRes {
			t.Fatalf("Expected to get result %s, got %s", expectedRes, actualRes)
		}
	}
}

func TestRegisterFunc(t *testing.T) {
	vm := GetLuaVm()
	expectedRes := utils.RandStringRunes(6)
	testFunc := func(L *lua.LState) int {
		param := L.ToString(-1)
		L.Push(lua.LString(param))
		return 1
	}
	vm.RegisterFunc("TestFunc", testFunc)
	if res, err := vm.CallFunc("TestFunc", 1, lua.LString(expectedRes)); err != nil {
		t.Fatal("Expected to CallFunc successfully, got error: ", err)
	} else {
		if len(res) != 1 {
			t.Fatal("Expected to get one result, got ", len(res))
		}
		if actualRes := lua.LVAsString(res[0]); actualRes != expectedRes {
			t.Fatalf("Expected to get result %s, got %s", expectedRes, actualRes)
		}
	}
}
