package edgeone

import (
	"errors"
	"strconv"
	"strings"
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
	Priority int    `json:"Priority,omitempty"`
	RecordId string `json:"RecordId,omitempty"`
	Status   string `json:"Status,omitempty"`
	TTL      int64  `json:"TTL,omitempty"`
	Type     string `json:"Type"`
	Weight   int    `json:"Weight,omitempty"`
	ZoneId   string `json:"ZoneId"`
}

type Filter struct {
	Name   string   `json:"Name"`
	Values []string `json:"Values"`
	Fuzzy  bool     `json:"Fuzzy,omitempty"`
}

type record struct {
	Type  string
	Name  string
	Value string
	TTL   time.Duration
	MX    int
}

func (r record) libdnsRecord() (libdns.Record, error) {
	if r.Type == "MX" {
		r.Value = strconv.Itoa(r.MX) + " " + r.Value
	}
	return libdns.RR{
		Type: r.Type,
		Name: r.Name,
		Data: r.Value,
		TTL:  r.TTL,
	}.Parse()
}

func fromLibdnsRecord(zone string, r libdns.Record) record {
	rr := r.RR()

	if rr.TTL == 0 {
		rr.TTL = 600
	}
	name, _ := idna.ToASCII(strings.TrimSuffix(libdns.AbsoluteName(rr.Name, zone), "."))

	return record{
		Type:  rr.Type,
		Name:  name,
		Value: rr.Data,
		TTL:   rr.TTL,
	}
}
