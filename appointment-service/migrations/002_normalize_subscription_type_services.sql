BEGIN;

CREATE TABLE IF NOT EXISTS public.subscription_type_services (
    subscription_type_id integer NOT NULL,
    service_id integer NOT NULL,
    CONSTRAINT subscription_type_services_pkey
        PRIMARY KEY (subscription_type_id, service_id),
    CONSTRAINT subscription_type_services_subscription_type_id_fkey
        FOREIGN KEY (subscription_type_id)
        REFERENCES public.subscription_types(subscription_types_id)
        ON DELETE CASCADE,
    CONSTRAINT subscription_type_services_service_id_fkey
        FOREIGN KEY (service_id)
        REFERENCES public.services(service_id)
);

CREATE INDEX IF NOT EXISTS subscription_type_services_service_id_idx
    ON public.subscription_type_services (service_id);

INSERT INTO public.subscription_type_services (subscription_type_id, service_id)
SELECT
    st.subscription_types_id,
    service_id
FROM public.subscription_types AS st
CROSS JOIN LATERAL unnest(COALESCE(st.service_ids, '{}'::integer[])) AS service_id
ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION public.sync_subscription_type_services_from_array()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF pg_trigger_depth() > 1 THEN
        RETURN NEW;
    END IF;

    DELETE FROM public.subscription_type_services
    WHERE subscription_type_id = NEW.subscription_types_id;

    INSERT INTO public.subscription_type_services (subscription_type_id, service_id)
    SELECT
        NEW.subscription_types_id,
        service_id
    FROM unnest(COALESCE(NEW.service_ids, '{}'::integer[])) AS service_id
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS sync_subscription_type_services_from_array_trigger
    ON public.subscription_types;

CREATE TRIGGER sync_subscription_type_services_from_array_trigger
AFTER INSERT OR UPDATE OF service_ids ON public.subscription_types
FOR EACH ROW
EXECUTE FUNCTION public.sync_subscription_type_services_from_array();

COMMIT;
