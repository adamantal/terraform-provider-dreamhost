package dreamhost

import (
	"context"
	"errors"
	"sync"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
)

type cachedDreamhostClient struct {
	sync.RWMutex

	client        *dreamhostapi.Client
	cachedRecords []dreamhostapi.DNSRecord
}

func newDreamhostClient(client *dreamhostapi.Client) *cachedDreamhostClient {
	return &cachedDreamhostClient{
		client: client,
	}
}

func (c *cachedDreamhostClient) AddDNSRecord(ctx context.Context, recordInput dreamhostapi.DNSRecordInput) error {
	return c.client.AddDNSRecord(ctx, recordInput)
}

func (c *cachedDreamhostClient) GetDNSRecord(
	ctx context.Context, recordInput dreamhostapi.DNSRecordInput, enableCache bool,
) (*dreamhostapi.DNSRecord, error) {
	if enableCache && c.cachedRecords != nil {
		for _, record := range c.cachedRecords {
			if record.Record == recordInput.Record &&
				record.Type == recordInput.Type &&
				(record.Value == recordInput.Value || record.Value+"." == recordInput.Value) {
				// record found
				return &record, nil
			}
		}
		return nil, errors.New("record not found")
	}
	if _, err := c.ListDNSRecords(ctx); err != nil {
		return nil, errors.New("failed to refresh cache")
	}
	return c.GetDNSRecord(ctx, recordInput, true)
}

func (c *cachedDreamhostClient) ListDNSRecords(ctx context.Context) ([]dreamhostapi.DNSRecord, error) {
	records, err := c.client.ListDNSRecords(ctx)
	if err != nil {
		return nil, err
	}

	c.cachedRecords = records
	return records, err
}

func (c *cachedDreamhostClient) RemoveDNSRecord(ctx context.Context, recordInput dreamhostapi.DNSRecordInput) error {
	return c.client.RemoveDNSRecord(ctx, recordInput)
}
