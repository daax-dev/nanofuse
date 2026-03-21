# Java GitHub Actions CI/CD - Complete Build Pipeline

## Basic CI/CD Workflow

### Maven-Based CI Pipeline
```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  JAVA_VERSION: '21'
  MAVEN_OPTS: '-Xmx2048m'

jobs:
  test:
    name: Test on ${{ matrix.os }}
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        java: ['17', '21']

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Fetch full history for SonarQube

    - name: Set up JDK ${{ matrix.java }}
      uses: actions/setup-java@v4
      with:
        java-version: ${{ matrix.java }}
        distribution: 'temurin'
        cache: maven

    - name: Cache Maven dependencies
      uses: actions/cache@v4
      with:
        path: ~/.m2/repository
        key: ${{ runner.os }}-maven-${{ hashFiles('**/pom.xml') }}
        restore-keys: |
          ${{ runner.os }}-maven-

    - name: Run tests
      run: mvn clean test

    - name: Generate test report
      uses: dorny/test-reporter@v1
      if: success() || failure()
      with:
        name: Maven Tests (${{ matrix.os }}, Java ${{ matrix.java }})
        path: target/surefire-reports/*.xml
        reporter: java-junit

    - name: Upload test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: test-results-${{ matrix.os }}-java${{ matrix.java }}
        path: |
          target/surefire-reports/
          target/failsafe-reports/

  code-quality:
    name: Code Quality Analysis
    runs-on: ubuntu-latest
    needs: test

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: ${{ env.JAVA_VERSION }}
        distribution: 'temurin'
        cache: maven

    - name: Run tests with coverage
      run: mvn clean verify jacoco:report

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: target/site/jacoco/jacoco.xml
        flags: unittests
        name: codecov-umbrella

    - name: SonarQube analysis
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
      run: |
        mvn sonar:sonar \
          -Dsonar.projectKey=${{ github.repository_owner }}_${{ github.event.repository.name }} \
          -Dsonar.organization=${{ github.repository_owner }} \
          -Dsonar.host.url=https://sonarcloud.io

    - name: SpotBugs analysis
      run: mvn spotbugs:check

    - name: Checkstyle analysis
      run: mvn checkstyle:check

    - name: Upload quality reports
      uses: actions/upload-artifact@v4
      with:
        name: quality-reports
        path: |
          target/site/jacoco/
          target/spotbugsXml.xml
          target/checkstyle-result.xml

  security-scan:
    name: Security Scanning
    runs-on: ubuntu-latest
    needs: test

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: ${{ env.JAVA_VERSION }}
        distribution: 'temurin'
        cache: maven

    - name: OWASP Dependency Check
      run: |
        mvn org.owasp:dependency-check-maven:check \
          -DfailBuildOnCVSS=8 \
          -DsuppressionsLocation=owasp-suppressions.xml

    - name: Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy results to GitHub Security
      uses: github/codeql-action/upload-sarif@v3
      if: always()
      with:
        sarif_file: 'trivy-results.sarif'

    - name: Upload security reports
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: security-reports
        path: |
          target/dependency-check-report.html
          trivy-results.sarif

  build-and-publish:
    name: Build and Publish
    runs-on: ubuntu-latest
    needs: [test, code-quality, security-scan]
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    outputs:
      version: ${{ steps.version.outputs.version }}
      image-digest: ${{ steps.build.outputs.digest }}

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: ${{ env.JAVA_VERSION }}
        distribution: 'temurin'
        cache: maven

    - name: Get version
      id: version
      run: echo "version=$(mvn help:evaluate -Dexpression=project.version -q -DforceStdout)" >> $GITHUB_OUTPUT

    - name: Build application
      run: mvn clean package -DskipTests

    - name: Build Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        push: false
        tags: |
          myapp:${{ steps.version.outputs.version }}
          myapp:latest
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Upload artifacts
      uses: actions/upload-artifact@v4
      with:
        name: application-jar
        path: target/*.jar
        retention-days: 30
```

### Gradle-Based CI Pipeline
```yaml
# .github/workflows/gradle-ci.yml
name: Gradle CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  JAVA_VERSION: '21'

jobs:
  test:
    name: Test and Build
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: ${{ env.JAVA_VERSION }}
        distribution: 'temurin'

    - name: Setup Gradle
      uses: gradle/gradle-build-action@v2
      with:
        gradle-home-cache-cleanup: true

    - name: Make gradlew executable
      run: chmod +x ./gradlew

    - name: Run tests
      run: ./gradlew test jacocoTestReport

    - name: Run integration tests
      run: ./gradlew integrationTest

    - name: Build application
      run: ./gradlew build

    - name: Upload test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: test-results
        path: |
          build/test-results/
          build/reports/

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: build/reports/jacoco/test/jacocoTestReport.xml
```

## Advanced CI/CD Features

### Matrix Strategy with Multiple Versions
```yaml
# .github/workflows/matrix-ci.yml
name: Matrix CI

on: [push, pull_request]

jobs:
  test:
    name: Test Java ${{ matrix.java }} on ${{ matrix.os }}
    runs-on: ${{ matrix.os }}

    strategy:
      fail-fast: false
      matrix:
        java: ['11', '17', '21']
        os: [ubuntu-latest, windows-latest, macos-latest]
        exclude:
          # Exclude Java 11 on macOS (example exclusion)
          - java: '11'
            os: macos-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up JDK ${{ matrix.java }}
      uses: actions/setup-java@v4
      with:
        java-version: ${{ matrix.java }}
        distribution: 'temurin'
        cache: maven

    - name: Run tests
      run: mvn test

    - name: Conditional steps for Java 21
      if: matrix.java == '21'
      run: |
        echo "Running Java 21 specific tests"
        mvn test -Dtest=Java21SpecificTest
```

### Database Integration Testing
```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '21'
        distribution: 'temurin'
        cache: maven

    - name: Wait for services
      run: |
        timeout 30s bash -c 'until nc -z localhost 5432; do sleep 1; done'
        timeout 30s bash -c 'until nc -z localhost 6379; do sleep 1; done'

    - name: Run integration tests
      env:
        DATABASE_URL: postgresql://testuser:testpass@localhost:5432/testdb
        REDIS_URL: redis://localhost:6379
      run: mvn failsafe:integration-test failsafe:verify

    - name: Upload integration test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: integration-test-results
        path: target/failsafe-reports/
```

### Performance Testing Pipeline
```yaml
# .github/workflows/performance.yml
name: Performance Tests

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  performance-test:
    name: Performance Benchmarks
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '21'
        distribution: 'temurin'
        cache: maven

    - name: Build application
      run: mvn clean package -DskipTests

    - name: Run JMH benchmarks
      run: |
        java -jar target/benchmarks.jar \
          -rf json \
          -rff benchmark-results.json

    - name: Store benchmark results
      uses: benchmark-action/github-action-benchmark@v1
      with:
        name: Java Benchmark
        tool: 'jmh'
        output-file-path: benchmark-results.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        auto-push: true

    - name: Upload performance reports
      uses: actions/upload-artifact@v4
      with:
        name: performance-reports
        path: benchmark-results.json
```

## Deployment Workflows

### Docker Build and Publish
```yaml
# .github/workflows/docker.yml
name: Docker Build and Publish

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    name: Build and Push Docker Image
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '21'
        distribution: 'temurin'
        cache: maven

    - name: Build JAR
      run: mvn clean package -DskipTests

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=sha

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        platforms: linux/amd64,linux/arm64
```

### Multi-Stage Dockerfile
```dockerfile
# Dockerfile
# Multi-stage build for optimal image size
FROM eclipse-temurin:21-jdk-alpine AS builder

WORKDIR /app

# Copy Maven wrapper and pom.xml
COPY .mvn/ .mvn/
COPY mvnw pom.xml ./

# Download dependencies (for better caching)
RUN ./mvnw dependency:go-offline

# Copy source code and build
COPY src ./src
RUN ./mvnw clean package -DskipTests

# Production image
FROM eclipse-temurin:21-jre-alpine

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install necessary packages
RUN apk --no-cache add curl

WORKDIR /app

# Copy JAR from builder stage
COPY --from=builder /app/target/*.jar app.jar

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/actuator/health || exit 1

# Expose port
EXPOSE 8080

# JVM tuning for containers
ENV JAVA_OPTS="-XX:+UseContainerSupport \
               -XX:MaxRAMPercentage=75.0 \
               -XX:+UseG1GC \
               -XX:+UseStringDeduplication"

# Run application
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### Kubernetes Deployment
```yaml
# .github/workflows/k8s-deploy.yml
name: Deploy to Kubernetes

on:
  workflow_run:
    workflows: ["Docker Build and Publish"]
    types:
      - completed
    branches: [main]

jobs:
  deploy:
    name: Deploy to K8s
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup kubectl
      uses: azure/setup-kubectl@v3
      with:
        version: 'latest'

    - name: Set up Kustomize
      run: |
        curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
        sudo mv kustomize /usr/local/bin/

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2

    - name: Update kubeconfig
      run: aws eks update-kubeconfig --name my-cluster --region us-west-2

    - name: Deploy to staging
      run: |
        cd k8s/overlays/staging
        kustomize edit set image app=ghcr.io/${{ github.repository }}:${{ github.sha }}
        kustomize build . | kubectl apply -f -

    - name: Wait for rollout
      run: kubectl rollout status deployment/my-app -n staging --timeout=300s

    - name: Run smoke tests
      run: |
        kubectl wait --for=condition=ready pod -l app=my-app -n staging --timeout=300s
        curl -f http://staging.example.com/actuator/health

    - name: Deploy to production (manual approval)
      if: github.ref == 'refs/heads/main'
      uses: trstringer/manual-approval@v1
      with:
        secret: ${{ secrets.GITHUB_TOKEN }}
        approvers: admin-user1,admin-user2
        minimum-approvals: 2
        issue-title: "Deploy to production"
        issue-body: "Please approve production deployment for commit ${{ github.sha }}"

    - name: Deploy to production
      if: github.ref == 'refs/heads/main'
      run: |
        cd k8s/overlays/production
        kustomize edit set image app=ghcr.io/${{ github.repository }}:${{ github.sha }}
        kustomize build . | kubectl apply -f -
        kubectl rollout status deployment/my-app -n production --timeout=600s
```

### Release Workflow
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '21'
        distribution: 'temurin'
        cache: maven

    - name: Build release artifacts
      run: |
        mvn clean package -DskipTests
        mvn javadoc:javadoc

    - name: Generate changelog
      id: changelog
      uses: requarks/changelog-action@v1
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
        tag: ${{ github.ref_name }}

    - name: Create GitHub Release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref_name }}
        release_name: Release ${{ github.ref_name }}
        body: ${{ steps.changelog.outputs.changes }}
        draft: false
        prerelease: false

    - name: Upload JAR to release
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ steps.create_release.outputs.upload_url }}
        asset_path: target/my-app-${{ github.ref_name }}.jar
        asset_name: my-app-${{ github.ref_name }}.jar
        asset_content_type: application/java-archive

    - name: Publish to Maven Central
      env:
        MAVEN_GPG_PRIVATE_KEY: ${{ secrets.MAVEN_GPG_PRIVATE_KEY }}
        MAVEN_GPG_PASSPHRASE: ${{ secrets.MAVEN_GPG_PASSPHRASE }}
        MAVEN_CENTRAL_TOKEN: ${{ secrets.MAVEN_CENTRAL_TOKEN }}
      run: |
        echo "$MAVEN_GPG_PRIVATE_KEY" | gpg --batch --import
        mvn deploy -DskipTests -Prelease
```

### Dependency Updates
```yaml
# .github/workflows/dependency-updates.yml
name: Dependency Updates

on:
  schedule:
    - cron: '0 10 * * MON'  # Weekly on Monday
  workflow_dispatch:

jobs:
  update-dependencies:
    name: Update Dependencies
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        token: ${{ secrets.DEPENDENCY_UPDATE_TOKEN }}

    - name: Set up JDK
      uses: actions/setup-java@v4
      with:
        java-version: '21'
        distribution: 'temurin'

    - name: Update Maven dependencies
      run: |
        mvn versions:use-latest-releases
        mvn versions:update-properties

    - name: Run tests
      run: mvn clean test

    - name: Create Pull Request
      uses: peter-evans/create-pull-request@v5
      with:
        token: ${{ secrets.DEPENDENCY_UPDATE_TOKEN }}
        commit-message: 'chore: update dependencies'
        title: 'chore: update dependencies'
        body: |
          This PR updates project dependencies to their latest versions.

          Please review the changes and ensure all tests pass.
        branch: dependency-updates
        delete-branch: true
```

This comprehensive GitHub Actions setup provides enterprise-grade CI/CD capabilities for Java applications with testing, quality gates, security scanning, and automated deployment.