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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const REDIS_VAR_PREFIX = "redis_"

type Redis struct {
}

func (r *Redis) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.EcacheInstanceList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Redis TF generation started. =====>")
		kms, kmsClientErr := client.TenantGetTenantKmsKey(config.TenantId)
		for _, redis := range *list {
			shortName := redis.Identifier[len("duplo-"):len(redis.Identifier)]
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo Redis Instance : %s", redis.Identifier)

			varFullPrefix := REDIS_VAR_PREFIX + resourceName + "_"
			inputVars := generateRedisVars(redis, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "redis-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			redisBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecache_instance",
					resourceName})
			redisBody := redisBlock.Body()
			redisBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// redisBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			name := shortName + "-${local.tenant_name}"
			redisNameTokens := hclwrite.Tokens{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenIdent, Bytes: []byte(name)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
			}
			// redisBody.SetAttributeValue("name",
			// 	cty.StringVal(shortName+"-"+config.TenantName))
			redisBody.SetAttributeRaw("name", redisNameTokens)

			redisBody.SetAttributeValue("cache_type",
				cty.NumberIntVal(int64(0)))
			redisBody.SetAttributeTraversal("replicas", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "replicas",
				},
			})
			redisBody.SetAttributeTraversal("size", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "size",
				},
			})
			redisBody.SetAttributeValue("encryption_at_rest",
				cty.BoolVal(redis.EncryptionAtRest))
			redisBody.SetAttributeValue("encryption_in_transit",
				cty.BoolVal(redis.EncryptionInTransit))
			if len(redis.AuthToken) > 0 {
				redisBody.SetAttributeValue("auth_token",
					cty.StringVal(redis.AuthToken))
			}
			if len(redis.KMSKeyID) > 0 {
				if kms != nil && kmsClientErr == nil && (redis.KMSKeyID == kms.KeyArn || redis.KMSKeyID == kms.KeyID) {
					redisBody.SetAttributeTraversal("kms_key_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "data.duplocloud_tenant_aws_kms_key.tenant_kms",
						},
						hcl.TraverseAttr{
							Name: "key_id",
						},
					})
				} else {
					redisBody.SetAttributeValue("kms_key_id",
						cty.StringVal(redis.KMSKeyID))
				}
			}

			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo redis instance : %s", redis.Identifier)

			outVars := generateRedisOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_ecache_instance." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ECacheDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== redis TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateRedisVars(duplo duplosdk.DuploEcacheInstance, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "replicas",
		DefaultVal: strconv.Itoa(duplo.Replicas),
		TypeVal:    "number",
	}
	varConfigs["replicas"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "size",
		DefaultVal: duplo.Size,
		TypeVal:    "string",
	}
	varConfigs["size"] = var2

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateRedisOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_ecache_instance." + resourceName + ".identifier",
		DescVal:       "The full name of the elasticache instance.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_ecache_instance." + resourceName + ".arn",
		DescVal:       "The ARN of the elasticache instance.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "endpoint",
		ActualVal:     "duplocloud_ecache_instance." + resourceName + ".endpoint",
		DescVal:       "The endpoint of the elasticache instance.",
		RootTraversal: true,
	}
	outVarConfigs["endpoint"] = var3

	var4 := common.OutputVarConfig{
		Name:          prefix + "host",
		ActualVal:     "duplocloud_ecache_instance." + resourceName + ".host",
		DescVal:       "The DNS hostname of the elasticache instance.",
		RootTraversal: true,
	}
	outVarConfigs["host"] = var4

	var5 := common.OutputVarConfig{
		Name:          prefix + "port",
		ActualVal:     "duplocloud_ecache_instance." + resourceName + ".port",
		DescVal:       "The listening port of the elasticache instance.",
		RootTraversal: true,
	}
	outVarConfigs["port"] = var5

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
