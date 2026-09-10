package operator

import (
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
)

func hasManifestNamed(manifests []manifest, substr string) bool {
	for _, m := range manifests {
		if strings.Contains(m.name, substr) {
			return true
		}
	}
	return false
}

func TestAppendManifestsByPlatformExternal(t *testing.T) {
	cases := []struct {
		name          string
		external      *configv1.ExternalPlatformStatus
		wantCoreDNS   bool
		wantKeepalive bool
	}{
		{
			name: "managed LB with VIPs deploys coredns and keepalived",
			external: &configv1.ExternalPlatformStatus{
				APIServerInternalIPs: []string{"192.168.1.5"},
				IngressIPs:           []string{"192.168.1.6"},
				LoadBalancer:         &configv1.ExternalPlatformLoadBalancer{Type: configv1.LoadBalancerTypeOpenShiftManagedDefault},
			},
			wantCoreDNS:   true,
			wantKeepalive: true,
		},
		{
			name: "user-managed LB with VIPs deploys coredns only",
			external: &configv1.ExternalPlatformStatus{
				APIServerInternalIPs: []string{"192.168.1.5"},
				LoadBalancer:         &configv1.ExternalPlatformLoadBalancer{Type: configv1.LoadBalancerTypeUserManaged},
			},
			wantCoreDNS:   true,
			wantKeepalive: false,
		},
		{
			name:          "empty VIPs deploys nothing",
			external:      &configv1.ExternalPlatformStatus{},
			wantCoreDNS:   false,
			wantKeepalive: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			infra := &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						Type:     configv1.ExternalPlatformType,
						External: tc.external,
					},
				},
			}
			manifests := appendManifestsByPlatform(nil, infra)

			if got := hasManifestNamed(manifests, "coredns"); got != tc.wantCoreDNS {
				t.Errorf("coredns present = %v, want %v", got, tc.wantCoreDNS)
			}
			if got := hasManifestNamed(manifests, "keepalived"); got != tc.wantKeepalive {
				t.Errorf("keepalived present = %v, want %v", got, tc.wantKeepalive)
			}
			// On-prem CoreDNS must be sourced, not the cloud-platform-alt-dns variant.
			if tc.wantCoreDNS && hasManifestNamed(manifests, "cloud-platform-alt-dns") {
				t.Errorf("expected on-prem coredns source, got cloud-platform-alt-dns")
			}
		})
	}
}
