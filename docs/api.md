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
- `grid_rows` and `grid_cols` must be ≥ 1

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
| `INVALID_GRID_SIZE` | 400 | grid_rows or grid_cols < 1 |
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
- `active`, `queenless`, `ready_for_harvest` default to `false` if omitted
- `grid_row` and `grid_col` are 0-indexed and must fall within the apiary's `grid_rows` × `grid_cols` bounds
- Each position within an apiary must be unique

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
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `INVALID_GRID_POSITION` | 400 | Position is outside apiary grid bounds |
| `POSITION_OCCUPIED` | 409 | Another hive already occupies that position |
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

Updates a hive's name, type, and active status. Both owners and members can edit hives.

**Request**
```json
{
  "name": "Renamed Hive",
  "type": "top_bar",
  "active": false,
  "queenless": true,
  "ready_for_harvest": false,
  "frames": 12
}
```

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
| `APIARY_NOT_FOUND` | 404 | Apiary does not exist or user is not a member |
| `HIVE_NOT_FOUND` | 404 | Hive does not exist in this apiary |
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
  "frames_honey": 4,
  "frames_pollen": 2,
  "queen_cells_count": 0,
  "aggressiveness": "calm",
  "frames_added_foundation": 1,
  "frames_added_drawn": null,
  "frames_added_honey": null,
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
  "frames_honey": 4,
  "frames_pollen": 2,
  "queen_cells_count": 0,
  "aggressiveness": "calm",
  "frames_added_foundation": 1,
  "frames_added_drawn": null,
  "frames_added_honey": null,
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

Images are stored on the server under a Docker volume. Accepted MIME types: `image/jpeg`, `image/png`, `image/webp`. Maximum file size: **10 MB**.

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
| `IMAGE_TOO_LARGE` | 413 | File exceeds 10 MB |
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
