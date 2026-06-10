package core

type HookFunc func()
type Hook struct {
	entries map[string][]HookFunc
}

func NewHook() *Hook {
	return &Hook{
		entries: make(map[string][]HookFunc),
	}
}

func (h *Hook) AddFunc(pos string, fn HookFunc) {
	if _, ok := h.entries[pos]; !ok {
		h.entries[pos] = []HookFunc{}
	}

	h.entries[pos] = append(h.entries[pos], fn)
}

func (h *Hook) RunFunc(pos string) {
	if _, ok := h.entries[pos]; ok {
		for _, fn := range h.entries[pos] {
			fn()
		}
	}
}
