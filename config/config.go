package config

import (
	"github.com/ProjectAthenaa/sonic-core/sonic"
	"os"
	"strings"
)


var Module *sonic.Module

func init() {
	var name = "walmart"

	if podName := os.Getenv("POD_NAME"); podName != "" {
		name = strings.Split(podName, "-")[0]
	}

	fieldKey := "LOOKUP_PID"

	Module = &sonic.Module{
		Name: name,
		Fields: []*sonic.ModuleField{
			{
				Validation: "\\d+",
				Type:       sonic.FieldTypeText,
				Label:      "Product ID",
				FieldKey:   &fieldKey,
			},
		},
	}
}
