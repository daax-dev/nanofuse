Below is a short, battle-tested plan + concrete examples (so you can ship this in one sitting).

# 1) Recommended flow (fast, compatible with Slicer)

Start from Slicer’s base image for the right arch:

FROM ghcr.io/openfaasltd/slicer-systemd:5.10.240-x86_64-latest (x86_64) or the arm64 tag for aarch64. 
docs.slicervm.com

Add packages, systemd unit files, app files, etc., in a Dockerfile.

Important: don’t try to systemctl start during build — enable the unit(s) (systemctl enable) or add one-shot unit files that run on first boot. Slicer docs call this out explicitly. 
docs.slicervm.com

Build and push to a registry (GHCR / Docker Hub / internal registry).

docker build -t ghcr.io/yourorg/slicer-myimage:1.0 .
docker push ghcr.io/yourorg/slicer-myimage:1.0


Edit your vm-image.yaml and set image: "ghcr.io/yourorg/slicer-myimage:1.0" (or use insecure_registry: true for local registries). Then slicer up ./vm-image.yaml. Slicer will pull the image and use it as the VM rootfs (storage: image). 
docs.slicervm.com

Why this is the best path:

It’s the documented + supported method. Slicer builds its images the same way (Dockerfile layers). You get easy iteration (build/push) and simple YAML updates. 
docs.slicervm.com

# 2) Example minimal Dockerfile (systemd-based image)
FROM ghcr.io/openfaasltd/slicer-systemd:5.10.240-x86_64-latest

### noninteractive installs
ENV DEBIAN_FRONTEND=noninteractive

### add packages (example)
RUN apt-get update && \
    apt-get install -y nginx curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

### enable systemd service (do NOT start it in the Dockerfile)
RUN systemctl enable nginx.service

### add website files
COPY my-website /var/www/html

### add one-shot firstboot unit if you need to run commands on first boot
COPY my-firstboot.service /etc/systemd/system/my-firstboot.service
RUN systemctl enable my-firstboot.service

### do not change CMD (image expects systemd)


Notes: adapt packages and services for your needs. If something cannot run inside a Docker build (e.g., interactive steps), either put it into a one-shot systemd unit or run it via userdata in the Slicer YAML. 
docs.slicervm.com

# 3) If you must build a Firecracker kernel + rootfs from scratch (advanced)

Use this only if you need a completely different OS or very tiny specialized image. It’s harder and Slicer says “not recommended / not documented” for totally custom OSes — reach out to Slicer if you go that route. 
docs.slicervm.com

High-level steps:

Build (or grab) an uncompressed Linux kernel image that supports virtio (virtio_blk, virtio_net, virtio_pci, devtmpfs, ext4, etc.). Use Firecracker’s kernel config as a starting point.

Produce an ext4 rootfs image (debootstrap or docker export → chroot → create /etc/fstab, add /etc/hostname, SSH keys, systemd units, etc.).

Kernel cmdline example Firecracker commonly expects:

console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k


Boot/test locally with firecracker / firectl or firecracker-containerd before integrating with Slicer. There are numerous guides and the Firecracker repo/docs for rootfs & kernel. 
cocalc.com

## Caveats & gotchas if you go raw:

Newer kernels sometimes need special kernel cmdline/ACPI/virtio setup — people hit /dev/vda not found errors. Test kernels thoroughly. 
GitHub

You must include console on ttyS0 (Firecracker uses that) so you can see boot logs in Slicer. 
Medium

If you depend on PCI passthrough (GPU), Slicer says Cloud Hypervisor is required for mounting PCI devices; for most cases Firecracker is preferred. 
docs.slicervm.com

# 4) How to make it “pull and run exactly like Slicer images”

Publish the Docker image to a public/private registry.

Make sure the tag you publish matches the architecture (x86_64 vs arm64).

Put that repo:tag into vm-image.yaml’s image: field — Slicer will pull and use it as the VM image. Example YAML snippet is in Slicer docs. 
docs.slicervm.com

# 5) Bonus: CI / delivery pipeline

Add a GitHub Actions workflow to docker build on push and publish to GHCR. Use image tags like :latest and :sha-<short> for reproducibility.

Optionally publish a small README and a sample vm-image.yaml showing how to use the image with Slicer.

TL;DR (plain talk)

If you want the easiest, most maintainable path: extend the Slicer base via Dockerfile, push the container image, point vm-image.yaml at it. Building raw kernel+ext4 is doable and sometimes necessary, but it’s fiddly and much slower to iterate. Slicer explicitly supports the Dockerfile-extension model — so use it. 
docs.slicervm.com