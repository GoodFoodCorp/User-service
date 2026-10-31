package domain

import "testing"

// Marketing needs consent: never on before the customer turns it on.
func TestDefaultPreferencesHaveNoMarketing(t *testing.T) {
	p := DefaultNotificationPreferences("u1")
	if p.EmailPromos {
		t.Fatal("promotional emails must be off by default")
	}
	if !p.EmailOrders {
		t.Fatal("order emails stay on by default")
	}
}
