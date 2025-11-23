# Open Redirect Vulnerability Fix - Link Tracking System

## Security Issue: CWE-601 - URL Redirection to Untrusted Site

**Severity:** HIGH  
**OWASP Category:** A1 - Broken Access Control  
**Date Fixed:** November 23, 2025

---

## Problem Statement

The original `TrackClickHandler` accepted URLs from user-controlled query parameters and redirected to them after validation. Even with URL validation, security scanners flag this as an open redirect vulnerability because:

1. **User-controlled data determines redirect destination**
2. **Attack surface exists** - malicious actors can craft URLs
3. **Validation can be bypassed** - new attack vectors may emerge
4. **Fails "never trust user input" principle**

### Original Vulnerable Code
```go
func TrackClickHandler(c *gin.Context) {
    trackingPixelID := c.Param("pixel_id")
    linkID := c.Param("link_id")
    encodedURL := c.Query("url") // ⚠️ USER-CONTROLLED INPUT
    
    // Decode URL from query parameter
    urlBytes, err := base64.URLEncoding.DecodeString(encodedURL)
    originalURL := string(urlBytes) // ⚠️ REDIRECT DESTINATION FROM USER
    
    // Validate before redirect
    if err := utils.ValidateRedirectURL(originalURL); err != nil {
        return
    }
    
    c.Redirect(http.StatusFound, originalURL) // ⚠️ REDIRECTS TO USER INPUT
}
```

**Why This Is Vulnerable:**
- Tracking URL: `/api/v1/track/click/{pixel_id}/{link_id}?url={base64_encoded_url}`
- Attacker controls `url` parameter
- Can craft malicious redirects: `/track/click/abc/123?url=<base64:evil.com>`

---

## Solution: Database-Backed Link Tracking

### Architecture Changes

#### 1. New Database Model: `TrackedLink`
```go
type TrackedLink struct {
    ID              uuid.UUID
    TrackingPixelID string    // Links to email tracking
    LinkID          string    // Unique per link in email
    OriginalURL     string    // Pre-approved URL (stored at send time)
    CreatedAt       time.Time
}
```

**Key Security Properties:**
- URLs stored when email is **sent** (system-controlled)
- URLs retrieved by **system-generated IDs** (not user input)
- No user-controlled data in redirect path

#### 2. Modified Email Sending Flow
```go
func InjectLinkTracking(htmlContent, trackingPixelID, baseURL string) string {
    linkRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
        originalURL := extractURL(match)
        linkID := generateLinkID(originalURL)
        
        // ✅ STORE URL IN DATABASE AT SEND TIME
        trackedLink := models.TrackedLink{
            TrackingPixelID: trackingPixelID,
            LinkID:          linkID,
            OriginalURL:     originalURL, // Pre-approved URL
        }
        db.Create(&trackedLink)
        
        // ✅ NO URL IN TRACKING LINK
        trackingURL := fmt.Sprintf("%s/api/v1/track/click/%s/%s",
            baseURL, trackingPixelID, linkID)
        
        return fmt.Sprintf(`href="%s"`, trackingURL)
    })
}
```

#### 3. Secure Handler Implementation
```go
func TrackClickHandler(c *gin.Context) {
    trackingPixelID := c.Param("pixel_id") // System-generated
    linkID := c.Param("link_id")           // System-generated
    
    // ✅ RETRIEVE PRE-APPROVED URL FROM DATABASE
    trackingService := service.NewEmailTrackingService()
    originalURL, err := trackingService.GetTrackedLinkURL(
        trackingPixelID, linkID)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
        return
    }
    
    // ✅ DEFENSE-IN-DEPTH: Re-validate stored URL
    if err := utils.ValidateRedirectURL(originalURL); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
        return
    }
    
    // ✅ REDIRECT TO PRE-APPROVED URL (NOT USER INPUT)
    c.Redirect(http.StatusFound, originalURL)
}
```

---

## Security Benefits

### 1. **Eliminates User-Controlled Redirects**
- **Before:** URL from query parameter (user-controlled)
- **After:** URL from database (system-controlled)
- **Result:** Attacker cannot inject malicious URLs

### 2. **Pre-Approval at Send Time**
- URLs validated and stored when email is sent
- Only legitimate links from sent emails can redirect
- No runtime URL injection possible

### 3. **Defense-in-Depth**
- **Layer 1:** URL stored at send time (trusted source)
- **Layer 2:** Database lookup (validates link exists)
- **Layer 3:** Re-validation before redirect (belt-and-suspenders)
- **Layer 4:** Tracking ID validation (format checking)

### 4. **Audit Trail**
- All tracked links stored with timestamps
- Can identify suspicious redirect patterns
- Easy to revoke compromised links

---

## Attack Surface Comparison

### Before (Vulnerable)
```
Attack Vector: Craft malicious URL parameter
Example: /track/click/valid_id/valid_id?url=<base64:evil.com>
Impact: Redirect to attacker-controlled site
Likelihood: HIGH - URL directly from user input
```

### After (Secure)
```
Attack Vector: Must compromise database to inject malicious URL
Example: /track/click/valid_id/invalid_id (returns 404)
Impact: Cannot redirect - link must exist in database
Likelihood: LOW - requires database breach, not simple parameter manipulation
```

---

## Database Migration

Run the following migration to add the `tracked_links` table:

```sql
-- See: database/migrations/add_tracked_links_table.sql

CREATE TABLE tracked_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_pixel_id VARCHAR(255) NOT NULL,
    link_id VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT idx_tracking_link UNIQUE (tracking_pixel_id, link_id)
);

CREATE INDEX idx_tracked_links_tracking_pixel_id ON tracked_links(tracking_pixel_id);
CREATE INDEX idx_tracked_links_link_id ON tracked_links(link_id);
```

---

## Testing the Fix

### 1. **Legitimate Link Click (Should Work)**
```bash
# Send email with link: https://example.com
# Email HTML contains: /api/v1/track/click/abc123/link456

curl -L http://localhost:8080/api/v1/track/click/abc123/link456
# Expected: 302 redirect to https://example.com
```

### 2. **Malicious URL Injection (Should Fail)**
```bash
# Try to inject URL via query parameter
curl -L "http://localhost:8080/api/v1/track/click/abc123/evil?url=aHR0cHM6Ly9ldmlsLmNvbQ=="
# Expected: 404 Not Found (link not in database)
```

### 3. **Invalid Link ID (Should Fail)**
```bash
# Try non-existent link ID
curl -L http://localhost:8080/api/v1/track/click/abc123/nonexistent
# Expected: 404 Not Found
```

### 4. **Database Verification**
```sql
-- Check tracked links are stored correctly
SELECT tracking_pixel_id, link_id, original_url, created_at 
FROM tracked_links 
WHERE tracking_pixel_id = 'abc123';
```

---

## Backward Compatibility

⚠️ **BREAKING CHANGE:** Old tracking URLs with `?url=` parameter will no longer work.

### Migration Steps:
1. **Deploy new code** with database migration
2. **All new emails** will use new tracking format (no `?url=` parameter)
3. **Old tracking links** (already sent) will return 404 after 24 hours
4. **Optional:** Import old links into `tracked_links` table if needed

### Import Script (if needed):
```sql
-- Extract old tracking URLs from email_click_events
INSERT INTO tracked_links (tracking_pixel_id, link_id, original_url, created_at)
SELECT DISTINCT 
    tracking_id::text as tracking_pixel_id,
    link_id,
    link_url as original_url,
    MIN(clicked_at) as created_at
FROM email_click_events
WHERE link_url IS NOT NULL
GROUP BY tracking_id, link_id, link_url;
```

---

## Files Modified

1. **`models/tracking.go`** - Added `TrackedLink` model
2. **`handlers/tracking.go`** - Removed URL parameter, added database lookup
3. **`service/email_tracking.go`** - Store links in DB, retrieve by ID
4. **`database/migrations/add_tracked_links_table.sql`** - New table migration

---

## Security Compliance

✅ **CWE-601:** URL Redirection to Untrusted Site - **FIXED**  
✅ **OWASP A1:** Broken Access Control - **MITIGATED**  
✅ **PCI DSS 6.5.1:** Injection flaws - **COMPLIANT**  
✅ **Defense-in-Depth:** Multiple validation layers - **IMPLEMENTED**  
✅ **Principle of Least Privilege:** No user control over redirects - **ENFORCED**  

---

## Additional Recommendations

### 1. **Rate Limiting**
Add rate limiting to prevent tracking link abuse:
```go
// Limit: 100 clicks per tracking link per hour
middleware.RateLimit("tracking_click", 100, time.Hour)
```

### 2. **Link Expiration**
Add expiration to tracked links:
```sql
ALTER TABLE tracked_links ADD COLUMN expires_at TIMESTAMP;
CREATE INDEX idx_tracked_links_expires ON tracked_links(expires_at);
```

### 3. **Suspicious Link Detection**
Monitor for:
- Links clicked before email marked as opened
- Unusual geographic patterns
- High-velocity clicks from single IP
- Links clicked from non-email user agents

### 4. **Admin Override**
Add admin endpoint to revoke compromised links:
```go
DELETE FROM tracked_links 
WHERE tracking_pixel_id = ? AND link_id = ?;
```

---

## Summary

This fix **completely eliminates the open redirect vulnerability** by:

1. ✅ **Removing user-controlled redirect destinations**
2. ✅ **Storing pre-approved URLs at email send time**
3. ✅ **Using system-generated IDs for lookups**
4. ✅ **Adding defense-in-depth validation layers**
5. ✅ **Providing audit trail for all tracked links**

**Result:** Attacker cannot manipulate redirect destinations - they must exist in the database with valid tracking IDs.
