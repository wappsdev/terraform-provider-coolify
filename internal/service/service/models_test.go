package service_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-coolify/internal/api"
	service "terraform-provider-coolify/internal/service/service"
)

// Coolify GET /services/{uuid} environment_uuid ve destination_uuid dondurmez.
// Config'te verilmeyen bu Optional+Computed alanlar create plan'inda Unknown
// gelir; apply sonrasi state'e Unknown yazilirsa Terraform "Provider returned
// invalid result object after apply" hatasi verir ve kaynak tainted olur.
func TestServiceFromAPI_resolvesUnknownComputedToNull(t *testing.T) {
	uuid := "svc123"
	name := "teamspeak"
	plan := service.ServiceModel{
		ServerUuid:      types.StringValue("srv"),
		ProjectUuid:     types.StringValue("proj"),
		EnvironmentName: types.StringValue("production"),
		EnvironmentUuid: types.StringUnknown(),
		DestinationUuid: types.StringUnknown(),
		InstantDeploy:   types.BoolValue(true),
		Compose:         types.StringValue("services: {}"),
	}

	got := service.ServiceModel{}.FromAPI(&api.Service{Uuid: &uuid, Name: &name}, plan)

	if !got.EnvironmentUuid.IsNull() {
		t.Errorf("environment_uuid: want null, got %s", got.EnvironmentUuid)
	}
	if !got.DestinationUuid.IsNull() {
		t.Errorf("destination_uuid: want null, got %s", got.DestinationUuid)
	}
}

// Coolify's PATCH /services/{uuid} rejects destination_uuid with 422
// "This field is not allowed" — it is create-only, like server/project/env.
func TestServiceToAPIUpdate_omitsDestinationUuid(t *testing.T) {
	m := service.ServiceModel{
		DestinationUuid: types.StringValue("dest"),
		Compose:         types.StringValue("services: {}"),
	}

	got := m.ToAPIUpdate()

	if got.DestinationUuid != nil {
		t.Errorf("destination_uuid must not be sent on update, got %q", *got.DestinationUuid)
	}
}

func TestServiceFromAPI_keepsConfiguredUuids(t *testing.T) {
	uuid := "svc123"
	plan := service.ServiceModel{
		EnvironmentUuid: types.StringValue("env"),
		DestinationUuid: types.StringValue("dest"),
	}

	got := service.ServiceModel{}.FromAPI(&api.Service{Uuid: &uuid}, plan)

	if got.EnvironmentUuid.ValueString() != "env" {
		t.Errorf("environment_uuid: want env, got %s", got.EnvironmentUuid)
	}
	if got.DestinationUuid.ValueString() != "dest" {
		t.Errorf("destination_uuid: want dest, got %s", got.DestinationUuid)
	}
}
