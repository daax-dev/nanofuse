# 17 React Best Practices for 2024

> **Source Video**: [17 React Best Practices](https://www.youtube.com/watch?v=5r25Y9Vg2P4)
>
> This document outlines critical best practices that separate junior from senior React developers. These practices focus on structure, reusability, consistency, and creating robust web applications.

---

## Table of Contents

1. [Extract Hardcoded Values (Magic Numbers/Strings)](#1-extract-hardcoded-values-magic-numbersstrings)
2. [Organize with Proper Folder Structure](#2-organize-with-proper-folder-structure)
3. [Create Components Strategically](#3-create-components-strategically)
4. [Avoid Unnecessary Divs - Use Fragments](#4-avoid-unnecessary-divs---use-fragments)
5. [Prevent Prop Drilling with Children Pattern](#5-prevent-prop-drilling-with-children-pattern)
6. [Keep Components Simple and Dumb](#6-keep-components-simple-and-dumb)
7. [Use Functional State Updates](#7-use-functional-state-updates)
8. [Pass Handler Functions, Not Setters](#8-pass-handler-functions-not-setters)
9. [Follow Naming Conventions for Event Props](#9-follow-naming-conventions-for-event-props)
10. [Optimize with useMemo for Expensive Calculations](#10-optimize-with-usememo-for-expensive-calculations)
11. [Optimize with useCallback for Functions](#11-optimize-with-usecallback-for-functions)
12. [Prevent Re-renders with React.memo](#12-prevent-re-renders-with-reactmemo)
13. [Use Single State for Related Values](#13-use-single-state-for-related-values)
14. [Avoid useEffect for Data Fetching](#14-avoid-useeffect-for-data-fetching)
15. [Use Specialized Libraries for Data Fetching](#15-use-specialized-libraries-for-data-fetching)
16. [Break Down Large Components](#16-break-down-large-components)
17. [Create Custom Hooks for Reusable Logic](#17-create-custom-hooks-for-reusable-logic)

---

## 1. Extract Hardcoded Values (Magic Numbers/Strings)

### The Problem

Hardcoding values directly in components makes code harder to maintain and scale.

```tsx
// ❌ Bad - Hardcoded values scattered throughout component
if (todos.length >= 3 && !isAuthenticated) {
  alert('You need to sign in to add more than 3 todos');
}

const sensitiveWords = ['password', 'credit card'];
if (sensitiveWords.includes(content)) {
  // ...
}
```

### The Solution

Extract constants to a dedicated file for centralized management.

```tsx
// ✅ Good - constants.ts
export const MAXIMUM_FREE_TODOS = 3;
export const SENSITIVE_WORDS = ['password', 'credit card'];

// Component file
if (todos.length >= MAXIMUM_FREE_TODOS && !isAuthenticated) {
  alert(`You need to sign in to add more than ${MAXIMUM_FREE_TODOS} todos`);
}
```

### Benefits

- **Easy to update**: Change value in one place, updates everywhere
- **Better organization**: All business logic values in one location
- **Automatic propagation**: All instances automatically updated
- **Reduced errors**: No need to hunt down multiple hardcoded values

### References

- [React Best Practices - Constants](https://react.dev/learn/sharing-state-between-components#lifting-state-up-by-example)
- [Clean Code: Constants](https://github.com/ryanmcdermott/clean-code-javascript#variables)

---

## 2. Organize with Proper Folder Structure

### The Principle

There's no single "correct" folder structure, but consistency and semantic organization are crucial.

### Recommended Structure

```
src/
├── components/
│   ├── Header/
│   │   ├── Header.tsx
│   │   ├── Logo.tsx
│   │   └── Header.css
│   ├── auth/
│   │   ├── LoginButton.tsx
│   │   └── ProfileComponent.tsx
│   └── Button.tsx
├── contexts/       # or 'store/' if using Redux/Zustand
│   └── TodosContext.tsx
├── lib/           # or 'utils/'
│   ├── constants.ts
│   ├── hooks.ts
│   ├── helpers.ts
│   └── types.ts
└── App.tsx
```

### Key Principles

- **Consistency**: Use the same pattern throughout
- **Semantic grouping**: Group by feature or component type
- **Easy onboarding**: New developers should understand structure quickly
- **Separate concerns**: Components ≠ State Management ≠ Utilities

### References

- [React File Structure Best Practices](https://react.dev/learn/thinking-in-react#step-1-break-the-ui-into-a-component-hierarchy)
- [Folder Structure Conventions](https://github.com/alan2207/bulletproof-react/blob/master/docs/project-structure.md)

---

## 3. Create Components Strategically

### When to Create Components

1. **For Reusability**: Any UI pattern used more than once
2. **For Organization**: Even single-use components improve code structure
3. **Semantic Clarity**: Component names document what markup represents

### Example

```tsx
// ❌ Bad - Everything in one component
function App() {
  return (
    <div>
      <h1 className="background-heading">Todo App</h1>
      <main>
        <ul>
          {todos.map(todo => <li key={todo.id}>{todo.content}</li>)}
        </ul>
      </main>
      <footer>
        <p>&copy; 2024 Todo App</p>
      </footer>
    </div>
  );
}

// ✅ Good - Separated into semantic components
function App() {
  return (
    <div>
      <BackgroundHeading />
      <main>
        <TodoList />
      </main>
      <Footer />
    </div>
  );
}
```

### Benefits

- **Better organization**: Clear component hierarchy
- **Easier maintenance**: Update one component, not scattered markup
- **Semantic meaning**: Names describe purpose
- **Reusability**: Ready for multi-instance use

### References

- [Thinking in React - Breaking UI into Components](https://react.dev/learn/thinking-in-react)
- [Component Design Patterns](https://www.patterns.dev/posts/presentational-container-pattern)

---

## 4. Avoid Unnecessary Divs - Use Fragments

### The Problem

Wrapping elements in `<div>` tags when you need to return multiple elements can break layouts.

```tsx
// ❌ Bad - Unnecessary div breaks layout
function Sidebar() {
  return (
    <div className="sidebar">
      {isAuthenticated ? (
        <Button>Logout</Button>
      ) : (
        <div> {/* Breaks flexbox/grid layout! */}
          <Button>Login</Button>
          <Button>Register</Button>
        </div>
      )}
    </div>
  );
}
```

### The Solution

Use React Fragments which don't add to the DOM.

```tsx
// ✅ Good - Fragment doesn't affect layout
function Sidebar() {
  return (
    <div className="sidebar">
      {isAuthenticated ? (
        <Button>Logout</Button>
      ) : (
        <> {/* or <React.Fragment> */}
          <Button>Login</Button>
          <Button>Register</Button>
        </>
      )}
    </div>
  );
}
```

### Benefits

- **Preserves layout**: No extra DOM nodes
- **Cleaner HTML**: Inspecting DOM shows expected structure
- **CSS compatibility**: Flexbox/Grid work as intended

### References

- [React Fragments Documentation](https://react.dev/reference/react/Fragment)
- [Why React Fragments are Better](https://kentcdodds.com/blog/avoid-nesting-when-youre-testing)

---

## 5. Prevent Prop Drilling with Children Pattern

### The Problem

Passing props through components that don't need them creates fragile chains.

```tsx
// ❌ Bad - Prop drilling
function App() {
  const [todos, setTodos] = useState([]);
  return <Sidebar setTodos={setTodos} />;
}

function Sidebar({ setTodos }) {
  // Sidebar doesn't use setTodos, just passes it down
  return <AddTodoForm setTodos={setTodos} />;
}
```

### The Solution

Use the children pattern to pass components directly.

```tsx
// ✅ Good - Children pattern eliminates prop drilling
function App() {
  const [todos, setTodos] = useState([]);
  return (
    <Sidebar>
      <AddTodoForm setTodos={setTodos} />
    </Sidebar>
  );
}

function Sidebar({ children }) {
  return <div className="sidebar">{children}</div>;
}
```

### Additional Benefits

- **Performance**: Children don't re-render when parent re-renders
- **Simpler components**: Less props = easier to understand
- **Flexibility**: Easy to add/remove nested components

### Alternative Solutions

- **Context API**: For app-wide state
- **Zustand/Redux**: For complex state management
- **React Query**: For server state

### References

- [Composition vs Inheritance](https://react.dev/learn/passing-props-to-a-component#passing-jsx-as-children)
- [Prop Drilling Solutions](https://kentcdodds.com/blog/prop-drilling)

---

## 6. Keep Components Simple and Dumb

### The Problem

Components shouldn't know about complex business logic they don't directly use.

```tsx
// ❌ Bad - Component knows too much
function AddTodoForm({ setTodos, setModelOpen, todos }) {
  const handleSubmit = (e) => {
    e.preventDefault();
    setTodos(prev => [...prev, newTodo]);

    if (todos.length === 10) {
      setModelOpen(true); // Too much responsibility!
    }
  };

  return <form onSubmit={handleSubmit}>...</form>;
}
```

### The Solution

Lift logic up and pass down simple handler functions.

```tsx
// ✅ Good - Component is simple and focused
function App() {
  const [todos, setTodos] = useState([]);
  const [modelOpen, setModelOpen] = useState(false);

  const handleAddTodo = (content) => {
    setTodos(prev => [...prev, { id: prev.length + 1, content, completed: false }]);

    if (todos.length === 10) {
      setModelOpen(true);
    }
  };

  return <AddTodoForm onAddTodo={handleAddTodo} />;
}

function AddTodoForm({ onAddTodo }) {
  const handleSubmit = (e) => {
    e.preventDefault();
    onAddTodo(content);
  };

  return <form onSubmit={handleSubmit}>...</form>;
}
```

### Benefits

- **Easier to test**: Simple components with clear inputs
- **Better reusability**: No tight coupling to specific logic
- **Clear responsibility**: State logic centralized near state

### References

- [Smart vs Dumb Components](https://medium.com/@dan_abramov/smart-and-dumb-components-7ca2f9a7c7d0)
- [Container/Presentational Pattern](https://www.patterns.dev/posts/presentational-container-pattern)

---

## 7. Use Functional State Updates

### The Problem

Directly referencing state when updating can cause issues with closures and batching.

```tsx
// ❌ Bad - Depends on stale state
const [count, setCount] = useState(0);

setCount(count + 1); // May use stale value
setCount(count + 1); // Won't work as expected
```

### The Solution

Use the updater function when new state depends on old state.

```tsx
// ✅ Good - Always uses current state
setCount(prev => prev + 1);
setCount(prev => prev + 1); // Works correctly!

setTodos(prev => [...prev, newTodo]); // Reliable
```

### When to Use

- State depends on previous value
- Inside closures (setTimeout, setInterval)
- Multiple rapid updates
- Passing to child components

### References

- [useState Updater Functions](https://react.dev/reference/react/useState#updating-state-based-on-the-previous-state)
- [State Updates and Closures](https://epicreact.dev/how-react-uses-closures-to-avoid-bugs/)

---

## 8. Pass Handler Functions, Not Setters

### The Principle

Encapsulate complex update logic in handler functions instead of exposing raw setters.

```tsx
// ❌ Bad - Passing raw setter
function App() {
  const [todos, setTodos] = useState([]);
  return <AddTodoForm setTodos={setTodos} />;
}

// ✅ Good - Passing handler with encapsulated logic
function App() {
  const [todos, setTodos] = useState([]);
  const [modelOpen, setModelOpen] = useState(false);

  const handleAddTodo = (content) => {
    setTodos(prev => [
      ...prev,
      { id: prev.length + 1, content, completed: false }
    ]);

    // Additional logic centralized here
    if (prev.length + 1 === 10) {
      setModelOpen(true);
    }
  };

  const handleDeleteTodo = (id) => {
    setTodos(prev => prev.filter(todo => todo.id !== id));
  };

  const handleEditTodo = (id, newContent) => {
    setTodos(prev => prev.map(todo =>
      todo.id === id ? { ...todo, content: newContent } : todo
    ));
  };

  return <TodoApp onAddTodo={handleAddTodo} />;
}
```

### Benefits

- **Centralized state logic**: All state updates in one place
- **Simpler child components**: Less knowledge required
- **Easier to extend**: Add side effects without modifying children
- **Better testing**: Test handlers independently

### References

- [Lifting State Up](https://react.dev/learn/sharing-state-between-components)
- [React Patterns - Compound Components](https://kentcdodds.com/blog/compound-components-with-react-hooks)

---

## 9. Follow Naming Conventions for Event Props

### The Convention

Follow HTML/DOM naming patterns for consistency and predictability.

```tsx
// ✅ Good - Follows DOM conventions
<button onClick={handleClick}>Click me</button>
<input onChange={handleChange} />

// ✅ Good - Custom components follow same pattern
<AddTodoForm onAddTodo={handleAddTodo} />
<TodoItem
  onEdit={handleEdit}
  onDelete={handleDelete}
  onToggle={handleToggle}
/>

// Component definition
interface AddTodoFormProps {
  onAddTodo: (content: string) => void; // Prop name: on[Event]
}

function AddTodoForm({ onAddTodo }: AddTodoFormProps) {
  const handleSubmit = (e) => { // Handler name: handle[Event]
    e.preventDefault();
    onAddTodo(content);
  };

  return <form onSubmit={handleSubmit}>...</form>;
}
```

### Pattern

- **Props**: `on[EventName]` (onAddTodo, onClick, onChange)
- **Handlers**: `handle[EventName]` (handleAddTodo, handleClick, handleChange)

### Benefits

- **Consistency**: Matches native React/DOM patterns
- **Predictability**: Developers know what to expect
- **Clear intent**: Signals this is an event callback

### References

- [React Event Naming Conventions](https://react.dev/learn/responding-to-events#naming-event-handler-props)
- [JavaScript Naming Conventions](https://github.com/airbnb/javascript#naming-conventions)

---

## 10. Optimize with useMemo for Expensive Calculations

### The Problem

Expensive computations run on every render, even when dependencies haven't changed.

```tsx
// ❌ Bad - Recalculates every render
function TodoList({ todos }) {
  const completedCount = todos.filter(todo => todo.completed).length;
  const stats = calculateComplexStats(todos); // Runs every render!

  return <div>...</div>;
}
```

### The Solution

Use `useMemo` to cache calculations and only recompute when dependencies change.

```tsx
// ✅ Good - Only recalculates when todos change
function TodoList({ todos }) {
  const completedCount = useMemo(() => {
    return todos.filter(todo => todo.completed).length;
  }, [todos]);

  const stats = useMemo(() => {
    return calculateComplexStats(todos);
  }, [todos]);

  return <div>...</div>;
}
```

### When to Use

- Expensive calculations (filtering, sorting, complex math)
- Derived state from props/state
- Objects/arrays passed as dependencies to useEffect
- Objects/arrays passed as props to memoized components

### Note

The React compiler (future) will auto-optimize these cases, but currently manual memoization is required.

### References

- [useMemo Hook](https://react.dev/reference/react/useMemo)
- [When to useMemo and useCallback](https://kentcdodds.com/blog/usememo-and-usecallback)

---

## 11. Optimize with useCallback for Functions

### The Problem

Functions are recreated on every render, potentially causing child re-renders.

```tsx
// ❌ Bad - New function every render
function App() {
  const [todos, setTodos] = useState([]);

  const handleAddTodo = (content) => {
    setTodos(prev => [...prev, newTodo]);
  };

  return <AddTodoForm onAddTodo={handleAddTodo} />;
}
```

### The Solution

Use `useCallback` to memoize functions between renders.

```tsx
// ✅ Good - Function only recreated when todos change
function App() {
  const [todos, setTodos] = useState([]);

  const handleAddTodo = useCallback((content) => {
    setTodos(prev => [
      ...prev,
      { id: prev.length + 1, content, completed: false }
    ]);
  }, [todos]); // Or [] if using functional updates

  return <AddTodoForm onAddTodo={handleAddTodo} />;
}
```

### When to Use

- Functions passed to memoized child components
- Functions used in dependency arrays
- Event handlers passed down multiple levels
- Functions with expensive creation cost

### Works Best With

Combine with `React.memo` on child components for maximum benefit.

### References

- [useCallback Hook](https://react.dev/reference/react/useCallback)
- [useCallback vs useMemo](https://kentcdodds.com/blog/usememo-and-usecallback)

---

## 12. Prevent Re-renders with React.memo

### The Problem

Child components re-render even when their props haven't changed.

```tsx
// Component re-renders on every parent render
function ExpensiveComponent({ data, onAction }) {
  // Expensive rendering logic
  return <div>...</div>;
}
```

### The Solution

Wrap components in `React.memo` to skip re-renders when props are unchanged.

```tsx
// ✅ Good - Only re-renders when props change
const ExpensiveComponent = React.memo(({ data, onAction }) => {
  // Expensive rendering logic
  return <div>...</div>;
});

// Or
export default React.memo(ExpensiveComponent);
```

### Critical Requirement

Props must be stable (memoized) for this to work:

```tsx
function Parent() {
  const [count, setCount] = useState(0);
  const data = useMemo(() => ({ /* ... */ }), []);
  const handleAction = useCallback(() => { /* ... */ }, []);

  // Now memo works because props don't change unnecessarily
  return <ExpensiveComponent data={data} onAction={handleAction} />;
}
```

### When to Use

- Components with expensive render logic
- Components that render frequently
- Components deep in the tree
- When profiling shows unnecessary re-renders

### References

- [React.memo](https://react.dev/reference/react/memo)
- [React Performance Optimization](https://react.dev/learn/render-and-commit#optimizing-performance)

---

## 13. Use Single State for Related Values

### The Problem

Multiple boolean states for mutually exclusive conditions.

```tsx
// ❌ Bad - Multiple booleans for related states
const [isLoading, setIsLoading] = useState(false);
const [isError, setIsError] = useState(false);
const [isEmpty, setIsEmpty] = useState(false);

// Hard to maintain consistency
setIsLoading(false);
setIsError(true);
setIsEmpty(false);
```

### The Solution

Use a single state with discrete values for related conditions.

```tsx
// ✅ Good - Single state with all possibilities
type Status = 'idle' | 'loading' | 'error' | 'empty' | 'success';
const [status, setStatus] = useState<Status>('idle');

// Clear and atomic updates
setStatus('loading');
setStatus('error');
setStatus('empty');

// Easy to check
{status === 'loading' && <Spinner />}
{status === 'error' && <ErrorMessage />}
{status === 'empty' && <EmptyState />}
```

### Benefits

- **Atomic updates**: Impossible to have inconsistent state
- **Type safety**: TypeScript ensures valid values
- **Easier to reason about**: One source of truth
- **Prevents bugs**: Can't be both loading and error

### References

- [Managing State](https://react.dev/learn/managing-state)
- [State Machines in React](https://kentcdodds.com/blog/implementing-a-simple-state-machine-library-in-javascript)

---

## 14. Avoid useEffect for Data Fetching

### The Problem

Using `useEffect` for data fetching leads to race conditions, no caching, and boilerplate code.

```tsx
// ❌ Bad - Manual data fetching with useEffect
function JobItem({ id }) {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setIsLoading(true);
    fetch(`/api/jobs/${id}`)
      .then(res => res.json())
      .then(data => {
        setData(data);
        setIsLoading(false);
      })
      .catch(err => {
        setError(err);
        setIsLoading(false);
      });
  }, [id]); // Race condition if id changes quickly!

  // No caching - fetches same data repeatedly
  // No request deduplication
  // Manual loading/error state management
}
```

### Issues with useEffect Data Fetching

- Race conditions when dependencies change
- No automatic caching
- Repeated fetches for same data
- Manual loading/error state
- Network waterfalls
- No request deduplication

### References

- [You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect)
- [Common useEffect Mistakes](https://kentcdodds.com/blog/useeffect-vs-uselayouteffect)

---

## 15. Use Specialized Libraries for Data Fetching

### The Solution

Use libraries designed for data fetching and server state management.

```tsx
// ✅ Good - Using React Query
function useJobItem(id) {
  return useQuery({
    queryKey: ['job', id],
    queryFn: () => fetch(`/api/jobs/${id}`).then(res => res.json())
  });
}

function JobItem({ id }) {
  const { data, isLoading, error } = useJobItem(id);

  if (isLoading) return <Spinner />;
  if (error) return <Error error={error} />;

  return <div>{data.title}</div>;
}
```

### Benefits of React Query / SWR

- **Automatic caching**: Fetches once, reuses data
- **Background refetching**: Keeps data fresh
- **Request deduplication**: Multiple components requesting same data = one request
- **Optimistic updates**: Update UI before server responds
- **Pagination/infinite scroll**: Built-in support
- **Automatic retries**: Configurable retry logic
- **Stale-while-revalidate**: Show cached data while fetching fresh
- **DevTools**: Inspect cache and queries

### Recommended Libraries

- **[React Query (TanStack Query)](https://tanstack.com/query)**: Most popular, feature-rich
- **[SWR](https://swr.vercel.app/)**: Lightweight, by Vercel
- **[RTK Query](https://redux-toolkit.js.org/rtk-query/overview)**: If using Redux Toolkit
- **Next.js**: Built-in with Server Components and data caching

### Example: Automatic Caching

```tsx
// First click - fetches from server
<JobItem id={1} /> // Fetches...

// Second click on same item - instant from cache!
<JobItem id={1} /> // Cached! ⚡

// Different item - fetches
<JobItem id={2} /> // Fetches...

// Back to first - still cached!
<JobItem id={1} /> // Cached! ⚡
```

### References

- [React Query Documentation](https://tanstack.com/query/latest/docs/react/overview)
- [SWR Documentation](https://swr.vercel.app/)
- [Data Fetching Patterns](https://www.patterns.dev/posts/data-fetching)

---

## 16. Break Down Large Components

### The Problem

Large components with mixed concerns are hard to read, test, and maintain.

```tsx
// ❌ Bad - Everything in one component
function TodoList() {
  const [todos, setTodos] = useState([]);

  useEffect(() => {
    localStorage.setItem('todos', JSON.stringify(todos));
  }, [todos]);

  return (
    <div>
      {todos.length === 0 ? (
        <p>No todos yet. Add one to get started!</p>
      ) : null}

      <ul>
        {todos.map(todo => (
          <li key={todo.id}>
            <input
              type="checkbox"
              checked={todo.completed}
              onChange={() => {/* toggle logic */}}
            />
            <span>
              {todo.content.charAt(0).toUpperCase() + todo.content.slice(1)}
            </span>
            <button onClick={() => {/* delete logic */}}>Delete</button>
          </li>
        ))}
      </ul>

      <button onClick={() => setTodos([])}>Reset</button>
      <button onClick={() => {/* mark all complete logic */}}>
        Complete All
      </button>
    </div>
  );
}
```

### The Solution

Extract components and utilities for better organization.

```tsx
// ✅ Good - Broken into focused pieces

// Utility function
function capitalizeFirstLetter(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

// Small focused components
function EmptyState() {
  return <p>No todos yet. Add one to get started!</p>;
}

function TodoItem({ todo, onToggle, onDelete }) {
  return (
    <li>
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo.id)}
      />
      <span>{capitalizeFirstLetter(todo.content)}</span>
      <DeleteButton onDelete={() => onDelete(todo.id)} />
    </li>
  );
}

function DeleteButton({ onDelete }) {
  return <button onClick={onDelete}>Delete</button>;
}

// Main component - clean and readable
function TodoList() {
  const [todos, setTodos] = useLocalStorage('todos', []);

  const handleToggle = (id) => {
    setTodos(prev => prev.map(todo =>
      todo.id === id ? { ...todo, completed: !todo.completed } : todo
    ));
  };

  const handleDelete = (id) => {
    setTodos(prev => prev.filter(todo => todo.id !== id));
  };

  const handleReset = () => setTodos([]);

  const handleCompleteAll = () => {
    setTodos(prev => prev.map(todo => ({ ...todo, completed: true })));
  };

  return (
    <div>
      {todos.length === 0 && <EmptyState />}

      <ul>
        {todos.map(todo => (
          <TodoItem
            key={todo.id}
            todo={todo}
            onToggle={handleToggle}
            onDelete={handleDelete}
          />
        ))}
      </ul>

      <Button onClick={handleReset}>Reset</Button>
      <Button onClick={handleCompleteAll}>Complete All</Button>
    </div>
  );
}
```

### Refactoring Strategy

1. **Extract markup into components**: Reusable UI elements
2. **Extract logic into utility functions**: Reusable operations
3. **Extract hooks into custom hooks**: Reusable stateful logic
4. **Name components semantically**: Document what markup represents

### Benefits

- **Easier to understand**: Scan names instead of parsing code
- **Easier to test**: Test small units in isolation
- **Easier to reuse**: Components and utilities available elsewhere
- **Easier to maintain**: Changes isolated to relevant files

### References

- [Extracting Components](https://react.dev/learn/thinking-in-react#step-2-build-a-static-version-in-react)
- [Single Responsibility Principle](https://www.patterns.dev/posts/presentational-container-pattern)

---

## 17. Create Custom Hooks for Reusable Logic

### The Principle

Extract stateful logic into custom hooks for reusability and cleaner components.

### When to Create Custom Hooks

**Create a custom hook when you want to reuse logic that uses React hooks.**

```tsx
// ❌ Bad - Stateful logic duplicated across components
function TodoList() {
  const [todos, setTodos] = useState([]);

  useEffect(() => {
    const saved = localStorage.getItem('todos');
    if (saved) setTodos(JSON.parse(saved));
  }, []);

  useEffect(() => {
    localStorage.setItem('todos', JSON.stringify(todos));
  }, [todos]);

  // Component logic...
}

function Settings() {
  const [settings, setSettings] = useState({});

  useEffect(() => {
    const saved = localStorage.getItem('settings');
    if (saved) setSettings(JSON.parse(saved));
  }, []);

  useEffect(() => {
    localStorage.setItem('settings', JSON.stringify(settings));
  }, [settings]);

  // Component logic...
}
```

### The Solution

```tsx
// ✅ Good - Custom hook encapsulates logic
function useLocalStorage<T>(key: string, initialValue: T) {
  const [value, setValue] = useState<T>(() => {
    const saved = localStorage.getItem(key);
    return saved ? JSON.parse(saved) : initialValue;
  });

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value));
  }, [key, value]);

  return [value, setValue] as const;
}

// Now reusable everywhere
function TodoList() {
  const [todos, setTodos] = useLocalStorage('todos', []);
  // Clean component logic...
}

function Settings() {
  const [settings, setSettings] = useLocalStorage('settings', {});
  // Clean component logic...
}
```

### The Reusability Hierarchy

```
For Markup       → Create a React Component
For Logic        → Create a Utility Function
For Logic + Hook → Create a Custom Hook
```

### Benefits

- **Reusability**: Share stateful logic across components
- **Cleaner components**: Extract complex logic
- **Easier testing**: Test hooks independently
- **Better organization**: Dedicated files for complex logic
- **Semantic clarity**: Names describe what logic does

### Naming Convention

Always prefix custom hooks with `use` to follow React's rules and enable lint checks.

```tsx
useLocalStorage  ✅
useWindowSize    ✅
useDebounce      ✅
useAuth          ✅

getLocalStorage  ❌ (if it uses hooks)
windowSize       ❌ (if it uses hooks)
```

### References

- [Building Your Own Hooks](https://react.dev/learn/reusing-logic-with-custom-hooks)
- [Custom Hook Patterns](https://usehooks.com/)
- [React Hooks Best Practices](https://kentcdodds.com/blog/react-hooks-pitfalls)

---

## Summary

These 17 best practices form the foundation of professional React development:

### Code Organization
1. Extract constants
2. Organize folders semantically
3. Create components strategically
4. Use fragments instead of divs

### Component Design
5. Prevent prop drilling
6. Keep components simple
7. Use functional state updates
8. Pass handlers, not setters
9. Follow naming conventions

### Performance
10. Use useMemo for calculations
11. Use useCallback for functions
12. Use React.memo for components
13. Single state for related values

### Data Management
14. Avoid useEffect for fetching
15. Use specialized libraries

### Code Quality
16. Break down large components
17. Create custom hooks

### Master React Professionally

Following these practices will:
- Improve code maintainability
- Enhance application performance
- Make debugging easier
- Enable better team collaboration
- Create more robust applications

The difference between junior and senior React developers isn't just knowing React—it's having internalized these patterns and practices.

---

## Additional Resources

- **Official React Documentation**: [https://react.dev](https://react.dev)
- **React Patterns**: [https://www.patterns.dev/react](https://www.patterns.dev/react)
- **React TypeScript Cheatsheet**: [https://react-typescript-cheatsheet.netlify.app/](https://react-typescript-cheatsheet.netlify.app/)
- **Epic React by Kent C. Dodds**: [https://epicreact.dev](https://epicreact.dev)
- **TanStack Query**: [https://tanstack.com/query](https://tanstack.com/query)
- **React Performance**: [https://react.dev/learn/render-and-commit](https://react.dev/learn/render-and-commit)

---

**Last Updated**: January 2025
**Video Source**: [17 React Best Practices You Need to Know](https://www.youtube.com/watch?v=5r25Y9Vg2P4)
