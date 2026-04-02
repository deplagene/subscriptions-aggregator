package subscriptions

import "errors"

var ErrSubscriptionNotFound = errors.New("subscription not found")
var ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
