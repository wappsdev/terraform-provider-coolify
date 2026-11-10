package service

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// preserveGitRepository keeps the state value only when it names the same repository as Coolify's answer in another
// form (a URL in state, "owner/repo" from the API). It used to keep the state value for ANY difference, so a stale
// repository in state (kreeva-web's app still read "wappsdev/vaulter-web") never refreshed and gave a diff on every
// plan (10 Nov 2026).
func TestPreserveGitRepository(t *testing.T) {
	tests := []struct {
		name, state, api, want string
	}{
		{"same value", "wappsdev/royco", "wappsdev/royco", "wappsdev/royco"},
		{"https URL in state, same repo", "https://github.com/wappsdev/royco", "wappsdev/royco", "https://github.com/wappsdev/royco"},
		{"https URL with .git, same repo", "https://github.com/wappsdev/royco.git", "wappsdev/royco", "https://github.com/wappsdev/royco.git"},
		{"ssh URL in state, same repo", "git@github.com:wappsdev/royco.git", "wappsdev/royco", "git@github.com:wappsdev/royco.git"},
		{"different repository: take the API", "wappsdev/vaulter-web", "wappsdev/kreeva-web", "wappsdev/kreeva-web"},
		{"URL of a different repository: take the API", "https://github.com/wappsdev/vaulter-web", "wappsdev/kreeva-web", "wappsdev/kreeva-web"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preserveGitRepository(types.StringValue(tt.state), types.StringValue(tt.api))
			if got.ValueString() != tt.want {
				t.Errorf("got %q, want %q", got.ValueString(), tt.want)
			}
		})
	}
	if got := preserveGitRepository(types.StringNull(), types.StringValue("wappsdev/royco")); got.ValueString() != "wappsdev/royco" {
		t.Errorf("null state: got %q", got.ValueString())
	}
}
