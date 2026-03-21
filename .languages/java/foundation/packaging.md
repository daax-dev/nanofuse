# Java Packaging and Distribution - Complete Guide

## JAR (Java Archive) Packaging

### Basic JAR Creation
```bash
# Create JAR from compiled classes
jar cf myapp.jar -C target/classes .

# Create JAR with manifest file
jar cfm myapp.jar MANIFEST.MF -C target/classes .

# Create JAR with specific main class
jar cfe myapp.jar com.example.Main -C target/classes .

# View JAR contents
jar tf myapp.jar

# Extract JAR contents
jar xf myapp.jar

# Update JAR
jar uf myapp.jar NewClass.class

# Verbose output
jar cvf myapp.jar -C target/classes .
```

### Manifest File Configuration
```
# MANIFEST.MF
Manifest-Version: 1.0
Main-Class: com.example.Application
Class-Path: lib/dependency1.jar lib/dependency2.jar
Implementation-Title: My Application
Implementation-Version: 1.0.0
Implementation-Vendor: Example Corp
Built-By: build-system
Build-Jdk: 21.0.1
Created-By: Apache Maven 3.9.6
Specification-Title: Application Specification
Specification-Version: 1.0
Specification-Vendor: Example Corp
```

### Executable JAR with Maven
```xml
<!-- Maven Shade Plugin for Fat JAR -->
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-shade-plugin</artifactId>
    <version>3.5.1</version>
    <executions>
        <execution>
            <phase>package</phase>
            <goals>
                <goal>shade</goal>
            </goals>
            <configuration>
                <createDependencyReducedPom>false</createDependencyReducedPom>
                <transformers>
                    <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                        <mainClass>com.example.Application</mainClass>
                    </transformer>
                    <!-- Merge service provider files -->
                    <transformer implementation="org.apache.maven.plugins.shade.resource.ServicesResourceTransformer"/>
                    <!-- Handle Spring handlers -->
                    <transformer implementation="org.apache.maven.plugins.shade.resource.AppendingTransformer">
                        <resource>META-INF/spring.handlers</resource>
                    </transformer>
                    <transformer implementation="org.apache.maven.plugins.shade.resource.AppendingTransformer">
                        <resource>META-INF/spring.schemas</resource>
                    </transformer>
                </transformers>
                <filters>
                    <filter>
                        <!-- Exclude signature files -->
                        <artifact>*:*</artifact>
                        <excludes>
                            <exclude>META-INF/*.SF</exclude>
                            <exclude>META-INF/*.DSA</exclude>
                            <exclude>META-INF/*.RSA</exclude>
                        </excludes>
                    </filter>
                </filters>
            </configuration>
        </execution>
    </executions>
</plugin>
```

### Assembly Plugin for Custom Distributions
```xml
<!-- Maven Assembly Plugin -->
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-assembly-plugin</artifactId>
    <version>3.6.0</version>
    <configuration>
        <descriptors>
            <descriptor>src/assembly/distribution.xml</descriptor>
        </descriptors>
        <finalName>${project.artifactId}-${project.version}</finalName>
        <appendAssemblyId>false</appendAssemblyId>
    </configuration>
    <executions>
        <execution>
            <id>create-archive</id>
            <phase>package</phase>
            <goals>
                <goal>single</goal>
            </goals>
        </execution>
    </executions>
</plugin>
```

```xml
<!-- src/assembly/distribution.xml -->
<assembly xmlns="http://maven.apache.org/ASSEMBLY/2.1.1"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/ASSEMBLY/2.1.1
                              http://maven.apache.org/xsd/assembly-2.1.1.xsd">

    <id>distribution</id>
    <formats>
        <format>tar.gz</format>
        <format>zip</format>
    </formats>

    <includeBaseDirectory>true</includeBaseDirectory>
    <baseDirectory>${project.artifactId}-${project.version}</baseDirectory>

    <fileSets>
        <!-- Main JAR -->
        <fileSet>
            <directory>target</directory>
            <outputDirectory>lib</outputDirectory>
            <includes>
                <include>${project.artifactId}-${project.version}.jar</include>
            </includes>
        </fileSet>

        <!-- Dependencies -->
        <fileSet>
            <directory>target/dependency</directory>
            <outputDirectory>lib</outputDirectory>
            <includes>
                <include>*.jar</include>
            </includes>
        </fileSet>

        <!-- Configuration files -->
        <fileSet>
            <directory>src/main/config</directory>
            <outputDirectory>config</outputDirectory>
            <includes>
                <include>**/*</include>
            </includes>
        </fileSet>

        <!-- Scripts -->
        <fileSet>
            <directory>src/main/scripts</directory>
            <outputDirectory>bin</outputDirectory>
            <includes>
                <include>**/*.sh</include>
                <include>**/*.bat</include>
            </includes>
            <fileMode>0755</fileMode>
        </fileSet>

        <!-- Documentation -->
        <fileSet>
            <directory>.</directory>
            <outputDirectory>docs</outputDirectory>
            <includes>
                <include>README.md</include>
                <include>LICENSE</include>
                <include>CHANGELOG.md</include>
            </includes>
        </fileSet>
    </fileSets>

</assembly>
```

## Gradle Packaging

### Gradle Fat JAR (Shadow Plugin)
```groovy
plugins {
    id 'com.github.johnrengelman.shadow' version '8.1.1'
}

shadowJar {
    archiveBaseName = 'myapp'
    archiveClassifier = ''
    archiveVersion = ''

    // Merge service provider files
    mergeServiceFiles()

    // Transform package names to avoid conflicts
    relocate 'org.apache.commons', 'shadow.org.apache.commons'

    // Exclude files
    exclude 'META-INF/*.SF'
    exclude 'META-INF/*.DSA'
    exclude 'META-INF/*.RSA'

    manifest {
        attributes 'Main-Class': 'com.example.Application'
        attributes 'Implementation-Title': 'My Application'
        attributes 'Implementation-Version': project.version
    }
}

// Replace default jar task
jar {
    enabled = false
    dependsOn(shadowJar)
}
```

### Gradle Application Plugin
```groovy
plugins {
    id 'application'
}

application {
    mainClass = 'com.example.Application'
    applicationName = 'myapp'
    executableDir = 'bin'
}

distributions {
    main {
        distributionBaseName = 'myapp'
        contents {
            from('src/dist') {
                into '/'
            }
            from('config') {
                into 'config'
            }
            from('docs') {
                into 'docs'
            }
        }
    }
}

startScripts {
    // Customize startup scripts
    doLast {
        // Unix script customization
        def unixScript = file getUnixScript()
        unixScript.text = unixScript.text.replace(
            'DEFAULT_JVM_OPTS=""',
            'DEFAULT_JVM_OPTS="-Xms512m -Xmx2g"'
        )

        // Windows script customization
        def windowsScript = file getWindowsScript()
        windowsScript.text = windowsScript.text.replace(
            'set DEFAULT_JVM_OPTS=',
            'set DEFAULT_JVM_OPTS=-Xms512m -Xmx2g'
        )
    }
}
```

## Spring Boot Packaging

### Spring Boot Maven Plugin
```xml
<plugin>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-maven-plugin</artifactId>
    <version>3.2.1</version>
    <configuration>
        <mainClass>com.example.Application</mainClass>
        <layout>JAR</layout>

        <!-- Layered JARs for Docker -->
        <layers>
            <enabled>true</enabled>
        </layers>

        <!-- Build info -->
        <buildInfo>
            <time>${maven.build.timestamp}</time>
        </buildInfo>

        <!-- Executable JAR -->
        <executable>true</executable>

        <!-- JVM arguments -->
        <jvmArguments>
            -Xmx2048m -Xms1024m
        </jvmArguments>

        <!-- System properties -->
        <systemPropertyVariables>
            <spring.profiles.active>production</spring.profiles.active>
        </systemPropertyVariables>

        <!-- Exclude dependencies from fat JAR -->
        <excludes>
            <exclude>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-devtools</artifactId>
            </exclude>
        </excludes>
    </configuration>
    <executions>
        <execution>
            <goals>
                <goal>repackage</goal>
                <goal>build-info</goal>
            </goals>
        </execution>
    </executions>
</plugin>
```

### Spring Boot Layered JARs
```bash
# Extract layers from Spring Boot JAR
java -Djarmode=layertools -jar myapp.jar list
java -Djarmode=layertools -jar myapp.jar extract

# Layers structure:
# dependencies/
# spring-boot-loader/
# snapshot-dependencies/
# application/
```

### Custom Layer Configuration
```xml
<!-- Custom layers.xml -->
<layers xmlns="http://www.springframework.org/schema/boot/layers"
        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
        xsi:schemaLocation="http://www.springframework.org/schema/boot/layers
                           https://www.springframework.org/schema/boot/layers/layers-2.6.xsd">
    <application>
        <into layer="spring-boot-loader">
            <include>org/springframework/boot/loader/**</include>
        </into>
        <into layer="application" />
    </application>
    <dependencies>
        <into layer="dependencies">
            <include>*:*</include>
            <exclude>org.springframework:*</exclude>
        </into>
        <into layer="spring-framework">
            <include>org.springframework:*</include>
        </into>
    </dependencies>
    <layerOrder>
        <layer>dependencies</layer>
        <layer>spring-framework</layer>
        <layer>spring-boot-loader</layer>
        <layer>application</layer>
    </layerOrder>
</layers>
```

## WAR (Web Archive) Packaging

### Maven WAR Plugin
```xml
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-war-plugin</artifactId>
    <version>3.4.0</version>
    <configuration>
        <webResources>
            <resource>
                <directory>src/main/webapp</directory>
                <filtering>true</filtering>
                <includes>
                    <include>**/*.xml</include>
                    <include>**/*.properties</include>
                </includes>
            </resource>
        </webResources>

        <!-- Custom manifest -->
        <archive>
            <manifestEntries>
                <Built-By>Build System</Built-By>
                <Build-Time>${maven.build.timestamp}</Build-Time>
                <Implementation-Version>${project.version}</Implementation-Version>
            </manifestEntries>
        </archive>

        <!-- Overlay configuration -->
        <overlays>
            <overlay>
                <groupId>com.example</groupId>
                <artifactId>base-webapp</artifactId>
                <type>war</type>
            </overlay>
        </overlays>
    </configuration>
</plugin>
```

### Spring Boot WAR Deployment
```java
@SpringBootApplication
public class Application extends SpringBootServletInitializer {

    @Override
    protected SpringApplicationBuilder configure(SpringApplicationBuilder application) {
        return application.sources(Application.class);
    }

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}
```

## Module System Packaging (Java 9+)

### Module Descriptor
```java
// module-info.java
module com.example.myapp {
    requires java.base;
    requires java.logging;
    requires transitive java.sql;

    exports com.example.myapp.api;
    exports com.example.myapp.model to com.example.myapp.persistence;

    provides com.example.myapp.spi.PluginProvider
        with com.example.myapp.impl.DefaultPluginProvider;

    uses com.example.myapp.spi.PluginProvider;
}
```

### Modular JAR Creation
```bash
# Compile modular application
javac -d target/classes --module-path lib src/main/java/module-info.java src/main/java/com/example/**/*.java

# Create modular JAR
jar --create --file=target/myapp.jar --main-class=com.example.Application -C target/classes .

# Run modular application
java --module-path target:lib --module com.example.myapp/com.example.Application
```

### JLink Custom Runtime
```bash
# Create custom runtime with jlink
jlink --module-path $JAVA_HOME/jmods:target:lib \
      --add-modules com.example.myapp \
      --launcher myapp=com.example.myapp/com.example.Application \
      --output target/myapp-runtime \
      --compress=2 \
      --no-header-files \
      --no-man-pages

# Run custom runtime
target/myapp-runtime/bin/myapp
```

## Native Image with GraalVM

### GraalVM Native Image Configuration
```xml
<!-- GraalVM Native Maven Plugin -->
<plugin>
    <groupId>org.graalvm.buildtools</groupId>
    <artifactId>native-maven-plugin</artifactId>
    <version>0.9.28</version>
    <configuration>
        <mainClass>com.example.Application</mainClass>
        <imageName>myapp</imageName>
        <buildArgs>
            <buildArg>--no-fallback</buildArg>
            <buildArg>--report-unsupported-elements-at-runtime</buildArg>
            <buildArg>--allow-incomplete-classpath</buildArg>
            <buildArg>-H:+ReportExceptionStackTraces</buildArg>
            <buildArg>-H:IncludeResources=.*\.properties$</buildArg>
        </buildArgs>
    </configuration>
    <executions>
        <execution>
            <id>build-native</id>
            <goals>
                <goal>compile-no-fork</goal>
            </goals>
            <phase>package</phase>
        </execution>
    </executions>
</plugin>
```

### Native Image Configuration Files
```json
// META-INF/native-image/reflect-config.json
[
  {
    "name": "com.example.model.User",
    "allDeclaredConstructors": true,
    "allPublicConstructors": true,
    "allDeclaredMethods": true,
    "allPublicMethods": true,
    "allDeclaredFields": true,
    "allPublicFields": true
  }
]
```

```json
// META-INF/native-image/resource-config.json
{
  "resources": {
    "includes": [
      {"pattern": ".*\\.properties$"},
      {"pattern": ".*\\.xml$"},
      {"pattern": ".*\\.json$"}
    ]
  }
}
```

## Docker Packaging

### Multi-stage Dockerfile
```dockerfile
# Multi-stage build for Java applications
FROM maven:3.9.6-eclipse-temurin-21 AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY pom.xml .
COPY .mvn .mvn
COPY mvnw .

# Download dependencies
RUN ./mvnw dependency:go-offline -B

# Copy source and build
COPY src ./src
RUN ./mvnw clean package -DskipTests

# Production image
FROM eclipse-temurin:21-jre-alpine

# Create application user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy JAR from builder stage
COPY --from=builder /app/target/*.jar app.jar

# Install required packages
RUN apk add --no-cache curl

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/actuator/health || exit 1

# Expose port
EXPOSE 8080

# JVM optimization for containers
ENV JAVA_OPTS="-XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0"

# Run application
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### Layered Docker Build (Spring Boot)
```dockerfile
# Optimized Spring Boot Dockerfile with layers
FROM eclipse-temurin:21-jdk-alpine AS builder

WORKDIR /app
COPY . .
RUN ./mvnw clean package -DskipTests

# Extract layers
RUN java -Djarmode=layertools -jar target/*.jar extract

# Production image
FROM eclipse-temurin:21-jre-alpine

RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring

WORKDIR /app

# Copy layers in order of change frequency
COPY --from=builder app/dependencies/ ./
COPY --from=builder app/spring-boot-loader/ ./
COPY --from=builder app/snapshot-dependencies/ ./
COPY --from=builder app/application/ ./

EXPOSE 8080

ENTRYPOINT ["java", "org.springframework.boot.loader.JarLauncher"]
```

## RPM and DEB Package Creation

### Maven RPM Plugin
```xml
<plugin>
    <groupId>org.codehaus.mojo</groupId>
    <artifactId>rpm-maven-plugin</artifactId>
    <version>2.2.0</version>
    <configuration>
        <group>Applications/System</group>
        <description>My Java Application</description>
        <summary>My Java Application RPM Package</summary>
        <copyright>2024 Example Corp</copyright>
        <packager>Build System</packager>
        <prefix>/opt</prefix>
        <changelogFile>src/rpm/CHANGELOG</changelogFile>
        <defineStatements>
            <defineStatement>_unpackaged_files_terminate_build 0</defineStatement>
        </defineStatements>
        <mappings>
            <mapping>
                <directory>/opt/myapp/lib</directory>
                <filemode>644</filemode>
                <sources>
                    <source>
                        <location>target/${project.build.finalName}.jar</location>
                    </source>
                </sources>
            </mapping>
            <mapping>
                <directory>/opt/myapp/bin</directory>
                <filemode>755</filemode>
                <sources>
                    <source>
                        <location>src/main/scripts/myapp.sh</location>
                    </source>
                </sources>
            </mapping>
            <mapping>
                <directory>/etc/myapp</directory>
                <configuration>true</configuration>
                <filemode>644</filemode>
                <sources>
                    <source>
                        <location>src/main/config/application.properties</location>
                    </source>
                </sources>
            </mapping>
        </mappings>
        <preinstallScriptlet>
            <script>
                getent group myapp >/dev/null || groupadd -r myapp
                getent passwd myapp >/dev/null || useradd -r -g myapp -d /opt/myapp -s /sbin/nologin myapp
            </script>
        </preinstallScriptlet>
        <postinstallScriptlet>
            <script>
                chown -R myapp:myapp /opt/myapp
                systemctl daemon-reload
                systemctl enable myapp
            </script>
        </postinstallScriptlet>
    </configuration>
    <executions>
        <execution>
            <goals>
                <goal>rpm</goal>
            </goals>
        </execution>
    </executions>
</plugin>
```

## Distribution Scripts

### Universal Startup Script
```bash
#!/bin/bash
# Universal startup script for Java applications

# Application configuration
APP_NAME="myapp"
APP_JAR="myapp.jar"
APP_USER="myapp"
APP_HOME="/opt/myapp"
PID_FILE="/var/run/myapp.pid"
LOG_FILE="/var/log/myapp.log"

# Java configuration
JAVA_HOME="${JAVA_HOME:-/usr/lib/jvm/java-21}"
JAVA_OPTS="${JAVA_OPTS:--Xms512m -Xmx2g -XX:+UseG1GC}"

# Application arguments
APP_ARGS="${APP_ARGS:-}"

# Source configuration file if exists
[ -f "/etc/default/myapp" ] && source /etc/default/myapp

# Helper functions
is_running() {
    [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null
}

start() {
    if is_running; then
        echo "Application is already running (PID: $(cat $PID_FILE))"
        return 1
    fi

    echo "Starting $APP_NAME..."

    # Create directories
    mkdir -p "$(dirname $PID_FILE)"
    mkdir -p "$(dirname $LOG_FILE)"

    # Start application
    if [ "$APP_USER" != "$(whoami)" ]; then
        sudo -u "$APP_USER" "$JAVA_HOME/bin/java" \
            $JAVA_OPTS \
            -jar "$APP_HOME/$APP_JAR" \
            $APP_ARGS \
            >> "$LOG_FILE" 2>&1 &
    else
        "$JAVA_HOME/bin/java" \
            $JAVA_OPTS \
            -jar "$APP_HOME/$APP_JAR" \
            $APP_ARGS \
            >> "$LOG_FILE" 2>&1 &
    fi

    local pid=$!
    echo $pid > "$PID_FILE"

    # Wait for application to start
    local count=0
    while [ $count -lt 30 ]; do
        if kill -0 $pid 2>/dev/null; then
            echo "Started $APP_NAME (PID: $pid)"
            return 0
        fi
        sleep 1
        ((count++))
    done

    echo "Failed to start $APP_NAME"
    return 1
}

stop() {
    if ! is_running; then
        echo "Application is not running"
        return 1
    fi

    local pid=$(cat "$PID_FILE")
    echo "Stopping $APP_NAME (PID: $pid)..."

    # Graceful shutdown
    kill $pid

    # Wait for graceful shutdown
    local count=0
    while [ $count -lt 30 ]; do
        if ! kill -0 $pid 2>/dev/null; then
            rm -f "$PID_FILE"
            echo "Stopped $APP_NAME"
            return 0
        fi
        sleep 1
        ((count++))
    done

    # Force kill
    echo "Force killing $APP_NAME..."
    kill -9 $pid
    rm -f "$PID_FILE"
    echo "Force stopped $APP_NAME"
}

status() {
    if is_running; then
        echo "$APP_NAME is running (PID: $(cat $PID_FILE))"
        return 0
    else
        echo "$APP_NAME is not running"
        return 1
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop && sleep 2 && start
        ;;
    status)
        status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac

exit $?
```

This comprehensive packaging guide covers all major Java distribution formats and deployment scenarios, from simple JARs to complex enterprise distributions and container deployments.