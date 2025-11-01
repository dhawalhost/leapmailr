# LeapMailr Feature Roadmap: Path to EmailJS Parity

This document outlines the key features missing in LeapMailr when compared to EmailJS. It serves as a high-level roadmap for future development work.

## High Priority

### 1. CAPTCHA Verification ✅ COMPLETED

*   **Gap**: LeapMailr lacks spam protection for public forms, which is a critical security feature.
*   **Status**: ✅ **IMPLEMENTED**
*   **Implementation**:
    *   **Backend**:
        - ✅ Created `CaptchaConfig` model to store CAPTCHA provider, site key, and secret key
        - ✅ Added database migration for `captcha_configs` table
        - ✅ Implemented CRUD API endpoints (`/api/v1/captcha`)
        - ✅ Created CAPTCHA validation service supporting Google reCAPTCHA v2 and hCaptcha
        - ✅ Integrated CAPTCHA validation into `/send-form` endpoint
    *   **Frontend**:
        - ✅ Created CAPTCHA settings page at `/dashboard/settings/captcha`
        - ✅ Users can add, view, activate/deactivate, and delete CAPTCHA configurations
        - ✅ Display usage instructions for developers

### 2. Suppressions List ✅ COMPLETED

*   **Status**: Fully Implemented
*   **Implementation Details**:
    *   **Backend**:
        1.  ✅ Created `suppressions` table with email, reason (bounce/complaint/unsubscribe/manual), source (webhook/manual/api), and metadata
        2.  ✅ Implemented webhook handlers for SendGrid, Mailgun, and generic providers
        3.  ✅ Integrated suppression checks into SendEmail and SendBulkEmail functions
        4.  ✅ Created API endpoints: POST /suppressions, POST /suppressions/bulk, GET /suppressions, DELETE /suppressions/:id
    *   **Frontend**:
        1.  ✅ Created suppressions management page at /dashboard/settings/suppressions
        2.  ✅ Added single and bulk email suppression features
        3.  ✅ Implemented filtering by reason, source, and search functionality

## Medium Priority

### 3. Auto-Reply Feature ✅ COMPLETED

*   **Status**: Fully Implemented
*   **Implementation Details**:
    *   **Backend**:
        1.  ✅ Created `AutoReplyConfig` model with support for multiple configurations per user
        2.  ✅ Implemented auto-reply service layer with variable replacement ({{variable}} syntax)
        3.  ✅ Added auto-reply triggers for both `/send-form` and `/send-email` endpoints
        4.  ✅ Created API endpoints: POST /autoreplies, GET /autoreplies, GET /autoreplies/:id, PUT /autoreplies/:id, DELETE /autoreplies/:id, POST /autoreplies/:id/test, GET /autoreplies/logs
        5.  ✅ Implemented async auto-reply sending to not block main email flow
        6.  ✅ Added configurable delay, custom from/reply-to addresses, service-specific and global auto-replies
    *   **Frontend**:
        1.  ✅ Created auto-reply management page at /dashboard/settings/auto-reply
        2.  ✅ Added UI for creating, editing, and testing auto-reply configurations
        3.  ✅ Implemented activation/deactivation toggle for auto-replies
        4.  ✅ Added trigger configuration (form submissions vs API calls)


### 4. Enhanced API Key Management ✅ COMPLETED

*   **Status**: Fully Implemented
*   **Implementation Details**:
    *   **Backend**:
        1.  ✅ Created `APIKeyPair` model with public/private key structure (pk_live_* / sk_live_*)
        2.  ✅ Implemented key generation using crypto/rand with base64 encoding
        3.  ✅ Created comprehensive API endpoints: POST /api-keys, GET /api-keys, GET /api-keys/:id, PUT /api-keys/:id, POST /api-keys/:id/revoke, DELETE /api-keys/:id, POST /api-keys/:id/rotate, GET /api-keys/:id/usage
        4.  ✅ Updated AuthMiddleware to accept X-Public-Key and X-Private-Key headers
        5.  ✅ Added usage tracking with `APIKeyUsageLog` model
        6.  ✅ Implemented rate limiting, expiration, and revocation features
        7.  ✅ Added key rotation functionality for security
    *   **Frontend**:
        1.  ✅ Created API keys management page at /dashboard/settings/api-keys
        2.  ✅ Implemented secure key display (private key shown only once)
        3.  ✅ Added usage statistics and rate limit tracking
        4.  ✅ Built activation/deactivation, revocation, and deletion controls
        5.  ✅ Added key rotation with secure warning prompts

## Low Priority

### 5. Contact Management ✅ COMPLETED

*   **Status**: Fully Implemented
*   **Implementation Details**:
    *   **Backend**:
        1.  ✅ Created `Contact` and `ContactList` models with email (unique per user), name, phone, company, source, metadata (JSONB), tags (JSONB), subscription status
        2.  ✅ Implemented contact deduplication - updates submission_count if contact already exists
        3.  ✅ Created comprehensive service layer with CreateContact, GetContact, ListContacts (with search/filters), UpdateContact, DeleteContact, ImportContacts (CSV), GetContactStats, ExportContacts (CSV)
        4.  ✅ Added API endpoints: POST /contacts, GET /contacts, GET /contacts/:id, PUT /contacts/:id, DELETE /contacts/:id, POST /contacts/import, GET /contacts/stats, GET /contacts/export
        5.  ✅ Integrated automatic contact collection into `/send-form` endpoint with `collect_contact` flag (defaults to true)
        6.  ✅ Metadata auto-population from template parameters and tagging with template ID
    *   **Frontend**:
        1.  ✅ Created contacts management page at /dashboard/contacts
        2.  ✅ Implemented contact list with search, filtering by source/subscription/tags
        3.  ✅ Added statistics dashboard showing total, subscribed, unsubscribed, and recent contacts
        4.  ✅ Built CRUD operations with inline editing and deletion
        5.  ✅ Implemented CSV import/export functionality
        6.  ✅ Added contact detail view with metadata, tags, and submission count


