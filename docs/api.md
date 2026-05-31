# BeeTrack API Documentation

Base URL: `http://localhost:8080/api/v1`

All responses are JSON. Errors follow the format:
```json
{ "code": "ERROR_CODE", "message": "human readable message" }
```

---

## Protected Routes

Protected routes require a valid JWT access token in the `Authorization` header:
```
Authorization: Bearer <access_token>
```

If the token is missing or invalid:
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |

---

## Authentication

### POST /auth/register

Creates a new user account.

**Request**
```json
{
  "email": "user@example.com",
  "name": "John",
  "password": "password123"
}
```

**Response** `201 Created`
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John",
  "created_at": "2026-06-01T12:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_EMAIL` | 400 | Email format invalid |
| `WEAK_PASSWORD` | 400 | Password shorter than 8 characters |
| `EMAIL_TAKEN` | 409 | Email already registered |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /auth/login

Authenticates a user and returns token pair.

**Request**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** `200 OK`
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<random string>"
}
```

- `access_token` — JWT, valid for 15 minutes. Send in `Authorization: Bearer <token>` header on protected routes.
- `refresh_token` — valid for 30 days, stored in DB. Use to get a new token pair via `/auth/refresh`.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_CREDENTIALS` | 401 | Wrong email or password |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /auth/refresh

Exchanges a refresh token for a new token pair. The old refresh token is invalidated (rotation).

**Request**
```json
{
  "refresh_token": "<refresh token>"
}
```

**Response** `200 OK`
```json
{
  "access_token": "<new jwt>",
  "refresh_token": "<new refresh token>"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_REFRESH_TOKEN` | 401 | Token not found |
| `TOKEN_EXPIRED` | 401 | Refresh token has expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /auth/logout

Revokes the refresh token. The user must log in again to get a new token pair.

**Request**
```json
{
  "refresh_token": "<refresh token>"
}
```

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Apiaries

### POST /apiaries 🔒

Creates a new apiary. The authenticated user becomes the owner.

**Request**
```json
{
  "name": "My Apiary",
  "lat": 52.23,
  "lng": 21.01,
  "grid_rows": 3,
  "grid_cols": 4
}
```

- `lat` and `lng` are optional
- `grid_rows` and `grid_cols` must be ≥ 1

**Response** `201 Created`
```json
{
  "id": 1,
  "name": "My Apiary",
  "lat": 52.23,
  "lng": 21.01,
  "grid_rows": 3,
  "grid_cols": 4,
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T12:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_BODY` | 400 | Malformed JSON |
| `NAME_REQUIRED` | 400 | Name field is empty |
| `INVALID_GRID_SIZE` | 400 | grid_rows or grid_cols < 1 |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
