package edgeone

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const (
	endpoint = "https://teo.tencentcloudapi.com"

	DescribeZones      = "DescribeZones"
	DescribeDnsRecords = "DescribeDnsRecords"
	CreateDnsRecord    = "CreateDnsRecord"
	ModifyDnsRecords   = "ModifyDnsRecords"
	DeleteDnsRecords   = "DeleteDnsRecords"
)

func (p *Provider) listRecords(ctx context.Context, zoneId string, zone string) ([]libdns.Record, error) {
	requestData := DescribeDnsRecordsRequest{
		ZoneId: zoneId,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	resp, err := p.sendRequest(ctx, DescribeDnsRecords, string(payload))
	if err != nil {
		return nil, err
	}

	var response DescribeDnsRecordsResponse
	if err = json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if response.Response.Error != nil {
		err = errors.New(response.Response.Error.Message)
		return nil, err
	}

	list := make([]libdns.Record, 0, len(response.Response.DnsRecords))
	for _, txRecord := range response.Response.DnsRecords {
		rr := record{
			Type:  txRecord.Type,
			Name:  libdns.RelativeName(txRecord.Name, zone),
			Value: txRecord.Content,
			TTL:   time.Duration(txRecord.TTL) * time.Second,
			MX:    txRecord.Priority,
		}
		libdnsRecord, err := rr.libdnsRecord()
		if err != nil {
			return nil, err
		}
		list = append(list, libdnsRecord)
	}

	return list, nil
}

func (p *Provider) createDnsRecord(ctx context.Context, zoneId string, r record) error {
	requestData := CreateDnsRecordRequest{
		DnsRecord: DnsRecord{
			ZoneId:   zoneId,
			Name:     r.Name,
			Type:     r.Type,
			Content:  r.Value,
			Location: "Default",
			TTL:      int64(r.TTL.Seconds()),
		},
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	resp, err := p.sendRequest(ctx, CreateDnsRecord, string(payload))
	if err != nil {
		return err
	}

	var response CreateDnsRecordResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return err
	}

	if response.Response.Error != nil {
		err = errors.New(response.Response.Error.Message)
		return err
	}

	if response.Response.RecordId == "" {
		return ErrNotValid
	}

	return nil
}

func (p *Provider) modifyDnsRecords(ctx context.Context, zoneId string, domain string, recordMap map[string]libdns.Record) error {

	var dnsRecords []DnsRecord

	for id := range recordMap {
		r := fromLibdnsRecord(domain, recordMap[id])
		dnsRecords = append(dnsRecords, DnsRecord{
			RecordId: id,
			ZoneId:   zoneId,
			Name:     r.Name,
			Type:     r.Type,
			Content:  r.Value,
			Location: "Default",
			Status:   "enable",
			TTL:      int64(r.TTL.Seconds()),
		})
	}

	requestData := ModifyDnsRecordsRequest{
		ZoneId:     zoneId,
		DnsRecords: dnsRecords,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	_, err = p.sendRequest(ctx, ModifyDnsRecords, string(payload))
	return err
}

func (p *Provider) deleteDnsRecords(ctx context.Context, zoneId string, ids []string) error {
	if len(ids) <= 0 {
		return nil
	}
	requestData := DeleteDnsRecordsRequest{
		ZoneId:    zoneId,
		RecordIds: ids,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	_, err = p.sendRequest(ctx, DeleteDnsRecords, string(payload))
	return err
}

func (p *Provider) findRecord(ctx context.Context, zoneId string, r record, matchContent bool) ([]string, error) {
	filters := []Filter{
		{Name: "name", Values: []string{r.Name}},
		{Name: "type", Values: []string{r.Type}},
	}
	if matchContent {
		filters = append(filters, Filter{Name: "content", Values: []string{r.Value}})
	}
	requestData := DescribeDnsRecordsRequest{
		ZoneId:  zoneId,
		Limit:   1000,
		Filters: filters,
		SortBy:  "created-on",
	}
	payload, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	resp, err := p.sendRequest(ctx, DescribeDnsRecords, string(payload))
	if err != nil {
		return nil, err
	}

	var response DescribeDnsRecordsResponse
	if err = json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	if response.Response.Error != nil {
		err = errors.New(response.Response.Error.Message)
		return nil, err
	}

	var recordId []string
	for _, record := range response.Response.DnsRecords {
		if record.Status == "enable" {
			recordId = append(recordId, record.RecordId)
		}
	}

	return recordId, nil
}

func (p *Provider) getZoneId(ctx context.Context, zone string) (string, error) {
	domain := strings.TrimSuffix(zone, ".")
	requestData := DescribeZonesRequest{
		Filters: []Filter{{Name: "zone-name", Values: []string{domain}}},
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}

	resp, err := p.sendRequest(ctx, DescribeZones, string(payload))
	if err != nil {
		return "", err
	}

	var response DescribeZonesResponse
	if err = json.Unmarshal(resp, &response); err != nil {
		return "", err
	}
	if response.Response.Error != nil {
		err = errors.New(response.Response.Error.Message)
		return "", err
	}

	if len(response.Response.Zones) <= 0 {
		return "", errors.New("Zone not found: " + zone)
	}
	return response.Response.Zones[0].ZoneId, nil
}

func (p *Provider) sendRequest(ctx context.Context, action string, data string) ([]byte, error) {
	endpointUrl := endpoint
	if p.Region != "" {
		endpointUrl = "https://teo." + p.Region + ".tencentcloudapi.com"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpointUrl, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Version", "2022-09-01")

	SignRequest(p.SecretId, p.SecretKey, p.SessionToken, req, action, data)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
