# Security Considerations - Frontend

This document outlines security measures implemented in the frontend to protect against common web vulnerabilities.

## 🔒 Security Measures Implemented

### 1. Input Validation

**Location:** `src/utils/security.ts:validateCode()`

All code submissions are validated before sending to the backend:

- ✅ **Size Limits**: Maximum 1MB code size (matches backend limit)
- ✅ **Empty Check**: Prevents empty submissions
- ✅ **Null Byte Detection**: Prevents null byte injection attacks
- ✅ **Client-Side Validation**: Fast feedback without server roundtrip

```typescript
// Validates code size, empty input, and null bytes
const validation = validateCode(code)
if (!validation.valid) {
  // Show error to user
}
```

### 2. Output Sanitization (XSS Prevention)

**Location:** `src/utils/security.ts:sanitizeOutput()`, `src/components/CompilerOutput/CompilerOutput.tsx:23-25`

All compiler output (stdout, stderr, errors) is sanitized before display:

- ✅ **Control Character Removal**: Dangerous control characters removed
- ✅ **OSC Sequence Filtering**: Operating System Command sequences blocked
- ✅ **ANSI Code Whitelisting**: Only safe color codes allowed
- ✅ **Escape Sequence Sanitization**: Malicious escape sequences removed

**Why This Matters:**
Compiler output could contain ANSI escape sequences or control characters that could:
- Move cursor to overwrite sensitive data on screen
- Clear the screen
- Execute terminal commands (in some environments)
- Inject malicious HTML/JavaScript (if not handled properly)

```typescript
const safeStdout = sanitizeOutput(result.stdout)  // Always sanitized before display
```

### 3. Safe Base64 Encoding

**Location:** `src/utils/security.ts:safeBase64Encode()`

Unicode-safe base64 encoding replaces the built-in `btoa()`:

- ✅ **Unicode Support**: Properly handles UTF-8 characters
- ✅ **Error Handling**: Clear error messages for encoding failures
- ✅ **Prevents btoa() Crashes**: Native `btoa()` fails with Unicode

**Why This Matters:**
The native `btoa()` function throws errors with Unicode characters (emojis, special characters), which could crash the application or be exploited.

```typescript
const encodedCode = safeBase64Encode(code)  // Handles Unicode properly
```

### 4. Client-Side Rate Limiting

**Location:** `src/utils/security.ts:RateLimiter`, `src/pages/Home.tsx:32`

Prevents abuse by limiting compilation requests:

- ✅ **10 requests per minute** per client
- ✅ **Token bucket algorithm** for smooth rate limiting
- ✅ **User feedback**: Shows retry time when limit exceeded
- ✅ **Complements backend** rate limiting

```typescript
const rateLimit = rateLimiterRef.current.checkLimit()
if (!rateLimit.allowed) {
  // Show "wait X seconds" message
}
```

### 5. Secure Context Detection

**Location:** `src/utils/security.ts:isSecureContext()`, `src/pages/Home.tsx:38-50`

Warns users when not using HTTPS:

- ✅ **HTTPS Check**: Detects non-secure connections
- ✅ **Warning Display**: Shows security notice banner
- ✅ **Localhost Exception**: No warning for local development
- ✅ **Console Warnings**: Logs security warnings

**Why This Matters:**
Without HTTPS, code and API responses are transmitted in plain text and could be intercepted.

### 6. Error Message Sanitization

**Location:** `src/utils/security.ts:sanitizeErrorMessage()`

Prevents information leakage through error messages:

- ✅ **Path Redaction**: File paths replaced with `[path]`
- ✅ **IP Address Redaction**: IP addresses replaced with `[ip]`
- ✅ **Token Redaction**: Long strings (potential secrets) replaced with `[redacted]`
- ✅ **Generic Fallback**: Unknown errors show generic message

```typescript
const safeError = sanitizeErrorMessage(error)  // Prevents info leakage
```

### 7. Job ID Validation

**Location:** `src/utils/security.ts:isValidJobId()`

Validates job IDs before API requests:

- ✅ **UUID Format Check**: Ensures proper UUID format
- ✅ **Prevents Injection**: Blocks malicious job ID values
- ✅ **Type Safety**: TypeScript validation

### 8. React XSS Protection

React provides built-in XSS protection:

- ✅ **Automatic Escaping**: React escapes JSX expressions by default
- ✅ **No dangerouslySetInnerHTML**: We never use this dangerous API
- ✅ **Safe Text Rendering**: All user content rendered as text, not HTML

## 🚨 Vulnerabilities Prevented

### Cross-Site Scripting (XSS)
- ✅ Output sanitization prevents malicious scripts in compiler output
- ✅ React's automatic escaping prevents HTML injection
- ✅ No use of `dangerouslySetInnerHTML`

### Injection Attacks
- ✅ Input validation prevents null byte injection
- ✅ Job ID validation prevents ID manipulation
- ✅ Base64 encoding ensures safe data transmission

### Denial of Service (DoS)
- ✅ Client-side rate limiting (10 req/min)
- ✅ Code size limits (1MB max)
- ✅ Backend has additional protections

### Information Disclosure
- ✅ Error message sanitization removes sensitive data
- ✅ Secure context warnings for HTTPS
- ✅ No exposure of internal paths or IPs

### Unicode Exploitation
- ✅ Safe base64 encoding handles all Unicode properly
- ✅ Prevents crashes from special characters

## 🛡️ Backend Dependencies

The frontend relies on these backend security measures:

1. **CORS Configuration** - Restricts API access to allowed origins
2. **Rate Limiting** - Server-side request throttling
3. **Input Validation** - Server validates all requests
4. **Container Isolation** - Code runs in sandboxed Docker containers
5. **Resource Limits** - CPU, memory, time limits on compilation
6. **Seccomp Profiles** - Syscall restrictions in containers

See `../../CLAUDE.md` section "Security Layers" for backend security details.

## 📋 Security Best Practices for Developers

### When Adding New Features

1. **Validate All Inputs**
   ```typescript
   const validation = validateCode(input)
   if (!validation.valid) {
     // Handle error
   }
   ```

2. **Sanitize All Outputs**
   ```typescript
   const safeOutput = sanitizeOutput(untrustedOutput)
   ```

3. **Never Use dangerouslySetInnerHTML**
   ```typescript
   // ❌ NEVER DO THIS
   <div dangerouslySetInnerHTML={{ __html: userContent }} />

   // ✅ DO THIS
   <pre>{safeContent}</pre>
   ```

4. **Check Rate Limits for New API Calls**
   ```typescript
   if (!rateLimiter.checkLimit().allowed) {
     // Show error
     return
   }
   ```

5. **Sanitize Error Messages**
   ```typescript
   catch (error) {
     const safeError = sanitizeErrorMessage(error)
     // Display safeError to user
   }
   ```

### Testing Security

```bash
# Test with malicious inputs
curl -X POST http://localhost:3000/api/v1/compile \
  -H "Content-Type: application/json" \
  -d '{"code": "console.log(\"\x00\")"}' # Null byte test

# Test rate limiting
for i in {1..15}; do curl http://localhost:3000/api/v1/compile; done

# Test large code size
# Create 2MB file and submit
```

## 🔍 Security Checklist for Code Reviews

- [ ] All user inputs validated
- [ ] All outputs sanitized
- [ ] No use of `dangerouslySetInnerHTML`
- [ ] Error messages don't leak sensitive info
- [ ] Rate limiting checked for new endpoints
- [ ] No hardcoded secrets or API keys
- [ ] HTTPS required in production
- [ ] Dependencies up to date (no known CVEs)

## 🚀 Production Deployment Security

### Required HTTP Headers

Configure these headers in your web server (nginx/Apache):

```nginx
# Content Security Policy
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' http://localhost:8080 https://your-api-domain.com;" always;

# Prevent clickjacking
add_header X-Frame-Options "SAMEORIGIN" always;

# XSS Protection (legacy browsers)
add_header X-XSS-Protection "1; mode=block" always;

# Prevent MIME sniffing
add_header X-Content-Type-Options "nosniff" always;

# Force HTTPS
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

# Referrer Policy
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
```

### Environment Variables

**Never commit these to git:**
```bash
# .env.local (git-ignored)
VITE_API_URL=https://api.your-domain.com/api/v1  # Production API
```

**Safe to commit:**
```bash
# .env.example (template only)
VITE_API_URL=/api/v1  # Uses Vite proxy in development
```

### Dependency Security

```bash
# Check for vulnerabilities
npm audit

# Fix vulnerabilities
npm audit fix

# Update dependencies
npm update

# Check for outdated packages
npm outdated
```

### CORS Configuration

Ensure backend CORS is properly configured:

```go
// internal/api/middleware.go
AllowOrigins: []string{
  "https://your-frontend-domain.com",  // Production
  "http://localhost:3000",              // Development
}
```

## 📚 Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [React Security](https://react.dev/reference/react-dom/server)
- [ANSI Escape Code Security](https://dgl.cx/2023/09/ansi-terminal-security)

## 🐛 Reporting Security Issues

If you discover a security vulnerability:

1. **DO NOT** open a public issue
2. Email security contact (see main README)
3. Provide detailed description and reproduction steps
4. Allow time for patch before public disclosure

---

**Last Updated**: 2025-11-14
**Security Review Date**: 2025-11-14
**Next Review**: 2025-12-14 (monthly reviews recommended)
