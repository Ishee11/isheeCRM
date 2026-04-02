--
-- PostgreSQL database dump
--

\restrict fXldq8LRqBZmy9tA9qVAwHsEufeczF9xaUAsfunUbFA7Oqefg8MrILz0VUJNCiI

-- Dumped from database version 17rc1
-- Dumped by pg_dump version 17rc1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: set_appointment_cost(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_appointment_cost() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Получаем стоимость услуги из таблицы services
    SELECT price INTO NEW.cost
    FROM services
    WHERE services.service_id = NEW.service_id;

    RETURN NEW;
END;
$$;


--
-- Name: set_default_price(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_default_price() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Устанавливаем стоимость по умолчанию из таблицы services
    IF NEW.price IS NULL THEN
        SELECT price INTO NEW.price
        FROM services
        WHERE services.id = NEW.service_id;
    END IF;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: appointments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.appointments (
    id integer NOT NULL,
    service_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    cost numeric(10,2),
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    client_id integer,
    start_time timestamp without time zone NOT NULL,
    appointment_status character varying(20) DEFAULT 'pending'::character varying,
    payment_status character varying(20) DEFAULT 'unpaid'::character varying,
    CONSTRAINT appointments_appointment_status_check CHECK (((appointment_status)::text = ANY ((ARRAY['pending'::character varying, 'confirmed'::character varying, 'arrived'::character varying, 'no-show'::character varying])::text[]))),
    CONSTRAINT appointments_payment_status_check CHECK (((payment_status)::text = ANY (ARRAY[('unpaid'::character varying)::text, ('partially_paid'::character varying)::text, ('paid'::character varying)::text, ('subscription'::character varying)::text])))
);


--
-- Name: appointments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.appointments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: appointments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.appointments_id_seq OWNED BY public.appointments.id;


--
-- Name: clients; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clients (
    name text,
    phone text,
    email text,
    categories text,
    birth_date date,
    paid integer,
    spent numeric,
    gender text,
    discount numeric,
    last_visit timestamp without time zone,
    first_visit timestamp without time zone,
    visit_count integer,
    comment text,
    clients_id integer NOT NULL
);


--
-- Name: clients_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.clients_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: clients_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.clients_id_seq OWNED BY public.clients.clients_id;


--
-- Name: financial_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.financial_operations (
    id integer NOT NULL,
    operation_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    document_number character varying(50) NOT NULL,
    counterparty character varying(100),
    purpose character varying(255),
    cashbox character varying(100),
    comment text,
    author character varying(100),
    amount numeric(10,2) NOT NULL,
    cashbox_balance numeric(10,2),
    service_or_product character varying(100),
    client_id integer,
    appointment_id integer
);


--
-- Name: financial_operations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.financial_operations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: financial_operations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.financial_operations_id_seq OWNED BY public.financial_operations.id;


--
-- Name: services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.services (
    service_id integer NOT NULL,
    name character varying(100) NOT NULL,
    duration integer,
    price numeric(10,2) NOT NULL
);


--
-- Name: services_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.services_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: services_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.services_id_seq OWNED BY public.services.service_id;


--
-- Name: subscription_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscription_types (
    subscription_types_id integer NOT NULL,
    name character varying(100) NOT NULL,
    cost numeric(10,2) NOT NULL,
    sessions_count integer NOT NULL,
    service_ids integer[]
);


--
-- Name: subscription_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.subscription_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: subscription_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.subscription_types_id_seq OWNED BY public.subscription_types.subscription_types_id;


--
-- Name: subscription_visits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscription_visits (
    id integer NOT NULL,
    subscription_id integer NOT NULL,
    visit_date timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: subscription_visits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.subscription_visits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: subscription_visits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.subscription_visits_id_seq OWNED BY public.subscription_visits.id;


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    subscriptions_id integer NOT NULL,
    subscription_types_id integer NOT NULL,
    current_balance integer DEFAULT 0,
    status character varying(20) DEFAULT 'active'::character varying,
    client_id integer,
    cost numeric(10,2),
    sale_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT subscriptions_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'expired'::character varying])::text[])))
);


--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.subscriptions_id_seq OWNED BY public.subscriptions.subscriptions_id;


--
-- Name: appointments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.appointments ALTER COLUMN id SET DEFAULT nextval('public.appointments_id_seq'::regclass);


--
-- Name: clients clients_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients ALTER COLUMN clients_id SET DEFAULT nextval('public.clients_id_seq'::regclass);


--
-- Name: financial_operations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_operations ALTER COLUMN id SET DEFAULT nextval('public.financial_operations_id_seq'::regclass);


--
-- Name: services service_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services ALTER COLUMN service_id SET DEFAULT nextval('public.services_id_seq'::regclass);


--
-- Name: subscription_types subscription_types_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_types ALTER COLUMN subscription_types_id SET DEFAULT nextval('public.subscription_types_id_seq'::regclass);


--
-- Name: subscription_visits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_visits ALTER COLUMN id SET DEFAULT nextval('public.subscription_visits_id_seq'::regclass);


--
-- Name: subscriptions subscriptions_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions ALTER COLUMN subscriptions_id SET DEFAULT nextval('public.subscriptions_id_seq'::regclass);


--
-- Data for Name: appointments; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.appointments (id, service_id, created_at, cost, updated_at, client_id, start_time, appointment_status, payment_status) FROM stdin;
46	1	2025-01-13 15:02:32.869336	2500.00	2025-01-13 15:02:32.869336	3	2025-01-19 12:00:00	pending	unpaid
33	1	2025-01-13 12:55:05.329903	2500.00	2025-01-13 12:55:05.329903	2	2025-01-09 19:00:00	arrived	subscription
34	1	2025-01-13 12:56:56.747116	2500.00	2025-01-13 12:56:56.747116	2	2025-01-05 17:00:00	arrived	subscription
35	1	2025-01-13 13:00:11.386908	2500.00	2025-01-13 13:00:11.386908	2	2024-12-26 17:30:00	arrived	subscription
36	1	2025-01-13 13:01:18.479135	2500.00	2025-01-13 13:01:18.479135	2	2024-11-23 12:00:00	arrived	subscription
64	1	2025-01-14 23:37:58.166217	2500.00	2025-01-14 23:37:58.166217	29	2025-01-08 18:30:00	pending	paid
47	1	2025-01-13 19:55:07.541982	2500.00	2025-01-13 19:55:07.541982	15	2025-01-05 16:00:00	arrived	paid
63	1	2025-01-14 23:35:59.309189	2500.00	2025-01-14 23:35:59.309189	29	2025-01-12 13:30:00	pending	paid
62	2	2025-01-14 23:16:00.13191	3600.00	2025-01-14 23:16:00.13191	28	2025-01-12 09:15:00	pending	partially_paid
22	1	2025-01-10 21:46:47.170033	2500.00	2025-01-10 21:46:47.170033	3	2025-01-12 11:03:00	arrived	paid
59	1	2025-01-14 15:26:22.998876	2500.00	2025-01-14 15:26:22.998876	15	2025-01-14 19:00:00	arrived	paid
66	2	2025-01-18 14:44:01.609553	3600.00	2025-01-18 14:44:01.609553	28	2025-01-05 10:00:00	pending	partially_paid
38	4	2025-01-13 13:44:11.674207	2500.00	2025-01-13 13:44:11.674207	9	2025-01-19 16:00:00	pending	unpaid
37	4	2025-01-13 13:42:49.833993	2500.00	2025-01-13 13:42:49.833993	8	2025-01-19 14:30:00	pending	unpaid
1	4	2025-01-13 13:42:49.833993	2500.00	2025-01-13 13:42:49.833993	8	2025-01-07 13:00:00	arrived	subscription
67	4	2025-01-13 13:44:11.674207	2500.00	2025-01-13 13:44:11.674207	9	2025-01-07 14:30:00	arrived	subscription
68	1	2025-01-18 15:09:35.262249	2500.00	2025-01-18 15:09:35.262249	31	2025-01-20 18:30:00	pending	unpaid
\.


--
-- Data for Name: clients; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.clients (name, phone, email, categories, birth_date, paid, spent, gender, discount, last_visit, first_visit, visit_count, comment, clients_id) FROM stdin;
Наталья Веселова Фотограф	79141866626	\N	\N	\N	124300	124300	\N	\N	2024-12-27 19:30:00	2021-07-27 12:30:00	39	\N	2
Ксюша Бенгальская	79294099444	\N	\N	\N	60000	60000	\N	0	2024-12-22 15:30:00	2024-04-28 15:00:00	24	Отверстие для лица в кушетке	3
Виктор Святоха	79141922324	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	8
Екатерина Глазкова	79842823540	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	9
Вадим Кудрявцев	79145415060	\N	\N	\N	362800	365700	\N	151	\N	2022-07-27 19:30:00	\N	\N	15
а/ц Катерина с юлы	79144117055	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	26
Степан Семёнов	79244028142	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	27
Елена Селиванова	79243031122	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	28
Наталья Олеговна Пустельникова	79147724264	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	29
Наталья Медведева	79621507898	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	30
Оксана Массаж	79242240677	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	31
\.


--
-- Data for Name: financial_operations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.financial_operations (id, operation_date, document_number, counterparty, purpose, cashbox, comment, author, amount, cashbox_balance, service_or_product, client_id, appointment_id) FROM stdin;
46	2025-01-18 14:38:51.268379	PAY-64-1737175131	\N	Оплата за услугу	Основная касса	\N	\N	2500.00	2500.00	service	29	64
47	2025-01-18 14:39:02.02823	PAY-47-1737175142	\N	Оплата за услугу	Основная касса	\N	\N	2500.00	2500.00	service	15	47
48	2025-01-18 14:42:56.048743	PAY-63-1737175376	\N	Оплата за услугу	Основная касса	\N	\N	2500.00	2500.00	service	29	63
49	2025-01-18 14:43:04.786529	PAY-62-1737175384	\N	Оплата за услугу	Основная касса	\N	\N	2600.00	2600.00	service	28	62
50	2025-01-18 14:43:12.007061	PAY-22-1737175392	\N	Оплата за услугу	Основная касса	\N	\N	2500.00	2500.00	service	3	22
51	2025-01-18 14:43:17.247358	PAY-59-1737175397	\N	Оплата за услугу	Основная касса	\N	\N	2500.00	2500.00	service	15	59
52	2025-01-18 14:44:10.514675	PAY-66-1737175450	\N	Оплата за услугу	Основная касса	\N	\N	2600.00	2600.00	service	28	66
\.


--
-- Data for Name: services; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.services (service_id, name, duration, price) FROM stdin;
1	Оздоровительный массаж 60 минут	60	2500.00
3	Оздоровительный массаж 120 минут	120	4200.00
5	Юмейхо детский 45 минут	45	2000.00
4	Юмейхо 60 минут	60	2500.00
2	Оздоровительный массаж 90 минут	90	3600.00
\.


--
-- Data for Name: subscription_types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.subscription_types (subscription_types_id, name, cost, sessions_count, service_ids) FROM stdin;
1	10 сеансов массажа по 60 минут	25000.00	10	{1,4}
2	10 сеансов массажа по 90 минут	36000.00	10	{1,4}
\.


--
-- Data for Name: subscription_visits; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.subscription_visits (id, subscription_id, visit_date, created_at) FROM stdin;
3	1	2024-11-23 12:00:00	2025-01-13 13:03:07.531444
4	1	2024-12-26 17:30:00	2025-01-13 13:03:41.997161
5	1	2025-01-05 17:00:00	2025-01-13 13:04:16.473411
6	1	2025-01-09 19:00:00	2025-01-13 13:04:32.063039
7	1	2025-01-14 22:17:00	2025-01-14 22:19:48.532514
\.


--
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.subscriptions (subscriptions_id, subscription_types_id, current_balance, status, client_id, cost, sale_date) FROM stdin;
2	1	3	active	8	17500.00	2024-11-15 00:00:00
5	1	10	active	8	17500.00	2024-11-15 00:00:00
1	1	5	active	2	17500.00	2024-11-19 00:00:00
3	1	6	active	9	17500.00	2024-11-15 00:00:00
\.


--
-- Name: appointments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.appointments_id_seq', 68, true);


--
-- Name: clients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.clients_id_seq', 31, true);


--
-- Name: financial_operations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.financial_operations_id_seq', 52, true);


--
-- Name: services_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.services_id_seq', 10, true);


--
-- Name: subscription_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.subscription_types_id_seq', 1, true);


--
-- Name: subscription_visits_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.subscription_visits_id_seq', 7, true);


--
-- Name: subscriptions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.subscriptions_id_seq', 9, true);


--
-- Name: appointments appointments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_pkey PRIMARY KEY (id);


--
-- Name: clients clients_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (clients_id);


--
-- Name: financial_operations financial_operations_document_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_operations
    ADD CONSTRAINT financial_operations_document_number_key UNIQUE (document_number);


--
-- Name: financial_operations financial_operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_operations
    ADD CONSTRAINT financial_operations_pkey PRIMARY KEY (id);


--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (service_id);


--
-- Name: subscription_types subscription_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_types
    ADD CONSTRAINT subscription_types_pkey PRIMARY KEY (subscription_types_id);


--
-- Name: subscription_visits subscription_visits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_visits
    ADD CONSTRAINT subscription_visits_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (subscriptions_id);


--
-- Name: appointments set_cost_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_cost_trigger BEFORE INSERT ON public.appointments FOR EACH ROW EXECUTE FUNCTION public.set_appointment_cost();


--
-- Name: appointments appointments_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(clients_id) ON DELETE CASCADE;


--
-- Name: financial_operations fk_appointment; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_operations
    ADD CONSTRAINT fk_appointment FOREIGN KEY (appointment_id) REFERENCES public.appointments(id);


--
-- Name: financial_operations fk_client; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_operations
    ADD CONSTRAINT fk_client FOREIGN KEY (client_id) REFERENCES public.clients(clients_id) ON DELETE SET NULL;


--
-- Name: appointments fk_service; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT fk_service FOREIGN KEY (service_id) REFERENCES public.services(service_id);


--
-- Name: subscriptions fk_subscription_types; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT fk_subscription_types FOREIGN KEY (subscription_types_id) REFERENCES public.subscription_types(subscription_types_id) ON DELETE CASCADE;


--
-- Name: subscription_visits subscription_visits_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscription_visits
    ADD CONSTRAINT subscription_visits_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(subscriptions_id) ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.clients(clients_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict fXldq8LRqBZmy9tA9qVAwHsEufeczF9xaUAsfunUbFA7Oqefg8MrILz0VUJNCiI

