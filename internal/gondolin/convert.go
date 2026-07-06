package gondolin

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"unicode"

	"github.com/daax-dev/nanofuse/internal/client"
	"gopkg.in/yaml.v3"
)

// Default resource hints, matching `nanofuse vm create` defaults. Gondolin has
// no resource model, so these are nanofuse assumptions disclosed in the report.
const (
	DefaultVCPUs     = 2
	DefaultMemoryMiB = 512
)

// Severity classifies a divergence between gondolin and nanofuse.
type Severity string

const (
	// SeverityInfo: a disclosed nanofuse assumption (e.g. defaulted resources).
	SeverityInfo Severity = "info"
	// SeverityWarn: a safe degrade that always proceeds. The nanofuse result is
	// at least as restrictive as gondolin (e.g. an unrepresentable allow-list
	// becomes a locked-down default-deny egress policy).
	SeverityWarn Severity = "warn"
	// SeverityBlocking: a gondolin feature with no faithful nanofuse equivalent
	// whose omission could change behaviour in an unsafe or surprising way.
	// Blocking divergences fail the conversion closed unless AllowLossy is set,
	// which downgrades them to warnings and drops the feature loudly.
	SeverityBlocking Severity = "blocking"
)

// Divergence is one reported difference between the gondolin request and the
// nanofuse spec produced for it.
type Divergence struct {
	// Feature is the gondolin flag/field, e.g. "--host-secret".
	Feature  string
	Severity Severity
	// Detail explains what could not be represented and what nanofuse did.
	Detail string
}

// Options controls conversion policy.
type Options struct {
	// AllowLossy downgrades blocking divergences to warnings and proceeds instead
	// of failing closed. The unrepresentable features are not faithfully
	// translated — most are dropped, and a few are mapped best-effort (e.g. --dns
	// sets the coarse AllowDNS toggle) — each reported loudly with a LOSSY marker.
	AllowLossy bool
	// ResolveEgress opts in to resolving literal allow_host hostnames (no
	// wildcards, no paths) to /32 egress rules. Off by default: an L7 HTTP
	// allowlist cannot be faithfully expressed as an L3/L4 CIDR policy, so the
	// default is drop-and-warn rather than a false approximation.
	ResolveEgress bool
	// Resolver resolves a hostname to IPs. Injected for determinism/testing.
	// When nil and ResolveEgress is set, resolution is skipped and reported.
	Resolver func(host string) ([]string, error)
}

// Convert maps a gondolin mirror sandbox to a nanofuse CreateVMRequest and a
// divergence report. It returns a non-nil error when blocking divergences exist
// and opts.AllowLossy is false (fail-closed); in that case the request and the
// divergences are both returned so callers can display the full report
// alongside the error. Input-validation failures (nil sandbox, missing image,
// invalid resources) instead return a nil request and nil/partial divergences
// with the validation error.
func Convert(sb *Sandbox, opts Options) (*client.CreateVMRequest, []Divergence, error) {
	if sb == nil {
		return nil, nil, fmt.Errorf("nil sandbox")
	}

	var divs []Divergence

	// --- image (clean) ---
	image := strings.TrimSpace(sb.Image)
	if image == "" {
		return nil, nil, fmt.Errorf("image is required (gondolin --image)")
	}
	// Reject embedded control/whitespace: such a value is not a valid image
	// reference and would carry a control char into the rendered spec (which
	// preserves newlines), enabling output injection.
	for _, r := range image {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return nil, nil, fmt.Errorf("image %q contains invalid whitespace or control characters", image)
		}
	}

	// --- dns mode: validate before use so a typo/unknown value fails closed
	// rather than silently enabling DNS (especially under --allow-lossy) ---
	if mode := strings.TrimSpace(sb.DNS); mode != "" {
		switch mode {
		case "synthetic", "trusted", "open":
		default:
			return nil, nil, fmt.Errorf("unknown gondolin dns mode %q (expected synthetic, trusted, or open)", mode)
		}
	}

	// --- resources (nanofuse-authored; disclose defaults) ---
	vcpus := DefaultVCPUs
	memory := DefaultMemoryMiB
	switch {
	case sb.Resources == nil:
		divs = append(divs, Divergence{
			Feature:  "resources",
			Severity: SeverityInfo,
			Detail: fmt.Sprintf("gondolin has no CPU/memory model; nanofuse defaults assumed: vcpus=%d, memory_mib=%d",
				DefaultVCPUs, DefaultMemoryMiB),
		})
	default:
		switch {
		case sb.Resources.VCPUs == nil:
			divs = append(divs, Divergence{
				Feature:  "resources.vcpus",
				Severity: SeverityInfo,
				Detail:   fmt.Sprintf("vcpus unset; nanofuse default assumed: %d", DefaultVCPUs),
			})
		case *sb.Resources.VCPUs <= 0:
			return nil, divs, fmt.Errorf("invalid resources: vcpus=%d must be > 0", *sb.Resources.VCPUs)
		default:
			vcpus = *sb.Resources.VCPUs
		}
		switch {
		case sb.Resources.MemoryMiB == nil:
			divs = append(divs, Divergence{
				Feature:  "resources.memory_mib",
				Severity: SeverityInfo,
				Detail:   fmt.Sprintf("memory_mib unset; nanofuse default assumed: %d", DefaultMemoryMiB),
			})
		case *sb.Resources.MemoryMiB <= 0:
			return nil, divs, fmt.Errorf("invalid resources: memory_mib=%d must be > 0", *sb.Resources.MemoryMiB)
		default:
			memory = *sb.Resources.MemoryMiB
		}
	}

	// Normalize allow_host: drop empty/whitespace-only entries so a list that is
	// effectively empty is not treated as "present" (which would emit a spurious
	// default-deny policy and blank entries in the report).
	sb.AllowHost = trimmedNonEmpty(sb.AllowHost)

	// --- egress: allow_host (L7) -> default-deny + warn (safe degrade) ---
	network := client.NetworkConfig{Mode: "nat"}
	if len(sb.AllowHost) > 0 || strings.TrimSpace(sb.DNS) != "" {
		egressDivs, policy := buildEgressPolicy(sb, opts)
		divs = append(divs, egressDivs...)
		network.EgressPolicy = policy
	}

	// Fields with no gondolin equivalent (KernelArgs, Disks, Mounts, Secrets, …)
	// are left at their zero value; the daemon owns their defaults. KernelArgs is
	// omitempty on client.VMConfig, so an unset value is omitted from a request
	// and the daemon applies its default (rather than this converter hardcoding —
	// and drifting from — the daemon's boot args). RenderSpecYAML shows only the
	// conversion-relevant subset.
	req := &client.CreateVMRequest{
		Image: image,
		Config: client.VMConfig{
			VCPUs:     vcpus,
			MemoryMiB: memory,
			Network:   network,
		},
	}

	// --- unrepresentable features -> one distinct divergence each ---
	// Each has no faithful nanofuse equivalent. Blocking by default; dropped
	// loudly under AllowLossy. Never silently translated (avoids false
	// equivalence / false security).
	addBlocking := func(feature string, present bool, detail string) {
		if !present {
			return
		}
		divs = append(divs, Divergence{Feature: feature, Severity: SeverityBlocking, Detail: detail})
	}

	addBlocking("--vmm", strings.TrimSpace(sb.VMM) != "",
		fmt.Sprintf("gondolin vmm %q is qemu/krun; nanofuse runs only firecracker. Hypervisor cannot be honoured.", strings.TrimSpace(sb.VMM)))
	addBlocking("--cwd", strings.TrimSpace(sb.Cwd) != "",
		"nanofuse CreateVMRequest has no guest working-directory field; cwd cannot be set at create time.")
	addBlocking("--env", len(sb.Env) > 0,
		fmt.Sprintf("nanofuse has no guest environment injection; %d env var(s) (%s) cannot be delivered to the guest.",
			len(sb.Env), strings.Join(sortedKeys(sb.Env), ", ")))
	addBlocking("--host-secret", len(sb.HostSecret) > 0,
		fmt.Sprintf("nanofuse has no host-secret injection; %d host-secret entr(y/ies) cannot be delivered "+
			"(entries omitted from this report in case they contain secret values).",
			len(sb.HostSecret)))
	addBlocking("--mount-hostfs", len(sb.MountHostFS) > 0,
		fmt.Sprintf("nanofuse CreateVMRequest exposes no host-directory bind mounts; %d mount(s) (%s) cannot be represented.",
			len(sb.MountHostFS), strings.Join(sb.MountHostFS, ", ")))
	addBlocking("--mount-memfs", len(sb.MountMemFS) > 0,
		fmt.Sprintf("nanofuse has no in-memory mount primitive; %d memfs mount(s) (%s) cannot be represented.",
			len(sb.MountMemFS), strings.Join(sb.MountMemFS, ", ")))
	addBlocking("--ssh-allow-host", len(sb.SSHAllowHost) > 0,
		fmt.Sprintf("nanofuse has no SSH egress broker; %d ssh-allow-host rule(s) (%s) cannot be represented.",
			len(sb.SSHAllowHost), strings.Join(sb.SSHAllowHost, ", ")))
	addBlocking("--tcp-map", len(sb.TCPMap) > 0,
		fmt.Sprintf("nanofuse egress is CIDR-based and has no guest->upstream TCP remapping; %d tcp-map rule(s) (%s) cannot be represented.",
			len(sb.TCPMap), strings.Join(sb.TCPMap, ", ")))
	addBlocking("--rootfs-size", strings.TrimSpace(sb.RootfsSize) != "",
		fmt.Sprintf("nanofuse CreateVMRequest has no rootfs-size control; requested size %q cannot be applied.", strings.TrimSpace(sb.RootfsSize)))

	// Sort divergences for stable, deterministic output (golden tests).
	sortDivergences(divs)

	if opts.AllowLossy {
		// Downgrade blocking -> warn and proceed. The feature is not faithfully
		// translated; some are dropped, some kept only best-effort — so use a
		// neutral marker rather than claiming every one was "dropped".
		for i := range divs {
			if divs[i].Severity == SeverityBlocking {
				divs[i].Severity = SeverityWarn
				divs[i].Detail = "LOSSY (--allow-lossy): " + divs[i].Detail
			}
		}
		return req, divs, nil
	}

	if blocking := blockingFeatures(divs); len(blocking) > 0 {
		return req, divs, fmt.Errorf(
			"fail-closed: %d gondolin feature(s) have no faithful nanofuse equivalent: %s. "+
				"Re-run with --allow-lossy to drop them and proceed",
			len(blocking), strings.Join(blocking, ", "))
	}

	return req, divs, nil
}

// buildEgressPolicy maps gondolin's L7 allow_host list and dns mode onto a
// nanofuse L3/L4 EgressPolicy. Because an HTTP host allowlist cannot be
// faithfully expressed as CIDR rules, the default is a locked-down
// default-deny policy plus a warning (safe degrade). With ResolveEgress and a
// Resolver, literal hostnames (no wildcards/paths) become /32 allow rules, and
// wildcard/path rules are still dropped and reported.
func buildEgressPolicy(sb *Sandbox, opts Options) ([]Divergence, *client.EgressPolicy) {
	var divs []Divergence

	policy := &client.EgressPolicy{
		Enabled:       true,
		DefaultAction: "deny", // nanofuse egress vocabulary is "deny"/"allow"
	}

	if mode := strings.TrimSpace(sb.DNS); mode != "" {
		// nanofuse only has a coarse AllowDNS toggle; the specific gondolin DNS
		// mode (synthetic/trusted/open) and its resolver semantics are lost.
		policy.AllowDNS = true
		divs = append(divs, Divergence{
			Feature:  "--dns",
			Severity: SeverityBlocking,
			Detail: fmt.Sprintf("nanofuse egress has only a coarse allow-DNS toggle; gondolin dns mode %q "+
				"(synthetic/trusted/open resolver semantics) cannot be represented. AllowDNS set true as best effort.", mode),
		})
	}

	if len(sb.AllowHost) == 0 {
		return divs, policy
	}

	dropped := make([]string, 0, len(sb.AllowHost))
	seenCIDR := make(map[string]struct{}) // dedupe rules across resolved IPs
	for _, host := range sb.AllowHost {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if resolveHostToRules(host, opts, policy, seenCIDR) {
			continue
		}
		dropped = append(dropped, host)
	}

	switch {
	case opts.ResolveEgress && opts.Resolver != nil:
		if len(dropped) > 0 {
			divs = append(divs, Divergence{
				Feature:  "--allow-host",
				Severity: SeverityWarn,
				Detail: fmt.Sprintf("L7 HTTP allowlist resolved to point-in-time /32 CIDR rules where possible; "+
					"%d entr(y/ies) not resolvable to a literal host (wildcards/paths/errors) dropped under default-deny: %s. "+
					"Resolved rules are a snapshot and do not track DNS changes.",
					len(dropped), strings.Join(dropped, ", ")),
			})
		}
		if len(policy.AllowRules) > 0 {
			divs = append(divs, Divergence{
				Feature:  "--allow-host (resolved)",
				Severity: SeverityWarn,
				Detail:   "allow-host hostnames resolved to /32 rules restricted to TCP/443; L7 path/method filtering is not enforced.",
			})
		}
	case opts.ResolveEgress && opts.Resolver == nil:
		// Resolution was requested but no resolver is wired; do not tell the user
		// to "use --resolve-egress" (they already did). Report the missing resolver.
		divs = append(divs, Divergence{
			Feature:  "--allow-host",
			Severity: SeverityWarn,
			Detail: fmt.Sprintf("--resolve-egress was requested but no resolver is configured; "+
				"gondolin L7 HTTP allowlist (%s) left under default-deny (no outbound allowed).",
				strings.Join(sb.AllowHost, ", ")),
		})
	default:
		divs = append(divs, Divergence{
			Feature:  "--allow-host",
			Severity: SeverityWarn,
			Detail: fmt.Sprintf("gondolin L7 HTTP allowlist (%s) cannot be expressed as nanofuse L3/L4 CIDR rules; "+
				"egress locked to default-deny (safe degrade, no outbound allowed). "+
				"Use --resolve-egress to opt in to point-in-time hostname->CIDR resolution.",
				strings.Join(sb.AllowHost, ", ")),
		})
	}

	// Resolver output order (e.g. net.LookupHost) is not guaranteed stable across
	// runs/hosts; sort the rules so RenderSpecYAML output is deterministic.
	sort.SliceStable(policy.AllowRules, func(i, j int) bool {
		if policy.AllowRules[i].CIDR != policy.AllowRules[j].CIDR {
			return policy.AllowRules[i].CIDR < policy.AllowRules[j].CIDR
		}
		return policy.AllowRules[i].Port < policy.AllowRules[j].Port
	})

	return divs, policy
}

// isLiteralHost reports whether host is a plain hostname (no wildcard, no path,
// no scheme) that can be resolved to an IP.
// resolveHostToRules attempts to resolve a literal host to deduped egress allow
// rules (HTTPS/443), appending them to policy. It reports true when at least one
// rule was produced (or already covered by an earlier duplicate), meaning the
// host is handled; false means the caller should record it as dropped.
func resolveHostToRules(host string, opts Options, policy *client.EgressPolicy, seenCIDR map[string]struct{}) bool {
	if !opts.ResolveEgress || opts.Resolver == nil || !isLiteralHost(host) {
		return false
	}
	ips, err := opts.Resolver(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	added := false
	for _, ip := range ips {
		cidr, ok := hostCIDR(ip)
		if !ok {
			continue // resolver returned a non-IP string; skip it
		}
		added = true
		if _, dup := seenCIDR[cidr]; dup {
			continue // already covered by an earlier host/IP
		}
		seenCIDR[cidr] = struct{}{}
		policy.AllowRules = append(policy.AllowRules, client.EgressRule{
			CIDR:        cidr,
			Protocol:    "tcp",
			Port:        443,
			Description: "resolved from gondolin allow-host " + host + " (point-in-time; HTTPS/443 only)",
		})
	}
	return added
}

// hostCIDR returns a single-host CIDR for a resolved IP — /32 for IPv4 (incl.
// IPv4-in-IPv6, canonicalized) and /128 for IPv6 — with ok=true. If the string
// is not a parseable IP it returns ("", false) and the caller drops the entry.
func hostCIDR(ip string) (string, bool) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", false // not a valid IP; caller drops it
	}
	// Canonicalize so an IPv4-in-IPv6 form (e.g. ::ffff:1.2.3.4) becomes a plain
	// IPv4 /32 rather than an overly broad IPv6 rule.
	if v4 := parsed.To4(); v4 != nil {
		return v4.String() + "/32", true
	}
	return parsed.String() + "/128", true
}

// trimmedNonEmpty returns the input with each element trimmed and any
// empty-after-trim element dropped.
func trimmedNonEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func isLiteralHost(host string) bool {
	if host == "" {
		return false
	}
	if strings.ContainsAny(host, "*?/ ") {
		return false
	}
	if strings.Contains(host, "://") {
		return false
	}
	// Reject any control character or other whitespace (tab/newline/etc.): such a
	// value is not a valid hostname and, if resolved, would carry the control
	// char into a rule Description and enable terminal line-spoofing.
	for _, r := range host {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func blockingFeatures(divs []Divergence) []string {
	var out []string
	for _, d := range divs {
		if d.Severity == SeverityBlocking {
			out = append(out, d.Feature)
		}
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortDivergences(divs []Divergence) {
	sort.SliceStable(divs, func(i, j int) bool {
		if divs[i].Feature != divs[j].Feature {
			return divs[i].Feature < divs[j].Feature
		}
		return divs[i].Detail < divs[j].Detail
	})
}

// --- rendering -----------------------------------------------------------

// renderSpec is a presentation view of the fields this conversion actually
// populates on client.CreateVMRequest (image, resources, network/egress), using
// stable YAML keys that match the nanofuse schema. It is a human-readable
// preview, not a full API payload: fields the conversion never sets (e.g.
// kernel_args, disks, mounts, secrets) are intentionally omitted so the daemon
// applies its own defaults. Rendered deterministically for display and golden
// tests.
type renderSpec struct {
	Name   string       `yaml:"name,omitempty"`
	Image  string       `yaml:"image"`
	Config renderConfig `yaml:"config"`
}

type renderConfig struct {
	VCPUs     int           `yaml:"vcpus"`
	MemoryMiB int           `yaml:"memory_mib"`
	Network   renderNetwork `yaml:"network"`
}

type renderNetwork struct {
	Mode         string        `yaml:"mode"`
	EgressPolicy *renderEgress `yaml:"egress_policy,omitempty"`
}

type renderEgress struct {
	Enabled       bool         `yaml:"enabled"`
	DefaultAction string       `yaml:"default_action,omitempty"`
	AllowDNS      bool         `yaml:"allow_dns,omitempty"`
	AllowRules    []renderRule `yaml:"allow_rules,omitempty"`
}

type renderRule struct {
	CIDR        string `yaml:"cidr"`
	Protocol    string `yaml:"protocol"`
	Port        int    `yaml:"port"`
	Description string `yaml:"description,omitempty"`
}

// RenderSpecYAML renders the nanofuse spec as deterministic YAML.
func RenderSpecYAML(req *client.CreateVMRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	rs := renderSpec{
		Name:  req.Name,
		Image: req.Image,
		Config: renderConfig{
			VCPUs:     req.Config.VCPUs,
			MemoryMiB: req.Config.MemoryMiB,
			Network: renderNetwork{
				Mode: req.Config.Network.Mode,
			},
		},
	}
	if ep := req.Config.Network.EgressPolicy; ep != nil {
		re := &renderEgress{
			Enabled:       ep.Enabled,
			DefaultAction: ep.DefaultAction,
			AllowDNS:      ep.AllowDNS,
		}
		for _, r := range ep.AllowRules {
			re.AllowRules = append(re.AllowRules, renderRule{
				CIDR:        r.CIDR,
				Protocol:    r.Protocol,
				Port:        r.Port,
				Description: r.Description,
			})
		}
		rs.Config.Network.EgressPolicy = re
	}
	return yaml.Marshal(rs)
}
