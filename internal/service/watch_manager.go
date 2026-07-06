package service

import "sync"

type WatchManager struct {
	mu      sync.Mutex
	waiters map[string][]chan struct{}
}

func NewWatchManager() *WatchManager {
	return &WatchManager{
		waiters: make(map[string][]chan struct{}),
	}
}

func watchKey(app, env string) string {
	return app + ":" + env
}

func (m *WatchManager) Add(app, env string) chan struct{} {
	ch := make(chan struct{}, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	key := watchKey(app, env)
	m.waiters[key] = append(m.waiters[key], ch)

	return ch
}

func (m *WatchManager) Remove(app, env string, target chan struct{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := watchKey(app, env)
	list := m.waiters[key]

	filtered := list[:0]
	for _, ch := range list {
		if ch != target {
			filtered = append(filtered, ch)
		}
	}

	if len(filtered) == 0 {
		delete(m.waiters, key)
		return
	}
	m.waiters[key] = filtered
}

func (m *WatchManager) Notify(app, env string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := watchKey(app, env)
	for _, ch := range m.waiters[key] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
