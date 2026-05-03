package entity

import (
	"errors"
	"strings"
)

// SubscriptionType отражает тип абонемента для интерфейса
type SubscriptionType struct {
	ID            int     `json:"subscription_types_id"`
	Name          string  `json:"name"`
	Cost          float64 `json:"cost"`
	SessionsCount int     `json:"sessions_count"`
	ServiceIDs    []int   `json:"service_ids"`
}

func NewSubscriptionType(
	name string,
	cost float64,
	sessions int,
	serviceIDs []int,
) (SubscriptionType, error) {

	if strings.TrimSpace(name) == "" {
		return SubscriptionType{}, errors.New("name is required")
	}
	if cost <= 0 {
		return SubscriptionType{}, errors.New("cost must be positive")
	}
	if sessions <= 0 {
		return SubscriptionType{}, errors.New("sessions must be positive")
	}
	if len(serviceIDs) == 0 {
		return SubscriptionType{}, errors.New("services required")
	}

	return SubscriptionType{
		Name:          name,
		Cost:          cost,
		SessionsCount: sessions,
		ServiceIDs:    serviceIDs,
	}, nil
}
