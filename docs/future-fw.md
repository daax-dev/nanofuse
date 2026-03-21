# Advanced Firewall Capabilities Reference

**Status:** Reference Documentation
**Related:** [GOALS.md](GOALS.md), [firecracker-runner-design.md](firecracker-runner-design.md)
**Last Updated:** 2025-12-21

---

## Overview

This document provides a comprehensive capability map comparing basic L3/L4 firewall controls with advanced security features. It serves as a reference for Nanofuse network security implementation decisions.

**Organization:** What it protects, what it enforces, who typically owns it, and example commercial + open-source implementations. Includes relevant IETF RFCs where applicable.

---

## 0) Baseline controls (the “basic” stuff)

### 0.1 Stateless L3/L4 ACLs (packet filters)
**Protects against**
- Accidental exposure (open ports), basic segmentation failures
- Coarse control for east/west or north/south

**Enforces**
- src/dst IP/CIDR, protocol (TCP/UDP/ICMP), ports, direction
- No connection tracking: you must allow return traffic explicitly

**Who implements/owns**
- Network team (routers/switches), cloud networking/platform teams

**Commercial / cloud examples**
- AWS VPC Network ACLs (stateless): https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html
- AWS Prescriptive note on NACL statelessness: https://docs.aws.amazon.com/prescriptive-guidance/latest/robust-network-design-control-tower/nacl.html
- Azure NSG (conceptually similar L3/4 filtering; not stateless like NACL but often used as “basic network rules”)
- GCP VPC firewall rules

**Open source / OS**
- Linux nftables/iptables (host-level packet filtering)
- BSD PF (pfctl)

**IETF protocols**
- IP itself (IPv4/IPv6) and transport protocols (TCP/UDP/ICMP) — these are the layers you’re controlling here

---

### 0.2 Stateful L3/L4 firewall (connection tracking)
**Protects against**
- Same as above, plus “return traffic” is automatically handled
- Cleaner policies for client/server patterns

**Enforces**
- 5-tuple + state (NEW/ESTABLISHED/RELATED), basic TCP sanity

**Who implements/owns**
- Network security team (appliances), platform/SRE (host firewalls), cloud platform team (SG equivalents)

**Commercial / cloud examples**
- AWS Security Groups (stateful “virtual firewall”): 
  - https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-groups.html
  - https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-security-groups.html

**Open source / OS**
- nftables connection tracking (conntrack)
- Windows Defender Firewall (host-based, policy-managed)

**IETF protocols**
- Still mostly L3/L4 primitives (IP/TCP/UDP/ICMP)

---

## 1) “Advanced firewall” capability families (the full menu)

> Rule of thumb: once you’re doing *identity + L7 parsing + threat prevention + content inspection*, you’re in “advanced” territory.

---

## A) NGFW (Next-Generation Firewall) — L7-aware inline enforcement
**Protects against**
- Malware delivery, exploit attempts, command-and-control, suspicious apps over “allowed” ports (hello, 443)
- Shadow IT usage and risky SaaS/app behaviors

**Enforces (feature checklist)**
- Application-ID (identify apps regardless of port)
- User/Group awareness (directory / IdP mapping)
- Deep packet inspection (protocol decoding)
- TLS decryption/inspection (with governance + exceptions)
- Integrated IDS/IPS signatures + anomaly checks
- URL/category filtering (often bundled)
- Threat intel: IP/domain reputation, C2 feeds, sinkhole
- File inspection/sandbox detonation (varies by vendor)
- Policy-based routing / SD-WAN integration (common now)
- Centralized policy, logging, reporting

**Who implements/owns**
- Network security engineering / firewall team
- SOC helps tune IPS/threat policies
- Often platform/cloud networking teams for virtual appliances

**Commercial examples**
- Palo Alto Networks NGFW
- Fortinet FortiGate
- Check Point Quantum
- Cisco Secure Firewall (Firepower)

**Open source-ish building blocks**
- Suricata (IDS/IPS engine, often embedded by vendors): https://suricata.io/
- Zeek (network security monitoring; not an IPS but commonly paired): https://zeek.org/
- Snort (IDS/IPS): https://www.snort.org/

**IETF protocols (commonly relevant)**
- TLS 1.3 (encryption you may inspect or at least fingerprint): https://datatracker.ietf.org/doc/html/rfc8446
- QUIC (changes “inspection” game; many controls shift to metadata): https://datatracker.ietf.org/doc/html/rfc9000
- HTTP/2 (common on the wire behind TLS): https://datatracker.ietf.org/doc/html/rfc7540
- HTTP/3 (over QUIC): https://datatracker.ietf.org/doc/html/rfc9114

---

## B) IDS vs IPS (often embedded in NGFW, but can be standalone)
**Protects against**
- Known exploit patterns, suspicious protocol behaviors, scanning/recon
- IPS blocks inline; IDS detects/alerts (out-of-band)

**Enforces**
- Signature-based detection, protocol anomaly detection
- Inline block/quarantine (IPS) or alert-only (IDS)

**Who implements/owns**
- Security engineering / SOC (tuning, alert triage)
- Network team (placement, traffic mirroring / inline path)

**Commercial examples**
- Trellix/FireEye legacy appliance lines (varies by org footprint)
- Cisco / Palo / Check Point IPS modules
- Managed cloud IPS offerings (often inside “network firewall” products)

**Open source**
- Suricata: https://suricata.io/
- Snort: https://www.snort.org/

**IETF protocols**
- Depends on what’s being decoded: DNS/HTTP/TLS/IPsec/etc.

---

## C) WAF (Web Application Firewall) — HTTP/API request protection
**Protects against**
- OWASP-style web attacks (SQLi/XSS/etc.), malicious bots, credential stuffing
- HTTP-layer abuse and “virtual patching” for known vuln patterns

**Enforces (feature checklist)**
- Managed rule sets + custom rules (headers/URI/body)
- Positive security options (allow only known-good patterns)
- Rate limiting / throttling / quotas
- Bot management (signal-based, challenges, fingerprints)
- API protection add-ons (schema validation, JWT validation hooks, etc.)
- Logging/analytics + SIEM export

**Who implements/owns**
- AppSec + platform teams (you need app context)
- Edge team if implemented at CDN/ingress

**Commercial examples**
- Cloudflare WAF
- Akamai App & API Protector
- Imperva WAF
- F5 Advanced WAF / NGINX App Protect
- AWS WAF / Azure WAF / Google Cloud Armor (cloud-native)

**Open source**
- ModSecurity (classic OSS WAF engine): https://github.com/owasp-modsecurity/ModSecurity
- OWASP Core Rule Set (CRS): https://coreruleset.org/
- NGINX (OSS) + ModSecurity module (common pattern)

**IETF protocols**
- HTTP semantics (what you’re filtering):
  - HTTP/1.1: https://datatracker.ietf.org/doc/html/rfc9112
  - HTTP semantics: https://datatracker.ietf.org/doc/html/rfc9110
- TLS 1.3 (most WAFs see HTTPS): https://datatracker.ietf.org/doc/html/rfc8446
- JWT (often used for API auth; sometimes enforced at gateway/WAF layer): https://datatracker.ietf.org/doc/html/rfc7519

---

## D) DDoS protection (volumetric L3/L4 + L7 floods)
**Protects against**
- Bandwidth exhaustion, protocol floods (SYN/UDP reflection), L7 HTTP floods
- The stuff that overwhelms links *before* your firewall gets a vote

**Enforces (feature checklist)**
- Anycast absorption / edge scrubbing
- Auto-detection + mitigation
- Rate-based controls at L7 (often with WAF)
- Upstream BGP-based diversion to scrubbing centers (enterprise setups)

**Who implements/owns**
- Network/security + ISP/CDN/cloud provider (because physics)

**Commercial examples**
- AWS Shield (integrates with WAF for L7): https://docs.aws.amazon.com/waf/latest/developerguide/ddos-overview.html
- Cloudflare DDoS protection (Magic Transit / Spectrum, etc.)
- Akamai Prolexic
- Google Cloud Armor / Azure DDoS Protection

**Open source**
- There’s no real “OSS scrubbing center” that competes with global anycast. OSS helps *detect*:
  - FastNetMon (detection/mitigation automation): https://fastnetmon.com/
  - Zeek (visibility): https://zeek.org/

**IETF protocols (frequently involved)**
- BGP (traffic engineering / diversion):
  - BGP-4: https://datatracker.ietf.org/doc/html/rfc4271
- (Operational patterns like RTBH/Flowspec are BGP-based; confirm your exact method per ISP/vendor)

---

## E) DNS Firewall / Protective DNS (high ROI, often neglected)
**Protects against**
- Malware C2 via DNS, phishing domains, newly registered domains
- DNS tunneling / exfil (some solutions)

**Enforces**
- Domain/category blocking, reputation, sinkholing, logging

**Who implements/owns**
- Security + network (central resolvers) / endpoint team (agent-based DNS)

**Commercial examples**
- Infoblox (DNS security)
- Cisco Umbrella
- Cloudflare Gateway (DNS policies)
- AWS Route 53 Resolver DNS Firewall (cloud-native)

**Open source**
- Unbound (resolver; can be policy-enforced with RPZ patterns): https://nlnetlabs.nl/projects/unbound/about/
- BIND + RPZ (Response Policy Zones): https://kb.isc.org/docs/aa-01251
- Pi-hole (lightweight DNS blocking): https://pi-hole.net/

**IETF protocols**
- DNS over HTTPS (DoH): https://datatracker.ietf.org/doc/html/rfc8484
- DNS over TLS (DoT): https://datatracker.ietf.org/doc/html/rfc7858
- DNS Security Extensions (DNSSEC) architecture:
  - DNSSEC intro/spec set (start here): https://datatracker.ietf.org/doc/html/rfc4033

---

## F) Secure Web Gateway (SWG) — egress control for user web traffic
**Protects against**
- Malicious browsing, risky categories, download controls
- Shadow IT and data leakage via web

**Enforces**
- URL/category policy, TLS inspection, file type controls, inline malware scanning, sometimes DLP

**Who implements/owns**
- Security (often), with IAM/endpoint collaboration

**Commercial examples**
- Zscaler Internet Access (ZIA)
- Netskope SWG
- Palo Alto Prisma Access
- Cloudflare Gateway

**Open source-ish**
- Squid proxy (policy + auth + logging): http://www.squid-cache.org/
- e2guardian (content filtering; niche)

**IETF protocols**
- TLS 1.3: https://datatracker.ietf.org/doc/html/rfc8446
- (Often uses HTTP CONNECT proxying; modern posture is more “identity-aware proxy” than pure proxy)

---

## G) CASB (Cloud Access Security Broker) — SaaS control plane security
**Protects against**
- Unsanctioned SaaS, risky sharing, data exposure inside SaaS

**Enforces**
- App discovery, policy, DLP in SaaS, session controls (depends on product)

**Who implements/owns**
- Security + IAM + governance/risk

**Commercial examples**
- Microsoft Defender for Cloud Apps
- Netskope CASB
- Palo Alto Prisma SaaS
- Cisco Umbrella SIG

**Open source**
- Not really a full CASB equivalent; you can assemble partial coverage with:
  - SaaS audit APIs + SIEM + DLP tooling + IdP conditional access

**IETF protocols**
- OAuth 2.0 (SaaS API authorization): https://datatracker.ietf.org/doc/html/rfc6749
- JWT (tokens commonly encountered): https://datatracker.ietf.org/doc/html/rfc7519

---

## H) ZTNA / Identity-aware access proxy (replaces “VPN to the whole network”)
**Protects against**
- Lateral movement from over-broad network access
- Exposing private apps behind a flat VPN

**Enforces**
- Per-app access, identity + device posture, continuous evaluation
- Often integrates SWG/DLP for “full SASE”

**Who implements/owns**
- Security + IAM, often with network engineering

**Commercial examples**
- Zscaler Private Access (ZPA)
- Cloudflare Access
- Palo Alto Prisma Access (ZTNA flavors)
- Netskope Private Access

**Open source building blocks**
- OAuth2-Proxy (app auth front-door): https://oauth2-proxy.github.io/oauth2-proxy/
- Keycloak (IdP, policy glue): https://www.keycloak.org/
- Pomerium (identity-aware proxy): https://www.pomerium.com/
- Teleport (access proxy for infra): https://goteleport.com/

**IETF protocols**
- OAuth 2.0: https://datatracker.ietf.org/doc/html/rfc6749
- OAuth 2.0 PKCE: https://datatracker.ietf.org/doc/html/rfc7636
- mTLS (implemented via TLS; “mTLS” is a pattern, not a separate RFC): https://datatracker.ietf.org/doc/html/rfc8446

---

## I) VPN / secure tunnels (site-to-site, remote access)
**Protects against**
- Eavesdropping/tampering over untrusted networks
- Provides encrypted transport for network connectivity (not app-layer security)

**Enforces**
- Encryption + peer authentication + routing policy

**Who implements/owns**
- Network/security team; sometimes platform for cloud connectivity

**Commercial examples**
- Most NGFWs include IPsec VPN
- Cloud VPN services (AWS/Azure/GCP)

**Open source**
- strongSwan (IPsec/IKE): https://www.strongswan.org/
- OpenVPN (TLS-based VPN): https://openvpn.net/community/
- WireGuard (widely used; standardization is not purely IETF RFC-based): https://www.wireguard.com/

**IETF protocols**
- IPsec architecture: https://datatracker.ietf.org/doc/html/rfc4301
- IKEv2 (key exchange for IPsec): https://datatracker.ietf.org/doc/html/rfc7296
- TLS 1.3 (for TLS-based VPNs): https://datatracker.ietf.org/doc/html/rfc8446

---

## J) Microsegmentation / Distributed firewall (east-west control at workload level)
**Protects against**
- Lateral movement inside the DC/VPC/VNet/K8s
- Blast radius reduction when (not if) something gets popped

**Enforces**
- Policy at host/VM/container boundary, often label/identity-based
- Sometimes L7 rules (service identity, HTTP methods, etc.)

**Who implements/owns**
- Platform/SRE + security engineering (shared ownership is the only survivable model)

**Commercial examples**
- Illumio
- VMware NSX Distributed Firewall
- Cisco Secure Workload (Tetration lineage)
- Guardicore (Akamai)

**Open source**
- Kubernetes NetworkPolicy engines:
  - Cilium: https://cilium.io/
  - Calico: https://www.tigera.io/project-calico/
- Service mesh policy (L7 identity + mTLS):
  - Istio: https://istio.io/
  - Linkerd: https://linkerd.io/

**IETF protocols**
- TLS 1.3 (mTLS between workloads): https://datatracker.ietf.org/doc/html/rfc8446
- (Service mesh identity standards often rely on X.509/PKI patterns; enforcement is typically above pure IETF protocol definitions)

---

## K) API Gateway security (often overlaps WAF, but is its own control plane)
**Protects against**
- Abusive clients, auth bypass attempts, API-specific threats (schema abuse, request bombs)
- Governance: consistent authn/authz and rate controls

**Enforces**
- Authn (JWT validation), authz hooks, quotas, rate limits, request/response transforms, schema validation

**Who implements/owns**
- Platform engineering + AppSec

**Commercial examples**
- Apigee
- Kong Enterprise
- MuleSoft
- AWS API Gateway (plus WAF), Azure API Management

**Open source**
- Kong OSS: https://konghq.com/
- Envoy Gateway / Envoy: https://www.envoyproxy.io/
- NGINX OSS: https://nginx.org/

**IETF protocols**
- OAuth 2.0: https://datatracker.ietf.org/doc/html/rfc6749
- JWT: https://datatracker.ietf.org/doc/html/rfc7519
- HTTP semantics: https://datatracker.ietf.org/doc/html/rfc9110

---

## L) Certificate automation & trust plumbing (the “boring” bit that powers mTLS everywhere)
**Protects against**
- Manual cert ops errors, expired cert outages, inconsistent trust distribution

**Enforces**
- Automated issuance/renewal/revocation workflows, policy around identities

**Who implements/owns**
- Platform + security PKI team (or whoever owns “the CA headache”)

**Commercial examples**
- Venafi, Keyfactor, DigiCert tooling ecosystems
- Cloud-native certificate managers (managed CA services)

**Open source**
- cert-manager (Kubernetes): https://cert-manager.io/
- step-ca (Smallstep): https://smallstep.com/docs/step-ca/
- HashiCorp Vault PKI (community): https://developer.hashicorp.com/vault/docs/secrets/pki

**IETF protocols**
- ACME (Let’s Encrypt automation protocol): https://datatracker.ietf.org/doc/html/rfc8555
- TLS 1.3: https://datatracker.ietf.org/doc/html/rfc8446

---

# The “who implements what” reality map (quick)
- **Network team**: L3/L4 ACLs, routing, VPN, DDoS provider integration, perimeter/edge placement
- **Firewall/NetSec team**: NGFW policy, TLS inspection governance, IPS tuning, egress control strategy
- **AppSec**: WAF rules, API abuse controls, bot/credential stuffing mitigations, false-positive ownership
- **Platform/SRE**: security groups, host firewalls, K8s NetworkPolicy, service mesh mTLS, cert automation
- **Cloud/CDN/ISP**: real DDoS absorption/scrubbing at scale

---

# Two sanity rules that prevent pain
1) **L3/L4-only controls prevent accidents**. Attackers live at L7 + identity + content.
2) **DDoS is upstream or it’s wishful thinking**. If your link is saturated, your firewall can’t block what it never receives.

If you tell me your target environment (on-prem, AWS/Azure/GCP, k8s yes/no, public APIs yes/no, remote workforce yes/no), I’ll turn this into a reference architecture with:
- enforcement points (where each control sits),
- minimum viable feature set per point,
- “do this first” ordering (ROI vs complexity),
- and the ugly parts to plan for (TLS inspection, false positives, policy sprawl).

