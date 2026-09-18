-- Reserve support ticket numbers atomically so concurrent creates cannot pick
-- the same TIC-* value and fail on the unique ticket_id constraint.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class
        WHERE relkind = 'S'
          AND relname = 'support_ticket_number_seq'
    ) THEN
        CREATE SEQUENCE support_ticket_number_seq START WITH 1001;
    END IF;
END $$;

SELECT setval(
    'support_ticket_number_seq',
    GREATEST(
        COALESCE(
            (
                SELECT MAX(SUBSTRING(ticket_id FROM 5)::BIGINT)
                FROM support_tickets
                WHERE ticket_id ~ '^TIC-[0-9]+$'
            ),
            1000
        ),
        1000
    ),
    true
);
