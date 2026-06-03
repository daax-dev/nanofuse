package types

import (
	"fmt"
	"path"
	"strings"
)

// NormalizeAndValidateMounts applies defaults (type=bind) and validates the
// operator-supplied mount list. Targets must be absolute and unique; bind and
// volume mounts require a source; tmpfs must not carry a source.
func NormalizeAndValidateMounts(mounts []Mount) ([]Mount, error) {
	if len(mounts) == 0 {
		return nil, nil
	}
	out := make([]Mount, 0, len(mounts))
	seenTargets := make(map[string]struct{}, len(mounts))
	for i := range mounts {
		m := mounts[i]
		m.Source = strings.TrimSpace(m.Source)
		m.Target = strings.TrimSpace(m.Target)
		m.Type = strings.ToLower(strings.TrimSpace(m.Type))
		if m.Type == "" {
			m.Type = MountTypeBind
		}
		switch m.Type {
		case MountTypeBind, MountTypeVolume, MountTypeTmpfs:
		default:
			return nil, fmt.Errorf("mount[%d]: unsupported type %q (want bind|volume|tmpfs)", i, m.Type)
		}
		if m.Target == "" {
			return nil, fmt.Errorf("mount[%d]: target is required", i)
		}
		if !strings.HasPrefix(m.Target, "/") {
			return nil, fmt.Errorf("mount[%d]: target %q must be an absolute path", i, m.Target)
		}
		// Normalize so equivalent paths (/data, /data/, /a//b) dedupe consistently.
		m.Target = path.Clean(m.Target)
		if _, dup := seenTargets[m.Target]; dup {
			return nil, fmt.Errorf("mount[%d]: duplicate target %q", i, m.Target)
		}
		seenTargets[m.Target] = struct{}{}
		switch m.Type {
		case MountTypeBind, MountTypeVolume:
			if m.Source == "" {
				return nil, fmt.Errorf("mount[%d]: source is required for %s mount", i, m.Type)
			}
		case MountTypeTmpfs:
			if m.Source != "" {
				return nil, fmt.Errorf("mount[%d]: tmpfs mount must not set a source", i)
			}
		}
		out = append(out, m)
	}
	return out, nil
}

// NormalizeAndValidateSecrets applies defaults (type=env, target defaults to
// name for env) and validates the operator-supplied secret reference list.
// Secret references never carry values; only names, sources, and targets.
func NormalizeAndValidateSecrets(secrets []SecretRef) ([]SecretRef, error) {
	if len(secrets) == 0 {
		return nil, nil
	}
	out := make([]SecretRef, 0, len(secrets))
	seenNames := make(map[string]struct{}, len(secrets))
	for i := range secrets {
		s := secrets[i]
		s.Name = strings.TrimSpace(s.Name)
		s.Source = strings.TrimSpace(s.Source)
		s.Type = strings.ToLower(strings.TrimSpace(s.Type))
		s.Target = strings.TrimSpace(s.Target)
		if s.Type == "" {
			s.Type = SecretTypeEnv
		}
		if s.Name == "" {
			return nil, fmt.Errorf("secret[%d]: name is required", i)
		}
		if _, dup := seenNames[s.Name]; dup {
			return nil, fmt.Errorf("secret[%d]: duplicate name %q", i, s.Name)
		}
		seenNames[s.Name] = struct{}{}
		switch s.Type {
		case SecretTypeEnv:
			if s.Target == "" {
				s.Target = s.Name
			}
		case SecretTypeFile:
			if s.Target == "" {
				return nil, fmt.Errorf("secret[%d]: file secret %q requires a target path", i, s.Name)
			}
			if !strings.HasPrefix(s.Target, "/") {
				return nil, fmt.Errorf("secret[%d]: file secret target %q must be an absolute path", i, s.Target)
			}
			// Normalize so equivalent paths refer to one stable inventory entry.
			s.Target = path.Clean(s.Target)
		default:
			return nil, fmt.Errorf("secret[%d]: unsupported type %q (want env|file)", i, s.Type)
		}
		out = append(out, s)
	}
	return out, nil
}
