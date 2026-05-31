# Nanofuse Launch One-Liners

Nanofuse VM execution runs on a Linux/KVM host. macOS and Windows are API clients only; they do not run Firecracker locally.

There is no tray/menu application in this repo yet. The tray app is a planned API client, not a shipped binary.

Primary runtime requirement: Firecracker requires Linux KVM and read/write access to `/dev/kvm`. Apple documents nested virtualization support as available on Macs with M3 chips and later. Parallels documents nested virtualization as not supported on Mac computers with Apple silicon for its current nested Hyper-V path, and the local Parallels VM on this Apple M2 Max does not expose `/dev/kvm`.

## Tested Status

| Path | Status |
|------|--------|
| macOS CLI binary starts | Tested on macOS arm64 |
| macOS CLI reaches remote Linux/KVM API through SSH tunnel | Tested from this Mac to `dublin-wg` on local port `18082` |
| macOS CLI creates and starts a Firecracker VM through the API | Tested: VM `mac-api-smoke` reached `running` and console logs showed Ubuntu boot |
| Linux/KVM direct Firecracker boot | Tested on `dublin-wg` with `/dev/kvm`, Firecracker `1.7.0`, kernel `6.1.77`, and Ubuntu rootfs |
| Linux/KVM rootless API daemon with no VM networking | Tested with `network.setup=false` and VM `network=none` |
| Windows one-liner | Syntax only; not tested in this workspace |
| Local macOS Parallels Vagrant as Firecracker host | Tested and failed on this Apple M2 Max with Parallels 26.1.1: guest has no `/dev/kvm`; `--nested-virt on` makes the VM fail to start; `--hypervisor-type parallels` is rejected |

If the TCP API is down, confirm reachability with `curl http://127.0.0.1:18080/health` or `curl http://linux-kvm-host:8080/health`.

## Linux/KVM Host

Run this on the machine that has Linux, `/dev/kvm`, and Firecracker for the normal privileged daemon:

```bash
cd /path/to/nanofuse && ./scripts/ensure-mage.sh && mage build && sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
```

For remote clients, keep the daemon bound to `127.0.0.1:8080` and use an SSH tunnel from the client machine.

For an explicit no-network rootless validation daemon, use a config with:

```yaml
network:
  setup: false
```

Then create VMs with `--network none`.

## Local Parallels Result

This was tested locally on this Mac:

```bash
swift -e 'import Virtualization; print(VZGenericPlatformConfiguration.isNestedVirtualizationSupported)'
```

Result:

```text
false
```

Guest KVM check:

```bash
vagrant ssh -c 'ls -l /dev/kvm 2>/dev/null || echo /dev/kvm-missing'
```

Result:

```text
/dev/kvm-missing
```

Trying to enable nested virtualization in the Parallels VM config succeeds, but the VM does not start:

```bash
prlctl set 8eda22c1-1ee1-4069-94f9-5b5befdb2be8 --nested-virt on
prlctl start 8eda22c1-1ee1-4069-94f9-5b5befdb2be8
```

Result:

```text
Failed to start the VM: Unable to start the virtual machine. The virtual machine cannot be started.
```

Trying to switch the same Apple Silicon VM to the Parallels hypervisor is rejected:

```bash
prlctl set 8eda22c1-1ee1-4069-94f9-5b5befdb2be8 --hypervisor-type parallels
```

Result:

```text
Unable to commit VM configuration: Unable to start the virtual machine.The virtual machine configuration is invalid.
```

The current local Parallels VM cannot be the Firecracker runtime host on this Mac. It can still be used for API/client tests that do not boot Firecracker.

Sources:

- Firecracker Getting Started: https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md
- Apple `isNestedVirtualizationSupported`: https://developer.apple.com/documentation/virtualization/vzgenericplatformconfiguration/isnestedvirtualizationsupported
- Parallels nested virtualization KB: https://kb.parallels.com/en/116239

## macOS Client

Run this from the Mac after replacing `user@linux-kvm-host`:

```bash
ssh -fN -L 18080:127.0.0.1:8080 user@linux-kvm-host && NANOFUSE_API_URL=http://127.0.0.1:18080 ./bin/nanofuse health
```

This one-liner requires SSH key authentication or another non-interactive SSH setup. If SSH prompts for a password or host-key approval, run the tunnel in one terminal instead:

```bash
ssh -L 18080:127.0.0.1:8080 user@linux-kvm-host
```

Then run the health check in another terminal:

```bash
NANOFUSE_API_URL=http://127.0.0.1:18080 ./bin/nanofuse health
```

Verified working from this Mac to `dublin-wg`:

```bash
ssh -fN -L 18082:127.0.0.1:18080 dublin-wg && ./bin/nanofuse --api-url http://127.0.0.1:18082 health
```

Verified VM launch through that API:

```bash
./bin/nanofuse --api-url http://127.0.0.1:18082 vm run nanofuse-ci:latest mac-api-smoke --network none --vcpus 1 --memory 256 --kernel-args 'console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw'
```

Verified status and logs:

```bash
./bin/nanofuse --api-url http://127.0.0.1:18082 vm status 04b5c05f-4a45-4a96-a2a6-37cbadb51e58
./bin/nanofuse --api-url http://127.0.0.1:18082 vm logs 04b5c05f-4a45-4a96-a2a6-37cbadb51e58 --tail 80
```

## Windows PowerShell Client

Run this from Windows PowerShell after replacing `user@linux-kvm-host`:

```powershell
Start-Process ssh -ArgumentList '-N','-L','18080:127.0.0.1:8080','user@linux-kvm-host' -WindowStyle Hidden; $env:NANOFUSE_API_URL='http://127.0.0.1:18080'; .\nanofuse.exe health
```

This assumes `ssh.exe` is installed and the Windows `nanofuse.exe` binary is in the current directory.

## Direct TCP Client

Use this only on a trusted management network or behind another authenticated transport:

```bash
NANOFUSE_API_URL=http://linux-kvm-host:8080 ./bin/nanofuse health
```

```powershell
$env:NANOFUSE_API_URL='http://linux-kvm-host:8080'; .\nanofuse.exe health
```
