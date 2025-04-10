package telegram

import (
	"sync"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// UserStateManager управляет состояниями пользователей в боте
type UserStateManager struct {
	states map[int64]models.UserState
	mu     sync.RWMutex
}

// NewUserStateManager создает новый менеджер состояний пользователей
func NewUserStateManager() *UserStateManager {
	return &UserStateManager{
		states: make(map[int64]models.UserState),
	}
}

// SetState устанавливает состояние для пользователя
func (m *UserStateManager) SetState(userID int64, state string, contextData map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[userID] = models.UserState{
		State:       state,
		ContextData: contextData,
	}
}

// GetState получает текущее состояние пользователя
func (m *UserStateManager) GetState(userID int64) (models.UserState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[userID]
	return state, exists
}

// ClearState удаляет состояние пользователя
func (m *UserStateManager) ClearState(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, userID)
}

// UpdateContextData обновляет данные контекста пользователя, сохраняя текущее состояние
func (m *UserStateManager) UpdateContextData(userID int64, newData map[string]interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[userID]
	if !exists {
		return false
	}

	// Обновляем существующие данные новыми
	for key, value := range newData {
		state.ContextData[key] = value
	}

	m.states[userID] = state
	return true
}

// GetContextValue получает значение определенного ключа из контекста состояния пользователя
func (m *UserStateManager) GetContextValue(userID int64, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return nil, false
	}

	value, exists := state.ContextData[key]
	return value, exists
}

// SetContextValue устанавливает значение определенного ключа в контексте состояния пользователя
func (m *UserStateManager) SetContextValue(userID int64, key string, value interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[userID]
	if !exists {
		return false
	}

	state.ContextData[key] = value
	m.states[userID] = state
	return true
}

// HasState проверяет, имеет ли пользователь активное состояние
func (m *UserStateManager) HasState(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.states[userID]
	return exists
}

// HasStateWithName проверяет, имеет ли пользователь определенное состояние
func (m *UserStateManager) HasStateWithName(userID int64, stateName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[userID]
	return exists && state.State == stateName
}
