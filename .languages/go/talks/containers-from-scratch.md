# Building Containers from Scratch in Go

> A talk by Liz Rice demonstrating how to build a container runtime in ~50 lines of Go code

## Overview

This talk demonstrates the fundamental concepts behind containers by building one from scratch using Go. It shows how containers are created using three main Linux concepts:
- **Namespaces**: Limit what a process can see
- **Chroot**: Change the root filesystem
- **Control Groups (cgroups)**: Limit resources a process can use

## Key Concepts

### 1. Namespaces

Namespaces isolate and virtualize system resources for a collection of processes. Linux provides several namespace types:

- **UTS (Unix Time Sharing)**: Isolates hostname and domain name
- **PID**: Isolates process IDs
- **Mount**: Isolates mount points
- **Network**: Isolates network interfaces
- **IPC**: Isolates inter-process communication
- **User**: Isolates user and group IDs

### 2. Process Isolation

The implementation creates a parent-child process structure:
1. Parent process (`run`) creates new namespaces
2. Child process (`child`) executes within those namespaces
3. Uses `clone` syscall with flags like `CLONE_NEWUTS` and `CLONE_NEWPID`

### 3. Filesystem Isolation

- Uses `chroot` to change the root directory for the container
- Mounts `/proc` as a pseudo-filesystem for process information
- Creates isolated view of the filesystem using a pre-prepared root filesystem (like Ubuntu)

### 4. Control Groups (cgroups)

Control groups limit resources that containers can use:
- Memory limits
- CPU usage
- I/O bandwidth
- Process count limits

Example demonstrated: Creating a cgroup that limits processes to 20, protecting against fork bombs.

## Implementation Steps

### Basic Container Structure

```go
func main() {
    switch os.Args[1] {
    case "run":
        run()
    case "child":
        child()
    default:
        panic("bad command")
    }
}

func run() {
    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
        Unshareflags: syscall.CLONE_NEWNS,
    }

    must(cmd.Run())
}

func child() {
    // Set hostname
    syscall.Sethostname([]byte("container"))

    // Change root filesystem
    syscall.Chroot("/home/rootfs")
    syscall.Chdir("/")

    // Mount proc
    syscall.Mount("proc", "proc", "proc", 0, "")

    // Run command
    cmd := exec.Command(os.Args[2], os.Args[3:]...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    must(cmd.Run())

    // Cleanup
    syscall.Unmount("proc", 0)
}
```

### Setting up cgroups

```go
func cg() {
    cgroups := "/sys/fs/cgroup/"
    pids := filepath.Join(cgroups, "pids")

    // Create new cgroup
    os.Mkdir(filepath.Join(pids, "liz"), 0755)

    // Set max processes to 20
    must(ioutil.WriteFile(filepath.Join(pids, "liz/pids.max"), []byte("20"), 0700))

    // Add current process to cgroup
    must(ioutil.WriteFile(filepath.Join(pids, "liz/cgroup.procs"),
        []byte(strconv.Itoa(os.Getpid())), 0700))
}
```

## Key Takeaways

1. **Containers are Linux processes** - They're not VMs, just isolated processes
2. **Namespaces provide isolation** - Different views of the system for different processes
3. **chroot limits filesystem view** - Container sees only its designated root
4. **cgroups limit resources** - Prevent containers from consuming all system resources
5. **Container images are just filesystems** - Packaged filesystem that becomes the container's root

## Security Considerations

- Containers share the kernel - kernel exploits can affect all containers
- Need proper vulnerability scanning of container images
- Root in container can still be dangerous without proper restrictions
- Consider running containers in VMs for additional isolation in multi-tenant environments

## Production Considerations

While this demonstration shows the core concepts, production container runtimes like Docker add:
- Network configuration and isolation
- Storage drivers and volume management
- Image layer management and caching
- Security features (AppArmor, SELinux, seccomp)
- Resource monitoring and logging
- Container orchestration integration

## Resources

- GitHub: Container implementation examples
- Container vulnerability scanning tools
- Understanding the difference between containers and VMs
- Linux kernel documentation on namespaces and cgroups