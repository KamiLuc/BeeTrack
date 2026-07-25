#!/usr/bin/env python3
"""Generates docs/openapi.yaml from docs/api.md.

docs/api.md is the single source of truth for the BeeTrack API surface. This
script parses its consistent Markdown structure (### METHOD /path headings,
**Request**/**Response**/**Errors** sections) into an OpenAPI 3.0 document,
so the spec can never silently drift from the prose docs the way it drifted
from the actual Go routes before this script existed (see INF-05-BE).

Regenerate after editing docs/api.md:
    python docs/generate_openapi.py

This is a documentation-generation tool, not a full CommonMark parser — it
relies on api.md's headings/tables/fenced-code-blocks being formatted the
way every existing entry in this file already is. If you add a wildly
different shape of section, check the generated output.
"""
import json
import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent
API_MD = ROOT / "api.md"
OUT = ROOT / "openapi.yaml"

HTTP_METHODS = {"GET", "POST", "PATCH", "DELETE", "PUT"}

ERROR_SCHEMA_REF = {"$ref": "#/components/schemas/Error"}

ADMIN_BASE_ERRORS = [
    ("MISSING_TOKEN", 401, "No Bearer token in header"),
    ("INVALID_TOKEN", 401, "Token invalid or expired"),
    ("NOT_ADMIN", 403, "Caller is authenticated but not an admin"),
]


def infer_schema(value):
    """Builds a minimal JSON Schema fragment describing value, for use as
    both `schema` and `example` in the generated spec."""
    if value is None:
        return {"type": "string", "nullable": True}
    if isinstance(value, bool):
        return {"type": "boolean"}
    if isinstance(value, int):
        return {"type": "integer"}
    if isinstance(value, float):
        return {"type": "number"}
    if isinstance(value, str):
        return {"type": "string"}
    if isinstance(value, list):
        if not value:
            return {"type": "array", "items": {}}
        return {"type": "array", "items": infer_schema(value[0])}
    if isinstance(value, dict):
        props = {k: infer_schema(v) for k, v in value.items()}
        return {"type": "object", "properties": props}
    return {}


def split_sections(text):
    """Yields (tag, section_body) for each '## Heading' block, in order."""
    parts = re.split(r"^## (.+)$", text, flags=re.MULTILINE)
    # parts[0] is preamble before the first ##; then alternating heading/body.
    for i in range(1, len(parts), 2):
        yield parts[i].strip(), parts[i + 1]


def split_endpoints(section_body):
    """Yields (heading_line, block_body) for each '### ...' block."""
    parts = re.split(r"^### (.+)$", section_body, flags=re.MULTILINE)
    for i in range(1, len(parts), 2):
        yield parts[i].strip(), parts[i + 1]


def parse_heading(heading):
    m = re.match(r"^(GET|POST|PATCH|DELETE|PUT)\s+(\S+)", heading)
    if not m:
        return None
    method, path = m.group(1), m.group(2)
    locked = "🔒" in heading
    return method, path, locked


def extract_query_params(body):
    m = re.search(
        r"\*\*Query [Pp]arameters\*\*\s*\n\|.*\|\s*\n\|[-\s|]+\|\s*\n((?:\|.*\|\s*\n?)+)",
        body,
    )
    params = []
    if not m:
        return params
    for line in m.group(1).strip().splitlines():
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        if len(cells) < 2:
            continue
        name = cells[0].strip("`")
        default = cells[1] if len(cells) > 2 else None
        schema_type = (
            "integer" if name in ("limit", "offset") else "string"
        )
        params.append(
            {
                "name": name,
                "in": "query",
                "required": False,
                "schema": {"type": schema_type},
                "description": cells[-1],
            }
        )
    return params


def extract_json_block(text):
    m = re.search(r"```json\s*\n(.*?)\n```", text, re.DOTALL)
    if not m:
        return None
    try:
        return json.loads(m.group(1))
    except json.JSONDecodeError:
        return None


def extract_request_body(body):
    m = re.search(
        r"\*\*Request(?: [Bb]ody)?\*\*[^\n]*\n(.*?)(?=\n\*\*|\n---|\Z)",
        body,
        re.DOTALL,
    )
    if not m:
        return None
    example = extract_json_block(m.group(1))
    if example is None:
        return None
    return {
        "content": {
            "application/json": {
                "schema": infer_schema(example),
                "example": example,
            }
        }
    }


def extract_responses(body, inherited_errors, last_full_errors):
    responses = {}

    # Success/primary response lines: "**Response** `200 OK` — optional note"
    for m in re.finditer(
        r"\*\*(?:Response|Error response)\*\*\s*`(\d{3})[^`]*`([^\n]*)\n?(.*?)(?=\n\*\*|\n---|\Z)",
        body,
        re.DOTALL,
    ):
        status, note, rest = m.group(1), m.group(2), m.group(3)
        example = extract_json_block(rest)
        content = {}
        if example is not None:
            content["application/json"] = {
                "schema": infer_schema(example),
                "example": example,
            }
        elif "HTML" in note or "html" in rest[:80]:
            content["text/html"] = {"schema": {"type": "string"}}
        elif "PDF" in note or "pdf" in note:
            content["application/pdf"] = {
                "schema": {"type": "string", "format": "binary"}
            }
        elif "image" in note.lower() or "PNG" in note or "png" in note:
            content["image/png"] = {"schema": {"type": "string", "format": "binary"}}
        desc = note.strip(" —-:") or "Success"
        responses[status] = {"description": desc or "Response"}
        if content:
            responses[status]["content"] = content

    # Errors table, with the two back-reference idioms api.md uses.
    errors_match = re.search(
        r"\*\*Errors?\*\*([^\n]*)\n"
        r"(?:\|.*\|\s*\n\|[-\s|]+\|\s*\n((?:\|.*\|\s*\n?)+))?",
        body,
    )
    error_rows = []
    if errors_match:
        prefix = errors_match.group(1)
        table = errors_match.group(2)
        if "Admin" in prefix and "header errors" in prefix:
            error_rows.extend(ADMIN_BASE_ERRORS)
        elif "same as" in prefix and "above" in prefix:
            error_rows.extend(last_full_errors)
            # "...plus `X_NOT_FOUND` 404." style inline addition.
            for em in re.finditer(r"`([A-Z0-9_]+)`\s+(\d{3})", prefix):
                error_rows.append((em.group(1), int(em.group(2)), ""))
        if table:
            for line in table.strip().splitlines():
                cells = [c.strip() for c in line.strip().strip("|").split("|")]
                if len(cells) < 3:
                    continue
                code = cells[0].strip("`")
                try:
                    status = int(cells[1])
                except ValueError:
                    continue
                error_rows.append((code, status, cells[2]))

    if error_rows:
        by_status = {}
        for code, status, desc in error_rows:
            by_status.setdefault(status, []).append((code, desc))
        for status, entries in by_status.items():
            codes = ", ".join(f"`{c}`" for c, _ in entries)
            responses[str(status)] = {
                "description": "; ".join(d for _, d in entries if d) or codes,
                "content": {
                    "application/json": {
                        "schema": ERROR_SCHEMA_REF,
                        "examples": {
                            code: {"value": {"code": code, "message": desc or code}}
                            for code, desc in entries
                        },
                    }
                },
            }
        requires_auth = any(code == "MISSING_TOKEN" for code, _, _ in error_rows)
        return responses, error_rows, requires_auth

    requires_auth = any(code == "MISSING_TOKEN" for code, _, _ in inherited_errors)
    return responses, inherited_errors, requires_auth


def build_operation(tag, method, path, locked, body, last_full_errors):
    desc_match = re.match(r"\s*(.*?)(?=\n\*\*|\n```|\Z)", body, re.DOTALL)
    description = (desc_match.group(1).strip() if desc_match else "").replace("\n", " ")
    summary = re.split(r"(?<=[.!?])\s", description)[0] if description else f"{method} {path}"

    path_params = []
    for name in re.findall(r"\{(\w+)\}", path):
        path_params.append(
            {
                "name": name,
                "in": "path",
                "required": True,
                "schema": {"type": "string" if name == "token" else "integer"},
            }
        )

    query_params = extract_query_params(body)
    responses, new_last_errors, requires_auth = extract_responses(body, [], last_full_errors)
    if not responses:
        responses["200"] = {"description": "Success"}

    if len(summary) > 120:
        cut = summary.rfind(" ", 0, 117)
        summary = (summary[:cut] if cut > 0 else summary[:117]) + "..."

    op = {
        "tags": [tag],
        "summary": summary,
        "description": description or summary,
        "parameters": path_params + query_params,
        "responses": responses,
    }
    op["security"] = [{"bearerAuth": []}] if (locked or requires_auth) else []

    req_body = extract_request_body(body)
    if req_body is not None and method in ("POST", "PATCH", "PUT"):
        op["requestBody"] = req_body

    if not op["parameters"]:
        del op["parameters"]

    return op, new_last_errors


def main():
    text = API_MD.read_text(encoding="utf-8")

    paths = {}
    for tag, section_body in split_sections(text):
        if tag.startswith("Protected Routes"):
            continue
        last_full_errors = []
        for heading, block in split_endpoints(section_body):
            parsed = parse_heading(heading)
            if not parsed:
                continue
            method, path, locked = parsed
            if not path.startswith("/"):
                continue
            is_bare_html_page = "(HTML page)" in heading
            full_path = path if is_bare_html_page else "/api/v1" + path
            op, last_full_errors = build_operation(
                tag, method, path, locked, block, last_full_errors
            )
            paths.setdefault(full_path, {})[method.lower()] = op

    spec = {
        "openapi": "3.0.3",
        "info": {
            "title": "BeeTrack API",
            "version": "1.0.0",
            "description": (
                "Generated from docs/api.md — do not hand-edit; run "
                "`python docs/generate_openapi.py` after changing api.md."
            ),
        },
        "servers": [{"url": "http://localhost:8080", "description": "Local dev (docker compose up)"}],
        "components": {
            "securitySchemes": {
                "bearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
            },
            "schemas": {
                "Error": {
                    "type": "object",
                    "properties": {
                        "code": {"type": "string"},
                        "message": {"type": "string"},
                    },
                    "required": ["code", "message"],
                }
            },
        },
        "paths": dict(sorted(paths.items())),
    }

    OUT.write_text(
        "# Generated by docs/generate_openapi.py from docs/api.md — do not hand-edit.\n"
        + yaml.safe_dump(spec, sort_keys=False, allow_unicode=True, width=100),
        encoding="utf-8",
    )

    count = sum(len(methods) for methods in paths.values())
    print(f"Wrote {OUT} with {count} operations across {len(paths)} paths.")


if __name__ == "__main__":
    sys.exit(main())
