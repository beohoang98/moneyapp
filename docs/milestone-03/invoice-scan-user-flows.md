# Invoice Document Scanning — User Flow Diagrams

**Feature scope**: Milestone 3 — Document Scanning (Epic 8, SC-01 – SC-14)  
**Last updated**: 2026-04-27  
**Author**: QA/UX Tester agent

**Route source of truth**: HTTP paths and payloads in this doc are aligned with **`docs/tickets.md`** (Epic 8 + API appendix). If anything drifts during implementation, update tickets first, then this file.

---

## 1. Settings Flow — Configure Scanning

Users reach Scanning Settings via **Settings → Document Scanning** in the sidebar navigation.

### 1.1 Settings Configuration Flowchart

```mermaid
flowchart TD
    A([User opens Settings page]) --> B[Navigate to\nDocument Scanning section]
    B --> C{Scanning currently\nenabled?}

    C -- No --> D[Toggle is OFF\nAll fields disabled / greyed-out]
    C -- Yes --> E[Toggle is ON\nFields are active]

    D --> F[User flips Enable\nDocument Scanning toggle ON]
    F --> G[Fields become editable\nwith default values pre-filled]

    G --> H[Base URL\ndefault: http://localhost:11434/v1]
    G --> I[Model name\ndefault: qwen3-vl:4b]
    G --> J[API Key optional\nplaceholder: uses OPENAI_API_KEY env]

    H --> K[User edits fields\nas needed]
    I --> K
    J --> K

    E --> K

    K --> L{User clicks\nTest Connection}

    L --> M[/POST /api/settings/scanning/test/]
    M --> N{API responds\nwithin timeout?}

    N -- Yes, healthy --> O[✅ Success banner:\nConnected to model name]
    N -- Timeout / unreachable --> P[❌ Error banner:\nCould not reach Base URL]
    N -- Auth error 401 --> Q[❌ Error banner:\nInvalid API key]
    N -- Model not found 404 --> R[❌ Error banner:\nModel not found on server]

    O --> S[User clicks Save Settings]
    P --> S2{User corrects\nfields?}
    Q --> S2
    R --> S2

    S2 -- Yes --> K
    S2 -- No, save anyway --> S

    S --> T[/PUT /api/settings/scanning/]
    T --> U[✅ Settings persisted\nto database / config]
    U --> V[Page reloads → settings\nsurvive browser refresh]

    K2[User flips Enable toggle OFF] --> K3[/PUT /api/settings/scanning enabled=false/]
    K3 --> K4[Scan affordance disabled\nacross entire app]
```

### 1.2 Settings Data Model

| Field | Type | Default | Notes |
|---|---|---|---|
| `enabled` | boolean | `false` | Master gate for scan feature |
| `base_url` | string | `http://localhost:11434/v1` | OpenAI-compatible endpoint |
| `model` | string | `qwen3-vl:4b` | Vision model identifier |
| `api_key` | string | `""` | Optional; falls back to `OPENAI_API_KEY` env var |

### 1.3 Test Connection — Sequence Diagram

```mermaid
sequenceDiagram
    actor User
    participant FE as Frontend (SettingsPage)
    participant BE as Backend (/api/settings/scanning)
    participant LLM as Vision API (Ollama / OpenAI-compat)

    User->>FE: Clicks "Test Connection"
    FE->>FE: Disable button, show spinner
    FE->>BE: POST /api/settings/scanning/test\n{base_url, model, api_key}
    BE->>LLM: GET {base_url}/models  [or lightweight probe]
    alt API healthy, model found
        LLM-->>BE: 200 OK  {models: [...]}
        BE-->>FE: 200 {ok: true, message: "Connected to qwen3-vl:4b"}
        FE->>User: ✅ "Connected to qwen3-vl:4b"
    else API unreachable / timeout > 5 s
        LLM-->>BE: connection refused / timeout
        BE-->>FE: 200 {ok: false, message: "Cannot reach base_url"}
        FE->>User: ❌ "Could not reach http://localhost:11434/v1"
    else API key rejected (401)
        LLM-->>BE: 401 Unauthorized
        BE-->>FE: 200 {ok: false, message: "Invalid API key"}
        FE->>User: ❌ "Invalid API key"
    else Model not found (404)
        LLM-->>BE: 404 Not Found
        BE-->>FE: 200 {ok: false, message: "Model qwen3-vl:4b not found"}
        FE->>User: ❌ "Model not found — is it pulled on Ollama?"
    end
    FE->>FE: Re-enable button
```

---

## 2. Usage Flow — Scan an Invoice

Users initiate scanning from the **Invoices** page (primary entry point per M3 exit criteria).

### 2.1 Scan Feature Gate

```mermaid
flowchart TD
    A([User opens Invoices page]) --> B[/GET /api/scanning/health/]
    B --> C{Health status?}

    C -- disabled / not configured --> D[Scan button greyed-out\nTooltip: Enable scanning in Settings]
    C -- configured but unhealthy --> E[Scan button greyed-out\nTooltip: Scanning service unreachable.\nCheck Settings → Document Scanning]
    C -- healthy --> F[Scan Invoice button active\n+ pulsing indicator optional]

    D --> G{User clicks\ngreyed button?}
    E --> G
    G -- Yes --> H[Toast / inline hint:\nGo to Settings → Document Scanning\nto configure]
    G -- No --> I([User uses manual\ncreate instead])

    F --> J[User clicks Scan Invoice]
    J --> K[[Continue to Scan Flow →\nsee section 2.2]]
```

### 2.2 Full Scan Flow — Happy Path

```mermaid
flowchart TD
    START([User clicks Scan Invoice]) --> A[Open Scan Modal / Drawer]
    A --> B[User selects image file\nJPEG · PNG · WEBP · HEIC · max 10 MB]

    B --> C{Client-side validation}
    C -- invalid type or > 10 MB --> D[❌ Client error:\nFile too large or unsupported format]
    D --> B

    C -- valid --> E[Preview thumbnail shown\nin modal]
    E --> F[User clicks Scan]

    F --> G[/POST /api/scanning/invoice\nmultipart: image file/]
    G --> H[Backend: validate MIME + size\nstore image to ObjectStore\nget temp storage_key]
    H --> I[Backend: call Vision API\nwith image + structured JSON prompt]

    I --> J{Vision API response\nwithin 60 s?}

    J -- Success, JSON parseable --> K[Backend: parse extraction\nreturn structured result]
    J -- Timeout > 60 s --> L[❌ 504: Scan timed out]
    J -- Vision API error 5xx --> M[❌ 502: Scan service error]
    J -- Partial JSON / required fields missing --> N[⚠️ Partial result:\nmissing fields left blank]

    K --> O[/200 JSON:\nvendor_name · date · total_amount\ncurrency · line_items array · storage_key/]
    N --> O

    L --> P[Modal shows error\nUser can retry or enter manually]
    M --> P

    O --> Q[[Continue to Review Form →\nsee section 2.3]]
```

### 2.3 Review Form — Confirm or Cancel

```mermaid
flowchart TD
    A([Scan result received]) --> B[Review Form renders\nwith pre-filled fields]

    B --> C[Editable fields:\n• Vendor / Payee\n• Invoice date\n• Total amount\n• Currency\n• Description / notes\n• Category dropdown]
    B --> D[Line items table\nSC-07: desc + amount per row\nread-only list per E8-S4]

    C --> E{User reviews\nand edits}
    D --> E

    E --> F{Action?}

    F -- Confirm Save --> G[/POST /api/invoices\nwith reviewed fields + storage_key/]
    G --> H[Backend: create invoice record\npromote temp image via\nPOST /api/attachments\nentity_type=invoice · entity_id=new_id\nsource_storage_key=scan-tmp/...]
    H --> I[✅ Invoice created\nImage attached automatically\nSC-10]
    I --> J[Modal closes\nInvoice list refreshes\nNew row visible]

    F -- Cancel / Close X --> K{Confirm discard?}
    K -- User confirms discard --> L[/DELETE /api/scanning/temp\nJSON body: storage_key/]
    L --> M[Backend: delete temp image\nfrom ObjectStore immediately\nOQ9 — no orphan files]
    M --> N[Modal closes\nNo record created SC-06\nNo attachment stored]

    K -- User stays --> E
```

### 2.4 Full API Sequence — Scan to Save

```mermaid
sequenceDiagram
    actor User
    participant FE as Frontend
    participant BE as Backend
    participant Store as ObjectStore (local/S3)
    participant LLM as Vision API

    User->>FE: Select image file + click Scan
    FE->>FE: Validate MIME type + file size (client-side)
    FE->>BE: POST /api/scanning/invoice  (multipart/form-data)

    BE->>BE: Validate MIME type + size (NF-25, max 10 MB)
    BE->>Store: Upload image → returns storage_key (temp)
    Store-->>BE: storage_key

    BE->>LLM: POST {base_url}/chat/completions\n{model, image_data_url, extraction_prompt}
    Note over BE,LLM: Timeout: 60 s

    alt Happy path
        LLM-->>BE: 200 JSON with extracted fields
        BE->>BE: Parse + validate extraction result
        BE-->>FE: 200 {vendor_name, date, total_amount,\ncurrency, line_items, storage_key}
        FE->>User: Review form pre-filled
    else Timeout
        LLM-->>BE: (no response after 60 s)
        BE->>Store: DELETE temp image
        BE-->>FE: 504 {error: "scan_timeout"}
        FE->>User: Error state — enter manually
    else LLM API error
        LLM-->>BE: 5xx
        BE->>Store: DELETE temp image
        BE-->>FE: 502 {error: "vision_api_error"}
        FE->>User: Error state — enter manually
    end

    User->>FE: Review fields, edit if needed
    User->>FE: Click "Save Invoice"

    FE->>BE: POST /api/invoices  {fields...}
    BE->>BE: Create invoice record (DB)
    BE-->>FE: 201 {invoice_id}

    FE->>BE: POST /api/attachments\n{entity_type: invoice, entity_id,\nsource_storage_key: temp key,\nfilename, mime_type}
    BE->>BE: Promote temp object + create attachment row
    BE-->>FE: 201 {attachment_id}

    FE->>User: ✅ Invoice saved + image attached

    Note over User,FE: — CANCEL PATH —
    User->>FE: Click Cancel
    FE->>BE: DELETE /api/scanning/temp\n{storage_key: "scan-tmp/..."}
    BE->>Store: Delete temp image
    BE-->>FE: 204 No Content
    FE->>User: Modal closed, no records written
```

---

## 3. Edge Cases — QA Verification Checklist

The following edge cases **must** be covered in manual or automated testing before M3 sign-off:

1. **Ollama not running** — Test Connection and Scan both fail with a clear, user-readable message (not a raw HTTP error). The app must not crash or freeze.

2. **Ollama running but model not pulled** — `/api/models` returns 200 but `qwen3-vl:4b` is absent. Backend must distinguish "API reachable" from "model available" and surface a specific error: _"Model qwen3-vl:4b not found. Run `ollama pull qwen3-vl:4b`."_

3. **Vision API timeout (> 60 s)** — Backend enforces a hard 60-second timeout. The temp image stored in ObjectStore must be deleted on timeout; no orphan file is left behind.

4. **Unreadable / blurry image** — LLM returns a response but with no extractable financial fields. System treats this as a partial result (SC-05): review form opens with all fields blank, banner reads _"Could not extract data from image — please fill in manually."_ No hard error.

5. **Partial JSON from Vision API** — LLM response is valid JSON but missing required fields (e.g., `total_amount` absent). Backend must accept the partial result, return `null` for missing fields, and let the user fill gaps in the review form rather than failing the entire scan (D9 risk).

6. **Malformed JSON from Vision API** — LLM response cannot be parsed as JSON at all. Backend must catch the parse error and return a degraded result (all fields null / empty) rather than propagating a 500. Review form opens in manual-entry mode.

7. **User cancels after image uploaded but before scanning** — If the modal is closed before the scan API is called, no image has been written to ObjectStore yet, so no cleanup is needed. If the scan is in-flight, the frontend must wait for the response (or abort with a cancel signal), then send the DELETE temp request with the `storage_key` from the response.

8. **User cancels after scan succeeds but before saving** — The review form has a `temp_storage_key` from the scan response. Clicking Cancel must trigger `DELETE /api/scanning/temp` with JSON body `{ "storage_key": "scan-tmp/..." }` and receive 204. Verify in ObjectStore that the file is gone (no orphan per OQ9).

9. **Settings persistence across reload** — After saving scan settings, reload the page. All values — enabled toggle, base URL, model name, API key presence (masked) — must be restored from the database without re-entry.

10. **API key env-var fallback** — If the API key field is left blank in Settings, the backend must silently use the `OPENAI_API_KEY` environment variable (NF-26). A scan should succeed when the env var is set and the field is empty; fail with "Invalid API key" only when neither the field nor the env var provides a key.

---

## 4. Out-of-Scope Reminders (Do Not Test in M3)

- Batch scanning of multiple images (SC-16 — Won't have)
- On-device OCR / Tesseract.js (SC-15 — Won't have)
- Camera capture via `<input capture>` (SC-12 / E8-S8 — deferred to M4)
- Scan audit log / `scan_results` table (SC-13 — Could have, deferred)
- Category auto-suggestion from vendor name (SC-14 — Could have, deferred)
- Multi-user / shared scanning credentials (out of scope entirely)
