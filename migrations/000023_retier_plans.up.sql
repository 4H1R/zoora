-- Re-tier plans: the code-defined catalog is now "<tier>_<size>"
-- (free/plus/pro/max × 50/100/200/500/1000 members).
-- Legacy keys are remapped to the nearest new plan at or above their old
-- capacity: free (10 users) -> free_50, pro (150 users) -> pro_200,
-- enterprise (unlimited) -> max_1000.

ALTER TABLE organizations ALTER COLUMN plan SET DEFAULT 'free_50';

UPDATE organizations SET plan = 'free_50'  WHERE plan = 'free';
UPDATE organizations SET plan = 'pro_200'  WHERE plan = 'pro';
UPDATE organizations SET plan = 'max_1000' WHERE plan = 'enterprise';

UPDATE plan_prices SET plan = 'pro_200'  WHERE plan = 'pro';
UPDATE plan_prices SET plan = 'max_1000' WHERE plan = 'enterprise';

UPDATE invoice_items SET plan = 'pro_200'  WHERE plan = 'pro';
UPDATE invoice_items SET plan = 'max_1000' WHERE plan = 'enterprise';
