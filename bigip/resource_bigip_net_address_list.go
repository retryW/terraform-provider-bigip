/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"context"
	"fmt"
	"log"
	"strings"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBigipNetAddressList() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipNetAddressListCreate,
		ReadContext:   resourceBigipNetAddressListRead,
		UpdateContext: resourceBigipNetAddressListUpdate,
		DeleteContext: resourceBigipNetAddressListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the address list",
				ForceNew:     true,
				ValidateFunc: validateF5NameWithDirectory,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the address list",
			},
			"addresses": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Addresses with mask",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceBigipNetAddressListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Println("[INFO] Creating address list" + name)
	addressList := addressListFromConfig(d)

	err := client.AddAddressList(ctx, addressList)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating address list (%s): %s", name, err))
	}

	return resourceBigipNetAddressListRead(ctx, d, meta)
}

func resourceBigipNetAddressListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	_ = d.Set("name", name)

	log.Println("[INFO] Reading address list " + name)
	addressList, err := client.GetAddressList(ctx, name)
	if err != nil && strings.Contains(err.Error(), "not found") {
		log.Printf("[WARN] address list (%s) not found, removing from state", name)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("description", addressList.Description)
	_ = d.Set("addresses", addressList.Addresses)

	return nil
}

func resourceBigipNetAddressListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	addressList := &bigip.AddressList{
		Description: d.Get("description").(string),
		Addresses:   listToAddressesListAddresses(d.Get("addresses").([]interface{})),
	}

	fmt.Println("[INFO] Attempting to update address list ", name)
	err := client.ModifyAddressList(ctx, name, addressList)
	if err != nil {
		fmt.Printf("[ERROR] Failed to update address list (%s) (%v)", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipNetAddressListRead(ctx, d, meta)
}

func resourceBigipNetAddressListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	fmt.Println("[INFO] Attempting to delete traffing matching criteria " + name)
	err := client.DeleteAddressList(ctx, name)
	if err != nil {
		fmt.Printf("[ERROR] Failed to delete address list (%s) (%v)", name, err)
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func addressListFromConfig(d *schema.ResourceData) *bigip.AddressList {
	al := &bigip.AddressList{
		Name:        d.Id(),
		Description: d.Get("description").(string),
	}

	if p, ok := d.GetOk("addresses"); ok {
		al.Addresses = listToAddressesListAddresses(p.([]interface{}))
	}

	return al
}

func listToAddressesListAddresses(list []interface{}) []bigip.AddressListAddress {
	ala := make([]bigip.AddressListAddress, len(list))
	for i, v := range list {
		ala[i].Name = v.(string)
	}

	return ala
}
