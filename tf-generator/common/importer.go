package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Importer struct {
}

type ImportConfig struct {
	ResourceAddress string
	ResourceId      string
	WorkingDir      string
}

func (i *Importer) Import(config *Config, importConfig *ImportConfig) {
	log.Println("[TRACE] <================================== TF Import in progress. ==================================>")
	log.Printf("[TRACE] Importing terraform resource  : (%s, %s).", importConfig.ResourceAddress, importConfig.ResourceId)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.11")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(importConfig.WorkingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	err = tf.Import(context.Background(), importConfig.ResourceAddress, importConfig.ResourceId)
	if err != nil {
		log.Fatalf("error running Import: %s", err)
	}
	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	stateJson, err := json.Marshal(state.Values)
	fmt.Println(string(stateJson))

	log.Printf("[TRACE] Terraform resource (%s, %s) is imported.", importConfig.ResourceAddress, importConfig.ResourceId)
	log.Println("[TRACE] <====================================================================>")
}
