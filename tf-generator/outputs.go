package tfgenerator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type OutputVarConfig struct {
	Name          string
	ActualVal     string
	DescVal       string
	RootTraversal bool
}

type OutputVars struct {
	targetLocation string
	outputVars     []OutputVarConfig
}

func (ov *OutputVars) Generate() {
	log.Println("[TRACE] <====== Output Variables TF generation started. =====>")

	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()
	// create new file on system
	path := filepath.Join(ov.targetLocation, "outputs.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// initialize the body of the new file object
	rootBody := hclFile.Body()
	for _, outVarConfig := range ov.outputVars {
		if len(outVarConfig.Name) > 0 {
			outputVarblock := rootBody.AppendNewBlock("output",
				[]string{outVarConfig.Name})
			outputVarBody := outputVarblock.Body()
			if len(outVarConfig.ActualVal) > 0 {
				if outVarConfig.RootTraversal {
					outputVarBody.SetAttributeTraversal("value", hcl.Traversal{
						hcl.TraverseRoot{
							Name: outVarConfig.ActualVal,
						},
					})
				} else {
					outputVarBody.SetAttributeValue("value",
						cty.StringVal(outVarConfig.ActualVal))
				}
			}

			if len(outVarConfig.DescVal) > 0 {
				outputVarBody.SetAttributeValue("description",
					cty.StringVal(outVarConfig.DescVal))
			}
			rootBody.AppendNewline()
		}

	}

	fmt.Printf("%s", hclFile.Bytes())
	tfFile.Write(hclFile.Bytes())
	log.Println("[TRACE] <====== Output Variables TF generation done. =====>")
}
