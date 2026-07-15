package productimage

type AddImageRequest struct {
	ImageURL string `json:"image_url"`
}

type ImageResponse struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImageURL  string `json:"image_url"`
}
