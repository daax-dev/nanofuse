# Plan: Building a Custom Firecracker MicroVM Image (Ubuntu 24.04 Base)

This plan outlines a phased approach to create a custom Firecracker MicroVM image using Ubuntu 24.04 as a minimal base. The goal is to produce a tiny base image (with no extra packages) that can be extended for deep learning workloads, and to automate its build and distribution via GitHub Actions and a private GitHub Container Registry (GHCR). The image will serve as a foundation for isolated microVMs running the two components of Trigger.dev (web UI and supervisor/worker), similar to how SlicerVM manages MicroVM images.

**Note:** Firecracker requires two main components for a MicroVM: **(1)** an **uncompressed Linux kernel** (`vmlinux`), and **(2)** an **ext4 root filesystem image** for the guest OS:contentReference[oaicite:0]{index=0}. We will build a custom kernel and a minimal root filesystem image that includes an init system (systemd) and essential services, following best practices from Firecracker and SlicerVM.

## Phase 1: Define Requirements and Base OS Selection

- **Base OS:** Use **Ubuntu 24.04** (Jammy+1 LTS) as the guest OS. Ubuntu 24.04 is supported for Firecracker in SlicerVM’s images:contentReference[oaicite:1]{index=1} and provides a familiar environment. We will create a minimal Ubuntu 24.04 root filesystem (tiny base image) that can be extended with additional layers as needed (e.g. adding deep learning libraries or application code later). No unnecessary packages will be pre-installed – this keeps the image lightweight for fast boot and allows customization in higher layers (the principle is to start with a minimal OS and add only what’s needed):contentReference[oaicite:2]{index=2}.

- **Init System:** Use **systemd** as the init system in the guest. SlicerVM relies on systemd to manage services and background agents:contentReference[oaicite:3]{index=3}, and using systemd ensures compatibility with Slicer’s expectations and general Linux tooling. Other init systems (OpenRC, etc.) could work, but are not officially supported in Slicer’s context:contentReference[oaicite:4]{index=4}. Systemd will handle service startup (including our custom app services) and system tuning inside the VM.

- **Networking & Access:** Plan to include an **OpenSSH server** to allow SSH access into the VM for administration or Slicer’s serial-over-SSH console. We will ensure a default user (e.g. `ubuntu`) exists in the image with appropriate permissions (passwordless sudo) and that SSH key-based login can be configured (likely via userdata or by baking an authorized_keys file for development). The image should also have basic networking tools (e.g. `iproute2` for configuring networking) and a DHCP client or static network config, depending on how networking is managed (Slicer can set up a TAP interface and provide DHCP or static IPs to VMs). We will make sure a **serial console** is enabled (such as a `getty` on `ttyS0`) so that the VM can be accessed via the Firecracker serial port (Slicer uses this for its SOS console in case network access fails).

- **Deep Learning Support:** For CPU-only deep learning tasks, the base image may include libraries like Python and possibly CUDA drivers if NVIDIA GPUs are intended. However, Firecracker **does not support PCI pass-through**, so if GPU support is needed, using **Cloud Hypervisor (CH)** instead of Firecracker is required:contentReference[oaicite:5]{index=5}. In this plan, we focus on Firecracker for simplicity (fast, lightweight VMs). If GPU acceleration is required in the future (for example, attaching an NVIDIA GPU to the VM for deep learning), we would plan a variant of the image compatible with Cloud Hypervisor (with a kernel that supports PCIe/VFIO and NVIDIA drivers installed):contentReference[oaicite:6]{index=6}. For now, the base image will be CPU-focused but structured such that we can extend or rebuild for CH if needed.

## Phase 2: Build a Custom Linux Kernel for Firecracker

In Firecracker, the **VM kernel is provided by the host** (as an uncompressed binary). We will build a custom kernel to ensure it’s optimized for MicroVM use and includes any features needed for our workloads:

- **Kernel Version:** We can base our kernel on a stable LTS version. Slicer’s official images use Linux **5.10.240** (an LTS kernel):contentReference[oaicite:7]{index=7} for x86_64, which is known to work with Firecracker. Alternatively, we might choose a newer LTS (e.g. 6.1 or 6.2) for updated hardware support, but using a proven version like 5.10.x or 5.15.x can be safer initially. For deep learning tasks, newer kernels could help with improved container isolation or newer drivers, but the difference is minor unless GPU support is involved. We’ll proceed with 5.10.x to match Slicer’s kernel (ensuring compatibility and stability), and can upgrade later once the base setup is working.

- **Kernel Configuration:** Use Firecracker’s recommended kernel config as a baseline. The Firecracker team provides a sample minimal config (`microvm-kernel-x86_64.config`) that includes virtio drivers and excludes unnecessary devices (no ACPI, no PCI, etc.):contentReference[oaicite:8]{index=8}. We will start from this config to enable fast boot and small image size. Key settings include enabling virtio-block, virtio-net drivers (for disk and network), console on ttyS0, and disabling modules (built-in drivers for faster boot). We will also include support for KVM (for nested virt, if running inside another VM or for potential use of KVM hypervisor) and any debugging features we might need. If we plan to use Cloud Hypervisor for GPU later, we’d ensure PCIe, VFIO, and related drivers can be enabled (likely as loadable modules). These can be compiled as modules and included in the image if not built-in.

- **Build Process:** Compile the kernel on an x86_64 Linux build host. This can be done locally or in CI. We will integrate it into GitHub Actions (see Phase 4), but initially, the steps are:
  1. **Fetch Kernel Source:** Clone the Linux kernel source for the chosen version (e.g. v5.10.240) or download an official tarball.
  2. **Apply Config:** Copy the Firecracker recommended config as `.config` and adjust if needed (e.g., enable `CONFIG_KVM` or others if desired):contentReference[oaicite:9]{index=9}. We can run `make menuconfig` (locally) to tweak options, but in automation we might skip interactive steps by using the provided config directly.
  3. **Compile** the kernel using `make -j$(nproc) vmlinux`. We specifically want an **uncompressed vmlinux** (not bzImage) since Firecracker expects an uncompressed image:contentReference[oaicite:10]{index=10}. This will produce a `vmlinux` file at the top of the build directory if successful:contentReference[oaicite:11]{index=11}.
  4. **Output:** The result is a `vmlinux` binary (several MBs in size). If any kernel modules are enabled (hopefully minimal or none), gather those as well (`make modules_install` to an output directory) so we can include them in the rootfs under `/lib/modules/<version>`.

- **Kernel Optimization:** To keep boot times low, we’ll pass appropriate kernel boot parameters at runtime (Firecracker does this via its API or config file). Typical parameters include `console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw`:contentReference[oaicite:12]{index=12}. These ensure the kernel uses the serial console, doesn’t use legacy PIC or reboot on panic, and treats the FS as read-write. We will make sure these or similar are set when launching the VM. Additionally, because we compile our kernel, we can strip out features not needed (e.g. no sound, no unnecessary filesystems, etc.) to reduce boot overhead. The chosen config already aligns with a minimal VM use case:contentReference[oaicite:13]{index=13}.

- **(Optional) Using Existing Kernel:** As a shortcut, one could use the pre-built kernel from Slicer or Firecracker’s demo (for example, Firecracker provides a demo kernel “hello-vmlinux.bin” that can be downloaded:contentReference[oaicite:14]{index=14}). However, since this is a **complete rebuild**, we will build our own to have full control (and to allow adding features like VFIO later). Still, it’s worth noting that initially using a known-good kernel can speed up testing. The plan assumes we compile our kernel in CI for repeatability.

## Phase 3: Create the Minimal Ubuntu 24.04 Root Filesystem

Next, we construct the **root filesystem image** (which will become the MicroVM’s disk). We want a minimal Ubuntu 24.04 environment with systemd, networking, and SSH – essentially similar to an Ubuntu Cloud image but stripped down. We will leverage Docker to build this rootfs, as it provides a convenient chroot-like environment to install packages and configure the image, which we can then export as an ext4 disk image:contentReference[oaicite:15]{index=15}:contentReference[oaicite:16]{index=16}.

**Steps to build the rootfs:**

- **Dockerfile Setup:** We will write a `Dockerfile` to automate assembling the filesystem:
  - Start `FROM ubuntu:24.04` – this gives a base file system for Ubuntu 24.04. Note that the official Ubuntu container image is minimal and **does not include an init system by default** (containers don’t run a full init):contentReference[oaicite:17]{index=17}. We will therefore install the `systemd` package (and related packages) ourselves.
  - `RUN apt-get update && apt-get install -y systemd systemd-sysv openssh-server sudo` – install systemd (the `systemd-sysv` package provides the `/sbin/init` symlink to systemd), OpenSSH server, and sudo. We also include any other core utilities we need (for example, `iproute2` for networking, `curl` or `wget` for basic web requests, etc., though we keep it minimal). This mirrors the approach recommended by Firecracker docs and others: **ensure an init is present** because Ubuntu containers don’t have one by default:contentReference[oaicite:18]{index=18}.
  - (Optional) Install cloud-init? – Likely we **do not include cloud-init** to keep things lightweight, unless we want to leverage it for user-data scripts. SlicerVM docs suggest using either user-data scripts or building images via Docker:contentReference[oaicite:19]{index=19}. We can manage initialization via our own scripts or systemd units, so cloud-init is probably unnecessary overhead for our case.
  - **Enable essential services:** After installing, we’ll enable necessary services to start on boot:
    - `systemctl enable ssh` (to ensure the SSH server starts on boot and allows login).
    - Ensure `systemd-networkd` or `networking` service is active if we rely on DHCP. Alternatively, we might place a simple `/etc/netplan` config or `/etc/network/interfaces` if using ifupdown. Slicer likely sets up networking externally (via TAP and static IP or DHCP), but having DHCP client in the image (e.g. `systemd-networkd` with a DHCP config for `eth0`) would let it acquire an IP automatically. We can configure `systemd-networkd` with a default config for `eth0` to DHCP.
    - Enable a serial console getty: e.g., `systemctl enable serial-getty@ttyS0.service` to allow login via Firecracker’s serial port. This ensures if network is not reachable, we can still connect through the console (Slicer’s *Serial Over SSH* relies on this console login).
  - **Create default user:** Add an `ubuntu` user to mirror Ubuntu cloud images:
    - `RUN useradd -m -s /bin/bash ubuntu && usermod -aG sudo ubuntu` – create user with home and bash shell, add to sudo group.
    - Optionally set `ubuntu` password to none (lock it) or a known default (cloud images usually have it locked and rely on SSH keys). We can leave it locked and rely on key auth.
    - Ensure `sudo` is passwordless for convenience: we can drop a `/etc/sudoers.d/ubuntu` file with `%sudo ALL=(ALL) NOPASSWD:ALL`.
  - **Clean up:** Remove package caches (`apt-get clean` and truncate logs, etc.) to reduce size.

- **Incorporate the Kernel:** We need the VM’s kernel available. One approach is to **COPY the compiled `vmlinux` into the image** (e.g., `COPY --from=kernel-builder /path/to/vmlinux /boot/vmlinux`). If we built the kernel in a separate stage or earlier step, we can include it in the Docker build context. By placing `vmlinux` in `/boot` or another known location, we allow our launcher (or Slicer) to retrieve it. Slicer images likely bundle the kernel in the container image itself (their image tags include the kernel version, implying the kernel binary is part of the image) – when using Slicer, it may extract the kernel from the image. Including it ensures the image is self-contained. We’ll also include `/lib/modules/<kernel-version>` if we have any modules, so that the guest OS has them available (e.g., if we built VFIO or ext4 as modules, though in our case we try to have a mostly monolithic kernel to avoid needing an initramfs).

- **Resulting Artifact:** The Docker build will produce an OCI image (in our case, effectively a container image with Ubuntu 24.04 + systemd + our config). We are treating this Docker image as a **packaging format for the VM rootfs**. Slicer’s documentation confirms that their images are built with Docker and meant to be extended or used as base layers:contentReference[oaicite:20]{index=20}. We’re following the same pattern, except we started from scratch (Ubuntu base) rather than Slicer’s pre-made image. For example, if we wanted to add Nginx in this image later, we could do so with a Dockerfile RUN instruction and enabling its service:contentReference[oaicite:21]{index=21}.

- **Export for Testing (optional):** To test locally, we can export the filesystem to an ext4 file:
  - Run the built image in a container, mount a blank disk file, and copy the filesystem, as Julia Evans describes:contentReference[oaicite:22]{index=22}. E.g., using a script: create an empty file (say 1 GiB), format with ext4, mount it, then `docker cp` the container’s `/.` into it:contentReference[oaicite:23]{index=23}:contentReference[oaicite:24]{index=24}. After that, the `rootfs.ext4` file can be used with Firecracker directly. This step can be done in a dev environment to validate the VM boots with our image before pushing to registry. However, when using Slicer, it may not require us to manually create an ext4—Slicer can pull the OCI image and handle conversion to a block device internally. Still, having the ext4 is useful for manual Firecracker runs or integration tests.

## Phase 4: Set Up GitHub Actions CI Pipeline

We want the build process to be reproducible and automated. We’ll use **GitHub Actions** to compile the kernel and build the image, using a Makefile to orchestrate tasks where appropriate:

- **Repository Structure:** In a GitHub repo (private or internal to our org), we will store:
  - The Dockerfile for the base image (e.g., `Dockerfile.base`).
  - Optionally a Dockerfile for building the kernel in a container (or we just use Makefile + Actions steps on the runner).
  - A Makefile with targets like `kernel`, `image`, `push`, etc., to allow local builds and to document the process.
  - Kernel config file (`firecracker_kernel.config`) if we want to version control our kernel configuration.
  - Scripts if needed (maybe a script to assist in image export or testing).

- **GitHub Actions Workflow:** We create `.github/workflows/build.yml` (for example) with jobs to build and publish:
  - Use an Ubuntu runner (which is x86_64) for compatibility.
  - **Step 1: Checkout code.**
  - **Step 2: Set up dependencies** – e.g., install packages needed for kernel build (`sudo apt-get install build-essential libssl-dev bc flex bison libncurses-dev qemu-utils` etc. on the runner). Alternatively, use a Docker container with these tools.
  - **Step 3: Build Kernel** – run the Makefile target or direct commands:
    - `make olddefconfig && make -j$(nproc) vmlinux` using our config. This produces `vmlinux`.
    - Save the `vmlinux` artifact (and possibly modules). We might copy `vmlinux` into the repository workspace so Docker can access it. Ensure it’s not compressed (if `make` produces `bzImage`, we use `vmlinux` target specifically as noted).
  - **Step 4: Build OCI Image** – use Docker or Buildx:
    - If using Docker CLI: `docker build -f Dockerfile.base -t ghcr.io/<org>/firecracker-base:5.10.240-x86_64 .` (tag includes kernel version and arch for clarity:contentReference[oaicite:25]{index=25}).
    - This build will COPY the `vmlinux` from the workspace into the image, and run the steps to install packages, etc.
    - After build, log in to GHCR (use `GITHUB_TOKEN` or a PAT with `packages:write` scope, configured in Actions secrets) and push the image: `docker push ghcr.io/<org>/firecracker-base:5.10.240-x86_64`.
    - We could also use the GitHub Actions **Build-Push action** for a streamlined process (which handles buildx and multi-arch if needed). For now, x86_64 is our target, but we could extend to arm64 later by cross-compiling (Ubuntu 24.04 arm64 rootfs and an arm64 kernel with similar steps).
  - **Step 5: (Optional) Test Run** – If we have the ability, we might try to boot the image in a Firecracker instance within CI. However, GitHub hosted runners likely do not allow KVM (for Firecracker) and running Firecracker in CI is non-trivial (needs nested virt). We might skip automated boot testing and instead rely on manual testing, or run a basic qemu check. Alternatively, as a lightweight check, we could run the container and ensure systemd is installed and the expected files exist (not a full boot, but a sanity check).
  - **Step 6: Publish** – Mark the GHCR image as `latest` or a specific version tag. We will also consider versioning the image (for example, include a date or git commit in the tag for unique identification, or use semantic versions for our image). Slicer uses `latest` tag for convenience:contentReference[oaicite:26]{index=26}, but since this is internal, we might use explicit version tags to avoid surprises when deploying.

- **Makefile usage:** The Makefile might define:
  - `make kernel` (compiles kernel with config),
  - `make image` (builds the Docker image, perhaps depends on kernel target),
  - `make push` (pushes to registry).
  This abstracts away the raw commands, and the GitHub Actions can simply call `make push` after setting up environment, making the process easier to maintain.

- **Continuous Updates:** Schedule this workflow to run periodically (maybe monthly) to pull in security updates for Ubuntu packages. Since the base image is minimal, updates are mainly for critical libraries or kernel patches. This ensures our image stays up to date. Each run could push a new tag (like `5.10.240-x86_64-<date>`). In a private registry, storage of multiple versions is fine, and one can clean up older images as needed.

## Phase 5: Publish Image to GitHub Container Registry (Private)

Now that the GH Actions build pushes the image, we have our base image stored in GHCR:

- **Private Registry Configuration:** The image is in a private GHCR repository (e.g. `ghcr.io/<org>/firecracker-base:5.10.240-x86_64`). By default, GHCR images pushed from GitHub Actions in your repo are private unless you configure them otherwise. We’ll keep it private since this is an internal base image. To use it, the host that will run Firecracker/Slicer needs credentials to pull from GHCR. Typically, you can create a fine-grained PAT or use the GitHub CLI to auth. Alternatively, if running Slicer on the same GitHub org’s infrastructure, you might use the GitHub Actions runner’s token. But since Slicer likely runs on your own hardware, you’ll do a `docker login ghcr.io` on that machine with a username and PAT that has read access to the image.

- **Image Tagging Scheme:** We’ve tagged the image including the kernel version and OS:
  - e.g. `firecracker-base:5.10.240-x86_64` (similar to Slicer’s `slicer-systemd-2404:5.10.240-x86_64-latest` naming:contentReference[oaicite:27]{index=27}). This makes it clear which kernel and arch it’s for. We may add `-latest` for convenience on the main tag. For example, after testing, we might push also `:latest` pointing to this tag for ease of reference.
  - If we build a Cloud Hypervisor/GPU-enabled variant later, we could tag it distinctly (Slicer uses `*-ch` in the tag for Cloud Hypervisor images:contentReference[oaicite:28]{index=28}). For instance, `firecracker-base-ch:5.10.240-x86_64` for a CH variant.

- **Verification of Image in Registry:** We should verify the image manifest in GHCR contains what we expect (it should have the layers of the rootfs). One of those layers likely includes everything (since we started `FROM ubuntu:24.04`, many base packages are there). Our custom additions (systemd, etc.) are layered on top. The `vmlinux` file is baked into the image layers as well. So everything needed to instantiate the VM is contained in this image.

- **Size Consideration:** The base image might be on the order of a few hundred MB (Ubuntu base ~30MB, plus systemd ~50MB, plus other packages, and the kernel ~10MB, etc.). This is acceptable, though we aim to minimize it. We did not include any large AI frameworks yet; those would be added in specialized images later. By keeping this base small, pulling it is faster and using it as a clone source (Slicer uses copy-on-write cloning) is efficient.

- **Access Control:** The GHCR image is private, so ensure that any system or pipeline that needs it is authenticated. If using Slicer, you might need to configure Slicer to use credentials when pulling the image. Slicer’s docs mention the option to allow an insecure registry for local development:contentReference[oaicite:29]{index=29}, but in production one would use secure registry with auth. We will plan to either:
  - Pre-pull the image onto the host (using `docker pull` with auth) and let Slicer use it from local cache.
  - Or configure Slicer’s YAML to include registry credentials (if supported) or rely on the host’s Docker daemon credentials.
  - As a simpler approach, if the environment is closed, one could temporarily make the image public or use a self-hosted registry on the same network without auth:contentReference[oaicite:30]{index=30}. For our use, we’ll likely just handle the auth manually (login before running Slicer or build Slicer with the image imported).

## Phase 6: Testing the Custom MicroVM Image

With the kernel and rootfs image ready and published, thoroughly test it in a MicroVM to ensure it functions as expected:

- **Local Firecracker Test:** On a machine with KVM (and Firecracker installed), do a manual test:
  - Download/copy the `vmlinux` and `rootfs.ext4` (if you exported one) or use the OCI image.
  - Start Firecracker or use `firectl` with our kernel and rootfs:
    ```bash
    sudo firectl \
      --kernel=/path/to/vmlinux \
      --root-drive=/path/to/rootfs.ext4 \
      --kernel-opts="console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw" \
      --tap-device=tap0/AA:FC:00:00:00:01
    ``` 
    (This example uses typical Firecracker kernel options:contentReference[oaicite:31]{index=31} and sets up a TAP interface. You’d need to configure `tap0` on the host similarly to how Firecracker demos do:contentReference[oaicite:32]{index=32}, giving the VM an IP. The MAC here is arbitrary.)
  - Ensure the VM boots to login prompt (watch the Firecracker console output). It should initialize systemd (which might take a second or two). Check that you can log in (either via console login using the root password if set, or via SSH if network is up).
  - Test that systemd services are running: e.g., `ssh` is running (from host, try `ssh ubuntu@<vm-ip>` using an injected key or if we set a password for testing). Also check memory usage, etc.
  - If things fail (e.g. the VM doesn’t boot or cannot mount rootfs), debug accordingly (common issues: missing drivers in kernel for virtio, missing init, etc.). Our steps should prevent those (we included systemd and virtio support in kernel).

- **SlicerVM Integration Test:** If the aim is to use this with **Slicer**, perform these steps:
  1. Install and configure Slicer on the host (per Slicer documentation).
  2. In Slicer’s YAML config (which defines the VM image to use for slicing up), **point it to our custom image**. For example:
     ```yaml
     vm:
       image: ghcr.io/<org>/firecracker-base:5.10.240-x86_64-latest
       kernel_image: ""        # (if Slicer requires separate kernel spec, but likely it reads from image tag)
       # other Slicer configs like vcpu, memory, etc.
     ```
     Slicer’s docs indicate you can replace the `image:` field with your own image name:contentReference[oaicite:33]{index=33}.
  3. If the GHCR image is private, ensure Slicer can pull it (you might need to run Slicer within a Docker context that has auth, or use `slicer pull` commands if available). Alternatively, as a quick workaround, you could push the image to Docker Hub (private or public) and use that, since Slicer natively can pull from Docker Hub. In the example docs, Alex Ellis pushes to Docker Hub and references it:contentReference[oaicite:34]{index=34}.
  4. Run `slicer up` (or the equivalent command to start a VM or cluster). Slicer should fetch the image and create a MicroVM using Firecracker. Monitor the Slicer logs to see the VM booting. It should reach the “SSH accessible” state quickly (Slicer sets up networking and uses the Serial Over SSH if configured).
  5. Verify that both console and SSH access works via Slicer. Check that the `ubuntu` user’s authorized keys are in place. (Slicer can fetch your GitHub public key if you provided `github_user` in config, and insert it into `~ubuntu/.ssh/authorized_keys` on boot:contentReference[oaicite:35]{index=35} – ensure this mechanism works. It usually uses a cloud-init style user-data or a one-shot script.)
  6. Try launching a simple workload in the VM via Slicer (for example, Slicer’s one-shot task or just SSH in and run a command) to ensure the environment is functional.

- **Performance & Boot Time:** Measure the boot time. Firecracker VMs with systemd typically boot in ~1-2 seconds:contentReference[oaicite:36]{index=36}:contentReference[oaicite:37]{index=37}. Our minimal image should be in that range. Ensure no significant delays (if we accidentally left something like waiting on cloud-init or a generator). If boot is slow, use the console log to identify hang-ups (e.g., a service timing out).

- **Iterate if Needed:** If any issues are found (missing packages, misconfigurations), update the Dockerfile or kernel config accordingly and re-run the GitHub Actions workflow to rebuild and push a new image. Because this is automated, making a fix and getting a new image is straightforward. Just remember to update the image tag in Slicer or your Firecracker launch script to the new version if you don’t overwrite the old tag.

## Phase 7: Extend the Base Image for Trigger.dev Components (Web & Worker)

With a solid base image in place, we can create **custom images for each Trigger.dev component** on top of it. The Trigger.dev self-hosted architecture has (at least) two main services: the **web (UI/API server)** and the **worker/supervisor** process that executes tasks. We want to isolate these in separate microVMs. To do so, we’ll build two new images derived from our base:

- **Web VM Image:** This image will contain the Trigger.dev web application (likely a Node.js application or similar). Steps:
  - Create a `Dockerfile.web` using our base image as the parent:
    ```dockerfile
    FROM ghcr.io/<org>/firecracker-base:5.10.240-x86_64-latest
    # Install runtime dependencies, e.g., Node.js, and the web app code
    RUN apt-get update && apt-get install -y nodejs npm  # (or install from NodeSource)
    COPY ./web-app /opt/triggerdev/web  # copy application code (or binaries)
    RUN cd /opt/triggerdev/web && npm install --production  # install deps (if not containerized build)
    ```
    We might instead build the web app separately and just copy the built artifacts (to avoid compiling in the VM image). For simplicity, we can build inside or use a multi-stage Docker build.
  - Configure a **systemd service** to run the web server on startup. For example, create `/etc/systemd/system/trigger-web.service` with commands to start the web server (e.g., `ExecStart=/usr/bin/node /opt/triggerdev/web/dist/server.js`). Enable the service: `RUN systemctl enable trigger-web`.
    - If the web is just an HTTP server, ensure it listens on a proper interface (likely `0.0.0.0` and some port, e.g., 8080). Slicer can map VM ports to host, or you might use a reverse proxy on host. Initially, we can allow it to listen directly on the VM’s IP.
  - Expose necessary ports (if using Slicer, ensure the networking allows host to connect to the VM’s port 8080 or whatever the app uses).
  - After building this image, push it to GHCR (e.g., `firecracker-trigger-web:latest`). This image will inherently include everything from the base (so it has systemd, ssh, etc., plus Node and the app).

- **Worker VM Image:** Similarly, prepare `Dockerfile.worker`:
  - `FROM ghcr.io/<org>/firecracker-base:5.10.240-x86_64-latest`
  - Install runtime for the worker (which might also be Node.js if the worker is a Node process, or perhaps a separate service).
  - Copy the worker service code.
  - Install any needed libraries (for example, if the worker executes user workflows, it might need Docker or other tools; include them if necessary).
  - Create a systemd unit `trigger-worker.service` to start the worker on boot (running the command that launches the Trigger.dev worker process).
  - Enable the service (`systemctl enable trigger-worker`). Also ensure any dependency (maybe a Redis or DB connection details) are configured via environment variables or config files included.
  - Build and push this image as `firecracker-trigger-worker:latest` to GHCR.

- **Automation:** Extend the GitHub Actions workflow to also build these two images whenever code changes:
  - Possibly have separate jobs or a matrix for `role in [web, worker]` that:
    - Checks out code, (builds app if needed), 
    - Calls `docker build` for each Dockerfile, 
    - Tags and pushes images to GHCR.
  - We will likely trigger these builds when the application code or Dockerfiles change. The base image build (Phase 4) might be separate and triggered less often (only on base image changes or schedule).

- **Deployment:** Now we have:
  - Base image: `firecracker-base:...` (foundation).
  - Web image: `firecracker-trigger-web:...`.
  - Worker image: `firecracker-trigger-worker:...`.
  
  To deploy Trigger.dev in microVMs, we will:
  - Launch a VM using the web image for the web component.
  - Launch another VM using the worker image for the background tasks.
  Each VM will use the same kernel (the kernel baked in or referenced by the image tag). If using Slicer, we can define two VM profiles or just manually start two VMs with different images.
  
  For example, in Slicer’s YAML we might define two VM groups or simply run `slicer run ... --image ghcr.io/<org>/firecracker-trigger-web:latest` for one and similarly for the worker. If Slicer is more oriented to clusters of identical nodes, we may just treat them separately. Alternatively, if not using Slicer’s orchestration for this, a custom script could use Firecracker directly to start the two VMs: each with its kernel and rootfs (extracted from the respective images).
  
- **Networking Setup:** Ensure the two VMs can communicate if needed (the worker might need to talk to the web or vice versa, or both to a database). Possible setups:
  - **Shared bridge:** Attach both VMs’ network interfaces to a Linux bridge on the host (like the `firecracker0` bridge concept). They will then get IPs on the same subnet (via DHCP or static config) and can talk to each other. We might repurpose the Docker bridge or create a dedicated bridge as Julia Evans did:contentReference[oaicite:38]{index=38}. This way, the web VM can call the worker on an internal address if needed.
  - **Host access:** If the web UI is user-facing, we might forward its port to the host or use a reverse proxy. For instance, forward VM port 80 to host port 80. Firecracker doesn’t have built-in port forwarding, so one might rely on Slicer features or simply assign the VM an IP that the host can route to. Slicer likely sets up iptables DNAT for port forwarding if configured; otherwise, we can run an Nginx on host that proxies to the VM’s IP.
  - Document how this is done in deployment scripts so that all pieces can talk as required.

- **Validate Trigger.dev in VMs:** Finally, test the full setup:
  - Boot the web VM and worker VM.
  - Ensure the web UI is reachable (open the site, ensure it loads).
  - Ensure the worker is connected to the web (trigger some background job and see if the worker processes it).
  - Monitor logs via `journalctl` inside each VM (since systemd is running, our app logs will go there unless otherwise directed).
  - This gives us a working Trigger.dev deployment where each part is inside its own secure microVM. This isolation adds security (if one component crashes or is compromised, it doesn’t directly affect the other or the host) and can allow scaling each part separately by launching more microVMs if needed.

## Phase 8: Ongoing Maintenance and Enhancements

With the custom microVM images in place and working, we should plan for future needs and maintenance:

- **Security Updates:** Regularly rebuild the base image to pull in the latest security patches for Ubuntu 24.04. Our GH Actions can be set to run monthly as noted, or we do it when we get notifications of important CVEs. Because the images are versioned, we can test a new image in staging before switching production to use it. Ubuntu 24.04 is supported through 2029, so we’ll get updates for a while:contentReference[oaicite:39]{index=39}.
- **Upgrading Kernel:** Keep an eye on Firecracker’s supported kernel versions. We might upgrade to a newer LTS kernel (e.g., 5.15 or 6.1) for performance improvements or new features once we are comfortable. Test compatibility with Firecracker and the apps when doing so.
- **Deep Learning Stack:** If the goal is to support deep learning workloads *inside* these microVMs (for example, running AI inference or training in Trigger.dev tasks), consider creating specialized VM images for those tasks:
  - You might build an image that includes Python, CUDA (if GPUs are available), and frameworks like PyTorch or TensorFlow. This could be separate from the Trigger.dev core images; perhaps spun up on-demand for specific tasks. Since our pipeline and base are set, we can create more derivatives as needed.
  - If GPUs are involved, implement the Cloud Hypervisor path: compile a kernel with necessary GPU support and use CH to launch such VMs. According to Slicer, Firecracker is preferred for most cases and CH only when PCI devices (GPUs) are needed:contentReference[oaicite:40]{index=40}. So we could maintain two sets of images (Firecracker/CPU and CH/GPU). For example, a `firecracker-base-gpu` image that has NVIDIA drivers installed and is used with CH on a host with an NVIDIA GPU (via VFIO passthrough). This is a complex task (requires matching driver versions, etc.), so it can be a future phase on its own.
- **Snapshot/Restore:** Firecracker supports snapshotting VMs. Down the line, to optimize cold start, we could snapshot a fully initialized VM and boot from that snapshot for nearly instant startup. Trigger.dev’s roadmap mentions sub-500ms cold starts with MicroVMs – likely using techniques like Firecracker snapshot/restore. We could experiment with taking a snapshot after systemd is booted and our app is running, then use that as a base for new VMs (this is advanced, but an interesting enhancement).
- **Monitoring and Logging:** Ensure we have a way to aggregate logs and metrics from inside the VMs. Since each VM is isolated, we might install an agent or use Prometheus node exporter, etc. Alternatively, leverage Slicer’s mechanisms or simply have the VMs push metrics to a central service. (This is outside the scope of building the image, but important operationally.)
- **Backups:** If the microVMs have any state (they typically shouldn’t, except maybe caches), ensure we have persistence as needed. In our case, Trigger.dev likely uses an external database for state, so the VMs themselves can be treated as ephemeral and easily replaceable.

By following this phased plan, we create a robust pipeline to build custom MicroVM images and use them just like Slicer’s official images. We’ve built a minimal Ubuntu 24.04 base with systemd (for compatibility and manageability), packaged it as a Docker/OCI image:contentReference[oaicite:41]{index=41}, and stored it in a private registry. This image can be **pulled down and run via Slicer exactly like the official images**, giving us the same ease of deployment but with our own customizations. Finally, we extended the base to containerize the Trigger.dev application into two MicroVMs (web and worker), achieving strong isolation and potentially better cold-start performance for the platform. Each step is automated and documented, ensuring that updating or modifying the images in the future is straightforward and reliable.

**Sources:**

- Firecracker documentation and community Q&A on building custom kernel and rootfs images:contentReference[oaicite:42]{index=42}:contentReference[oaicite:43]{index=43}  
- SlicerVM official docs on image creation and extension:contentReference[oaicite:44]{index=44}:contentReference[oaicite:45]{index=45}:contentReference[oaicite:46]{index=46}  
- Julia Evans’ blog on using Firecracker and building images with Docker export (for reference on methodology):contentReference[oaicite:47]{index=47}:contentReference[oaicite:48]{index=48}  
- Trigger.dev documentation (for context on microVM plans and architecture):contentReference[oaicite:49]{index=49}:contentReference[oaicite:50]{index=50}  
