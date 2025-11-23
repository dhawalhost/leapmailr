package utils

import (
	"testing"
)

func TestValidateRedirectURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL",
			url:     "https://example.com/page",
			wantErr: false,
		},
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com/page",
			wantErr: false,
		},
		{
			name:    "Valid URL with query params",
			url:     "https://example.com/page?param=value&foo=bar",
			wantErr: false,
		},
		{
			name:    "Valid URL with path and fragment",
			url:     "https://example.com/path/to/page#section",
			wantErr: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "Invalid scheme - javascript",
			url:     "javascript:alert('XSS')",
			wantErr: true,
		},
		{
			name:    "Invalid scheme - data",
			url:     "data:text/html,<script>alert('XSS')</script>",
			wantErr: true,
		},
		{
			name:    "Invalid scheme - vbscript",
			url:     "vbscript:msgbox",
			wantErr: true,
		},
		{
			name:    "Invalid scheme - file",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "Invalid scheme - ftp",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "Localhost blocked",
			url:     "http://localhost:8080/admin",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 blocked",
			url:     "http://127.0.0.1/admin",
			wantErr: true,
		},
		{
			name:    "Private network 192.168.x.x blocked",
			url:     "http://192.168.1.1/admin",
			wantErr: true,
		},
		{
			name:    "Private network 10.x.x.x blocked",
			url:     "http://10.0.0.1/admin",
			wantErr: true,
		},
		{
			name:    "Private network 172.16.x.x blocked",
			url:     "http://172.16.0.1/admin",
			wantErr: true,
		},
		{
			name:    "IPv6 localhost blocked",
			url:     "http://[::1]/admin",
			wantErr: true,
		},
		{
			name:    "URL without host",
			url:     "http://",
			wantErr: true,
		},
		{
			name:    "Relative URL",
			url:     "/admin/users",
			wantErr: true,
		},
		{
			name:    "URL too long",
			url:     "https://example.com/" + string(make([]byte, 3000)),
			wantErr: true,
		},
		{
			name:    "Invalid URL format",
			url:     "not a valid url at all",
			wantErr: true,
		},
		{
			name:    "URL with embedded javascript",
			url:     "https://example.com/page?redirect=javascript:alert(1)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRedirectURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRedirectURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsAllowedDomain(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		allowedDomains []string
		wantAllowed    bool
		wantErr        bool
	}{
		{
			name:           "Exact domain match",
			url:            "https://example.com/page",
			allowedDomains: []string{"example.com"},
			wantAllowed:    true,
			wantErr:        false,
		},
		{
			name:           "Subdomain match",
			url:            "https://sub.example.com/page",
			allowedDomains: []string{"example.com"},
			wantAllowed:    true,
			wantErr:        false,
		},
		{
			name:           "Multiple allowed domains",
			url:            "https://trusted.com/page",
			allowedDomains: []string{"example.com", "trusted.com", "safe.org"},
			wantAllowed:    true,
			wantErr:        false,
		},
		{
			name:           "Domain not in allowed list",
			url:            "https://evil.com/page",
			allowedDomains: []string{"example.com", "trusted.com"},
			wantAllowed:    false,
			wantErr:        false,
		},
		{
			name:           "Empty allowed list - allow all",
			url:            "https://anything.com/page",
			allowedDomains: []string{},
			wantAllowed:    true,
			wantErr:        false,
		},
		{
			name:           "Relative URL (parsed but not a valid full URL)",
			url:            "/relative/path",
			allowedDomains: []string{"example.com"},
			wantAllowed:    false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := IsAllowedDomain(tt.url, tt.allowedDomains)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if allowed != tt.wantAllowed {
				t.Errorf("IsAllowedDomain() = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestSanitizeHostHeader(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid domain",
			host:    "example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "Valid domain with port",
			host:    "example.com:443",
			want:    "example.com:443",
			wantErr: false,
		},
		{
			name:    "Valid subdomain",
			host:    "api.example.com",
			want:    "api.example.com",
			wantErr: false,
		},
		{
			name:    "Empty host",
			host:    "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Localhost blocked",
			host:    "localhost",
			want:    "",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 blocked",
			host:    "127.0.0.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Private IP blocked",
			host:    "192.168.1.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid characters",
			host:    "evil.com<script>",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Host with spaces",
			host:    "evil .com",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeHostHeader(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeHostHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeHostHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeRequestURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid root path",
			uri:     "/",
			want:    "/",
			wantErr: false,
		},
		{
			name:    "Valid path",
			uri:     "/api/users",
			want:    "/api/users",
			wantErr: false,
		},
		{
			name:    "Valid path with query",
			uri:     "/api/users?id=123&name=test",
			want:    "/api/users?id=123&name=test",
			wantErr: false,
		},
		{
			name:    "Valid path with fragment",
			uri:     "/page#section",
			want:    "/page#section",
			wantErr: false,
		},
		{
			name:    "Empty URI defaults to /",
			uri:     "",
			want:    "/",
			wantErr: false,
		},
		{
			name:    "Absolute URL blocked",
			uri:     "http://evil.com/page",
			want:    "",
			wantErr: true,
		},
		{
			name:    "HTTPS URL blocked",
			uri:     "https://evil.com/page",
			want:    "",
			wantErr: true,
		},
		{
			name:    "JavaScript injection blocked",
			uri:     "/page?redirect=javascript:alert(1)",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Data URI blocked",
			uri:     "/page?data:text/html,<script>alert(1)</script>",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Relative path without leading slash",
			uri:     "admin/users",
			want:    "",
			wantErr: true,
		},
		{
			name:    "URI too long",
			uri:     "/" + string(make([]byte, 3000)),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeRequestURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeRequestURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeRequestURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildSecureHTTPSURL(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid domain and path",
			host:    "example.com",
			uri:     "/api/users",
			want:    "https://example.com/api/users",
			wantErr: false,
		},
		{
			name:    "Valid domain with port and path",
			host:    "example.com:8443",
			uri:     "/api/users",
			want:    "https://example.com:8443/api/users",
			wantErr: false,
		},
		{
			name:    "Valid subdomain",
			host:    "api.example.com",
			uri:     "/v1/users?limit=10",
			want:    "https://api.example.com/v1/users?limit=10",
			wantErr: false,
		},
		{
			name:    "Root path",
			host:    "example.com",
			uri:     "/",
			want:    "https://example.com/",
			wantErr: false,
		},
		{
			name:    "Invalid host - localhost",
			host:    "localhost",
			uri:     "/admin",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid host - private IP",
			host:    "192.168.1.1",
			uri:     "/admin",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid URI - absolute URL",
			host:    "example.com",
			uri:     "http://evil.com/phishing",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid host - special characters",
			host:    "evil<script>.com",
			uri:     "/page",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty host",
			host:    "",
			uri:     "/page",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildSecureHTTPSURL(tt.host, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildSecureHTTPSURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildSecureHTTPSURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
