-- No CHECK constraint: the catalog is code-defined (Plan.Valid() validates at
-- the admin endpoint) and adding a tier later must not require a migration.
-- An unknown stored value resolves to Free via EffectiveEntitlements.
ALTER TABLE organizations
    ADD COLUMN plan VARCHAR(20) NOT NULL DEFAULT 'free',
    ADD COLUMN plan_expires_at TIMESTAMPTZ;

-- Existing orgs default to 'free' with NULL (perpetual) expiry.
