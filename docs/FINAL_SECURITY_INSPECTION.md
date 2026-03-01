# ClaraVerse Security Guide

Comprehensive security inspection and documentation for ClaraVerseAI.

## Security Overview

ClaraVerse implements defense-in-depth security with multiple layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Layers                           │
├─────────────────────────────────────────────────────────────┤
│  1. Network Layer     │ CORS, TLS, Rate Limiting            │
│  2. Authentication    │ JWT (HS256), Argon2id passwords     │
│  3. Authorization     │ Role-based, resource ownership      │
│  4. Data Protection   │ AES-256-GCM encryption, HKDF        │
│  5. Input Validation  │ Schema validation, sanitization     │
│  6. Logging & Audit   │ Security events, access logs        │
└─────────────────────────────────────────────────────────────┘
```

---

## Authentication

### JWT Token System

ClaraVerse v2.0 uses local JWT authentication with HS256 signing.

**Implementation:** `backend/pkg/auth/local_jwt.go`

| Token Type | Default Expiry | Configurable Via |
|------------|----------------|------------------|
| Access Token | 15 minutes | `JWT_ACCESS_TOKEN_EXPIRY` |
| Refresh Token | 7 days | `JWT_REFRESH_TOKEN_EXPIRY` |

**Token Claims:**
```go
type JWTClaims struct {
    UserID  string `json:"sub"`
    Email   string `json:"email"`
    Role    string `json:"role"`
    TokenID string `json:"jti"`  // For refresh token tracking
    jwt.RegisteredClaims
}
```

**Security Properties:**
- Signed with HS256 (HMAC-SHA256)
- Issuer validation (`claraverse-local`)
- Expiration enforced
- Unique token IDs for refresh tokens

### Password Hashing

**Algorithm:** Argon2id (OWASP recommended)

**Implementation:** `backend/pkg/auth/local_jwt.go:146-150`

```go
const (
    argon2Time    = 3           // Number of iterations
    argon2Memory  = 64 * 1024   // 64MB memory
    argon2Threads = 4           // Parallelism factor
    argon2KeyLen  = 32          // 256-bit key
    saltLength    = 16          // 128-bit salt
)
```

**Why Argon2id:**
- Memory-hard: Resistant to GPU/ASIC attacks
- Side-channel resistant (id variant)
- Winner of Password Hashing Competition
- OWASP recommended parameters

### Authentication Middleware

**Implementation:** `backend/internal/middleware/auth.go`

**Security Checks:**
1. Token extraction (Authorization header or query param)
2. Signature verification
3. Expiration validation
4. Claims extraction and storage in context

**Production Safety:**
```go
// CRITICAL: Never allow auth bypass in production
if environment == "production" {
    log.Fatal("CRITICAL SECURITY ERROR: Auth not configured in production")
}
```

---

## Authorization

### Role-Based Access Control

**Roles:**
| Role | Access Level |
|------|--------------|
| `user` | Standard user endpoints |
| `admin` | Admin panel + all user endpoints |

### Admin Access

**Implementation:** `backend/internal/middleware/admin.go`

**Two methods for admin access:**

1. **Role field (preferred):**
   ```go
   if role == "admin" {
       // Grant admin access
   }
   ```

2. **SUPERADMIN_USER_IDS (legacy fallback):**
   ```bash
   SUPERADMIN_USER_IDS=user-id-1,user-id-2
   ```

### Resource Ownership

All user resources are scoped by `user_id`:
- Agents
- Workflows
- Credentials
- Memories
- Chat sessions
- Uploads

**Example ownership check:**
```go
if file.UserID != userID {
    return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
        "error": "Access denied to this file",
    })
}
```

---

## Data Encryption

### Encryption at Rest

**Algorithm:** AES-256-GCM (Authenticated Encryption)

**Implementation:** `backend/internal/crypto/encryption.go`

**Key Derivation:**
```
ENCRYPTION_MASTER_KEY (32 bytes hex)
         │
         ▼
    ┌─────────┐
    │  HKDF   │◀── User ID (salt) + "claraverse-user-key" (info)
    │ SHA-256 │
    └────┬────┘
         │
         ▼
   Per-User Key (32 bytes)
         │
         ▼
    ┌─────────┐
    │ AES-256 │◀── Random 12-byte nonce per encryption
    │   GCM   │
    └────┬────┘
         │
         ▼
   Encrypted Data
   (nonce || ciphertext || auth_tag)
```

**What's Encrypted:**
- User credentials (API keys, OAuth tokens)
- Chat sync data (when using server-side sync)
- Sensitive preferences

**Security Properties:**
- Per-user key isolation
- Authentication tag prevents tampering
- Random nonce prevents replay attacks
- Master key never stored in database

### Browser-Local Storage

Chat conversations are stored in IndexedDB by default:
- **Zero-knowledge architecture** - server never sees chat content
- Encryption happens client-side
- Even admins cannot read user chats

---

## Rate Limiting

**Implementation:** `backend/internal/middleware/rate_limiter.go`

### Rate Limit Configuration

| Endpoint Type | Default | Environment Variable |
|---------------|---------|---------------------|
| Global API | 200/min | `RATE_LIMIT_GLOBAL_API` |
| Public Read | 120/min | `RATE_LIMIT_PUBLIC_READ` |
| Authenticated | 60/min | `RATE_LIMIT_AUTHENTICATED` |
| WebSocket | 20/min | `RATE_LIMIT_WEBSOCKET` |
| Image Proxy | 60/min | `RATE_LIMIT_IMAGE_PROXY` |
| Transcribe | 10/min | Hardcoded |
| Upload | 10/min | Hardcoded |

### Rate Limit Response

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": "Too many requests. Please wait before trying again."
}
```

### Key Generation

Rate limits are tracked by:
- User ID (authenticated requests)
- IP address (anonymous requests)

---

## CORS Configuration

**Implementation:** `backend/cmd/server/main.go`

```go
app.Use(cors.New(cors.Config{
    AllowOriginsFunc: func(origin string) bool {
        // Validate against ALLOWED_ORIGINS
    },
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
    MaxAge:           86400,
}))
```

**Configuration:**
```bash
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### Trigger Endpoints (Open CORS)

External trigger endpoints (`/api/trigger/*`) have permissive CORS:
- Required for webhook integrations
- Protected by API key authentication

---

## API Key Authentication

**Implementation:** `backend/internal/middleware/api_key.go`

### API Key Format
```
cv_live_<random_string>
cv_test_<random_string>
```

### Scopes

| Scope | Permissions |
|-------|-------------|
| `trigger` | Trigger agent executions |
| `upload` | Upload files |
| `read` | Read-only access |

### API Key Header
```http
X-API-Key: cv_live_abc123xyz...
```

---

## Input Validation

### Request Validation

All API inputs are validated:
- JSON schema validation
- Type checking
- Size limits
- Required fields

**Example:**
```go
type CreateAgentRequest struct {
    Name        string   `json:"name" validate:"required,min=1,max=100"`
    Description string   `json:"description" validate:"max=1000"`
    Model       string   `json:"model" validate:"required"`
    Tools       []string `json:"tools" validate:"max=20"`
}
```

### File Upload Validation

- File size limits
- MIME type validation
- Filename sanitization
- Virus scanning (if configured)

---

## Security Headers

Applied via middleware:

```go
c.Set("X-Content-Type-Options", "nosniff")
c.Set("X-Frame-Options", "DENY")
c.Set("X-XSS-Protection", "1; mode=block")
c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
c.Set("Content-Security-Policy", "default-src 'self'")
```

---

## Logging & Audit

### Security Event Logging

```go
log.Printf("✅ Authenticated user: %s (%s)", user.Email, user.ID)
log.Printf("🚫 Non-admin user %s attempted to access admin endpoint", userID)
log.Printf("⚠️  [RATE-LIMIT] Upload limit reached for user: %v", userID)
log.Printf("🚫 [SECURITY] User %s denied access to file %s", userID, fileID)
```

### Log Levels

| Level | Use Case |
|-------|----------|
| INFO | Normal operations |
| WARN | Security events (rate limits, access denied) |
| ERROR | Failures, exceptions |
| FATAL | Critical security failures (production without auth) |

---

## Production Security Checklist

### Required Configuration

- [ ] **Set strong secrets:**
  ```bash
  ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)
  JWT_SECRET=$(openssl rand -hex 64)
  ```

- [ ] **Configure CORS:**
  ```bash
  ALLOWED_ORIGINS=https://yourdomain.com
  ```

- [ ] **Set environment:**
  ```bash
  ENVIRONMENT=production
  ```

- [ ] **Configure TLS** (via reverse proxy)

- [ ] **Change default admin password**

### Recommended Configuration

- [ ] Enable rate limiting (enabled by default)
- [ ] Configure log aggregation
- [ ] Set up monitoring/alerting
- [ ] Regular security updates
- [ ] Database backups with encryption

### Security Anti-Patterns to Avoid

| Anti-Pattern | Risk | Mitigation |
|--------------|------|------------|
| Hardcoded secrets | Credential exposure | Use environment variables |
| Debug mode in production | Information leakage | Set `ENVIRONMENT=production` |
| Permissive CORS | CSRF attacks | Whitelist specific origins |
| Missing rate limits | DoS attacks | Enable rate limiting |
| Logging sensitive data | Data exposure | Sanitize logs |

---

## Vulnerability Handling

### Reporting Security Issues

Report security vulnerabilities via:
- GitHub Security Advisories (private)
- Email: security@claraverse.space

### Security Update Process

1. Issue triaged within 24 hours
2. Fix developed and tested
3. Coordinated disclosure
4. Security advisory published
5. Users notified

---

## Compliance

### GDPR Compliance

ClaraVerse supports GDPR requirements:

| Requirement | Implementation |
|-------------|----------------|
| Data export | `GET /api/user/data` |
| Data deletion | `DELETE /api/user/account` |
| Consent | Privacy policy acceptance |
| Data minimization | Browser-local storage by default |

### Data Retention

Configurable per tier:
- Chat history retention
- Execution log retention
- Automatic cleanup via background jobs

---

## Security Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Internet                                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                      ┌─────┴─────┐
                      │    TLS    │ (Reverse Proxy)
                      └─────┬─────┘
                            │
┌───────────────────────────┼─────────────────────────────────────┐
│                           │         ClaraVerse Backend          │
│                     ┌─────┴─────┐                               │
│                     │   CORS    │                               │
│                     └─────┬─────┘                               │
│                           │                                      │
│                     ┌─────┴─────┐                               │
│                     │   Rate    │                               │
│                     │  Limiter  │                               │
│                     └─────┬─────┘                               │
│                           │                                      │
│              ┌────────────┴────────────┐                        │
│              │                         │                        │
│        ┌─────┴─────┐             ┌─────┴─────┐                 │
│        │  Public   │             │   Auth    │                 │
│        │ Endpoints │             │ Required  │                 │
│        └───────────┘             └─────┬─────┘                 │
│                                        │                        │
│                                  ┌─────┴─────┐                 │
│                                  │    JWT    │                 │
│                                  │  Verify   │                 │
│                                  └─────┬─────┘                 │
│                                        │                        │
│                          ┌─────────────┴─────────────┐         │
│                          │                           │         │
│                    ┌─────┴─────┐               ┌─────┴─────┐  │
│                    │   User    │               │   Admin   │  │
│                    │  Access   │               │   Check   │  │
│                    └─────┬─────┘               └─────┬─────┘  │
│                          │                           │         │
│                    ┌─────┴─────┐               ┌─────┴─────┐  │
│                    │ Ownership │               │   Admin   │  │
│                    │   Check   │               │ Endpoints │  │
│                    └─────┬─────┘               └───────────┘  │
│                          │                                     │
│                    ┌─────┴─────┐                               │
│                    │  Handler  │                               │
│                    └─────┬─────┘                               │
│                          │                                      │
│                    ┌─────┴─────┐                               │
│                    │ Encrypted │                               │
│                    │  Storage  │                               │
│                    └───────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Related Documentation

- [Architecture Guide](ARCHITECTURE.md) - System design
- [API Reference](API_REFERENCE.md) - API documentation
- [Developer Guide](DEVELOPER_GUIDE.md) - Local setup
- [Admin Guide](ADMIN_GUIDE.md) - Administration
- [Quick Reference](QUICK_REFERENCE.md) - Common commands
