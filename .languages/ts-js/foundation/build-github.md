# GitHub Actions CI/CD for TypeScript/JavaScript

## Basic CI/CD Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  NODE_VERSION: '18'

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        node-version: [16, 18, 20]

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ matrix.node-version }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run linting
      run: npm run lint

    - name: Run type checking
      run: npm run typecheck

    - name: Run tests
      run: npm run test:coverage

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        token: ${{ secrets.CODECOV_TOKEN }}

  build:
    runs-on: ubuntu-latest
    needs: test

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Build project
      run: npm run build

    - name: Upload build artifacts
      uses: actions/upload-artifact@v4
      with:
        name: dist
        path: dist/

  deploy:
    runs-on: ubuntu-latest
    needs: [test, build]
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: dist

    - name: Deploy to production
      run: echo "Deploying to production"
```

## Advanced Multi-Environment Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy to Multiple Environments

on:
  push:
    branches: [ main, staging, develop ]
  release:
    types: [ published ]

jobs:
  test-and-build:
    runs-on: ubuntu-latest
    outputs:
      environment: ${{ steps.determine-env.outputs.environment }}

    steps:
    - name: Checkout
      uses: actions/checkout@v4

    - name: Determine environment
      id: determine-env
      run: |
        if [[ ${{ github.ref }} == 'refs/heads/main' ]]; then
          echo "environment=production" >> $GITHUB_OUTPUT
        elif [[ ${{ github.ref }} == 'refs/heads/staging' ]]; then
          echo "environment=staging" >> $GITHUB_OUTPUT
        else
          echo "environment=development" >> $GITHUB_OUTPUT
        fi

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run tests
      run: npm test

    - name: Build for environment
      run: |
        if [[ "${{ steps.determine-env.outputs.environment }}" == "production" ]]; then
          npm run build:prod
        elif [[ "${{ steps.determine-env.outputs.environment }}" == "staging" ]]; then
          npm run build:staging
        else
          npm run build:dev
        fi

    - name: Upload artifacts
      uses: actions/upload-artifact@v4
      with:
        name: build-${{ steps.determine-env.outputs.environment }}
        path: dist/

  deploy:
    needs: test-and-build
    runs-on: ubuntu-latest
    environment: ${{ needs.test-and-build.outputs.environment }}

    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v4
      with:
        name: build-${{ needs.test-and-build.outputs.environment }}

    - name: Deploy to ${{ needs.test-and-build.outputs.environment }}
      run: |
        echo "Deploying to ${{ needs.test-and-build.outputs.environment }}"
        # Add your deployment commands here
```

## Complete Package Publishing Pipeline

```yaml
# .github/workflows/publish.yml
name: Publish Package

on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      version-type:
        description: 'Version increment type'
        required: true
        type: choice
        options: ['patch', 'minor', 'major']

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write

    steps:
    - name: Checkout
      uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        registry-url: 'https://registry.npmjs.org'
        cache: 'npm'

    - name: Configure Git
      run: |
        git config user.name "github-actions[bot]"
        git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

    - name: Install dependencies
      run: npm ci

    - name: Run tests
      run: npm test

    - name: Build package
      run: npm run build

    - name: Version bump (if workflow_dispatch)
      if: github.event_name == 'workflow_dispatch'
      run: |
        npm version ${{ github.event.inputs.version-type }}
        git push origin main --tags

    - name: Publish to npm
      run: npm publish
      env:
        NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}

    - name: Setup Node.js for GitHub Packages
      uses: actions/setup-node@v4
      with:
        registry-url: 'https://npm.pkg.github.com'

    - name: Publish to GitHub Packages
      run: npm publish
      env:
        NODE_AUTH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

This comprehensive build system covers all aspects from basic CI/CD to advanced deployment pipelines, ensuring robust and reliable TypeScript/JavaScript project automation on GitHub.