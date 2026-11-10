package service_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-coolify/internal/service"
)

// Coolify's PATCH /applications/{uuid} accepts only nixpacks, static, dockerfile and dockercompose as build_pack.
// A Docker Image app reads back build_pack "dockerimage"; sending it in an update fails with 422 "The selected build
// pack is invalid", so every in-place update of such an app failed.
func TestToAPIUpdateBuildPack(t *testing.T) {
	cases := []struct {
		buildPack string
		wantSent  bool
	}{
		{"dockerimage", false},
		{"dockerfile", true},
		{"nixpacks", true},
		{"static", true},
		{"dockercompose", true},
	}
	for _, c := range cases {
		t.Run(c.buildPack, func(t *testing.T) {
			m := service.ApplicationModel{BuildPack: types.StringValue(c.buildPack)}
			body := m.ToAPIUpdate()
			if c.wantSent {
				if body.BuildPack == nil || string(*body.BuildPack) != c.buildPack {
					t.Fatalf("build_pack %q: want it in the update body, got %v", c.buildPack, body.BuildPack)
				}
			} else if body.BuildPack != nil {
				t.Fatalf("build_pack %q: want it omitted from the update body, got %q", c.buildPack, *body.BuildPack)
			}
		})
	}
}
