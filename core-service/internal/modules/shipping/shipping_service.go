package shipping

import (
	"context"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/external/rajaongkir"
)

// Service defines the shipping use cases exposed to the handler layer,
// backed by the RajaOngkir HTTP client.
type Service interface {
	SearchDestinations(ctx context.Context, search string) ([]DestinationResponse, error)
	CalculateCost(ctx context.Context, req CalculateCostRequest) ([]CostOptionResponse, error)
}

type service struct {
	client rajaongkir.HttpClient
}

// NewService returns the Service implementation backed by the given
// RajaOngkir client.
func NewService(client rajaongkir.HttpClient) Service {
	return &service{client: client}
}

func (s *service) SearchDestinations(_ context.Context, search string) ([]DestinationResponse, error) {
	data, err := s.client.GetDestinations(search)
	if err != nil {
		return nil, err
	}

	res := make([]DestinationResponse, 0, len(data))
	for _, d := range data {
		res = append(res, DestinationResponse{
			ID:              d.ID,
			Label:           d.Label,
			ProvinceName:    d.ProvinceName,
			CityName:        d.CityName,
			DistrictName:    d.DistrictName,
			SubdistrictName: d.SubdistrictName,
			ZipCode:         d.ZipCode,
		})
	}
	return res, nil
}

func (s *service) CalculateCost(_ context.Context, req CalculateCostRequest) ([]CostOptionResponse, error) {
	data, err := s.client.CalculateCost(req.Origin, req.Destination, req.Weight, req.Courier)
	if err != nil {
		return nil, err
	}

	res := make([]CostOptionResponse, 0, len(data))
	for _, c := range data {
		res = append(res, CostOptionResponse{
			Name:        c.Name,
			Code:        c.Code,
			Service:     c.Service,
			Description: c.Description,
			Cost:        c.Cost,
			ETD:         c.ETD,
		})
	}
	return res, nil
}
