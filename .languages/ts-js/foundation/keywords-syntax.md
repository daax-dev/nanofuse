# TypeScript/JavaScript Keywords and Syntax Reference

## JavaScript Keywords (ES2024)

### Reserved Keywords
These keywords cannot be used as identifiers (variable names, function names, etc.):

```javascript
// Control Flow
if, else, for, while, do, switch, case, default, break, continue

// Functions and Scope
function, return, var, let, const

// Object-Oriented
class, extends, super, static, new, this

// Modules
import, export, from, as

// Error Handling
try, catch, finally, throw

// Operators
typeof, instanceof, in, delete, void

// Literals and Primitives
true, false, null, undefined

// Debugging
debugger

// Async
async, await

// Generators
yield

// Future Reserved (Strict Mode)
implements, interface, package, private, protected, public
```

### Contextual Keywords
These have special meaning in certain contexts but can be used as identifiers:

```javascript
// Modules
default, as, from

// Classes
get, set, static

// Async/Generators
async, await, yield

// TypeScript additions
declare, namespace, module, type, readonly, keyof, infer, is, asserts
```

## JavaScript Operators

### Arithmetic Operators
```javascript
+    // Addition or unary plus
-    // Subtraction or unary minus
*    // Multiplication
/    // Division
%    // Modulo (remainder)
**   // Exponentiation (ES2016)
++   // Increment
--   // Decrement
```

### Assignment Operators
```javascript
=     // Assignment
+=    // Addition assignment
-=    // Subtraction assignment
*=    // Multiplication assignment
/=    // Division assignment
%=    // Modulo assignment
**=   // Exponentiation assignment (ES2016)
&&=   // Logical AND assignment (ES2021)
||=   // Logical OR assignment (ES2021)
??=   // Nullish coalescing assignment (ES2021)
```

### Comparison Operators
```javascript
==    // Equality (with type coercion)
!=    // Inequality (with type coercion)
===   // Strict equality
!==   // Strict inequality
<     // Less than
<=    // Less than or equal
>     // Greater than
>=    // Greater than or equal
```

### Logical Operators
```javascript
&&    // Logical AND
||    // Logical OR
!     // Logical NOT
??    // Nullish coalescing (ES2020)
```

### Bitwise Operators
```javascript
&     // Bitwise AND
|     // Bitwise OR
^     // Bitwise XOR
~     // Bitwise NOT
<<    // Left shift
>>    // Sign-propagating right shift
>>>   // Zero-fill right shift
&=    // Bitwise AND assignment
|=    // Bitwise OR assignment
^=    // Bitwise XOR assignment
<<=   // Left shift assignment
>>=   // Right shift assignment
>>>=  // Zero-fill right shift assignment
```

### Other Operators
```javascript
?:    // Ternary conditional
?.    // Optional chaining (ES2020)
...   // Spread syntax / Rest parameters (ES2015)
,     // Comma operator
```

## Syntax Structures

### Variable Declarations
```javascript
// var - function-scoped, hoisted
var name = 'value';\n\n// let - block-scoped, temporal dead zone
let count = 0;\n\n// const - block-scoped, immutable binding
const PI = 3.14159;
const obj = { key: 'value' }; // Object contents can change
```

### Function Declarations and Expressions
```javascript
// Function declaration (hoisted)
function add(a, b) {
  return a + b;
}

// Function expression
const multiply = function(a, b) {
  return a * b;
};

// Arrow function (ES2015)
const subtract = (a, b) => a - b;
const divide = (a, b) => {
  return a / b;
};

// Method shorthand
const obj = {
  method() {
    return 'shorthand method';
  }
};

// Async functions
async function fetchData() {
  return await fetch('/api/data');
}

// Generator functions
function* generateNumbers() {
  yield 1;
  yield 2;
  yield 3;
}
```

### Classes
```javascript
// Class declaration
class Animal {
  constructor(name) {
    this.name = name;
  }

  speak() {
    console.log(`${this.name} makes a sound`);
  }

  static species() {
    return 'Unknown';
  }
}

// Class expression
const Bird = class {
  fly() {
    console.log('Flying');
  }
};

// Inheritance
class Dog extends Animal {
  constructor(name, breed) {
    super(name);
    this.breed = breed;
  }

  speak() {
    console.log(`${this.name} barks`);
  }
}

// Private fields (ES2022)
class BankAccount {
  #balance = 0;

  deposit(amount) {
    this.#balance += amount;
  }

  get balance() {
    return this.#balance;
  }
}
```

### Control Flow
```javascript
// if-else
if (condition) {
  // code
} else if (otherCondition) {
  // code
} else {
  // code
}

// switch
switch (value) {
  case 'a':
    // code
    break;
  case 'b':
  case 'c':
    // code for b or c
    break;
  default:
    // default code
}

// Loops
for (let i = 0; i < 10; i++) {
  // code
}

for (const item of iterable) {
  // code
}

for (const key in object) {
  // code
}

while (condition) {
  // code
}

do {
  // code
} while (condition);

// Labels and break/continue
outer: for (let i = 0; i < 3; i++) {
  for (let j = 0; j < 3; j++) {
    if (i === 1 && j === 1) {
      break outer; // breaks out of both loops
    }
  }
}
```

### Error Handling
```javascript
try {
  // code that might throw
  riskyOperation();
} catch (error) {
  // handle error
  console.error('Error:', error.message);
} finally {
  // cleanup code (always runs)
  cleanup();
}

// Throwing errors
throw new Error('Something went wrong');
throw new TypeError('Invalid type');

// Custom error classes
class CustomError extends Error {
  constructor(message) {
    super(message);
    this.name = 'CustomError';
  }
}
```

### Modules
```javascript
// Exporting
export const variable = 'value';
export function myFunction() {}
export class MyClass {}

// Default export
export default function() {}
export default class {}
export { variable as default };

// Re-exporting
export { something } from './other-module';
export * from './other-module';

// Importing
import defaultExport from './module';
import { namedExport } from './module';
import { originalName as alias } from './module';
import * as namespace from './module';
import './module'; // side effects only

// Dynamic imports
const module = await import('./module');
```

## TypeScript-Specific Syntax

### Type Annotations
```typescript
// Basic types
let name: string = 'John';
let age: number = 30;
let isActive: boolean = true;
let items: string[] = ['a', 'b', 'c'];
let values: Array<number> = [1, 2, 3];

// Function types
function greet(name: string): string {
  return `Hello, ${name}`;
}

const add: (a: number, b: number) => number = (a, b) => a + b;

// Object types
const person: { name: string; age: number } = {
  name: 'John',
  age: 30
};
```

### Interfaces and Types
```typescript
// Interface declaration
interface User {
  id: number;
  name: string;
  email?: string; // optional property
  readonly created: Date; // readonly property
}

// Extending interfaces
interface Admin extends User {
  permissions: string[];
}

// Type aliases
type Status = 'pending' | 'approved' | 'rejected';
type UserID = number | string;

// Generic types
interface Container<T> {
  value: T;
}

type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };
```

### Advanced Type Features
```typescript
// Union types
let value: string | number | boolean;

// Intersection types
type Employee = Person & { employeeId: string };

// Literal types
let direction: 'north' | 'south' | 'east' | 'west';

// Tuple types
let coordinates: [number, number] = [10, 20];
let namedTuple: [x: number, y: number] = [10, 20];

// Mapped types
type Partial<T> = {
  [P in keyof T]?: T[P];
};

// Conditional types
type NonNullable<T> = T extends null | undefined ? never : T;

// Template literal types
type EventName<T extends string> = `on${Capitalize<T>}`;

// Index signatures
interface Dictionary {
  [key: string]: any;
}
```

### Type Guards and Assertions
```typescript
// Type guards
function isString(value: any): value is string {
  return typeof value === 'string';
}

// Type assertions
let someValue: any = 'hello';
let strLength: number = (someValue as string).length;
let strLength2: number = (<string>someValue).length;

// Non-null assertion
function processUser(user: User | null) {
  console.log(user!.name); // Assert user is not null
}
```

### Utility Types
```typescript
// Built-in utility types
type UserUpdate = Partial<User>; // All properties optional
type UserCreate = Omit<User, 'id' | 'created'>; // Exclude properties
type UserKeys = keyof User; // Union of property names
type UserName = Pick<User, 'name'>; // Extract properties
type UserRequired = Required<User>; // All properties required
type UserReadonly = Readonly<User>; // All properties readonly

// Record type
type Roles = Record<string, string[]>; // { [key: string]: string[] }
```

### Decorators (Experimental)
```typescript
// Class decorators
@Component({
  selector: 'app-user'
})
class UserComponent {}

// Method decorators
class Service {
  @logged
  process() {
    // method implementation
  }
}

// Property decorators
class Model {
  @required
  @maxLength(50)
  name: string;
}

// Parameter decorators
class Controller {
  update(@Body() data: UpdateData) {
    // method implementation
  }
}
```

### Namespaces and Modules
```typescript
// Namespace declaration
namespace Geometry {
  export interface Point {
    x: number;
    y: number;
  }

  export function distance(p1: Point, p2: Point): number {
    return Math.sqrt((p2.x - p1.x) ** 2 + (p2.y - p1.y) ** 2);
  }
}

// Module augmentation
declare module 'existing-library' {
  interface ExistingInterface {
    newProperty: string;
  }
}

// Global augmentation
declare global {
  interface Window {
    customProperty: string;
  }
}
```

## Special Syntax Features

### Template Literals
```javascript
// Basic template literals
const name = 'World';
const greeting = `Hello, ${name}!`;

// Multi-line strings
const multiLine = `
  Line 1
  Line 2
  Line 3
`;

// Tagged template literals
function highlight(strings, ...values) {
  return strings.reduce((result, string, i) => {
    return result + string + (values[i] ? `<mark>${values[i]}</mark>` : '');
  }, '');
}

const highlighted = highlight`Hello ${name}, you are ${age} years old`;
```

### Destructuring
```javascript
// Array destructuring
const [first, second, ...rest] = [1, 2, 3, 4, 5];
const [a, , c] = [1, 2, 3]; // Skip second element

// Object destructuring
const { name, age, email = 'default@email.com' } = user;
const { name: userName, ...otherProps } = user;

// Nested destructuring
const { address: { city, country } } = user;

// Parameter destructuring
function processUser({ name, age }: User) {
  // function body
}
```

### Spread and Rest
```javascript
// Spread in arrays
const arr1 = [1, 2, 3];
const arr2 = [...arr1, 4, 5]; // [1, 2, 3, 4, 5]

// Spread in objects
const obj1 = { a: 1, b: 2 };
const obj2 = { ...obj1, c: 3 }; // { a: 1, b: 2, c: 3 }

// Rest in function parameters
function sum(...numbers: number[]): number {
  return numbers.reduce((total, num) => total + num, 0);
}

// Rest in destructuring
const [first, ...remaining] = [1, 2, 3, 4];
const { name, ...otherProps } = user;
```

### Object Shorthand
```javascript
// Property shorthand
const name = 'John';
const age = 30;
const user = { name, age }; // { name: 'John', age: 30 }

// Method shorthand
const obj = {
  method() {
    return 'shorthand method';
  },

  // Computed property names
  [dynamicKey]: 'value',
  [`${prefix}_property`]: 'computed value'
};

// Getter and setter shorthand
const obj = {
  _value: 0,
  get value() {
    return this._value;
  },
  set value(val) {
    this._value = val;
  }
};
```

## Regular Expressions
```javascript
// Literal notation
const regex1 = /pattern/flags;
const regex2 = /[a-z]+/gi;

// Constructor notation
const regex3 = new RegExp('pattern', 'flags');
const regex4 = new RegExp('[0-9]+', 'g');

// Common flags
// g - global
// i - case insensitive
// m - multiline
// s - dotAll
// u - unicode
// y - sticky

// Methods
const result = text.match(/pattern/);
const isMatch = /pattern/.test(text);
const replaced = text.replace(/pattern/g, 'replacement');
```

## Comments
```javascript
// Single line comment

/*
  Multi-line comment
  Can span multiple lines
*/

/**
 * JSDoc comment
 * @param {string} name - The name parameter
 * @returns {string} The greeting message
 */
function greet(name) {
  return `Hello, ${name}`;
}
```

## Strict Mode
```javascript
'use strict';

// Strict mode changes:
// - Variables must be declared
// - Function parameters must be unique
// - Octal literals are forbidden
// - eval and arguments cannot be used as identifiers
// - this is undefined in functions (not global object)
```

## Semicolon Rules
```javascript
// Semicolons are optional but recommended
const a = 1
const b = 2

// ASI (Automatic Semicolon Insertion) can cause issues:
const obj = {
  method: function() {
    return
      {
        status: 'OK'
      }
  }
}
// Returns undefined due to ASI after 'return'

// Safe approach - always use semicolons:
const obj = {
  method: function() {
    return {
      status: 'OK'
    };
  }
};
```

## TypeScript Configuration Syntax
```json
// tsconfig.json
{
  \"compilerOptions\": {
    \"target\": \"ES2020\",
    \"module\": \"ESNext\",
    \"moduleResolution\": \"node\",
    \"strict\": true,
    \"esModuleInterop\": true,
    \"allowSyntheticDefaultImports\": true,
    \"resolveJsonModule\": true,
    \"isolatedModules\": true,
    \"noEmit\": true,
    \"jsx\": \"react-jsx\"
  },
  \"include\": [\"src/**/*\"],
  \"exclude\": [\"node_modules\", \"dist\"]
}
```

This comprehensive reference covers all major keywords and syntax elements in both JavaScript and TypeScript, providing a complete foundation for understanding and using these languages effectively.