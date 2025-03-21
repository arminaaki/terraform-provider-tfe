package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOrganizationModuleSharing() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEOrganizationModuleSharingCreate,
		Read:   resourceTFEOrganizationModuleSharingRead,
		Update: resourceTFEOrganizationModuleSharingUpdate,
		Delete: resourceTFEOrganizationModuleSharingDelete,
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},

			"module_consumers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceTFEOrganizationModuleSharingCreate(d *schema.ResourceData, meta interface{}) error {
	// Get the organization name that will share "produce" modules
	producer := d.Get("organization").(string)

	log.Printf("[DEBUG] Create %s module consumers", producer)
	d.SetId(producer)

	return resourceTFEOrganizationModuleSharingUpdate(d, meta)
}

func resourceTFEOrganizationModuleSharingUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	var consumers []string
	for _, name := range d.Get("module_consumers").([]interface{}) {
		// ignore empty strings
		if name == nil {
			continue
		}
		consumers = append(consumers, name.(string))
	}

	log.Printf("[DEBUG] Update %s module consumers", d.Id())
	err := tfeClient.Admin.Organizations.UpdateModuleConsumers(ctx, d.Id(), consumers)
	if err != nil {
		return fmt.Errorf("error updating module consumers to %s: %w", d.Id(), err)
	}

	return resourceTFEOrganizationModuleSharingRead(d, meta)
}

func resourceTFEOrganizationModuleSharingRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	options := tfe.AdminOrganizationListModuleConsumersOptions{}

	log.Printf("[DEBUG] Read configuration of module sharing for organization: %s", d.Id())
	for {
		consumerList, err := tfeClient.Admin.Organizations.ListModuleConsumers(ctx, d.Id(), options)
		if err != nil {
			if err == tfe.ErrResourceNotFound {
				log.Printf("[DEBUG] Organization %s does not longer exist", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error reading organization %s module consumer list: %w", d.Id(), err)
		}

		if consumerList.CurrentPage >= consumerList.TotalPages {
			break
		}

		options.PageNumber = consumerList.NextPage
	}

	return nil
}

func resourceTFEOrganizationModuleSharingDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Disable module sharing for organization: %s", d.Id())
	err := tfeClient.Admin.Organizations.UpdateModuleConsumers(ctx, d.Id(), []string{})
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete module sharing for organization %s: %w", d.Id(), err)
	}

	return nil
}
