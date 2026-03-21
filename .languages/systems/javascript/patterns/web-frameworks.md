# JavaScript Web Framework Patterns & Integration

## Frontend Framework Architectures

### **React Ecosystem Patterns**

#### Component Composition
```jsx
// Higher-Order Component (HOC) Pattern
const withAuth = (WrappedComponent) => {
  return function AuthenticatedComponent(props) {
    const { user, loading } = useAuth();
    
    if (loading) return <LoadingSpinner />;
    if (!user) return <LoginPrompt />;
    
    return <WrappedComponent {...props} user={user} />;
  };
};

// Custom Hook Pattern
function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged((user) => {
      setUser(user);
      setLoading(false);
    });
    
    return unsubscribe;
  }, []);
  
  return { user, loading };
}

// Compound Component Pattern
function Tabs({ children, defaultTab }) {
  const [activeTab, setActiveTab] = useState(defaultTab);
  
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

Tabs.List = function TabList({ children }) {
  return <div className="tab-list">{children}</div>;
};

Tabs.Tab = function Tab({ id, children }) {
  const { activeTab, setActiveTab } = useContext(TabsContext);
  return (
    <button 
      className={activeTab === id ? 'active' : ''}
      onClick={() => setActiveTab(id)}
    >
      {children}
    </button>
  );
};

Tabs.Panel = function TabPanel({ id, children }) {
  const { activeTab } = useContext(TabsContext);
  return activeTab === id ? <div>{children}</div> : null;
};
```

#### State Management Patterns
```jsx
// Context + Reducer Pattern
const AppStateContext = createContext();
const AppDispatchContext = createContext();

function appReducer(state, action) {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, loading: action.payload };
    case 'SET_USER':
      return { ...state, user: action.payload, loading: false };
    case 'SET_ERROR':
      return { ...state, error: action.payload, loading: false };
    default:
      throw new Error(`Unhandled action type: ${action.type}`);
  }
}

function AppProvider({ children }) {
  const [state, dispatch] = useReducer(appReducer, {
    user: null,
    loading: true,
    error: null
  });
  
  return (
    <AppStateContext.Provider value={state}>
      <AppDispatchContext.Provider value={dispatch}>
        {children}
      </AppDispatchContext.Provider>
    </AppStateContext.Provider>
  );
}

// Custom hooks for state access
function useAppState() {
  const context = useContext(AppStateContext);
  if (!context) {
    throw new Error('useAppState must be used within AppProvider');
  }
  return context;
}

function useAppDispatch() {
  const context = useContext(AppDispatchContext);
  if (!context) {
    throw new Error('useAppDispatch must be used within AppProvider');
  }
  return context;
}
```

### **Vue.js Composition Patterns**

#### Composable Functions
```javascript
// Data fetching composable
export function useApi(url, options = {}) {
  const data = ref(null);
  const error = ref(null);
  const loading = ref(false);
  
  async function execute() {
    loading.value = true;
    error.value = null;
    
    try {
      const response = await fetch(url, options);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      data.value = await response.json();
    } catch (err) {
      error.value = err.message;
    } finally {
      loading.value = false;
    }
  }
  
  // Auto-execute on mount if not explicitly disabled
  if (options.immediate !== false) {
    onMounted(execute);
  }
  
  return {
    data: readonly(data),
    error: readonly(error),
    loading: readonly(loading),
    execute
  };
}

// Form handling composable
export function useForm(initialValues, validationSchema) {
  const values = reactive({ ...initialValues });
  const errors = reactive({});
  const touched = reactive({});
  
  function validateField(field) {
    if (validationSchema[field]) {
      const result = validationSchema[field](values[field]);
      if (result !== true) {
        errors[field] = result;
        return false;
      } else {
        delete errors[field];
        return true;
      }
    }
    return true;
  }
  
  function handleSubmit(onSubmit) {
    return async (event) => {
      event.preventDefault();
      
      // Validate all fields
      let isValid = true;
      Object.keys(validationSchema).forEach(field => {
        touched[field] = true;
        if (!validateField(field)) isValid = false;
      });
      
      if (isValid) {
        await onSubmit(values);
      }
    };
  }
  
  return {
    values,
    errors: readonly(errors),
    touched: readonly(touched),
    validateField,
    handleSubmit
  };
}
```

#### Plugin Architecture
```javascript
// Vue plugin pattern
export default {
  install(app, options = {}) {
    // Global properties
    app.config.globalProperties.$api = createApiClient(options.apiUrl);
    
    // Global components
    app.component('LoadingSpinner', LoadingSpinner);
    app.component('ErrorBoundary', ErrorBoundary);
    
    // Directives
    app.directive('focus', {
      mounted(el) {
        el.focus();
      }
    });
    
    // Provide/inject
    app.provide('theme', reactive({ 
      dark: options.darkMode || false 
    }));
  }
};

// Usage in main.js
import { createApp } from 'vue';
import MyPlugin from './plugins/MyPlugin';

const app = createApp(App);
app.use(MyPlugin, {
  apiUrl: 'https://api.example.com',
  darkMode: true
});
```

## Backend Framework Patterns

### **Express.js Middleware Architecture**

#### Middleware Composition
```javascript
// Error handling middleware
function errorHandler(err, req, res, next) {
  console.error(err.stack);
  
  if (err.name === 'ValidationError') {
    return res.status(400).json({
      error: 'Validation Error',
      details: err.details
    });
  }
  
  if (err.name === 'UnauthorizedError') {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Invalid or missing token'
    });
  }
  
  res.status(500).json({
    error: 'Internal Server Error',
    message: process.env.NODE_ENV === 'development' ? err.message : 'Something went wrong'
  });
}

// Authentication middleware factory
function requireAuth(options = {}) {
  return async (req, res, next) => {
    try {
      const token = req.headers.authorization?.replace('Bearer ', '');
      
      if (!token) {
        throw new UnauthorizedError('No token provided');
      }
      
      const decoded = jwt.verify(token, process.env.JWT_SECRET);
      const user = await User.findById(decoded.userId);
      
      if (!user) {
        throw new UnauthorizedError('User not found');
      }
      
      // Check permissions if required
      if (options.permissions) {
        const hasPermission = options.permissions.every(
          permission => user.permissions.includes(permission)
        );
        
        if (!hasPermission) {
          throw new ForbiddenError('Insufficient permissions');
        }
      }
      
      req.user = user;
      next();
    } catch (error) {
      next(error);
    }
  };
}

// Rate limiting middleware
function rateLimit(options = {}) {
  const { windowMs = 15 * 60 * 1000, max = 100, message } = options;
  const requests = new Map();
  
  return (req, res, next) => {
    const key = req.ip;
    const now = Date.now();
    
    // Clean old entries
    for (const [ip, data] of requests.entries()) {
      if (now - data.resetTime > windowMs) {
        requests.delete(ip);
      }
    }
    
    const requestData = requests.get(key) || { count: 0, resetTime: now };
    
    if (now - requestData.resetTime > windowMs) {
      requestData.count = 0;
      requestData.resetTime = now;
    }
    
    requestData.count++;
    requests.set(key, requestData);
    
    if (requestData.count > max) {
      return res.status(429).json({
        error: 'Too Many Requests',
        message: message || `Rate limit exceeded. Try again in ${Math.ceil(windowMs / 1000)} seconds.`
      });
    }
    
    res.set({
      'X-RateLimit-Limit': max,
      'X-RateLimit-Remaining': Math.max(0, max - requestData.count),
      'X-RateLimit-Reset': new Date(requestData.resetTime + windowMs)
    });
    
    next();
  };
}
```

#### Route Organization
```javascript
// routes/users.js
import { Router } from 'express';
import { body, param } from 'express-validator';
import { requireAuth, validateRequest } from '../middleware/index.js';
import * as userController from '../controllers/userController.js';

const router = Router();

// Validation schemas
const createUserSchema = [
  body('email').isEmail().normalizeEmail(),
  body('password').isLength({ min: 8 }).matches(/^(?=.*[A-Za-z])(?=.*\d)/),
  body('name').isLength({ min: 2, max: 50 }).trim()
];

const updateUserSchema = [
  param('id').isUUID(),
  body('name').optional().isLength({ min: 2, max: 50 }).trim(),
  body('email').optional().isEmail().normalizeEmail()
];

// Routes with middleware composition
router.post('/users', 
  createUserSchema,
  validateRequest,
  userController.createUser
);

router.get('/users/:id', 
  param('id').isUUID(),
  validateRequest,
  requireAuth(),
  userController.getUserById
);

router.put('/users/:id',
  updateUserSchema,
  validateRequest,
  requireAuth({ permissions: ['user:update'] }),
  userController.updateUser
);

export default router;
```

### **Fastify Plugin Pattern**

```javascript
// plugins/database.js
import fp from 'fastify-plugin';
import { MongoClient } from 'mongodb';

async function dbPlugin(fastify, options) {
  const client = new MongoClient(options.uri);
  await client.connect();
  
  const db = client.db(options.database);
  
  fastify.decorate('mongo', {
    client,
    db,
    ObjectId: import('mongodb').ObjectId
  });
  
  fastify.addHook('onClose', async () => {
    await client.close();
  });
}

export default fp(dbPlugin, {
  name: 'database-plugin'
});

// plugins/auth.js
import fp from 'fastify-plugin';
import jwt from '@fastify/jwt';

async function authPlugin(fastify, options) {
  await fastify.register(jwt, {
    secret: options.secret
  });
  
  fastify.decorate('authenticate', async (request, reply) => {
    try {
      await request.jwtVerify();
    } catch (err) {
      reply.send(err);
    }
  });
  
  fastify.decorate('requirePermissions', (permissions) => {
    return async (request, reply) => {
      await request.jwtVerify();
      
      const userPermissions = request.user.permissions || [];
      const hasAllPermissions = permissions.every(
        permission => userPermissions.includes(permission)
      );
      
      if (!hasAllPermissions) {
        reply.code(403).send({
          error: 'Insufficient permissions',
          required: permissions
        });
      }
    };
  });
}

export default fp(authPlugin, {
  name: 'auth-plugin',
  dependencies: ['database-plugin']
});
```

## Full-Stack Integration Patterns

### **Next.js Full-Stack Pattern**

#### API Routes with Middleware
```javascript
// lib/middleware.js
export function withAuth(handler) {
  return async (req, res) => {
    try {
      const token = req.headers.authorization?.replace('Bearer ', '');
      if (!token) {
        return res.status(401).json({ error: 'No token provided' });
      }
      
      const decoded = jwt.verify(token, process.env.JWT_SECRET);
      req.user = decoded;
      
      return handler(req, res);
    } catch (error) {
      return res.status(401).json({ error: 'Invalid token' });
    }
  };
}

export function withValidation(schema) {
  return (handler) => async (req, res) => {
    try {
      const validated = await schema.parseAsync(req.body);
      req.body = validated;
      return handler(req, res);
    } catch (error) {
      return res.status(400).json({
        error: 'Validation Error',
        details: error.errors
      });
    }
  };
}

// pages/api/users/[id].js
import { z } from 'zod';
import { withAuth, withValidation } from '../../../lib/middleware';

const updateUserSchema = z.object({
  name: z.string().min(2).max(50).optional(),
  email: z.string().email().optional()
});

async function handler(req, res) {
  const { id } = req.query;
  
  switch (req.method) {
    case 'GET':
      const user = await getUserById(id);
      return res.json(user);
      
    case 'PUT':
      const updated = await updateUser(id, req.body);
      return res.json(updated);
      
    default:
      res.setHeader('Allow', ['GET', 'PUT']);
      return res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}

export default withAuth(withValidation(updateUserSchema)(handler));
```

#### Server-Side Rendering with Data Fetching
```javascript
// pages/users/[id].js
import { GetServerSideProps } from 'next';
import { z } from 'zod';

const UserPageProps = z.object({
  user: z.object({
    id: z.string(),
    name: z.string(),
    email: z.string().email()
  }),
  posts: z.array(z.object({
    id: z.string(),
    title: z.string(),
    createdAt: z.string()
  }))
});

export default function UserPage({ user, posts }) {
  return (
    <div>
      <h1>{user.name}</h1>
      <p>{user.email}</p>
      
      <h2>Recent Posts</h2>
      {posts.map(post => (
        <div key={post.id}>
          <h3>{post.title}</h3>
          <time>{new Date(post.createdAt).toLocaleDateString()}</time>
        </div>
      ))}
    </div>
  );
}

export const getServerSideProps: GetServerSideProps = async ({ params }) => {
  try {
    const [user, posts] = await Promise.all([
      fetch(`${process.env.API_URL}/users/${params?.id}`),
      fetch(`${process.env.API_URL}/users/${params?.id}/posts`)
    ]);
    
    const userData = await user.json();
    const postsData = await posts.json();
    
    const props = UserPageProps.parse({
      user: userData,
      posts: postsData
    });
    
    return { props };
  } catch (error) {
    return { notFound: true };
  }
};
```

### **GraphQL Integration Pattern**

```javascript
// Apollo Server with Express
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { buildSubgraphSchema } from '@apollo/subgraph';
import { gql } from 'graphql-tag';

const typeDefs = gql`
  type User @key(fields: "id") {
    id: ID!
    name: String!
    email: String!
    posts: [Post!]!
  }
  
  type Post {
    id: ID!
    title: String!
    content: String!
    author: User!
  }
  
  type Query {
    users: [User!]!
    user(id: ID!): User
    posts: [Post!]!
  }
  
  type Mutation {
    createUser(input: CreateUserInput!): User!
    updateUser(id: ID!, input: UpdateUserInput!): User!
  }
  
  input CreateUserInput {
    name: String!
    email: String!
  }
  
  input UpdateUserInput {
    name: String
    email: String
  }
`;

const resolvers = {
  Query: {
    users: async () => {
      return await User.findAll();
    },
    user: async (_, { id }) => {
      return await User.findByPk(id);
    },
    posts: async () => {
      return await Post.findAll({ include: [User] });
    }
  },
  
  Mutation: {
    createUser: async (_, { input }, { user }) => {
      if (!user) throw new AuthenticationError('Must be authenticated');
      return await User.create(input);
    },
    
    updateUser: async (_, { id, input }, { user }) => {
      if (!user || user.id !== id) {
        throw new ForbiddenError('Cannot update other users');
      }
      
      const [updatedRowsCount] = await User.update(input, { 
        where: { id },
        returning: true 
      });
      
      return await User.findByPk(id);
    }
  },
  
  User: {
    posts: async (user) => {
      return await Post.findAll({ where: { userId: user.id } });
    }
  },
  
  Post: {
    author: async (post) => {
      return await User.findByPk(post.userId);
    }
  }
};

const server = new ApolloServer({
  schema: buildSubgraphSchema({ typeDefs, resolvers }),
  context: async ({ req }) => {
    const token = req.headers.authorization || '';
    let user = null;
    
    if (token) {
      try {
        const decoded = jwt.verify(token.replace('Bearer ', ''), process.env.JWT_SECRET);
        user = await User.findByPk(decoded.userId);
      } catch (error) {
        console.warn('Invalid token');
      }
    }
    
    return { user };
  }
});

app.use('/graphql', expressMiddleware(server));
```

These patterns provide a comprehensive foundation for building scalable, maintainable web applications across the JavaScript ecosystem, from component-based frontends to robust backend APIs and full-stack integrations.