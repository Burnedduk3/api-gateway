package entities_test

import (
	"testing"

	"api-gateway/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestAuthPolicy_RequiresAuth(t *testing.T) {
	tests := []struct {
		name     string
		policy   *entities.AuthPolicy
		expected bool
	}{
		{
			name: "requires auth",
			policy: &entities.AuthPolicy{
				Type:    entities.AuthTypeAPIKey,
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "auth disabled",
			policy: &entities.AuthPolicy{
				Type:    entities.AuthTypeAPIKey,
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "no auth type",
			policy: &entities.AuthPolicy{
				Type:    entities.AuthTypeNone,
				Enabled: true,
			},
			expected: false,
		},
		{
			name:     "nil policy",
			policy:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			if tt.policy != nil {
				result = tt.policy.RequiresAuth()
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthPolicy_GetType(t *testing.T) {
	policy := &entities.AuthPolicy{
		Type: entities.AuthTypeAPIKey,
	}

	assert.Equal(t, entities.AuthTypeAPIKey, policy.GetType())
}

func TestAuthPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  *entities.AuthPolicy
		wantErr bool
	}{
		{
			name: "valid API key policy",
			policy: &entities.AuthPolicy{
				Type:    entities.AuthTypeAPIKey,
				Enabled: true,
				Config: map[string]interface{}{
					"valid_keys": []string{"key1"},
				},
			},
			wantErr: false,
		},
		{
			name: "none type always valid",
			policy: &entities.AuthPolicy{
				Type:    entities.AuthTypeNone,
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
