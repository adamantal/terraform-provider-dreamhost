package dreamhost

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	idParts = 3
)

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDNSRecordCreate,
		ReadContext:   resourceDNSRecordRead,
		UpdateContext: nil,
		DeleteContext: resourceDNSRecordDelete,
		Schema: map[string]*schema.Schema{
			"record": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				// TODO(adamantal): add validatation for DNS record types
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// computed values
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"editable": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDNSRecordCreate(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	api, ok := config.(*dreamhostapi.Client) // nolint:varnamelen
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	record, ok := data.Get("record").(string)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve record property for DNS record creation")
	}
	value, ok := data.Get("value").(string)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve value property for DNS record creation")
	}
	typ, ok := data.Get("type").(dreamhostapi.RecordType)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve type property for DNS record creation")
	}
	recordInput := dreamhostapi.DNSRecordInput{
		Record: record,
		Value:  value,
		Type:   typ,
	}

	err := api.AddDNSRecord(ctx, recordInput)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(recordInputToID(recordInput))

	resourceDNSRecordRead(ctx, data, config)

	return diags
}

func resourceDNSRecordRead(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	api, ok := config.(*dreamhostapi.Client)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	recordID := data.Id()
	recordInput, err := idToRecordInput(recordID)
	if err != nil {
		return diag.FromErr(err)
	}

	records, err := api.ListDNSRecords(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, record := range records {
		if record.Record == recordInput.Record &&
			record.Type == recordInput.Type &&
			record.Value == recordInput.Value {
			// record found

			data.Set("record", record.Record)
			data.Set("value", record.Value)
			data.Set("type", record.Type)

			// computed values
			data.Set("comment", record.Comment)
			data.Set("account_id", record.AccountID)
			data.Set("zone", record.Zone)
			data.Set("editable", record.Editable)

			return nil
		}
	}

	return diag.FromErr(errors.New("record not found"))
}

func resourceDNSRecordDelete(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	api, ok := config.(*dreamhostapi.Client)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	recordID := data.Id()
	recordInput, err := idToRecordInput(recordID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.RemoveDNSRecord(ctx, *recordInput)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	data.SetId("")

	return diags
}

func recordInputToID(record dreamhostapi.DNSRecordInput) string {
	return fmt.Sprintf("%s|%s|%s", record.Type, record.Record, record.Value)
}

func idToRecordInput(id string) (*dreamhostapi.DNSRecordInput, error) {
	parts := strings.Split(id, "|")
	if len(parts) != idParts {
		return nil, errors.New("could not determine record from input ID")
	}
	return &dreamhostapi.DNSRecordInput{
		Type:   dreamhostapi.RecordType(parts[0]),
		Record: parts[1],
		Value:  parts[2],
	}, nil
}
