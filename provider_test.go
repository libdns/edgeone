package edgeone

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/libdns/libdns"
)

var provider = &Provider{
	SecretId:  os.Getenv("TC_SECRET_ID"),
	SecretKey: os.Getenv("TC_SECRET_KEY"),
}

var (
	zone = os.Getenv("TC_ZONE")
)

func TestAppendRecords(t *testing.T) {
	recs, err := provider.AppendRecords(context.Background(), zone, []libdns.Record{
		libdns.RR{Name: "one", TTL: 10 * time.Minute, Type: "A", Data: "1.1.1.1"},
	})
	if err != nil {
		t.Fatalf("AppendRecords: %v", err)
	}
	fmt.Println("AppendRecords:", recs)
}

func TestSetRecords(t *testing.T) {
	recs, err := provider.SetRecords(context.Background(), zone, []libdns.Record{
		libdns.RR{Name: "one", TTL: 10 * time.Minute, Type: "A", Data: "1.1.1.2"},
		libdns.TXT{Name: "one3", TTL: 10 * time.Minute, Text: "hello world"},
	})
	if err != nil {
		t.Fatalf("SetRecords: %v", err)
	}
	fmt.Println("SetRecords:", recs)
}

func TestGetRecords(t *testing.T) {
	recs, err := provider.GetRecords(context.Background(), zone)
	if err != nil {
		t.Fatalf("GetRecords: %v", err)
	}
	fmt.Println("GetRecords:", recs)
}

func TestDeleteRecords(t *testing.T) {
	recs, err := provider.DeleteRecords(context.Background(), zone, []libdns.Record{
		libdns.RR{Name: "two", TTL: 10 * time.Minute, Type: "AAAA", Data: "2606:4700:4700::1111"},
		libdns.RR{Name: "one", TTL: 10 * time.Minute, Type: "TXT", Data: "hello world2"},
		libdns.RR{Name: "one", TTL: 10 * time.Minute, Type: "A"},
		libdns.TXT{Name: "one3", TTL: 10 * time.Minute, Text: "hello world"},
		libdns.TXT{Name: "@", TTL: 10 * time.Minute},
	})
	if err != nil {
		t.Fatalf("DeleteRecords: %v", err)
	}
	fmt.Println("DeleteRecords:", recs)
}
