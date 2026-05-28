package service

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Tests useStateForUnknownUnlessNull (String). The key regression: when the
// attribute is omitted from config (ConfigValue null) on an UPDATE where the
// framework left PlanValue Unknown, the modifier must pin PlanValue to the prior
// state — otherwise the value stays Unknown after apply, producing
// "Provider returned invalid result object after apply" (vaulter Spec 3 import).
func TestUseStateForUnknownUnlessNullString(t *testing.T) {
	tests := []struct {
		name        string
		config      types.String
		plan        types.String
		state       types.String
		wantPlan    types.String
		wantChanged bool // whether the modifier overrides resp.PlanValue
	}{
		{
			name:        "config null, plan unknown, state known (UPDATE) → pin to state",
			config:      types.StringNull(),
			plan:        types.StringUnknown(),
			state:       types.StringValue("keep-me"),
			wantPlan:    types.StringValue("keep-me"),
			wantChanged: true,
		},
		{
			name:        "config null, plan null, state known (UPDATE) → pin to state",
			config:      types.StringNull(),
			plan:        types.StringNull(),
			state:       types.StringValue("keep-me"),
			wantPlan:    types.StringValue("keep-me"),
			wantChanged: true,
		},
		{
			name:        "config null, state null (CREATE) → unknown",
			config:      types.StringNull(),
			plan:        types.StringNull(),
			state:       types.StringNull(),
			wantPlan:    types.StringUnknown(),
			wantChanged: true,
		},
		{
			name:        "config null, state unknown (CREATE) → unknown",
			config:      types.StringNull(),
			plan:        types.StringUnknown(),
			state:       types.StringUnknown(),
			wantPlan:    types.StringUnknown(),
			wantChanged: true,
		},
		{
			name:        "config set → no-op (respect config, leave plan untouched)",
			config:      types.StringValue("user-set"),
			plan:        types.StringValue("user-set"),
			state:       types.StringValue("old"),
			wantPlan:    types.StringValue("user-set"),
			wantChanged: false,
		},
	}

	m := useStateForUnknownUnlessNull{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				ConfigValue: tt.config,
				PlanValue:   tt.plan,
				StateValue:  tt.state,
			}
			resp := &planmodifier.StringResponse{PlanValue: tt.plan}
			m.PlanModifyString(context.Background(), req, resp)
			if !resp.PlanValue.Equal(tt.wantPlan) {
				t.Errorf("PlanValue = %v, want %v", resp.PlanValue, tt.wantPlan)
			}
		})
	}
}
