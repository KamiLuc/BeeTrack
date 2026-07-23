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

Creates a new unverified user account and sends a confirmation email. The user must click the link in that email before they can log in.

**Request**
```json
{
  "email": "user@example.com",
  "name": "John",
  "password": "password123",
  "lang": "en"
}
```

- `lang` — optional, `"en"` or `"pl"` (default `"en"`). Determines the language of the confirmation email.

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

Authenticates a user and returns token pair. The account must be verified first.

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
  "refresh_token": "<random string>",
  "name": "<display name>"
}
```

- `access_token` — JWT, valid for 15 minutes. Send in `Authorization: Bearer <token>` header on protected routes.
- `refresh_token` — valid for 30 days, stored in DB. Use to get a new token pair via `/auth/refresh`.
- `name` — the user's display name.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_CREDENTIALS` | 401 | Wrong email or password |
| `EMAIL_NOT_VERIFIED` | 403 | Account email not yet verified |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /auth/resend-verification

Sends a new verification email. Always returns 204 regardless of whether the email is registered or already verified (to avoid enumeration).

**Request**
```json
{
  "email": "user@example.com",
  "lang": "en"
}
```

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /auth/verify-email

Called from the link in the verification email. Validates the token, marks the account as verified, and returns a localized HTML confirmation page (not JSON). Intended to be opened directly in a browser.

**Query parameters**
| Parameter | Description |
|-----------|-------------|
| `token` | Verification token |
| `lang` | Optional — `en` or `pl`; controls the language of the HTML page |

**Response** `200 OK` — HTML confirmation page ("Email Verified")

**Error response** `400 Bad Request` — HTML error page ("Verification Failed") if token is missing, expired, or already used

---

### POST /auth/forgot-password

Initiates a password reset by sending a reset link to the given email. Always returns 204 to avoid email enumeration. The reset link points to `GET /auth/reset-password-form` on the API.

**Request**
```json
{
  "email": "user@example.com",
  "lang": "en"
}
```

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /auth/reset-password-form

Serves the HTML password reset form. Intended to be opened directly in a browser from the link in the reset email.

**Query parameters**
| Parameter | Description |
|-----------|-------------|
| `token` | Reset token from the email link |
| `lang` | Optional — `en` or `pl` |

**Response** `200 OK` — HTML form with a password input

**Error response** `400 Bad Request` — HTML error page if token is missing or expired

---

### POST /auth/reset-password-form

Handles HTML form submission from `GET /auth/reset-password-form`. Accepts `application/x-www-form-urlencoded`.

**Form fields**

| Field | Description |
|-------|-------------|
| `token` | Reset token (from hidden input) |
| `password` | New password (min 8 chars) |
| `lang` | Language (from hidden input) |

**Response** `200 OK` — HTML success page on success; re-renders the form with an error message on weak password; HTML expired page on invalid/expired token

---

### POST /auth/reset-password

Validates the reset token and updates the user's password. For API clients (mobile). Invalidates all existing reset tokens for the user.

**Request**
```json
{
  "token": "<reset token>",
  "password": "newpassword123"
}
```

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_RESET_TOKEN` | 400 | Token not found or expired |
| `WEAK_PASSWORD` | 400 | Password shorter than 8 characters |
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
- `grid_rows` and `grid_cols` must be between 1 and 25

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
| `NAME_TOO_LONG` | 400 | `name` exceeds 50 characters |
| `INVALID_GRID_SIZE` | 400 | grid_rows or grid_cols < 1 |
| `GRID_SIZE_TOO_LARGE` | 400 | grid_rows or grid_cols > 25 |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries 🔒

Returns all apiaries the authenticated user belongs to (as owner or member), ordered by creation date descending.

**Response** `200 OK`
```json
[
  {
    "id": 1,
    "name": "My Apiary",
    "lat": 52.23,
    "lng": 21.01,
    "grid_rows": 3,
    "grid_cols": 4,
    "hive_count": 2,
    "user_role": "owner",
    "created_at": "2026-06-01T12:00:00Z",
    "updated_at": "2026-06-01T12:00:00Z"
  }
]
```

- `user_role` — `"owner"` or `"member"`
- Returns an empty array if the user has no apiaries

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /apiaries/{id} 🔒

Updates an apiary. Only the owner can edit.

**Request**
```json
{
  "name": "Updated Name",
  "lat": 52.23,
  "lng": 21.01,
  "grid_rows": 4,
  "grid_cols": 5
}
```

- `lat` and `lng` are optional (omit or pass `null` to clear)
- `grid_rows` and `grid_cols` must be between 1 and 25

**Response** `200 OK`
```json
{
  "id": 1,
  "name": "Updated Name",
  "lat": 52.23,
  "lng": 21.01,
  "grid_rows": 4,
  "grid_cols": 5,
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T13:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `NAME_REQUIRED` | 400 | Name field is empty |
| `NAME_TOO_LONG` | 400 | `name` exceeds 50 characters |
| `INVALID_GRID_SIZE` | 400 | grid_rows or grid_cols < 1 |
| `GRID_SIZE_TOO_LARGE` | 400 | grid_rows or grid_cols > 25 |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `FORBIDDEN` | 403 | Caller is a member, not the owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /apiaries/{id}/copy 🔒

Creates a deep copy of an apiary the user is a member of. The copy is owned by the requesting
user and includes all hives, hive diseases, inspections, and inspection diseases.
Members, invitations, and inspection images are not copied.

**Request body** (optional)

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Name for the new apiary. If omitted or empty, defaults to the source name suffixed with ` (copy)`. |

**Response** `201 Created`

```json
{
  "id": 7,
  "name": "My Apiary (copy)",
  "lat": 52.23,
  "lng": 21.01,
  "grid_rows": 3,
  "grid_cols": 4,
  "created_at": "2025-06-07T12:00:00Z",
  "updated_at": "2025-06-07T12:00:00Z"
}
```

**Errors**

| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No auth token |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id} 🔒

Deletes an apiary and all its members. Only the owner can delete.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `FORBIDDEN` | 403 | Caller is a member, not the owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Hives

### GET /apiaries/{id}/hives 🔒

Returns all hives in an apiary ordered by grid position (`grid_row ASC, grid_col ASC`). Caller must be a member.

**Response** `200 OK`
```json
[
  {
    "id": 1,
    "apiary_id": 1,
    "name": "Hive A",
    "type": "langstroth",
    "active": true,
    "queenless": false,
    "ready_for_harvest": false,
    "needs_food": false,
    "frames": 10,
    "grid_row": 0,
    "grid_col": 0,
    "diseases": [],
    "last_inspected_at": "2026-06-01T10:00:00Z",
    "created_at": "2026-06-01T12:00:00Z",
    "updated_at": "2026-06-01T12:00:00Z"
  }
]
```

- Returns an empty array if the apiary has no hives
- `frames` — total frame capacity of the hive; `0` means not configured
- `last_inspected_at` is `null` when no inspections exist
- `diseases` is an array of `{ "id": 1, "disease": "varroa", "created_at": "..." }`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /apiaries/{id}/hives 🔒

Adds a hive to an apiary. Both owners and members can add hives.

**Request**
```json
{
  "name": "Hive A",
  "type": "langstroth",
  "frames": 10,
  "grid_row": 0,
  "grid_col": 0
}
```

- `type` is optional — defaults to `"langstroth"`
- `frames` is optional — defaults to `0` (not configured)
- `active`, `queenless`, `ready_for_harvest`, `needs_food` default to `false` if omitted
- `grid_row` and `grid_col` are 0-indexed and must fall within the apiary's `grid_rows` × `grid_cols` bounds
- Each position within an apiary must be unique
- `name` must be unique within the apiary, case-insensitive

**Response** `201 Created`
```json
{
  "id": 1,
  "apiary_id": 1,
  "name": "Hive A",
  "type": "langstroth",
  "active": true,
  "queenless": false,
  "ready_for_harvest": false,
  "needs_food": false,
  "frames": 10,
  "grid_row": 0,
  "grid_col": 0,
  "diseases": [],
  "last_inspected_at": null,
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T12:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `NAME_REQUIRED` | 400 | Name field is empty |
| `NAME_TOO_LONG` | 400 | `name` exceeds 50 characters |
| `TYPE_TOO_LONG` | 400 | `type` exceeds 50 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `INVALID_GRID_POSITION` | 400 | Position is outside apiary grid bounds |
| `POSITION_OCCUPIED` | 409 | Another hive already occupies that position |
| `DUPLICATE_HIVE_NAME` | 409 | A hive with this name already exists in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId} 🔒

Returns a single hive. Caller must be a member of the apiary.

**Response** `200 OK`
```json
{
  "id": 1,
  "apiary_id": 1,
  "name": "Hive A",
  "type": "langstroth",
  "active": true,
  "queenless": false,
  "ready_for_harvest": false,
  "needs_food": false,
  "frames": 10,
  "grid_row": 0,
  "grid_col": 0,
  "diseases": [],
  "last_inspected_at": "2026-06-01T10:00:00Z",
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T12:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /apiaries/{id}/hives/{hiveId} 🔒

Updates a hive's name, type, and status flags (`active`, `queenless`, `ready_for_harvest`, `needs_food`). Both owners and members can edit hives.

**Request**
```json
{
  "name": "Renamed Hive",
  "type": "top_bar",
  "active": false,
  "queenless": true,
  "ready_for_harvest": false,
  "needs_food": true,
  "frames": 12
}
```

- `name` must be unique within the apiary, case-insensitive

**Response** `200 OK`
```json
{
  "id": 1,
  "apiary_id": 1,
  "name": "Renamed Hive",
  "type": "top_bar",
  "active": false,
  "queenless": true,
  "ready_for_harvest": false,
  "needs_food": true,
  "frames": 12,
  "grid_row": 0,
  "grid_col": 0,
  "diseases": [],
  "last_inspected_at": "2026-06-01T10:00:00Z",
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T13:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `NAME_REQUIRED` | 400 | Name field is empty |
| `NAME_TOO_LONG` | 400 | `name` exceeds 50 characters |
| `TYPE_TOO_LONG` | 400 | `type` exceeds 50 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `DUPLICATE_HIVE_NAME` | 409 | A hive with this name already exists in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /apiaries/{id}/hives/{hiveId}/frames 🔒

Atomically increments the hive's frame count by `delta`. Used by the app after saving an inspection that added frames, to avoid a race condition with read-modify-write.

**Request**
```json
{
  "delta": 3
}
```

- `delta` must be a positive integer

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /apiaries/{id}/hives/{hiveId}/position 🔒

Moves a hive to a new grid position. Both owners and members can move hives. Moving to the current position is a no-op.

**Request**
```json
{
  "grid_row": 2,
  "grid_col": 3
}
```

- Position must be within the apiary's `grid_rows` × `grid_cols` bounds (0-indexed)
- Target position must be unoccupied by another hive

**Response** `200 OK`
```json
{
  "id": 1,
  "apiary_id": 1,
  "name": "Hive A",
  "type": "langstroth",
  "active": true,
  "queenless": false,
  "ready_for_harvest": false,
  "needs_food": false,
  "frames": 10,
  "grid_row": 2,
  "grid_col": 3,
  "diseases": [],
  "last_inspected_at": "2026-06-01T10:00:00Z",
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T13:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INVALID_GRID_POSITION` | 400 | Position is outside apiary grid bounds |
| `POSITION_OCCUPIED` | 409 | Another hive already occupies that position |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/hives/{hiveId} 🔒

Deletes a hive. Both owners and members can delete hives.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /apiaries/{id}/hives/{hiveId}/transfer 🔒

Moves a hive from its current apiary to a different one. The hive is placed at the first available grid position in the target apiary.

- The target apiary must not already have a hive with the same name (case-insensitive)

**Request**
```json
{ "target_apiary_id": 5 }
```

**Response** `200 OK` — updated hive object (same shape as GET /apiaries/{id}/hives/{hiveId}).

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{hiveId}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `SAME_APIARY` | 400 | `target_apiary_id` equals the source apiary |
| `APIARY_NOT_FOUND` | 404 | Source or target apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in the source apiary |
| `TARGET_APIARY_FULL` | 409 | All grid cells in the target apiary are occupied |
| `DUPLICATE_HIVE_NAME` | 409 | The target apiary already has a hive with this name |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Hive Diseases

### POST /apiaries/{id}/hives/{hiveId}/diseases 🔒

Adds a disease to a hive.

**Request**
```json
{
  "disease": "varroa"
}
```

Valid disease values: `varroa`, `nosema`, `dwv`, `american_foulbrood`, `chalkbrood`, `european_foulbrood`, `laying_workers`

**Response** `201 Created`
```json
{
  "id": 1,
  "disease": "varroa",
  "created_at": "2026-06-01T12:00:00Z"
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_DISEASE` | 400 | Value not in allowed set |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/hives/{hiveId}/diseases/{diseaseId} 🔒

Removes a disease from a hive.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `HIVE_DISEASE_NOT_FOUND` | 404 | Disease does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Users

### GET /users/me 🔒

Returns the authenticated caller's own profile.

**Response** `200 OK`
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John",
  "role": "user",
  "verified": true
}
```

- `role` — `"user"` or `"admin"`. Client-side UX only (e.g. gating the admin panel's login screen) — every admin route is still enforced server-side by `RequireAdmin`.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `USER_NOT_FOUND` | 404 | Caller's user record no longer exists |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /users/me/name 🔒

Updates the authenticated user's display name.

**Request**
```json
{
  "name": "New Name"
}
```

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_BODY` | 400 | Malformed JSON |
| `NAME_REQUIRED` | 400 | Name field is empty |
| `NAME_TOO_LONG` | 400 | `name` exceeds 50 characters |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Inspections

All inspection endpoints are nested under a hive: `/apiaries/{id}/hives/{hiveId}/inspections`.

### Inspection object

```json
{
  "id": 1,
  "hive_id": 10,
  "inspected_by": 1,
  "inspected_by_name": "John Doe",
  "inspected_at": "2026-06-04T10:00:00Z",
  "queen_status": "seen",
  "brood_pattern": "good",
  "frames_brood": 5,
  "frames_feed": 4,
  "frames_pollen": 2,
  "queen_cells_count": 0,
  "aggressiveness": "calm",
  "frames_added_foundation": 1,
  "frames_added_drawn": null,
  "frames_added_brood": null,
  "frames_added_feed": null,
  "queen_added": false,
  "notes": "Colony looks healthy.",
  "photo_count": 2,
  "diseases": [],
  "created_at": "2026-06-04T10:05:00Z",
  "updated_at": "2026-06-04T10:05:00Z"
}
```

- All observation fields are optional — omit or send `null` to leave unrecorded
- `queen_status` valid values: `seen`, `not_seen`
- `brood_pattern` valid values: `excellent`, `good`, `poor`, `none`
- `aggressiveness` valid values: `calm`, `mild`, `aggressive`, `very_aggressive`
- `frames_brood` — nullable int, frames of brood observed
- `frames_added_foundation`, `frames_added_drawn`, `frames_added_brood`, `frames_added_feed` — signed frame-delta counts for this inspection; positive means frames were added to the hive, negative means frames were taken/removed
- `photo_count` — number of images attached to this inspection (only present in list responses)
- `diseases` — array of disease objects (see below); always included in responses

---

### POST /apiaries/{id}/hives/{hiveId}/inspections 🔒

Creates a new inspection for the hive. Caller must be a member of the apiary.

**Request**
```json
{
  "inspected_at": "2026-06-04T10:00:00Z",
  "queen_status": "seen",
  "brood_pattern": "good",
  "frames_brood": 5,
  "frames_feed": 4,
  "frames_pollen": 2,
  "queen_cells_count": 0,
  "aggressiveness": "calm",
  "frames_added_foundation": 1,
  "frames_added_drawn": null,
  "frames_added_brood": null,
  "frames_added_feed": null,
  "queen_added": false,
  "notes": "Colony looks healthy."
}
```

**Response** `201 Created` — inspection object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `INSPECTED_AT_REQUIRED` | 400 | `inspected_at` missing or zero |
| `INVALID_QUEEN_STATUS` | 400 | Value not in allowed set |
| `INVALID_BROOD_PATTERN` | 400 | Value not in allowed set |
| `INVALID_AGGRESSIVENESS` | 400 | Value not in allowed set |
| `NOTES_TOO_LONG` | 400 | `notes` exceeds 5000 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId}/inspections 🔒

Returns a paginated list of inspections for the hive ordered by `inspected_at` descending. Each item includes `photo_count`.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |

**Response** `200 OK`
```json
{
  "items": [ /* inspection objects */ ],
  "total": 42
}
```
- `total` — total number of inspections for the hive (used for pagination)

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} 🔒

Returns a single inspection.

**Response** `200 OK` — inspection object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} 🔒

Overwrites all mutable fields of an inspection. Send the complete desired state.

**Request** — same shape as POST

**Response** `200 OK` — updated inspection object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `INSPECTED_AT_REQUIRED` | 400 | `inspected_at` missing or zero |
| `INVALID_QUEEN_STATUS` | 400 | Value not in allowed set |
| `INVALID_BROOD_PATTERN` | 400 | Value not in allowed set |
| `INVALID_AGGRESSIVENESS` | 400 | Value not in allowed set |
| `NOTES_TOO_LONG` | 400 | `notes` exceeds 5000 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} 🔒

Deletes an inspection and all its diseases (cascade).

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Inspection Diseases

Diseases are nested under inspections. Valid `disease` values: `varroa`, `nosema`, `dwv`, `american_foulbrood`, `chalkbrood`, `european_foulbrood`, `laying_workers`.

The `diseases` array is always included in inspection responses — no separate list endpoint needed.

### Disease object

```json
{
  "id": 1,
  "disease": "nosema",
  "created_at": "2026-06-04T10:05:00Z"
}
```

---

### POST /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases 🔒

Adds a disease to an inspection.

**Request**
```json
{
  "disease": "nosema"
}
```

**Response** `201 Created` — disease object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `INVALID_DISEASE` | 400 | Value not in allowed set |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases/{diseaseId} 🔒

Removes a disease from an inspection.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `DISEASE_NOT_FOUND` | 404 | Disease does not exist for this inspection |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Inspection Images

Images are stored on the server under a Docker volume. Accepted MIME types: `image/jpeg`, `image/png`, `image/webp`. Maximum file size: **5 MB**.

Images are cascade-deleted when the parent inspection is deleted. File cleanup on disk is performed before the DB row is removed.

**Image object**
```json
{
  "id": 1,
  "inspection_id": 5,
  "mime_type": "image/jpeg",
  "created_at": "2025-06-06T10:00:00Z"
}
```

---

### POST /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images 🔒

Uploads an image. Send as `multipart/form-data` with field name `image`.

**Response** `201 Created` — image object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `MISSING_FILE` | 400 | `image` field missing from form |
| `INVALID_IMAGE_TYPE` | 400 | MIME type not allowed |
| `IMAGE_TOO_LARGE` | 413 | File exceeds 5 MB |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images 🔒

Returns all image metadata for an inspection ordered by id ascending.

**Response** `200 OK` — array of image objects

---

### GET /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}/file 🔒

Serves the raw image bytes with the correct `Content-Type` header. Cached for 24 hours (`Cache-Control: private, max-age=86400`).

**Response** `200 OK` — image binary

---

### DELETE /apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId} 🔒

Deletes an image from the DB and from disk.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
| `INSPECTION_NOT_FOUND` | 404 | Inspection does not exist for this hive |
| `IMAGE_NOT_FOUND` | 404 | Image does not exist for this inspection |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Invitations

Apiary owners can invite registered users by email. Invited users see pending invitations and can accept or decline them.

### Invitation object (apiary view)

```json
{
  "id": 1,
  "apiary_id": 5,
  "invited_email": "guest@example.com",
  "status": "pending",
  "created_at": "2026-06-07T12:00:00Z"
}
```

### Member object (apiary view)

```json
{
  "user_id": 3,
  "email": "guest@example.com",
  "name": "Guest User",
  "role": "member",
  "joined_at": "2026-06-07T12:00:00Z"
}
```

### My invitation object

```json
{
  "id": 1,
  "apiary_id": 5,
  "apiary_name": "My Apiary",
  "invited_by_name": "Owner Name",
  "created_at": "2026-06-07T12:00:00Z"
}
```

---

### POST /apiaries/{id}/invitations 🔒

Sends an invitation to the given email. Only the apiary owner can invite. The email must belong to a registered account.

**Request**
```json
{
  "email": "guest@example.com"
}
```

**Response** `201 Created` — invitation object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON or missing email |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist |
| `FORBIDDEN` | 403 | Caller is not the owner |
| `CANNOT_INVITE_SELF` | 400 | Owner tried to invite themselves |
| `USER_NOT_FOUND` | 404 | No account found for that email |
| `ALREADY_MEMBER` | 409 | User is already a member |
| `INVITATION_PENDING` | 409 | A pending invitation already exists for this email |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/invitations 🔒

Returns members and pending invitations for the apiary. Only the owner can call this.

**Response** `200 OK`
```json
{
  "members": [ /* member objects */ ],
  "invitations": [ /* invitation objects */ ]
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist |
| `FORBIDDEN` | 403 | Caller is not the owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/invitations/{invitationId} 🔒

Cancels a pending invitation. Only the owner can cancel invitations.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist |
| `FORBIDDEN` | 403 | Caller is not the owner |
| `INVITATION_NOT_FOUND` | 404 | Invitation does not exist for this apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/members/{userId} 🔒

Removes a member from the apiary. Only the owner can remove members. The owner cannot remove themselves.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path id is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist |
| `FORBIDDEN` | 403 | Caller is not the owner |
| `CANNOT_REMOVE_OWNER` | 400 | Cannot remove the apiary owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /apiaries/{id}/leave 🔒

Leaves an apiary. The owner cannot leave — delete the apiary instead.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `CANNOT_LEAVE_AS_OWNER` | 400 | Owner cannot leave the apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /invitations 🔒

Returns all pending invitations addressed to the authenticated user.

**Response** `200 OK` — array of my invitation objects

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /invitations/count 🔒

Returns the count of pending invitations for the authenticated user. Used for the badge indicator.

**Response** `200 OK`
```json
{
  "count": 3
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /invitations/{id}/accept 🔒

Accepts a pending invitation. The caller must be the invited user.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVITATION_NOT_FOUND` | 404 | Invitation does not exist |
| `FORBIDDEN` | 403 | Invitation does not belong to the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /invitations/{id}/decline 🔒

Declines a pending invitation. The caller must be the invited user.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVITATION_NOT_FOUND` | 404 | Invitation does not exist |
| `FORBIDDEN` | 403 | Invitation does not belong to the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Medicines

### GET /medicines

Returns the list of known medicine names. No authentication required. Intended for autocomplete in treatment forms and AI/MCP voice assistants.

**Response** `200 OK`
```json
["Api Life Var","Api-Bioxal","Apiguard","Apivar","Apistan","Apiwarol","Bayvarol","Biowar 500","Formicpro","MAQS","Oxuvar","PolyVar Yellow","VarroMed"]
```

---

## Treatments

### POST /apiaries/{id}/treatments/bulk 🔒

Creates one treatment record per hive in the apiary, all inside a single transaction. Same request body as the single-hive create endpoint.

**Request**
```json
{
  "treated_at": "2026-06-08T10:00:00Z",
  "medicine_name": "Apiwarol",
  "dose": "2",
  "notes": "applied evenly"
}
```

- `dose` — optional; defaults to `"1"` if omitted or empty.

**Response** `201 Created`
```json
{ "count": 5 }
```

- `count` — number of treatments inserted (equals number of hives in the apiary at the time of the call). `0` if the apiary has no hives.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `TREATED_AT_REQUIRED` | 400 | `treated_at` missing or zero |
| `MEDICINE_NAME_REQUIRED` | 400 | `medicine_name` empty |
| `MEDICINE_NAME_TOO_LONG` | 400 | `medicine_name` exceeds 50 characters |
| `DOSE_TOO_LONG` | 400 | `dose` exceeds 20 characters |
| `NOTES_TOO_LONG` | 400 | `notes` exceeds 5000 characters |

---

### GET /apiaries/{id}/hives/{hiveId}/treatments 🔒

Returns paginated treatments for a hive, ordered by `treated_at` descending.

**Query Parameters**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Max items to return |
| `offset` | int | 0 | Number of items to skip |

**Response** `200 OK`
```json
{
  "items": [
    {
      "id": 1,
      "hive_id": 10,
      "treated_by": 5,
      "treated_by_name": "Alice",
      "treated_at": "2025-06-01T10:00:00Z",
      "medicine_name": "Apiwarol",
      "dose": "2",
      "notes": "Applied on frames",
      "created_at": "2025-06-01T10:05:00Z",
      "updated_at": "2025-06-01T10:05:00Z"
    }
  ],
  "total": 1
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or caller is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not belong to the apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /apiaries/{id}/hives/{hiveId}/treatments 🔒

Creates a new treatment record. Caller must be a member of the apiary.

**Request Body**
```json
{
  "treated_at": "2025-06-01T10:00:00Z",
  "medicine_name": "Apiwarol",
  "dose": "2",
  "notes": "Applied on frames"
}
```

`dose` defaults to `"1"` if omitted or empty.

**Response** `201 Created` — treatment object (same shape as list item above)

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `TREATED_AT_REQUIRED` | 400 | `treated_at` is missing or zero |
| `MEDICINE_NAME_REQUIRED` | 400 | `medicine_name` is empty |
| `MEDICINE_NAME_TOO_LONG` | 400 | `medicine_name` exceeds 50 characters |
| `DOSE_TOO_LONG` | 400 | `dose` exceeds 20 characters |
| `NOTES_TOO_LONG` | 400 | `notes` exceeds 5000 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary not found / not a member |
| `HIVE_NOT_FOUND` | 404 | Hive not found |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} 🔒

Returns a single treatment.

**Response** `200 OK` — treatment object

**Errors** — same as POST above plus `TREATMENT_NOT_FOUND` 404.

---

### PATCH /apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} 🔒

Overwrites all mutable fields of an existing treatment.

**Request Body** — same as POST (all fields required)

**Response** `200 OK` — updated treatment object

**Errors** — same as POST above plus `TREATMENT_NOT_FOUND` 404.

---

### DELETE /apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} 🔒

Deletes a treatment record.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `TREATMENT_NOT_FOUND` | 404 | Treatment not found |
| `APIARY_NOT_FOUND` | 404 | Apiary not found / not a member |
| `HIVE_NOT_FOUND` | 404 | Hive not found |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Harvests

### GET /apiaries/{id}/hives/{hiveId}/harvests 🔒

Returns paginated harvests for a hive, ordered by `harvested_at` descending.

**Query Parameters**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Max items to return |
| `offset` | int | 0 | Number of items to skip |

**Response** `200 OK`
```json
{
  "items": [
    {
      "id": 1,
      "hive_id": 10,
      "harvested_by": 5,
      "harvested_by_name": "Alice",
      "harvested_at": "2025-08-10T09:00:00Z",
      "frames": 8,
      "half_frames": 2,
      "kilograms": 24.50,
      "notes": "Good harvest",
      "created_at": "2025-08-10T09:05:00Z",
      "updated_at": "2025-08-10T09:05:00Z"
    }
  ],
  "total": 1
}
```

`harvested_by_name` is `null` when the recording user is unknown (harvests created before attribution tracking was added).

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or caller is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not belong to the apiary |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /apiaries/{id}/hives/{hiveId}/harvests 🔒

Creates a new harvest record. Caller must be a member of the apiary. `harvested_by` is set automatically from the authenticated user.

**Request Body**
```json
{
  "harvested_at": "2025-08-10T09:00:00Z",
  "frames": 8,
  "half_frames": 2,
  "kilograms": 24.50,
  "notes": "Good harvest"
}
```

`frames` and `half_frames` default to `0` if omitted. `notes` defaults to `""`.

**Response** `201 Created` — harvest object (same shape as list item above)

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `HARVESTED_AT_REQUIRED` | 400 | `harvested_at` is missing or zero |
| `HARVEST_FRAMES_REQUIRED` | 400 | Both `frames` and `half_frames` are zero |
| `HARVEST_KILOGRAMS_REQUIRED` | 400 | `kilograms` is zero or negative |
| `NOTES_TOO_LONG` | 400 | `notes` exceeds 5000 characters |
| `APIARY_NOT_FOUND` | 404 | Apiary not found / not a member |
| `HIVE_NOT_FOUND` | 404 | Hive not found |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /apiaries/{id}/hives/{hiveId}/harvests/{harvestId} 🔒

Returns a single harvest.

**Response** `200 OK` — harvest object

**Errors** — same as POST above plus `HARVEST_NOT_FOUND` 404.

---

### PATCH /apiaries/{id}/hives/{hiveId}/harvests/{harvestId} 🔒

Overwrites all mutable fields of an existing harvest.

**Request Body** — same as POST (all fields required)

**Response** `200 OK` — updated harvest object

**Errors** — same as POST above plus `HARVEST_NOT_FOUND` 404.

---

### DELETE /apiaries/{id}/hives/{hiveId}/harvests/{harvestId} 🔒

Deletes a harvest record.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `HARVEST_NOT_FOUND` | 404 | Harvest not found |
| `APIARY_NOT_FOUND` | 404 | Apiary not found / not a member |
| `HIVE_NOT_FOUND` | 404 | Hive not found |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Honey Batches

Honey batches record a harvest lot for certification. Every endpoint is scoped to the authenticated caller as owner — a batch belongs to whichever user created it, with no apiary scoping.

### Honey batch object

```json
{
  "id": 1,
  "verification_token": "b2e1c9e0-....-....-....-............",
  "verification_url": "http://localhost:8080/verify/b2e1c9e0-....-....-....-............",
  "gathering_date": "2026-07-01T00:00:00Z",
  "amount_grams": 5000,
  "processing_method": "raw",
  "honey_type": "Wildflower",
  "pdf_file_hash": "b7e2...",
  "pdf_filename": "lab-report.pdf",
  "metadata_hash": "9f1a...",
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-01T10:00:00Z",
  "certification": null,
  "certification_request": null
}
```

- `verification_token` — opaque UUID identifying the batch publicly
- `verification_url` — the public verification page URL for this batch, i.e. `GET /verify/{token}` (see below); built from the server's own public base URL
- `processing_method` valid values: `raw`, `filtered`, `pasteurized`
- `pdf_file_hash` — SHA-256 hex digest of the uploaded lab PDF
- `pdf_filename` — original filename of the uploaded lab PDF
- `metadata_hash` — SHA-256 hex digest of the batch's certifiable metadata, recomputed whenever mutable fields change
- `certification` — `null` until an admin approves a certification request (see `certification_request` below) and the background worker claims the resulting blockchain job; otherwise `{status, chain_id, contract_address, transaction_hash, block_number, gas_used, confirmation_timestamp, created_at}` reflecting the real row's latest status.
- `certification_request` — `null` if certification was never requested; otherwise `{status, rejection_reason, created_at}` describing the pending/approved/rejected admin review request. `status` is `"pending"` immediately after requesting certification (via `POST /honey-batches` with `request_certification=true` or `POST /honey-batches/{id}/retry-certification`); it's only once an admin approves the request that a `blockchain_jobs` row is enqueued and `certification` starts reflecting real on-chain progress.

---

### POST /honey-batches 🔒

Creates a honey batch. Request is `multipart/form-data` (not JSON) because it includes the lab PDF file.

**Form fields**
| Field | Type | Description |
|-------|------|-------------|
| `gathering_date` | string | `YYYY-MM-DD` |
| `amount_grams` | int | Grams of honey in the batch; must be > 0 and ≤ 100,000,000 |
| `processing_method` | string | `raw`, `filtered`, or `pasteurized` |
| `honey_type` | string | Free text, max 50 characters |
| `request_certification` | string | `"true"` or `"false"`; defaults to `"false"` if omitted |
| `lab_pdf` | file | Lab analysis PDF, content type must be `application/pdf`, max 10 MB |

**Response** `201 Created` — honey batch object. `certification` is always `null` on creation (never enqueued directly). `certification_request` is `null` unless `request_certification` was `"true"`, in which case it's `{"status": "pending", ...}` — an admin must approve it before any blockchain job is enqueued.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_BODY` | 400 | Malformed multipart form |
| `INVALID_DATE` | 400 | `gathering_date` is not `YYYY-MM-DD` |
| `INVALID_AMOUNT` | 400 | `amount_grams` missing/not an integer, or out of range (0 < n ≤ 100,000,000) |
| `HONEY_TYPE_REQUIRED` | 400 | `honey_type` is empty |
| `HONEY_TYPE_TOO_LONG` | 400 | `honey_type` exceeds 50 characters |
| `INVALID_PROCESSING_METHOD` | 400 | `processing_method` not one of the allowed values |
| `MISSING_FILE` | 400 | `lab_pdf` field missing from form |
| `INVALID_PDF_TYPE` | 400 | `lab_pdf` content type is not `application/pdf` |
| `PDF_TOO_LARGE` | 413 | `lab_pdf` exceeds 10 MB |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /honey-batches/{id} 🔒

Returns a single batch owned by the caller.

**Response** `200 OK` — honey batch object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /honey-batches 🔒

Returns a paginated list of the caller's batches, each with its latest certification.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |

**Response** `200 OK`
```json
{
  "items": [ /* honey batch objects */ ],
  "total": 3
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /honey-batches/{id} 🔒

Updates a batch's `gathering_date`, `amount_grams`, `processing_method`, and `honey_type` — all mutable fields. Recomputes `metadata_hash` from the new values. Request is `multipart/form-data` (not JSON), same field names as `POST /honey-batches`. Locked (409) once the batch has any certification attempt (any certification row, or a pending/unclaimed blockchain job).

**Form fields**
| Field | Type | Description |
|-------|------|-------------|
| `gathering_date` | string | `YYYY-MM-DD` |
| `amount_grams` | int | Grams of honey in the batch; must be > 0 and ≤ 100,000,000 |
| `processing_method` | string | `raw`, `filtered`, or `pasteurized` |
| `honey_type` | string | Free text, max 50 characters |
| `lab_pdf` | file | Optional. If provided, replaces the batch's existing lab PDF (same validation as Create: content type must be `application/pdf`, max 10 MB). If omitted, the existing PDF is left untouched. |
| `remove_pdf` | string | Optional. Set to `"true"` to clear the batch's existing lab PDF (`lab_pdf_url`, `pdf_filename`, `pdf_file_hash`) and delete the stored file. Ignored if `lab_pdf` is also provided — a new upload always takes precedence. |

**Response** `200 OK` — updated honey batch object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed multipart form |
| `INVALID_DATE` | 400 | `gathering_date` is not `YYYY-MM-DD` |
| `INVALID_AMOUNT` | 400 | `amount_grams` is not greater than 0 or exceeds 100,000,000 |
| `INVALID_PROCESSING_METHOD` | 400 | `processing_method` is not one of `raw`, `filtered`, `pasteurized` |
| `HONEY_TYPE_REQUIRED` | 400 | `honey_type` is empty |
| `HONEY_TYPE_TOO_LONG` | 400 | `honey_type` exceeds 50 characters |
| `INVALID_PDF_TYPE` | 400 | `lab_pdf` content type is not `application/pdf` |
| `PDF_TOO_LARGE` | 413 | `lab_pdf` exceeds 10 MB |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `BATCH_LOCKED` | 409 | Batch has an existing certification attempt and can no longer be edited |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /honey-batches/{id} 🔒

Soft-deletes a batch owned by the caller. Any existing on-chain certification is untouched.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /honey-batches/{id}/pdf 🔒

Serves the lab PDF for a batch owned by the caller. Not gated on certification status.

**Response** `200 OK` — `application/pdf` binary

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /honey-batches/{id}/certifications 🔒

Returns the full certification history for a batch owned by the caller — every attempt, most recent first (not just the latest, which is what the honey batch object's `certification` field carries).

**Response** `200 OK`:

```json
{
  "items": [
    {
      "status": "confirmed",
      "chain_id": 80002,
      "contract_address": "0x...",
      "transaction_hash": "0x...",
      "block_number": 12345,
      "gas_used": 123456,
      "confirmation_timestamp": "2026-07-01T10:05:00Z",
      "created_at": "2026-07-01T10:00:00Z",
      "on_chain_pdf_hash": "b7e2...",
      "on_chain_metadata_hash": "9f1a..."
    }
  ]
}
```

`on_chain_pdf_hash`/`on_chain_metadata_hash` (hex strings, no `0x` prefix) are only attached to the current live/confirmed row (`status` one of `submitted`, `pending_confirmation`, `confirmed`), fetched live from the chain with a 5-second timeout. They're omitted entirely — never an error — if blockchain isn't configured on the server or the RPC read fails/times out; this endpoint never fails because of the chain read.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /honey-batches/{id}/retry-certification 🔒

Submits a batch owned by the caller for admin certification review — a first-time request if certification was never requested, or a retry if the latest attempt is `failed`/`reverted`. Creates a `HoneyBatchCertificationRequest` awaiting admin approval; it no longer enqueues a blockchain job directly. No request body.

**Response** `200 OK` — honey batch object. `certification_request` is `{"status": "pending", ...}`.

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist or is not owned by the caller |
| `BATCH_ALREADY_CERTIFIED` | 409 | Batch already has a live certification (`submitted`, `pending_confirmation`, or `confirmed`) |
| `CERTIFICATION_REQUEST_PENDING` | 409 | Batch already has a certification request awaiting admin review |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /verify/{token}

Public, no authentication. Looks up a batch by its public `verification_token` (opaque UUID, not the numeric `id`) and returns the batch's public representation — same fields as the honey batch object above, minus the internal numeric `id` and `certification_request` (that state is owner-only).

**Response** `200 OK`:

```json
{
  "verification_token": "b2e1c9e0-....-....-....-............",
  "verification_url": "http://localhost:8080/verify/b2e1c9e0-....-....-....-............",
  "gathering_date": "2026-07-01T00:00:00Z",
  "amount_grams": 5000,
  "processing_method": "raw",
  "honey_type": "Wildflower",
  "pdf_file_hash": "b7e2...",
  "pdf_filename": "lab-report.pdf",
  "metadata_hash": "9f1a...",
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-01T10:00:00Z",
  "certification": null
}
```

`certification` is `null`, or the real certification object (`{status, chain_id, contract_address, transaction_hash, block_number, gas_used, confirmation_timestamp, created_at}`) reflecting its full lifecycle status (`queued`, `submitting`, `submitted`, `pending_confirmation`, `confirmed`, `failed`, `reverted`).

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `BATCH_NOT_FOUND` | 404 | Token does not resolve to a batch |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /verify/{token} (HTML page)

**Not under the `/api/v1` base URL** — a plain top-level route: `GET {apiURL}/verify/{token}` (`apiURL` is the backend's own public URL, e.g. `http://localhost:8080/verify/{token}`). Public, no authentication.

This is the exact URL encoded in a batch's QR code and its "verification_url" field — a self-contained, dependency-free HTML page (Go `html/template`, no JS, no SPA) so the link opens directly in any browser, including one launched straight from a phone's stock camera app scanning the QR. There is no in-app QR scanner or Flutter verification screen; this page is the entire verification experience.

Shows honey type, processing method, gathering date, amount, batch ID (the verification token, never the internal numeric id), certification status, lab PDF hash, and metadata hash. Once the certification is confirmed, also shows the contract address, block number, transaction hash, and a link to the Polygon Amoy block explorer (`https://amoy.polygonscan.com/tx/{hash}`).

An intro paragraph above the hashes (and a one-line explainer under each) explains what the two hashes are and why the live check below matters. Once the certification is confirmed, the page also does a live read against the deployed smart contract (bounded by a 5-second timeout so a slow RPC never hangs the page) and shows a badge under each hash: "Matches the record on the blockchain" (green) if it equals the on-chain value, "Does not match — data may have changed" (red) if it doesn't, or "Live check unavailable right now" (grey) if blockchain isn't configured on the server or the RPC call errors/times out. This live check never causes the page itself to fail — it only affects the badges.

Once the certification is confirmed, a "Download lab PDF" button (pointing at the public `GET /verify/{token}/pdf` endpoint above) is shown alongside the "View on block explorer" button, in a row at the bottom of the proof card.

Bilingual: `?lang=pl` or `?lang=en` query param overrides the language; otherwise the `Accept-Language` header is sniffed for `pl`, defaulting to English.

**Response** `200 OK` (or `404 Not Found` if the token doesn't resolve) — `text/html; charset=utf-8`

---

### GET /verify/{token}/qr-code

Public, no authentication. Serves a 512x512 PNG QR code encoding the verification URL `{apiURL}/verify/{token}` (the HTML page above). Requires the batch to have a confirmed certification — a QR pointing at an uncertified batch would be misleading.

**Response** `200 OK` — `image/png` binary, `Cache-Control: public, max-age=31536000, immutable`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `BATCH_NOT_FOUND` | 404 | Token does not resolve to a batch |
| `BATCH_NOT_CERTIFIED` | 409 | Batch has no confirmed certification |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /verify/{token}/qr-code/download

Public, no authentication. Serves the identical PNG as `GET /verify/{token}/qr-code`, but adds a `Content-Disposition: attachment` header so the response triggers a browser download/save dialog instead of rendering inline. Same requirement (confirmed certification) and same caching.

The download filename is derived from the batch's metadata: `{gathering_date}_{honey_type}_{weight}kg.png` (e.g. `2024-05-01_wildflower_1.5kg.png`), with the honey type lowercased and non-alphanumeric characters collapsed to hyphens.

**Response** `200 OK` — `image/png` binary, `Content-Disposition: attachment; filename="{gathering_date}_{honey_type}_{weight}kg.png"`, `Cache-Control: public, max-age=31536000, immutable`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `BATCH_NOT_FOUND` | 404 | Token does not resolve to a batch |
| `BATCH_NOT_CERTIFIED` | 409 | Batch has no confirmed certification |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /verify/{token}/pdf

Public, no authentication. Serves the lab PDF for a batch. Requires a confirmed certification — no public exposure of lab data for an uncertified batch.

**Response** `200 OK` — `application/pdf` binary, `Cache-Control: public, max-age=86400`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `BATCH_NOT_FOUND` | 404 | Token does not resolve to a batch |
| `BATCH_NOT_CERTIFIED` | 409 | Batch has no confirmed certification |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Marketplace Listings

Listings are public classifieds posted by users. Auth is optional on read routes (`GET`) — an authenticated caller sees their own hidden and pending listings and can filter to `mine=true`; anonymous callers only see visible, approved listings.

New listings default to a `pending` moderation status and are hidden from public search/`GET` until an admin approves them (see [Admin: Marketplace Listing Moderation](#admin-marketplace-listing-moderation)). Editing an already-approved listing resets its status back to `pending`. The listing object returned by these routes does not include the `status`/`rejection_reason` fields — those are only exposed on the admin endpoints.

### Listing object

```json
{
  "id": 1,
  "user_id": 3,
  "title": "Wildflower honey, 500g jars",
  "description": "Raw, unfiltered, harvested this summer.",
  "category": "HONEY",
  "price": 25.00,
  "quantity": "10 jars",
  "address": "Warsaw, Poland",
  "apiary_id": 5,
  "apiary_name": "My Apiary",
  "lat": 52.23,
  "lng": 21.01,
  "distance_km": 3.4,
  "contact_phone": "+48123456789",
  "contact_email": "seller@example.com",
  "is_hidden": false,
  "created_at": "2026-07-01T10:00:00Z",
  "updated_at": "2026-07-01T10:00:00Z",
  "images": [
    {
      "id": 1,
      "listing_id": 1,
      "url": "/api/v1/listings/1/images/1/file",
      "display_order": 0,
      "created_at": "2026-07-01T10:00:00Z"
    }
  ],
  "honey_batch_id": 12,
  "honey_batch": {
    "id": 12,
    "honey_type": "wildflower",
    "gathering_date": "2026-06-15T00:00:00Z",
    "amount_grams": 5000,
    "processing_method": "raw",
    "certification_status": "confirmed",
    "has_pdf": true,
    "verification_url": "https://.../verify/<token>",
    "pdf_url": "https://.../verify/<token>/pdf"
  }
}
```

- `category` valid values: `HONEY`, `POLLEN`, `BEE_COLONIES`, `QUEEN_BEES`, `BEEHIVES`, `POPULATED_BEEHIVES`, `EQUIPMENT`, `EXTRACTION_EQUIPMENT`, `FEED`, `SUPPLIES`, `WAX_FOUNDATION`, `BEESWAX`, `PROPOLIS`, `SERVICES`, `OTHER`
- `price` — nullable number
- `apiary_id` — nullable; if set, must be an apiary the caller belongs to; `apiary_name` is populated via JOIN
- `lat`/`lng` — required, independent of `apiary_id`; not derived from an attached apiary
- `distance_km` — only present when the `GET /listings` request included a `near_lat`/`near_lng`/`radius_km` filter
- `images` — max 3 per listing
- `honey_batch_id` — nullable; may only be set when `category` is `HONEY`, must reference a batch owned by the caller with a confirmed on-chain certification, and a batch can only be attached to one listing at a time
- `honey_batch` — nested object with the attached batch's public details, `null`/absent when no batch is attached; `pdf_url` is only included when `has_pdf` is `true`. This enrichment is only present on `GET /listings/{id}`, not on `GET /listings` search/list results.

---

### POST /listings 🔒

Creates a listing owned by the authenticated user. The listing starts in `pending` status and is invisible to other users until an admin approves it. A user may have at most 20 listings.

**Request**
```json
{
  "title": "Wildflower honey, 500g jars",
  "description": "Raw, unfiltered, harvested this summer.",
  "category": "HONEY",
  "price": 25.00,
  "quantity": "10 jars",
  "address": "Warsaw, Poland",
  "apiary_id": 5,
  "lat": 52.23,
  "lng": 21.01,
  "contact_phone": "+48123456789",
  "contact_email": "seller@example.com",
  "image_urls": ["https://.../image.jpg"],
  "honey_batch_id": 12
}
```

- `title`, `category`, `lat`, and `lng` are required
- `image_urls` — optional array of strings, max 3
- `honey_batch_id` — optional; only valid when `category` is `HONEY`, must be a batch owned by the caller with a confirmed on-chain certification, and can only be attached to one listing at a time

**Response** `201 Created` — listing object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_BODY` | 400 | Malformed JSON |
| `TITLE_REQUIRED` | 400 | Title field is empty |
| `CATEGORY_INVALID` | 400 | Category not in allowed set |
| `TOO_MANY_IMAGES` | 400 | More than 3 `image_urls` given |
| `TITLE_TOO_LONG` | 400 | `title` exceeds 150 characters |
| `DESCRIPTION_TOO_LONG` | 400 | `description` exceeds 500 characters |
| `QUANTITY_TOO_LONG` | 400 | `quantity` exceeds 50 characters |
| `ADDRESS_TOO_LONG` | 400 | `address` exceeds 150 characters |
| `LOCATION_REQUIRED` | 400 | `lat`/`lng` missing |
| `INVALID_GPS` | 400 | `lat` not between -90 and 90, or `lng` not between -180 and 180 |
| `CONTACT_PHONE_TOO_LONG` | 400 | `contact_phone` exceeds 20 characters |
| `CONTACT_EMAIL_TOO_LONG` | 400 | `contact_email` exceeds 150 characters |
| `PRICE_TOO_LARGE` | 400 | `price` magnitude is >= 100,000,000 |
| `LISTING_LIMIT_REACHED` | 400 | Caller already has 20 listings |
| `APIARY_NOT_FOUND` | 404 | `apiary_id` set but caller is not a member of that apiary |
| `HONEY_BATCH_CATEGORY_MISMATCH` | 400 | `honey_batch_id` set but `category` is not `HONEY` |
| `HONEY_BATCH_NOT_FOUND` | 404 | `honey_batch_id` does not exist or is not owned by the caller |
| `HONEY_BATCH_NOT_CERTIFIED` | 400 | Batch does not have a confirmed on-chain certification |
| `HONEY_BATCH_ALREADY_ATTACHED` | 409 | Batch is already attached to another listing |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /listings

Searches/filters listings. Public — auth optional.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `category` | — | Filter by exact category |
| `keyword` | — | Matches title/description |
| `price_min` | — | Minimum price |
| `price_max` | — | Maximum price |
| `posted_after` | — | Only listings created after this timestamp |
| `near_lat` | — | Latitude to measure distance from; must be given together with `near_lng` and `radius_km` |
| `near_lng` | — | Longitude to measure distance from; must be given together with `near_lat` and `radius_km` |
| `radius_km` | — | Max distance in km from `near_lat`/`near_lng`, capped at 20,000; results are sorted nearest-first |
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |
| `mine` | false | If `true`, requires auth; returns only the caller's own listings, including hidden ones |

`near_lat`, `near_lng`, and `radius_km` only apply if all three are present and valid; otherwise the distance filter is silently ignored (not an error).

Non-owner and anonymous callers never see hidden listings.

**Response** `200 OK`
```json
{
  "items": [ /* listing objects */ ],
  "total": 1
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | `mine=true` given without a Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /listings/{id}

Returns a single listing. Public — auth optional. Hidden listings return `404` unless the caller is the owner.

**Response** `200 OK` — listing object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist, or is hidden and caller is not the owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /listings/{id} 🔒

Updates a listing. Only the owner can edit. Same body as create; when `image_urls` is provided, it replaces the existing images. If the listing was previously `approved`, editing resets it back to `pending` admin review.

**Request** — same shape as POST

**Response** `200 OK` — updated listing object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `TITLE_REQUIRED` | 400 | Title field is empty |
| `CATEGORY_INVALID` | 400 | Category not in allowed set |
| `TOO_MANY_IMAGES` | 400 | More than 3 `image_urls` given |
| `TITLE_TOO_LONG` | 400 | `title` exceeds 150 characters |
| `DESCRIPTION_TOO_LONG` | 400 | `description` exceeds 500 characters |
| `QUANTITY_TOO_LONG` | 400 | `quantity` exceeds 50 characters |
| `ADDRESS_TOO_LONG` | 400 | `address` exceeds 150 characters |
| `LOCATION_REQUIRED` | 400 | `lat`/`lng` missing |
| `INVALID_GPS` | 400 | `lat` not between -90 and 90, or `lng` not between -180 and 180 |
| `CONTACT_PHONE_TOO_LONG` | 400 | `contact_phone` exceeds 20 characters |
| `CONTACT_EMAIL_TOO_LONG` | 400 | `contact_email` exceeds 150 characters |
| `PRICE_TOO_LARGE` | 400 | `price` magnitude is >= 100,000,000 |
| `PHOTO_REQUIRED` | 400 | `image_urls` given as an empty array, removing all photos |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `APIARY_NOT_FOUND` | 404 | `apiary_id` set but caller is not a member of that apiary |
| `HONEY_BATCH_CATEGORY_MISMATCH` | 400 | `honey_batch_id` set but `category` is not `HONEY` |
| `HONEY_BATCH_NOT_FOUND` | 404 | `honey_batch_id` does not exist or is not owned by the caller |
| `HONEY_BATCH_NOT_CERTIFIED` | 400 | Batch does not have a confirmed on-chain certification |
| `HONEY_BATCH_ALREADY_ATTACHED` | 409 | Batch is already attached to a different listing |
| `NOT_OWNER` | 403 | Caller is not the listing owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### PATCH /listings/{id}/hide 🔒

Toggles listing visibility. Only the owner can hide/unhide.

**Request**
```json
{ "hidden": true }
```

**Response** `200 OK` — updated listing object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `NOT_OWNER` | 403 | Caller is not the listing owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /listings/{id} 🔒

Deletes a listing. Only the owner can delete. Images cascade.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `NOT_OWNER` | 403 | Caller is not the listing owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Listing Images

Images are stored on the server under a Docker volume. Accepted MIME types: `image/jpeg`, `image/png`, `image/webp`. Maximum file size: **5 MB**. Maximum **3 images per listing**.

Images are cascade-deleted when the parent listing is deleted.

**Image object**
```json
{
  "id": 1,
  "listing_id": 1,
  "url": "/api/v1/listings/1/images/1/file",
  "display_order": 0,
  "created_at": "2026-07-01T10:00:00Z"
}
```

- `url` — computed path to `GET /listings/{id}/images/{imageId}/file`, not stored directly

---

### POST /listings/{id}/images 🔒

Uploads an image. Send as `multipart/form-data` with field name `image`. Only the listing owner can upload.

**Response** `201 Created` — image object

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `MISSING_FILE` | 400 | `image` field missing from form |
| `INVALID_IMAGE_TYPE` | 400 | MIME type not allowed |
| `IMAGE_TOO_LARGE` | 413 | File exceeds 5 MB |
| `TOO_MANY_IMAGES` | 400 | Listing already has 3 images |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `NOT_OWNER` | 403 | Caller is not the listing owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /listings/{id}/images/{imageId}/file

Serves the raw image bytes with the correct `Content-Type` header. Public — hidden-listing rules do not apply. Cached for 24 hours (`Cache-Control: public, max-age=86400`).

**Response** `200 OK` — image binary

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` or `{imageId}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `IMAGE_NOT_FOUND` | 404 | Image does not exist for this listing |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /listings/{id}/images/{imageId} 🔒

Deletes an image from the DB and from disk. Only the listing owner can delete.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` or `{imageId}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `IMAGE_NOT_FOUND` | 404 | Image does not exist for this listing |
| `NOT_OWNER` | 403 | Caller is not the listing owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Listing Favorites

Lets a user save listings for later. A favorite is a per-user, per-listing bookmark.

---

### POST /listings/{id}/favorite 🔒

Saves the listing to the caller's favorites. Idempotent — favoriting an already-favorited listing succeeds with no change.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist, or is hidden and caller is not the owner |
| `CANNOT_FAVORITE_OWN_LISTING` | 403 | Caller is the listing's owner |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### DELETE /listings/{id}/favorite 🔒

Removes the listing from the caller's favorites. No-op if it wasn't favorited.

**Response** `204 No Content`

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /listings/{id}/favorite 🔒

Reports whether the caller has favorited the listing.

**Response** `200 OK`
```json
{
  "is_favorite": true
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Admin

All routes below require both a valid JWT (`Authorization: Bearer`) and the caller's user having `role = "admin"` in the database (set manually via SQL — there is no self-service promotion path).

| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token in header |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `NOT_ADMIN` | 403 | Caller is authenticated but not an admin |

### Admin: Marketplace Listing Moderation

Admin listing objects have the same shape as the public listing object (see [Listing object](#listing-object)), plus two extra fields:

```json
{
  "...": "...",
  "status": "pending",
  "rejection_reason": null,
  "is_edit": false
}
```

- `status` — `"pending"`, `"approved"`, or `"rejected"`
- `rejection_reason` — nullable string, set when `status` is `"rejected"`
- `is_edit` — `true` if this is an edit of a previously-approved listing (as opposed to a brand-new one); both sit at `status: "pending"` otherwise

---

### GET /admin/listings

Returns pending marketplace listings (new and edited), ordered oldest-first.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |

**Response** `200 OK`
```json
{
  "items": [ /* admin listing objects */ ],
  "total": 3
}
```

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /admin/listings/{id}

Returns a single listing regardless of status.

**Response** `200 OK` — admin listing object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /admin/listings/{id}/approve

Approves a pending listing, making it publicly visible. The listing must have at least one photo.

**Response** `200 OK` — updated admin listing object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `LISTING_NOT_PENDING` | 409 | Listing is not currently pending review |
| `PHOTO_REQUIRED` | 400 | Listing has no photos |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /admin/listings/{id}/reject

Rejects a pending listing with a reason. The listing stays invisible to the public.

**Request**
```json
{ "reason": "Photos don't match the description" }
```

**Response** `200 OK` — updated admin listing object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `REJECTION_REASON_REQUIRED` | 400 | `reason` is empty |
| `LISTING_NOT_FOUND` | 404 | Listing does not exist |
| `LISTING_NOT_PENDING` | 409 | Listing is not currently pending review |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### Admin: Honey Batch Certification Review

### Certification request object

```json
{
  "id": 1,
  "batch_id": 5,
  "requested_by": 3,
  "requester_email": "beekeeper@example.com",
  "status": "pending",
  "rejection_reason": null,
  "blockchain_job_id": null,
  "created_at": "2026-07-01T10:00:00Z",
  "gathering_date": "2026-06-15T00:00:00Z",
  "amount_grams": 2500,
  "honey_type": "Wildflower",
  "pdf_url": "/api/v1/admin/honey-batches/5/pdf"
}
```

- `status` — `"pending"`, `"approved"`, or `"rejected"`
- `blockchain_job_id` — nullable; set once approval enqueues the `blockchain_jobs` row the Epic 9 worker picks up
- `requester_email`, `gathering_date`, `amount_grams`, `honey_type` — joined from the batch and requesting user, for display in the admin queue/detail views
- `pdf_url` — convenience field pointing at `GET /api/v1/admin/honey-batches/{id}/pdf`

---

### GET /admin/certification-requests

Returns pending honey batch certification requests, ordered oldest-first.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |

**Response** `200 OK`
```json
{
  "items": [ /* certification request objects */ ],
  "total": 2
}
```

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /admin/certification-requests/{id}

Returns a single certification request.

**Response** `200 OK` — certification request object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `CERTIFICATION_REQUEST_NOT_FOUND` | 404 | Certification request does not exist |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /admin/certification-requests/{id}/approve

Approves a pending certification request. This is what actually creates the `blockchain_jobs` row the existing Epic 9 worker picks up — worker, idempotency, and retry logic are unchanged.

**Response** `200 OK` — updated certification request object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `CERTIFICATION_REQUEST_NOT_FOUND` | 404 | Certification request does not exist |
| `CERTIFICATION_REQUEST_NOT_PENDING` | 409 | Request is not currently pending review |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### POST /admin/certification-requests/{id}/reject

Rejects a pending certification request with a reason. No blockchain job is enqueued.

**Request**
```json
{ "reason": "Lab PDF is unreadable" }
```

**Response** `200 OK` — updated certification request object

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `INVALID_BODY` | 400 | Malformed JSON |
| `REJECTION_REASON_REQUIRED` | 400 | `reason` is empty |
| `CERTIFICATION_REQUEST_NOT_FOUND` | 404 | Certification request does not exist |
| `CERTIFICATION_REQUEST_NOT_PENDING` | 409 | Request is not currently pending review |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

### GET /admin/honey-batches/{id}/pdf

Serves a batch's lab PDF regardless of ownership, for admin review.

**Response** `200 OK` — `application/pdf` binary

**Errors** — see [Admin](#admin) header errors, plus:
| Code | Status | Description |
|------|--------|-------------|
| `INVALID_ID` | 400 | Path `{id}` is not a valid integer |
| `BATCH_NOT_FOUND` | 404 | Batch does not exist |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

### GET /favorites 🔒

Returns the caller's favorited listings, most recently favorited first. Hidden listings are excluded unless the caller owns them.

**Response** `200 OK`
```json
{
  "items": [ /* listing objects */ ],
  "total": 1
}
```

**Errors**
| Code | Status | Description |
|------|--------|-------------|
| `MISSING_TOKEN` | 401 | No Bearer token |
| `INVALID_TOKEN` | 401 | Token invalid or expired |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
