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

1. [UX Polish](#1-ux-polish)
2. [Honey Harvest Tracking](#2-honey-harvest-tracking)
3. [Reports & Analytics](#3-reports--analytics)
4. [Bulk Operations](#4-bulk-operations)
5. [Voice Logging](#5-voice-logging)
6. [Queen Recognition (AI Feature)](#6-queen-recognition-ai-feature)
7. [MCP Server](#7-mcp-server)
8. [Infrastructure & DevOps](#8-infrastructure--devops)

---

## 1. UX Polish

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| UX-01-FE | `FE` | `[x]` | Apiary copy — name picker modal | Instead of auto-suffixing, show a modal with a pre-filled name the user can edit before confirming |

---

## 2. Honey Harvest Tracking

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| HH-01-BE | `BE` | `[ ]` | Log harvest endpoint | Date, hive, frames extracted, kg/litres yield |
| HH-01-FE | `FE` | `[ ]` | Log harvest screen | |
| HH-02-BE | `BE` | `[ ]` | Edit / delete harvest endpoint | |
| HH-02-FE | `FE` | `[ ]` | Edit / delete harvest UI | |

---

## 3. Reports & Analytics

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| RP-01-BE | `BE` | `[ ]` | Dashboard data endpoint | Hive count, last inspection dates, active treatments |
| RP-01-FE | `FE` | `[ ]` | Dashboard screen | Overview of all hives at a glance |

---

## 4. Bulk Operations

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| BK-01-BE | `BE` | `[x]` | Bulk treatment endpoint | POST /api/v1/apiaries/{id}/treatments/bulk — same body as single treatment, inserts one record per hive in a transaction |
| BK-01-FE | `FE` | `[x]` | "Treat all hives" from apiary view | 3-dots menu option; reuses treatment form; shows "Treatment logged for N hives" snackbar on success |

---

## 5. Voice Logging

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| VC-01-BE | `BE` | `[ ]` | Voice endpoint | POST /api/v1/hives/{hiveId}/voice — accepts audio file, calls Whisper → Claude, dispatches to correct service |
| VC-02-BE | `BE` | `[ ]` | Claude intent parser | Given transcript + hive context, returns structured action (log_inspection / log_treatment / log_harvest) with fields filled |
| VC-03-FE | `FE` | `[ ]` | Hold-to-record mic button on hive detail screen | Uses `record` package; sends audio to VC-01 on release; saves immediately, no confirmation step |
| VC-04-FE | `FE` | `[ ]` | Result snackbar | Show what was saved ("Inspection logged: queen added, brood good") so user knows what was recorded; tap to edit if wrong |

---

## 6. Queen Recognition (AI Feature)

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

## 7. MCP Server

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

## 8. Infrastructure & DevOps

| ID | Layer | Status | Title | Notes |
|----|-------|--------|-------|-------|
| INF-05-BE | `BE` | `[ ]` | REST API — OpenAPI / Swagger spec | |
| INF-06-BE | `BE` | `[ ]` | Input validation & structured error responses | |
| INF-07-BE | `BE` | `[ ]` | Structured JSON logging | |

