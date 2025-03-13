package models

// Meta holds metadata for responses.
type Meta struct {
	Total int `json:"total,omitempty"`
}

type Response[T any] struct {
	Data T     `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

func NewResponse[T any](data T, meta *Meta) Response[T] {
	return Response[T]{
		Data: data,
		Meta: meta,
	}
}
