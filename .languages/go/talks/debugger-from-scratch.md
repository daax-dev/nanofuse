# Building a Debugger from Scratch in Go

> A talk by Liz Rice demonstrating container concepts through building a minimal container runtime

## Overview

This talk parallels the "Containers from Scratch" presentation, showing how to build a container in approximately 52 lines of Go code. The demonstration reveals that containers are fundamentally composed of Linux namespaces and control groups (cgroups), not magical isolation technology.

## Core Container Components

### 1. Linux Namespaces
Namespaces limit what a containerized process can see:
- **UTS (Unix Time Sharing System)**: Isolates hostname
- **PID**: Process ID isolation
- **Mount**: Filesystem mount points
- **Network**: Network interfaces
- **IPC**: Inter-process communication
- **User**: User and group IDs

### 2. Filesystem Isolation
- **chroot**: Changes the root directory visible to the process
- **Mount namespaces**: Isolate mount points from the host
- **/proc filesystem**: Special pseudo-filesystem for process information

### 3. Control Groups (cgroups)
Limit resources that containers can consume:
- CPU usage
- Memory allocation
- I/O bandwidth
- Process count

## Implementation Walkthrough

### Basic Structure

The implementation follows a parent-child process model:

```go
func main() {
    switch os.Args[1] {
    case "run":
        run()  // Parent process
    case "child":
        child()  // Child process in new namespace
    }
}
```

### Creating Namespaces

```go
func run() {
    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

    // Set up namespaces
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
    }

    cmd.Run()
}
```

### Inside the Container

```go
func child() {
    // Set container hostname
    syscall.Sethostname([]byte("container"))

    // Change root filesystem
    syscall.Chroot("/home/rootfs")
    syscall.Chdir("/")

    // Mount proc filesystem
    syscall.Mount("proc", "proc", "proc", 0, "")

    // Execute the requested command
    cmd := exec.Command(os.Args[2], os.Args[3:]...)
    cmd.Run()
}
```

## Key Demonstrations

### Process Isolation

When running `ps` inside the container:
- Shows only container processes
- Process IDs start from 1
- Host processes are invisible

### Filesystem Isolation

- Container sees only its designated root filesystem
- Cannot access files above its root
- `/proc` must be mounted separately for tools like `ps` to work

### Resource Limits with cgroups

Example limiting process count to prevent fork bombs:
```go
// Limit to 20 processes
ioutil.WriteFile("/sys/fs/cgroup/pids/container/pids.max", []byte("20"), 0700)
```

## Container Images Explained

Container images are essentially:
1. Packaged filesystems (like the Ubuntu filesystem used in demo)
2. Configuration metadata (environment variables, default commands)
3. Layered for efficiency and reuse

## Key Insights

### What Containers Really Are
- **NOT lightweight VMs** - they're isolated Linux processes
- **NOT magic** - built on established Linux kernel features
- **Shared kernel** - all containers on a host share the same kernel

### Security Implications
- Kernel vulnerabilities affect all containers
- Root in container ≠ root on host (but still dangerous)
- Need additional security measures in production

### "It's Just Namespaces and cgroups"
This common phrase in the container community means:
- **Namespaces**: Provide isolation (what you can see)
- **cgroups**: Provide resource limits (what you can use)
- Everything else is tooling built on these primitives

## Production vs Demo

Production container runtimes add:
- Network configuration
- Storage drivers
- Image management
- Security policies (AppArmor, SELinux)
- OCI compliance
- Orchestration support

## Practical Applications

Understanding these fundamentals helps with:
- Debugging container issues
- Security assessments
- Performance optimization
- Choosing appropriate isolation levels
- Understanding container behavior

## Tools and Projects

- **microBadger**: Inspect Docker image layers
- Container security scanning
- Understanding Dockerfile efficiency
- Debugging containerized applications

## Takeaway Message

Containers aren't mysterious - they're elegant applications of Linux kernel features. In just 52 lines of Go, you can create a functional container that demonstrates all the core concepts. This demystification helps developers and operators work more effectively with container technology.