# GitHub Actions for C# - Comprehensive CI/CD Guide

## Table of Contents

1. [GitHub Actions Fundamentals for .NET](#github-actions-fundamentals-for-net)
2. [Basic CI Workflow](#basic-ci-workflow)
3. [Multi-Platform Builds](#multi-platform-builds)
4. [Multi-Framework Targeting](#multi-framework-targeting)
5. [Advanced Testing and Code Coverage](#advanced-testing-and-code-coverage)
6. [Security Scanning and Dependency Checking](#security-scanning-and-dependency-checking)
7. [Artifact Publishing and Deployment](#artifact-publishing-and-deployment)
8. [Release Automation and Semantic Versioning](#release-automation-and-semantic-versioning)
9. [Docker Image Building and Publishing](#docker-image-building-and-publishing)
10. [Performance Testing and Monitoring](#performance-testing-and-monitoring)
11. [Branch Protection and Quality Gates](#branch-protection-and-quality-gates)
12. [Enterprise Patterns and Advanced Scenarios](#enterprise-patterns-and-advanced-scenarios)

---

## GitHub Actions Fundamentals for .NET

### Core Concepts

GitHub Actions for .NET leverages the cross-platform nature of .NET Core/.NET 5+ to provide robust CI/CD capabilities across Windows, Linux, and macOS environments.

### Key Components

- **Runners**: GitHub-hosted or self-hosted environments
- **Actions**: Reusable units of code
- **Workflows**: YAML files defining CI/CD processes
- **Jobs**: Groups of steps that execute on the same runner
- **Steps**: Individual tasks within a job

### Essential .NET Actions

```yaml
# Common .NET actions
- uses: actions/setup-dotnet@v4
- uses: actions/checkout@v4
- uses: actions/upload-artifact@v4
- uses: actions/cache@v4
```

### Runner Matrix for .NET

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    dotnet-version: ['6.0.x', '8.0.x']
    include:
      - os: ubuntu-latest
        dotnet-version: '9.0.x'
runs-on: ${{ matrix.os }}
```

---

## Basic CI Workflow

### Minimal CI Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --no-restore --configuration Release

    - name: Test
      run: dotnet test --no-build --configuration Release --verbosity normal
```

### Enhanced CI with Caching

```yaml
# .github/workflows/ci-enhanced.yml
name: Enhanced CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:

env:
  DOTNET_VERSION: '8.0.x'
  DOTNET_SKIP_FIRST_TIME_EXPERIENCE: 1
  DOTNET_NOLOGO: true
  DOTNET_CLI_TELEMETRY_OPTOUT: 1

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0 # Full clone for better analysis

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ env.DOTNET_VERSION }}

    - name: Cache NuGet packages
      uses: actions/cache@v4
      with:
        path: ~/.nuget/packages
        key: ${{ runner.os }}-nuget-${{ hashFiles('**/*.csproj') }}
        restore-keys: |
          ${{ runner.os }}-nuget-

    - name: Restore dependencies
      run: dotnet restore --verbosity minimal

    - name: Build solution
      run: |
        dotnet build \
          --no-restore \
          --configuration Release \
          --verbosity minimal \
          -p:TreatWarningsAsErrors=true

    - name: Run tests
      run: |
        dotnet test \
          --no-build \
          --configuration Release \
          --verbosity minimal \
          --logger trx \
          --results-directory "TestResults-${{ matrix.dotnet-version }}"

    - name: Upload test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: test-results
        path: TestResults-${{ env.DOTNET_VERSION }}/*.trx
```

### Code Quality Integration

```yaml
# .github/workflows/quality.yml
name: Code Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  lint-and-format:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install format tool
      run: dotnet tool install -g dotnet-format

    - name: Check formatting
      run: dotnet format --verify-no-changes --verbosity diagnostic

    - name: Run code analysis
      run: |
        dotnet build \
          --configuration Release \
          -p:RunAnalyzersDuringBuild=true \
          -p:TreatWarningsAsErrors=true \
          -p:WarningsAsErrors="" \
          -p:WarningsNotAsErrors="CS1591" # XML documentation warnings
```

---

## Multi-Platform Builds

### Cross-Platform Matrix Build

```yaml
# .github/workflows/cross-platform.yml
name: Cross Platform Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        configuration: [Debug, Release]
        include:
          # Add specific configurations for different OS
          - os: ubuntu-latest
            artifact-name: linux-build
            test-command: dotnet test --collect:"XPlat Code Coverage"
          - os: windows-latest
            artifact-name: windows-build
            test-command: dotnet test --collect:"Code Coverage"
          - os: macos-latest
            artifact-name: macos-build
            test-command: dotnet test --collect:"XPlat Code Coverage"

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Cache dependencies
      uses: actions/cache@v4
      with:
        path: ~/.nuget/packages
        key: ${{ runner.os }}-nuget-${{ hashFiles('**/*.csproj') }}

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration ${{ matrix.configuration }} --no-restore

    - name: Test
      run: ${{ matrix.test-command }} --configuration ${{ matrix.configuration }} --no-build

    - name: Publish artifacts (Release only)
      if: matrix.configuration == 'Release'
      run: |
        dotnet publish \
          --configuration Release \
          --no-build \
          --output ./publish/${{ matrix.os }}

    - name: Upload build artifacts
      if: matrix.configuration == 'Release'
      uses: actions/upload-artifact@v4
      with:
        name: ${{ matrix.artifact-name }}
        path: ./publish/${{ matrix.os }}
```

### Platform-Specific Optimizations

```yaml
# .github/workflows/platform-optimized.yml
name: Platform Optimized Build

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  windows-build:
    runs-on: windows-latest
    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Setup MSBuild
      uses: microsoft/setup-msbuild@v2

    - name: Build with MSBuild
      run: |
        msbuild /p:Configuration=Release /p:Platform="Any CPU" /p:OutputPath=bin/Release/ MyProject.sln

    - name: Run Windows-specific tests
      run: dotnet test --filter "Category=Windows" --configuration Release

  linux-build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install native dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y libgdiplus libc6-dev

    - name: Build
      run: dotnet build --configuration Release

    - name: Run Linux-specific tests
      run: dotnet test --filter "Category=Linux" --configuration Release

  macos-build:
    runs-on: macos-latest
    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Build
      run: dotnet build --configuration Release

    - name: Run macOS-specific tests
      run: dotnet test --filter "Category=macOS" --configuration Release
```

---

## Multi-Framework Targeting

### Multi-Target Framework Build

```yaml
# .github/workflows/multi-framework.yml
name: Multi-Framework Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    strategy:
      matrix:
        framework: ['net6.0', 'net8.0', 'net48']
        os: [ubuntu-latest, windows-latest]
        exclude:
          # .NET Framework only runs on Windows
          - framework: 'net48'
            os: ubuntu-latest

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: |
          6.0.x
          8.0.x

    - name: Build for specific framework
      run: |
        dotnet build \
          --framework ${{ matrix.framework }} \
          --configuration Release

    - name: Test for specific framework
      run: |
        dotnet test \
          --framework ${{ matrix.framework }} \
          --configuration Release \
          --no-build \
          --logger "trx;LogFileName=TestResults-${{ matrix.framework }}-${{ matrix.os }}.trx"

    - name: Upload test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: test-results-${{ matrix.framework }}-${{ matrix.os }}
        path: "**/*.trx"
```

### Complex Multi-Framework Scenario

```yaml
# .github/workflows/complex-multi-framework.yml
name: Complex Multi-Framework

on:
  workflow_dispatch:
    inputs:
      frameworks:
        description: 'Target frameworks (comma-separated)'
        required: false
        default: 'net6.0,net8.0,net48'

env:
  FRAMEWORKS: ${{ github.event.inputs.frameworks || 'net6.0,net8.0,net48' }}

jobs:
  prepare:
    runs-on: ubuntu-latest
    outputs:
      matrix: ${{ steps.set-matrix.outputs.matrix }}
    steps:
    - name: Set matrix
      id: set-matrix
      run: |
        frameworks=$(echo "${{ env.FRAMEWORKS }}" | jq -R -c 'split(",")')
        echo "matrix={\"framework\":$frameworks}" >> $GITHUB_OUTPUT

  build-and-test:
    needs: prepare
    strategy:
      matrix: ${{ fromJson(needs.prepare.outputs.matrix) }}

    runs-on: ${{ startsWith(matrix.framework, 'net4') && 'windows-latest' || 'ubuntu-latest' }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET (Multi-version)
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: |
          6.0.x
          8.0.x

    - name: Restore for framework
      run: dotnet restore --framework ${{ matrix.framework }}

    - name: Build for framework
      run: |
        dotnet build \
          --framework ${{ matrix.framework }} \
          --configuration Release \
          --no-restore \
          -p:TargetFramework=${{ matrix.framework }}

    - name: Test for framework
      run: |
        dotnet test \
          --framework ${{ matrix.framework }} \
          --configuration Release \
          --no-build \
          --collect:"XPlat Code Coverage" \
          --results-directory "coverage-${{ matrix.framework }}"

    - name: Generate coverage report
      uses: danielpalme/ReportGenerator-GitHub-Action@5.2.0
      with:
        reports: 'coverage-${{ matrix.framework }}/**/coverage.cobertura.xml'
        targetdir: 'coveragereport-${{ matrix.framework }}'
        reporttypes: 'Html;Cobertura'

    - name: Upload coverage
      uses: actions/upload-artifact@v4
      with:
        name: coverage-${{ matrix.framework }}
        path: coveragereport-${{ matrix.framework }}
```

---

## Advanced Testing and Code Coverage

### Comprehensive Testing Workflow

```yaml
# .github/workflows/testing.yml
name: Comprehensive Testing

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Run unit tests with coverage
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --verbosity minimal \
          --collect:"XPlat Code Coverage" \
          --results-directory ./coverage \
          --logger trx \
          --filter "Category=Unit"

    - name: Generate coverage report
      uses: danielpalme/ReportGenerator-GitHub-Action@5.2.0
      with:
        reports: './coverage/**/coverage.cobertura.xml'
        targetdir: 'coveragereport'
        reporttypes: 'Html;Cobertura;JsonSummary'
        verbosity: 'Verbose'

    - name: Upload coverage reports to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./coveragereport/Cobertura.xml
        flags: unittests
        name: codecov-umbrella
        fail_ci_if_error: true

    - name: Comment coverage on PR
      if: github.event_name == 'pull_request'
      uses: marocchino/sticky-pull-request-comment@v2
      with:
        recreate: true
        path: coveragereport/Summary.md

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
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
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Run integration tests
      env:
        ConnectionStrings__DefaultConnection: "Host=localhost;Database=testdb;Username=postgres;Password=postgres"
        ConnectionStrings__Redis: "localhost:6379"
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --verbosity minimal \
          --logger trx \
          --filter "Category=Integration" \
          --results-directory ./integration-results

    - name: Upload integration test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: integration-test-results
        path: ./integration-results/*.trx

  performance-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install NBomber CLI
      run: dotnet tool install -g NBomber.CLI

    - name: Run performance tests
      run: |
        dotnet run --project ./tests/PerformanceTests \
          --configuration Release \
          -- --output ./perf-results

    - name: Upload performance results
      uses: actions/upload-artifact@v4
      with:
        name: performance-results
        path: ./perf-results
```

### Advanced Code Coverage Configuration

```yaml
# .github/workflows/coverage-advanced.yml
name: Advanced Code Coverage

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  coverage:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0 # Needed for SonarQube

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Setup Java (for SonarQube)
      uses: actions/setup-java@v4
      with:
        distribution: 'temurin'
        java-version: '17'

    - name: Cache SonarQube packages
      uses: actions/cache@v4
      with:
        path: ~/.sonar/cache
        key: ${{ runner.os }}-sonar
        restore-keys: ${{ runner.os }}-sonar

    - name: Install SonarQube scanner
      run: |
        dotnet tool install --global dotnet-sonarscanner
        dotnet tool install --global dotnet-coverage

    - name: Begin SonarQube analysis
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
      run: |
        dotnet sonarscanner begin \
          /k:"your-project-key" \
          /o:"your-org" \
          /d:sonar.login="${{ secrets.SONAR_TOKEN }}" \
          /d:sonar.host.url="https://sonarcloud.io" \
          /d:sonar.cs.vscoveragexml.reportsPaths="**/coverage.xml" \
          /d:sonar.exclusions="**/bin/**,**/obj/**,**/*.Generated.cs" \
          /d:sonar.coverage.exclusions="**/*Tests*.cs,**/Program.cs"

    - name: Restore and build
      run: |
        dotnet restore
        dotnet build --configuration Release --no-restore

    - name: Run tests with coverage
      run: |
        dotnet coverage collect \
          "dotnet test --configuration Release --no-build --verbosity minimal" \
          -f xml \
          -o coverage.xml

    - name: End SonarQube analysis
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
      run: dotnet sonarscanner end /d:sonar.login="${{ secrets.SONAR_TOKEN }}"

    - name: Quality Gate check
      uses: sonarqube-quality-gate-action@master
      timeout-minutes: 5
      env:
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### Mutation Testing

```yaml
# .github/workflows/mutation-testing.yml
name: Mutation Testing

on:
  schedule:
    - cron: '0 2 * * 1' # Weekly on Monday at 2 AM
  workflow_dispatch:

jobs:
  mutation-testing:
    runs-on: ubuntu-latest
    timeout-minutes: 60

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install Stryker.NET
      run: dotnet tool install -g dotnet-stryker

    - name: Restore dependencies
      run: dotnet restore

    - name: Run mutation testing
      run: |
        dotnet stryker \
          --output ./mutation-report \
          --reporter html \
          --reporter json \
          --threshold-high 80 \
          --threshold-low 60 \
          --threshold-break 40

    - name: Upload mutation report
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: mutation-testing-report
        path: ./mutation-report

    - name: Comment mutation score
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v7
      with:
        script: |
          const fs = require('fs');
          const report = JSON.parse(fs.readFileSync('./mutation-report/mutation-report.json'));
          const score = report.summary.mutationScore;

          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: `## Mutation Testing Results\n\nMutation Score: **${score}%**\n\n${score >= 80 ? '✅' : score >= 60 ? '⚠️' : '❌'} ${score >= 80 ? 'Excellent' : score >= 60 ? 'Good' : 'Needs Improvement'}`
          });
```

---

## Security Scanning and Dependency Checking

### Comprehensive Security Workflow

```yaml
# .github/workflows/security.yml
name: Security Scanning

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 6 * * 1' # Weekly on Monday at 6 AM

jobs:
  dependency-check:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Check for vulnerable packages
      run: |
        dotnet list package --vulnerable --include-transitive 2>&1 | tee vulnerable-packages.txt
        if grep -q "has the following vulnerable packages" vulnerable-packages.txt; then
          echo "❌ Vulnerable packages found!"
          exit 1
        else
          echo "✅ No vulnerable packages found"
        fi

    - name: Check for deprecated packages
      run: |
        dotnet list package --deprecated --include-transitive

    - name: Upload vulnerability report
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: vulnerability-report
        path: vulnerable-packages.txt

  codeql-analysis:
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Initialize CodeQL
      uses: github/codeql-action/init@v3
      with:
        languages: csharp
        queries: +security-and-quality

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Build for CodeQL
      run: |
        dotnet restore
        dotnet build --configuration Release --no-restore

    - name: Perform CodeQL Analysis
      uses: github/codeql-action/analyze@v3
      with:
        category: "/language:csharp"

  semgrep-scan:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Run Semgrep
      uses: semgrep/semgrep-action@v1
      with:
        config: >-
          p/security-audit
          p/secrets
          p/csharp
          p/owasp-top-ten
        generateSarif: "1"

    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: semgrep.sarif
      if: always()

  snyk-security:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Run Snyk to check for vulnerabilities
      uses: snyk/actions/dotnet@master
      env:
        SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
      with:
        args: --severity-threshold=high --all-projects

    - name: Upload Snyk report
      uses: github/codeql-action/upload-sarif@v3
      if: always()
      with:
        sarif_file: snyk.sarif

  docker-security:
    runs-on: ubuntu-latest
    if: hashFiles('Dockerfile') != ''

    steps:
    - uses: actions/checkout@v4

    - name: Build Docker image
      run: docker build -t test-image .

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: 'test-image'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: 'trivy-results.sarif'
```

### Advanced Dependency Management

```yaml
# .github/workflows/dependency-management.yml
name: Dependency Management

on:
  schedule:
    - cron: '0 8 * * 1' # Weekly on Monday at 8 AM
  workflow_dispatch:

jobs:
  audit-licenses:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install license checker
      run: dotnet tool install -g dotnet-project-licenses

    - name: Generate license report
      run: |
        dotnet-project-licenses \
          --input . \
          --output-directory ./licenses \
          --export-license-texts \
          --print summary

    - name: Check for prohibited licenses
      run: |
        # Define prohibited licenses
        prohibited=("GPL" "AGPL" "LGPL")

        for license in "${prohibited[@]}"; do
          if grep -i "$license" ./licenses/*.json; then
            echo "❌ Prohibited license found: $license"
            exit 1
          fi
        done

        echo "✅ No prohibited licenses found"

    - name: Upload license report
      uses: actions/upload-artifact@v4
      with:
        name: license-report
        path: ./licenses

  update-dependencies:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install update tool
      run: dotnet tool install -g dotnet-outdated-tool

    - name: Check for outdated packages
      run: |
        dotnet outdated --output ./outdated-packages.json

    - name: Update patch versions
      run: |
        dotnet outdated --upgrade --version-lock Major

    - name: Test after updates
      run: |
        dotnet restore
        dotnet build --configuration Release
        dotnet test --configuration Release

    - name: Create Pull Request
      uses: peter-evans/create-pull-request@v6
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
        commit-message: 'chore: update NuGet packages (patch versions)'
        title: 'Automated NuGet Package Updates'
        body: |
          This PR updates NuGet packages to their latest patch versions.

          - Only patch version updates are included for safety
          - All tests pass with the updated dependencies

          Please review the changes before merging.
        branch: automated/nuget-updates
        delete-branch: true
```

---

## Artifact Publishing and Deployment

### NuGet Package Publishing

```yaml
# .github/workflows/nuget-publish.yml
name: Publish NuGet Package

on:
  push:
    tags: [ 'v*' ]
  workflow_dispatch:
    inputs:
      prerelease:
        description: 'Is this a prerelease?'
        required: false
        default: false
        type: boolean

env:
  DOTNET_VERSION: '8.0.x'
  PROJECT_PATH: './src/MyLibrary/MyLibrary.csproj'

jobs:
  publish:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0 # Needed for GitVersion

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ env.DOTNET_VERSION }}

    - name: Install GitVersion
      uses: gittools/actions/gitversion/setup@v1.1.1
      with:
        versionSpec: '5.x'

    - name: Determine Version
      uses: gittools/actions/gitversion/execute@v1.1.1
      id: gitversion
      with:
        useConfigFile: true

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: |
        dotnet build \
          --configuration Release \
          --no-restore \
          -p:Version=${{ steps.gitversion.outputs.semVer }} \
          -p:AssemblyVersion=${{ steps.gitversion.outputs.assemblySemVer }} \
          -p:FileVersion=${{ steps.gitversion.outputs.assemblySemFileVer }} \
          -p:InformationalVersion=${{ steps.gitversion.outputs.informationalVersion }}

    - name: Test
      run: dotnet test --configuration Release --no-build

    - name: Pack NuGet package
      run: |
        dotnet pack ${{ env.PROJECT_PATH }} \
          --configuration Release \
          --no-build \
          --output ./packages \
          -p:PackageVersion=${{ steps.gitversion.outputs.semVer }} \
          -p:RepositoryUrl=${{ github.server_url }}/${{ github.repository }} \
          -p:RepositoryCommit=${{ github.sha }}

    - name: Publish to NuGet.org
      run: |
        dotnet nuget push ./packages/*.nupkg \
          --api-key ${{ secrets.NUGET_API_KEY }} \
          --source https://api.nuget.org/v3/index.json \
          --skip-duplicate

    - name: Publish to GitHub Packages
      run: |
        dotnet nuget push ./packages/*.nupkg \
          --api-key ${{ secrets.GITHUB_TOKEN }} \
          --source https://nuget.pkg.github.com/${{ github.repository_owner }}/index.json \
          --skip-duplicate

    - name: Upload package artifacts
      uses: actions/upload-artifact@v4
      with:
        name: nuget-packages
        path: ./packages/*.nupkg
```

### Application Deployment

```yaml
# .github/workflows/deploy.yml
name: Deploy Application

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        default: 'staging'
        type: choice
        options:
        - staging
        - production

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.gitversion.outputs.semVer }}

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Determine Version
      uses: gittools/actions/gitversion/execute@v1.1.1
      id: gitversion

    - name: Restore dependencies
      run: dotnet restore

    - name: Build and publish
      run: |
        dotnet publish ./src/MyApp/MyApp.csproj \
          --configuration Release \
          --output ./publish \
          --self-contained true \
          --runtime linux-x64 \
          -p:Version=${{ steps.gitversion.outputs.semVer }} \
          -p:PublishSingleFile=true \
          -p:PublishTrimmed=true

    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: published-app
        path: ./publish

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    if: github.ref == 'refs/heads/main' || github.event.inputs.environment == 'staging'

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: published-app
        path: ./app

    - name: Deploy to staging
      uses: azure/webapps-deploy@v3
      with:
        app-name: 'my-app-staging'
        publish-profile: ${{ secrets.AZURE_WEBAPP_STAGING_PUBLISH_PROFILE }}
        package: './app'

    - name: Run health check
      run: |
        sleep 30 # Wait for deployment
        curl -f https://my-app-staging.azurewebsites.net/health || exit 1

  deploy-production:
    needs: [build, deploy-staging]
    runs-on: ubuntu-latest
    environment: production
    if: startsWith(github.ref, 'refs/tags/v') || github.event.inputs.environment == 'production'

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: published-app
        path: ./app

    - name: Deploy to production
      uses: azure/webapps-deploy@v3
      with:
        app-name: 'my-app-production'
        publish-profile: ${{ secrets.AZURE_WEBAPP_PRODUCTION_PUBLISH_PROFILE }}
        package: './app'

    - name: Run health check
      run: |
        sleep 30 # Wait for deployment
        curl -f https://my-app-production.azurewebsites.net/health || exit 1

    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        text: |
          🚀 Production deployment completed successfully!
          Version: ${{ needs.build.outputs.version }}
          Commit: ${{ github.sha }}
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## Release Automation and Semantic Versioning

### Automated Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    branches: [ main ]
  workflow_dispatch:

jobs:
  release:
    runs-on: ubuntu-latest
    outputs:
      released: ${{ steps.release.outputs.released }}
      version: ${{ steps.release.outputs.version }}
      tag: ${{ steps.release.outputs.tag }}

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'

    - name: Install semantic-release
      run: |
        npm install -g semantic-release
        npm install -g @semantic-release/changelog
        npm install -g @semantic-release/git
        npm install -g @semantic-release/github

    - name: Create .releaserc.json
      run: |
        cat > .releaserc.json << 'EOF'
        {
          "branches": ["main"],
          "plugins": [
            "@semantic-release/commit-analyzer",
            "@semantic-release/release-notes-generator",
            "@semantic-release/changelog",
            [
              "@semantic-release/git",
              {
                "assets": ["CHANGELOG.md"],
                "message": "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}"
              }
            ],
            "@semantic-release/github"
          ]
        }
        EOF

    - name: Run semantic-release
      id: release
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: |
        npx semantic-release --dry-run > release-output.txt 2>&1 || true

        if grep -q "The next release version is" release-output.txt; then
          echo "released=true" >> $GITHUB_OUTPUT
          version=$(grep "The next release version is" release-output.txt | sed 's/.*The next release version is //')
          echo "version=$version" >> $GITHUB_OUTPUT
          echo "tag=v$version" >> $GITHUB_OUTPUT

          # Actually run the release
          npx semantic-release
        else
          echo "released=false" >> $GITHUB_OUTPUT
          echo "No release needed"
        fi

    - name: Upload release notes
      if: steps.release.outputs.released == 'true'
      uses: actions/upload-artifact@v4
      with:
        name: release-notes
        path: CHANGELOG.md

  build-and-publish:
    needs: release
    if: needs.release.outputs.released == 'true'
    uses: ./.github/workflows/nuget-publish.yml
    secrets: inherit
    with:
      version: ${{ needs.release.outputs.version }}

  create-github-release:
    needs: [release, build-and-publish]
    if: needs.release.outputs.released == 'true'
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        ref: ${{ needs.release.outputs.tag }}

    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: nuget-packages
        path: ./packages

    - name: Create GitHub Release
      uses: softprops/action-gh-release@v2
      with:
        tag_name: ${{ needs.release.outputs.tag }}
        name: Release ${{ needs.release.outputs.version }}
        body_path: CHANGELOG.md
        files: ./packages/*
        draft: false
        prerelease: ${{ contains(needs.release.outputs.version, '-') }}
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### GitVersion Configuration

```yaml
# GitVersion.yml
mode: ContinuousDelivery
branches:
  main:
    regex: ^master$|^main$
    mode: ContinuousDelivery
    tag: ''
    increment: Patch
    prevent-increment-of-merged-branch-version: true
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: false
  release:
    regex: ^releases?[/-]
    mode: ContinuousDelivery
    tag: beta
    increment: Patch
    prevent-increment-of-merged-branch-version: true
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: true
  feature:
    regex: ^features?[/-]
    mode: ContinuousDelivery
    tag: useBranchName
    increment: Inherit
    prevent-increment-of-merged-branch-version: false
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: false
  pull-request:
    regex: ^(pull|pull\-requests|pr)[/-]
    mode: ContinuousDelivery
    tag: PullRequest
    increment: Inherit
    prevent-increment-of-merged-branch-version: false
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: false
  hotfix:
    regex: ^hotfix(es)?[/-]
    mode: ContinuousDelivery
    tag: beta
    increment: Patch
    prevent-increment-of-merged-branch-version: false
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: false
  support:
    regex: ^support[/-]
    mode: ContinuousDelivery
    tag: ''
    increment: Patch
    prevent-increment-of-merged-branch-version: true
    track-merge-target: false
    tracks-release-branches: false
    is-release-branch: false
  develop:
    regex: ^dev(elop)?(ment)?$
    mode: ContinuousDeployment
    tag: unstable
    increment: Minor
    prevent-increment-of-merged-branch-version: false
    track-merge-target: true
    tracks-release-branches: true
    is-release-branch: false
ignore:
  sha: []
merge-message-formats: {}
```

---

## Docker Image Building and Publishing

### Docker Multi-Stage Build Workflow

```yaml
# .github/workflows/docker.yml
name: Docker Build and Push

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

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
          type=semver,pattern={{major}}
          type=raw,value=latest,enable={{is_default_branch}}
        labels: |
          org.opencontainers.image.title=My Application
          org.opencontainers.image.description=A sample .NET application
          org.opencontainers.image.vendor=My Company

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        build-args: |
          BUILDKIT_INLINE_CACHE=1
          VERSION=${{ github.sha }}

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: 'trivy-results.sarif'
```

### Optimized Dockerfile

```dockerfile
# Dockerfile
# Use the official .NET SDK image for building
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy csproj files and restore dependencies
COPY ["src/MyApp/MyApp.csproj", "src/MyApp/"]
COPY ["src/MyApp.Core/MyApp.Core.csproj", "src/MyApp.Core/"]
RUN dotnet restore "src/MyApp/MyApp.csproj"

# Copy source code and build
COPY . .
WORKDIR "/src/src/MyApp"
RUN dotnet build "MyApp.csproj" -c Release -o /app/build --no-restore

# Publish the application
RUN dotnet publish "MyApp.csproj" \
    -c Release \
    -o /app/publish \
    --no-restore \
    --self-contained false \
    /p:PublishTrimmed=true \
    /p:PublishSingleFile=false

# Use the official ASP.NET Core runtime image
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

# Create non-root user
RUN adduser --disabled-password --gecos '' appuser && chown -R appuser /app
USER appuser

# Copy published application
COPY --from=build /app/publish .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose port
EXPOSE 8080

ENTRYPOINT ["dotnet", "MyApp.dll"]
```

### Docker Compose for Testing

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: runtime
    environment:
      - ASPNETCORE_ENVIRONMENT=Testing
      - ConnectionStrings__DefaultConnection=Host=postgres;Database=testdb;Username=testuser;Password=testpass
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser -d testdb"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Advanced Docker Testing

```yaml
# .github/workflows/docker-test.yml
name: Docker Integration Tests

on:
  pull_request:
    branches: [ main ]
    paths:
      - 'Dockerfile'
      - 'docker-compose*.yml'
      - 'src/**'

jobs:
  docker-test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Build test image
      run: |
        docker build -t test-app:latest .

    - name: Test container startup
      run: |
        # Start container and wait for health check
        docker run -d --name test-container -p 8080:8080 test-app:latest

        # Wait for container to be healthy
        timeout 60 bash -c 'until docker inspect --format="{{.State.Health.Status}}" test-container | grep -q healthy; do sleep 2; done'

        # Test basic functionality
        curl -f http://localhost:8080/health

        # Clean up
        docker stop test-container
        docker rm test-container

    - name: Run Docker Compose tests
      run: |
        docker-compose -f docker-compose.test.yml up --build -d

        # Wait for services to be ready
        sleep 30

        # Run integration tests
        docker-compose -f docker-compose.test.yml exec -T app dotnet test --logger trx

        # Clean up
        docker-compose -f docker-compose.test.yml down -v

    - name: Test multi-platform build
      if: github.event_name == 'pull_request'
      run: |
        docker buildx build \
          --platform linux/amd64,linux/arm64 \
          --tag multi-platform-test:latest \
          .
```

---

## Performance Testing and Monitoring

### Comprehensive Performance Testing

```yaml
# .github/workflows/performance.yml
name: Performance Testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 4 * * 1' # Weekly performance baseline

jobs:
  benchmark-tests:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Run BenchmarkDotNet tests
      run: |
        dotnet run --project ./tests/PerformanceTests \
          --configuration Release \
          --framework net8.0 \
          -- --filter "*" --exporters json

    - name: Store benchmark results
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'benchmarkdotnet'
        output-file-path: ./tests/PerformanceTests/BenchmarkDotNet.Artifacts/results/combined-results.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        auto-push: true
        comment-on-alert: true
        alert-threshold: '200%'
        fail-on-alert: true

    - name: Upload benchmark artifacts
      uses: actions/upload-artifact@v4
      with:
        name: benchmark-results
        path: ./tests/PerformanceTests/BenchmarkDotNet.Artifacts/

  load-testing:
    runs-on: ubuntu-latest
    services:
      app:
        image: myapp:latest
        ports:
          - 8080:8080
        options: --health-cmd "curl -f http://localhost:8080/health" --health-interval 10s

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install NBomber CLI
      run: dotnet tool install -g NBomber.CLI

    - name: Wait for service
      run: |
        timeout 60 bash -c 'until curl -f http://localhost:8080/health; do sleep 2; done'

    - name: Run load tests
      run: |
        dotnet run --project ./tests/LoadTests \
          --configuration Release \
          -- --target http://localhost:8080 \
             --duration 120s \
             --rate 50 \
             --output ./load-test-results

    - name: Analyze load test results
      run: |
        # Extract key metrics
        avg_response_time=$(jq '.ScenarioStats[0].Ok.Response.Mean' ./load-test-results/stats.json)
        p95_response_time=$(jq '.ScenarioStats[0].Ok.Response.Percentile95' ./load-test-results/stats.json)
        error_rate=$(jq '.ScenarioStats[0].Fail.Response.Count / .ScenarioStats[0].AllOkCount * 100' ./load-test-results/stats.json)

        echo "Average Response Time: ${avg_response_time}ms"
        echo "95th Percentile: ${p95_response_time}ms"
        echo "Error Rate: ${error_rate}%"

        # Fail if performance degrades
        if (( $(echo "$avg_response_time > 500" | bc -l) )); then
          echo "❌ Average response time too high: ${avg_response_time}ms"
          exit 1
        fi

        if (( $(echo "$p95_response_time > 1000" | bc -l) )); then
          echo "❌ 95th percentile too high: ${p95_response_time}ms"
          exit 1
        fi

        if (( $(echo "$error_rate > 1" | bc -l) )); then
          echo "❌ Error rate too high: ${error_rate}%"
          exit 1
        fi

    - name: Upload load test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: load-test-results
        path: ./load-test-results

  memory-profiling:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Install dotMemory CLI
      run: |
        wget -O dotmemory.tar.gz "https://download.jetbrains.com/resharper/dotUltimate.2023.3.4/JetBrains.dotMemoryUnit.3.2.20231218.164431.tar.gz"
        tar -xzf dotmemory.tar.gz
        chmod +x dotMemoryUnit/dotMemoryUnit

    - name: Build application
      run: dotnet build --configuration Release

    - name: Run memory profiling tests
      run: |
        dotnet test ./tests/MemoryTests \
          --configuration Release \
          --logger trx \
          --results-directory ./memory-results

    - name: Analyze memory usage
      run: |
        # Parse memory profiling results
        # This would integrate with your memory profiling tool
        echo "Memory profiling completed"

    - name: Upload memory profiling results
      uses: actions/upload-artifact@v4
      with:
        name: memory-profiling-results
        path: ./memory-results

  application-insights:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Query Application Insights
      uses: azure/CLI@v1
      with:
        azcliversion: 2.0.72
        inlineScript: |
          # Query performance metrics from Application Insights
          az monitor app-insights query \
            --app "${{ secrets.APPINSIGHTS_APP_ID }}" \
            --analytics-query "
              requests
              | where timestamp > ago(1h)
              | summarize
                  avg_duration = avg(duration),
                  p95_duration = percentile(duration, 95),
                  request_count = count(),
                  error_rate = countif(success == false) * 100.0 / count()
              by bin(timestamp, 5m)
              | order by timestamp desc
              | limit 12
            " \
            --out table
      env:
        AZURE_CLIENT_ID: ${{ secrets.AZURE_CLIENT_ID }}
        AZURE_CLIENT_SECRET: ${{ secrets.AZURE_CLIENT_SECRET }}
        AZURE_TENANT_ID: ${{ secrets.AZURE_TENANT_ID }}
```

### Sample BenchmarkDotNet Test

```csharp
// tests/PerformanceTests/StringBenchmarks.cs
using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Running;

[MemoryDiagnoser]
[SimpleJob(BenchmarkDotNet.Jobs.RuntimeMoniker.Net80)]
public class StringBenchmarks
{
    private const string TestString = "Hello, World!";
    private readonly string[] _strings = Enumerable.Range(0, 1000).Select(i => $"String {i}").ToArray();

    [Benchmark]
    public string StringConcatenation()
    {
        var result = "";
        for (int i = 0; i < 100; i++)
        {
            result += TestString;
        }
        return result;
    }

    [Benchmark]
    public string StringBuilderAppend()
    {
        var sb = new StringBuilder();
        for (int i = 0; i < 100; i++)
        {
            sb.Append(TestString);
        }
        return sb.ToString();
    }

    [Benchmark]
    public string StringJoin()
    {
        return string.Join("", _strings.Take(100));
    }
}

public class Program
{
    public static void Main(string[] args)
    {
        BenchmarkRunner.Run<StringBenchmarks>();
    }
}
```

---

## Branch Protection and Quality Gates

### Branch Protection Rules Setup

```yaml
# .github/workflows/setup-protection.yml
name: Setup Branch Protection

on:
  workflow_dispatch:

jobs:
  setup-protection:
    runs-on: ubuntu-latest
    steps:
    - name: Setup Branch Protection Rules
      uses: actions/github-script@v7
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const protection = {
            required_status_checks: {
              strict: true,
              contexts: [
                'ci',
                'security',
                'quality-gate',
                'build-and-test'
              ]
            },
            enforce_admins: false,
            required_pull_request_reviews: {
              dismiss_stale_reviews: true,
              require_code_owner_reviews: true,
              required_approving_review_count: 2,
              require_last_push_approval: true
            },
            restrictions: null,
            required_linear_history: true,
            allow_force_pushes: false,
            allow_deletions: false,
            required_conversation_resolution: true
          };

          await github.rest.repos.updateBranchProtection({
            owner: context.repo.owner,
            repo: context.repo.repo,
            branch: 'main',
            ...protection
          });
```

### Quality Gate Workflow

```yaml
# .github/workflows/quality-gate.yml
name: Quality Gate

on:
  pull_request:
    branches: [ main ]

jobs:
  quality-gate:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: dotnet build --configuration Release --no-restore

    - name: Run tests with coverage
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --collect:"XPlat Code Coverage" \
          --results-directory ./coverage

    - name: Generate coverage report
      uses: danielpalme/ReportGenerator-GitHub-Action@5.2.0
      with:
        reports: './coverage/**/coverage.cobertura.xml'
        targetdir: 'coveragereport'
        reporttypes: 'JsonSummary'

    - name: Check code coverage threshold
      run: |
        coverage=$(jq -r '.summary.linecoverage' coveragereport/Summary.json)
        echo "Current coverage: $coverage%"

        if (( $(echo "$coverage < 80" | bc -l) )); then
          echo "❌ Code coverage below threshold: $coverage% < 80%"
          exit 1
        fi

        echo "✅ Code coverage meets threshold: $coverage% >= 80%"

    - name: Check for code smells
      run: |
        # Run static analysis
        dotnet build --verbosity normal 2>&1 | tee build-output.txt

        # Check for warnings
        warning_count=$(grep -c "warning" build-output.txt || echo "0")

        if [ "$warning_count" -gt 5 ]; then
          echo "❌ Too many warnings: $warning_count"
          exit 1
        fi

        echo "✅ Warning count acceptable: $warning_count"

    - name: Check commit message format
      run: |
        # Check if PR title follows conventional commits
        pr_title="${{ github.event.pull_request.title }}"

        if ! echo "$pr_title" | grep -qE '^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .+'; then
          echo "❌ PR title doesn't follow conventional commits format"
          echo "Expected: type(scope): description"
          echo "Actual: $pr_title"
          exit 1
        fi

        echo "✅ PR title follows conventional commits format"

    - name: Check for security issues
      run: |
        dotnet list package --vulnerable --include-transitive 2>&1 | tee security-check.txt

        if grep -q "has the following vulnerable packages" security-check.txt; then
          echo "❌ Vulnerable packages detected"
          cat security-check.txt
          exit 1
        fi

        echo "✅ No vulnerable packages found"

    - name: Performance regression check
      run: |
        # This would compare current performance with baseline
        # For example, using benchmark results
        echo "✅ No performance regressions detected"

    - name: Update PR status
      if: always()
      uses: actions/github-script@v7
      with:
        script: |
          const status = '${{ job.status }}' === 'success' ? 'success' : 'failure';
          const description = status === 'success'
            ? 'All quality gates passed'
            : 'Quality gate checks failed';

          github.rest.repos.createCommitStatus({
            owner: context.repo.owner,
            repo: context.repo.repo,
            sha: context.payload.pull_request.head.sha,
            state: status,
            target_url: `${context.serverUrl}/${context.repo.owner}/${context.repo.repo}/actions/runs/${context.runId}`,
            description: description,
            context: 'quality-gate'
          });
```

### Advanced PR Validation

```yaml
# .github/workflows/pr-validation.yml
name: PR Validation

on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches: [ main ]

jobs:
  validate-changes:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Get changed files
      id: changed-files
      uses: tj-actions/changed-files@v44
      with:
        files: |
          src/**/*.cs
          tests/**/*.cs
          *.csproj
          *.sln

    - name: Validate file changes
      if: steps.changed-files.outputs.any_changed == 'true'
      run: |
        echo "Changed files:"
        echo "${{ steps.changed-files.outputs.all_changed_files }}"

        # Check for large files
        for file in ${{ steps.changed-files.outputs.all_changed_files }}; do
          if [ -f "$file" ]; then
            size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file")
            if [ "$size" -gt 1048576 ]; then # 1MB
              echo "❌ File too large: $file ($size bytes)"
              exit 1
            fi
          fi
        done

    - name: Check for breaking changes
      run: |
        # This would use tools like Microsoft.DotNet.ApiCompat
        # to detect breaking API changes
        echo "Checking for breaking changes..."

        # For demonstration, we'll just check if any public interfaces changed
        git diff origin/main..HEAD --name-only | grep -E "\.cs$" | xargs -I {} \
          git diff origin/main..HEAD -- {} | grep -E "^\+.*public.*interface" && {
            echo "⚠️  Potential breaking change detected in public interfaces"
            echo "Please ensure this is intentional and properly versioned"
          } || echo "✅ No obvious breaking changes detected"

    - name: Validate documentation
      run: |
        # Check if code changes include corresponding documentation updates
        has_code_changes=$(echo "${{ steps.changed-files.outputs.all_changed_files }}" | grep -E "\.(cs)$" | wc -l)
        has_doc_changes=$(echo "${{ steps.changed-files.outputs.all_changed_files }}" | grep -E "\.(md|xml)$" | wc -l)

        if [ "$has_code_changes" -gt 0 ] && [ "$has_doc_changes" -eq 0 ]; then
          echo "⚠️  Code changes without documentation updates"
          echo "Consider updating relevant documentation"
        else
          echo "✅ Documentation consideration met"
        fi

    - name: Check test coverage impact
      if: steps.changed-files.outputs.any_changed == 'true'
      run: |
        # This is a simplified check - in practice, you'd run coverage on the branch
        # and compare with main branch coverage

        changed_src_files=$(echo "${{ steps.changed-files.outputs.all_changed_files }}" | grep -E "src/.*\.cs$" | wc -l)
        changed_test_files=$(echo "${{ steps.changed-files.outputs.all_changed_files }}" | grep -E "tests/.*\.cs$" | wc -l)

        if [ "$changed_src_files" -gt 0 ] && [ "$changed_test_files" -eq 0 ]; then
          echo "⚠️  Source code changes without test updates"
          echo "Consider adding or updating tests for your changes"
        else
          echo "✅ Test coverage consideration met"
        fi

  dependency-review:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Dependency Review
      uses: actions/dependency-review-action@v4
      with:
        fail-on-severity: moderate
        allow-licenses: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, GPL-3.0
        deny-licenses: GPL-2.0, LGPL-2.1
```

---

## Enterprise Patterns and Advanced Scenarios

### Mono-repo Multi-Project Build

```yaml
# .github/workflows/monorepo.yml
name: Monorepo Build

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      projects: ${{ steps.projects.outputs.projects }}

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Detect changed projects
      id: projects
      run: |
        # Get list of changed files
        if [ "${{ github.event_name }}" = "pull_request" ]; then
          changed_files=$(git diff --name-only origin/${{ github.base_ref }}..HEAD)
        else
          changed_files=$(git diff --name-only HEAD~1..HEAD)
        fi

        # Detect which projects are affected
        projects=()

        for project_dir in src/*/; do
          project_name=$(basename "$project_dir")

          if echo "$changed_files" | grep -q "^$project_dir"; then
            projects+=("$project_name")
          fi
        done

        # Always build shared libraries if any project changes
        if [ ${#projects[@]} -gt 0 ]; then
          projects+=("Shared.Core" "Shared.Infrastructure")
        fi

        # Remove duplicates and convert to JSON array
        unique_projects=($(printf "%s\n" "${projects[@]}" | sort -u))
        json_projects=$(printf '"%s"\n' "${unique_projects[@]}" | paste -sd, | sed 's/,/","/g' | sed 's/^/["/' | sed 's/$/"]/')

        echo "projects=$json_projects" >> $GITHUB_OUTPUT
        echo "Detected projects: $json_projects"

  build:
    needs: detect-changes
    if: needs.detect-changes.outputs.projects != '[]'

    strategy:
      matrix:
        project: ${{ fromJson(needs.detect-changes.outputs.projects) }}
        os: [ubuntu-latest, windows-latest]

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Cache NuGet packages
      uses: actions/cache@v4
      with:
        path: ~/.nuget/packages
        key: ${{ runner.os }}-nuget-${{ matrix.project }}-${{ hashFiles(format('src/{0}/*.csproj', matrix.project)) }}
        restore-keys: |
          ${{ runner.os }}-nuget-${{ matrix.project }}-
          ${{ runner.os }}-nuget-

    - name: Restore project
      run: dotnet restore src/${{ matrix.project }}

    - name: Build project
      run: |
        dotnet build src/${{ matrix.project }} \
          --configuration Release \
          --no-restore

    - name: Test project
      run: |
        test_project="tests/${{ matrix.project }}.Tests"
        if [ -d "$test_project" ]; then
          dotnet test "$test_project" \
            --configuration Release \
            --no-build \
            --collect:"XPlat Code Coverage" \
            --results-directory "./coverage/${{ matrix.project }}-${{ matrix.os }}"
        else
          echo "No tests found for ${{ matrix.project }}"
        fi

    - name: Upload coverage
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: coverage-${{ matrix.project }}-${{ matrix.os }}
        path: ./coverage/${{ matrix.project }}-${{ matrix.os }}

    - name: Package project (if applicable)
      if: matrix.os == 'ubuntu-latest' && (contains(matrix.project, 'Library') || contains(matrix.project, 'Shared'))
      run: |
        dotnet pack src/${{ matrix.project }} \
          --configuration Release \
          --no-build \
          --output ./packages

    - name: Upload packages
      if: matrix.os == 'ubuntu-latest' && (contains(matrix.project, 'Library') || contains(matrix.project, 'Shared'))
      uses: actions/upload-artifact@v4
      with:
        name: packages-${{ matrix.project }}
        path: ./packages

  integration-tests:
    needs: [detect-changes, build]
    if: needs.detect-changes.outputs.projects != '[]'
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: integrationdb
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
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Run integration tests
      env:
        ConnectionStrings__DefaultConnection: "Host=localhost;Database=integrationdb;Username=postgres;Password=postgres"
        ConnectionStrings__Redis: "localhost:6379"
      run: |
        dotnet test tests/Integration.Tests \
          --configuration Release \
          --logger trx \
          --collect:"XPlat Code Coverage" \
          --results-directory ./integration-results

    - name: Upload integration results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: integration-test-results
        path: ./integration-results
```

### Multi-Environment Deployment Pipeline

```yaml
# .github/workflows/multi-env-deploy.yml
name: Multi-Environment Deployment

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        default: 'staging'
        type: choice
        options:
          - development
          - staging
          - production

      services:
        description: 'Services to deploy (comma-separated)'
        required: false
        default: 'all'

jobs:
  prepare:
    runs-on: ubuntu-latest
    outputs:
      environment: ${{ steps.env.outputs.environment }}
      services: ${{ steps.services.outputs.services }}
      version: ${{ steps.version.outputs.version }}

    steps:
    - uses: actions/checkout@v4

    - name: Determine environment
      id: env
      run: |
        if [ "${{ github.event_name }}" = "workflow_dispatch" ]; then
          echo "environment=${{ github.event.inputs.environment }}" >> $GITHUB_OUTPUT
        elif [ "${{ github.ref }}" = "refs/heads/main" ]; then
          echo "environment=staging" >> $GITHUB_OUTPUT
        elif [[ "${{ github.ref }}" =~ refs/tags/v.* ]]; then
          echo "environment=production" >> $GITHUB_OUTPUT
        else
          echo "environment=development" >> $GITHUB_OUTPUT
        fi

    - name: Determine services
      id: services
      run: |
        if [ "${{ github.event.inputs.services }}" = "all" ] || [ -z "${{ github.event.inputs.services }}" ]; then
          services='["api", "web", "worker"]'
        else
          # Convert comma-separated list to JSON array
          services=$(echo "${{ github.event.inputs.services }}" | jq -R -c 'split(",")')
        fi
        echo "services=$services" >> $GITHUB_OUTPUT

    - name: Determine version
      id: version
      run: |
        if [[ "${{ github.ref }}" =~ refs/tags/v.* ]]; then
          echo "version=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT
        else
          echo "version=${{ github.sha }}" >> $GITHUB_OUTPUT
        fi

  build:
    runs-on: ubuntu-latest
    needs: prepare
    strategy:
      matrix:
        service: ${{ fromJson(needs.prepare.outputs.services) }}

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: '8.0.x'

    - name: Build and publish service
      run: |
        dotnet publish src/${{ matrix.service }} \
          --configuration Release \
          --output ./publish/${{ matrix.service }} \
          --self-contained true \
          --runtime linux-x64 \
          -p:Version=${{ needs.prepare.outputs.version }}

    - name: Upload artifacts
      uses: actions/upload-artifact@v4
      with:
        name: ${{ matrix.service }}-build
        path: ./publish/${{ matrix.service }}

  deploy-development:
    needs: [prepare, build]
    if: needs.prepare.outputs.environment == 'development'
    runs-on: ubuntu-latest
    environment: development

    strategy:
      matrix:
        service: ${{ fromJson(needs.prepare.outputs.services) }}

    steps:
    - name: Deploy to development
      run: |
        echo "Deploying ${{ matrix.service }} to development environment"
        # Development deployment logic here

  deploy-staging:
    needs: [prepare, build]
    if: needs.prepare.outputs.environment == 'staging'
    runs-on: ubuntu-latest
    environment: staging

    strategy:
      matrix:
        service: ${{ fromJson(needs.prepare.outputs.services) }}

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: ${{ matrix.service }}-build
        path: ./app

    - name: Deploy to staging
      uses: azure/webapps-deploy@v3
      with:
        app-name: 'myapp-${{ matrix.service }}-staging'
        publish-profile: ${{ secrets[format('AZURE_WEBAPP_STAGING_{0}_PUBLISH_PROFILE', upper(matrix.service))] }}
        package: './app'

    - name: Health check
      run: |
        sleep 30
        curl -f https://myapp-${{ matrix.service }}-staging.azurewebsites.net/health

  deploy-production:
    needs: [prepare, build]
    if: needs.prepare.outputs.environment == 'production'
    runs-on: ubuntu-latest
    environment: production

    strategy:
      matrix:
        service: ${{ fromJson(needs.prepare.outputs.services) }}

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: ${{ matrix.service }}-build
        path: ./app

    - name: Blue-Green Deployment
      run: |
        # Implement blue-green deployment logic
        echo "Deploying ${{ matrix.service }} to production using blue-green strategy"

    - name: Traffic switching
      run: |
        # Switch traffic to new version
        echo "Switching traffic to new version of ${{ matrix.service }}"

    - name: Smoke tests
      run: |
        # Run production smoke tests
        curl -f https://api.myapp.com/health
        curl -f https://myapp.com

  notify:
    needs: [prepare, deploy-development, deploy-staging, deploy-production]
    if: always()
    runs-on: ubuntu-latest

    steps:
    - name: Notify deployment status
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        text: |
          Deployment to ${{ needs.prepare.outputs.environment }} completed
          Services: ${{ join(fromJson(needs.prepare.outputs.services), ', ') }}
          Version: ${{ needs.prepare.outputs.version }}
          Status: ${{ job.status }}
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Matrix Testing with External Dependencies

```yaml
# .github/workflows/matrix-testing.yml
name: Matrix Testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *' # Nightly

jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        dotnet: ['6.0.x', '8.0.x']
        database: [postgresql, sqlserver, sqlite]
        exclude:
          # SQL Server only on Windows and Linux
          - os: macos-latest
            database: sqlserver
        include:
          # Add specific configurations
          - os: ubuntu-latest
            dotnet: '8.0.x'
            database: postgresql
            run-performance: true
          - os: windows-latest
            dotnet: '8.0.x'
            database: sqlserver
            run-integration: true

    runs-on: ${{ matrix.os }}

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      sqlserver:
        image: mcr.microsoft.com/mssql/server:2022-latest
        env:
          SA_PASSWORD: 'YourStrong@Passw0rd'
          ACCEPT_EULA: 'Y'
        options: >-
          --health-cmd "/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P YourStrong@Passw0rd -Q 'SELECT 1'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 1433:1433

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET ${{ matrix.dotnet }}
      uses: actions/setup-dotnet@v4
      with:
        dotnet-version: ${{ matrix.dotnet }}

    - name: Setup database connection
      run: |
        case "${{ matrix.database }}" in
          postgresql)
            echo "DATABASE_URL=Host=localhost;Database=testdb;Username=postgres;Password=postgres" >> $GITHUB_ENV
            ;;
          sqlserver)
            echo "DATABASE_URL=Server=localhost;Database=testdb;User Id=sa;Password=YourStrong@Passw0rd;TrustServerCertificate=true" >> $GITHUB_ENV
            ;;
          sqlite)
            echo "DATABASE_URL=Data Source=test.db" >> $GITHUB_ENV
            ;;
        esac

    - name: Restore dependencies
      run: dotnet restore

    - name: Build
      run: |
        dotnet build \
          --configuration Release \
          --no-restore \
          -p:DatabaseProvider=${{ matrix.database }}

    - name: Run unit tests
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --filter "Category=Unit" \
          --logger trx \
          --collect:"XPlat Code Coverage" \
          --results-directory "TestResults-${{ matrix.os }}-${{ matrix.dotnet }}-${{ matrix.database }}"

    - name: Run integration tests
      if: matrix.run-integration == true
      env:
        ConnectionStrings__DefaultConnection: ${{ env.DATABASE_URL }}
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --filter "Category=Integration" \
          --logger trx \
          --results-directory "IntegrationResults-${{ matrix.os }}-${{ matrix.dotnet }}-${{ matrix.database }}"

    - name: Run performance tests
      if: matrix.run-performance == true
      run: |
        dotnet test \
          --configuration Release \
          --no-build \
          --filter "Category=Performance" \
          --logger trx \
          --results-directory "PerformanceResults-${{ matrix.os }}-${{ matrix.dotnet }}-${{ matrix.database }}"

    - name: Upload test results
      uses: actions/upload-artifact@v4
      if: always()
      with:
        name: test-results-${{ matrix.os }}-${{ matrix.dotnet }}-${{ matrix.database }}
        path: |
          TestResults-*/**/*.trx
          IntegrationResults-*/**/*.trx
          PerformanceResults-*/**/*.trx
          TestResults-*/**/*.xml
```

### Self-Hosted Runner Configuration

```yaml
# .github/workflows/self-hosted.yml
name: Self-Hosted Runners

on:
  push:
    branches: [ main ]
    paths:
      - 'src/HighPerformance/**'
      - 'tests/Performance/**'

jobs:
  performance-testing:
    runs-on: [self-hosted, performance, gpu]

    steps:
    - uses: actions/checkout@v4

    - name: Setup .NET
      run: |
        # Install .NET if not present on self-hosted runner
        if ! command -v dotnet &> /dev/null; then
          wget https://dot.net/v1/dotnet-install.sh -O dotnet-install.sh
          chmod +x dotnet-install.sh
          ./dotnet-install.sh --version latest
          export PATH=$PATH:$HOME/.dotnet
        fi

    - name: Check hardware capabilities
      run: |
        echo "CPU Info:"
        lscpu
        echo "Memory Info:"
        free -h
        echo "GPU Info:"
        nvidia-smi || echo "No NVIDIA GPU detected"

    - name: Build with native optimizations
      run: |
        dotnet build \
          --configuration Release \
          -p:EnableNativeOptimizations=true \
          -p:TargetArchitecture=x64

    - name: Run GPU-accelerated tests
      run: |
        dotnet test tests/Performance.GPU \
          --configuration Release \
          --logger trx

    - name: Benchmark critical paths
      run: |
        dotnet run --project benchmarks/CriticalPath \
          --configuration Release \
          -- --exporters json html

    - name: Upload performance results
      uses: actions/upload-artifact@v4
      with:
        name: gpu-performance-results
        path: |
          BenchmarkDotNet.Artifacts/**/*
          **/*.trx

  secure-build:
    runs-on: [self-hosted, secure, isolated]

    steps:
    - name: Secure checkout
      uses: actions/checkout@v4
      with:
        token: ${{ secrets.SECURE_CHECKOUT_TOKEN }}

    - name: Verify runner security
      run: |
        # Check for required security tools
        command -v clamav-scan >/dev/null 2>&1 || { echo "ClamAV not found"; exit 1; }
        command -v rkhunter >/dev/null 2>&1 || { echo "rkhunter not found"; exit 1; }

    - name: Scan for malware
      run: |
        clamscan -r . --exclude-dir=.git

    - name: Build with security scanning
      run: |
        dotnet build --configuration Release

        # Run additional security scans
        dotnet list package --vulnerable --include-transitive

    - name: Secure artifact handling
      run: |
        # Package artifacts securely
        tar -czf secure-build.tar.gz -C publish .
        gpg --symmetric --cipher-algo AES256 secure-build.tar.gz

    - name: Upload encrypted artifacts
      uses: actions/upload-artifact@v4
      with:
        name: secure-build
        path: secure-build.tar.gz.gpg
```

---

## Best Practices Summary

### Essential Practices

1. **Always Use Caching**
   ```yaml
   - uses: actions/cache@v4
     with:
       path: ~/.nuget/packages
       key: ${{ runner.os }}-nuget-${{ hashFiles('**/*.csproj') }}
   ```

2. **Implement Proper Error Handling**
   ```yaml
   - name: Step with error handling
     run: |
       command_that_might_fail || {
         echo "Command failed, but continuing..."
         echo "STEP_FAILED=true" >> $GITHUB_ENV
       }
   ```

3. **Use Matrix Builds Efficiently**
   ```yaml
   strategy:
     fail-fast: false
     matrix:
       os: [ubuntu-latest, windows-latest]
       exclude:
         - os: windows-latest
           feature: linux-only
   ```

4. **Secure Secret Handling**
   ```yaml
   - name: Use secrets properly
     env:
       SECRET_VALUE: ${{ secrets.SECRET_NAME }}
     run: |
       # Never echo secrets directly
       echo "Using secret (hidden)"
   ```

5. **Implement Comprehensive Testing**
   - Unit tests with high coverage (>80%)
   - Integration tests with real dependencies
   - Performance benchmarks
   - Security scanning
   - Mutation testing (when applicable)

6. **Use Semantic Versioning**
   - Implement conventional commits
   - Automate version bumping
   - Generate changelogs automatically

7. **Monitor and Alert**
   - Set up deployment notifications
   - Monitor performance metrics
   - Track security vulnerabilities
   - Alert on quality gate failures

This comprehensive guide provides a solid foundation for implementing professional CI/CD practices with GitHub Actions for C# projects, covering everything from basic builds to enterprise-scale deployments with advanced security and monitoring capabilities.