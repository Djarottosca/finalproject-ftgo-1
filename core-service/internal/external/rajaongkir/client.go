package rajaongkir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

type HttpClient interface {
	GetDestinations(search string) ([]DestinationData, error)
	CalculateCost(origin string, destination string, weight int, courier string) ([]CostData, error)
}

func NewHttpClient(cfg *config.RajaOngkirConfig) HttpClient {
	client := http.Client{Timeout: 10 * time.Second}

	return &httpClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.ApiKey,
		client:  &client,
	}
}

type httpClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func (c *httpClient) GetDestinations(search string) ([]DestinationData, error) {
	url := fmt.Sprintf("%s/destination/domestic-destination?search=%s", c.baseURL, search)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GetDestinationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Meta.Status != "success" {
		return nil, fmt.Errorf("there was an error: %s", result.Meta.Message)
	}

	return result.Data, nil
}

func (c *httpClient) CalculateCost(origin string, destination string, weight int, courier string) ([]CostData, error) {
	apiURL := fmt.Sprintf("%s/calculate/domestic-cost", c.baseURL)
	body := url.Values{}
	body.Set("origin", origin)
	body.Set("destination", destination)
	body.Set("weight", fmt.Sprintf("%d", weight))
	body.Set("courier", courier)
	body.Set("price", "lowest")

	payload := bytes.NewBufferString(body.Encode())
	req, err := http.NewRequest("POST", apiURL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("key", c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CalculateCostResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Meta.Status != "success" {
		return nil, fmt.Errorf("there was an error: %s", result.Meta.Message)
	}

	logger.Log.Info().Any("result", result).Msg("Calculate Cost Response")

	return result.Data, nil
}
