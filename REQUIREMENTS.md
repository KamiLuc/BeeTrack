# BeeTrack — Product Requirements & Backlog

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

---

## 1. Auth & User Management

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| AU-01 | `[ ]`  | User registration (email + password) | Hash passwords with bcrypt |
| AU-02 | `[ ]`  | User login — JWT-based auth | Access token + refresh token |
| AU-03 | `[ ]`  | Refresh token rotation | Store refresh tokens in DB, invalidate on use |
| AU-04 | `[ ]`  | Logout / token revocation | |

---

## 2. Apiary & Hive Management

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| HV-01 | `[ ]`  | Create apiary (name, location, GPS coords, grid size rows×cols) | Creator becomes apiary owner |
| HV-02 | `[ ]`  | Edit / delete apiary | |
| HV-03 | `[ ]`  | List all apiaries | |
| HV-04 | `[ ]`  | Apiary grid view — display hives on their grid positions | Empty cells shown as placeholders |
| HV-05 | `[ ]`  | Add hive to apiary (name, type, install date, grid position) | Position = (row, col); must be within grid bounds and unoccupied |
| HV-06 | `[ ]`  | Move hive to a different grid position | Drag-and-drop or manual coordinate input |
| HV-07 | `[ ]`  | Rename hive | |
| HV-08 | `[ ]`  | Edit other hive fields / delete hive | |
| HV-09 | `[ ]`  | Hive detail screen | Shows hive info + linked inspections |
| HV-10 | `[ ]`  | Hive status field (active, inactive, dead-out, sold) | |
| HV-11 | `[ ]`  | Invite another user to an apiary (by email) | Invited user gets member role |
| HV-12 | `[ ]`  | Apiary roles: owner vs member | Owner can invite/remove members and delete apiary; members can manage hives and inspections |
| HV-13 | `[ ]`  | List members of an apiary | |
| HV-14 | `[ ]`  | Remove a member from apiary / leave apiary | Owner cannot leave without transferring ownership |

---

## 3. Inspection Logging

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| IN-01 | `[ ]`  | Create inspection record for a hive | Date, duration, weather |
| IN-02 | `[ ]`  | Record queen status (seen, not seen, capped cells, eggs) | Enum + free-text |
| IN-03 | `[ ]`  | Record brood pattern (excellent / good / poor / none) | |
| IN-04 | `[ ]`  | Record frames of bees, honey, pollen (counts) | |
| IN-05 | `[ ]`  | Record Varroa mite count result | Numeric field + method |
| IN-06 | `[ ]`  | Free-text notes per inspection | |
| IN-07 | `[ ]`  | Edit / delete inspection | |
| IN-08 | `[ ]`  | Inspection history list per hive (paginated) | |

---

## 4. Health & Treatment Tracking

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| TR-01 | `[ ]`  | Log treatment (type, product, dosage, date applied, date ended) | e.g., Apivar, oxalic acid, Apiguard |
| TR-02 | `[ ]`  | Mark treatment complete | |
| TR-03 | `[ ]`  | Treatment history per hive | |

---

## 5. Honey Harvest Tracking

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| HH-01 | `[ ]`  | Log harvest (date, hive, frames extracted, kg/litres yield) | |
| HH-02 | `[ ]`  | Edit / delete harvest entry | |

---

## 6. Reports & Analytics

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| RP-01 | `[ ]`  | Dashboard — overview of all hives at a glance | Hive count, last inspection dates, active treatments |

---

## 7. Queen Recognition (AI Feature)

> **Deferred — implement after core app is stable.**

| ID    | Status | Title | Notes |
|-------|--------|-------|-------|
| QR-01 | `[ ]`  | Camera capture screen in Flutter | Live preview + capture button |
| QR-02 | `[ ]`  | Upload frame photo to backend for analysis | Multipart form POST |
| QR-03 | `[ ]`  | Backend: CV model inference endpoint | Go service calls Python/ONNX/TFLite model or external microservice |
| QR-04 | `[ ]`  | Return bounding box + confidence score to app | |
| QR-05 | `[ ]`  | Overlay bounding box on captured image in Flutter | |
| QR-06 | `[ ]`  | Collect user feedback (correct / incorrect) for model retraining | |
| QR-07 | `[ ]`  | Model training pipeline (dataset, annotations, training script) | Thesis core — YOLO / EfficientDet / custom CNN |
| QR-08 | `[ ]`  | Model evaluation metrics (mAP, precision, recall) | Document for thesis |

---

## 8. MCP Server

> **Deferred — enables AI voice assistant integration.**

| ID     | Status | Title | Notes |
|--------|--------|-------|-------|
| MCP-01 | `[ ]`  | MCP server endpoint in Go backend (HTTP+SSE transport) | Runs alongside REST API |
| MCP-02 | `[ ]`  | Tool: `create_inspection` | |
| MCP-03 | `[ ]`  | Tool: `log_treatment` | |
| MCP-04 | `[ ]`  | Tool: `log_harvest` | |
| MCP-05 | `[ ]`  | Tool: `get_hive_summary` | Returns latest inspection + active treatments |
| MCP-06 | `[ ]`  | Tool: `list_hives` | |
| MCP-07 | `[ ]`  | Auth for MCP clients (API key or OAuth) | |
| MCP-08 | `[ ]`  | Voice pipeline integration (STT → LLM + MCP → backend) | e.g., Whisper → Claude/GPT with MCP tools |

---

## 9. Infrastructure & DevOps

| ID     | Status | Title | Notes |
|--------|--------|-------|-------|
| INF-01 | `[x]`  | Docker Compose setup (Go API + PostgreSQL) | |
| INF-02 | `[x]`  | Database schema & migrations (golang-migrate or goose) | Using goose |
| INF-03 | `[x]`  | Environment config via `.env` | |
| INF-04 | `[x]`  | Go project structure (cmd/, internal/, pkg/) | |
| INF-05 | `[ ]`  | REST API — OpenAPI / Swagger spec | |
| INF-06 | `[ ]`  | Input validation & structured error responses | |
| INF-07 | `[ ]`  | Structured JSON logging | |
| INF-08 | `[ ]`  | CORS configuration for web client | |

---

## Data Model (Draft)

```
User
  id, email, password_hash, name, created_at

Apiary
  id, owner_user_id, name, location_name, lat, lng, grid_rows, grid_cols, created_at

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
