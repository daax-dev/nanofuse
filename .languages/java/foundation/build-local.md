# Java Local Build Toolchain - Complete Guide

## Maven Build System

### Project Structure and POM Configuration
```xml
<!-- pom.xml - Maven Project Object Model -->
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <!-- Project Coordinates -->
    <groupId>com.example</groupId>
    <artifactId>my-java-app</artifactId>
    <version>1.0.0-SNAPSHOT</version>
    <packaging>jar</packaging>

    <!-- Project Information -->
    <name>My Java Application</name>
    <description>A comprehensive Java application example</description>
    <url>https://github.com/example/my-java-app</url>

    <!-- Properties -->
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <junit.version>5.10.1</junit.version>
        <spring.version>6.1.2</spring.version>
        <maven.surefire.version>3.2.2</maven.surefire.version>
    </properties>

    <!-- Dependencies -->
    <dependencies>
        <!-- Spring Framework -->
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-context</artifactId>
            <version>${spring.version}</version>
        </dependency>

        <!-- Logging -->
        <dependency>
            <groupId>ch.qos.logback</groupId>
            <artifactId>logback-classic</artifactId>
            <version>1.4.14</version>
        </dependency>

        <!-- JSON Processing -->
        <dependency>
            <groupId>com.fasterxml.jackson.core</groupId>
            <artifactId>jackson-databind</artifactId>
            <version>2.16.1</version>
        </dependency>

        <!-- Test Dependencies -->
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>${junit.version}</version>
            <scope>test</scope>
        </dependency>

        <dependency>
            <groupId>org.mockito</groupId>
            <artifactId>mockito-core</artifactId>
            <version>5.8.0</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <!-- Build Configuration -->
    <build>
        <plugins>
            <!-- Compiler Plugin -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.12.1</version>
                <configuration>
                    <source>21</source>
                    <target>21</target>
                    <compilerArgs>
                        <arg>--enable-preview</arg>
                    </compilerArgs>
                </configuration>
            </plugin>

            <!-- Surefire Plugin (Unit Tests) -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>${maven.surefire.version}</version>
                <configuration>
                    <argLine>--enable-preview</argLine>
                </configuration>
            </plugin>

            <!-- Failsafe Plugin (Integration Tests) -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-failsafe-plugin</artifactId>
                <version>3.2.2</version>
                <configuration>
                    <argLine>--enable-preview</argLine>
                </configuration>
                <executions>
                    <execution>
                        <goals>
                            <goal>integration-test</goal>
                            <goal>verify</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>

            <!-- JaCoCo Plugin (Code Coverage) -->
            <plugin>
                <groupId>org.jacoco</groupId>
                <artifactId>jacoco-maven-plugin</artifactId>
                <version>0.8.11</version>
                <executions>
                    <execution>
                        <goals>
                            <goal>prepare-agent</goal>
                        </goals>
                    </execution>
                    <execution>
                        <id>report</id>
                        <phase>test</phase>
                        <goals>
                            <goal>report</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>

            <!-- SpotBugs Plugin (Static Analysis) -->
            <plugin>
                <groupId>com.github.spotbugs</groupId>
                <artifactId>spotbugs-maven-plugin</artifactId>
                <version>4.8.2.0</version>
                <configuration>
                    <effort>Max</effort>
                    <threshold>Low</threshold>
                </configuration>
            </plugin>

            <!-- Checkstyle Plugin -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-checkstyle-plugin</artifactId>
                <version>3.3.1</version>
                <configuration>
                    <configLocation>checkstyle.xml</configLocation>
                </configuration>
            </plugin>
        </plugins>
    </build>

    <!-- Profiles for different environments -->
    <profiles>
        <profile>
            <id>development</id>
            <properties>
                <environment>dev</environment>
            </properties>
            <activation>
                <activeByDefault>true</activeByDefault>
            </activation>
        </profile>

        <profile>
            <id>production</id>
            <properties>
                <environment>prod</environment>
            </properties>
            <build>
                <plugins>
                    <!-- Shade Plugin for Fat JAR -->
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
                                    </transformers>
                                </configuration>
                            </execution>
                        </executions>
                    </plugin>
                </plugins>
            </build>
        </profile>
    </profiles>
</project>
```

### Maven Commands and Lifecycle
```bash
# Basic Maven commands
mvn clean                    # Clean build artifacts
mvn compile                  # Compile main source code
mvn test-compile            # Compile test source code
mvn test                    # Run unit tests
mvn package                 # Create JAR/WAR file
mvn install                 # Install to local repository
mvn deploy                  # Deploy to remote repository

# Combined commands
mvn clean compile           # Clean and compile
mvn clean test             # Clean, compile, and test
mvn clean package          # Full build cycle
mvn clean install         # Build and install locally

# Skip tests
mvn clean package -DskipTests        # Skip test execution
mvn clean package -Dmaven.test.skip=true  # Skip test compilation and execution

# Profile activation
mvn clean package -Pproduction      # Activate production profile

# Dependency management
mvn dependency:tree                  # Show dependency tree
mvn dependency:analyze              # Analyze dependencies
mvn dependency:resolve              # Download dependencies

# Plugin execution
mvn spotbugs:check                  # Run SpotBugs analysis
mvn checkstyle:check               # Run Checkstyle
mvn jacoco:report                  # Generate coverage report

# Site generation
mvn site                           # Generate project site
```

## Gradle Build System

### Build Script Configuration
```groovy
// build.gradle
plugins {
    id 'java'
    id 'application'
    id 'jacoco'
    id 'com.github.spotbugs' version '6.0.4'
    id 'checkstyle'
    id 'org.springframework.boot' version '3.2.1'
    id 'io.spring.dependency-management' version '1.1.4'
}

// Project configuration
group = 'com.example'
version = '1.0.0-SNAPSHOT'
java.sourceCompatibility = JavaVersion.VERSION_21

// Application main class
application {
    mainClass = 'com.example.Application'
}

// Repositories
repositories {
    mavenCentral()
    gradlePluginPortal()
}

// Dependencies
dependencies {
    // Spring Boot
    implementation 'org.springframework.boot:spring-boot-starter'
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'

    // Database
    runtimeOnly 'com.h2database:h2'
    runtimeOnly 'org.postgresql:postgresql'

    // JSON Processing
    implementation 'com.fasterxml.jackson.core:jackson-databind'

    // Logging
    implementation 'ch.qos.logback:logback-classic'

    // Test dependencies
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testImplementation 'org.junit.jupiter:junit-jupiter'
    testImplementation 'org.mockito:mockito-core'
    testImplementation 'org.testcontainers:junit-jupiter'
    testImplementation 'org.testcontainers:postgresql'
}

// Java compilation configuration
java {
    withJavadocJar()
    withSourcesJar()
}

compileJava {
    options.compilerArgs += ['--enable-preview']
    options.encoding = 'UTF-8'
}

compileTestJava {
    options.compilerArgs += ['--enable-preview']
    options.encoding = 'UTF-8'
}

// Test configuration
test {
    useJUnitPlatform()
    jvmArgs '--enable-preview'

    // Test output
    testLogging {
        events "passed", "skipped", "failed"
        exceptionFormat = 'full'
    }

    // Parallel execution
    maxParallelForks = Runtime.runtime.availableProcessors().intdiv(2) ?: 1
}

// JaCoCo configuration
jacoco {
    toolVersion = "0.8.11"
}

jacocoTestReport {
    dependsOn test
    reports {
        xml.required = true
        html.required = true
    }

    afterEvaluate {
        classDirectories.setFrom(files(classDirectories.files.collect {
            fileTree(dir: it, exclude: [
                '**/Application.class',
                '**/config/**',
                '**/dto/**'
            ])
        }))
    }
}

// SpotBugs configuration
spotbugs {
    ignoreFailures = false
    effort = 'max'
    reportLevel = 'low'
}

spotbugsMain {
    reports {
        html.required = true
        xml.required = false
    }
}

// Checkstyle configuration
checkstyle {
    toolVersion = '10.12.5'
    configFile = file('config/checkstyle/checkstyle.xml')
}

// Custom tasks
task integrationTest(type: Test) {
    description = 'Runs integration tests'
    group = 'verification'

    testClassesDirs = sourceSets.test.output.classesDirs
    classpath = sourceSets.test.runtimeClasspath

    shouldRunAfter test

    include '**/*IT.class'
}

task fatJar(type: Jar) {
    description = 'Create a fat JAR with all dependencies'
    group = 'build'

    archiveClassifier = 'all'
    from {
        configurations.runtimeClasspath.collect { it.isDirectory() ? it : zipTree(it) }
    }
    with jar

    manifest {
        attributes 'Main-Class': application.mainClass.get()
    }
}

// Build configuration for different environments
if (project.hasProperty('env') && project.property('env') == 'production') {
    jar {
        enabled = false
        dependsOn(fatJar)
    }
}

// Quality gate - ensure all checks pass
check.dependsOn jacocoTestReport, spotbugsMain, checkstyle, integrationTest

// Wrapper configuration
wrapper {
    gradleVersion = '8.5'
    distributionType = Wrapper.DistributionType.ALL
}
```

### Gradle Commands
```bash
# Basic Gradle commands
./gradlew clean                  # Clean build artifacts
./gradlew compileJava           # Compile main source
./gradlew compileTestJava       # Compile test source
./gradlew test                  # Run unit tests
./gradlew build                 # Full build
./gradlew run                   # Run application

# Custom tasks
./gradlew integrationTest       # Run integration tests
./gradlew fatJar               # Create fat JAR

# Code quality
./gradlew check                 # Run all checks
./gradlew jacocoTestReport     # Generate coverage report
./gradlew spotbugsMain         # Run SpotBugs
./gradlew checkstyleMain       # Run Checkstyle

# Dependencies
./gradlew dependencies         # Show dependency tree
./gradlew dependencyInsight --dependency jackson-databind

# Parallel execution
./gradlew build --parallel     # Enable parallel execution

# Build cache
./gradlew build --build-cache  # Use build cache

# Profiles/environments
./gradlew build -Penv=production  # Production build

# Debugging
./gradlew build --debug        # Debug mode
./gradlew build --info         # Info logging
./gradlew build --scan         # Generate build scan
```

## IDE Integration

### IntelliJ IDEA Configuration
```xml
<!-- .idea/compiler.xml -->
<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="CompilerConfiguration">
    <annotationProcessing>
      <profile name="Maven default" enabled="true">
        <sourceOutputDir name="target/generated-sources/annotations" />
        <sourceTestOutputDir name="target/generated-test-sources/test-annotations" />
        <outputRelativeToContentRoot value="true" />
      </profile>
    </annotationProcessing>
    <bytecodeTargetLevel>
      <module name="my-java-app" target="21" />
    </bytecodeTargetLevel>
  </component>
</project>
```

### Eclipse Configuration
```xml
<!-- .project -->
<?xml version="1.0" encoding="UTF-8"?>
<projectDescription>
    <name>my-java-app</name>
    <comment></comment>
    <projects></projects>
    <buildSpec>
        <buildCommand>
            <name>org.eclipse.jdt.core.javabuilder</name>
        </buildCommand>
        <buildCommand>
            <name>org.eclipse.m2e.core.maven2Builder</name>
        </buildCommand>
    </buildSpec>
    <natures>
        <nature>org.eclipse.jdt.core.javanature</nature>
        <nature>org.eclipse.m2e.core.maven2Nature</nature>
    </natures>
</projectDescription>
```

## Build Automation Scripts

### Comprehensive Build Script
```bash
#!/bin/bash
# build.sh - Comprehensive build script

set -e  # Exit on any error

# Configuration
PROJECT_NAME="my-java-app"
BUILD_TOOL="maven"  # or "gradle"
JAVA_VERSION="21"
COVERAGE_THRESHOLD=80

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Java version
    if ! command -v java &> /dev/null; then
        log_error "Java is not installed"
        exit 1
    fi

    JAVA_VER=$(java -version 2>&1 | grep -oP 'version "([0-9]+)' | grep -oP '([0-9]+)$')
    if [ "$JAVA_VER" -lt "$JAVA_VERSION" ]; then
        log_error "Java $JAVA_VERSION or higher required, found $JAVA_VER"
        exit 1
    fi

    # Check build tool
    if [ "$BUILD_TOOL" = "maven" ] && ! command -v mvn &> /dev/null; then
        log_error "Maven is not installed"
        exit 1
    fi

    if [ "$BUILD_TOOL" = "gradle" ] && ! [ -f "./gradlew" ]; then
        log_error "Gradle wrapper not found"
        exit 1
    fi

    log_info "Prerequisites satisfied"
}

# Clean build artifacts
clean_build() {
    log_info "Cleaning build artifacts..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn clean
    else
        ./gradlew clean
    fi
}

# Compile source code
compile_sources() {
    log_info "Compiling source code..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn compile test-compile
    else
        ./gradlew compileJava compileTestJava
    fi
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn test
    else
        ./gradlew test
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn failsafe:integration-test failsafe:verify
    else
        ./gradlew integrationTest
    fi
}

# Generate code coverage report
generate_coverage() {
    log_info "Generating code coverage report..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn jacoco:report
        COVERAGE_FILE="target/site/jacoco/index.html"
    else
        ./gradlew jacocoTestReport
        COVERAGE_FILE="build/reports/jacoco/test/html/index.html"
    fi

    # Check coverage threshold (simplified)
    log_info "Coverage report generated: $COVERAGE_FILE"
}

# Run static analysis
run_static_analysis() {
    log_info "Running static analysis..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn spotbugs:check checkstyle:check
    else
        ./gradlew spotbugsMain checkstyleMain
    fi
}

# Package application
package_application() {
    log_info "Packaging application..."

    if [ "$BUILD_TOOL" = "maven" ]; then
        mvn package -DskipTests
    else
        ./gradlew build -x test
    fi
}

# Create distribution
create_distribution() {
    log_info "Creating distribution..."

    DIST_DIR="dist"
    mkdir -p "$DIST_DIR"

    if [ "$BUILD_TOOL" = "maven" ]; then
        cp target/*.jar "$DIST_DIR/"
    else
        cp build/libs/*.jar "$DIST_DIR/"
    fi

    # Copy configuration files
    cp -r src/main/resources/config "$DIST_DIR/" 2>/dev/null || true

    # Create startup script
    cat > "$DIST_DIR/start.sh" << EOF
#!/bin/bash
java -jar *.jar
EOF
    chmod +x "$DIST_DIR/start.sh"

    log_info "Distribution created in $DIST_DIR/"
}

# Main build function
main() {
    log_info "Starting build for $PROJECT_NAME"

    case "${1:-full}" in
        "clean")
            check_prerequisites
            clean_build
            ;;
        "compile")
            check_prerequisites
            compile_sources
            ;;
        "test")
            check_prerequisites
            compile_sources
            run_unit_tests
            ;;
        "integration")
            check_prerequisites
            compile_sources
            run_unit_tests
            run_integration_tests
            ;;
        "analyze")
            check_prerequisites
            compile_sources
            run_static_analysis
            ;;
        "package")
            check_prerequisites
            clean_build
            compile_sources
            run_unit_tests
            package_application
            ;;
        "full")
            check_prerequisites
            clean_build
            compile_sources
            run_unit_tests
            run_integration_tests
            generate_coverage
            run_static_analysis
            package_application
            create_distribution
            ;;
        *)
            echo "Usage: $0 {clean|compile|test|integration|analyze|package|full}"
            exit 1
            ;;
    esac

    log_info "Build completed successfully!"
}

# Run main function with all arguments
main "$@"
```

### Development Environment Setup
```bash
#!/bin/bash
# setup-dev.sh - Development environment setup

# Install SDKMAN
curl -s "https://get.sdkman.io" | bash
source ~/.sdkman/bin/sdkman-init.sh

# Install Java 21
sdk install java 21.0.1-oracle
sdk use java 21.0.1-oracle

# Install Maven
sdk install maven 3.9.6

# Install Gradle
sdk install gradle 8.5

# Create project directories
mkdir -p src/{main,test}/{java,resources}
mkdir -p src/main/java/com/example
mkdir -p src/test/java/com/example

# Initialize Git repository
git init
echo "target/" > .gitignore
echo "build/" >> .gitignore
echo ".idea/" >> .gitignore
echo "*.iml" >> .gitignore

# Create basic project files
cat > README.md << EOF
# My Java Application

## Build Instructions

### Maven
\`\`\`bash
mvn clean install
\`\`\`

### Gradle
\`\`\`bash
./gradlew build
\`\`\`

## Running
\`\`\`bash
java -jar target/my-java-app-1.0.0-SNAPSHOT.jar
\`\`\`
EOF

echo "Development environment setup complete!"
```

This comprehensive build toolchain covers Maven and Gradle build systems with modern Java features, quality gates, and automation scripts for efficient local development.