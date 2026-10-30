-- Deliveries across the EU: the member state of each saved address.
ALTER TABLE addresses ADD COLUMN IF NOT EXISTS country_code CHAR(2) NOT NULL DEFAULT 'FR';
