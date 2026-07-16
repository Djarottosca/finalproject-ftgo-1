package rajaongkir

type RajaOngkirMeta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

type GetDestinationResponse struct {
	Meta RajaOngkirMeta    `json:"meta"`
	Data []DestinationData `json:"data"`
}

type DestinationData struct {
	ID              int    `json:"id"`
	Label           string `json:"label"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	ZipCode         string `json:"zip_code"`
}

type CalculateCostResponse struct {
	Meta RajaOngkirMeta `json:"meta"`
	Data []CostData     `json:"data"`
}

type CostData struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	ETD         string `json:"etd"`
}
