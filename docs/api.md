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
  "grid_row": 0,
  "grid_col": 0
}
```

- `type` is optional — defaults to `"langstroth"`
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
  "ready_for_harvest": false
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

Returns inspections for the hive ordered by `inspected_at` descending.

**Query parameters**
| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | 20 | Maximum number of records to return |
| `offset` | 0 | Number of records to skip |

**Response** `200 OK` — array of inspection objects

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
