package crm

import "time"

const (
	EventCustomerUpdated = "customer.updated"
)

// CustomerUpdatedV1 - обновлены данные покупателя в CRM-системе
// schemagen: customer.updated.v1
type CustomerUpdatedV1 struct {
	Id        int       `json:"id" schema:"id,required"`
	Name      string    `json:"name" schema:"name,required"`
	Birthday  time.Time `json:"birthday" schema:"birthday,required"`
	IsBlocked bool      `json:"is_blocked" schema:"is_blocked,required"`
	Comment   string    `json:"comment" schema:"-"`
	Contacts  struct {
		Phone int    `json:"phone" schema:"phone"`
		Email string `json:"email" schema:"email"`
	} `json:"contacts" schema:"contacts"`
	Cards []struct {
		Number    string `json:"number" schema:"number,required"`
		IsBlocked bool   `json:"is_blocked" schema:"is_blocked,required"`
	} `json:"cards" schema:"card"`
}
