# Nanofuse Tray App Requirements

**Date:** 2026-05-30

## Decision

The tray/menu app must be an API client. It must not embed hypervisor control, shell into the runtime host, edit Nanofuse storage directly, or manipulate Firecracker sockets. The app talks to `nanofused` through the REST API described by `api/openapi.yaml`.

This PR captures requirements and API prerequisites only. It does not add Electron, Tauri, Wails, Swift, .NET, or another desktop runtime because that is a repo-wide dependency and packaging decision.

## Target Platforms

- macOS menu bar app.
- Windows tray app.
- Linux tray support is optional after macOS and Windows work.

## Required Views

| View | Required behavior |
|------|-------------------|
| Connection profiles | Add/edit/select API profiles, including URL, timeout, optional tunnel instructions, and future auth material reference. |
| Health and capabilities | Show `/health` and `/capabilities`; disable VM actions when `native_runtime` is false or API is unreachable. |
| VM list | Show VM name, ID, state, image, CPU, memory, uptime, owner/group identity, and last update. |
| VM actions | Create, start, stop, kill, delete, pause/resume where implemented, and open logs. |
| Image list | Show pulled images, tags, architecture, size, labels, and delete where safe. |
| Image pull | Pull image by reference and poll `/images/jobs/{id}` with progress. |
| Logs | Tail VM console logs with copy/save controls and no hidden credential display. |
| Egress policy | Select default-deny, proxy-only, DNS allow, and allow-rule templates during VM create. |
| Notifications | Surface daemon unreachable, VM state transitions, image pull completion/failure, and KVM capability failures. |

## Security Requirements

- Store only API profile metadata and references to credentials in the OS keychain or credential manager.
- Never store raw upstream LLM/API/MCP secrets in app config.
- Treat raw TCP API profiles as insecure unless the profile uses localhost, SSH tunnel, WireGuard, or an authenticated reverse proxy.
- Do not display guest logs as trusted UI content; render as plain text.
- Require confirmation for destructive VM/image actions.

## API Requirements Before Implementation

- `GET /capabilities` exists and is used for capability gating.
- `api/openapi.yaml` is the source of generated client types.
- Remote API auth profile is designed before any non-localhost default profile ships.
- VM create supports egress policy selection from the app.
- Image pull job polling is stable enough for progress UI.

## Candidate Desktop Stacks

| Stack | Cost | Fit |
|-------|------|-----|
| Tauri v2 | Adds Rust and web frontend toolchain. Smaller app footprint, strong tray support. | Good fit if the repo accepts Rust/Node packaging. |
| Wails | Adds Go desktop/webview build path. Keeps backend language aligned with Nanofuse. | Good fit if native webview differences are acceptable. |
| Electron | Adds Node/Chromium and larger packaging surface. | Fastest ecosystem, largest runtime footprint. |
| Native Swift + .NET/WinUI | Best native behavior. Two codebases. | Only justified if native UX is more important than shared implementation. |

No stack is selected in this PR.
