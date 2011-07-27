package registry

import (
	"container/list"
	"strings"
	"sync/atomic"
	"sync"
)

type HandlerID uint32
type Handler struct {
	Run func(...interface{})
	ID	HandlerID
}

var hidCounter uint32 = 0

func NewHandlerID() HandlerID {
	return HandlerID(atomic.AddUint32(&hidCounter, 1))
}

func NewHandler(h func(...interface{})) *Handler {
	return &Handler{h, NewHandlerID()}
}

type EventRegistry interface {
	AddHandler(name string, h *Handler)
	DelHandler(h *Handler)
	Dispatch(name string, ev ...interface{})
	ClearEvents(name string)
}

type registry struct {
	// Event registry as a lockable map of linked-lists
	sync.RWMutex
	events	   map[string]*list.List
	Dispatcher func(r *registry, name string, ev ...interface{})
}	

func NewRegistry() *registry {
	r := &registry{events: make(map[string]*list.List)}
	r.Parallel()
	return r
}

func (r *registry) AddHandler(name string, h *Handler) {
	name = strings.ToLower(name)
	r.Lock()
	defer r.Unlock()
	if _, ok := r.events[name]; !ok {
		r.events[name] = list.New()
	}
	for e := r.events[name].Front(); e != nil; e = e.Next() {
		// Check we're not adding a duplicate handler to this event
		if e.Value.(*Handler).ID == h.ID {
			return
		}
	}
	r.events[name].PushBack(h)
}

func (r *registry) DelHandler(h *Handler) {
	// This is a bit brute-force. Don't add too many handlers!
	r.Lock()
	defer r.Unlock()
	for _, l := range r.events {
		for e := l.Front(); e != nil; e = e.Next() {
			if e.Value.(*Handler).ID == h.ID {
				l.Remove(e)
			}
		}
	}
}

func (r *registry) Dispatch(name string, ev ...interface{}) {
	r.Dispatcher(r, name, ev)
}

func (r *registry) Parallel() {
	r.Dispatcher = (*registry).parallelDispatch
}

func (r *registry) Serial() {
	r.Dispatcher = (*registry).serialDispatch
}

func (r *registry) parallelDispatch(name string, ev ...interface{}) {
	name = strings.ToLower(name)
	r.RLock()
	defer r.RUnlock()
	if l, ok := r.events[name]; ok {
		for e := l.Front(); e != nil; e = e.Next() {
			h := e.Value.(*Handler)
			go h.Run(ev)
		}
	}
}

func (r *registry) serialDispatch(name string, ev ...interface{}) {
	name = strings.ToLower(name)
	r.RLock()
	defer r.RUnlock()
	if l, ok := r.events[name]; ok {
		hlist := make([]*Handler, 0, l.Len())
		for e := l.Front(); e != nil; e = e.Next() {
			hlist = append(hlist, e.Value.(*Handler))
		}
		go func() {
			for _, h := range hlist {
				h.Run(ev)
			}
		}()
	}
}
