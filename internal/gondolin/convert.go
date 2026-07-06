package gondolin

import (
	"fmt"
	"net"
	"sort"
	"strings"

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
	// AllowLossy downgrades blocking divergences to warnings and proceeds,
	// dropping the unrepresentable feature loudly instead of failing closed.
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
// and opts.AllowLossy is false (fail-closed). The request and divergences are
// always returned so callers can display the full report alongside the error.
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
		if sb.Resources.VCPUs > 0 {
			vcpus = sb.Resources.VCPUs
		} else {
			divs = append(divs, Divergence{
				Feature:  "resources.vcpus",
				Severity: SeverityInfo,
				Detail:   fmt.Sprintf("vcpus unset; nanofuse default assumed: %d", DefaultVCPUs),
			})
		}
		if sb.Resources.MemoryMiB > 0 {
			memory = sb.Resources.MemoryMiB
		} else {
			divs = append(divs, Divergence{
				Feature:  "resources.memory_mib",
				Severity: SeverityInfo,
				Detail:   fmt.Sprintf("memory_mib unset; nanofuse default assumed: %d", DefaultMemoryMiB),
			})
		}
	}
	if vcpus <= 0 || memory <= 0 {
		return nil, divs, fmt.Errorf("invalid resources: vcpus=%d memory_mib=%d must be > 0", vcpus, memory)
	}

	// --- egress: allow_host (L7) -> default-deny + warn (safe degrade) ---
	network := client.NetworkConfig{Mode: "nat"}
	if len(sb.AllowHost) > 0 || strings.TrimSpace(sb.DNS) != "" {
		egressDivs, policy := buildEgressPolicy(sb, opts)
		divs = append(divs, egressDivs...)
		network.EgressPolicy = policy
	}

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
		fmt.Sprintf("nanofuse has no host-secret injection; %d secret(s) (%s) cannot be delivered.",
			len(sb.HostSecret), strings.Join(sb.HostSecret, ", ")))
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
	for _, host := range sb.AllowHost {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if opts.ResolveEgress && opts.Resolver != nil && isLiteralHost(host) {
			ips, err := opts.Resolver(host)
			if err == nil && len(ips) > 0 {
				added := false
				for _, ip := range ips {
					cidr, ok := hostCIDR(ip)
					if !ok {
						continue // resolver returned a non-IP string; skip it
					}
					policy.AllowRules = append(policy.AllowRules, client.EgressRule{
						CIDR:        cidr,
						Protocol:    "tcp",
						Port:        443,
						Description: "resolved from gondolin allow-host " + host + " (point-in-time; HTTPS/443 only)",
					})
					added = true
				}
				if added {
					continue
				}
			}
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

	return divs, policy
}

// isLiteralHost reports whether host is a plain hostname (no wildcard, no path,
// no scheme) that can be resolved to an IP.
// hostCIDR returns a single-host CIDR for a resolved IP: /32 for IPv4, /128 for
// IPv6. Falls back to /32 only if the string is not a parseable IP (defensive;
// resolver output is expected to be valid).
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

// renderSpec is a presentation view of client.CreateVMRequest with stable YAML
// keys mirroring the nanofuse JSON schema. Rendered deterministically for
// display and golden tests.
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
