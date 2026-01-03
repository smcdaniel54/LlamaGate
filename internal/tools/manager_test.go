package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_NewManager(t *testing.T) {
	manager := NewManager()
	assert.NotNil(t, manager)
	assert.Len(t, manager.GetAllTools(), 0)
}

