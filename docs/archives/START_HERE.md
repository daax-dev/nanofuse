# NanoFuse - Start Here

**Status**: Work in progress. Nothing has successfully built and booted yet.

This document explains what you actually need to do to build the base image.

## What You're Trying to Do

Build a minimal Ubuntu 24.04 microVM image that runs in Firecracker with systemd, SSH, and networking.

## The Simple Path (Just Build the Image)

The base image lives in `./images/base/`. That's where all the actual build logic is.

```bash
# Go to the base image directory
cd ./images/base

# Build the image (creates rootfs.ext4, vmlinux kernel, manifest.json)
sudo make build

# Verify artifacts were created
ls -lh ./build/
```

That's it. If that works, you'll have:
- `./build/rootfs.ext4` - The filesystem
- `./build/vmlinux` - The kernel
- `./build/manifest.json` - Metadata

## Testing the Image (If Build Works)

Once the image builds successfully:

```bash
cd ./images/base

# Test it boots in Firecracker
sudo ./test-boot.sh ./build/vmlinux ./build/rootfs.ext4

# Or validate the artifacts
./validate-build.sh
```

## Cleaning Up

```bash
cd ./images/base

# Remove all build artifacts
./clean.sh              # User-writable only
sudo ./clean-sudo.sh    # Everything including root-owned files
```

## The Whole Project (Not Needed to Build Image)

The rest of the nanofuse project in root is:

- `test-build-and-boot.sh` - Tries to do a complete E2E test (includes daemon, VM creation, etc.) - hasn't worked yet
- `setup-service.sh` - Sets up the nanofused daemon (not needed to build the image)
- `clean-all.sh` - Cleans everything including daemon state
- `RELEASE.md` - Release process (not relevant for development)
- `NOW.md` - Status tracking
- `DONE.md` - Completed features list
- Other markdown files in `./docs/` - Supporting documentation

## Focus

**To get something working**, focus on just the base image:

1. `cd ./images/base`
2. Run `sudo make build`
3. Does it work? If not, fix it.
4. Once it works, run `./test-boot.sh`

Everything else is extra complexity that doesn't matter until the base image actually builds.

## Important Files

- `./images/base/Dockerfile` - Defines what goes in the image
- `./images/base/Makefile` - Orchestrates the build
- `./images/base/build.sh` - Does the actual work (called by make)
- `./images/base/README.md` - Full documentation for base image
- `./images/base/BUILD.md` - Build instructions
- `./images/base/TEST.md` - Testing instructions

## Why Nothing Works

Looking at the git history, there's been constant attempts to make various things work:
- Different kernel versions (5.10.204 vs 6.1.90)
- Different networking configurations
- Port forwarding issues
- SSH key issues
- etc.

The core issue is probably in the base image build itself. Fix that first, then worry about everything else.

## Next Steps

1. Try building: `cd ./images/base && sudo make build`
2. What error do you get?
3. Fix that specific error
4. Repeat until it works

Don't try to run test-build-and-boot.sh or setup services until step 1 works.
