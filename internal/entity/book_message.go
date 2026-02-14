package entity

// BookMessage is the Kafka message model for book creation with fibonacci metadata.
type BookMessage struct {
	Book        Book  `json:"book"`
	Position    int64 `json:"position"`
	IsFibonacci bool  `json:"is_fibonacci"`
}
