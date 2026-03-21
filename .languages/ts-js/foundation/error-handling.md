# Error Handling in TypeScript/JavaScript

## Table of Contents
- [Error Types and Categories](#error-types-and-categories)
- [Core Error Handling Mechanisms](#core-error-handling-mechanisms)
- [Advanced Error Handling Patterns](#advanced-error-handling-patterns)
- [Async Error Handling](#async-error-handling)
- [TypeScript Error Types](#typescript-error-types)
- [Testing Error Scenarios](#testing-error-scenarios)
- [Production Error Handling](#production-error-handling)

## Error Types and Categories

### Built-in Error Types
```javascript
// Generic Error
throw new Error('Something went wrong');

// Type-specific errors
throw new TypeError('Expected a string, got number');
throw new ReferenceError('Variable is not defined');
throw new RangeError('Number out of range');
throw new SyntaxError('Invalid syntax');
throw new URIError('Invalid URI');

// Specialized errors
throw new EvalError('Error in eval()'); // Rarely used
throw new AggregateError(errors, 'Multiple errors occurred');
```

### Custom Error Classes
```javascript
// Basic custom error
class CustomError extends Error {
  constructor(message) {
    super(message);
    this.name = 'CustomError';

    // Maintains proper stack trace (V8)
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, CustomError);
    }
  }
}

// Rich custom error with additional properties
class ValidationError extends Error {
  constructor(message, field, value) {
    super(message);
    this.name = 'ValidationError';
    this.field = field;
    this.value = value;
    this.timestamp = new Date().toISOString();

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, ValidationError);
    }
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      field: this.field,
      value: this.value,
      timestamp: this.timestamp,
      stack: this.stack
    };
  }
}

// Domain-specific errors
class NetworkError extends Error {
  constructor(message, statusCode, url) {
    super(message);
    this.name = 'NetworkError';
    this.statusCode = statusCode;
    this.url = url;
  }
}

class DatabaseError extends Error {
  constructor(message, query, code) {
    super(message);
    this.name = 'DatabaseError';
    this.query = query;
    this.code = code;
  }
}
```

## Core Error Handling Mechanisms

### Try-Catch-Finally
```javascript
// Basic try-catch
try {
  const result = riskyOperation();
  return result;
} catch (error) {
  console.error('Operation failed:', error.message);
  return null;
}

// With finally block
try {
  const resource = acquireResource();
  return processResource(resource);
} catch (error) {
  logError(error);
  throw error; // Re-throw after logging
} finally {
  // Always executes, even if return or throw in try/catch
  releaseResource(resource);
}

// Multiple catch blocks simulation
try {
  riskyOperation();
} catch (error) {
  if (error instanceof TypeError) {
    handleTypeError(error);
  } else if (error instanceof NetworkError) {
    handleNetworkError(error);
  } else {
    handleGenericError(error);
  }
}

// Nested try-catch for granular handling
try {
  const data = parseJSON(input);
  try {
    const result = processData(data);
    return result;
  } catch (processingError) {
    logError('Processing failed', processingError);
    return getDefaultResult();
  }
} catch (parseError) {
  logError('Parsing failed', parseError);
  throw new ValidationError('Invalid input format');
}
```

### Error Propagation Strategies
```javascript
// Immediate re-throw
function processUser(userData) {
  try {
    return validateAndSave(userData);
  } catch (error) {
    // Add context and re-throw
    error.context = 'processUser';
    error.userData = sanitizeForLogging(userData);
    throw error;
  }
}

// Transform and re-throw
function apiCall(endpoint) {
  try {
    return fetch(endpoint);
  } catch (error) {
    // Transform internal error to domain error
    throw new NetworkError(
      `Failed to fetch ${endpoint}`,
      error.status,
      endpoint
    );
  }
}

// Catch and return error as value
function safeOperation(input) {
  try {
    return { success: true, data: riskyOperation(input) };
  } catch (error) {
    return { success: false, error: error.message };
  }
}
```

## Advanced Error Handling Patterns

### Result Pattern (Either Type)
```javascript
// Result type implementation
class Result {
  constructor(isSuccess, value, error) {
    this.isSuccess = isSuccess;
    this.value = value;
    this.error = error;
  }

  static success(value) {
    return new Result(true, value, null);
  }

  static failure(error) {
    return new Result(false, null, error);
  }

  map(fn) {
    if (this.isSuccess) {
      try {
        return Result.success(fn(this.value));
      } catch (error) {
        return Result.failure(error);
      }
    }
    return this;
  }

  flatMap(fn) {
    if (this.isSuccess) {
      try {
        return fn(this.value);
      } catch (error) {
        return Result.failure(error);
      }
    }
    return this;
  }

  getOrElse(defaultValue) {
    return this.isSuccess ? this.value : defaultValue;
  }
}

// Usage
function divideResult(a, b) {
  if (b === 0) {
    return Result.failure(new Error('Division by zero'));
  }
  return Result.success(a / b);
}

const result = divideResult(10, 2)
  .map(x => x * 2)
  .map(x => x + 1);

if (result.isSuccess) {
  console.log('Result:', result.value);
} else {
  console.error('Error:', result.error.message);
}
```

### Option Pattern (Maybe Type)
```javascript
class Option {
  constructor(value) {
    this.value = value;
  }

  static some(value) {
    return new Option(value);
  }

  static none() {
    return new Option(null);
  }

  static of(value) {
    return value != null ? Option.some(value) : Option.none();
  }

  isSome() {
    return this.value != null;
  }

  isNone() {
    return this.value == null;
  }

  map(fn) {
    return this.isSome() ? Option.of(fn(this.value)) : this;
  }

  flatMap(fn) {
    return this.isSome() ? fn(this.value) : this;
  }

  filter(predicate) {
    return this.isSome() && predicate(this.value) ? this : Option.none();
  }

  getOrElse(defaultValue) {
    return this.isSome() ? this.value : defaultValue;
  }
}

// Usage
function findUser(id) {
  const user = database.findById(id);
  return Option.of(user);
}

const userName = findUser(123)
  .map(user => user.name)
  .filter(name => name.length > 0)
  .getOrElse('Unknown User');
```

### Circuit Breaker Pattern
```javascript
class CircuitBreaker {
  constructor(options = {}) {
    this.failureThreshold = options.failureThreshold || 5;
    this.timeout = options.timeout || 60000;
    this.monitoringPeriod = options.monitoringPeriod || 10000;

    this.failureCount = 0;
    this.lastFailureTime = null;
    this.state = 'CLOSED'; // CLOSED, OPEN, HALF_OPEN
  }

  async call(fn, ...args) {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime > this.timeout) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker is OPEN');
      }
    }

    try {
      const result = await fn(...args);
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failureCount = 0;
    this.state = 'CLOSED';
  }

  onFailure() {
    this.failureCount++;
    this.lastFailureTime = Date.now();

    if (this.failureCount >= this.failureThreshold) {
      this.state = 'OPEN';
    }
  }
}

// Usage
const breaker = new CircuitBreaker({
  failureThreshold: 3,
  timeout: 30000
});

async function callExternalAPI() {
  return breaker.call(fetch, '/api/external');
}
```

### Retry Pattern
```javascript
async function retry(fn, options = {}) {
  const {
    maxAttempts = 3,
    delay = 1000,
    backoffFactor = 2,
    shouldRetry = () => true
  } = options;

  let lastError;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      if (attempt === maxAttempts || !shouldRetry(error)) {
        break;
      }

      const waitTime = delay * Math.pow(backoffFactor, attempt - 1);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }
  }

  throw lastError;
}

// Usage with exponential backoff
async function fetchWithRetry(url) {
  return retry(
    () => fetch(url),
    {
      maxAttempts: 3,
      delay: 1000,
      shouldRetry: (error) => error.status >= 500
    }
  );
}
```

## Async Error Handling

### Promise Error Handling
```javascript
// Promise chain error handling
fetchUser(userId)
  .then(user => processUser(user))
  .then(result => saveResult(result))
  .catch(error => {
    console.error('Pipeline failed:', error);
    return getDefaultResult();
  })
  .finally(() => {
    cleanup();
  });

// Multiple promise error handling
Promise.allSettled([
  fetchUser(1),
  fetchUser(2),
  fetchUser(3)
]).then(results => {
  const successful = results.filter(r => r.status === 'fulfilled');
  const failed = results.filter(r => r.status === 'rejected');

  console.log(`${successful.length} succeeded, ${failed.length} failed`);

  failed.forEach(failure => {
    console.error('Failed request:', failure.reason);
  });
});

// Promise.any for fallback operations
Promise.any([
  fetchFromPrimaryAPI(),
  fetchFromSecondaryAPI(),
  fetchFromCache()
]).then(result => {
  console.log('Got result from fastest source:', result);
}).catch(aggregateError => {
  console.error('All sources failed:', aggregateError.errors);
});
```

### Async/Await Error Handling
```javascript
// Basic async/await error handling
async function processData(id) {
  try {
    const data = await fetchData(id);
    const processed = await processRawData(data);
    const result = await saveProcessedData(processed);
    return result;
  } catch (error) {
    console.error('Processing failed:', error);
    throw new ProcessingError('Failed to process data', id);
  }
}

// Parallel operations with error handling
async function processMultipleItems(ids) {
  const results = await Promise.allSettled(
    ids.map(id => processData(id))
  );

  const successful = [];
  const errors = [];

  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      successful.push(result.value);
    } else {
      errors.push({
        id: ids[index],
        error: result.reason
      });
    }
  });

  return { successful, errors };
}

// Graceful degradation
async function getUser(id) {
  try {
    return await fetchUserFromAPI(id);
  } catch (apiError) {
    console.warn('API failed, trying cache:', apiError.message);

    try {
      return await getUserFromCache(id);
    } catch (cacheError) {
      console.warn('Cache failed, using default:', cacheError.message);
      return getDefaultUser();
    }
  }
}
```

### Event-Driven Error Handling
```javascript
// EventEmitter error handling
const EventEmitter = require('events');

class DataProcessor extends EventEmitter {
  async processData(data) {
    this.emit('processing:start', data);

    try {
      const result = await this.doProcessing(data);
      this.emit('processing:success', result);
      return result;
    } catch (error) {
      this.emit('processing:error', error, data);
      throw error;
    }
  }
}

const processor = new DataProcessor();

processor.on('processing:error', (error, data) => {
  console.error('Processing failed:', error);
  // Log to monitoring system
  logger.error('Data processing error', {
    error: error.message,
    data: sanitize(data),
    stack: error.stack
  });
});

// Unhandled rejection handling
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  // Log the error and gracefully shutdown
});

process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
  // Log and exit
  process.exit(1);
});
```

## TypeScript Error Types

### Strongly Typed Error Handling
```typescript
// Error union types
type ApiError =
  | { type: 'network'; message: string; status: number }
  | { type: 'validation'; message: string; field: string }
  | { type: 'unauthorized'; message: string }
  | { type: 'server'; message: string; code: string };

// Result type with specific error types
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

// Function with typed errors
function parseUserData(input: string): Result<User, ApiError> {
  try {
    const data = JSON.parse(input);

    if (!isValidUser(data)) {
      return {
        success: false,
        error: {
          type: 'validation',
          message: 'Invalid user data',
          field: 'user'
        }
      };
    }

    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: {
        type: 'validation',
        message: 'Invalid JSON',
        field: 'input'
      }
    };
  }
}

// Usage with type narrowing
const result = parseUserData(jsonString);

if (result.success) {
  // TypeScript knows result.data is User
  console.log(result.data.name);
} else {
  // TypeScript knows result.error is ApiError
  switch (result.error.type) {
    case 'validation':
      console.error(`Validation error in ${result.error.field}: ${result.error.message}`);
      break;
    case 'network':
      console.error(`Network error ${result.error.status}: ${result.error.message}`);
      break;
    default:
      console.error('Unknown error:', result.error.message);
  }
}
```

### Exception Specifications (TypeScript)
```typescript
// Tagged union for exceptions
type Exception<T extends string, P = {}> = {
  readonly _tag: T;
} & P;

type ValidationException = Exception<'ValidationException', {
  field: string;
  value: unknown;
}>;

type NetworkException = Exception<'NetworkException', {
  statusCode: number;
  url: string;
}>;

type AppException = ValidationException | NetworkException;

// Function that can throw specific exceptions
function validateUser(data: unknown): User | never {
  if (typeof data !== 'object' || data === null) {
    const error: ValidationException = {
      _tag: 'ValidationException',
      field: 'root',
      value: data
    };
    throw error;
  }

  // Additional validation...
  return data as User;
}

// Catch and handle typed exceptions
try {
  const user = validateUser(input);
  return user;
} catch (error) {
  if (isValidationException(error)) {
    handleValidationError(error);
  } else if (isNetworkException(error)) {
    handleNetworkError(error);
  } else {
    handleUnknownError(error);
  }
}

// Type guards for exceptions
function isValidationException(error: any): error is ValidationException {
  return error && error._tag === 'ValidationException';
}

function isNetworkException(error: any): error is NetworkException {
  return error && error._tag === 'NetworkException';
}
```

## Testing Error Scenarios

### Unit Testing Error Conditions
```javascript
// Jest error testing
describe('UserService', () => {
  describe('validateUser', () => {
    it('should throw ValidationError for missing name', () => {
      const invalidUser = { email: 'test@example.com' };

      expect(() => {
        userService.validateUser(invalidUser);
      }).toThrow(ValidationError);

      expect(() => {
        userService.validateUser(invalidUser);
      }).toThrow('Name is required');
    });

    it('should handle async errors', async () => {
      const mockError = new NetworkError('Connection failed', 500, '/api/users');
      jest.spyOn(apiClient, 'fetchUser').mockRejectedValue(mockError);

      await expect(userService.getUser(123)).rejects.toThrow(NetworkError);
      await expect(userService.getUser(123)).rejects.toThrow('Connection failed');
    });

    it('should test error recovery', async () => {
      // Mock primary API to fail
      jest.spyOn(primaryApi, 'getUser').mockRejectedValue(new Error('API down'));
      // Mock fallback to succeed
      jest.spyOn(fallbackApi, 'getUser').mockResolvedValue({ id: 123, name: 'John' });

      const result = await userService.getUserWithFallback(123);

      expect(result).toEqual({ id: 123, name: 'John' });
      expect(primaryApi.getUser).toHaveBeenCalled();
      expect(fallbackApi.getUser).toHaveBeenCalled();
    });
  });
});

// Testing custom error properties
describe('CustomError', () => {
  it('should preserve error properties', () => {
    const error = new ValidationError('Invalid email', 'email', 'invalid-email');

    expect(error.name).toBe('ValidationError');
    expect(error.message).toBe('Invalid email');
    expect(error.field).toBe('email');
    expect(error.value).toBe('invalid-email');
    expect(error.stack).toBeDefined();
  });

  it('should serialize error correctly', () => {
    const error = new ValidationError('Invalid email', 'email', 'invalid-email');
    const serialized = JSON.stringify(error);
    const parsed = JSON.parse(serialized);

    expect(parsed.name).toBe('ValidationError');
    expect(parsed.field).toBe('email');
    expect(parsed.timestamp).toBeDefined();
  });
});
```

### Integration Testing Error Scenarios
```javascript
// Testing error propagation in integration scenarios
describe('User Registration Flow', () => {
  it('should handle database errors gracefully', async () => {
    const userData = { name: 'John', email: 'john@example.com' };

    // Mock database to throw error
    jest.spyOn(database, 'save').mockRejectedValue(
      new DatabaseError('Connection timeout', 'INSERT INTO users...', 'TIMEOUT')
    );

    const result = await registrationService.registerUser(userData);

    expect(result.success).toBe(false);
    expect(result.error.type).toBe('database');
    expect(result.error.retry).toBe(true);
  });

  it('should test error handling middleware', async () => {
    const request = {
      body: { invalid: 'data' }
    };
    const response = {
      status: jest.fn().mockReturnThis(),
      json: jest.fn()
    };
    const next = jest.fn();

    await errorHandlingMiddleware(request, response, next);

    expect(response.status).toHaveBeenCalledWith(400);
    expect(response.json).toHaveBeenCalledWith({
      error: 'Validation failed',
      details: expect.any(Object)
    });
  });
});
```

## Production Error Handling

### Error Monitoring and Logging
```javascript
// Structured error logging
class ErrorLogger {
  static log(error, context = {}) {
    const errorInfo = {
      timestamp: new Date().toISOString(),
      message: error.message,
      name: error.name,
      stack: error.stack,
      context,
      severity: this.getSeverity(error),
      fingerprint: this.generateFingerprint(error)
    };

    // Send to monitoring service
    this.sendToMonitoring(errorInfo);

    // Local logging
    console.error('[ERROR]', JSON.stringify(errorInfo, null, 2));
  }

  static getSeverity(error) {
    if (error instanceof ValidationError) return 'low';
    if (error instanceof NetworkError) return 'medium';
    if (error instanceof DatabaseError) return 'high';
    return 'medium';
  }

  static generateFingerprint(error) {
    // Create unique identifier for error grouping
    return btoa(`${error.name}:${error.message}:${error.stack?.split('\n')[1]}`);
  }

  static sendToMonitoring(errorInfo) {
    // Integration with Sentry, DataDog, etc.
    if (window.Sentry) {
      window.Sentry.captureException(error, {
        extra: errorInfo.context,
        fingerprint: [errorInfo.fingerprint]
      });
    }
  }
}

// Global error handlers
window.addEventListener('error', (event) => {
  ErrorLogger.log(event.error, {
    type: 'uncaught_exception',
    filename: event.filename,
    line: event.lineno,
    column: event.colno
  });
});

window.addEventListener('unhandledrejection', (event) => {
  ErrorLogger.log(event.reason, {
    type: 'unhandled_promise_rejection'
  });
});
```

### Error Recovery Strategies
```javascript
// Graceful degradation with fallbacks
class ResilientService {
  constructor() {
    this.fallbacks = [
      () => this.primaryAPI(),
      () => this.secondaryAPI(),
      () => this.cacheAPI(),
      () => this.defaultResponse()
    ];
  }

  async getData(id) {
    let lastError;

    for (const fallback of this.fallbacks) {
      try {
        const result = await fallback(id);
        if (result) return result;
      } catch (error) {
        lastError = error;
        ErrorLogger.log(error, {
          context: 'fallback_attempt',
          fallback: fallback.name,
          id
        });
      }
    }

    throw new Error(`All fallbacks failed. Last error: ${lastError?.message}`);
  }

  async primaryAPI(id) {
    // Primary implementation
    return await fetch(`/api/primary/${id}`);
  }

  async secondaryAPI(id) {
    // Secondary implementation
    return await fetch(`/api/secondary/${id}`);
  }

  async cacheAPI(id) {
    // Cache implementation
    return localStorage.getItem(`cache_${id}`);
  }

  defaultResponse() {
    // Default fallback
    return { id: null, data: 'default' };
  }
}

// Rate limiting and backoff
class RateLimitedService {
  constructor() {
    this.requestCount = 0;
    this.resetTime = Date.now() + 60000; // Reset every minute
    this.maxRequests = 100;
  }

  async makeRequest(url) {
    if (Date.now() > this.resetTime) {
      this.requestCount = 0;
      this.resetTime = Date.now() + 60000;
    }

    if (this.requestCount >= this.maxRequests) {
      const waitTime = this.resetTime - Date.now();
      throw new Error(`Rate limit exceeded. Try again in ${waitTime}ms`);
    }

    this.requestCount++;
    return await fetch(url);
  }
}
```

This comprehensive guide covers all aspects of error handling in TypeScript and JavaScript, from basic try-catch blocks to advanced patterns for production applications. Proper error handling is crucial for building resilient, maintainable applications that provide good user experiences even when things go wrong.