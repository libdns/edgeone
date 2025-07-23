package edgeone

import (
	"context"

	"github.com/libdns/libdns"
)

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zoneId, err := p.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}
	return p.listRecords(ctx, zoneId, zone)
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		r := fromLibdnsRecord(zone, record)
		if err := p.createDnsRecord(ctx, zoneId, r); err != nil {
			return nil, err
		}
	}

	return records, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}
	recordMap := make(map[string]libdns.Record)
	p_matches := make(map[string]int)
	matches := make(map[string][]string)
	for _, record := range records {
		r := fromLibdnsRecord(zone, record)
		if _, ok := matches[r.Name]; !ok {
			matches[r.Name], err = p.findRecord(ctx, zoneId, r, false)
			p_matches[r.Name] = len(matches[r.Name])
		}
		if err != nil {
			return nil, err
		}
		if p_matches[r.Name] > 0 {
			p_matches[r.Name] = p_matches[r.Name] - 1
			recordMap[matches[r.Name][p_matches[r.Name]]] = record
		} else if err = p.createDnsRecord(ctx, zoneId, r); err != nil {
			return nil, err
		}
	}

	if err = p.modifyDnsRecords(ctx, zoneId, zone, recordMap); err != nil {
		return nil, err
	}

	return records, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	var ids []string
	matches := make(map[string][]string)
	for _, record := range records {
		r := fromLibdnsRecord(zone, record)
		item := r.Name
		matchContent := record.RR().Data != ""
		if matchContent {
			item = r.Name + r.Value
		}
		if _, ok := matches[item]; !ok {
			matches[item], err = p.findRecord(ctx, zoneId, r, matchContent)
		}
		if err != nil {
			return nil, err
		}
	}
	for s := range matches {
		ids = append(ids, matches[s]...)
	}
	if len(ids) > 0 {
		if err := p.deleteDnsRecords(ctx, zoneId, ids); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
