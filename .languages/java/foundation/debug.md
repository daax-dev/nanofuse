# Java Debugging - Comprehensive Guide

## IDE-Based Debugging

### IntelliJ IDEA Debugging
```java
public class DebuggingExample {

    public static void main(String[] args) {
        DebuggingExample example = new DebuggingExample();

        // Set breakpoint here (Ctrl+F8)
        int result = example.calculateSum(10, 20);

        // Conditional breakpoint example
        for (int i = 0; i < 100; i++) {
            if (i % 10 == 0) {
                System.out.println("Processing: " + i);  // Breakpoint with condition: i == 50
            }
        }

        System.out.println("Result: " + result);
    }

    public int calculateSum(int a, int b) {
        // Step into this method (F7)
        int sum = a + b;

        // Evaluate expression here (Alt+F8)
        // Try: sum * 2, Math.pow(sum, 2)

        return sum;
    }
}
```

#### IntelliJ IDEA Debug Features
- **F8**: Step Over - Execute current line, don't step into method calls
- **F7**: Step Into - Step into method calls
- **Shift+F8**: Step Out - Complete current method and return to caller
- **Alt+F9**: Run to Cursor - Run until cursor position
- **Ctrl+F8**: Toggle Breakpoint
- **Ctrl+Shift+F8**: View/Edit Breakpoints
- **Alt+F8**: Evaluate Expression
- **F9**: Resume Program

### Eclipse Debugging
```java
// Debug configuration in Eclipse
public class EclipseDebugExample {

    public void debugMethod() {
        String message = "Debug message";

        // Right-click line number to set breakpoint
        System.out.println(message);

        // Use Debug perspective (Window -> Perspective -> Debug)
        // Variables view shows local variables and their values
        // Expressions view for custom expressions
        // Debug view shows call stack
    }
}
```

## Command Line Debugging with JDB

### Basic JDB Usage
```bash
# Compile with debug information
javac -g MyClass.java

# Run with debug agent
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005 MyClass

# Connect with JDB
jdb -attach localhost:5005

# JDB Commands
# help - Show available commands
# stop at MyClass:10 - Set breakpoint at line 10
# stop in MyClass.methodName - Set breakpoint at method entry
# run - Start execution
# cont - Continue execution
# step - Step to next line
# step up - Step out of current method
# next - Step over method calls
# list - Show current source lines
# print variableName - Print variable value
# locals - Show local variables
# where - Show call stack
# quit - Exit debugger
```

## Remote Debugging

### JVM Debug Options
```bash
# Java 9+ (recommended)
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 MyApp

# Java 8 and earlier
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005 MyApp

# Alternative syntax (older)
java -Xdebug -Xrunjdwp:transport=dt_socket,server=y,suspend=n,address=5005 MyApp

# Debug options explained:
# transport=dt_socket - Use TCP sockets
# server=y - JVM listens for debugger connection
# suspend=n - Don't wait for debugger (y = wait)
# address=*:5005 - Listen on all interfaces, port 5005
```

### Docker Remote Debugging
```dockerfile
# Dockerfile for debugging
FROM openjdk:21-jdk

COPY app.jar /app.jar

# Expose debug port
EXPOSE 8080 5005

# Run with debug options
CMD ["java", "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005", "-jar", "/app.jar"]
```

```bash
# Run Docker container with debug port
docker run -p 8080:8080 -p 5005:5005 myapp:latest

# Connect IDE debugger to localhost:5005
```

### Kubernetes Debugging
```yaml
# k8s-debug.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app-debug
spec:
  replicas: 1
  selector:
    matchLabels:
      app: java-app
  template:
    metadata:
      labels:
        app: java-app
    spec:
      containers:
      - name: java-app
        image: myapp:latest
        ports:
        - containerPort: 8080
        - containerPort: 5005
        env:
        - name: JAVA_OPTS
          value: "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005"
---
apiVersion: v1
kind: Service
metadata:
  name: java-app-debug-service
spec:
  selector:
    app: java-app
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: debug
    port: 5005
    targetPort: 5005
  type: LoadBalancer
```

## Profiling and Performance Debugging

### Java Flight Recorder (JFR)
```bash
# Start application with JFR
java -XX:+FlightRecorder \
     -XX:StartFlightRecording=duration=60s,filename=app-profile.jfr \
     -jar myapp.jar

# Start recording on running JVM
jcmd <pid> JFR.start duration=30s filename=runtime-profile.jfr

# View available recordings
jcmd <pid> JFR.check

# Dump recording
jcmd <pid> JFR.dump filename=current-state.jfr

# Stop recording
jcmd <pid> JFR.stop
```

### JFR Custom Events
```java
import jdk.jfr.*;

@Name("com.example.UserLogin")
@Label("User Login Event")
@Category("Application")
public class UserLoginEvent extends Event {

    @Label("User ID")
    String userId;

    @Label("Login Duration")
    @Timespan(Timespan.MILLISECONDS)
    long duration;

    @Label("Success")
    boolean success;
}

// Usage
public class UserService {

    public boolean login(String userId, String password) {
        UserLoginEvent event = new UserLoginEvent();
        event.userId = userId;
        event.begin();

        try {
            boolean result = authenticateUser(userId, password);
            event.success = result;
            return result;
        } finally {
            event.end();
            event.commit();
        }
    }
}
```

### Memory Analysis with JVisualVM
```bash
# Start JVisualVM
jvisualvm

# Connect to running JVM
# PID-based connection automatically detected
# Remote connection: Add remote host

# Heap dump analysis
jcmd <pid> GC.run_finalization
jcmd <pid> GC.run
jmap -dump:live,format=b,file=heap-dump.hprof <pid>

# Analyze heap dump with Eclipse MAT
# Install Eclipse Memory Analyzer Tool
# Open heap-dump.hprof file
```

### GC Analysis
```bash
# Enable GC logging (Java 11+)
java -Xlog:gc*:gc.log:time,tags -jar myapp.jar

# Java 8 GC logging
java -Xloggc:gc.log -XX:+PrintGCDetails -XX:+PrintGCTimeStamps -jar myapp.jar

# Analyze GC logs with GCViewer
# Download GCViewer tool
java -jar gcviewer-1.36.jar gc.log
```

## Advanced Debugging Techniques

### Thread Debugging
```java
public class ThreadDebuggingExample {

    public static void main(String[] args) {
        // Create multiple threads for debugging
        Thread thread1 = new Thread(new Worker("Worker-1"), "Thread-1");
        Thread thread2 = new Thread(new Worker("Worker-2"), "Thread-2");

        thread1.start();
        thread2.start();

        // Generate thread dump
        generateThreadDump();
    }

    static class Worker implements Runnable {
        private String name;

        Worker(String name) {
            this.name = name;
        }

        @Override
        public void run() {
            for (int i = 0; i < 100; i++) {
                // Set breakpoint and examine different threads
                System.out.println(name + " - Count: " + i);

                try {
                    Thread.sleep(100);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }
        }
    }

    public static void generateThreadDump() {
        ThreadMXBean threadMX = ManagementFactory.getThreadMXBean();
        ThreadInfo[] threadInfos = threadMX.dumpAllThreads(true, true);

        for (ThreadInfo info : threadInfos) {
            System.out.println("Thread: " + info.getThreadName());
            System.out.println("State: " + info.getThreadState());

            StackTraceElement[] stack = info.getStackTrace();
            for (StackTraceElement element : stack) {
                System.out.println("  " + element.toString());
            }
            System.out.println();
        }
    }
}
```

### Deadlock Detection
```java
public class DeadlockDetection {
    private static final Object lock1 = new Object();
    private static final Object lock2 = new Object();

    public static void main(String[] args) {
        // Start deadlock detection
        startDeadlockMonitoring();

        // Create potential deadlock situation
        Thread t1 = new Thread(() -> {
            synchronized (lock1) {
                System.out.println("Thread 1: Locked resource 1");

                try { Thread.sleep(100); } catch (InterruptedException e) {}

                synchronized (lock2) {
                    System.out.println("Thread 1: Locked resource 2");
                }
            }
        });

        Thread t2 = new Thread(() -> {
            synchronized (lock2) {
                System.out.println("Thread 2: Locked resource 2");

                try { Thread.sleep(100); } catch (InterruptedException e) {}

                synchronized (lock1) {
                    System.out.println("Thread 2: Locked resource 1");
                }
            }
        });

        t1.start();
        t2.start();
    }

    private static void startDeadlockMonitoring() {
        ThreadMXBean threadBean = ManagementFactory.getThreadMXBean();

        Thread monitor = new Thread(() -> {
            while (true) {
                long[] deadlocked = threadBean.findDeadlockedThreads();
                if (deadlocked != null) {
                    ThreadInfo[] threadInfos = threadBean.getThreadInfo(deadlocked);
                    System.out.println("DEADLOCK DETECTED!");

                    for (ThreadInfo info : threadInfos) {
                        System.out.println("Deadlocked thread: " + info.getThreadName());
                        System.out.println("Blocked on: " + info.getLockName());
                        System.out.println("Lock owned by: " + info.getLockOwnerName());
                    }
                }

                try {
                    Thread.sleep(1000);
                } catch (InterruptedException e) {
                    break;
                }
            }
        });

        monitor.setDaemon(true);
        monitor.start();
    }
}
```

### Conditional Debugging
```java
public class ConditionalDebugging {
    private static final Logger logger = LoggerFactory.getLogger(ConditionalDebugging.class);

    public void processData(List<DataItem> items) {
        for (int i = 0; i < items.size(); i++) {
            DataItem item = items.get(i);

            // Conditional breakpoint: i == 50 || item.isSpecial()
            // Log point instead of breakpoint
            if (item.isSpecial()) {
                logger.debug("Processing special item at index {}: {}", i, item);
            }

            processItem(item);
        }
    }

    private void processItem(DataItem item) {
        // Exception breakpoint - break when specific exception occurs
        try {
            item.validate();
            item.process();
        } catch (ValidationException e) {
            // IntelliJ: Run -> View Breakpoints -> Exception Breakpoints
            // Add ValidationException to break when thrown
            logger.error("Validation failed for item: {}", item, e);
            throw e;
        }
    }

    static class DataItem {
        private String data;
        private boolean special;

        public boolean isSpecial() { return special; }
        public void validate() throws ValidationException {
            if (data == null) throw new ValidationException("Data is null");
        }
        public void process() { /* processing logic */ }

        @Override
        public String toString() { return "DataItem{data='" + data + "'}"; }
    }

    static class ValidationException extends RuntimeException {
        public ValidationException(String message) { super(message); }
    }
}
```

## Debugging Tools and Commands

### jstack - Thread Stack Traces
```bash
# Generate thread dump
jstack <pid> > thread-dump.txt

# Multiple dumps for comparison
for i in {1..3}; do
    jstack <pid> > thread-dump-$i.txt
    sleep 5
done

# Analyze thread dump for patterns
grep -A 10 "java.lang.Thread.State: BLOCKED" thread-dump.txt
```

### jmap - Memory Analysis
```bash
# Heap summary
jmap -heap <pid>

# Generate heap dump
jmap -dump:live,format=b,file=heap.hprof <pid>

# Object histogram
jmap -histo <pid> | head -20

# Class loader statistics
jmap -clstats <pid>
```

### jstat - JVM Statistics
```bash
# GC statistics
jstat -gc <pid> 5s 10    # Every 5 seconds, 10 times

# Memory utilization
jstat -gccapacity <pid>

# Compilation statistics
jstat -compiler <pid>

# Class loading statistics
jstat -class <pid>
```

### jcmd - Multi-purpose Tool
```bash
# List available commands
jcmd <pid> help

# VM flags
jcmd <pid> VM.flags

# System properties
jcmd <pid> VM.system_properties

# Class histogram
jcmd <pid> GC.class_histogram

# Thread dump
jcmd <pid> Thread.print

# GC run
jcmd <pid> GC.run
```

## Debugging Configuration Files

### Logback Configuration for Debugging
```xml
<!-- logback-debug.xml -->
<configuration debug="true">

    <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>

    <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>debug.log</file>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>debug.%d{yyyy-MM-dd}.%i.log</fileNamePattern>
            <maxFileSize>100MB</maxFileSize>
            <maxHistory>7</maxHistory>
            <totalSizeCap>1GB</totalSizeCap>
        </rollingPolicy>
        <encoder>
            <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>

    <!-- Application loggers -->
    <logger name="com.example" level="DEBUG" />

    <!-- Framework loggers -->
    <logger name="org.springframework" level="INFO" />
    <logger name="org.hibernate" level="DEBUG" />
    <logger name="org.hibernate.SQL" level="DEBUG" />
    <logger name="org.hibernate.type.descriptor.sql.BasicBinder" level="TRACE" />

    <!-- Root logger -->
    <root level="INFO">
        <appender-ref ref="CONSOLE" />
        <appender-ref ref="FILE" />
    </root>

</configuration>
```

### JVM Debug Startup Script
```bash
#!/bin/bash
# debug-start.sh

APP_JAR="myapp.jar"
MAIN_CLASS="com.example.Application"
DEBUG_PORT="5005"
JMX_PORT="9999"

# Debug options
DEBUG_OPTS="-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:${DEBUG_PORT}"

# JMX options
JMX_OPTS="-Dcom.sun.management.jmxremote \
          -Dcom.sun.management.jmxremote.port=${JMX_PORT} \
          -Dcom.sun.management.jmxremote.authenticate=false \
          -Dcom.sun.management.jmxremote.ssl=false"

# JFR options
JFR_OPTS="-XX:+FlightRecorder \
          -XX:StartFlightRecording=duration=0,filename=debug-session.jfr"

# Memory options
MEM_OPTS="-Xms512m -Xmx2g -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=./heapdump.hprof"

# GC logging
GC_OPTS="-Xlog:gc*:gc-debug.log:time,tags"

# Combine all options
JAVA_OPTS="${DEBUG_OPTS} ${JMX_OPTS} ${JFR_OPTS} ${MEM_OPTS} ${GC_OPTS}"

echo "Starting application with debug options..."
echo "Debug port: ${DEBUG_PORT}"
echo "JMX port: ${JMX_PORT}"

java ${JAVA_OPTS} -jar ${APP_JAR}
```

This comprehensive debugging guide provides tools and techniques for troubleshooting Java applications at all levels, from simple IDE debugging to advanced production diagnostics.