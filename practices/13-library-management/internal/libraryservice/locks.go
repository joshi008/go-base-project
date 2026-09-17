package libraryservice

import (
	"sync"

	"github.com/xtgo/uuid"
)

type lockManager struct {
	mu        sync.Mutex
	userLocks map[uuid.UUID]*sync.Mutex
	bookLocks map[uuid.UUID]*sync.Mutex
}

func newLockManager() *lockManager {
	return &lockManager{
		userLocks: make(map[uuid.UUID]*sync.Mutex),
		bookLocks: make(map[uuid.UUID]*sync.Mutex),
	}
}

func (m *lockManager) user(id uuid.UUID) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.userLocks[id] == nil {
		m.userLocks[id] = &sync.Mutex{}
	}
	return m.userLocks[id]
}

func (m *lockManager) book(id uuid.UUID) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bookLocks[id] == nil {
		m.bookLocks[id] = &sync.Mutex{}
	}
	return m.bookLocks[id]
}

func (m *lockManager) lock(userID, bookID uuid.UUID) func() {
	userLock := m.user(userID)
	bookLock := m.book(bookID)
	userLock.Lock()
	bookLock.Lock()
	return func() {
		bookLock.Unlock()
		userLock.Unlock()
	}
}
