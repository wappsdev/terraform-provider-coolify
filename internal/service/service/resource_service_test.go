package service_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"terraform-provider-coolify/internal/acctest"
	service "terraform-provider-coolify/internal/service/service"
)

func TestServiceResource_implementsImportState(t *testing.T) {
	rs := service.NewServiceResource()
	if _, ok := rs.(tfresource.ResourceWithImportState); !ok {
		t.Error("ServiceResource must implement ResourceWithImportState")
	}
}

func TestServiceValidateCreatePlan_okEnvNameOnly(t *testing.T) {
	plan := service.ServiceModel{
		EnvironmentName: types.StringValue("production"),
		Compose:         types.StringValue("services:\n  app:\n    image: nginx"),
	}
	diags := service.ValidateCreatePlan(plan)
	if diags.HasError() {
		t.Errorf("expected no error, got: %s", diags.Errors())
	}
}

func TestServiceValidateCreatePlan_okEnvUuidOnly(t *testing.T) {
	plan := service.ServiceModel{
		EnvironmentUuid: types.StringValue("envuuid123"),
		Compose:         types.StringValue("services:\n  app:\n    image: nginx"),
	}
	diags := service.ValidateCreatePlan(plan)
	if diags.HasError() {
		t.Errorf("expected no error, got: %s", diags.Errors())
	}
}

func TestServiceValidateCreatePlan_failsEnvNeither(t *testing.T) {
	plan := service.ServiceModel{
		Compose: types.StringValue("services:\n  app:\n    image: nginx"),
	}
	diags := service.ValidateCreatePlan(plan)
	if !diags.HasError() {
		t.Error("expected error for missing env_name and env_uuid")
	}
}

func TestServiceValidateCreatePlan_failsComposeEmpty(t *testing.T) {
	plan := service.ServiceModel{
		EnvironmentName: types.StringValue("production"),
		Compose:         types.StringValue(""),
	}
	diags := service.ValidateCreatePlan(plan)
	if !diags.HasError() {
		t.Error("expected error for empty compose")
	}
}

func TestServiceValidateCreatePlan_failsComposeNull(t *testing.T) {
	plan := service.ServiceModel{
		EnvironmentName: types.StringValue("production"),
	}
	diags := service.ValidateCreatePlan(plan)
	if !diags.HasError() {
		t.Error("expected error for null compose")
	}
}

func TestServiceResourceSchema_descriptionHasPlanModifier(t *testing.T) {
	ctx := context.Background()
	rs := service.NewServiceResource()

	req := tfresource.SchemaRequest{}
	resp := &tfresource.SchemaResponse{}
	rs.Schema(ctx, req, resp)

	attr, ok := resp.Schema.Attributes["description"].(schema.StringAttribute)
	if !ok {
		t.Fatal("description attribute missing or wrong type")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("description should have a PlanModifier (UseStateForUnknown) to prevent drift on Optional+Computed")
	}
}

func TestServiceResourceSchema_hasConnectToDockerNetwork(t *testing.T) {
	ctx := context.Background()
	rs := service.NewServiceResource()

	req := tfresource.SchemaRequest{}
	resp := &tfresource.SchemaResponse{}
	rs.Schema(ctx, req, resp)

	attr, ok := resp.Schema.Attributes["connect_to_docker_network"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("connect_to_docker_network attribute missing or wrong type")
	}
	if !attr.Optional {
		t.Error("connect_to_docker_network should be Optional")
	}
	if !attr.Computed {
		t.Error("connect_to_docker_network should be Computed")
	}
}

func TestServiceResourceSchema_environmentUuidComputedWithModifier(t *testing.T) {
	ctx := context.Background()
	rs := service.NewServiceResource()

	req := tfresource.SchemaRequest{}
	resp := &tfresource.SchemaResponse{}
	rs.Schema(ctx, req, resp)

	attr, ok := resp.Schema.Attributes["environment_uuid"].(schema.StringAttribute)
	if !ok {
		t.Fatal("environment_uuid attribute missing or wrong type")
	}
	if !attr.Computed {
		t.Error("environment_uuid should be Computed (Coolify Read returns it)")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("environment_uuid should have UseStateForUnknown PlanModifier")
	}
}

func TestServiceResourceSchema_destinationUuidNoStaticDefault(t *testing.T) {
	ctx := context.Background()
	rs := service.NewServiceResource()

	req := tfresource.SchemaRequest{}
	resp := &tfresource.SchemaResponse{}
	rs.Schema(ctx, req, resp)

	attr, ok := resp.Schema.Attributes["destination_uuid"].(schema.StringAttribute)
	if !ok {
		t.Fatal("destination_uuid attribute missing or wrong type")
	}
	if attr.Default != nil {
		t.Error("destination_uuid should not have a static Default (causes empty-string drift); use UseStateForUnknown modifier instead")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("destination_uuid should have a UseStateForUnknown PlanModifier")
	}
}

func TestServiceResourceSchema_nameHasPlanModifier(t *testing.T) {
	ctx := context.Background()
	rs := service.NewServiceResource()

	req := tfresource.SchemaRequest{}
	resp := &tfresource.SchemaResponse{}
	rs.Schema(ctx, req, resp)

	attr, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("name attribute missing or wrong type")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("name should have a PlanModifier (UseStateForUnknown) to prevent drift on Optional+Computed")
	}
}

func TestAccServiceResource(t *testing.T) {
	resName := "coolify_service.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // Create and Read testing
				Config: `
				resource "coolify_service" "test" {
					name        = "TerraformAccTest"
					description = "Terraform acceptance testing"

					server_uuid = "` + acctest.ServerUUID + `"
					project_uuid = "` + acctest.ProjectUUID + `"
					environment_name = "` + acctest.EnvironmentName + `"
					destination_uuid = "` + acctest.DestinationUUID + `"

					instant_deploy = false
  				compose = <<EOF
services:
  whoami:
    image: "containous/whoami"
    container_name: "simple-service"
EOF
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resName, "description", "Terraform acceptance testing"),
					resource.TestCheckResourceAttr(resName, "server_uuid", acctest.ServerUUID),
					resource.TestCheckResourceAttr(resName, "project_uuid", acctest.ProjectUUID),
					resource.TestCheckResourceAttr(resName, "environment_name", acctest.EnvironmentName),
					resource.TestCheckResourceAttr(resName, "instant_deploy", "false"),

					resource.TestCheckResourceAttrSet(resName, "uuid"),
					resource.TestCheckResourceAttrSet(resName, "compose"),
				),
			},
			{ // ImportState testing
				ResourceName:                         resName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ExpectError: regexp.MustCompile(
					`("instant_deploy")`,
				),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r := s.RootModule().Resources[resName].Primary.Attributes
					return fmt.Sprintf("%s/%s/%s/%s",
						r["server_uuid"],
						r["project_uuid"],
						r["environment_name"],
						r["uuid"],
					), nil
				},
			},
			{ // Update and Read testing
				Config: `
				resource "coolify_service" "test" {
					name        = "TerraformAccTestUpdated"
					description = "Terraform acceptance testing"

					server_uuid = "` + acctest.ServerUUID + `"
					project_uuid = "` + acctest.ProjectUUID + `"
					environment_name = "` + acctest.EnvironmentName + `"
					destination_uuid = "` + acctest.DestinationUUID + `"

					instant_deploy = false

  				compose = <<EOF
services:
  whoami2:
    image: "containous/whoami"
    container_name: "simple-service"
EOF

				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resName, tfjsonpath.New("name"), knownvalue.StringExact("TerraformAccTestUpdated")),
						plancheck.ExpectKnownValue(resName, tfjsonpath.New("description"), knownvalue.StringExact("Terraform acceptance testing")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "uuid"),
					resource.TestCheckResourceAttr(resName, "name", "TerraformAccTestUpdated"),
					resource.TestCheckResourceAttr(resName, "description", "Terraform acceptance testing"),
					resource.TestCheckResourceAttr(resName, "server_uuid", acctest.ServerUUID),
				),
			},
		},
	})
}
