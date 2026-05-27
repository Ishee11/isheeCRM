BEGIN;

CREATE OR REPLACE VIEW public.client_stats AS
SELECT
    c.clients_id,
    COALESCE(pay.total_paid, 0::numeric) AS paid,
    COALESCE(spend.total_spent, 0::numeric) AS spent,
    visits.last_visit,
    visits.first_visit,
    COALESCE(visits.visit_count, 0::bigint)::integer AS visit_count
FROM public.clients AS c
LEFT JOIN (
    SELECT
        fo.client_id,
        SUM(fo.amount) AS total_paid
    FROM public.financial_operations AS fo
    WHERE fo.client_id IS NOT NULL
    GROUP BY fo.client_id
) AS pay
    ON pay.client_id = c.clients_id
LEFT JOIN (
    SELECT
        a.client_id,
        SUM(a.cost) AS total_spent
    FROM public.appointments AS a
    WHERE a.client_id IS NOT NULL
      AND a.deleted_at IS NULL
      AND a.appointment_status = 'arrived'
    GROUP BY a.client_id
) AS spend
    ON spend.client_id = c.clients_id
LEFT JOIN (
    SELECT
        a.client_id,
        MAX(a.start_time) AS last_visit,
        MIN(a.start_time) AS first_visit,
        COUNT(*) AS visit_count
    FROM public.appointments AS a
    WHERE a.client_id IS NOT NULL
      AND a.deleted_at IS NULL
    GROUP BY a.client_id
) AS visits
    ON visits.client_id = c.clients_id
WHERE c.deleted_at IS NULL;

COMMIT;
