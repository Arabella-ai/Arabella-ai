-- Make all templates non-premium for testing Wan AI
UPDATE templates SET is_premium = false WHERE is_premium = true;

-- Show updated templates
SELECT id, name, is_premium, is_active, credit_cost FROM templates;







