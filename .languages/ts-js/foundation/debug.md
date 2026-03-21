# Debugging TypeScript/JavaScript Applications

## Browser DevTools

### Console Debugging
```javascript
// Basic logging
console.log('Debug message');
console.error('Error message');
console.warn('Warning message');
console.info('Info message');

// Structured logging
console.table([{name: 'John', age: 30}, {name: 'Jane', age: 25}]);
console.group('User Details');
console.log('Name: John');
console.log('Age: 30');
console.groupEnd();

// Performance tracking
console.time('operation');
// ... some operation
console.timeEnd('operation');

// Stack traces
console.trace('Execution path');
```

### Breakpoint Debugging
```javascript
// Programmatic breakpoints
debugger; // Pauses execution when DevTools are open

// Conditional breakpoints in DevTools
function processUser(user) {
  // Right-click line number in DevTools, add conditional breakpoint
  // Condition: user.age > 25
  return user.name.toUpperCase();
}
```

### Advanced DevTools Features
```javascript
// Memory profiling
const users = [];
for (let i = 0; i < 10000; i++) {
  users.push({ id: i, name: `User ${i}` });
}

// Performance profiling
performance.mark('start-processing');
// ... processing code
performance.mark('end-processing');
performance.measure('processing-time', 'start-processing', 'end-processing');
```

## Node.js Debugging

### Built-in Debugger
```bash
# Start Node.js with debugger
node --inspect server.js
node --inspect-brk server.js  # Break on first line

# Debug with Chrome DevTools
# Open chrome://inspect in Chrome browser
```

### VS Code Debugging
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug TypeScript",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/src/index.ts",
      "outFiles": ["${workspaceFolder}/dist/**/*.js"],
      "sourceMaps": true,
      "runtimeArgs": ["-r", "ts-node/register"],
      "env": {
        "NODE_ENV": "development"
      }
    },
    {
      "name": "Debug Tests",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/node_modules/.bin/jest",
      "args": ["--runInBand", "--no-cache"],
      "console": "integratedTerminal",
      "internalConsoleOptions": "neverOpen"
    }
  ]
}
```

## TypeScript-Specific Debugging

### Source Maps Configuration
```json
// tsconfig.json
{
  "compilerOptions": {
    "sourceMap": true,
    "inlineSourceMap": false,
    "sourceRoot": "/",
    "mapRoot": "/"
  }
}
```

### Debugging Compiled Code
```javascript
// Original TypeScript
class UserService {
  private users: User[] = [];

  addUser(user: User): void {
    this.users.push(user);
  }

  getUser(id: string): User | undefined {
    debugger; // This will work in both TS and compiled JS
    return this.users.find(u => u.id === id);
  }
}
```

## Testing and Debugging

### Jest Debugging
```json
// package.json
{
  "scripts": {
    "test:debug": "node --inspect-brk node_modules/.bin/jest --runInBand",
    "test:watch": "jest --watch"
  }
}
```

```javascript
// Debug specific test
describe('UserService', () => {
  it('should find user by id', () => {
    const service = new UserService();
    const user = { id: '1', name: 'John' };

    service.addUser(user);
    debugger; // Breakpoint for debugging

    const found = service.getUser('1');
    expect(found).toBe(user);
  });
});
```

## Error Tracking and Logging

### Structured Error Handling
```typescript
interface ErrorContext {
  userId?: string;
  operation: string;
  timestamp: Date;
  metadata?: Record<string, any>;
}

class ApplicationError extends Error {
  constructor(
    message: string,
    public context: ErrorContext,
    public originalError?: Error
  ) {
    super(message);
    this.name = 'ApplicationError';
  }
}

// Usage
try {
  await processUser(userId);
} catch (error) {
  throw new ApplicationError(
    'Failed to process user',
    {
      userId,
      operation: 'processUser',
      timestamp: new Date(),
      metadata: { additionalInfo: 'some context' }
    },
    error
  );
}
```

### Production Debugging Tools
```javascript
// Winston logging setup
import winston from 'winston';

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.errors({ stack: true }),
    winston.format.json()
  ),
  defaultMeta: { service: 'user-service' },
  transports: [
    new winston.transports.File({ filename: 'error.log', level: 'error' }),
    new winston.transports.File({ filename: 'combined.log' }),
    new winston.transports.Console({
      format: winston.format.simple()
    })
  ]
});

// Usage
logger.info('User created successfully', { userId: '123' });
logger.error('Database connection failed', { error: err });
```

This debugging guide provides comprehensive techniques for troubleshooting TypeScript/JavaScript applications across different environments and scenarios.