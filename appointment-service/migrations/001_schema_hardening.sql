BEGIN;

ALTER TABLE public.clients
    ADD COLUMN IF NOT EXISTS deleted_at timestamp without time zone;

ALTER TABLE public.appointments
    ADD COLUMN IF NOT EXISTS deleted_at timestamp without time zone;

ALTER TABLE public.services
    ADD COLUMN IF NOT EXISTS deleted_at timestamp without time zone;

ALTER TABLE public.subscriptions
    ADD COLUMN IF NOT EXISTS deleted_at timestamp without time zone;

UPDATE public.subscriptions AS s
SET cost = st.cost
FROM public.subscription_types AS st
WHERE s.subscription_types_id = st.subscription_types_id
  AND COALESCE(s.cost, 0) = 0
  AND st.cost > 0;

UPDATE public.financial_operations AS fo
SET amount = s.cost,
    cashbox_balance = COALESCE(NULLIF(fo.cashbox_balance, 0), s.cost)
FROM public.subscriptions AS s
WHERE fo.service_or_product = 'subscription'
  AND fo.amount = 0
  AND s.client_id = fo.client_id
  AND s.sale_date::date = fo.operation_date::date
  AND s.cost > 0;

UPDATE public.financial_operations AS fo
SET amount = st.cost,
    cashbox_balance = COALESCE(NULLIF(fo.cashbox_balance, 0), st.cost)
FROM public.subscription_types AS st
WHERE fo.service_or_product = 'subscription'
  AND fo.amount = 0
  AND fo.purpose = ('Покупка абонемента: ' || st.name)
  AND st.cost > 0;

ALTER TABLE public.clients
    ALTER COLUMN phone SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'clients_phone_format_check'
          AND conrelid = 'public.clients'::regclass
    ) THEN
        ALTER TABLE public.clients
            ADD CONSTRAINT clients_phone_format_check
            CHECK (phone ~ '^7[0-9]{10}$');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscriptions_current_balance_nonnegative'
          AND conrelid = 'public.subscriptions'::regclass
    ) THEN
        ALTER TABLE public.subscriptions
            ADD CONSTRAINT subscriptions_current_balance_nonnegative
            CHECK (current_balance >= 0);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'appointments_cost_nonnegative'
          AND conrelid = 'public.appointments'::regclass
    ) THEN
        ALTER TABLE public.appointments
            ADD CONSTRAINT appointments_cost_nonnegative
            CHECK (cost IS NULL OR cost >= 0);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'financial_operations_amount_positive'
          AND conrelid = 'public.financial_operations'::regclass
    ) THEN
        ALTER TABLE public.financial_operations
            ADD CONSTRAINT financial_operations_amount_positive
            CHECK (amount > 0);
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'appointments_client_id_fkey'
          AND conrelid = 'public.appointments'::regclass
    ) THEN
        ALTER TABLE public.appointments
            DROP CONSTRAINT appointments_client_id_fkey;
    END IF;

    ALTER TABLE public.appointments
        ADD CONSTRAINT appointments_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES public.clients(clients_id);
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscriptions_client_id_fkey'
          AND conrelid = 'public.subscriptions'::regclass
    ) THEN
        ALTER TABLE public.subscriptions
            DROP CONSTRAINT subscriptions_client_id_fkey;
    END IF;

    ALTER TABLE public.subscriptions
        ADD CONSTRAINT subscriptions_client_id_fkey
        FOREIGN KEY (client_id)
        REFERENCES public.clients(clients_id);
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'subscription_visits_appointment_id_fkey'
          AND conrelid = 'public.subscription_visits'::regclass
    ) THEN
        ALTER TABLE public.subscription_visits
            ADD CONSTRAINT subscription_visits_appointment_id_fkey
            FOREIGN KEY (appointment_id)
            REFERENCES public.appointments(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS clients_phone_active_uidx
    ON public.clients (phone)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS appointments_start_time_idx
    ON public.appointments (start_time)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS appointments_client_id_idx
    ON public.appointments (client_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS appointments_payment_status_start_time_idx
    ON public.appointments (payment_status, start_time)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS financial_operations_operation_date_idx
    ON public.financial_operations (operation_date);

CREATE INDEX IF NOT EXISTS financial_operations_appointment_id_idx
    ON public.financial_operations (appointment_id);

CREATE INDEX IF NOT EXISTS subscriptions_client_id_idx
    ON public.subscriptions (client_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS subscription_visits_appointment_id_uidx
    ON public.subscription_visits (appointment_id)
    WHERE appointment_id IS NOT NULL;

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
      AND a.appointment_status = 'arrived'
    GROUP BY a.client_id
) AS visits
    ON visits.client_id = c.clients_id
WHERE c.deleted_at IS NULL;

COMMIT;
