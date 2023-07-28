package dreamhost

import (
	"context"
	"sync"

	dreamhostapi "github.com/adamantal/go-dreamhost/api"
	"github.com/pkg/errors"
)

type cache struct {
	sync.Mutex

	cachedRecords []dreamhostapi.DNSRecord
}

func (c *cache) GetRecords(ctx context.Context, client *cachedDreamhostClient) ([]dreamhostapi.DNSRecord, error) {
	c.Lock()
	defer c.Unlock()

	if c.cachedRecords == nil {
		records, err := client.ListDNSRecords(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list DNS records")
		}
		c.cachedRecords = records
	}

	return c.cachedRecords, nil
}
