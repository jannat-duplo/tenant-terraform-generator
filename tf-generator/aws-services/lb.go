package awsservices

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

const LB_VAR_PREFIX = "lb_"

type LoadBalancer struct {
}

func (lb *LoadBalancer) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantGetApplicationLBList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Load balancer TF generation started. =====>")
		for _, lb := range *list {
			shortName, err := extractLbShortName(client, config.TenantId, lb.Name)
			resourceName := common.GetResourceName(shortName)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			settings, err := client.TenantGetApplicationLbSettings(config.TenantId, lb.Arn)
			if err != nil {
				fmt.Println(err)
				settings = nil
			}
			log.Printf("[TRACE] Generating terraform config for duplo aws load balancer : %s", shortName)

			varFullPrefix := LB_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "lb-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			rootBody := hclFile.Body()

			lbBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_load_balancer",
					resourceName})
			lbBody := lbBlock.Body()
			lbBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			lbBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			lbBody.SetAttributeValue("enable_access_logs",
				cty.BoolVal(lb.EnableAccessLogs))
			lbBody.SetAttributeValue("is_internal",
				cty.BoolVal(lb.IsInternal))

			if lb.LbType != nil {
				lbBody.SetAttributeValue("load_balancer_type",
					cty.StringVal(lb.LbType.Value))
			}

			if settings != nil {
				lbBody.SetAttributeValue("drop_invalid_headers",
					cty.BoolVal(settings.DropInvalidHeaders))
				if len(settings.WebACLID) > 0 {
					lbBody.SetAttributeValue("web_acl_id",
						cty.StringVal(settings.WebACLID))
				}
			}

			// Fetch all listeners
			listeners, clientErr := client.TenantListApplicationLbListeners(config.TenantId, shortName)
			if clientErr != nil {
				fmt.Println(err)
				listeners = nil
			}
			rootBody.AppendNewline()
			log.Printf("[TRACE] Terraform config is generation started for duplo aws load balancer listener : %s", shortName)
			if listeners != nil {
				for _, listener := range *listeners {
					listenerBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_load_balancer_listener",
							resourceName + "_listener_" + strconv.Itoa(listener.Port)})
					listenerBody := listenerBlock.Body()

					listenerBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})
					listenerBody.SetAttributeTraversal("load_balancer_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_load_balancer",
						},
						hcl.TraverseAttr{
							Name: resourceName + ".name",
						},
					})

					listenerBody.SetAttributeValue("protocol",
						cty.StringVal(listener.Protocol.Value))
					listenerBody.SetAttributeValue("port",
						cty.NumberIntVal(int64(listener.Port)))

					if len(listener.DefaultActions) > 0 {
						listenerBody.SetAttributeValue("target_group_arn",
							cty.StringVal(listener.DefaultActions[0].TargetGroupArn))
					}
					rootBody.AppendNewline()

					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_aws_load_balancer_listener." + resourceName + "_listener_" + strconv.Itoa(listener.Port),
						ResourceId:      config.TenantId + "/" + shortName + "/" + listener.ListenerArn,
						WorkingDir:      workingDir,
					})

					getReq := duplosdk.DuploTargetGroupAttributesGetReq{
						TargetGroupArn: listener.DefaultActions[0].TargetGroupArn,
					}
					targetGrpAttrs, _ := client.DuploAwsTargetGroupAttributesGet(config.TenantId, getReq)
					if targetGrpAttrs != nil && len(*targetGrpAttrs) > 0 {
						tgAttrBlock := rootBody.AppendNewBlock("resource",
							[]string{"duplocloud_aws_target_group_attributes",
								resourceName + "_listener_" + strconv.Itoa(listener.Port) + "_tg_attributes"})
						tgAttrBody := tgAttrBlock.Body()
						tgAttrBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "local",
							},
							hcl.TraverseAttr{
								Name: "tenant_id",
							},
						})
						tgAttrBody.SetAttributeTraversal("target_group_arn", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "duplocloud_aws_load_balancer_listener." + resourceName + "_listener_" + strconv.Itoa(listener.Port),
							},
							hcl.TraverseAttr{
								Name: "target_group_arn",
							},
						})
						for _, tgAttr := range *targetGrpAttrs {
							if len(tgAttr.Key) > 0 && len(tgAttr.Value) > 0 {
								attrBlock := tgAttrBody.AppendNewBlock("dimension",
									nil)
								attrBody := attrBlock.Body()
								attrBody.SetAttributeValue("key", cty.StringVal(tgAttr.Key))
								attrBody.SetAttributeValue("value", cty.StringVal(tgAttr.Value))
							}
						}
						importConfigs = append(importConfigs, common.ImportConfig{
							ResourceAddress: "duplocloud_aws_target_group_attributes." + resourceName + "_listener_" + strconv.Itoa(listener.Port) + "_tg_attributes",
							ResourceId:      config.TenantId + "/" + listener.DefaultActions[0].TargetGroupArn,
							WorkingDir:      workingDir,
						})
					}
				}
			}

			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer listener.: %s", shortName)
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer : %s", shortName)

			outVars := generateLBOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_load_balancer." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
		tfContext.ImportConfigs = importConfigs
		log.Println("[TRACE] <====== Load balancer TF generation done. =====>")
	}

	return &tfContext, nil
}

func extractLbShortName(client *duplosdk.Client, tenantID string, fullName string) (string, error) {
	prefix, err := client.GetResourcePrefix("duplo3", tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, fullName)
	return name, nil
}

func generateLBOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".fullname",
		DescVal:       "The full name of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".arn",
		DescVal:       "The ARN of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "dns_name",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".dns_name",
		DescVal:       "The DNS name of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["dns_name"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
