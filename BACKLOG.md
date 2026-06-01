# BeeTrack — Backlog

> Living document. Treat each item like a Jira ticket — update status as work progresses.
> **Stack:** Flutter (Android + Web) · Go (backend API) · PostgreSQL · Docker

---

## Status Legend

| Symbol | Meaning |
|--------|---------|
| `[ ]`  | To do |
| `[~]`  | In progress |
| `[x]`  | Done |
| `[!]`  | Blocked |

---

## Epics Overview

1. [Auth & User Management](#1-auth--user-management)
2. [Apiary & Hive Management](#2-apiary--hive-management)
3. [Inspection Logging](#3-inspection-logging)
4. [Health & Treatment Tracking](#4-health--treatment-tracking)
5. [Honey Harvest Tracking](#5-honey-harvest-tracking)
6. [Reports & Analytics](#6-reports--analytics)
7. [Queen Recognition (AI Feature)](#7-queen-recognition-ai-feature)
8. [MCP Server](#8-mcp-server)
9. [Infrastructure & DevOps](#9-infrastructure--devops)
10. [Localization](#10-localization)

---

## 1. Auth & User Management

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| AU-01-BE | `BE` | `[x]` | User registration endpoint | Hash passwords with bcrypt |
| AU-01-FE | `FE` | `[x]` | Registration screen | Email, name, password fields |
| AU-02-BE | `BE` | `[x]` | Login endpoint — JWT-based auth | Access token + refresh token |
| AU-02-FE | `FE` | `[ ]` | Login screen + token interceptor | Store tokens; auto-refresh on 401 |
| AU-03-BE | `BE` | `[x]` | Refresh token rotation | Invalidate on use |
| AU-04-BE | `BE` | `[x]` | Logout endpoint | Delete refresh token from DB |
| AU-04-FE | `FE` | `[ ]` | Logout action | Clear stored tokens, redirect to login |
| AU-05-BE | `BE` | `[x]` | Update display name endpoint | PATCH /api/v1/users/me/name |
| AU-05-FE | `FE` | `[ ]` | Edit display name screen | Name defaults to email on registration; user changes it here |

---

## 2. Apiary & Hive Management

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| HV-01-BE | `BE` | `[x]` | Create apiary endpoint | POST /api/v1/apiaries; creator becomes owner |
| HV-01-FE | `FE` | `[ ]` | Create apiary screen | Form: name, location, GPS coords, grid size |
| HV-02-BE | `BE` | `[x]` | Edit / delete apiary endpoints | |
| HV-02-FE | `FE` | `[ ]` | Edit / delete apiary UI | |
| HV-03-BE | `BE` | `[ ]` | List apiaries endpoint | GET /api/v1/apiaries |
| HV-03-FE | `FE` | `[ ]` | Apiaries list screen | |
| HV-04-BE | `BE` | `[x]` | Apiary hive positions endpoint | Returns hives with grid coords |
| HV-04-FE | `FE` | `[ ]` | Apiary grid view | Display hives on grid; empty cells as placeholders |
| HV-05-BE | `BE` | `[x]` | Add hive to apiary endpoint | Validate grid position within bounds and unoccupied |
| HV-05-FE | `FE` | `[ ]` | Add hive screen | Name, type, install date, grid position picker |
| HV-06-BE | `BE` | `[x]` | Move hive endpoint | Update grid position |
| HV-06-FE | `FE` | `[ ]` | Move hive UI | Drag-and-drop or coordinate input |
| HV-07-BE | `BE` | `[ ]` | Rename hive endpoint | |
| HV-07-FE | `FE` | `[ ]` | Rename hive UI | |
| HV-08-BE | `BE` | `[x]` | Edit / delete hive endpoints | |
| HV-08-FE | `FE` | `[ ]` | Edit / delete hive UI | |
| HV-09-BE | `BE` | `[x]` | Hive detail endpoint | Returns hive info + latest inspection |
| HV-09-FE | `FE` | `[ ]` | Hive detail screen | |
| HV-10-BE | `BE` | `[ ]` | Hive status field | active, inactive, dead-out, sold |
| HV-10-FE | `FE` | `[ ]` | Hive status UI | Dropdown / selector |
| HV-11-BE | `BE` | `[ ]` | Invite user to apiary endpoint | Invited user gets member role |
| HV-11-FE | `FE` | `[ ]` | Invite user UI | Input email, send invite |
| HV-12-BE | `BE` | `[ ]` | Apiary roles enforcement | Owner can invite/remove/delete; member can manage hives and inspections |
| HV-13-BE | `BE` | `[ ]` | List apiary members endpoint | |
| HV-13-FE | `FE` | `[ ]` | Members list screen | |
| HV-14-BE | `BE` | `[ ]` | Remove member / leave apiary endpoint | Owner cannot leave without transferring ownership |
| HV-14-FE | `FE` | `[ ]` | Remove member / leave apiary UI | |

---

## 3. Inspection Logging

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| IN-01-BE | `BE` | `[ ]` | Create inspection endpoint | Date, duration, weather |
| IN-01-FE | `FE` | `[ ]` | Create inspection screen | |
| IN-02-BE | `BE` | `[ ]` | Queen status field | Enum: seen, not seen, capped cells, eggs |
| IN-02-FE | `FE` | `[ ]` | Queen status UI | Selector + free-text |
| IN-03-BE | `BE` | `[ ]` | Brood pattern field | Enum: excellent, good, poor, none |
| IN-03-FE | `FE` | `[ ]` | Brood pattern UI | |
| IN-04-BE | `BE` | `[ ]` | Frames count fields | Bees, honey, pollen |
| IN-04-FE | `FE` | `[ ]` | Frames count UI | Numeric inputs |
| IN-05-BE | `BE` | `[ ]` | Varroa mite count field | Numeric + method |
| IN-05-FE | `FE` | `[ ]` | Varroa count UI | |
| IN-06-BE | `BE` | `[ ]` | Free-text notes field | |
| IN-06-FE | `FE` | `[ ]` | Notes UI | Text area |
| IN-07-BE | `BE` | `[ ]` | Edit / delete inspection endpoints | |
| IN-07-FE | `FE` | `[ ]` | Edit / delete inspection UI | |
| IN-08-BE | `BE` | `[ ]` | Inspection history endpoint | Paginated |
| IN-08-FE | `FE` | `[ ]` | Inspection history list screen | |

---

## 4. Health & Treatment Tracking

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| TR-01-BE | `BE` | `[ ]` | Log treatment endpoint | Type, product, dosage, date applied, date ended |
| TR-01-FE | `FE` | `[ ]` | Log treatment screen | e.g., Apivar, oxalic acid, Apiguard |
| TR-02-BE | `BE` | `[ ]` | Mark treatment complete endpoint | |
| TR-02-FE | `FE` | `[ ]` | Mark treatment complete UI | |
| TR-03-BE | `BE` | `[ ]` | Treatment history endpoint | |
| TR-03-FE | `FE` | `[ ]` | Treatment history screen | |

---

## 5. Honey Harvest Tracking

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| HH-01-BE | `BE` | `[ ]` | Log harvest endpoint | Date, hive, frames extracted, kg/litres yield |
| HH-01-FE | `FE` | `[ ]` | Log harvest screen | |
| HH-02-BE | `BE` | `[ ]` | Edit / delete harvest endpoint | |
| HH-02-FE | `FE` | `[ ]` | Edit / delete harvest UI | |

---

## 6. Reports & Analytics

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| RP-01-BE | `BE` | `[ ]` | Dashboard data endpoint | Hive count, last inspection dates, active treatments |
| RP-01-FE | `FE` | `[ ]` | Dashboard screen | Overview of all hives at a glance |

---

## 7. Queen Recognition (AI Feature)

> **Deferred — implement after core app is stable.**

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| QR-01-FE | `FE` | `[ ]` | Camera capture screen | Live preview + capture button |
| QR-02-BE | `BE` | `[ ]` | Upload photo endpoint | Multipart form POST |
| QR-02-FE | `FE` | `[ ]` | Upload photo from app | |
| QR-03-BE | `BE` | `[ ]` | CV model inference endpoint | Go service calls Python/ONNX/TFLite model |
| QR-04-BE | `BE` | `[ ]` | Return bounding box + confidence score | |
| QR-05-FE | `FE` | `[ ]` | Overlay bounding box on image | |
| QR-06-BE | `BE` | `[ ]` | Collect user feedback endpoint | correct / incorrect for retraining |
| QR-06-FE | `FE` | `[ ]` | User feedback UI | |
| QR-07-BE | `BE` | `[ ]` | Model training pipeline | Dataset, annotations, training script — thesis core |
| QR-08-BE | `BE` | `[ ]` | Model evaluation metrics | mAP, precision, recall — document for thesis |

---

## 8. MCP Server

> **Deferred — enables AI voice assistant integration.**

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| MCP-01-BE | `BE` | `[ ]` | MCP server endpoint | HTTP+SSE transport, runs alongside REST API |
| MCP-02-BE | `BE` | `[ ]` | Tool: `create_inspection` | |
| MCP-03-BE | `BE` | `[ ]` | Tool: `log_treatment` | |
| MCP-04-BE | `BE` | `[ ]` | Tool: `log_harvest` | |
| MCP-05-BE | `BE` | `[ ]` | Tool: `get_hive_summary` | Latest inspection + active treatments |
| MCP-06-BE | `BE` | `[ ]` | Tool: `list_hives` | |
| MCP-07-BE | `BE` | `[ ]` | Auth for MCP clients | API key or OAuth |
| MCP-08-BE | `BE` | `[ ]` | Voice pipeline integration | Whisper → Claude/GPT with MCP tools |

---

## 9. Infrastructure & DevOps

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| INF-01-BE | `BE` | `[x]` | Docker Compose setup | Go API + PostgreSQL |
| INF-02-BE | `BE` | `[x]` | Database schema & migrations | Using goose |
| INF-03-BE | `BE` | `[x]` | Environment config via `.env` | |
| INF-04-BE | `BE` | `[x]` | Go project structure | cmd/, internal/, pkg/ |
| INF-05-BE | `BE` | `[ ]` | REST API — OpenAPI / Swagger spec | |
| INF-06-BE | `BE` | `[ ]` | Input validation & structured error responses | |
| INF-07-BE | `BE` | `[ ]` | Structured JSON logging | |
| INF-08-BE | `BE` | `[x]` | CORS configuration for web client | CORS_ALLOWED_ORIGINS env var, defaults to * |

---

## 10. Localization

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| LO-01-BE | `BE` | `[x]` | Structured error codes in API responses | `{ code, message }` — frontend uses code for translation |
| LO-02-FE | `FE` | `[ ]` | Flutter i18n — English + Polish | ARB files, flutter_localizations |

---

## Data Model (Draft)

```
User
  id, email, password_hash, name, created_at

Apiary
  id, owner_user_id, name, lat, lng, grid_rows, grid_cols, created_at, updated_at

ApiaryMember
  apiary_id, user_id, role (owner|member), joined_at

Hive
  id, apiary_id, name, type, status, install_date, grid_row, grid_col, notes

Inspection
  id, hive_id, inspected_at, duration_min, weather,
  queen_status, brood_pattern, frames_bees, frames_honey,
  frames_pollen, varroa_count, varroa_method, notes

Treatment
  id, hive_id, product, type, dosage, started_at, ends_at, completed

Harvest
  id, hive_id, harvested_at, frames_extracted, yield_kg, notes
```

---

## Open Questions

- [x] Multi-user / shared apiary support — covered by HV-11 to HV-14
- [ ] Offline-first mode on mobile (local SQLite sync)?
