package service

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type useStateForUnknownUnlessNull struct{}

func (m useStateForUnknownUnlessNull) Description(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNull) MarkdownDescription(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNull) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Config explicitly sets a value → respect it (normal flow).
	if !req.ConfigValue.IsNull() {
		return
	}

	// Config is null (attribute omitted):
	//   - create (no prior resource state) → mark Unknown (Computed value resolves on apply)
	//   - update (the resource exists) → preserve the prior state value, null included: Coolify answers null for
	//     an unused field (e.g. custom_network_aliases), and an Unknown there gave a diff on every plan.
	//
	// We must pin to state even when the framework left PlanValue Unknown (the
	// common case for Optional+Computed on update). Returning early on an Unknown
	// plan value would leave it Unknown after apply → "Provider returned invalid
	// result object after apply" (surfaced by the vaulter Spec 3 import, where an
	// unrelated forced update left an omitted custom_labels Unknown).
	if req.State.Raw.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	resp.PlanValue = req.StateValue
}

func UseStateForUnknownUnlessNullString() planmodifier.String {
	return useStateForUnknownUnlessNull{}
}

type useStateForUnknownUnlessNullInt64 struct{}

func (m useStateForUnknownUnlessNullInt64) Description(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNullInt64) MarkdownDescription(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNullInt64) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// See PlanModifyString for rationale: pin to state on update even when the
	// framework left PlanValue Unknown, else an omitted Optional+Computed field
	// stays Unknown after a forced update → "invalid result object after apply".
	if !req.ConfigValue.IsNull() {
		return
	}

	if req.State.Raw.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.Int64Unknown()
		return
	}

	resp.PlanValue = req.StateValue
}

func UseStateForUnknownUnlessNullInt64() planmodifier.Int64 {
	return useStateForUnknownUnlessNullInt64{}
}

type useStateForUnknownUnlessNullBool struct{}

func (m useStateForUnknownUnlessNullBool) Description(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNullBool) MarkdownDescription(ctx context.Context) string {
	return "Handles Optional+Computed fields: Unknown on create when omitted, prior state (null included) on update"
}

func (m useStateForUnknownUnlessNullBool) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// See PlanModifyString for rationale: pin to state on update even when the
	// framework left PlanValue Unknown, else an omitted Optional+Computed field
	// stays Unknown after a forced update → "invalid result object after apply".
	if !req.ConfigValue.IsNull() {
		return
	}

	if req.State.Raw.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.BoolUnknown()
		return
	}

	resp.PlanValue = req.StateValue
}

func UseStateForUnknownUnlessNullBool() planmodifier.Bool {
	return useStateForUnknownUnlessNullBool{}
}
