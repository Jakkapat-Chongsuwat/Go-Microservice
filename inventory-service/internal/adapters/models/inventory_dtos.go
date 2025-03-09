package models

type CreateProductRequest struct {
	Name     string  `json:"name" example:"Widget"`
	Quantity int     `json:"quantity" example:"100"`
	Price    float64 `json:"price" example:"9.99"`
}

type UpdateProductMetadataRequest struct {
	Name  string  `json:"name" example:"Updated Widget"`
	Price float64 `json:"price" example:"12.99"`
}

type UpdateProductStockQuantityRequest struct {
	QuantityChange int `json:"quantityChange" example:"-10"`
}

type ProductResponse struct {
	ID       string  `json:"id" example:"a1b2c3d4"`
	Name     string  `json:"name" example:"Widget"`
	Quantity int     `json:"quantity" example:"100"`
	Price    float64 `json:"price" example:"9.99"`
}
