# Nanofuse Tray App Requirements

**Date:** 2026-05-30

## Decision

The tray/menu app must be an API client. It must not embed hypervisor control, shell into the runtime host, edit Nanofuse storage directly, or manipulate Firecracker sockets. The app talks to `nanofused` through the REST API described by `api/openapi.yaml`.

The current implementation adds a minimal Go tray client at `cmd/nanofuse-tray` using `github.com/getlantern/systray`. This intentionally avoids Electron, Tauri, Wails, Swift, .NET, or another larger desktop runtime until packaging, signing, and installer requirements are selected.

## Implemented Slice

| Capability | Status |
|------------|--------|
| macOS menu bar app | Implemented and built locally as `bin/nanofuse-tray`. |
| Windows tray app | Implemented and cross-built as `nanofuse-tray.exe`; desktop runtime click testing still needs a Windows session. |
| Health/capabilities | Implemented through `GET /health` and `GET /capabilities`. |
| VM list | Implemented through `GET /vms`, limited to 25 menu rows and showing readable name, state, image, and port context. |
| Image list | Implemented through `GET /images`, limited to 25 menu rows. |
| Create/start VM from image | Implemented through `POST /vms` followed by `POST /vms/{id}/start` for either a typed OCI image reference or a selected cached image. |
| Add image | Implemented as a prompt that calls `POST /images/pull`; progress polling is still future UI work. |
| VM start/stop | Implemented as per-VM row actions through `POST /vms/{id}/start` and `POST /vms/{id}/stop`. |
| VM kill/delete | Implemented as per-VM row actions with second-click confirmation through `POST /vms/{id}/kill` and `DELETE /vms/{id}`. |
| Smoke mode | Implemented with `nanofuse-tray --smoke --api-url ...` for non-GUI validation. |

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
| VM actions | Create from selected image or typed OCI reference, start, stop, kill, delete, pause/resume where implemented, and open logs. Current app implements prompt-based create/start plus per-row start, stop, kill, and delete. |
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

## Remaining API Requirements

- `GET /capabilities` exists and is used for capability gating.
- `api/openapi.yaml` is the source of generated client types.
- `POST /vms/{id}/exec` exists for runtime-backed command execution; interactive SSH still requires guest sshd and a port forward to port 22.
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

The selected current stack is Go plus `getlantern/systray`. This is reversible because API behavior is isolated in `internal/trayapp` and the OS tray loop is isolated in `cmd/nanofuse-tray`.
