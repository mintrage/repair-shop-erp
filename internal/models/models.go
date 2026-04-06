package models

type Part struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type RepairOrder struct {
	ID                 int    `json:"id"`
	ClientName         string `json:"client_name"`
	PhoneNumber        string `json:"phone_number"`
	DeviceModel        string `json:"device_model"`
	ProblemDescription string `json:"problem_description"`
	Status             string `json:"status"`
}

type RepairOrderPartInput struct {
	PartID   int `json:"part_id"`
	Quantity int `json:"quantity"`
}

type RepairOrderInput struct {
	ClientName         string                 `json:"client_name"`
	PhoneNumber        string                 `json:"phone_number"`
	DeviceModel        string                 `json:"device_model"`
	ProblemDescription string                 `json:"problem_description"`
	Status             string                 `json:"status"`
	Parts              []RepairOrderPartInput `json:"items"`
}

type AddPartToOrderInput struct {
	OrderID  int `json:"order_id"`
	PartID   int `json:"part_id"`
	Quantity int `json:"quantity"`
}
