package dreamhost

import (
	"context"
	"fmt"
	"strings"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "the name of the DNS record",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "the value of the DNS record",
			},
			"type": {
				// TODO(adamantal): add validatation for DNS record types
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "the type of the DNS record (e.g. A, CNAME, TXT)",
			},

			// computed values
			"comment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "any comment attached to the DNS record",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the account ID belonging to the DNS record",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the zone of the DNS record (used in a multi-zone setup)",
			},
			"editable": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "whether the record is editable",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDNSRecordCreate(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	api, ok := config.(*cachedDreamhostClient) // nolint:varnamelen
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	record, ok := data.Get("record").(string)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve record property for DNS record creation")
	}
	typ, ok := data.Get("type").(string)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve type property for DNS record creation")
	}
	actualType := dreamhostapi.RecordType(typ)
	value, ok := data.Get("value").(string)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve value property for DNS record creation")
	}
	// workaround: Dreamhost would do this anyways, let's save the resource adding the trailing dot in Terrafor
	if actualType == dreamhostapi.CNAMERecordType {
		value += "."
	}
	recordInput := dreamhostapi.DNSRecordInput{
		Record: record,
		Value:  value,
		Type:   dreamhostapi.RecordType(typ),
	}

	err := api.AddDNSRecord(ctx, recordInput)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(recordInputToID(recordInput))

	dnsRecord, err := api.GetDNSRecord(ctx, recordInput, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if dnsRecord == nil {
		return diag.Errorf("API error - failed to create DNS record")
	}
	if err := refreshDataFromRecord(data, *dnsRecord); err != nil {
		return diag.Errorf("failed to refresh data from record")
	}

	return diags
}

func resourceDNSRecordRead(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	api, ok := config.(*cachedDreamhostClient)
	if !ok {
		return diag.Errorf("internal error: failed to retrieve dreamhost API client")
	}

	recordID := data.Id()
	recordInput, err := idToRecordInput(recordID)
	if err != nil {
		return diag.FromErr(err)
	}

	record, err := api.GetDNSRecord(ctx, *recordInput, true)
	if err != nil {
		return diag.FromErr(err)
	}

	// record is completely missing
	if !data.IsNewResource() && record == nil {
		data.SetId("")
		return diags
	}

	// record is found, refresh data
	if err := refreshDataFromRecord(data, *record); err != nil {
		return diag.Errorf("failed to refresh data from record")
	}

	return diags
}

func refreshDataFromRecord(data *schema.ResourceData, record dreamhostapi.DNSRecord) error {
	if err := data.Set("record", record.Record); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	if err := data.Set("value", record.Value); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	if err := data.Set("type", record.Type); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}

	// computed values
	if err := data.Set("comment", record.Comment); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	if err := data.Set("account_id", record.AccountID); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	if err := data.Set("zone", record.Zone); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	if err := data.Set("editable", record.Editable); err != nil {
		return errors.Wrap(err, "failed to set field `record`")
	}
	return nil
}

func resourceDNSRecordDelete(ctx context.Context, data *schema.ResourceData, config interface{}) diag.Diagnostics {
	api, ok := config.(*cachedDreamhostClient)
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
