# Python GitHub Actions CI/CD Pipeline

Comprehensive guide to setting up GitHub Actions for Python projects with best practices for testing, building, and deployment.

## Basic CI Workflow

### Simple Test Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        python-version: [3.9, 3.10, 3.11, 3.12]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
        pip install -r requirements-dev.txt
    
    - name: Run tests
      run: |
        python -m pytest --cov=src --cov-report=xml
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml
        flags: unittests
        name: codecov-umbrella
```

### Advanced Multi-Job Workflow

```yaml
# .github/workflows/advanced-ci.yml
name: Advanced CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 0'  # Weekly

env:
  PYTHON_DEFAULT_VERSION: "3.11"

jobs:
  lint:
    name: Code Quality
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: ${{ env.PYTHON_DEFAULT_VERSION }}
    
    - name: Cache pip packages
      uses: actions/cache@v3
      with:
        path: ~/.cache/pip
        key: ${{ runner.os }}-pip-${{ hashFiles('**/requirements*.txt') }}
        restore-keys: |
          ${{ runner.os }}-pip-
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements-dev.txt
    
    - name: Run black
      run: black --check src tests
    
    - name: Run isort
      run: isort --check-only src tests
    
    - name: Run flake8
      run: flake8 src tests
    
    - name: Run mypy
      run: mypy src
    
    - name: Run bandit security check
      run: bandit -r src/
    
    - name: Run safety check
      run: safety check

  test:
    name: Tests
    runs-on: ${{ matrix.os }}
    needs: lint
    
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        python-version: ["3.9", "3.10", "3.11", "3.12"]
        exclude:
          - os: windows-latest
            python-version: "3.9"
          - os: macos-latest
            python-version: "3.9"
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Cache pip packages
      uses: actions/cache@v3
      with:
        path: ~/.cache/pip
        key: ${{ runner.os }}-pip-${{ hashFiles('**/requirements*.txt') }}
        restore-keys: |
          ${{ runner.os }}-pip-
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -e .[dev]
    
    - name: Run tests
      run: |
        python -m pytest -v --cov=src --cov-report=xml --cov-report=term
    
    - name: Upload coverage to Codecov
      if: matrix.os == 'ubuntu-latest' && matrix.python-version == env.PYTHON_DEFAULT_VERSION
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml
        flags: unittests
        name: codecov-umbrella

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'
    
    - name: Upload Trivy scan results to GitHub Security tab
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'

  docs:
    name: Documentation
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: ${{ env.PYTHON_DEFAULT_VERSION }}
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements-dev.txt
    
    - name: Build documentation
      run: |
        cd docs
        make html
    
    - name: Upload documentation artifacts
      uses: actions/upload-artifact@v3
      with:
        name: documentation
        path: docs/_build/html
```

## Poetry-Based Workflows

### Poetry CI Workflow

```yaml
# .github/workflows/poetry-ci.yml
name: Poetry CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        python-version: ["3.9", "3.10", "3.11", "3.12"]

    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install Poetry
      uses: snok/install-poetry@v1
      with:
        virtualenvs-create: true
        virtualenvs-in-project: true
        installer-parallel: true
    
    - name: Load cached venv
      id: cached-poetry-dependencies
      uses: actions/cache@v3
      with:
        path: .venv
        key: venv-${{ runner.os }}-${{ matrix.python-version }}-${{ hashFiles('**/poetry.lock') }}
    
    - name: Install dependencies
      if: steps.cached-poetry-dependencies.outputs.cache-hit != 'true'
      run: poetry install --no-interaction --no-root
    
    - name: Install project
      run: poetry install --no-interaction
    
    - name: Run tests
      run: |
        source .venv/bin/activate
        pytest --cov=src --cov-report=xml
    
    - name: Upload coverage
      if: matrix.os == 'ubuntu-latest' && matrix.python-version == '3.11'
      uses: codecov/codecov-action@v3
```

## Docker-based Workflows

### Docker Build and Test

```yaml
# .github/workflows/docker.yml
name: Docker

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Log in to Container Registry
      if: github.event_name != 'pull_request'
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

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

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
```

### Multi-stage Dockerfile

```dockerfile
# Dockerfile
FROM python:3.11-slim as base

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Development stage
FROM base as dev
WORKDIR /app
COPY requirements-dev.txt .
RUN pip install -r requirements-dev.txt
COPY . .
RUN pip install -e .
CMD ["python", "-m", "pytest"]

# Production stage
FROM base as prod
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY src/ ./src/
COPY setup.py pyproject.toml ./
RUN pip install .

# Create non-root user
RUN useradd --create-home --shell /bin/bash app
USER app

CMD ["python", "-m", "my_project"]
```

## Release and Deployment

### Automated Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install build twine
    
    - name: Build package
      run: python -m build
    
    - name: Publish to PyPI
      env:
        TWINE_USERNAME: __token__
        TWINE_PASSWORD: ${{ secrets.PYPI_API_TOKEN }}
      run: |
        twine upload dist/*
    
    - name: Create GitHub Release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ github.ref }}
        draft: false
        prerelease: false
    
    - name: Upload Release Assets
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ steps.create_release.outputs.upload_url }}
        asset_path: ./dist/
        asset_name: python-package
        asset_content_type: application/zip
```

### Semantic Release

```yaml
# .github/workflows/semantic-release.yml
name: Semantic Release

on:
  push:
    branches:
      - main

jobs:
  release:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
        token: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Python Semantic Release
      uses: relekang/python-semantic-release@master
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        repository_username: __token__
        repository_password: ${{ secrets.PYPI_API_TOKEN }}
```

## Deployment Workflows

### Deploy to AWS Lambda

```yaml
# .github/workflows/deploy-lambda.yml
name: Deploy to AWS Lambda

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt -t .
    
    - name: Create deployment package
      run: |
        zip -r deployment-package.zip .
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
    
    - name: Deploy to Lambda
      run: |
        aws lambda update-function-code \
          --function-name my-python-function \
          --zip-file fileb://deployment-package.zip
```

### Deploy to Heroku

```yaml
# .github/workflows/deploy-heroku.yml
name: Deploy to Heroku

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to Heroku
      uses: akhileshns/heroku-deploy@v3.12.12
      with:
        heroku_api_key: ${{ secrets.HEROKU_API_KEY }}
        heroku_app_name: "my-python-app"
        heroku_email: "your-email@example.com"
```

## Advanced Workflows

### Matrix Testing with Services

```yaml
# .github/workflows/services.yml
name: Test with Services

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11']
        database: [postgresql, mysql, sqlite]
    
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_USER: postgres
          POSTGRES_DB: test
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
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements-test.txt
    
    - name: Run tests
      env:
        DATABASE_URL: postgresql://postgres:postgres@localhost:5432/test
        REDIS_URL: redis://localhost:6379
      run: |
        python -m pytest --cov=src
```

### Performance Benchmarking

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmark

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements-dev.txt
        pip install pytest-benchmark
    
    - name: Run benchmarks
      run: |
        python -m pytest benchmarks/ --benchmark-json=benchmark.json
    
    - name: Store benchmark result
      uses: benchmark-action/github-action-benchmark@v1
      with:
        name: Python Benchmark
        tool: 'pytest'
        output-file-path: benchmark.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        auto-push: true
        comment-on-alert: true
        alert-threshold: '200%'
        fail-on-alert: true
```

## Environment-specific Configurations

### Development Environment

```yaml
# .github/workflows/dev.yml
name: Development

on:
  push:
    branches: [ develop ]

jobs:
  deploy-dev:
    runs-on: ubuntu-latest
    environment: development
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to development
      env:
        DEV_API_URL: ${{ secrets.DEV_API_URL }}
        DEV_API_KEY: ${{ secrets.DEV_API_KEY }}
      run: |
        echo "Deploying to development environment"
        # Deployment commands here
```

### Production Environment

```yaml
# .github/workflows/prod.yml
name: Production

on:
  release:
    types: [published]

jobs:
  deploy-prod:
    runs-on: ubuntu-latest
    environment: production
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Deploy to production
      env:
        PROD_API_URL: ${{ secrets.PROD_API_URL }}
        PROD_API_KEY: ${{ secrets.PROD_API_KEY }}
      run: |
        echo "Deploying to production environment"
        # Production deployment commands here
```

## Workflow Best Practices

### Conditional Steps

```yaml
steps:
- name: Run only on main branch
  if: github.ref == 'refs/heads/main'
  run: echo "This runs only on main branch"

- name: Run only on PR
  if: github.event_name == 'pull_request'
  run: echo "This runs only on pull requests"

- name: Run only if files changed
  uses: dorny/paths-filter@v2
  id: changes
  with:
    filters: |
      python:
        - 'src/**/*.py'
        - 'tests/**/*.py'

- name: Run tests if Python files changed
  if: steps.changes.outputs.python == 'true'
  run: pytest
```

### Reusable Workflows

```yaml
# .github/workflows/reusable-test.yml
name: Reusable Test Workflow

on:
  workflow_call:
    inputs:
      python-version:
        required: false
        default: '3.11'
        type: string
      os:
        required: false
        default: 'ubuntu-latest'
        type: string

jobs:
  test:
    runs-on: ${{ inputs.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: ${{ inputs.python-version }}
    
    - name: Run tests
      run: pytest
```

```yaml
# Using reusable workflow
name: CI

on: [push, pull_request]

jobs:
  test:
    uses: ./.github/workflows/reusable-test.yml
    with:
      python-version: '3.11'
      os: 'ubuntu-latest'
```

This comprehensive GitHub Actions setup provides robust CI/CD pipelines for Python projects, covering testing, security, documentation, and deployment scenarios.