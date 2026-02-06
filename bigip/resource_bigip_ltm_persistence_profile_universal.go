package bigip

import (
	"context"
	"log"
	"strconv"

	bigip "github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBigipLtmPersistenceProfileUniversal() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigipLtmPersistenceProfileUniversalCreate,
		ReadContext:   resourceBigipLtmPersistenceProfileUniversalRead,
		UpdateContext: resourceBigipLtmPersistenceProfileUniversalUpdate,
		DeleteContext: resourceBigipLtmPersistenceProfileUniversalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the persistence profile",
				ValidateFunc: validateF5Name,
			},

			"app_service": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"defaults_from": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Inherit defaults from parent profile",
				ValidateFunc: validateF5Name,
			},

			"match_across_pools": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "To enable _ disable match across pools with given persistence record",
			},

			"match_across_services": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "To enable _ disable match across services with given persistence record",
			},

			"match_across_virtuals": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "To enable _ disable match across virtual servers with given persistence record",
			},
			"method": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Specifies the type of cookie processing that the system uses",
			},

			"mirror": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "To enable _ disable",
			},

			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Timeout for persistence of the session",
			},

			"override_conn_limit": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "To enable _ disable that pool member connection limits are overridden for persisted clients. Per-virtual connection limits remain hard limits and are not overridden.",
			},

			// Specific to UniversalPersistenceProfile
			"rule": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The irule to use with this profile",
			},
		},
	}
}

func resourceBigipLtmPersistenceProfileUniversalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Get("name").(string)
	parent := d.Get("defaults_from").(string)

	err := client.CreateUniversalPersistenceProfile(name, parent)

	if err != nil {
		log.Printf("[ERROR] Unable to Create Universal Persistence Profile %s %v :", name, err)
		return diag.FromErr(err)
	}

	d.SetId(name)

	return resourceBigipLtmPersistenceProfileUniversalUpdate(ctx, d, meta)
}

func resourceBigipLtmPersistenceProfileUniversalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Println("[INFO] Fetching Cookie Persistence Profile " + name)

	pp, err := client.GetUniversalPersistenceProfile(name)
	if err != nil {
		log.Printf("[ERROR] Unable to retrieve Cookie Persistence Profile %s  %v : ", name, err)
		return diag.FromErr(err)
	}
	if pp == nil {
		log.Printf("[WARN] Cookie Persistence Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	_ = d.Set("name", name)
	_ = d.Set("defaults_from", pp.DefaultsFrom)
	_ = d.Set("match_across_pools", pp.MatchAcrossPools)
	_ = d.Set("match_across_services", pp.MatchAcrossServices)
	_ = d.Set("match_across_virtuals", pp.MatchAcrossVirtuals)
	_ = d.Set("mirror", pp.Mirror)
	_ = d.Set("override_conn_limit", pp.OverrideConnectionLimit)
	_ = d.Set("method", pp.Method)
	if timeout, err := strconv.Atoi(pp.Timeout); err == nil {
		_ = d.Set("timeout", timeout)
	}

	if _, ok := d.GetOk("app_service"); ok {
		_ = d.Set("app_service", pp.AppService)
	}
	// Specific to UniversalPersistenceProfile
	if _, ok := d.GetOk("rule"); ok {
		_ = d.Set("rule", pp.Rule)
	}

	return nil
}

func resourceBigipLtmPersistenceProfileUniversalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	pp := &bigip.UniversalPersistenceProfile{
		PersistenceProfile: bigip.PersistenceProfile{
			AppService:          d.Get("app_service").(string),
			DefaultsFrom:        d.Get("defaults_from").(string),
			MatchAcrossPools:    d.Get("match_across_pools").(string),
			MatchAcrossServices: d.Get("match_across_services").(string),
			MatchAcrossVirtuals: d.Get("match_across_virtuals").(string),
			Mirror:              d.Get("mirror").(string),
			//  Method:                  d.Get("method").(string),
			OverrideConnectionLimit: d.Get("override_conn_limit").(string),
			Timeout:                 strconv.Itoa(d.Get("timeout").(int)),
		},
		// Specific to UniversalPersistenceProfile
		Rule: d.Get("rule").(string),
	}

	err := client.ModifyUniversalPersistenceProfile(name, pp)
	if err != nil {
		log.Printf("[ERROR] Unable to Modify Universal Persistence Profile %s %v ", name, err)
		if errdel := client.DeleteUniversalPersistenceProfile(name); errdel != nil {
			return diag.FromErr(errdel)
		}
		return diag.FromErr(err)
	}
	return resourceBigipLtmPersistenceProfileUniversalRead(ctx, d, meta)
}

func resourceBigipLtmPersistenceProfileUniversalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Deleting Universal Persistence Profile " + name)
	err := client.DeleteUniversalPersistenceProfile(name)
	if err != nil {
		log.Printf("[ERROR] Unable to Delete Universal Persistence Profile %s  %v : ", name, err)
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
