1) Product & Domain Scope

1.1 Audience & access tiers (anonymous vs. registered vs. admin; regional restrictions).
1.2 Content vs. user-generated data boundaries (what lives in CMS vs. app DB).
1.3 Core user journeys (pre-login, onboarding, post-login flows).
1.4 Compliance domain (PII, payments, minors, regulated industries).
1.5 SLA/uptime expectations and failure tolerance (read-only modes, degraded UX).

2) Information Architecture & Content Modeling

2.1 Content types, fields, taxonomies, relationships, and versioning needs.
2.2 URL and routing strategy (content slugs vs. IDs; canonicalization; trailing slashes).
2.3 Localization/internationalization (copy, media, currency, calendars; fallback rules).
2.4 Content lifecycle states (draft, review, scheduled, archived) and auditability.
2.5 Preview requirements (draft visibility, shareable previews, time-based previews).

3) Rendering & Navigation Model

3.1 Server-rendered vs. client-rendered vs. hybrid “islands” granularity.
3.2 First-load performance vs. interactivity trade-offs (TTFB, LCP, hydration cost).
3.3 Navigation semantics (document navigations vs. in-app transitions; back/forward cache).
3.4 SEO constraints (crawlability, structured data, robots rules).
3.5 Accessibility baseline (semantics, focus management, keyboard nav, ARIA budgets).

4) API Strategy & Service Boundaries

4.1 Backend-for-Frontend vs. direct service access from the client.
4.2 Contract type (REST vs. GraphQL vs. RPC) and schema evolution strategy.
4.3 Pagination/sorting/filtering conventions and determinism.
4.4 Rate limits, quotas, and abuse controls at API boundaries.
4.5 Idempotency expectations for mutations and retriable semantics.
4.6 Webhooks/eventing requirements (retries, signatures, ordering, dead-lettering).

5) Identity, AuthN, AuthZ

5.1 Account creation modes (email/password, passkeys, social, SSO; verification).
5.2 Session management (cookie vs. token; rotation; device/session lists).
5.3 Multi-factor options (TOTP, WebAuthn, SMS) and step-up triggers.
5.4 Authorization model (RBAC/ABAC/tenant-aware policies) and policy placement.
5.5 Account recovery flows (email ownership changes, locked accounts, takeover risk).
5.6 Bot/abuse and fraud considerations (signup friction, velocity, disposable emails).
5.7 Data residency for identity data and cross-border transfer constraints.

6) Data Storage & Modeling

6.1 System of record separation (CMS content vs. transactional user data).
6.2 Relational vs. document vs. key-value for primary workloads.
6.3 Schema evolution, migrations, and backward compatibility.
6.4 Soft deletion, legal holds, and retention schedules.
6.5 Referential integrity across services (IDs, slugs, denormalization strategies).
6.6 Time-travel/version history requirements (who changed what, when).

7) Caching & Performance

7.1 Caching layers (edge/CDN, app cache, client cache) and invalidation.
7.2 Cache keys for personalized vs. public content; Vary rules.
7.3 Stale-while-revalidate vs. hard purge strategies for content changes.
7.4 Query shaping (over-fetch vs. under-fetch; N+1 risk) and batching.
7.5 Media optimization (formats, responsive images, DPR, lazy loading).
7.6 Concurrency limits and head-of-line blocking on hot endpoints.

8) Search & Discovery

8.1 Indexing scope (which fields, which languages, stemming/stopwords).
8.2 Relevance tuning, synonyms, and typo tolerance.
8.3 Faceting/filtering requirements and index update latency.
8.4 Content vs. UGC indexing separation and privacy constraints.
8.5 Zero-results handling and fallback experiences.

9) Media & File Handling

9.1 Storage backend (object storage needs, lifecycle rules, multi-region).
9.2 Upload flows (direct-to-storage, size limits, chunking, resumability).
9.3 Access control (signed URLs, private vs. public, hotlink protection).
9.4 Transform pipeline (images, video transcode, thumbnails) and caching.
9.5 Copyright/DMCA processes and takedown workflows.

10) Editorial Workflow & Backoffice

10.1 Roles and permissions for editors, reviewers, approvers.
10.2 Draft review UX (side-by-side, comments, suggestions).
10.3 Scheduled publishing and mass updates (campaigns, rollbacks).
10.4 Audit trails for content and admin actions.
10.5 Impersonation/safe “view as user” for support teams.
10.6 Operational guardrails (publish freezes, maintenance windows).

11) Personalization & Experimentation

11.1 Audience segmentation sources (profile, behavior, geography).
11.2 Server-side vs. client-side experiments and flicker risk.
11.3 Variant exposure, bucketing, and telemetry attribution.
11.4 Consent boundaries for personalization and profiling.
11.5 Feature flags vs. experiments vs. kill-switches governance.

12) Notifications & Messaging

12.1 Transactional vs. marketing channels separation.
12.2 Deliverability (SPF/DKIM/DMARC), bounce handling, complaint loops.
12.3 User preferences center (per-type, per-channel, locale).
12.4 Throttling/quiet hours and legal compliance (CAN-SPAM, GDPR ePrivacy).
12.5 In-app messages vs. email vs. push; unsubscribe semantics.

13) Security & Privacy

13.1 Threat modeling (auth endpoints, upload endpoints, admin surface).
13.2 Input validation & output encoding; XSS/CSRF/SSRF/IDOR protections.
13.3 Secrets management, key rotation, and least privilege access.
13.4 Content sanitization for UGC (HTML, Markdown, file scanning).
13.5 CSP, Referrer-Policy, Permissions-Policy, and security headers baseline.
13.6 Data minimization, purpose limitation, and DPIA where required.
13.7 Breach detection, incident response runbooks, and disclosure timelines.

14) Compliance & Governance

14.1 Applicable regimes (GDPR/UK-GDPR, CCPA/CPRA, COPPA, HIPAA, PCI, SOC 2).
14.2 Consent capture, proof of consent, and granular purposes.
14.3 Data subject rights execution (access/export/delete; verification).
14.4 Vendor risk management and DPAs.
14.5 Logging/audit retention vs. privacy obligations.
14.6 Content moderation policy and enforcement workflows.

15) Observability & Quality

15.1 Logging strategy (PII scrubbing, structured logs, correlation IDs).
15.2 Metrics and SLOs (availability, latency, error budgets).
15.3 Tracing for cross-service requests and async jobs.
15.4 Real-user monitoring (web vitals) and synthetic tests.
15.5 Error handling UX (graceful fallbacks, user-visible error taxonomies).
15.6 Analytics taxonomy (events, identities, sessionization, attribution windows).

16) Background Jobs & Workflows

16.1 Queueing semantics (at-least-once, exactly-once illusions, ordering).
16.2 Retries, backoff, dead-letter queues, and poison message handling.
16.3 Scheduled jobs (cron) and daylight savings/timezone fun.
16.4 Idempotent job design and deduplication keys.
16.5 Outbox pattern and consistency with external systems.

17) Multi-Tenancy (if applicable)

17.1 Isolation model (row-level, schema-per-tenant, DB-per-tenant).
17.2 Noisy neighbor mitigation (rate limits, resource quotas).
17.3 Tenant-scoped configuration and theming/branding.
17.4 Cross-tenant data leakage prevention and auth boundaries.
17.5 Tenant lifecycle (provisioning, migration, export, deletion).

18) Performance & Capacity Planning

18.1 Read/write ratios and hot-path identification.
18.2 Vertical vs. horizontal scaling plans; auto-scaling signals.
18.3 Cache hit rate targets and miss penalties.
18.4 Cold start and warm-pool strategies for bursty traffic.
18.5 Cost/performance trade-offs and budget guardrails.

19) Deployment, Environments & Release

19.1 Environment strategy (dev/preview/stage/prod) and data parity.
19.2 Rollout mechanics (blue/green, canary, feature-flagged releases).
19.3 Schema and content migrations coordination with deploys.
19.4 Secret rotation and config management across environments.
19.5 Rollback procedures and version pinning for client/server contracts.

20) Testing Strategy

20.1 Unit, contract, integration, E2E balances and priorities.
20.2 Fixture/data seeding for realistic tests and previews.
20.3 Accessibility and visual regression coverage.
20.4 Load testing for auth, search, and content publish spikes.
20.5 Chaos/resilience tests for critical flows (auth, checkout, publish).

21) Backup, DR & Business Continuity

21.1 RPO/RTO goals and tiered recovery (DB, object storage, search index).
21.2 Cross-region replication and failover triggers.
21.3 Encrypted backups, key escrow, and restore drills.
21.4 Content reindex/rebuild pathways (time to full functionality).
21.5 Admin break-glass accounts and emergency playbooks.

22) Cost, FinOps & Vendor Strategy

22.1 Unit economics (per user, per page view, per asset, per job).
22.2 Egress and transform costs for media and search.
22.3 Build-vs-buy posture for CMS, auth, search, email, analytics.
22.4 Vendor lock-in risks and data portability.
22.5 Reserved capacity vs. on-demand and budget alerts.

23) Frontend UX Considerations

23.1 Form resilience (autosave, partial drafts, resume flows).
23.2 Error states and validation feedback (latency masking).
23.3 Offline/read-through behaviors and optimistic UI (if applicable).
23.4 Consent banners, cookie settings, and preference storage.
23.5 Keyboard and screen-reader friendly interactions; reduced motion modes.

24) Mobile & App Surface Area

24.1 Responsive breakpoints and interaction patterns per device class.
24.2 Installable PWA vs. native apps; shared auth tokens and deep links.
24.3 Push notification registration and revocation lifecycle.
24.4 Background sync and bandwidth constraints.
24.5 App store policy implications (if wrapping as native).

25) Moderation & Community Safety (if UGC)

25.1 Pre- vs. post-publication checks and escalation paths.
25.2 Automated signals (toxicity, spam, malware) vs. human review.
25.3 Appeals, transparency reports, and user blocking/reporting UX.
25.4 Jurisdictional content rules (geo-blocking, age gates).
25.5 Evidence preservation for disputes and legal holds.

26) Data Sharing & Integrations

26.1 Outbound data policies (to CRM, email, analytics, ads).
26.2 Inbound integrations (SSO, payments, webhooks) and trust boundaries.
26.3 Event schemas and governance (names, versions, PII classification).
26.4 Partner sandboxing and API keys/scopes/rotations.
26.5 Data export and account deletion propagation to third parties.

27) Team & Process

27.1 Editor vs. developer responsibilities and handoffs.
27.2 Change management for content models (who approves schema changes).
27.3 Runbooks for common incidents (failed publishes, cache stampedes).
27.4 On-call coverage and paging criteria for content vs. auth incidents.
27.5 Documentation standards (ADR, API docs, content model docs).

28) Future-Facing Considerations

28.1 International expansion (languages, currencies, compliance deltas).
28.2 Multi-brand/white-label needs (skinning, routing, data isolation).
28.3 AI-assisted workflows (editor suggestions, semantic search, moderation).
28.4 Sunset/archival policies for aging content and features.
28.5 Telemetry evolution (privacy-preserving analytics, server-side tagging).