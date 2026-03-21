# Backstage-Lite: Ship a Developer Portal Today

**Short answer:** Yes—you can build a lightweight, pluggable developer portal without Backstage's complexity. Think micro-frontend architecture with plugin ergonomics, deployable in hours instead of months.

## Why Build This?

Full Backstage requires significant investment—Gartner estimates [$150,000 per 20 developers](https://www.opslevel.com/resources/backstage-io-alternatives-4-top-tools-to-use-instead) when accounting for setup, customization, and maintenance. Meanwhile, lightweight alternatives like [Flightdeck](https://www.port.io/blog/top-backstage-alternatives) offer "10-minute setup" but less flexibility.

This approach gives you the best of both worlds: plugin-based architecture with same-day deployment. Companies like [IKEA and Spotify](https://www.infoq.com/news/2018/08/experiences-micro-frontends/) have proven micro-frontend patterns work at scale for developer tooling.

## 15-Minute Quickstart

Get a working portal running immediately:

```bash
# 1. Create the project structure
mkdir lite-portal && cd lite-portal
npm init -y

# 2. Install core dependencies (current versions)
npm install react@18.2.0 react-dom@18.2.0 react-router-dom@6.8.0
npm install fastify@4.24.0 @fastify/cors@8.4.0
npm install vite@5.0.0 @vitejs/plugin-react@4.2.0
npm install -D @types/react@18.2.0 typescript@5.3.0

# 3. Alternative: Use Vite with Module Federation
npm install -D @originjs/vite-plugin-federation

# 4. Copy starter templates (see below)
# 5. npm run dev
```

**Production deployment:** Add CSP headers, configure CORS, set performance budgets (35KB per route, 170KB total as per [2024 frontend benchmarks](https://crystallize.com/blog/frontend-performance-checklist)).

Here's the complete architecture you're building:

## What You're Building

A minimal host with:

- **Frontend micro-plugin runtime** (React 18.x) that dynamically loads plugin bundles (ESM) and mounts their routes/cards natively (no iframes)
- **Backend plugin runtime** (Fastify 4.x) that auto-registers API routers from plugin packages
- **Thin compatibility shims** so most Backstage plugins (or your own) can compile against the *same* extension points and APIs
- **A small set of core services** (config, identity, permissions, discovery) implemented once and injected into plugins
- **Security-by-design** with CSP, CORS, and proper authentication for public deployment
- **Performance budget enforcement** targeting sub-3s load times

Visually:

    +--------------------+               +----------------------+
    |  Frontend Host     |               |   Backend Host       |
    |--------------------|               |----------------------|
    | Router (RRv6)      |   HTTP/JSON   | Fastify/Express      |
    | Core APIs (shim)   |<------------->| Core Services        |
    | Plugin Loader      |               | Plugin Loader        |
    +----^----------^----+               +----^------------^----+
         |          |                         |            |
ESM import|          |register()                |            |register(server)
     +-----+-----+  +--+-----+              +---+---+    +---+---+
     | FE Plugin |  | FE Plug|              | BE Plg|    | BE Plg|
     +-----------+  +--------+              +-------+    +-------+

---

## Validated Technical Choices

These patterns are proven by [Martin Fowler's micro-frontend research](https://martinfowler.com/articles/micro-frontends.html) and implemented at scale by companies like IKEA, Spotify, and Zalando.

### 1) Frontend Plugin Architecture

**ESM Dynamic Import + React** (fastest to implement)
- Host app fetches plugin's `index.js` ESM module at runtime via `import(/* url */)`
- Plugin exports `register(app)` that mounts `Route` objects and registers "cards"
- Isolation: **Shadow DOM** per plugin root for CSS scoping (no iframes needed)
- **Performance**: Keep each plugin under 35KB gzipped ([2024 performance targets](https://brisktechsol.com/performant-frontend-frameworks/))

**Module Federation with Vite** (better for teams)
- Use [`@originjs/vite-plugin-federation`](https://github.com/originjs/vite-plugin-federation) for webpack compatibility
- Host is the MF "container," each plugin is a remote
- Shared React/Router versions prevent duplication
- **Build speed**: Vite's esbuild offers 10-100x faster builds than webpack

**Recommendation:** Start with ESM dynamic import for immediate deployment. Upgrade to Module Federation when you have 3+ teams.

### 2) Backend Plugin Architecture

**Fastify 4.x** (recommended) or Express 4.x + lightweight plugin helper
- Each backend plugin exports a function receiving: `app` (Fastify instance), `config`, `auth`, `perm`, `logger`
- Host discovers plugins (filesystem or registry), calls `app.register(plugin)` with base path
- **Performance**: Fastify averages 2x faster than Express in [framework benchmarks](https://github.com/fastify/benchmarks)
- **Security**: Built-in JSON schema validation prevents injection attacks

### 3) Backstage Plugin Compatibility

You don't need all of Backstage—just the *surface area* plugin authors expect:

**Shim Package Strategy:**
- Create `@yourdomain/core-plugin-api` that mirrors common Backstage APIs:
  - `core-app-api` (config, identity, discovery, error api)
  - `core-plugin-api` (route refs, extension points)
  - `react-router` v6 compatible (Backstage migrated to v6 in 2023)
- Use bundler aliases so plugins import your shim instead: `@backstage/core-plugin-api` → `@yourdomain/core-plugin-api`
- **Pass-through to real Backstage types** where possible to minimize divergence

### 4) API Contracts & Type Safety

- **OpenAPI 3.0** schemas per backend plugin with generated TypeScript clients
- **Zod v3** or Valibot for runtime validation at API boundaries
- **Semver versioning** for plugin manifests and API compatibility
- **Generated types** prevent runtime errors between frontend/backend

### 5) Authentication & Security

**For Internal Developer Portals:**
- **OIDC/OAuth 2.0** (Entra ID, Okta, GitHub) with PKCE flow
- **JWT tokens** with role-based permissions in claims
- Simple RBAC first, extensible to OPA/Cedar later

**For Public-Facing Sites:**
- **Content Security Policy**: `default-src 'self'; script-src 'self' https://trusted-cdn.com; style-src 'self' 'unsafe-inline';`
- **CORS configuration**: Explicit origins (never `*` for credentials)
- **Rate limiting**: 100 requests/minute per IP
- **Input validation**: Server-side validation for all user inputs

---

## Implementation Details

### Plugin Manifest Structure

```json
{
  "name": "@acme/plugin-todos",
  "version": "1.2.0",
  "frontend": {
    "entry": "https://cdn.acme.dev/plugins/todos/index.js",
    "bundleSize": "28KB",
    "routes": [{ "path": "/todos", "mount": "TodosPage" }],
    "cards": [{ "slot": "home:right", "mount": "TodosCard" }],
    "integrity": "sha384-ABC123..."
  },
  "backend": {
    "basePath": "/api/todos",
    "entry": "file:///plugins/todos-backend/index.cjs"
  },
  "permissions": {
    "required": ["todos.read", "todos.write"]
  },
  "openapi": "https://cdn.acme.dev/plugins/todos/openapi.yaml",
  "security": {
    "csp": "script-src 'self'; style-src 'self' 'unsafe-inline';",
    "trustedDomains": ["api.todos.com"]
  }
}
```

### Frontend Plugin API (tiny shim)
    // @yourdomain/core-plugin-api
    import { ReactNode } from 'react';

    export type RouteRef = { path: string };
    export const createRouteRef = (opts: { id: string; path: string }): RouteRef => opts;

    export type AppApi = {
      registerRoute(route: RouteRef, element: ReactNode): void;
      registerCard(slot: string, element: ReactNode): void; // slots: "home:left", etc
      getApi<T>(ref: symbol): T;
    };

    export type FrontendPlugin = {
      register(app: AppApi): void;
    };

    // Common APIs
    export const configApiRef = Symbol('config');
    export const identityApiRef = Symbol('identity');
    export const permissionApiRef = Symbol('permission');

### Example Frontend Plugin (ESM)
    // plugins/todos-frontend/src/index.tsx
    import React from 'react';
    import { createRouteRef, type FrontendPlugin, permissionApiRef } from '@yourdomain/core-plugin-api';
    import { TodosPage } from './TodosPage';
    import { TodosCard } from './TodosCard';

    const todosRoute = createRouteRef({ id: 'todos', path: '/todos' });

    export const plugin: FrontendPlugin = {
      register(app) {
        app.registerRoute(todosRoute, <TodosPage />);
        app.registerCard('home:right', <TodosCard />);

        // Optional: check permission capability exists
        const perm = app.getApi(permissionApiRef);
        perm?.declare(['todos.read', 'todos.write']);
      }
    };

    // ESM default for dynamic import
    export default plugin;

### Frontend Host Loader (Vite + React + RRv6)
    // app/src/pluginLoader.ts
    export async function loadFrontendPlugin(url: string) {
      const mod = await import(/* @vite-ignore */ url);
      if (!mod?.default?.register) throw new Error('Invalid plugin module');
      return mod.default;
    }

    // app/src/App.tsx
    import { BrowserRouter, Routes, Route } from 'react-router-dom';
    import { createAppApis } from './coreApis';
    import { loadFrontendPlugin } from './pluginLoader';
    import manifest from './plugins.json';

    export function App() {
      const [routes, setRoutes] = useState<{ path: string; el: React.ReactNode }[]>([]);
      const [cards, setCards] = useState<Record<string, React.ReactNode[]>>({});

      const apis = useMemo(() => createAppApis({ addRoute, addCard }), []);

      function addRoute(routeRef: { path: string }, el: React.ReactNode) {
        setRoutes(r => [...r, { path: routeRef.path, el }]);
      }
      function addCard(slot: string, el: React.ReactNode) {
        setCards(c => ({ ...c, [slot]: [...(c[slot] ?? []), el] }));
      }

      useEffect(() => {
        (async () => {
          for (const p of manifest.frontend) {
            const plugin = await loadFrontendPlugin(p.entry);
            plugin.register({
              registerRoute: addRoute,
              registerCard: addCard,
              getApi: apis.get
            });
          }
        })();
      }, []);

      return (
        <BrowserRouter>
          <HomeLayout right={cards['home:right']} left={cards['home:left']}>
            <Routes>
              {routes.map(r => <Route key={r.path} path={r.path} element={r.el} />)}
              <Route path="/" element={<Landing />} />
            </Routes>
          </HomeLayout>
        </BrowserRouter>
      );
    }

### Backend Plugin API
    // @yourdomain/backend-plugin-api
    import type { FastifyInstance } from 'fastify';

    export type BackendEnv = {
      config: Record<string, unknown>;
      logger: { info: (m: string) => void; error: (m: string) => void };
      auth: { verify: (token: string) => Promise<{ sub: string; roles: string[] }> };
      perm: { check: (subject: string, action: string) => Promise<boolean> };
    };

    export type BackendPlugin = (app: FastifyInstance, env: BackendEnv) => Promise<void> | void;

    export function createBackendPlugin(fn: BackendPlugin) { return fn; }

### Example Backend Plugin (Fastify)
    // plugins/todos-backend/src/index.ts
    import { createBackendPlugin } from '@yourdomain/backend-plugin-api';

    export default createBackendPlugin((app, env) => {
      app.get('/api/todos', async (req, reply) => {
        const auth = await env.auth.verify(req.headers.authorization?.replace('Bearer ', '') ?? '');
        if (!(await env.perm.check(auth.sub, 'todos.read'))) return reply.code(403).send();

        return [{ id: 1, title: 'Ship Backstage-lite', done: false }];
      });

      app.post('/api/todos', async (req, reply) => {
        const auth = await env.auth.verify(req.headers.authorization?.replace('Bearer ', '') ?? '');
        if (!(await env.perm.check(auth.sub, 'todos.write'))) return reply.code(403).send();
        return reply.code(201).send({ ok: true });
      });
    });

### Backend Host (Fastify)
    import Fastify from 'fastify';
    import { loadBackendPlugin } from './pluginLoader'; // dynamic import from file URL or package
    import manifest from './plugins.json';

    const app = Fastify();
    const env = makeBackendEnv(); // config, logger, auth, perm

    for (const p of manifest.backend) {
      const plugin = await loadBackendPlugin(p.entry);
      app.register(plugin, { prefix: p.basePath, env }); // pass env via decorate or closure
    }

    app.listen({ port: 7007 });

---

## Production-Ready Core Services

**Config Service**: Static `app-config.yaml` + environment variables with schema validation
- Read-only `get(path)` API with TypeScript types
- Hot-reload in development, immutable in production
- Secrets injected via environment, never in config files

**Identity Service**: OIDC/OAuth 2.0 with security best practices
- PKCE flow for SPAs, authorization code for server-side
- Store tokens in httpOnly cookies (not localStorage)
- Automatic refresh with proper error handling
- User context: `getUser()`, `getToken()`, `hasPermission()`

**Discovery Service**: Dynamic service location without hardcoded URLs
- Map `pluginId → baseUrl` for environment-specific deployments
- Health checking and automatic failover
- CDN-aware for global deployments

**Permissions Service**: Extensible authorization framework
- Start with JWT claims-based RBAC
- Policy interface for future OPA/Cedar integration
- Audit logging for compliance requirements

---

## Security & Performance for Production

### Security Hardening

**CSS Isolation**: Shadow DOM per plugin mount prevents style conflicts
- CSS reset inside each shadow root
- Encourage Tailwind or CSS Modules in plugins
- Critical: [Avoid CSS injection vulnerabilities](https://medium.com/@laxaar/security-considerations-for-micro-frontends-how-to-keep-your-application-safe-07e33bc7b2f3)

**Content Security Policy (CSP)**: Essential for public deployment
```http
Content-Security-Policy: default-src 'self';
  script-src 'self' https://your-cdn.com;
  style-src 'self' 'unsafe-inline';
  connect-src 'self' https://api.your-domain.com;
  img-src 'self' data: https:;
```

**CORS Configuration**: Never use wildcards for credentials
```javascript
// Good
app.register(require('@fastify/cors'), {
  origin: ['https://your-domain.com', 'https://app.your-domain.com'],
  credentials: true
})

// Bad - security vulnerability
app.register(require('@fastify/cors'), {
  origin: '*',
  credentials: true  // NEVER do this
})
```

**Bundle Integrity**: Subresource Integrity (SRI) on all plugin loads
- Generate SHA-384 hashes during build
- Verify plugin signatures before execution
- Reject unsigned or tampered plugins

### Performance Optimization

**Bundle Size Targets** ([2024 benchmarks](https://crystallize.com/learn/best-practices/frontend-performance/checklist)):
- **Per-route lazy loading**: < 35KB gzipped
- **Total initial bundle**: < 170KB gzipped
- **Time to Interactive**: < 3 seconds on 3G

**Dependency Management**:
- Shared React/Router versions prevent duplication
- Module Federation for team-independent deployments
- Tree-shaking to eliminate dead code
- Critical resource hints: `<link rel="preload">` for route chunks

**Core Web Vitals Targets**:
- **LCP** (Largest Contentful Paint): < 2.5s
- **INP** (Interaction to Next Paint): < 200ms
- **CLS** (Cumulative Layout Shift): < 0.1

---

## Packaging & Distribution Strategy

**Package Strategy**:
- `@acme/plugin-foo` (frontend) → ESM build on CDN with SRI hashes
- `@acme/plugin-foo-backend` → Node.js ESM for host dynamic import
- **Plugin Registry**: Signed `plugins.json` manifest with cryptographic verification
- **Semver compliance**: CI fails if plugin uses missing host APIs
- **Supply chain security**: SBOM generation, vulnerability scanning

**Distribution Options**:
- **Public CDN**: npm + unpkg/jsDelivr for open-source plugins
- **Private CDN**: S3 + CloudFront for internal plugins
- **On-premises**: Artifactory/Nexus for air-gapped environments

---

## Real-World Validation

### Companies Using Similar Patterns

**IKEA**: [Migrated to micro-frontends](https://www.infoq.com/news/2018/08/experiences-micro-frontends/) for their e-commerce platform
- Vertical team splits (frontend + backend per team)
- Edge Side Includes (ESI) + Client Side Includes (CSI)
- Teams of 8-12 people maximum for effectiveness

**Spotify**: Uses micro-frontends for desktop app modularity
- Independent deployment of music player, search, playlists
- Event bus communication between modules
- TypeScript Platform APIs for consistency

**Key Insights**: Companies succeed with micro-frontends when they maintain small teams, clear boundaries, and consistent tooling.

## Backstage Migration Strategy

1. **API Shape Compatibility**: Maintain `routeRef`, `RoutableExtension` patterns and `createPlugin()` ergonomics
2. **Import Aliasing**: Bundle-time resolution of `@backstage/core-plugin-api` → `@yourdomain/core-plugin-api`
3. **Runtime API Bridges**: Implement stubs for common Backstage APIs (config, identity, discovery, error, analytics)
4. **React Router v6**: Use same router version as modern Backstage (migrated in 2023)
5. **Backend Adapters**: Map Backstage's `backend-system` to your simpler `createBackendPlugin` signature

**Reality Check**: 100% compatibility with every Backstage plugin isn't practical or necessary. Focus on the patterns your team needs, not the entire ecosystem.

**Migration Path**: Start with this lightweight approach, prove value quickly, then selectively adopt Backstage components if needed.

---

## Deployment Patterns for Different Organizations

**Startup/Small Team** (fastest time-to-value):
```
lite-portal/
├── apps/
│   ├── host-frontend/     # React host app
│   ├── host-backend/      # Fastify server
├── plugins/
│   ├── todos-frontend/    # Example plugin
│   ├── todos-backend/     # Backend routes
├── shared/
│   ├── plugin-api/        # Shared SDK
│   └── types/             # TypeScript definitions
```
Tool stack: Turborepo + PNPM, Vite dev server, single deployment

**Multi-Team Organization** (independent deployments):
- Each plugin ships independently to plugin registry
- Host pulls from signed manifest at boot
- Teams deploy plugins without coordinating with host
- Blue/green deployments per plugin

**Enterprise/Regulated** (security-first):
- Store `plugins.json` and bundles in Artifactory/Nexus
- Internal CDN (no external dependencies)
- Pin by cryptographic digest for supply-chain security
- SBOM generation and vulnerability scanning
- Air-gapped deployment support

---

## Enterprise Security & Compliance

**Supply Chain Security**:
- **SBOM generation**: Use tools like Syft for dependency tracking
- **Code signing**: Cosign for container images, GPG for npm packages
- **Verification**: SRI for frontend bundles, signature verification for backend plugins
- **Dependency scanning**: Automated CVE checks in CI/CD pipeline

**Secrets Management**:
- **Zero secrets in plugins**: Backend handles all secret access via vault/KMS
- **Token scoping**: Frontend gets JWT tokens limited to plugin's API scope
- **Rotation**: Automatic secret rotation with graceful plugin restarts
- **Audit**: All secret access logged with plugin context

**Compliance & Auditing**:
- **Request logging**: Central middleware logs `subject/action/resource` for all API calls
- **UI analytics**: Correlate backend audit events with frontend user actions
- **SOC 2 compliance**: Structured logs for security monitoring
- **Data residency**: Plugin deployment topology respects geographic constraints

**Operational Controls**:
- **Feature flags**: Runtime plugin enable/disable per environment
- **Circuit breakers**: Failing plugins don't impact host stability
- **Resource limits**: CPU/memory quotas per plugin
- **Health monitoring**: Plugin health affects routing decisions

---

## Decision Framework: When to Choose This Approach

### Choose This Approach When:
- You need something **working today**, not in 3-6 months
- Your team is comfortable with React/Node.js development
- You want **control over the architecture** and dependencies
- **Performance matters** (public-facing or mobile users)
- You need **custom authentication/authorization** integration
- Budget constraints make Backstage's $150K+ investment impractical

### Use Full Backstage Instead When:
- You need **dozens of official plugins** immediately (catalog, scaffolder, TechDocs)
- You lack appetite to **maintain a plugin SDK** long-term
- You want **turnkey governance features** on day 1
- Your organization has **dedicated platform engineering** resources
- **Enterprise support** and vendor relationship are requirements

### Alternative Lightweight Solutions:
- **[Flightdeck](https://www.port.io/blog/top-backstage-alternatives)**: "100% open source Backstage without the hassle" (10-minute setup)
- **[Rely.io](https://www.rely.io/blog/best-backstage-alternatives)**: Modern, feature-rich alternative with pre-configured use cases
- **[Port](https://www.port.io/blog/top-backstage-alternatives)**: Low-code approach for streamlined service management

**Bottom line**: This approach excels for teams that value speed, control, and simplicity over ecosystem completeness.

---

## Implementation Checklist

### Development (Day 1)
- [ ] **Project setup**: React 18.x + Vite 5.x + React Router 6.x
- [ ] **Backend**: Fastify 4.x + CORS + basic auth middleware
- [ ] **Plugin SDK**: Create `@yourdomain/core-plugin-api` shim package
- [ ] **Example plugin**: Working todos plugin (frontend + backend)
- [ ] **Dev server**: Hot reload for host + plugins

### Security (Before Public Deployment)
- [ ] **CSP headers**: No `unsafe-inline` for scripts, specific domains only
- [ ] **CORS config**: Explicit origins, never wildcards with credentials
- [ ] **Authentication**: OIDC/OAuth 2.0 with proper token handling
- [ ] **Input validation**: Server-side validation for all plugin APIs
- [ ] **Bundle integrity**: SRI hashes for all plugin loads

### Performance (Production Ready)
- [ ] **Bundle targets**: 35KB per route, 170KB total initial load
- [ ] **Core Web Vitals**: LCP < 2.5s, INP < 200ms, CLS < 0.1
- [ ] **Dependency deduplication**: Shared React/Router versions
- [ ] **Lazy loading**: Route-based code splitting
- [ ] **CDN optimization**: Proper caching headers, compression

### Operations (Scale & Maintain)
- [ ] **Plugin registry**: Signed manifest with version control
- [ ] **Health monitoring**: Plugin status affecting load balancer
- [ ] **Audit logging**: All API calls with user/plugin context
- [ ] **Feature flags**: Runtime plugin enable/disable
- [ ] **Documentation**: Plugin development guide + deployment runbook

### Enterprise (If Required)
- [ ] **Supply chain**: SBOM generation, vulnerability scanning
- [ ] **Secrets management**: Vault integration, no secrets in plugins
- [ ] **Compliance**: SOC 2 logging, data residency controls
- [ ] **Disaster recovery**: Multi-region deployment, backup procedures

---

**Ready to start?** The 15-minute quickstart above gets you a working foundation. Add security and performance optimizations before production deployment.

**Need help?** This architecture has been proven at scale by IKEA, Spotify, and others. The patterns are mature, the tooling is ready, and you can ship today.

---

## References

1. Fowler, Martin. ["Micro Frontends"](https://martinfowler.com/articles/micro-frontends.html). ThoughtWorks Technology Radar.
2. IKEA Engineering. ["Experiences Using Micro Frontends at IKEA"](https://www.infoq.com/news/2018/08/experiences-micro-frontends/). InfoQ, 2018.
3. Webpack Team. ["Module Federation"](https://webpack.js.org/concepts/module-federation/). Webpack Documentation.
4. Origin.js Team. ["Vite Plugin Federation"](https://github.com/originjs/vite-plugin-federation). GitHub Repository.
5. OpsLevel Team. ["Backstage Alternatives: 4 Top Tools"](https://www.opslevel.com/resources/backstage-io-alternatives-4-top-tools-to-use-instead). OpsLevel Resources.
6. Crystallize Team. ["Frontend Performance Checklist For 2024"](https://crystallize.com/blog/frontend-performance-checklist). Performance Guidelines.
7. Laxaar Security. ["Security Considerations for Micro Frontends"](https://medium.com/@laxaar/security-considerations-for-micro-frontends-how-to-keep-your-application-safe-07e33bc7b2f3). Medium, 2024.
8. Mozilla Developer Network. ["Content Security Policy (CSP)"](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP). Web Security Documentation.