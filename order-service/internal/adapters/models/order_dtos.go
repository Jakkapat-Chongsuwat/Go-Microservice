package models

type CreateOrderRequest struct {
	UserID string             `json:"userId"`
	Items  []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type OrderItemResponse struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderResponse struct {
	OrderID string              `json:"orderId"`
	UserID  string              `json:"userId"`
	Status  string              `json:"status"`
	Items   []OrderItemResponse `json:"items"`
}
