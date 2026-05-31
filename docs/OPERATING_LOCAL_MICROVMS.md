# Operating Local MicroVMs

This runbook answers the common local operations questions for the API, CLI, and tray app.

## Start macOS Local Runtime

```bash
./scripts/run-tray-macos.sh --start-api --restart
export NANOFUSE_API_URL="http://127.0.0.1:18080"
```

Validate the daemon and runtime:

```bash
nanofuse health
curl "$NANOFUSE_API_URL/capabilities"
```

The expected macOS runtime is `driver=apple_container` with `native_runtime=true`.

## See Published Ports

Use the Nanofuse CLI first:

```bash
nanofuse vm ports
nanofuse vm ports <vm-id-or-name>
nanofuse vm list
nanofuse vm status <vm-id-or-name>
```

`nanofuse vm ports` shows configured host-to-VM forwards. TCP forwards also get a localhost reachability check.

Host-level checks:

```bash
lsof -nP -iTCP -sTCP:LISTEN | grep 18081
nc -vz 127.0.0.1 18081
```

API equivalent:

```bash
curl "$NANOFUSE_API_URL/vms" | jq '.vms[] | {id,name,state,ports:.config.network.port_forwards,runtime:.runtime}'
```

## Access a Running VM

First-party command execution goes through the daemon API:

```bash
nanofuse vm exec <vm-id-or-name> -- uname -a
nanofuse vm exec <vm-id-or-name> -- sh -lc 'cat /etc/os-release'
```

This is not SSH. It works when the selected runtime backend supports exec. The macOS Apple-container backend supports it.

SSH requires a guest image that has `sshd` installed and running, plus a port forward to guest port 22:

```bash
nanofuse vm run <ssh-capable-image> ssh-test --ssh-key ~/.ssh/id_ed25519.pub --port-forward 2222:22
ssh -p 2222 root@127.0.0.1
```

Plain images such as Alpine usually do not run `sshd` by default. Use `nanofuse vm exec` for those images unless the image explicitly includes and starts SSH.

## Launch More Than One VM

Give each VM a unique name. Give each VM a unique host port when publishing services:

```bash
nanofuse vm run alpine:3.20 api-1 --port-forward 18081:8080
nanofuse vm run alpine:3.20 api-2 --port-forward 18082:8080
nanofuse vm list
nanofuse vm ports
```

The tray app lists up to 25 VMs with state and port context. Select a VM from the VM list, then use start, stop, kill, or delete.

## Enable More Launchable Images

Pull or resolve the image through the API boundary, then refresh the tray:

```bash
nanofuse image pull docker.io/library/alpine:3.20
nanofuse image pull docker.io/library/ubuntu:24.04
nanofuse image list
```

On macOS with `runtime.driver=apple_container`, `nanofuse vm run <oci-ref>` can also resolve OCI image references through Apple container when creating the VM:

```bash
nanofuse vm run docker.io/library/alpine:3.20 alpine-test
nanofuse image list
```

API equivalent:

```bash
curl -X POST "$NANOFUSE_API_URL/images/pull" \
  -H "Content-Type: application/json" \
  -d '{"image_ref":"docker.io/library/ubuntu:24.04"}'

curl "$NANOFUSE_API_URL/images"
```

In the tray app, choose `Refresh`, select a cached image under `Images`, then choose `Create and Start VM From Image`.
