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

1. [Apiary & Hive Management](#1-apiary--hive-management)
2. [Health & Treatment Tracking](#2-health--treatment-tracking)
3. [Honey Harvest Tracking](#3-honey-harvest-tracking)
4. [Reports & Analytics](#4-reports--analytics)
5. [Queen Recognition (AI Feature)](#5-queen-recognition-ai-feature)
6. [MCP Server](#6-mcp-server)
7. [Infrastructure & DevOps](#7-infrastructure--devops)
8. [Localization](#8-localization)

---

## 1. Apiary & Hive Management

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| HV-11-BE | `BE` | `[ ]` | Invite user to apiary endpoint | Invited user gets member role |
| HV-11-FE | `FE` | `[ ]` | Invite user UI | Input email, send invite |
| HV-12-BE | `BE` | `[ ]` | Apiary roles enforcement | Owner can invite/remove/delete; member can manage hives and inspections |
| HV-13-BE | `BE` | `[ ]` | List apiary members endpoint | |
| HV-13-FE | `FE` | `[ ]` | Members list screen | |
| HV-14-BE | `BE` | `[ ]` | Remove member / leave apiary endpoint | Owner cannot leave without transferring ownership |
| HV-14-FE | `FE` | `[ ]` | Remove member / leave apiary UI | |

---

## 2. Health & Treatment Tracking

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| TR-01-BE | `BE` | `[ ]` | Log treatment endpoint | Type, product, dosage, date applied, date ended |
| TR-01-FE | `FE` | `[ ]` | Log treatment screen | e.g., Apivar, oxalic acid, Apiguard |
| TR-02-BE | `BE` | `[ ]` | Mark treatment complete endpoint | |
| TR-02-FE | `FE` | `[ ]` | Mark treatment complete UI | |
| TR-03-BE | `BE` | `[ ]` | Treatment history endpoint | |
| TR-03-FE | `FE` | `[ ]` | Treatment history screen | |

---

## 3. Honey Harvest Tracking

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| HH-01-BE | `BE` | `[ ]` | Log harvest endpoint | Date, hive, frames extracted, kg/litres yield |
| HH-01-FE | `FE` | `[ ]` | Log harvest screen | |
| HH-02-BE | `BE` | `[ ]` | Edit / delete harvest endpoint | |
| HH-02-FE | `FE` | `[ ]` | Edit / delete harvest UI | |

---

## 4. Reports & Analytics

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| RP-01-BE | `BE` | `[ ]` | Dashboard data endpoint | Hive count, last inspection dates, active treatments |
| RP-01-FE | `FE` | `[ ]` | Dashboard screen | Overview of all hives at a glance |

---

## 5. Queen Recognition (AI Feature)

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

## 6. MCP Server

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

## 7. Infrastructure & DevOps

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| INF-05-BE | `BE` | `[ ]` | REST API — OpenAPI / Swagger spec | |
| INF-06-BE | `BE` | `[ ]` | Input validation & structured error responses | |
| INF-07-BE | `BE` | `[ ]` | Structured JSON logging | |

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
  id, apiary_id, name, type, active, queenless, ready_for_harvest, frames,
  grid_row, grid_col, notes

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

- [ ] Offline-first mode on mobile (local SQLite sync)?
