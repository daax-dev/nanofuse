# Local Build Toolchain for TypeScript/JavaScript

## Overview

This guide covers setting up a comprehensive local development environment and build toolchain for TypeScript/JavaScript projects, from basic setups to advanced configurations.

## Node.js and Package Management

### Node.js Setup
```bash
# Install Node.js using Node Version Manager (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
# or with wget
wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# Restart terminal or source profile
source ~/.bashrc

# Install latest LTS Node.js
nvm install --lts
nvm use --lts
nvm alias default node

# Verify installation
node --version
npm --version
```

### Package Managers Comparison

#### npm (Default)
```bash
# Initialize project
npm init -y

# Install dependencies
npm install express
npm install --save-dev typescript @types/express

# Global installations
npm install -g typescript ts-node

# Scripts execution
npm run build
npm test

# Lock file: package-lock.json
```

#### Yarn (Alternative)
```bash
# Install Yarn
npm install -g yarn

# Initialize project
yarn init -y

# Install dependencies
yarn add express
yarn add -D typescript @types/express

# Global installations
yarn global add typescript ts-node

# Scripts execution
yarn build
yarn test

# Lock file: yarn.lock
```

#### pnpm (Fast, Efficient)
```bash
# Install pnpm
npm install -g pnpm

# Initialize project
pnpm init

# Install dependencies
pnpm add express
pnpm add -D typescript @types/express

# Global installations
pnpm add -g typescript ts-node

# Scripts execution
pnpm build
pnpm test

# Lock file: pnpm-lock.yaml
# Advantages: Faster, disk efficient, strict node_modules
```

## TypeScript Configuration

### Basic tsconfig.json
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "node",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "allowJs": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "removeComments": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": [
    "src/**/*"
  ],
  "exclude": [
    "node_modules",
    "dist",
    "**/*.test.ts",
    "**/*.spec.ts"
  ]
}
```

### Advanced TypeScript Configuration
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "node",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    
    // Input/Output
    "allowJs": true,
    "checkJs": false,
    "outDir": "./dist",
    "rootDir": "./src",
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"],
      "@utils/*": ["src/utils/*"]
    },
    
    // Type checking
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "strictBindCallApply": true,
    "strictPropertyInitialization": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    
    // Modules
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    
    // Emit
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "removeComments": false,
    "importHelpers": true,
    "downlevelIteration": true,
    
    // Advanced
    "skipLibCheck": true,
    "experimentalDecorators": true,
    "emitDecoratorMetadata": true
  },
  "include": [
    "src/**/*",
    "types/**/*"
  ],
  "exclude": [
    "node_modules",
    "dist",
    "build",
    "**/*.test.ts",
    "**/*.spec.ts"
  ],
  "ts-node": {
    "esm": true,
    "compilerOptions": {
      "module": "ESNext"
    }
  }
}
```

### Environment-Specific Configurations
```bash
# Multiple tsconfig files for different environments

# tsconfig.json (base)
# tsconfig.build.json (production build)
# tsconfig.test.json (testing)
# tsconfig.dev.json (development)
```

```json
// tsconfig.build.json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "outDir": "./dist",
    "removeComments": true,
    "sourceMap": false
  },
  "exclude": [
    "**/*.test.ts",
    "**/*.spec.ts",
    "src/**/__tests__/**"
  ]
}

// tsconfig.test.json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "types": ["node", "jest"],
    "esModuleInterop": true
  },
  "include": [
    "src/**/*",
    "**/*.test.ts",
    "**/*.spec.ts"
  ]
}
```

## Build Tools and Bundlers

### Webpack Configuration
```javascript
// webpack.config.js
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');

module.exports = {
  entry: './src/index.ts',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].[contenthash].js',
    clean: true
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@components': path.resolve(__dirname, 'src/components'),
      '@utils': path.resolve(__dirname, 'src/utils')
    }
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: [
          {
            loader: 'ts-loader',
            options: {
              configFile: 'tsconfig.build.json'
            }
          }
        ],
        exclude: /node_modules/
      },
      {
        test: /\.css$/i,
        use: ['style-loader', 'css-loader', 'postcss-loader']
      },
      {
        test: /\.(png|svg|jpg|jpeg|gif)$/i,
        type: 'asset/resource'
      }
    ]
  },
  plugins: [
    new CleanWebpackPlugin(),
    new HtmlWebpackPlugin({
      template: './src/index.html',
      minify: {
        removeComments: true,
        collapseWhitespace: true,
        removeRedundantAttributes: true
      }
    })
  ],
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all'
        }
      }
    }
  },
  devServer: {
    static: path.join(__dirname, 'dist'),
    port: 3000,
    hot: true,
    open: true
  }
};
```

### Vite Configuration (Modern Alternative)
```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    target: 'es2022',
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        nested: resolve(__dirname, 'nested/index.html')
      }
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@components': resolve(__dirname, 'src/components'),
      '@utils': resolve(__dirname, 'src/utils')
    }
  },
  server: {
    port: 3000,
    open: true,
    cors: true
  },
  plugins: [
    // Add plugins as needed
  ]
});
```

### Rollup Configuration (Library Building)
```javascript
// rollup.config.js
import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from 'rollup-plugin-terser';

export default {
  input: 'src/index.ts',
  output: [
    {
      file: 'dist/index.cjs.js',
      format: 'cjs',
      sourcemap: true
    },
    {
      file: 'dist/index.esm.js',
      format: 'esm',
      sourcemap: true
    },
    {
      file: 'dist/index.umd.js',
      format: 'umd',
      name: 'MyLibrary',
      sourcemap: true
    }
  ],
  plugins: [
    nodeResolve(),
    commonjs(),
    typescript({
      tsconfig: './tsconfig.build.json',
      declaration: true,
      declarationDir: './dist/types'
    }),
    terser()
  ],
  external: ['lodash', 'react', 'react-dom']
};
```

## Development Tools

### Essential Development Dependencies
```json
{
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^18.0.0",
    "ts-node": "^10.9.0",
    "nodemon": "^2.0.20",
    "concurrently": "^7.6.0",
    "cross-env": "^7.0.3",
    "rimraf": "^4.0.0"
  }
}
```

### ts-node Configuration
```json
// tsconfig.json
{
  "ts-node": {
    "esm": true,
    "experimentalSpecifierResolution": "node",
    "compilerOptions": {
      "module": "ESNext",
      "target": "ES2022",
      "moduleResolution": "node"
    }
  }
}
```

### Nodemon Configuration
```json
// nodemon.json
{
  "watch": ["src"],
  "ext": "ts,js,json",
  "ignore": [
    "src/**/*.spec.ts",
    "src/**/*.test.ts"
  ],
  "exec": "ts-node --esm src/index.ts",
  "env": {
    "NODE_ENV": "development"
  },
  "delay": "1000"
}
```

## Code Quality Tools

### ESLint Configuration
```json
// .eslintrc.json
{
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 2022,
    "sourceType": "module",
    "project": "./tsconfig.json"
  },
  "plugins": [
    "@typescript-eslint",
    "import",
    "promise"
  ],
  "extends": [
    "eslint:recommended",
    "@typescript-eslint/recommended",
    "@typescript-eslint/recommended-requiring-type-checking",
    "plugin:import/recommended",
    "plugin:import/typescript",
    "plugin:promise/recommended"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/explicit-function-return-type": "warn",
    "@typescript-eslint/no-explicit-any": "warn",
    "@typescript-eslint/prefer-const": "error",
    "import/order": [
      "error",
      {
        "groups": [
          "builtin",
          "external",
          "internal",
          "parent",
          "sibling",
          "index"
        ],
        "newlines-between": "always"
      }
    ]
  },
  "settings": {
    "import/resolver": {
      "typescript": {
        "alwaysTryTypes": true,
        "project": "./tsconfig.json"
      }
    }
  }
}
```

### Prettier Configuration
```json
// .prettierrc
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false,
  "bracketSpacing": true,
  "arrowParens": "avoid"
}
```

```javascript
// .prettierignore
dist/
node_modules/
coverage/
*.min.js
package-lock.json
yarn.lock
pnpm-lock.yaml
```

### Husky and lint-staged
```json
// package.json
{
  "husky": {
    "hooks": {
      "pre-commit": "lint-staged",
      "commit-msg": "commitlint -E HUSKY_GIT_PARAMS"
    }
  },
  "lint-staged": {
    "*.{ts,tsx,js,jsx}": [
      "eslint --fix",
      "prettier --write",
      "git add"
    ],
    "*.{json,md,yml,yaml}": [
      "prettier --write",
      "git add"
    ]
  }
}
```

## Testing Setup

### Jest Configuration
```json
// jest.config.js
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src'],
  testMatch: [
    '**/__tests__/**/*.ts',
    '**/?(*.)+(spec|test).ts'
  ],
  transform: {
    '^.+\.ts$': 'ts-jest'
  },
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.d.ts',
    '!src/**/__tests__/**',
    '!src/**/index.ts'
  ],
  coverageDirectory: 'coverage',
  coverageReporters: ['text', 'lcov', 'html'],
  setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@utils/(.*)$': '<rootDir>/src/utils/$1'
  }
};
```

### Vitest Configuration (Modern Alternative)
```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import { resolve } from 'path';

export default defineConfig({
  test: {
    environment: 'node',
    globals: true,
    coverage: {
      provider: 'c8',
      reporter: ['text', 'json', 'html']
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@components': resolve(__dirname, 'src/components'),
      '@utils': resolve(__dirname, 'src/utils')
    }
  }
});
```

## Package.json Scripts

### Comprehensive Scripts Setup
```json
{
  "scripts": {
    "dev": "nodemon",
    "dev:watch": "concurrently \"tsc --watch\" \"nodemon dist/index.js\"",
    "build": "rimraf dist && tsc -p tsconfig.build.json",
    "build:webpack": "webpack --mode production",
    "build:vite": "vite build",
    "start": "node dist/index.js",
    "start:dev": "ts-node --esm src/index.ts",
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:vitest": "vitest",
    "lint": "eslint src/**/*.ts",
    "lint:fix": "eslint src/**/*.ts --fix",
    "format": "prettier --write src/**/*.ts",
    "format:check": "prettier --check src/**/*.ts",
    "typecheck": "tsc --noEmit",
    "clean": "rimraf dist coverage node_modules/.cache",
    "prepublishOnly": "npm run clean && npm run build && npm test",
    "precommit": "lint-staged",
    "release": "standard-version",
    "release:dry": "standard-version --dry-run"
  }
}
```

## Environment Configuration

### Environment Variables
```bash
# .env
NODE_ENV=development
PORT=3000
DATABASE_URL=postgresql://localhost:5432/mydb
API_KEY=your-api-key-here
JWT_SECRET=your-jwt-secret
```

```typescript
// src/config/env.ts
import dotenv from 'dotenv';

dotenv.config();

export const env = {
  NODE_ENV: process.env.NODE_ENV || 'development',
  PORT: parseInt(process.env.PORT || '3000', 10),
  DATABASE_URL: process.env.DATABASE_URL!,
  API_KEY: process.env.API_KEY!,
  JWT_SECRET: process.env.JWT_SECRET!,
} as const;

// Type-safe environment validation
function validateEnv(): void {
  const required = ['DATABASE_URL', 'API_KEY', 'JWT_SECRET'] as const;
  
  for (const key of required) {
    if (!process.env[key]) {
      throw new Error(`Missing required environment variable: ${key}`);
    }
  }
}

validateEnv();
```

### Cross-platform Scripts
```json
{
  "scripts": {
    "clean": "cross-env rimraf dist",
    "build": "cross-env NODE_ENV=production tsc",
    "dev": "cross-env NODE_ENV=development nodemon",
    "test": "cross-env NODE_ENV=test jest"
  }
}
```

## Advanced Development Setup

### Multi-package Monorepo (Lerna/Nx)
```json
// lerna.json
{
  "version": "independent",
  "npmClient": "npm",
  "packages": [
    "packages/*"
  ],
  "command": {
    "publish": {
      "conventionalCommits": true,
      "registry": "https://registry.npmjs.org/"
    },
    "bootstrap": {
      "ignore": "component-*",
      "npmClientArgs": ["--no-package-lock"]
    }
  }
}
```

### Workspace Configuration
```json
// package.json (root)
{
  "name": "my-monorepo",
  "private": true,
  "workspaces": [
    "packages/*",
    "apps/*"
  ],
  "scripts": {
    "build:all": "lerna run build",
    "test:all": "lerna run test",
    "lint:all": "lerna run lint",
    "clean:all": "lerna run clean"
  },
  "devDependencies": {
    "lerna": "^6.0.0",
    "typescript": "^5.0.0"
  }
}
```

### Docker Development Environment
```dockerfile
# Dockerfile.dev
FROM node:18-alpine

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci --only=development

# Copy source
COPY . .

# Expose port
EXPOSE 3000

# Development command
CMD ["npm", "run", "dev"]
```

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
    depends_on:
      - database
  
  database:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### VS Code Integration
```json
// .vscode/settings.json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true,
    "source.organizeImports": true
  },
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "typescript.suggest.autoImports": true,
  "typescript.updateImportsOnFileMove.enabled": "always",
  "files.exclude": {
    "**/node_modules": true,
    "**/dist": true,
    "**/*.js.map": true
  },
  "search.exclude": {
    "**/node_modules": true,
    "**/dist": true
  }
}

// .vscode/extensions.json
{
  "recommendations": [
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-json"
  ]
}

// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "tsc: build - tsconfig.json",
      "type": "typescript",
      "tsconfig": "tsconfig.json",
      "option": "watch",
      "problemMatcher": ["$tsc-watch"],
      "group": "build"
    },
    {
      "label": "npm: dev",
      "type": "shell",
      "command": "npm run dev",
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    }
  ]
}
```

## Performance Optimization

### Build Performance
```typescript
// webpack.config.js - Performance optimizations
module.exports = {
  // ... other config
  cache: {
    type: 'filesystem',
    buildDependencies: {
      config: [__filename]
    }
  },
  optimization: {
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all'
        }
      }
    }
  },
  resolve: {
    // Faster module resolution
    modules: ['node_modules'],
    extensions: ['.ts', '.tsx', '.js', '.jsx']
  }
};
```

### TypeScript Performance
```json
// tsconfig.json - Performance optimizations
{
  "compilerOptions": {
    "skipLibCheck": true,
    "incremental": true,
    "tsBuildInfoFile": ".tsbuildinfo"
  },
  "exclude": [
    "node_modules",
    "**/*.spec.ts",
    "**/*.test.ts"
  ]
}
```

This comprehensive local build toolchain setup provides everything needed for efficient TypeScript/JavaScript development, from basic configurations to advanced enterprise-level setups.