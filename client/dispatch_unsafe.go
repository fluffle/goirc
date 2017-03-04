package client

import (
	"sync/atomic"
	"unsafe"
)

type hNodePtr struct {
	ptr unsafe.Pointer
}

func (p *hNodePtr) load() *hNode {
	return (*hNode)(atomic.LoadPointer(&p.ptr))
}
func (p *hNodePtr) store(new *hNode) {
	atomic.StorePointer(&p.ptr, unsafe.Pointer(new))
}
func (p *hNodePtr) swap(new *hNode) (old *hNode) {
	return (*hNode)(atomic.SwapPointer(&p.ptr, unsafe.Pointer(new)))
}
func (p *hNodePtr) compareAndSwap(old, new *hNode) (swapped bool) {
	return atomic.CompareAndSwapPointer(&p.ptr, unsafe.Pointer(old), unsafe.Pointer(new))
}

type hNodeState uintptr

func (s *hNodeState) load() hNodeState {
	return hNodeState(atomic.LoadUintptr((*uintptr)(s)))
}
func (s *hNodeState) store(new hNodeState) {
	atomic.StoreUintptr((*uintptr)(s), uintptr(new))
}
func (s *hNodeState) swap(new hNodeState) (old hNodeState) {
	return hNodeState(atomic.SwapUintptr((*uintptr)(s), uintptr(new)))
}
func (s *hNodeState) compareAndSwap(old, new hNodeState) (swapped bool) {
	return atomic.CompareAndSwapUintptr((*uintptr)(s), uintptr(old), uintptr(new))
}
