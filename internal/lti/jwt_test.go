package lti

import (
	"testing"
)

func TestLTIClaims_GetContextID(t *testing.T) {
	tests := []struct {
		name     string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "nil context",
			context:  nil,
			expected: "",
		},
		{
			name:     "empty context",
			context:  map[string]interface{}{},
			expected: "",
		},
		{
			name:     "context with id",
			context:  map[string]interface{}{"id": "course-123"},
			expected: "course-123",
		},
		{
			name:     "context with non-string id",
			context:  map[string]interface{}{"id": 123},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &LTIClaims{Context: tt.context}
			got := claims.GetContextID()
			if got != tt.expected {
				t.Errorf("GetContextID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLTIClaims_GetContextLabel(t *testing.T) {
	tests := []struct {
		name     string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "nil context",
			context:  nil,
			expected: "",
		},
		{
			name:     "empty context",
			context:  map[string]interface{}{},
			expected: "",
		},
		{
			name:     "context with label",
			context:  map[string]interface{}{"label": "GEOG-101"},
			expected: "GEOG-101",
		},
		{
			name:     "context with non-string label",
			context:  map[string]interface{}{"label": 101},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &LTIClaims{Context: tt.context}
			got := claims.GetContextLabel()
			if got != tt.expected {
				t.Errorf("GetContextLabel() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLTIClaims_HasRole(t *testing.T) {
	claims := &LTIClaims{
		Roles: []string{
			"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor",
			"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Faculty",
		},
	}

	tests := []struct {
		role     string
		expected bool
	}{
		{"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor", true},
		{"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Faculty", true},
		{"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			got := claims.HasRole(tt.role)
			if got != tt.expected {
				t.Errorf("HasRole(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}
}

func TestLTIClaims_HasRole_EmptyRoles(t *testing.T) {
	claims := &LTIClaims{Roles: nil}
	if claims.HasRole("any-role") {
		t.Error("expected HasRole to return false for nil roles")
	}

	claims.Roles = []string{}
	if claims.HasRole("any-role") {
		t.Error("expected HasRole to return false for empty roles")
	}
}

func TestLTIClaims_IsInstructor(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{
			name:     "membership instructor",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor"},
			expected: true,
		},
		{
			name:     "institution instructor",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Instructor"},
			expected: true,
		},
		{
			name:     "learner only",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner"},
			expected: false,
		},
		{
			name:     "instructor among multiple roles",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner", "http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor"},
			expected: true,
		},
		{
			name:     "empty roles",
			roles:    []string{},
			expected: false,
		},
		{
			name:     "nil roles",
			roles:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &LTIClaims{Roles: tt.roles}
			got := claims.IsInstructor()
			if got != tt.expected {
				t.Errorf("IsInstructor() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLTIClaims_IsLearner(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{
			name:     "membership learner",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner"},
			expected: true,
		},
		{
			name:     "institution student",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Student"},
			expected: true,
		},
		{
			name:     "instructor only",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor"},
			expected: false,
		},
		{
			name:     "learner among multiple roles",
			roles:    []string{"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner", "http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor"},
			expected: true,
		},
		{
			name:     "empty roles",
			roles:    []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &LTIClaims{Roles: tt.roles}
			got := claims.IsLearner()
			if got != tt.expected {
				t.Errorf("IsLearner() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewJWTValidator(t *testing.T) {
	v := NewJWTValidator()
	if v == nil {
		t.Fatal("expected validator to be created")
	}
	if v.jwksCache == nil {
		t.Error("expected jwksCache to be initialized")
	}
}
