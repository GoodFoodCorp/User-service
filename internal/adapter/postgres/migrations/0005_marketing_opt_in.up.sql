-- Promotional emails need the customer's prior consent (CPCE art. L34-5):
-- off by default. Existing choices are reset too: the switch used to be
-- on by default, and a pre-checked box is no valid consent.
ALTER TABLE notification_preferences ALTER COLUMN email_promos SET DEFAULT FALSE;
UPDATE notification_preferences SET email_promos = FALSE;
