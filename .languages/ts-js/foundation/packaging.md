# TypeScript/JavaScript Packaging Guide

## NPM Package Ecosystem Fundamentals

### Package Architecture Principles
```javascript
// Modern package structure follows these principles:
// 1. Support both ESM and CommonJS formats
// 2. Provide TypeScript declarations
// 3. Tree-shakable exports
// 4. Conditional exports for different environments
// 5. Optimized bundle sizes
```

### Complete package.json Configuration
```json
{
  "name": "@myorg/my-package",
  "version": "1.0.0",
  "description": "A comprehensive TypeScript package",
  "keywords": ["typescript", "javascript", "library", "esm", "commonjs"],
  "author": "Your Name <email@example.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/myorg/my-package.git"
  },
  "bugs": {
    "url": "https://github.com/myorg/my-package/issues"
  },
  "homepage": "https://github.com/myorg/my-package#readme",

  "main": "./dist/index.cjs",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts",
  "browser": "./dist/index.browser.js",

  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs",
      "require": "./dist/index.cjs",
      "browser": "./dist/index.browser.js"
    },
    "./utils": {
      "types": "./dist/utils.d.ts",
      "import": "./dist/utils.mjs",
      "require": "./dist/utils.cjs"
    },
    "./plugins/*": {
      "types": "./dist/plugins/*.d.ts",
      "import": "./dist/plugins/*.mjs",
      "require": "./dist/plugins/*.cjs"
    },
    "./package.json": "./package.json"
  },

  "files": [
    "dist",
    "README.md",
    "LICENSE",
    "CHANGELOG.md"
  ],

  "engines": {
    "node": ">=18.0.0",
    "npm": ">=8.0.0"
  },

  "scripts": {
    "build": "npm run clean && npm run build:types && npm run build:bundle",
    "build:types": "tsc --emitDeclarationOnly",
    "build:bundle": "rollup -c",
    "clean": "rimraf dist",
    "dev": "rollup -c --watch",
    "test": "jest",
    "test:coverage": "jest --coverage",
    "test:watch": "jest --watch",
    "lint": "eslint src --ext .ts,.js",
    "lint:fix": "eslint src --ext .ts,.js --fix",
    "typecheck": "tsc --noEmit",
    "prepublishOnly": "npm run lint && npm run typecheck && npm run test && npm run build",
    "release:patch": "npm version patch && npm publish",
    "release:minor": "npm version minor && npm publish",
    "release:major": "npm version major && npm publish",
    "release:beta": "npm version prerelease --preid=beta && npm publish --tag beta"
  },

  "dependencies": {},
  "devDependencies": {
    "typescript": "^5.3.0",
    "@types/node": "^20.0.0",
    "@rollup/plugin-typescript": "^11.0.0",
    "@rollup/plugin-node-resolve": "^15.0.0",
    "@rollup/plugin-commonjs": "^25.0.0",
    "rollup": "^4.0.0",
    "rollup-plugin-dts": "^6.0.0",
    "jest": "^29.0.0",
    "@types/jest": "^29.0.0",
    "ts-jest": "^29.0.0",
    "eslint": "^8.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "rimraf": "^5.0.0"
  },

  "peerDependencies": {
    "react": ">=16.8.0"
  },
  "peerDependenciesMeta": {
    "react": {
      "optional": true
    }
  }
}
```

### Advanced Package Exports
```json
{
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": {
        "development": "./dist/index.dev.mjs",
        "production": "./dist/index.prod.mjs",
        "default": "./dist/index.mjs"
      },
      "require": {
        "development": "./dist/index.dev.cjs",
        "production": "./dist/index.prod.cjs",
        "default": "./dist/index.cjs"
      },
      "browser": "./dist/index.browser.js",
      "react-native": "./dist/index.native.js"
    },
    "./server": {
      "types": "./dist/server.d.ts",
      "node": "./dist/server.node.js",
      "default": "./dist/server.js"
    },
    "./client": {
      "types": "./dist/client.d.ts",
      "browser": "./dist/client.browser.js",
      "default": "./dist/client.js"
    }
  }
}
```

## Build Configuration

### TypeScript for Libraries
```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "node",
    "declaration": true,
    "declarationMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src/**/*"],
  "exclude": ["**/*.test.ts", "**/*.spec.ts"]
}
```

### Modern Rollup Configuration
```javascript
// rollup.config.js
import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import terser from '@rollup/plugin-terser';
import dts from 'rollup-plugin-dts';
import { readFileSync } from 'fs';

const packageJson = JSON.parse(readFileSync('./package.json', 'utf8'));
const isProduction = process.env.NODE_ENV === 'production';

const external = [
  ...Object.keys(packageJson.dependencies || {}),
  ...Object.keys(packageJson.peerDependencies || {}),
  'react', 'react-dom', 'node:*'
];

const basePlugins = [
  nodeResolve({
    preferBuiltins: true,
    exportConditions: ['node']
  }),
  commonjs(),
  typescript({
    tsconfig: './tsconfig.json',
    exclude: ['**/*.test.ts', '**/*.spec.ts', 'examples/**/*']
  })
];

export default [
  // ESM Build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.mjs',
      format: 'esm',
      sourcemap: true,
      exports: 'named'
    },
    plugins: [
      ...basePlugins,
      isProduction && terser({
        compress: {
          pure_getters: true,
          unsafe: true,
          unsafe_comps: true
        }
      })
    ].filter(Boolean),
    external
  },

  // CommonJS Build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.cjs',
      format: 'cjs',
      sourcemap: true,
      exports: 'named'
    },
    plugins: [
      ...basePlugins,
      isProduction && terser()
    ].filter(Boolean),
    external
  },

  // Browser Build (IIFE)
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.browser.js',
      format: 'iife',
      name: 'MyPackage',
      sourcemap: true,
      globals: {
        'react': 'React',
        'react-dom': 'ReactDOM'
      }
    },
    plugins: [
      nodeResolve({ browser: true }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        target: 'ES2017'
      }),
      terser({
        compress: {
          pure_getters: true,
          unsafe: true
        }
      })
    ],
    external: ['react', 'react-dom']
  },

  // Type Definitions
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.d.ts',
      format: 'esm'
    },
    plugins: [dts()],
    external
  }
];
```

### Alternative: Modern Build with Vite
```javascript
// vite.config.ts
import { defineConfig } from 'vite';
import { resolve } from 'path';
import dts from 'vite-plugin-dts';

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'MyPackage',
      formats: ['es', 'cjs', 'iife'],
      fileName: (format) => {
        switch (format) {
          case 'es': return 'index.mjs'
          case 'cjs': return 'index.cjs'
          case 'iife': return 'index.browser.js'
          default: return `index.${format}.js`
        }
      }
    },
    rollupOptions: {
      external: ['react', 'react-dom'],
      output: {
        globals: {
          'react': 'React',
          'react-dom': 'ReactDOM'
        }
      }
    },
    sourcemap: true,
    minify: 'terser'
  },
  plugins: [
    dts({
      insertTypesEntry: true,
      rollupTypes: true
    })
  ]
});
```

## Distribution and Publishing

### Comprehensive NPM Workflow
```bash
# Package Development Cycle
npm init -y                           # Initialize package
npm install --save-dev typescript     # Install TypeScript
npm run build                         # Build the package
npm pack                              # Create tarball for testing
npm install ./my-package-1.0.0.tgz   # Test local installation

# Version Management (Semantic Versioning)
npm version patch     # 1.0.0 -> 1.0.1 (bug fixes)
npm version minor     # 1.0.0 -> 1.1.0 (new features)
npm version major     # 1.0.0 -> 2.0.0 (breaking changes)
npm version prerelease --preid=alpha  # 1.0.0 -> 1.0.1-alpha.0
npm version prerelease --preid=beta   # 1.0.0 -> 1.0.1-beta.0
npm version prerelease --preid=rc     # 1.0.0 -> 1.0.1-rc.0

# Publishing Workflow
npm whoami                   # Verify login
npm config list             # Check configuration
npm run prepublishOnly       # Run pre-publish checks
npm publish                  # Publish to registry
npm publish --tag beta       # Publish with beta tag
npm publish --dry-run        # Simulate publishing
npm publish --access public  # For scoped packages

# Registry Management
npm config set registry https://registry.npmjs.org/
npm config set @myorg:registry https://npm.pkg.github.com/
npm adduser --registry https://registry.npmjs.org/
npm adduser --registry https://npm.pkg.github.com/

# Package Management
npm deprecate my-package@1.0.0 "Please use version 2.0.0"
npm unpublish my-package@1.0.0 --force  # Within 72 hours only
npm owner add <username> <package>       # Add maintainer
npm owner list <package>                 # List maintainers

# Distribution Tags
npm dist-tag add my-package@1.0.1 latest
npm dist-tag add my-package@2.0.0-beta.1 beta
npm dist-tag list my-package
npm dist-tag rm my-package beta
```

### Package Release Automation
```yaml
# .github/workflows/release.yml
name: Release and Publish

on:
  push:
    tags: ['v*']
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release'
        required: true
        type: choice
        options: ['patch', 'minor', 'major', 'prerelease']

env:
  NODE_VERSION: '20'

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 21]
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js ${{ matrix.node-version }}
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

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          token: ${{ secrets.CODECOV_TOKEN }}

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build package
        run: npm run build

      - name: Pack package
        run: npm pack

      - name: Upload build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: package-build
          path: |
            dist/
            *.tgz

  publish-npm:
    needs: [test, build]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          registry-url: 'https://registry.npmjs.org'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build package
        run: npm run build

      - name: Determine release type
        id: release-type
        run: |
          if [[ "${{ github.ref }}" =~ -alpha\.|beta\.|rc\. ]]; then
            echo "tag=beta" >> $GITHUB_OUTPUT
          else
            echo "tag=latest" >> $GITHUB_OUTPUT
          fi

      - name: Publish to NPM
        run: |
          if [ "${{ steps.release-type.outputs.tag }}" = "beta" ]; then
            npm publish --tag beta
          else
            npm publish
          fi
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}

  publish-github:
    needs: [test, build]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js for GitHub Packages
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          registry-url: 'https://npm.pkg.github.com'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build package
        run: npm run build

      - name: Publish to GitHub Packages
        run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Advanced Publishing Strategies
```json
{
  "publishConfig": {
    "registry": "https://registry.npmjs.org/",
    "access": "public",
    "tag": "latest"
  },
  "scripts": {
    "release": "npm run build && npm run test && npm publish",
    "release:alpha": "npm version prerelease --preid=alpha && npm run release -- --tag alpha",
    "release:beta": "npm version prerelease --preid=beta && npm run release -- --tag beta",
    "release:rc": "npm version prerelease --preid=rc && npm run release -- --tag rc",
    "release:patch": "npm version patch && npm run release",
    "release:minor": "npm version minor && npm run release",
    "release:major": "npm version major && npm run release"
  }
}
```

## Package Quality and Distribution

### Essential Files
```bash
# Package structure
my-package/
├── src/                 # Source code
├── dist/               # Built output
├── docs/               # Documentation
├── examples/           # Usage examples
├── __tests__/          # Tests
├── package.json        # Package configuration
├── README.md           # Package description
├── LICENSE             # License file
├── CHANGELOG.md        # Version history
└── .npmignore          # Files to exclude from npm
```

### .npmignore Configuration
```gitignore
# Source and development files
src/
__tests__/
*.test.ts
*.spec.ts
*.test.js
*.spec.js

# Build configuration
tsconfig.json
rollup.config.js
webpack.config.js
.babelrc
jest.config.js

# Development tools
.eslintrc*
.prettierrc*
.editorconfig
.gitignore
.github/

# Documentation source
docs-src/
examples/

# Logs and caches
*.log
node_modules/
coverage/
.nyc_output/
```

### Security and Package Integrity

#### Package Security Best Practices
```bash
# Security Auditing
npm audit                    # Check for vulnerabilities
npm audit fix                # Auto-fix vulnerabilities
npm audit fix --force        # Force fixes (may cause breaking changes)
npm audit --audit-level high # Only show high/critical vulnerabilities

# Dependency Management
npm ls                       # List installed packages
npm ls --depth=0            # Only direct dependencies
npm outdated                # Check for outdated packages
npm update                  # Update packages within semver range

# Package Verification
npm pack --dry-run          # Preview package contents
npm view my-package         # View package info from registry
npm doctor                  # Run environment diagnostics
```

#### Package Lock and Reproducible Builds
```json
{
  "scripts": {
    "preinstall": "npx only-allow npm",
    "postinstall": "npm audit",
    "build:clean": "rimraf dist && npm ci --only=production && npm run build"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=8.0.0"
  },
  "engineStrict": true
}
```

### Monorepo Packaging with Workspaces

#### NPM Workspaces Configuration
```json
{
  "name": "my-monorepo",
  "workspaces": [
    "packages/*",
    "apps/*",
    "tools/*"
  ],
  "scripts": {
    "build": "npm run build --workspaces",
    "test": "npm run test --workspaces",
    "publish:all": "npm run build && npm publish --workspaces"
  }
}
```

#### Lerna Configuration
```json
{
  "version": "independent",
  "npmClient": "npm",
  "command": {
    "publish": {
      "conventionalCommits": true,
      "message": "chore(release): publish",
      "npmClientArgs": ["--no-git-reset"]
    },
    "version": {
      "allowBranch": ["master", "main"],
      "conventionalCommits": true
    }
  },
  "ignoreChanges": [
    "**/*.md",
    "**/*.test.ts",
    "**/*.spec.ts"
  ]
}
```

### Modern Package Distribution

#### CDN Distribution
```html
<!-- ESM from CDN -->
<script type="module">
  import { myFunction } from 'https://cdn.skypack.dev/my-package';
  import { myFunction } from 'https://esm.sh/my-package';
  import { myFunction } from 'https://unpkg.com/my-package?module';
</script>

<!-- UMD from CDN -->
<script src="https://unpkg.com/my-package"></script>
<script src="https://cdn.jsdelivr.net/npm/my-package"></script>
```

#### Package.json for CDN Optimization
```json
{
  "unpkg": "dist/index.browser.js",
  "jsdelivr": "dist/index.browser.js",
  "browser": {
    "./dist/index.js": "./dist/index.browser.js",
    "./dist/index.cjs": "./dist/index.browser.js"
  },
  "sideEffects": false,
  "type": "module"
}
```

### Package Maintenance and Lifecycle

#### Deprecation Strategy
```bash
# Deprecate specific version
npm deprecate my-package@1.0.0 "This version has security vulnerabilities"

# Deprecate all versions
npm deprecate my-package "Package no longer maintained"

# Deprecate version range
npm deprecate my-package@"<2.0.0" "Please upgrade to v2.x for security fixes"
```

#### Package Migration Guide
```markdown
## Migration Guide Template

### Breaking Changes in v2.0.0

#### Removed APIs
- `oldFunction()` → Use `newFunction()` instead
- `deprecatedClass` → Use `ModernClass` instead

#### Changed APIs
- `configure(options)` now requires `apiKey` parameter
- `process()` returns Promise instead of synchronous result

#### New Requirements
- Node.js 18+ required (was 16+)
- ESM-only package (CommonJS via compatibility layer)

### Migration Steps
1. Update dependencies: `npm install my-package@^2.0.0`
2. Replace deprecated APIs (see above)
3. Update import statements if using CommonJS
4. Run tests and update as needed
```

### Package Analytics and Monitoring

#### Package Health Metrics
```json
{
  "scripts": {
    "stats": "npm view my-package",
    "downloads": "npm view my-package --json | jq '.downloads'",
    "size": "bundlesize",
    "analyze": "webpack-bundle-analyzer dist/stats.json"
  },
  "bundlesize": [
    {
      "path": "./dist/*.js",
      "maxSize": "50kb"
    }
  ]
}
```

This comprehensive packaging guide covers the complete lifecycle of TypeScript/JavaScript package development, from initial setup through maintenance and deprecation, ensuring professional and sustainable package distribution.