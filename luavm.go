package luavm

import (
	"log"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	ThreadCapacity  = 10
	ThreadInitCount = 1
)

type LuaVm struct {
	state       *lua.LState
	stateMu     sync.RWMutex
	threads     chan *lua.LState
	threadCount int
	threadMu    sync.Mutex
}

// GetLuaVm ...
func GetLuaVm() *LuaVm {
	vm := &LuaVm{}
	vm.state = lua.NewState(lua.Options{
		RegistrySize:    1024 * 20,
		RegistryMaxSize: 1024 * 60,
		CallStackSize:   1024})
	vm.threads = make(chan *lua.LState, ThreadCapacity)
	for i := 0; i < ThreadInitCount; i++ {
		thread, _ := vm.state.NewThread()
		vm.threads <- thread
	}
	vm.threadCount = ThreadInitCount
	return vm
}

// GetGlobal ...
func (vm *LuaVm) GetGlobal(name string) lua.LValue {
	vm.stateMu.RLock()
	defer vm.stateMu.RUnlock()
	return vm.state.GetGlobal(name)
}

// SetGlobal ...
func (vm *LuaVm) SetGlobal(name string, val lua.LValue) {
	vm.stateMu.Lock()
	defer vm.stateMu.Unlock()
	vm.state.SetGlobal(name, val)
}

// NewThread ...
func (vm *LuaVm) NewThread() (thread *lua.LState) {
	thread, _ = vm.state.NewThread()
	vm.threadMu.Lock()
	vm.threadCount++
	vm.threadMu.Unlock()
	return
}

// GetThread ...
func (vm *LuaVm) GetThread() (thread *lua.LState) {
	select {
	case thread = <-vm.threads: // get thread from chan
	case <-time.After(50 * time.Millisecond): // all thread busy, create thread
		thread = vm.NewThread()
	}
	return
}

// PutThread ...
func (vm *LuaVm) PutThread(thread *lua.LState) {
	select {
	case vm.threads <- thread: // put thread to chan
	case <-time.After(time.Second): // chan full of threads, drop the thread
		vm.threadMu.Lock()
		vm.threadCount--
		vm.threadMu.Unlock()
		thread.Close()
	}
}

// DoString ...
func (vm *LuaVm) DoString(str string) error {
	vm.stateMu.Lock()
	defer vm.stateMu.Unlock()
	return vm.state.DoString(str)
}

// DoFile ...
func (vm *LuaVm) DoFile(filepath string) error {
	vm.stateMu.Lock()
	defer vm.stateMu.Unlock()
	return vm.state.DoFile(filepath)
}

// RegisterFunc register function in vm state
func (vm *LuaVm) RegisterFunc(globalFuncName string, fn lua.LGFunction) {
	vm.stateMu.Lock()
	vm.state.SetGlobal(globalFuncName, vm.state.NewFunction(fn))
	vm.stateMu.Unlock()
}

// CallFunc call function set in vm state
func (vm *LuaVm) CallFunc(globalFuncName string, retCnt int, args ...lua.LValue) (res []lua.LValue, err error) {
	thread := vm.GetThread()
	defer func() {
		go vm.PutThread(thread)
		if e := recover(); e != nil {
			log.Println("CallFunc recover: ", e)
		}
	}()

	if err = thread.CallByParam(lua.P{
		Fn:      vm.state.GetGlobal(globalFuncName),
		NRet:    retCnt,
		Protect: true,
	}, args...); err != nil {
		return
	}
	for i := retCnt; i > 0; i-- {
		res = append([]lua.LValue{thread.Get(-i)}, res...)
	}
	return
}
