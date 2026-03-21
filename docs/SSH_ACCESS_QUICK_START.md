# SSH Access Quick Start

**TL;DR**: Add your SSH public key to the Dockerfile, rebuild the image, then SSH normally.

---

## Current State

The base image comes with:
- ✅ SSH server enabled and running
- ✅ Root login allowed (key-only, no passwords)
- ❌ **No authorized_keys** - image is empty by default
- ❌ **You cannot SSH in yet** without adding your public key

---

## Step 1: Get Your SSH Public Key

### If you already have SSH keys

```bash
cat ~/.ssh/id_rsa.pub
```

Should output something like:
```
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7VJT...xYRZmQp/K7B8= user@laptop
```

### If you don't have SSH keys, generate them

```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa

# Press enter for passphrase (or set one)
# You'll get:
# - ~/.ssh/id_rsa (private key)
# - ~/.ssh/id_rsa.pub (public key)
```

Get the public key:
```bash
cat ~/.ssh/id_rsa.pub
```

---

## Step 2: Add Your Public Key to Dockerfile

Edit `images/base/Dockerfile`:

```dockerfile
# Configure SSH for security and access
# - Permit root login with key only (no password)
# - Add your SSH public key for access
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7VJT...xYRZmQp/K7B8= user@laptop" > /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
```

**Replace the long string with YOUR public key from Step 1.**

Example:
```dockerfile
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7VJT/uvtSTfW+9bx3LMWMRn3oDjOqEk/8tIKr9Jy2cvcEmTFJxYQVJL/RBDnNVXzYl1mJjbVd8XCw2u1Xrd2Vl+r5Jp7kXMhx2Ql4Qs5Y6QZ9pXmZj/u7X8X2Y3Z6X7Q8Y9Z2A3 user@laptop" > /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
```

---

## Step 3: Rebuild the Image

```bash
cd images/base
sudo make build

# Takes 2-3 minutes
# Output should show:
# ✓ Docker image built: nanofuse-base:latest
# ✓ Rootfs extracted
# ✓ Kernel downloaded
# ✓ Manifest generated
```

---

## Step 4: Create and Start a VM

```bash
# Create VM
./bin/nanofuse vm create default test-vm

# Start VM
./bin/nanofuse vm start test-vm

# Watch it boot (should take < 2 seconds)
./bin/nanofuse vm logs test-vm --follow
```

---

## Step 5: SSH Into the VM

```bash
# Get VM's IP address
VM_IP=$(./bin/nanofuse vm inspect test-vm | grep -i "ip" | head -1 | awk '{print $NF}')
echo "VM IP: $VM_IP"

# SSH in (should work without password prompt)
ssh root@$VM_IP
```

If successful, you'll see:
```bash
root@nanofuse-vm:~#
```

---

## Troubleshooting

### "Permission denied (publickey)"

**Cause**: Your public key isn't in the image

**Solution**:
1. Check your public key in Dockerfile is correct (no truncation)
2. Check SSH config: `./bin/nanofuse vm exec test-vm cat /root/.ssh/authorized_keys`
3. Verify permissions: `./bin/nanofuse vm exec test-vm ls -la /root/.ssh/`
4. Rebuild and try again

### "Connection refused"

**Cause**: SSH daemon isn't running

**Solution**:
```bash
# Check SSH status
./bin/nanofuse vm exec test-vm systemctl status ssh

# Restart if needed
./bin/nanofuse vm exec test-vm systemctl restart ssh

# Check it's listening on port 22
./bin/nanofuse vm exec test-vm ss -tulpn | grep 22
```

### "Could not resolve hostname"

**Cause**: VM IP address not reachable

**Solution**:
```bash
# Verify VM is running
./bin/nanofuse vm status test-vm

# Check it has an IP
./bin/nanofuse vm exec test-vm ip addr

# Ping it
ping <vm-ip>

# Check network bridge
ip addr show nanofuse0
```

### SSH says "Remote host identification has changed!"

**Cause**: You're connecting to a different VM with a different key but same IP

**Solution**:
```bash
# Remove old host key
ssh-keygen -R <vm-ip>

# Try again
ssh root@<vm-ip>
```

---

## Multiple Keys

If you have multiple keys, specify which one to use:

```bash
# Use specific key
ssh -i ~/.ssh/id_rsa root@<vm-ip>

# Or add all keys to ssh-agent
ssh-add ~/.ssh/id_rsa
ssh root@<vm-ip>
```

---

## For Multiple Users

If you have multiple people accessing the VM, add multiple keys:

```dockerfile
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    echo "ssh-rsa AAAA... user1@laptop1" >> /root/.ssh/authorized_keys && \
    echo "ssh-rsa AAAA... user2@laptop2" >> /root/.ssh/authorized_keys && \
    echo "ssh-rsa AAAA... user3@laptop3" >> /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys
```

Each line adds another key.

---

## Advanced: Runtime Key Injection (Future)

Currently, keys must be baked into the image. Future versions will support:

```bash
# (Not yet implemented)
./bin/nanofuse vm create default test-vm --ssh-key ~/.ssh/id_rsa.pub

# Or
./bin/nanofuse vm inject-key test-vm --key ~/.ssh/id_rsa.pub
```

For now, edit the Dockerfile.

---

## Security Notes

### ✅ Good Practices

- Use SSH keys, not passwords
- Keep private keys (id_rsa) secret
- Different keys for different environments
- Rotate keys periodically

### ❌ Bad Practices

- Don't share SSH private keys
- Don't commit Dockerfile with real keys to git
- Don't use default/test keys in production
- Don't enable password authentication

### For Development vs Production

**Development**: Hardcode your key in Dockerfile
```dockerfile
RUN echo "ssh-rsa AAAA..." > /root/.ssh/authorized_keys
```

**Production**: Inject keys at runtime (future feature)
```bash
./bin/nanofuse vm create image --inject-key ops-team.pub
```

---

## Quick Reference

```bash
# Generate new key
ssh-keygen -t rsa -b 4096

# Get public key
cat ~/.ssh/id_rsa.pub

# Edit Dockerfile to add key
nano images/base/Dockerfile

# Rebuild image
cd images/base && sudo make build

# Create VM
./bin/nanofuse vm create default vm1

# Start VM
./bin/nanofuse vm start vm1

# Get IP
./bin/nanofuse vm inspect vm1 | grep ip

# SSH in
ssh root@<vm-ip>

# SSH with specific key
ssh -i ~/.ssh/id_rsa root@<vm-ip>
```

---

## See Also

- [SSH_HOST_KEYS_ISSUE.md](SSH_HOST_KEYS_ISSUE.md) - Why each VM gets unique host keys
- [FIRECRACKER_IMAGE_BUILD_GUIDE.md](FIRECRACKER_IMAGE_BUILD_GUIDE.md) - Full build documentation
- [images/base/Dockerfile](../images/base/Dockerfile) - Where to add your key
