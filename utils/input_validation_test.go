package utils

import (
	"strings"
	"testing"
)

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		limitStr    string
		offsetStr   string
		wantLimit   int
		wantOffset  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "Default values when empty",
			limitStr:   "",
			offsetStr:  "",
			wantLimit:  50,
			wantOffset: 0,
			wantErr:    false,
		},
		{
			name:       "Valid limit and offset",
			limitStr:   "100",
			offsetStr:  "20",
			wantLimit:  100,
			wantOffset: 20,
			wantErr:    false,
		},
		{
			name:        "Limit below minimum",
			limitStr:    "0",
			offsetStr:   "",
			wantErr:     true,
			errContains: "must be at least 1",
		},
		{
			name:        "Limit above maximum",
			limitStr:    "1001",
			offsetStr:   "",
			wantErr:     true,
			errContains: "maximum is 1000",
		},
		{
			name:        "Negative offset",
			limitStr:    "",
			offsetStr:   "-1",
			wantErr:     true,
			errContains: "must be non-negative",
		},
		{
			name:        "Invalid limit - not a number",
			limitStr:    "abc",
			offsetStr:   "",
			wantErr:     true,
			errContains: "must be a number",
		},
		{
			name:        "Invalid offset - not a number",
			limitStr:    "",
			offsetStr:   "xyz",
			wantErr:     true,
			errContains: "must be a number",
		},
		{
			name:       "Edge case - limit = 1",
			limitStr:   "1",
			offsetStr:  "",
			wantLimit:  1,
			wantOffset: 0,
			wantErr:    false,
		},
		{
			name:       "Edge case - limit = 1000",
			limitStr:   "1000",
			offsetStr:  "",
			wantLimit:  1000,
			wantOffset: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset, err := ValidatePaginationParams(tt.limitStr, tt.offsetStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePaginationParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Error message '%s' does not contain '%s'", err.Error(), tt.errContains)
			}

			if !tt.wantErr {
				if limit != tt.wantLimit {
					t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
				}
				if offset != tt.wantOffset {
					t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
				}
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		allowedValues []string
		fieldName     string
		wantErr       bool
	}{
		{
			name:          "Valid value",
			value:         "active",
			allowedValues: []string{"active", "inactive", "pending"},
			fieldName:     "status",
			wantErr:       false,
		},
		{
			name:          "Valid value - case insensitive",
			value:         "ACTIVE",
			allowedValues: []string{"active", "inactive"},
			fieldName:     "status",
			wantErr:       false,
		},
		{
			name:          "Empty value - allowed",
			value:         "",
			allowedValues: []string{"active", "inactive"},
			fieldName:     "status",
			wantErr:       false,
		},
		{
			name:          "Invalid value",
			value:         "unknown",
			allowedValues: []string{"active", "inactive"},
			fieldName:     "status",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum(tt.value, tt.allowedValues, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeSearchQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr bool
	}{
		{
			name:    "Normal query",
			query:   "john doe",
			want:    "john doe",
			wantErr: false,
		},
		{
			name:    "Empty query",
			query:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Query with SQL wildcards",
			query:   "test%_query",
			want:    "test\\%\\_query",
			wantErr: false,
		},
		{
			name:    "Query with backslashes",
			query:   "test\\query",
			want:    "test\\\\query",
			wantErr: false,
		},
		{
			name:    "Query with control characters",
			query:   "test\x00\nquery",
			want:    "testquery",
			wantErr: false,
		},
		{
			name:    "Query too long",
			query:   strings.Repeat("a", 256),
			want:    "",
			wantErr: true,
		},
		{
			name:    "Query with whitespace trim",
			query:   "  test  ",
			want:    "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeSearchQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeSearchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		maxLength    int
		allowedChars string
		fieldName    string
		wantErr      bool
	}{
		{
			name:         "Valid alphanumeric",
			value:        "test123",
			maxLength:    50,
			allowedChars: "",
			fieldName:    "field",
			wantErr:      false,
		},
		{
			name:         "Valid with allowed chars",
			value:        "test-123_value",
			maxLength:    50,
			allowedChars: "-_",
			fieldName:    "field",
			wantErr:      false,
		},
		{
			name:         "Empty value",
			value:        "",
			maxLength:    50,
			allowedChars: "",
			fieldName:    "field",
			wantErr:      true,
		},
		{
			name:         "Exceeds max length",
			value:        strings.Repeat("a", 51),
			maxLength:    50,
			allowedChars: "",
			fieldName:    "field",
			wantErr:      true,
		},
		{
			name:         "Invalid characters",
			value:        "test@value",
			maxLength:    50,
			allowedChars: "",
			fieldName:    "field",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlphanumeric(tt.value, tt.maxLength, tt.allowedChars, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAlphanumeric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTrackingID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid tracking ID",
			id:      "abc123xyz",
			wantErr: false,
		},
		{
			name:    "Valid with hyphens and underscores",
			id:      "tracking-id_123",
			wantErr: false,
		},
		{
			name:    "Empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "Too long",
			id:      strings.Repeat("a", 256),
			wantErr: true,
		},
		{
			name:    "Invalid characters",
			id:      "tracking@id",
			wantErr: true,
		},
		{
			name:    "Spaces not allowed",
			id:      "tracking id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrackingID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTrackingID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "Valid provider - gmail",
			provider: "gmail",
			wantErr:  false,
		},
		{
			name:     "Valid provider - case insensitive",
			provider: "Gmail",
			wantErr:  false,
		},
		{
			name:     "Valid provider - sendgrid",
			provider: "sendgrid",
			wantErr:  false,
		},
		{
			name:     "Empty provider",
			provider: "",
			wantErr:  false,
		},
		{
			name:     "Invalid provider",
			provider: "unknown-provider",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProvider(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{
			name:    "Valid status - active",
			status:  "active",
			wantErr: false,
		},
		{
			name:    "Valid status - case insensitive",
			status:  "DELIVERED",
			wantErr: false,
		},
		{
			name:    "Empty status",
			status:  "",
			wantErr: false,
		},
		{
			name:    "Invalid status",
			status:  "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTagsList(t *testing.T) {
	tests := []struct {
		name    string
		tagsStr string
		want    int // Expected number of tags
		wantErr bool
	}{
		{
			name:    "Valid tags",
			tagsStr: "tag1,tag2,tag3",
			want:    3,
			wantErr: false,
		},
		{
			name:    "Tags with whitespace",
			tagsStr: " tag1 , tag2 , tag3 ",
			want:    3,
			wantErr: false,
		},
		{
			name:    "Duplicate tags removed",
			tagsStr: "tag1,tag2,tag1,TAG1",
			want:    2,
			wantErr: false,
		},
		{
			name:    "Empty string",
			tagsStr: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "Tags with hyphens and underscores",
			tagsStr: "tag-1,tag_2",
			want:    2,
			wantErr: false,
		},
		{
			name:    "Too many tags",
			tagsStr: strings.Repeat("tag,", 51),
			want:    0,
			wantErr: true,
		},
		{
			name:    "Tag too long",
			tagsStr: strings.Repeat("a", 51),
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid characters in tag",
			tagsStr: "tag@1,tag2",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateTagsList(tt.tagsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("ValidateTagsList() returned %d tags, want %d", len(got), tt.want)
			}
		})
	}
}

func TestValidateBooleanParam(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		want      bool
		wantErr   bool
	}{
		{
			name:      "True - lowercase",
			value:     "true",
			fieldName: "field",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "True - uppercase",
			value:     "TRUE",
			fieldName: "field",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "True - numeric",
			value:     "1",
			fieldName: "field",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "True - yes",
			value:     "yes",
			fieldName: "field",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "False - lowercase",
			value:     "false",
			fieldName: "field",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "False - numeric",
			value:     "0",
			fieldName: "field",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "Empty string",
			value:     "",
			fieldName: "field",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "Invalid value",
			value:     "maybe",
			fieldName: "field",
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateBooleanParam(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBooleanParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateBooleanParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		minLength int
		maxLength int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "Valid length",
			value:     "hello",
			minLength: 1,
			maxLength: 10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "Below minimum",
			value:     "",
			minLength: 1,
			maxLength: 10,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "Above maximum",
			value:     "hello world",
			minLength: 1,
			maxLength: 5,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "No minimum constraint",
			value:     "",
			minLength: 0,
			maxLength: 10,
			fieldName: "field",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.value, tt.minLength, tt.maxLength, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid filename",
			input:   "document.pdf",
			want:    "document.pdf",
			wantErr: false,
		},
		{
			name:    "Remove path separators",
			input:   "../etc/passwd",
			want:    "etcpasswd",
			wantErr: false,
		},
		{
			name:    "Remove backslashes",
			input:   "..\\windows\\system32",
			want:    "windowssystem32",
			wantErr: false,
		},
		{
			name:    "Remove control characters",
			input:   "file\x00name.txt",
			want:    "filename.txt",
			wantErr: false,
		},
		{
			name:    "Empty filename",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Too long",
			input:   strings.Repeat("a", 256),
			want:    "",
			wantErr: true,
		},
		{
			name:    "Only invalid characters",
			input:   "../../../",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeFilename(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
