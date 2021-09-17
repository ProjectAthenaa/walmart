package config

import (
	"github.com/ProjectAthenaa/sonic-core/sonic"
	"github.com/ProjectAthenaa/sonic-core/sonic/database/ent/product"
)


var Module *sonic.Module

func init() {

	fieldKey := "LOOKUP_PID"

	Module = &sonic.Module{
		Name: string(product.SiteWalmart),
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
