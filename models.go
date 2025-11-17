package edgeone

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
	"golang.org/x/net/idna"
)

var ErrNotValid = errors.New("returned value is not valid")

type Provider struct {
	SecretId     string
	SecretKey    string
	SessionToken string
	Region       string

	mu sync.Mutex
}

type ModifyDnsRecordsRequest struct {
	ZoneId     string      `json:"ZoneId"`
	DnsRecords []DnsRecord `json:"DnsRecords,omitempty"`
}

type CreateDnsRecordRequest struct {
	DnsRecord
}

type DescribeDnsRecordsRequest struct {
	ZoneId  string   `json:"ZoneId"`
	Limit   int64    `json:"Limit,omitempty"`
	Filters []Filter `json:"Filters,omitempty"`
	SortBy  string   `json:"SortBy,omitempty"`
}

type DeleteDnsRecordsRequest struct {
	ZoneId    string   `json:"ZoneId"`
	RecordIds []string `json:"RecordIds"`
}

type DescribeZonesRequest struct {
	Filters []Filter `json:"Filters,omitempty"`
}

type DescribeZonesResponse struct {
	Response struct {
		Error *Error `json:"Error,omitempty"`
		Zones []struct {
			ZoneId string `json:"ZoneId"`
		} `json:"Zones"`
	}
}

type DescribeDnsRecordsResponse struct {
	Response struct {
		Error      *Error      `json:"Error,omitempty"`
		DnsRecords []DnsRecord `json:"DnsRecords"`
	}
}

type CreateDnsRecordResponse struct {
	Response struct {
		Error    *Error `json:"Error,omitempty"`
		RecordId string `json:"RecordId"`
	}
}

type Error struct {
	Code    string
	Message string
}

type DnsRecord struct {
	Content  string `json:"Content"`
	Location string `json:"Location,omitempty"`
	Name     string `json:"Name"`
	Priority int64  `json:"Priority,omitempty"`
	RecordId string `json:"RecordId,omitempty"`
	Status   string `json:"Status,omitempty"`
	TTL      int64  `json:"TTL,omitempty"`
	Type     string `json:"Type"`
	Weight   int64  `json:"Weight,omitempty"`
	ZoneId   string `json:"ZoneId"`
}

type Filter struct {
	Name   string   `json:"Name"`
	Values []string `json:"Values"`
	Fuzzy  bool     `json:"Fuzzy,omitempty"`
}

func (r DnsRecord) libdnsRecord(zone string) (libdns.Record, error) {
	name := libdns.RelativeName(r.Name, zone)
	ttl := time.Duration(r.TTL) * time.Second
	switch r.Type {
	case "A", "AAAA":
		addr, err := netip.ParseAddr(r.Content)
		if err != nil {
			return libdns.Address{}, fmt.Errorf("invalid IP address %q: %v", r.Content, err)
		}
		return libdns.Address{
			Name: name,
			TTL:  ttl,
			IP:   addr,
		}, nil
	case "CNAME":
		return libdns.CNAME{
			Name:   name,
			TTL:    ttl,
			Target: r.Content,
		}, nil
	case "MX":
		return libdns.MX{
			Name:       name,
			TTL:        ttl,
			Preference: uint16(r.Priority),
			Target:     r.Content,
		}, nil
	case "NS":
		return libdns.NS{
			Name:   name,
			TTL:    ttl,
			Target: r.Content,
		}, nil
	case "TXT":
		return libdns.TXT{
			Name: name,
			TTL:  ttl,
			Text: r.Content,
		}, nil
	default:
		return libdns.RR{
			Type: r.Type,
			Name: name,
			Data: r.Content,
			TTL:  ttl,
		}.Parse()
	}
}

func edgeOneRecord(zone string, r libdns.Record) DnsRecord {
	rr := r.RR()
	var ttl time.Duration
	if rr.TTL == 0 {
		ttl = 300 * time.Second
	} else {
		ttl = max(min(rr.TTL, 86400*time.Second), 60*time.Second)
	}
	name, _ := idna.ToASCII(strings.TrimSuffix(libdns.AbsoluteName(rr.Name, zone), "."))

	record := DnsRecord{
		Name:     name,
		Type:     rr.Type,
		Content:  rr.Data,
		Location: "Default",
		TTL:      int64(ttl.Seconds()),
	}
	switch rec := r.(type) {
	case libdns.MX:
		record.Priority = int64(rec.Preference)
	}
	return record
}
