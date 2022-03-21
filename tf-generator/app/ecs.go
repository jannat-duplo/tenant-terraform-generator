package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type ECS struct {
}

func (ecs *ECS) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== Duplo ECS TF generation started. =====>")
	workingDir := filepath.Join("target", config.CustomerName, config.AppProject)

	list, clientErr := client.EcsServiceList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, ecs := range *list {

			taskDefObj, clientErr := client.EcsTaskDefinitionGet(config.TenantId, ecs.TaskDefinition)
			if clientErr != nil {
				fmt.Println(clientErr)
				return
			}
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "ecs-"+ecs.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			log.Printf("[TRACE] Generating terraform config for duplo task definition : %s", taskDefObj.Family)
			// Add duplocloud_aws_host resource
			tdBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecs_task_definition",
					ecs.Name})
			tdBody := tdBlock.Body()
			// svcBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			tdBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			tdBody.SetAttributeValue("family",
				cty.StringVal(taskDefObj.Family))
			tdBody.SetAttributeValue("cpu",
				cty.StringVal(taskDefObj.CPU))
			tdBody.SetAttributeValue("memory",
				cty.StringVal(taskDefObj.Memory))
			tdBody.SetAttributeValue("network_mode",
				cty.StringVal(taskDefObj.NetworkMode.Value))

			if taskDefObj.RequiresCompatibilities != nil && len(taskDefObj.RequiresCompatibilities) > 0 {
				var vals []cty.Value
				for _, s := range taskDefObj.RequiresCompatibilities {
					vals = append(vals, cty.StringVal(s))
				}
				tdBody.SetAttributeValue("requires_compatibilities",
					cty.ListVal(vals))
			}
			if taskDefObj.Volumes != nil && len(taskDefObj.Volumes) > 0 {
				volString, err := duplosdk.PrettyStruct(taskDefObj.Volumes)
				if err != nil {
					panic(err)
				}
				tdBody.SetAttributeTraversal("volumes", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + volString + ")",
					},
				})
			}
			if taskDefObj.ContainerDefinitions != nil && len(taskDefObj.ContainerDefinitions) > 0 {
				containerString, err := duplosdk.PrettyStruct(taskDefObj.ContainerDefinitions)
				if err != nil {
					panic(err)
				}
				tdBody.SetAttributeTraversal("container_definitions", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + containerString + ")",
					},
				})
			}
			rootBody.AppendNewline()
			log.Printf("[TRACE] Terraform config generated for duplo task definition : %s", taskDefObj.Family)

			log.Printf("[TRACE] Generating terraform config for duplo ECS service : %s", ecs.Name)

			ecsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecs_service",
					ecs.Name})
			ecsBody := ecsBlock.Body()
			// svcBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			ecsBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			ecsBody.SetAttributeValue("name",
				cty.StringVal(ecs.Name))
			ecsBody.SetAttributeTraversal("task_definition", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "duplocloud_ecs_task_definition." + ecs.Name,
				},
				hcl.TraverseAttr{
					Name: "arn",
				},
			})
			ecsBody.SetAttributeValue("replicas",
				cty.NumberIntVal(int64(ecs.Replicas)))
			if ecs.HealthCheckGracePeriodSeconds > 0 {
				ecsBody.SetAttributeValue("health_check_grace_period_seconds",
					cty.NumberIntVal(int64(ecs.HealthCheckGracePeriodSeconds)))
			}
			ecsBody.SetAttributeValue("old_task_definition_buffer_size",
				cty.NumberIntVal(int64(ecs.OldTaskDefinitionBufferSize)))
			ecsBody.SetAttributeValue("is_target_group_only",
				cty.BoolVal(ecs.IsTargetGroupOnly))
			if len(ecs.DNSPrfx) > 0 {
				ecsBody.SetAttributeValue("dns_prfx",
					cty.StringVal(ecs.DNSPrfx))
			}

			for _, serviceConfig := range *ecs.LBConfigurations {
				lbConfigBlock := ecsBody.AppendNewBlock("load_balancer",
					nil)
				lbConfigBlockBody := lbConfigBlock.Body()

				lbConfigBlockBody.SetAttributeValue("target_group_count",
					cty.NumberIntVal(int64(serviceConfig.TgCount)))
				lbConfigBlockBody.SetAttributeValue("lb_type",
					cty.NumberIntVal(int64(serviceConfig.LbType)))
				lbConfigBlockBody.SetAttributeValue("is_internal",
					cty.BoolVal(serviceConfig.IsInternal))
				port, err := strconv.Atoi(serviceConfig.Port)
				if err != nil {
					fmt.Println(err)
					return
				}
				lbConfigBlockBody.SetAttributeValue("port",
					cty.NumberIntVal(int64(port)))
				lbConfigBlockBody.SetAttributeValue("external_port",
					cty.NumberIntVal(int64(serviceConfig.ExternalPort)))
				lbConfigBlockBody.SetAttributeValue("protocol",
					cty.StringVal(serviceConfig.Protocol))
				lbConfigBlockBody.SetAttributeValue("backend_protocol",
					cty.StringVal(serviceConfig.BackendProtocol))
				lbConfigBlockBody.SetAttributeValue("health_check_url",
					cty.StringVal(serviceConfig.HealthCheckURL))
				lbConfigBlockBody.SetAttributeValue("certificate_arn",
					cty.StringVal(serviceConfig.CertificateArn))

				// TODO - Add health_check_config block
				if serviceConfig.HealthCheckConfig != nil && (serviceConfig.HealthCheckConfig.HealthyThresholdCount != 0 || serviceConfig.HealthCheckConfig.UnhealthyThresholdCount != 0 || serviceConfig.HealthCheckConfig.HealthCheckIntervalSeconds != 0 || serviceConfig.HealthCheckConfig.HealthCheckTimeoutSeconds != 0) {
					lbConfigBlockBody.AppendNewline()
					hccBlock := lbConfigBlockBody.AppendNewBlock("health_check_config",
						nil)
					hccBlockBody := hccBlock.Body()
					hccBlockBody.SetAttributeValue("healthy_threshold_count",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthyThresholdCount)))
					hccBlockBody.SetAttributeValue("unhealthy_threshold_count",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.UnhealthyThresholdCount)))
					hccBlockBody.SetAttributeValue("health_check_interval_seconds",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthCheckIntervalSeconds)))
					hccBlockBody.SetAttributeValue("health_check_timeout_seconds",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthCheckTimeoutSeconds)))
					if len(serviceConfig.HealthCheckConfig.HttpSuccessCode) > 0 {
						hccBlockBody.SetAttributeValue("http_success_code",
							cty.StringVal(serviceConfig.HealthCheckConfig.HttpSuccessCode))
					}
					if len(serviceConfig.HealthCheckConfig.GrpcSuccessCode) > 0 {
						hccBlockBody.SetAttributeValue("grpc_success_code",
							cty.StringVal(serviceConfig.HealthCheckConfig.GrpcSuccessCode))
					}
				}

				ecsBody.AppendNewline()
			}
			//}

			tfFile.Write(hclFile.Bytes())

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_ecs_task_definition." + ecs.Name,
					ResourceId:      "subscriptions/" + config.TenantId + "/EcsTaskDefinition/" + ecs.TaskDefinition,
					WorkingDir:      workingDir,
				})
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_ecs_service." + ecs.Name,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/EcsServiceApiV2/" + ecs.Name,
					WorkingDir:      workingDir,
				})
			}
		}
	}

	log.Println("[TRACE] <====== Duplo ECS TF generation done. =====>")
}