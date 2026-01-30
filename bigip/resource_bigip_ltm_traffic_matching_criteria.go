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

func resourceBigipLtmTrafficMatchingCriteria() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipLtmTrafficMatchingCriteriaCreate,
		ReadContext:   resourceBigipLtmTrafficMatchingCriteriaRead,
		UpdateContext: resourceBigipLtmTrafficMatchingCriteriaUpdate,
		DeleteContext: resourceBigipLtmTrafficMatchingCriteriaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the Traffic Matching Criteria",
				ForceNew:     true,
				ValidateFunc: validateF5NameWithDirectory,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Traffic Matching Criteria",
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The transport protocol used in the Traffic Matching Criteria",
			},
			"route_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The route domain used in the Traffic Matching Criteria",
			},
			"destination_address_list": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of an address list shared object to use for destination routing",
				ConflictsWith: []string{"destination_address_inline"},
			},
			"destination_address_inline": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "A specific address to use for destination routing",
				ConflictsWith: []string{"destination_address_list"},
			},
			"destination_port_list": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of an port list shared object to use for destiantion port routing",
				ConflictsWith: []string{"destination_port_inline"},
			},
			"destination_port_inline": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "A specific port to use for destination port routing",
				ConflictsWith: []string{"destination_port_list"},
			},
			"source_address_list": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of an address list shared object for source address matching",
				ConflictsWith: []string{"source_address_inline"},
			},
			"source_address_inline": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "A specific address and optional mask to use for source matching",
				ConflictsWith: []string{"source_address_list"},
			},
			"source_port_inline": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "A specific port to use for source port matching",
			},
		},
	}
}

func resourceBigipLtmTrafficMatchingCriteriaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Println("[INFO] Creating traffic matching criteria" + name)
	tmc := NewTmcFromResourceData(d)

	err := client.AddTrafficMatchingCriteria(ctx, tmc)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating traffic matching criteria (%s): %s", name, err))
	}

	d.SetId(name)

	return resourceBigipLtmTrafficMatchingCriteriaRead(ctx, d, meta)
}

func resourceBigipLtmTrafficMatchingCriteriaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	_ = d.Set("name", name)

	log.Println("[INFO] Reading traffic matching criteria " + name)
	tmc, err := client.GetTrafficMatchingCriteria(ctx, name)
	if err != nil && strings.Contains(err.Error(), "not found") {
		log.Printf("[WARN] Traffic Matching Criteria (%s) not found, removing from state", name)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("description", tmc.Description)
	_ = d.Set("protocol", tmc.Protocol)
	_ = d.Set("route_domain", tmc.RouteDomain)
	_ = d.Set("destination_address_list", tmc.DestinationAddressList)
	_ = d.Set("destination_address_inline", tmc.DestinationAddressInline)
	_ = d.Set("destination_port_list", tmc.DestinationPortList)
	_ = d.Set("destination_port_inline", tmc.DestinationPortInline)
	_ = d.Set("source_address_list", tmc.SourceAddressList)
	_ = d.Set("source_address_inline", tmc.SourceAddressInline)
	_ = d.Set("source_port_inline", tmc.SourcePortInline)

	return nil
}

func resourceBigipLtmTrafficMatchingCriteriaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	tmc := NewTmcFromResourceData(d)

	fmt.Println("[INFO] Attempting to update traffic matching criteria ", name)
	err := client.ModifyTrafficMatchingCriteria(ctx, name, tmc)
	if err != nil {
		fmt.Printf("[ERROR] Failed to update traffic matching criteria (%s) (%v)", name, err)
		return diag.FromErr(err)
	}

	return resourceBigipLtmTrafficMatchingCriteriaRead(ctx, d, meta)
}

func resourceBigipLtmTrafficMatchingCriteriaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)
	name := d.Id()

	fmt.Println("[INFO] Attempting to delete traffing matching criteria " + name)
	err := client.DeleteTrafficMatchingCriteria(ctx, name)
	if err != nil {
		fmt.Printf("[ERROR] Failed to delete traffic matching criteria (%s) (%v)", name, err)
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func NewTmcFromResourceData(d *schema.ResourceData) *bigip.TrafficMatchingCriteria {
	// Even if SourceAddressList is in use, SourceAddressInline must be 0.0.0.0
	sail := d.Get("source_address_inline").(string)
	if len(sail) == 0 {
		sail = "0.0.0.0"
	}
	return &bigip.TrafficMatchingCriteria{
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		Protocol:                 d.Get("protocol").(string),
		RouteDomain:              d.Get("route_domain").(string),
		DestinationAddressList:   d.Get("destination_address_list").(string),
		DestinationAddressInline: d.Get("destination_address_inline").(string),
		DestinationPortList:      d.Get("destination_port_list").(string),
		DestinationPortInline:    d.Get("destination_port_inline").(string),
		SourceAddressList:        d.Get("source_address_list").(string),
		SourceAddressInline:      d.Get("source_address_inline").(string),
		SourcePortInline:         d.Get("source_port_inline").(int),
	}
}
