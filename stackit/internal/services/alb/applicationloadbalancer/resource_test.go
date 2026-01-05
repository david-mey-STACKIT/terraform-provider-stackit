package alb

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stackitcloud/stackit-sdk-go/services/alb"
	"k8s.io/utils/ptr"
)

const (
	projectID       = "b8c3fbaa-3ab4-4a8e-9584-de22453d046f"
	region          = "eu01"
	lbName          = "example-lb2"
	externalAddress = "188.34.80.229"
	lbVersion       = "lb-1"
	lbID            = projectID + "," + region + "," + lbName
	sgLBID          = "8c06e3b6-531b-43a0-b965-3ae73da83d1b"
	sgTargetID      = "19cc8a91-d590-4166-b27d-211da3cb44d3"
	targetPoolName  = "my-pool"
	credentialsRef  = "credentials-nzkp4"
)

func fixtureModel(explicitBool *bool, mods ...func(m *Model)) *Model {
	resp := &Model{
		Id:                             types.StringValue(lbID),
		ProjectId:                      types.StringValue(projectID),
		DisableSecurityGroupAssignment: types.BoolPointerValue(explicitBool),
		Errors: types.SetValueMust(
			types.ObjectType{AttrTypes: errorsType},
			[]attr.Value{
				types.ObjectValueMust(
					errorsType,
					map[string]attr.Value{
						"description": types.StringValue("quota test error"),
						"type":        types.StringValue(string(alb.LOADBALANCERERRORTYPE_QUOTA_SECGROUP_EXCEEDED)),
					},
				),
				types.ObjectValueMust(
					errorsType,
					map[string]attr.Value{
						"description": types.StringValue("fip test error"),
						"type":        types.StringValue(string(alb.LOADBALANCERERRORTYPE_FIP_NOT_CONFIGURED)),
					},
				),
			},
		),
		ExternalAddress: types.StringValue(externalAddress),
		Labels: types.MapValueMust(types.StringType, map[string]attr.Value{
			"key":  types.StringValue("value"),
			"key2": types.StringValue("value2"),
		}),
		Listeners: types.SetValueMust(
			types.ObjectType{AttrTypes: listenerTypes},
			[]attr.Value{
				types.ObjectValueMust(
					listenerTypes,
					map[string]attr.Value{
						"name":            types.StringValue("http-80"),
						"port":            types.Int64Value(80),
						"protocol":        types.StringValue("PROTOCOL_HTTP"),
						"waf_config_name": types.StringValue("my-waf-config"),
						"http": types.ObjectValueMust(
							httpTypes,
							map[string]attr.Value{
								"hosts": types.SetValueMust(
									types.ObjectType{AttrTypes: hostConfigTypes},
									[]attr.Value{types.ObjectValueMust(
										hostConfigTypes,
										map[string]attr.Value{
											"host": types.StringValue("*"),
											"rules": types.ListValueMust(
												types.ObjectType{AttrTypes: ruleTypes},
												[]attr.Value{types.ObjectValueMust(
													ruleTypes,
													map[string]attr.Value{
														"target_pool": types.StringValue(targetPoolName),
														"web_socket":  types.BoolPointerValue(explicitBool),
														"path": types.ObjectValueMust(pathTypes,
															map[string]attr.Value{
																"exact_match": types.StringNull(),
																"prefix":      types.StringValue("/"),
															},
														),
														"headers": types.SetValueMust(
															types.ObjectType{AttrTypes: headersTypes},
															[]attr.Value{types.ObjectValueMust(
																headersTypes,
																map[string]attr.Value{
																	"name":        types.StringValue("a-header"),
																	"exact_match": types.StringValue("value"),
																}),
															},
														),
														"query_parameters": types.SetValueMust(
															types.ObjectType{AttrTypes: queryParameterTypes},
															[]attr.Value{types.ObjectValueMust(
																queryParameterTypes,
																map[string]attr.Value{
																	"name":        types.StringValue("a_query_parameter"),
																	"exact_match": types.StringValue("value"),
																}),
															},
														),
														"cookie_persistence": types.ObjectValueMust(
															cookiePersistenceTypes,
															map[string]attr.Value{
																"name": types.StringValue("cookie_name"),
																"ttl":  types.StringValue("3s"),
															},
														),
													},
												),
												}),
										},
									),
									}),
							},
						),
						"https": types.ObjectValueMust(
							httpsTypes,
							map[string]attr.Value{
								"certificate_config": types.ObjectValueMust(
									certificateConfigTypes,
									map[string]attr.Value{
										"certificate_ids": types.SetValueMust(
											types.StringType,
											[]attr.Value{
												types.StringValue(credentialsRef),
											},
										),
									},
								),
							},
						),
					},
				),
			},
		),
		LoadBalancerSecurityGroup: types.ObjectValueMust(
			loadBalancerSecurityGroupType,
			map[string]attr.Value{
				"id":   types.StringValue(sgLBID),
				"name": types.StringValue("loadbalancer/" + lbName + "/backend-port"),
			},
		),
		Name: types.StringValue(lbName),
		Networks: types.SetValueMust(
			types.ObjectType{AttrTypes: networkTypes},
			[]attr.Value{
				types.ObjectValueMust(
					networkTypes,
					map[string]attr.Value{
						"network_id": types.StringValue("c7c92cc1-a6bd-4e15-a129-b6e2b9899bbc"),
						"role":       types.StringValue("ROLE_LISTENERS"),
					},
				),
				types.ObjectValueMust(
					networkTypes,
					map[string]attr.Value{
						"network_id": types.StringValue("ed3f1822-ca1c-4969-bea6-74c6b3e9aa40"),
						"role":       types.StringValue("ROLE_TARGETS"),
					},
				),
			},
		),
		Options: types.ObjectValueMust(
			optionsTypes,
			map[string]attr.Value{
				"acl": types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("192.168.0.0"),
						types.StringValue("192.168.0.1"),
					},
				),
				"ephemeral_address":    types.BoolPointerValue(explicitBool),
				"private_network_only": types.BoolPointerValue(explicitBool),
				"observability": types.ObjectValueMust(
					observabilityTypes,
					map[string]attr.Value{
						"logs": types.ObjectValueMust(
							observabilityOptionTypes,
							map[string]attr.Value{
								"credentials_ref": types.StringValue(credentialsRef),
								"push_url":        types.StringValue("http://www.example.org/push"),
							},
						),
						"metrics": types.ObjectValueMust(
							observabilityOptionTypes,
							map[string]attr.Value{
								"credentials_ref": types.StringValue(credentialsRef),
								"push_url":        types.StringValue("http://www.example.org/pull"),
							},
						),
					},
				),
			},
		),
		PlanId:         types.StringValue("p10"),
		PrivateAddress: types.StringValue("10.1.11.0"),
		Region:         types.StringValue(region),
		Status:         types.StringValue("STATUS_READY"),
		TargetPools: types.SetValueMust(
			types.ObjectType{AttrTypes: targetPoolTypes},
			[]attr.Value{
				types.ObjectValueMust(
					targetPoolTypes,
					map[string]attr.Value{
						"name":        types.StringValue(targetPoolName),
						"target_port": types.Int64Value(80),
						"targets": types.SetValueMust(
							types.ObjectType{AttrTypes: targetTypes},
							[]attr.Value{
								types.ObjectValueMust(
									targetTypes,
									map[string]attr.Value{
										"display_name": types.StringValue("test-backend-server"),
										"ip":           types.StringValue("192.168.0.218"),
									},
								),
							},
						),
						"tls_config": types.ObjectValueMust(
							tlsConfigTypes,
							map[string]attr.Value{
								"enabled":                     types.BoolPointerValue(explicitBool),
								"custom_ca":                   types.StringNull(),
								"skip_certificate_validation": types.BoolPointerValue(explicitBool),
							},
						),
						"active_health_check": types.ObjectValueMust(
							activeHealthCheckTypes,
							map[string]attr.Value{
								"healthy_threshold":   types.Int64Value(1),
								"interval":            types.StringValue("2s"),
								"interval_jitter":     types.StringValue("3s"),
								"timeout":             types.StringValue("4s"),
								"unhealthy_threshold": types.Int64Value(5),
								"http_health_checks": types.ObjectValueMust(
									httpHealthChecksTypes,
									map[string]attr.Value{
										"ok_status": types.SetValueMust(
											types.StringType,
											[]attr.Value{
												types.StringValue("200"),
												types.StringValue("201"),
											},
										),
										"path": types.StringValue("/health"),
									},
								),
							},
						),
					},
				),
			},
		),
		TargetSecurityGroup: types.ObjectValueMust(
			targetSecurityGroupType,
			map[string]attr.Value{
				"id":   types.StringValue(sgTargetID),
				"name": types.StringValue("loadbalancer/" + lbName + "/backend"),
			},
		),
		Version: types.StringValue(lbVersion),
	}
	for _, mod := range mods {
		mod(resp)
	}
	return resp
}

func fixtureModelNull(mods ...func(m *Model)) *Model {
	resp := &Model{
		Id:                             types.StringNull(),
		ProjectId:                      types.StringNull(),
		DisableSecurityGroupAssignment: types.BoolNull(),
		Errors:                         types.SetNull(types.ObjectType{AttrTypes: errorsType}),
		ExternalAddress:                types.StringNull(),
		Labels:                         types.MapNull(types.StringType),
		Listeners:                      types.SetNull(types.ObjectType{AttrTypes: listenerTypes}),
		LoadBalancerSecurityGroup:      types.ObjectNull(loadBalancerSecurityGroupType),
		Name:                           types.StringNull(),
		Networks:                       types.SetNull(types.ObjectType{AttrTypes: networkTypes}),
		Options:                        types.ObjectNull(optionsTypes),
		PlanId:                         types.StringNull(),
		PrivateAddress:                 types.StringNull(),
		Region:                         types.StringNull(),
		Status:                         types.StringNull(),
		TargetPools:                    types.SetNull(types.ObjectType{AttrTypes: targetPoolTypes}),
		TargetSecurityGroup:            types.ObjectNull(targetSecurityGroupType),
		Version:                        types.StringNull(),
	}
	for _, mod := range mods {
		mod(resp)
	}
	return resp
}

func fixtureCreatePayload(lb *alb.LoadBalancer) *alb.CreateLoadBalancerPayload {
	(*lb.Listeners)[0].Name = nil // will be required in ALB API V2

	return &alb.CreateLoadBalancerPayload{
		DisableTargetSecurityGroupAssignment: lb.DisableTargetSecurityGroupAssignment,
		ExternalAddress:                      lb.ExternalAddress,
		Labels:                               lb.Labels,
		Listeners:                            lb.Listeners,
		Name:                                 lb.Name,
		Networks:                             lb.Networks,
		Options:                              lb.Options,
		PlanId:                               lb.PlanId,
		TargetPools:                          lb.TargetPools,
	}
}

func fixtureUpdatePayload(lb *alb.LoadBalancer) *alb.UpdateLoadBalancerPayload {
	(*lb.Listeners)[0].Name = nil // will be required in ALB API V2

	return &alb.UpdateLoadBalancerPayload{
		DisableTargetSecurityGroupAssignment: lb.DisableTargetSecurityGroupAssignment,
		ExternalAddress:                      lb.ExternalAddress,
		Labels:                               lb.Labels,
		Listeners:                            lb.Listeners,
		Name:                                 lb.Name,
		Networks:                             lb.Networks,
		Options:                              lb.Options,
		PlanId:                               lb.PlanId,
		TargetPools:                          lb.TargetPools,
		Version:                              lb.Version,
	}
}

func fixtureApplicationLoadBalancer(explicitBool *bool, mods ...func(m *alb.LoadBalancer)) *alb.LoadBalancer {
	resp := &alb.LoadBalancer{
		DisableTargetSecurityGroupAssignment: explicitBool,
		ExternalAddress:                      ptr.To(externalAddress),
		Errors: ptr.To([]alb.LoadBalancerError{
			{
				Description: ptr.To("quota test error"),
				Type:        ptr.To(alb.LOADBALANCERERRORTYPE_QUOTA_SECGROUP_EXCEEDED),
			},
			{
				Description: ptr.To("fip test error"),
				Type:        ptr.To(alb.LOADBALANCERERRORTYPE_FIP_NOT_CONFIGURED),
			},
		}),
		Name:           ptr.To(lbName),
		PlanId:         ptr.To("p10"),
		PrivateAddress: ptr.To("10.1.11.0"),
		Region:         ptr.To(region),
		Status:         ptr.To(alb.LoadBalancerStatus("STATUS_READY")),
		Version:        ptr.To(lbVersion),
		Labels: &map[string]string{
			"key":  "value",
			"key2": "value2",
		},
		Networks: &[]alb.Network{
			{
				NetworkId: ptr.To("c7c92cc1-a6bd-4e15-a129-b6e2b9899bbc"),
				Role:      ptr.To(alb.NetworkRole("ROLE_LISTENERS")),
			},
			{
				NetworkId: ptr.To("ed3f1822-ca1c-4969-bea6-74c6b3e9aa40"),
				Role:      ptr.To(alb.NetworkRole("ROLE_TARGETS")),
			},
		},
		Listeners: &[]alb.Listener{
			{
				Name:     ptr.To("http-80"),
				Port:     ptr.To(int64(80)),
				Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
				Http: &alb.ProtocolOptionsHTTP{
					Hosts: &[]alb.HostConfig{
						{
							Host: ptr.To("*"),
							Rules: &[]alb.Rule{
								{
									TargetPool: ptr.To(targetPoolName),
									WebSocket:  explicitBool,
									Path: &alb.Path{
										Prefix: ptr.To("/"),
									},
									Headers: &[]alb.HttpHeader{
										{Name: ptr.To("a-header"), ExactMatch: ptr.To("value")},
									},
									QueryParameters: &[]alb.QueryParameter{
										{Name: ptr.To("a_query_parameter"), ExactMatch: ptr.To("value")},
									},
									CookiePersistence: &alb.CookiePersistence{
										Name: ptr.To("cookie_name"),
										Ttl:  ptr.To("3s"),
									},
								},
							},
						},
					},
				},
				Https: &alb.ProtocolOptionsHTTPS{
					CertificateConfig: ptr.To(alb.CertificateConfig{
						CertificateIds: &[]string{
							credentialsRef,
						},
					}),
				},
				WafConfigName: ptr.To("my-waf-config"),
			},
		},
		TargetPools: &[]alb.TargetPool{
			{
				Name:       ptr.To(targetPoolName),
				TargetPort: ptr.To(int64(80)),
				Targets: &[]alb.Target{
					{
						DisplayName: ptr.To("test-backend-server"),
						Ip:          ptr.To("192.168.0.218"),
					},
				},
				TlsConfig: &alb.TargetPoolTlsConfig{
					Enabled:                   explicitBool,
					SkipCertificateValidation: explicitBool,
				},
				ActiveHealthCheck: &alb.ActiveHealthCheck{
					HealthyThreshold:   ptr.To(int64(1)),
					UnhealthyThreshold: ptr.To(int64(5)),
					Interval:           ptr.To("2s"),
					IntervalJitter:     ptr.To("3s"),
					Timeout:            ptr.To("4s"),
					HttpHealthChecks: &alb.HttpHealthChecks{
						Path:       ptr.To("/health"),
						OkStatuses: &[]string{"200", "201"},
					},
				},
			},
		},
		Options: ptr.To(alb.LoadBalancerOptions{
			EphemeralAddress:   explicitBool,
			PrivateNetworkOnly: explicitBool,
			Observability: &alb.LoadbalancerOptionObservability{
				Logs: &alb.LoadbalancerOptionLogs{
					CredentialsRef: ptr.To(credentialsRef),
					PushUrl:        ptr.To("http://www.example.org/push"),
				},
				Metrics: &alb.LoadbalancerOptionMetrics{
					CredentialsRef: ptr.To(credentialsRef),
					PushUrl:        ptr.To("http://www.example.org/pull"),
				},
			},
			AccessControl: &alb.LoadbalancerOptionAccessControl{
				AllowedSourceRanges: &[]string{"192.168.0.0", "192.168.0.1"},
			},
		}),
		LoadBalancerSecurityGroup: &alb.CreateLoadBalancerPayloadLoadBalancerSecurityGroup{
			Id:   ptr.To(sgLBID),
			Name: ptr.To("loadbalancer/" + lbName + "/backend-port"),
		},
		TargetSecurityGroup: &alb.CreateLoadBalancerPayloadTargetSecurityGroup{
			Id:   ptr.To(sgTargetID),
			Name: ptr.To("loadbalancer/" + lbName + "/backend"),
		},
	}
	for _, mod := range mods {
		mod(resp)
	}
	return resp
}

func TestToCreatePayload(t *testing.T) {
	tests := []struct {
		description string
		input       *Model
		expected    *alb.CreateLoadBalancerPayload
		isValid     bool
	}{
		{
			description: "valid",
			input:       fixtureModel(nil),
			expected:    fixtureCreatePayload(fixtureApplicationLoadBalancer(nil)),
			isValid:     true,
		},
		/*{
			"simple_values_ok",
			&Model{
				ExternalAddress: types.StringValue("external_address"),
				Listeners: types.SetValueMust(types.ObjectType{AttrTypes: listenerTypes}, []attr.Value{
					types.ObjectValueMust(listenerTypes, map[string]attr.Value{
						"display_name": types.StringValue("display_name"),
						"port":         types.Int64Value(80),
						"protocol":     types.StringValue(string(alb.LISTENERPROTOCOL_HTTP)),
						"target_pool":  types.StringValue("target_pool"),
					}),
				}),
				Name: types.StringValue("name"),
				Networks: types.SetValueMust(types.ObjectType{AttrTypes: networkTypes}, []attr.Value{
					types.ObjectValueMust(networkTypes, map[string]attr.Value{
						"network_id": types.StringValue("network_id"),
						"role":       types.StringValue(string(alb.NETWORKROLE_LISTENERS_AND_TARGETS)),
					}),
					types.ObjectValueMust(networkTypes, map[string]attr.Value{
						"network_id": types.StringValue("network_id_2"),
						"role":       types.StringValue(string(alb.NETWORKROLE_LISTENERS_AND_TARGETS)),
					}),
				}),
				Options: types.ObjectValueMust(
					optionsTypes,
					map[string]attr.Value{
						"acl": types.SetValueMust(
							types.StringType,
							[]attr.Value{types.StringValue("cidr")}),
						"private_network_only": types.BoolValue(true),
						"observability": types.ObjectValueMust(observabilityTypes, map[string]attr.Value{
							"logs": types.ObjectValueMust(observabilityOptionTypes, map[string]attr.Value{
								"credentials_ref": types.StringValue("logs-credentials_ref"),
								"push_url":        types.StringValue("logs-push_url"),
							}),
							"metrics": types.ObjectValueMust(observabilityOptionTypes, map[string]attr.Value{
								"credentials_ref": types.StringValue("metrics-credentials_ref"),
								"push_url":        types.StringValue("metrics-push_url"),
							}),
						}),
					},
				),
				TargetPools: types.SetValueMust(types.ObjectType{AttrTypes: targetPoolTypes}, []attr.Value{
					types.ObjectValueMust(targetPoolTypes, map[string]attr.Value{
						"active_health_check": types.ObjectValueMust(activeHealthCheckTypes, map[string]attr.Value{
							"healthy_threshold":   types.Int64Value(1),
							"interval":            types.StringValue("2s"),
							"interval_jitter":     types.StringValue("3s"),
							"timeout":             types.StringValue("4s"),
							"unhealthy_threshold": types.Int64Value(5),
						}),
						"name":        types.StringValue("name"),
						"target_port": types.Int64Value(80),
						"targets": types.SetValueMust(types.ObjectType{AttrTypes: targetTypes}, []attr.Value{
							types.ObjectValueMust(targetTypes, map[string]attr.Value{
								"display_name": types.StringValue("display_name"),
								"ip":           types.StringValue("ip"),
							}),
						}),
					}),
				}),
			},
			&alb.CreateLoadBalancerPayload{
				ExternalAddress: utils.Ptr("external_address"),
				Listeners: &[]alb.Listener{
					{
						Name:     utils.Ptr("display_name"),
						Port:     utils.Ptr(int64(80)),
						Protocol: alb.LISTENERPROTOCOL_HTTP.Ptr(),
					},
				},
				Name: utils.Ptr("name"),
				Networks: &[]alb.Network{
					{
						NetworkId: utils.Ptr("network_id"),
						Role:      alb.NETWORKROLE_LISTENERS_AND_TARGETS.Ptr(),
					},
					{
						NetworkId: utils.Ptr("network_id_2"),
						Role:      alb.NETWORKROLE_LISTENERS_AND_TARGETS.Ptr(),
					},
				},
				Options: &alb.LoadBalancerOptions{
					AccessControl: &alb.LoadbalancerOptionAccessControl{
						AllowedSourceRanges: &[]string{"cidr"},
					},
					PrivateNetworkOnly: utils.Ptr(true),
					Observability: &alb.LoadbalancerOptionObservability{
						Logs: &alb.LoadbalancerOptionLogs{
							CredentialsRef: utils.Ptr("logs-credentials_ref"),
							PushUrl:        utils.Ptr("logs-push_url"),
						},
						Metrics: &alb.LoadbalancerOptionMetrics{
							CredentialsRef: utils.Ptr("metrics-credentials_ref"),
							PushUrl:        utils.Ptr("metrics-push_url"),
						},
					},
				},
				TargetPools: &[]alb.TargetPool{
					{
						ActiveHealthCheck: &alb.ActiveHealthCheck{
							HealthyThreshold:   utils.Ptr(int64(1)),
							Interval:           utils.Ptr("2s"),
							IntervalJitter:     utils.Ptr("3s"),
							Timeout:            utils.Ptr("4s"),
							UnhealthyThreshold: utils.Ptr(int64(5)),
						},
						Name:       utils.Ptr("name"),
						TargetPort: utils.Ptr(int64(80)),
						Targets: &[]alb.Target{
							{
								DisplayName: utils.Ptr("display_name"),
								Ip:          utils.Ptr("ip"),
							},
						},
					},
				},
			},
			true,
		},
		{
			"service_plan_ok",
			&Model{
				PlanId:          types.StringValue("p10"),
				ExternalAddress: types.StringValue("external_address"),
				Listeners: types.SetValueMust(types.ObjectType{AttrTypes: listenerTypes}, []attr.Value{
					types.ObjectValueMust(listenerTypes, map[string]attr.Value{
						"display_name": types.StringValue("display_name"),
						"port":         types.Int64Value(80),
						"protocol":     types.StringValue(string(alb.LISTENERPROTOCOL_HTTP)),
						"target_pool":  types.StringValue("target_pool"),
					}),
				}),
				Name: types.StringValue("name"),
				Networks: types.SetValueMust(types.ObjectType{AttrTypes: networkTypes}, []attr.Value{
					types.ObjectValueMust(networkTypes, map[string]attr.Value{
						"network_id": types.StringValue("network_id"),
						"role":       types.StringValue(string(alb.NETWORKROLE_LISTENERS_AND_TARGETS)),
					}),
					types.ObjectValueMust(networkTypes, map[string]attr.Value{
						"network_id": types.StringValue("network_id_2"),
						"role":       types.StringValue(string(alb.NETWORKROLE_LISTENERS_AND_TARGETS)),
					}),
				}),
				Options: types.ObjectValueMust(
					optionsTypes,
					map[string]attr.Value{
						"acl": types.SetValueMust(
							types.StringType,
							[]attr.Value{types.StringValue("cidr")}),
						"private_network_only": types.BoolValue(true),
						"observability": types.ObjectValueMust(observabilityTypes, map[string]attr.Value{
							"logs": types.ObjectValueMust(observabilityOptionTypes, map[string]attr.Value{
								"credentials_ref": types.StringValue("logs-credentials_ref"),
								"push_url":        types.StringValue("logs-push_url"),
							}),
							"metrics": types.ObjectValueMust(observabilityOptionTypes, map[string]attr.Value{
								"credentials_ref": types.StringValue("metrics-credentials_ref"),
								"push_url":        types.StringValue("metrics-push_url"),
							}),
						}),
					},
				),
				TargetPools: types.SetValueMust(types.ObjectType{AttrTypes: targetPoolTypes}, []attr.Value{
					types.ObjectValueMust(targetPoolTypes, map[string]attr.Value{
						"active_health_check": types.ObjectValueMust(activeHealthCheckTypes, map[string]attr.Value{
							"healthy_threshold":   types.Int64Value(1),
							"interval":            types.StringValue("2s"),
							"interval_jitter":     types.StringValue("3s"),
							"timeout":             types.StringValue("4s"),
							"unhealthy_threshold": types.Int64Value(5),
						}),
						"name":        types.StringValue("name"),
						"target_port": types.Int64Value(80),
						"targets": types.SetValueMust(types.ObjectType{AttrTypes: targetTypes}, []attr.Value{
							types.ObjectValueMust(targetTypes, map[string]attr.Value{
								"display_name": types.StringValue("display_name"),
								"ip":           types.StringValue("ip"),
							}),
						}),
					}),
				}),
			},
			&alb.CreateLoadBalancerPayload{
				PlanId:          utils.Ptr("p10"),
				ExternalAddress: utils.Ptr("external_address"),
				Listeners: &[]alb.Listener{
					{
						Name:     utils.Ptr("display_name"),
						Port:     utils.Ptr(int64(80)),
						Protocol: alb.LISTENERPROTOCOL_HTTP.Ptr(),
					},
				},
				Name: utils.Ptr("name"),
				Networks: &[]alb.Network{
					{
						NetworkId: utils.Ptr("network_id"),
						Role:      alb.NETWORKROLE_LISTENERS_AND_TARGETS.Ptr(),
					},
					{
						NetworkId: utils.Ptr("network_id_2"),
						Role:      alb.NETWORKROLE_LISTENERS_AND_TARGETS.Ptr(),
					},
				},
				Options: &alb.LoadBalancerOptions{
					AccessControl: &alb.LoadbalancerOptionAccessControl{
						AllowedSourceRanges: &[]string{"cidr"},
					},
					PrivateNetworkOnly: utils.Ptr(true),
					Observability: &alb.LoadbalancerOptionObservability{
						Logs: &alb.LoadbalancerOptionLogs{
							CredentialsRef: utils.Ptr("logs-credentials_ref"),
							PushUrl:        utils.Ptr("logs-push_url"),
						},
						Metrics: &alb.LoadbalancerOptionMetrics{
							CredentialsRef: utils.Ptr("metrics-credentials_ref"),
							PushUrl:        utils.Ptr("metrics-push_url"),
						},
					},
				},
				TargetPools: &[]alb.TargetPool{
					{
						ActiveHealthCheck: &alb.ActiveHealthCheck{
							HealthyThreshold:   utils.Ptr(int64(1)),
							Interval:           utils.Ptr("2s"),
							IntervalJitter:     utils.Ptr("3s"),
							Timeout:            utils.Ptr("4s"),
							UnhealthyThreshold: utils.Ptr(int64(5)),
						},
						Name:       utils.Ptr("name"),
						TargetPort: utils.Ptr(int64(80)),
						Targets: &[]alb.Target{
							{
								DisplayName: utils.Ptr("display_name"),
								Ip:          utils.Ptr("ip"),
							},
						},
					},
				},
			},
			true,
		},
		{
			"nil_model",
			nil,
			nil,
			false,
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output, err := toCreatePayload(context.Background(), tt.input)
			if !tt.isValid && err == nil {
				t.Fatalf("Should have failed")
			}
			if tt.isValid && err != nil {
				t.Fatalf("Should not have failed: %v", err)
			}
			if tt.isValid {
				diff := cmp.Diff(output, tt.expected)
				if diff != "" {
					t.Fatalf("Data does not match: %s", diff)
				}
			}
		})
	}
}

func TestToTargetPoolUpdatePayload(t *testing.T) {
	tests := []struct {
		description string
		input       *Model
		expected    *alb.UpdateLoadBalancerPayload
		isValid     bool
	}{
		{
			description: "valid",
			input:       fixtureModel(nil),
			expected:    fixtureUpdatePayload(fixtureApplicationLoadBalancer(nil)),
			isValid:     true,
		}, /*
			{
				"simple_values_ok",
				fixtureModel(nil),
				fixtureUpdateLoadBalancerPayload(),
				true,
			},
			{
				"nil_target_pool",
				nil,
				nil,
				false,
			},*/
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output, err := toUpdatePayload(context.Background(), tt.input)
			if !tt.isValid && err == nil {
				t.Fatalf("Should have failed")
			}
			if tt.isValid && err != nil {
				t.Fatalf("Should not have failed: %v", err)
			}
			if tt.isValid {
				diff := cmp.Diff(output, tt.expected)
				if diff != "" {
					t.Fatalf("Data does not match: %s", diff)
				}
			}
		})
	}
}

func TestMapFields(t *testing.T) {
	const testRegion = "eu01"
	tests := []struct {
		description             string
		input                   *alb.LoadBalancer
		output                  *Model
		modelPrivateNetworkOnly *bool
		region                  string
		expected                *Model
		isValid                 bool
	}{
		{
			description: "valid full model",
			input:       fixtureApplicationLoadBalancer(nil),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region:   testRegion,
			expected: fixtureModel(nil),
			isValid:  true,
		},
		{
			description: "error alb nil",
			input:       nil,
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region:   testRegion,
			expected: fixtureModel(nil),
			isValid:  false,
		},
		{
			description: "error model nil",
			input:       fixtureApplicationLoadBalancer(nil),
			output:      nil,
			region:      testRegion,
			expected:    fixtureModel(nil),
			isValid:     false,
		},
		{
			description: "error no name",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Name = nil
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
				Name:      types.StringValue(""),
			},
			region:   testRegion,
			expected: fixtureModel(nil),
			isValid:  false,
		},
		{
			description: "valid name in model",
			input:       fixtureApplicationLoadBalancer(nil),
			output: &Model{
				ProjectId: types.StringValue(projectID),
				Name:      types.StringValue(lbName),
			},
			region:   testRegion,
			expected: fixtureModel(nil),
			isValid:  true,
		},
		{
			description: "false - explicitly set",
			input:       fixtureApplicationLoadBalancer(ptr.To(false)),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region:   testRegion,
			expected: fixtureModel(ptr.To(false)),
			isValid:  true,
		},
		{
			description: "true - explicitly set",
			input:       fixtureApplicationLoadBalancer(ptr.To(true)),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region:   testRegion,
			expected: fixtureModel(ptr.To(true)),
			isValid:  true,
		},
		{
			description: "false - only in model set",
			input:       fixtureApplicationLoadBalancer(nil),
			output:      fixtureModel(ptr.To(false)),
			region:      testRegion,
			expected:    fixtureModel(ptr.To(false)),
			isValid:     true,
		},
		{
			description: "true - only in model set",
			input:       fixtureApplicationLoadBalancer(nil),
			output:      fixtureModel(ptr.To(true)),
			region:      testRegion,
			expected:    fixtureModel(nil),
			isValid:     true,
		},
		{
			description: "valid empty",
			input:       &alb.LoadBalancer{},
			output: &Model{
				ProjectId: types.StringValue(projectID),
				Name:      types.StringValue(lbName),
			},
			region: testRegion,
			expected: fixtureModelNull(func(m *Model) {
				m.Id = types.StringValue(strings.Join([]string{projectID, region, lbName}, ","))
				m.ProjectId = types.StringValue(projectID)
				m.Name = types.StringValue(lbName)
				m.Region = types.StringValue(region)
			}),
			isValid: true,
		},
		{
			description: "mapTargets no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.TargetPools = &[]alb.TargetPool{
					{ // empty target pool
						ActiveHealthCheck: nil,
						Name:              nil,
						TargetPort:        nil,
						Targets:           nil,
						TlsConfig:         nil,
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.TargetPools = types.SetValueMust(
					types.ObjectType{AttrTypes: targetPoolTypes},
					[]attr.Value{
						types.ObjectValueMust(
							targetPoolTypes,
							map[string]attr.Value{
								"name":                types.StringNull(),
								"target_port":         types.Int64Null(),
								"targets":             types.SetNull(types.ObjectType{AttrTypes: targetTypes}),
								"tls_config":          types.ObjectNull(tlsConfigTypes),
								"active_health_check": types.ObjectNull(activeHealthCheckTypes),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapHttpHealthChecks no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.TargetPools = &[]alb.TargetPool{
					{
						Name:       ptr.To(targetPoolName),
						TargetPort: ptr.To(int64(80)),
						Targets: &[]alb.Target{
							{
								DisplayName: ptr.To("test-backend-server"),
								Ip:          ptr.To("192.168.0.218"),
							},
						},
						ActiveHealthCheck: &alb.ActiveHealthCheck{
							HealthyThreshold:   ptr.To(int64(1)),
							UnhealthyThreshold: ptr.To(int64(5)),
							Interval:           ptr.To("2s"),
							IntervalJitter:     ptr.To("3s"),
							Timeout:            ptr.To("4s"),
						},
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.TargetPools = types.SetValueMust(
					types.ObjectType{AttrTypes: targetPoolTypes},
					[]attr.Value{
						types.ObjectValueMust(
							targetPoolTypes,
							map[string]attr.Value{
								"name":        types.StringValue(targetPoolName),
								"target_port": types.Int64Value(80),
								"targets": types.SetValueMust(
									types.ObjectType{AttrTypes: targetTypes},
									[]attr.Value{
										types.ObjectValueMust(
											targetTypes,
											map[string]attr.Value{
												"display_name": types.StringValue("test-backend-server"),
												"ip":           types.StringValue("192.168.0.218"),
											},
										),
									},
								),
								"tls_config": types.ObjectNull(tlsConfigTypes),
								"active_health_check": types.ObjectValueMust(
									activeHealthCheckTypes,
									map[string]attr.Value{
										"healthy_threshold":   types.Int64Value(1),
										"interval":            types.StringValue("2s"),
										"interval_jitter":     types.StringValue("3s"),
										"timeout":             types.StringValue("4s"),
										"unhealthy_threshold": types.Int64Value(5),
										"http_health_checks":  types.ObjectNull(httpHealthChecksTypes),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapOptions no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Options = &alb.LoadBalancerOptions{
					AccessControl: nil,
					Observability: nil,
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Options = types.ObjectValueMust(optionsTypes,
					map[string]attr.Value{
						"acl":                  types.SetNull(types.StringType),
						"observability":        types.ObjectNull(observabilityTypes),
						"private_network_only": types.BoolNull(),
						"ephemeral_address":    types.BoolNull(),
					})
			}),
			isValid: true,
		},
		{
			description: "mapCertificates no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http: &alb.ProtocolOptionsHTTP{
							Hosts: &[]alb.HostConfig{
								{
									Host: ptr.To("*"),
									Rules: &[]alb.Rule{
										{
											TargetPool: ptr.To(targetPoolName),
											WebSocket:  nil,
											Path: &alb.Path{
												Prefix: ptr.To("/"),
											},
											Headers: &[]alb.HttpHeader{
												{Name: ptr.To("a-header"), ExactMatch: ptr.To("value")},
											},
											QueryParameters: &[]alb.QueryParameter{
												{Name: ptr.To("a_query_parameter"), ExactMatch: ptr.To("value")},
											},
											CookiePersistence: &alb.CookiePersistence{
												Name: ptr.To("cookie_name"),
												Ttl:  ptr.To("3s"),
											},
										},
									},
								},
							},
						},
						Https: &alb.ProtocolOptionsHTTPS{
							CertificateConfig: nil,
						},
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http": types.ObjectValueMust(
									httpTypes,
									map[string]attr.Value{
										"hosts": types.SetValueMust(
											types.ObjectType{AttrTypes: hostConfigTypes},
											[]attr.Value{types.ObjectValueMust(
												hostConfigTypes,
												map[string]attr.Value{
													"host": types.StringValue("*"),
													"rules": types.ListValueMust(
														types.ObjectType{AttrTypes: ruleTypes},
														[]attr.Value{types.ObjectValueMust(
															ruleTypes,
															map[string]attr.Value{
																"target_pool": types.StringValue(targetPoolName),
																"web_socket":  types.BoolPointerValue(nil),
																"path": types.ObjectValueMust(pathTypes,
																	map[string]attr.Value{
																		"exact_match": types.StringNull(),
																		"prefix":      types.StringValue("/"),
																	},
																),
																"headers": types.SetValueMust(
																	types.ObjectType{AttrTypes: headersTypes},
																	[]attr.Value{types.ObjectValueMust(
																		headersTypes,
																		map[string]attr.Value{
																			"name":        types.StringValue("a-header"),
																			"exact_match": types.StringValue("value"),
																		}),
																	},
																),
																"query_parameters": types.SetValueMust(
																	types.ObjectType{AttrTypes: queryParameterTypes},
																	[]attr.Value{types.ObjectValueMust(
																		queryParameterTypes,
																		map[string]attr.Value{
																			"name":        types.StringValue("a_query_parameter"),
																			"exact_match": types.StringValue("value"),
																		}),
																	},
																),
																"cookie_persistence": types.ObjectValueMust(
																	cookiePersistenceTypes,
																	map[string]attr.Value{
																		"name": types.StringValue("cookie_name"),
																		"ttl":  types.StringValue("3s"),
																	},
																),
															},
														),
														}),
												},
											),
											}),
									},
								),
								"https": types.ObjectValueMust(
									httpsTypes,
									map[string]attr.Value{
										"certificate_config": types.ObjectNull(certificateConfigTypes),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapHttps no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http: &alb.ProtocolOptionsHTTP{
							Hosts: &[]alb.HostConfig{
								{
									Host: ptr.To("*"),
									Rules: &[]alb.Rule{
										{
											TargetPool: ptr.To(targetPoolName),
											WebSocket:  nil,
											Path: &alb.Path{
												Prefix: ptr.To("/"),
											},
											Headers: &[]alb.HttpHeader{
												{Name: ptr.To("a-header"), ExactMatch: ptr.To("value")},
											},
											QueryParameters: &[]alb.QueryParameter{
												{Name: ptr.To("a_query_parameter"), ExactMatch: ptr.To("value")},
											},
											CookiePersistence: &alb.CookiePersistence{
												Name: ptr.To("cookie_name"),
												Ttl:  ptr.To("3s"),
											},
										},
									},
								},
							},
						},
						Https:         nil,
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http": types.ObjectValueMust(
									httpTypes,
									map[string]attr.Value{
										"hosts": types.SetValueMust(
											types.ObjectType{AttrTypes: hostConfigTypes},
											[]attr.Value{types.ObjectValueMust(
												hostConfigTypes,
												map[string]attr.Value{
													"host": types.StringValue("*"),
													"rules": types.ListValueMust(
														types.ObjectType{AttrTypes: ruleTypes},
														[]attr.Value{types.ObjectValueMust(
															ruleTypes,
															map[string]attr.Value{
																"target_pool": types.StringValue(targetPoolName),
																"web_socket":  types.BoolPointerValue(nil),
																"path": types.ObjectValueMust(pathTypes,
																	map[string]attr.Value{
																		"exact_match": types.StringNull(),
																		"prefix":      types.StringValue("/"),
																	},
																),
																"headers": types.SetValueMust(
																	types.ObjectType{AttrTypes: headersTypes},
																	[]attr.Value{types.ObjectValueMust(
																		headersTypes,
																		map[string]attr.Value{
																			"name":        types.StringValue("a-header"),
																			"exact_match": types.StringValue("value"),
																		}),
																	},
																),
																"query_parameters": types.SetValueMust(
																	types.ObjectType{AttrTypes: queryParameterTypes},
																	[]attr.Value{types.ObjectValueMust(
																		queryParameterTypes,
																		map[string]attr.Value{
																			"name":        types.StringValue("a_query_parameter"),
																			"exact_match": types.StringValue("value"),
																		}),
																	},
																),
																"cookie_persistence": types.ObjectValueMust(
																	cookiePersistenceTypes,
																	map[string]attr.Value{
																		"name": types.StringValue("cookie_name"),
																		"ttl":  types.StringValue("3s"),
																	},
																),
															},
														),
														}),
												},
											),
											}),
									},
								),
								"https": types.ObjectNull(httpsTypes),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapRules contents no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http: &alb.ProtocolOptionsHTTP{
							Hosts: &[]alb.HostConfig{
								{
									Host: ptr.To("*"),
									Rules: &[]alb.Rule{
										{
											TargetPool:        ptr.To(targetPoolName),
											WebSocket:         nil,
											Path:              nil,
											Headers:           nil,
											QueryParameters:   nil,
											CookiePersistence: nil,
										},
									},
								},
							},
						},
						Https: &alb.ProtocolOptionsHTTPS{
							CertificateConfig: ptr.To(alb.CertificateConfig{
								CertificateIds: &[]string{
									credentialsRef,
								},
							}),
						},
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http": types.ObjectValueMust(
									httpTypes,
									map[string]attr.Value{
										"hosts": types.SetValueMust(
											types.ObjectType{AttrTypes: hostConfigTypes},
											[]attr.Value{types.ObjectValueMust(
												hostConfigTypes,
												map[string]attr.Value{
													"host": types.StringValue("*"),
													"rules": types.ListValueMust(
														types.ObjectType{AttrTypes: ruleTypes},
														[]attr.Value{types.ObjectValueMust(
															ruleTypes,
															map[string]attr.Value{
																"target_pool":        types.StringValue(targetPoolName),
																"web_socket":         types.BoolPointerValue(nil),
																"path":               types.ObjectNull(pathTypes),
																"headers":            types.SetNull(types.ObjectType{AttrTypes: headersTypes}),
																"query_parameters":   types.SetNull(types.ObjectType{AttrTypes: queryParameterTypes}),
																"cookie_persistence": types.ObjectNull(cookiePersistenceTypes),
															},
														),
														}),
												},
											),
											}),
									},
								),
								"https": types.ObjectValueMust(
									httpsTypes,
									map[string]attr.Value{
										"certificate_config": types.ObjectValueMust(
											certificateConfigTypes,
											map[string]attr.Value{
												"certificate_ids": types.SetValueMust(
													types.StringType,
													[]attr.Value{
														types.StringValue(credentialsRef),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapRules no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http: &alb.ProtocolOptionsHTTP{
							Hosts: &[]alb.HostConfig{
								{
									Host:  ptr.To("*"),
									Rules: nil,
								},
							},
						},
						Https: &alb.ProtocolOptionsHTTPS{
							CertificateConfig: ptr.To(alb.CertificateConfig{
								CertificateIds: &[]string{
									credentialsRef,
								},
							}),
						},
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http": types.ObjectValueMust(
									httpTypes,
									map[string]attr.Value{
										"hosts": types.SetValueMust(
											types.ObjectType{AttrTypes: hostConfigTypes},
											[]attr.Value{types.ObjectValueMust(
												hostConfigTypes,
												map[string]attr.Value{
													"host":  types.StringValue("*"),
													"rules": types.ListNull(types.ObjectType{AttrTypes: ruleTypes}),
												}),
											},
										),
									},
								),
								"https": types.ObjectValueMust(
									httpsTypes,
									map[string]attr.Value{
										"certificate_config": types.ObjectValueMust(
											certificateConfigTypes,
											map[string]attr.Value{
												"certificate_ids": types.SetValueMust(
													types.StringType,
													[]attr.Value{
														types.StringValue(credentialsRef),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapHosts no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http: &alb.ProtocolOptionsHTTP{
							Hosts: nil,
						},
						Https: &alb.ProtocolOptionsHTTPS{
							CertificateConfig: ptr.To(alb.CertificateConfig{
								CertificateIds: &[]string{
									credentialsRef,
								},
							}),
						},
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http": types.ObjectValueMust(
									httpTypes,
									map[string]attr.Value{
										"hosts": types.SetNull(types.ObjectType{AttrTypes: hostConfigTypes}),
									},
								),
								"https": types.ObjectValueMust(
									httpsTypes,
									map[string]attr.Value{
										"certificate_config": types.ObjectValueMust(
											certificateConfigTypes,
											map[string]attr.Value{
												"certificate_ids": types.SetValueMust(
													types.StringType,
													[]attr.Value{
														types.StringValue(credentialsRef),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
		{
			description: "mapHttp no response",
			input: fixtureApplicationLoadBalancer(nil, func(m *alb.LoadBalancer) {
				m.Listeners = &[]alb.Listener{
					{
						Name:     ptr.To("http-80"),
						Port:     ptr.To(int64(80)),
						Protocol: ptr.To(alb.ListenerProtocol("PROTOCOL_HTTP")),
						Http:     nil,
						Https: &alb.ProtocolOptionsHTTPS{
							CertificateConfig: ptr.To(alb.CertificateConfig{
								CertificateIds: &[]string{
									credentialsRef,
								},
							}),
						},
						WafConfigName: ptr.To("my-waf-config"),
					},
				}
			}),
			output: &Model{
				ProjectId: types.StringValue(projectID),
			},
			region: testRegion,
			expected: fixtureModel(nil, func(m *Model) {
				m.Listeners = types.SetValueMust(
					types.ObjectType{AttrTypes: listenerTypes},
					[]attr.Value{
						types.ObjectValueMust(
							listenerTypes,
							map[string]attr.Value{
								"name":            types.StringValue("http-80"),
								"port":            types.Int64Value(80),
								"protocol":        types.StringValue("PROTOCOL_HTTP"),
								"waf_config_name": types.StringValue("my-waf-config"),
								"http":            types.ObjectNull(httpTypes),
								"https": types.ObjectValueMust(
									httpsTypes,
									map[string]attr.Value{
										"certificate_config": types.ObjectValueMust(
											certificateConfigTypes,
											map[string]attr.Value{
												"certificate_ids": types.SetValueMust(
													types.StringType,
													[]attr.Value{
														types.StringValue(credentialsRef),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				)
			}),
			isValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := mapFields(context.Background(), tt.input, tt.output, tt.region)
			if !tt.isValid && err == nil {
				t.Fatalf("Should have failed")
			}
			if tt.isValid && err != nil {
				t.Fatalf("Should not have failed: %v", err)
			}
			if tt.isValid {
				diff := cmp.Diff(tt.output, tt.expected)
				if diff != "" {
					t.Fatalf("Data does not match: %s", diff)
				}
			}
		})
	}
}
