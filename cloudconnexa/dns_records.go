package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrDNSRecordNotFound is returned when a DNS record is not found.
	ErrDNSRecordNotFound = errors.New("dns record not found")
)

// DNSRecord represents a DNS record in CloudConnexa.
type DNSRecord struct {
	ID            string   `json:"id"`
	Domain        string   `json:"domain"`
	Description   string   `json:"description"`
	IPV4Addresses []string `json:"ipv4Addresses"`
	IPV6Addresses []string `json:"ipv6Addresses"`
}

// DNSRecordPageResponse represents a paginated response of DNS records.
type DNSRecordPageResponse struct {
	Content          []DNSRecord `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

// DNSRecordsService provides methods for managing DNS records.
type DNSRecordsService service

// GetByPage retrieves DNS records using pagination.
func (c *DNSRecordsService) GetByPage(page int, pageSize int) (DNSRecordPageResponse, error) {
	endpoint := fmt.Sprintf("%s/dns-records?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return DNSRecordPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return DNSRecordPageResponse{}, err
	}

	var response DNSRecordPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return DNSRecordPageResponse{}, err
	}
	return response, nil
}

// GetDNSRecord retrieves a specific DNS record by ID.
func (c *DNSRecordsService) GetDNSRecord(recordID string) (*DNSRecord, error) {
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, record := range response.Content {
			if record.ID == recordID {
				return &record, nil
			}
		}

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return nil, ErrDNSRecordNotFound
}

// Create creates a new DNS record.
func (c *DNSRecordsService) Create(record DNSRecord) (*DNSRecord, error) {
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/dns-records", c.client.GetV1Url()), bytes.NewBuffer(recordJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var d DNSRecord
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Update updates an existing DNS record.
func (c *DNSRecordsService) Update(record DNSRecord) error {
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/dns-records/%s", c.client.GetV1Url(), record.ID), bytes.NewBuffer(recordJSON))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// Delete deletes a DNS record by ID.
func (c *DNSRecordsService) Delete(recordID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/dns-records/%s", c.client.GetV1Url(), recordID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
