package alb_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stackitcloud/stackit-sdk-go/services/alb/wait"
	"github.com/stackitcloud/terraform-provider-stackit/stackit/internal/core"

	stackitSdkConfig "github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/core/utils"
	"github.com/stackitcloud/stackit-sdk-go/services/alb"
	"github.com/stackitcloud/terraform-provider-stackit/stackit/internal/testutil"
)

//go:embed testfiles/resource-min.tf
var resourceMinConfig string

//go:embed testfiles/resource-max.tf
var resourceMaxConfig string

var testConfigVarsMin = config.Variables{
	"project_id":          config.StringVariable(testutil.ProjectId),
	"region":              config.StringVariable(testutil.Region),
	"loadbalancer_name":   config.StringVariable(fmt.Sprintf("tf-acc-l%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"network_role":        config.StringVariable("ROLE_LISTENERS_AND_TARGETS"),
	"network_name":        config.StringVariable(fmt.Sprintf("tf-acc-n%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"plan_id":             config.StringVariable("p10"),
	"listener_port":       config.StringVariable("5432"),
	"host":                config.StringVariable("*"),
	"path_prefix":         config.StringVariable("/"),
	"protocol_http":       config.StringVariable("PROTOCOL_HTTP"),
	"target_pool_name":    config.StringVariable("my-target-pool"),
	"target_pool_port":    config.StringVariable("5432"),
	"target_display_name": config.StringVariable("my-target"),
}

var testConfigVarsMax = config.Variables{
	"project_id":                        config.StringVariable(testutil.ProjectId),
	"plan_id":                           config.StringVariable("p750"),
	"disable_security_group_assignment": config.BoolVariable(true),
	"network_name_listener":             config.StringVariable(fmt.Sprintf("tf-acc-l%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"network_name_target":               config.StringVariable(fmt.Sprintf("tf-acc-t%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"loadbalancer_name":                 config.StringVariable(fmt.Sprintf("tf-acc-l%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"target_port":                       config.StringVariable("443"),
	"listener_port":                     config.StringVariable("443"),
	"target_display_name":               config.StringVariable("example-target"),
	"ephemeral_address":                 config.BoolVariable(true),
	"private_network_only":              config.StringVariable("false"),
	"acl":                               config.StringVariable("192.168.0.0/24"),
	"observability_logs_push_url":       config.StringVariable("https://logs.observability.dummy.stackit.cloud"),
	"observability_metrics_push_url":    config.StringVariable("https://metrics.observability.dummy.stackit.cloud"),
	"observability_credential_name":     config.StringVariable(fmt.Sprintf("tf-acc-l%s", acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum))),
	"observability_credential_username": config.StringVariable("obs-cred-username"),
	"observability_credential_password": config.StringVariable("obs-cred-password"),
}

func configVarsMinUpdated() config.Variables {
	tempConfig := make(config.Variables, len(testConfigVarsMin))
	maps.Copy(tempConfig, testConfigVarsMin)
	tempConfig["target_pool_port"] = config.StringVariable("5431")
	return tempConfig
}

func configVarsMaxUpdated() config.Variables {
	tempConfig := make(config.Variables, len(testConfigVarsMax))
	maps.Copy(tempConfig, testConfigVarsMax)
	return tempConfig
}

func TestAccALBResourceMin(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckALBDestroy,
		Steps: []resource.TestStep{
			// Creation
			{
				ConfigVariables: testConfigVarsMin,
				Config:          testutil.ALBProviderConfig() + resourceMinConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Load balancer instance resource
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMin["project_id"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "region", testutil.ConvertConfigVariable(testConfigVarsMin["region"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMin["loadbalancer_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "networks.0.role", testutil.ConvertConfigVariable(testConfigVarsMin["network_role"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "plan_id", testutil.ConvertConfigVariable(testConfigVarsMin["plan_id"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.0.port", testutil.ConvertConfigVariable(testConfigVarsMin["listener_port"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.0.http.hosts.0.host", testutil.ConvertConfigVariable(testConfigVarsMin["host"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.0.http.hosts.0.rules.0.path.prefix", testutil.ConvertConfigVariable(testConfigVarsMin["path_prefix"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.0.http.hosts.0.rules.0.target_pool", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.0.protocol", testutil.ConvertConfigVariable(testConfigVarsMin["protocol_http"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.0.name", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.0.target_port", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_port"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.0.targets.0.display_name", testutil.ConvertConfigVariable(testConfigVarsMin["target_display_name"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "networks.0.network_id"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_pools.0.targets.0.ip"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "external_address"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_security_group.id"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_security_group.name"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "load_balancer_security_group.id"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "load_balancer_security_group.name"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "version"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "status"),
					resource.TestCheckNoResourceAttr("stackit_alb.loadbalancer", "disable_security_group_assignment"),
					resource.TestCheckNoResourceAttr("stackit_alb.loadbalancer", "options"),
					resource.TestCheckNoResourceAttr("stackit_alb.loadbalancer", "labels"),
				),
			},
			// Data source
			{
				ConfigVariables: testConfigVarsMin,
				Config: fmt.Sprintf(`
						%s

						data "stackit_alb" "loadbalancer" {
							project_id     = stackit_alb.loadbalancer.project_id
							name    = stackit_alb.loadbalancer.name
						}
						`,
					testutil.ALBProviderConfig()+resourceMinConfig,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Load balancer instance
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMin["project_id"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "region", testutil.ConvertConfigVariable(testConfigVarsMin["region"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMin["loadbalancer_name"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "networks.0.role", testutil.ConvertConfigVariable(testConfigVarsMin["network_role"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "plan_id", testutil.ConvertConfigVariable(testConfigVarsMin["plan_id"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "listeners.0.port", testutil.ConvertConfigVariable(testConfigVarsMin["listener_port"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "listeners.0.http.hosts.0.host", testutil.ConvertConfigVariable(testConfigVarsMin["host"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "listeners.0.http.hosts.0.rules.0.path.prefix", testutil.ConvertConfigVariable(testConfigVarsMin["path_prefix"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "listeners.0.http.hosts.0.rules.0.target_pool", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_name"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "listeners.0.protocol", testutil.ConvertConfigVariable(testConfigVarsMin["protocol_http"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "target_pools.0.name", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_name"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "target_pools.0.target_port", testutil.ConvertConfigVariable(testConfigVarsMin["target_pool_port"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "target_pools.0.targets.0.display_name", testutil.ConvertConfigVariable(testConfigVarsMin["target_display_name"])),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "networks.0.network_id"),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "target_pools.0.targets.0.ip"),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "external_address"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_security_group.id"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_security_group.name"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "load_balancer_security_group.id"),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "load_balancer_security_group.name"),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "version"),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "status"),
					resource.TestCheckNoResourceAttr("data.stackit_alb.loadbalancer", "disable_security_group_assignment"),
					resource.TestCheckNoResourceAttr("data.stackit_alb.loadbalancer", "options"),
					resource.TestCheckNoResourceAttr("data.stackit_alb.loadbalancer", "labels"),
					resource.TestCheckNoResourceAttr("data.stackit_alb.loadbalancer", "errors"),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "project_id",
						"stackit_alb.loadbalancer", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "region",
						"stackit_alb.loadbalancer", "region",
					),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "name",
						"stackit_alb.loadbalancer", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "plan_id",
						"stackit_alb.loadbalancer", "plan_id",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "external_address",
						"data.stackit_alb.loadbalancer", "external_address",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "target_security_group.id",
						"data.stackit_alb.loadbalancer", "target_security_group.id",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "target_security_group.name",
						"data.stackit_alb.loadbalancer", "target_security_group.name",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "load_balancer_security_group.id",
						"data.stackit_alb.loadbalancer", "load_balancer_security_group.id",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "load_balancer_security_group.name",
						"data.stackit_alb.loadbalancer", "load_balancer_security_group.name",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "version",
						"data.stackit_alb.loadbalancer", "version",
					),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "status",
						"data.stackit_alb.loadbalancer", "status",
					),
				)},
			// Import
			{
				ConfigVariables: testConfigVarsMin,
				ResourceName:    "stackit_alb.loadbalancer",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r, ok := s.RootModule().Resources["stackit_alb.loadbalancer"]
					if !ok {
						return "", fmt.Errorf("couldn't find resource stackit_alb.loadbalancer")
					}
					name, ok := r.Primary.Attributes["name"]
					if !ok {
						return "", fmt.Errorf("couldn't find attribute name")
					}
					region, ok := r.Primary.Attributes["region"]
					if !ok {
						return "", fmt.Errorf("couldn't find attribute region")
					}
					return fmt.Sprintf("%s,%s,%s", testutil.ProjectId, region, name), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				ConfigVariables: configVarsMinUpdated(),
				Config:          testutil.ALBProviderConfig() + resourceMinConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMin["project_id"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMin["loadbalancer_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.0.target_port", testutil.ConvertConfigVariable(configVarsMinUpdated()["target_pool_port"])),
				),
			},
			// Deletion is done by the framework implicitly
		},
	})
}

func TestAccALBResourceMax(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckALBDestroy,
		Steps: []resource.TestStep{
			// Creation
			{
				ConfigVariables: testConfigVarsMax,
				Config:          testutil.ALBProviderConfig() + resourceMaxConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Load balancer instance resource
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMax["project_id"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMax["loadbalancer_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "plan_id", testutil.ConvertConfigVariable(testConfigVarsMax["plan_id"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "networks.0.network_id"),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "networks.0.role", testutil.ConvertConfigVariable(testConfigVarsMax["network_role"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "external_address"),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "disable_security_group_assignment", testutil.ConvertConfigVariable(testConfigVarsMax["disable_security_group_assignment"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "security_group_id"),

					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.0.targets.0.display_name", testutil.ConvertConfigVariable(testConfigVarsMax["target_display_name"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_pools.0.targets.0.ip"),

					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.display_name", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_display_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.port", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_port"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.protocol", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_protocol"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.target_pool", testutil.ConvertConfigVariable(testConfigVarsMax["udp_target_pool_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.udp.idle_timeout", testutil.ConvertConfigVariable(testConfigVarsMax["udp_idle_timeout"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.1.name", testutil.ConvertConfigVariable(testConfigVarsMax["udp_target_pool_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.1.target_port", testutil.ConvertConfigVariable(testConfigVarsMax["udp_target_port"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "target_pools.1.targets.0.display_name", testutil.ConvertConfigVariable(testConfigVarsMax["target_display_name"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "target_pools.1.targets.0.ip"),

					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "options.private_network_only", testutil.ConvertConfigVariable(testConfigVarsMax["private_network_only"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "options.acl.0", testutil.ConvertConfigVariable(testConfigVarsMax["acl"])),

					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "options.observability.logs.credentials_ref"),
					resource.TestCheckResourceAttrPair("stackit_loadbalancer_observability_credential.logs", "credentials_ref", "stackit_alb.loadbalancer", "options.observability.logs.credentials_ref"),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "options.observability.logs.push_url", testutil.ConvertConfigVariable(testConfigVarsMax["observability_logs_push_url"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "options.observability.metrics.credentials_ref"),
					resource.TestCheckResourceAttrPair("stackit_loadbalancer_observability_credential.metrics", "credentials_ref", "stackit_alb.loadbalancer", "options.observability.metrics.credentials_ref"),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "options.observability.metrics.push_url", testutil.ConvertConfigVariable(testConfigVarsMax["observability_metrics_push_url"])),

					// Loadbalancer observability credential resource
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.logs", "project_id", testutil.ConvertConfigVariable(testConfigVarsMin["project_id"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.logs", "display_name", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_logs_name"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.logs", "username", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_logs_username"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.logs", "password", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_logs_password"])),
					resource.TestCheckResourceAttrSet("stackit_loadbalancer_observability_credential.logs", "credentials_ref"),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.metrics", "project_id", testutil.ConvertConfigVariable(testConfigVarsMin["project_id"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.metrics", "display_name", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_metrics_name"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.metrics", "username", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_metrics_username"])),
					resource.TestCheckResourceAttr("stackit_loadbalancer_observability_credential.metrics", "password", testutil.ConvertConfigVariable(testConfigVarsMax["observability_credential_metrics_password"])),
					resource.TestCheckResourceAttrSet("stackit_loadbalancer_observability_credential.metrics", "credentials_ref"),
				),
			},
			// Data source
			{
				ConfigVariables: testConfigVarsMax,
				Config: fmt.Sprintf(`
						%s

						data "stackit_loadbalancer" "loadbalancer" {
							project_id     = stackit_alb.loadbalancer.project_id
							name    = stackit_alb.loadbalancer.name
						}
						`,
					testutil.ALBProviderConfig()+resourceMaxConfig,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Load balancer instance
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMax["project_id"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMax["loadbalancer_name"])),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "plan_id", testutil.ConvertConfigVariable(testConfigVarsMax["plan_id"])),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "project_id",
						"stackit_alb.loadbalancer", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.stackit_alb.loadbalancer", "name",
						"stackit_alb.loadbalancer", "name",
					),
					// Load balancer instance
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "networks.0.network_id"),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "networks.0.role", testutil.ConvertConfigVariable(testConfigVarsMax["network_role"])),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "external_address"),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "disable_security_group_assignment", testutil.ConvertConfigVariable(testConfigVarsMax["disable_security_group_assignment"])),
					resource.TestCheckResourceAttrSet("stackit_alb.loadbalancer", "security_group_id"),

					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.display_name", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_display_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.port", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_port"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.protocol", testutil.ConvertConfigVariable(testConfigVarsMax["udp_listener_protocol"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.target_pool", testutil.ConvertConfigVariable(testConfigVarsMax["udp_target_pool_name"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "listeners.1.udp.idle_timeout", testutil.ConvertConfigVariable(testConfigVarsMax["udp_idle_timeout"])),

					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "options.acl.0", testutil.ConvertConfigVariable(testConfigVarsMax["acl"])),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "options.observability.logs.credentials_ref"),
					resource.TestCheckResourceAttrPair("stackit_loadbalancer_observability_credential.logs", "credentials_ref", "data.stackit_alb.loadbalancer", "options.observability.logs.credentials_ref"),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "options.observability.logs.push_url", testutil.ConvertConfigVariable(testConfigVarsMax["observability_logs_push_url"])),
					resource.TestCheckResourceAttrSet("data.stackit_alb.loadbalancer", "options.observability.metrics.credentials_ref"),
					resource.TestCheckResourceAttrPair("stackit_loadbalancer_observability_credential.metrics", "credentials_ref", "data.stackit_alb.loadbalancer", "options.observability.metrics.credentials_ref"),
					resource.TestCheckResourceAttr("data.stackit_alb.loadbalancer", "options.observability.metrics.push_url", testutil.ConvertConfigVariable(testConfigVarsMax["observability_metrics_push_url"])),
					resource.TestCheckResourceAttrPair(
						"stackit_alb.loadbalancer", "security_group_id",
						"data.stackit_alb.loadbalancer", "security_group_id",
					),
				)},
			// Import
			{
				ConfigVariables: testConfigVarsMax,
				ResourceName:    "stackit_alb.loadbalancer",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r, ok := s.RootModule().Resources["stackit_alb.loadbalancer"]
					if !ok {
						return "", fmt.Errorf("couldn't find resource stackit_alb.loadbalancer")
					}
					name, ok := r.Primary.Attributes["name"]
					if !ok {
						return "", fmt.Errorf("couldn't find attribute name")
					}
					region, ok := r.Primary.Attributes["region"]
					if !ok {
						return "", fmt.Errorf("couldn't find attribute region")
					}
					return fmt.Sprintf("%s,%s,%s", testutil.ProjectId, region, name), nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"options.private_network_only"},
			},
			// Update
			{
				ConfigVariables: configVarsMaxUpdated(),
				Config:          testutil.ALBProviderConfig() + resourceMaxConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "project_id", testutil.ConvertConfigVariable(testConfigVarsMax["project_id"])),
					resource.TestCheckResourceAttr("stackit_alb.loadbalancer", "name", testutil.ConvertConfigVariable(testConfigVarsMax["loadbalancer_name"])),
				),
			},
			// Deletion is done by the framework implicitly
		},
	})
}

func testAccCheckALBDestroy(s *terraform.State) error {
	ctx := context.Background()
	var client *alb.APIClient
	var err error
	if testutil.ALBCustomEndpoint == "" {
		client, err = alb.NewAPIClient()
	} else {
		client, err = alb.NewAPIClient(
			stackitSdkConfig.WithEndpoint(testutil.ALBCustomEndpoint),
		)
	}
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	region := "eu01"
	if testutil.Region != "" {
		region = testutil.Region
	}
	loadbalancersToDestroy := []string{}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "stackit_loadbalancer" {
			continue
		}
		// loadbalancer terraform ID: = "[project_id],[name]"
		loadbalancerName := strings.Split(rs.Primary.ID, core.Separator)[1]
		loadbalancersToDestroy = append(loadbalancersToDestroy, loadbalancerName)
	}

	loadbalancersResp, err := client.ListLoadBalancers(ctx, testutil.ProjectId, region).Execute()
	if err != nil {
		return fmt.Errorf("getting loadbalancersResp: %w", err)
	}

	if loadbalancersResp.LoadBalancers == nil || (loadbalancersResp.LoadBalancers != nil && len(*loadbalancersResp.LoadBalancers) == 0) {
		fmt.Print("No load balancers found for project \n")
		return nil
	}

	items := *loadbalancersResp.LoadBalancers
	for i := range items {
		if items[i].Name == nil {
			continue
		}
		if utils.Contains(loadbalancersToDestroy, *items[i].Name) {
			_, err := client.DeleteLoadBalancerExecute(ctx, testutil.ProjectId, region, *items[i].Name)
			if err != nil {
				return fmt.Errorf("destroying load balancer %s during CheckDestroy: %w", *items[i].Name, err)
			}
			_, err = wait.DeleteLoadbalancerWaitHandler(ctx, client, testutil.ProjectId, region, *items[i].Name).WaitWithContext(ctx)
			if err != nil {
				return fmt.Errorf("destroying load balancer %s during CheckDestroy: waiting for deletion %w", *items[i].Name, err)
			}
		}
	}
	return nil
}
