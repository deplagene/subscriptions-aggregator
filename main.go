package main

import (
	"time"

	"github.com/google/uuid"
)

func main() {
}

// сущность подписка
type Subscription struct {
	SubscriptionID uuid.UUID // в тз не указано
	ServiceName    string
	Price          float64 // целое число рублей
	UserID         uuid.UUID
	// дата начала
	StartedAt time.Time
	// дата окончания
	EndedAt *time.Time
}

// по тз дефолт CRUDL - операции
// сам сервис - онлайн агрегатор подписок юзера
// бд - постгря с миграциями
// логи
// докер-композ
