--
-- PostgreSQL database dump
--

\restrict xofqGegNV31EQvdYS9HjqA6D4MZpKfWxvoJrUtMuH2CO7TRjGv7K4KVcD0Nmbaw

-- Dumped from database version 15.17
-- Dumped by pg_dump version 15.17

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: cache_control_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cache_control_user (
    user_id uuid NOT NULL,
    version integer DEFAULT 1 NOT NULL
);


ALTER TABLE public.cache_control_user OWNER TO postgres;

--
-- Name: COLUMN cache_control_user.user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.cache_control_user.user_id IS 'ID пользователя (UUID)';


--
-- Name: COLUMN cache_control_user.version; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.cache_control_user.version IS 'Текущая версия кеша (инкрементится при каждом изменении подписок пользователя)';


--
-- Name: subscription_templates; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscription_templates (
    id integer NOT NULL,
    service_name text NOT NULL,
    price integer NOT NULL,
    CONSTRAINT subscription_templates_price_check CHECK ((price >= 0))
);


ALTER TABLE public.subscription_templates OWNER TO postgres;

--
-- Name: TABLE subscription_templates; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.subscription_templates IS 'Шаблоны подписок, создаваемые админом';


--
-- Name: COLUMN subscription_templates.service_name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subscription_templates.service_name IS 'Название сервиса (уникальное, без учёта регистра и пробелов)';


--
-- Name: COLUMN subscription_templates.price; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subscription_templates.price IS 'Цена подписки в рублях (целое число)';


--
-- Name: subscription_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subscription_templates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.subscription_templates_id_seq OWNER TO postgres;

--
-- Name: subscription_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subscription_templates_id_seq OWNED BY public.subscription_templates.id;


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscriptions (
    id integer NOT NULL,
    user_id uuid NOT NULL,
    template_id integer,
    start_date date NOT NULL,
    end_date date
);


ALTER TABLE public.subscriptions OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.subscriptions_id_seq OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subscriptions_id_seq OWNED BY public.subscriptions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: subscription_templates id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscription_templates ALTER COLUMN id SET DEFAULT nextval('public.subscription_templates_id_seq'::regclass);


--
-- Name: subscriptions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions ALTER COLUMN id SET DEFAULT nextval('public.subscriptions_id_seq'::regclass);


--
-- Data for Name: cache_control_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cache_control_user (user_id, version) FROM stdin;
\.


--
-- Data for Name: subscription_templates; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subscription_templates (id, service_name, price) FROM stdin;
\.


--
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subscriptions (id, user_id, template_id, start_date, end_date) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, email, password_hash, role, created_at) FROM stdin;
\.


--
-- Name: subscription_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subscription_templates_id_seq', 1, false);


--
-- Name: subscriptions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subscriptions_id_seq', 1, false);


--
-- Name: cache_control_user cache_control_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cache_control_user
    ADD CONSTRAINT cache_control_user_pkey PRIMARY KEY (user_id);


--
-- Name: subscription_templates subscription_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscription_templates
    ADD CONSTRAINT subscription_templates_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_user_id_template_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_template_id_key UNIQUE (user_id, template_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_subscriptions_dates; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_dates ON public.subscriptions USING btree (start_date, end_date);


--
-- Name: idx_subscriptions_template_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_template_id ON public.subscriptions USING btree (template_id);


--
-- Name: idx_subscriptions_user_dates; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_user_dates ON public.subscriptions USING btree (user_id, start_date, end_date);


--
-- Name: idx_subscriptions_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_user_id ON public.subscriptions USING btree (user_id);


--
-- Name: idx_unique_service_name_ci; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_unique_service_name_ci ON public.subscription_templates USING btree (lower(TRIM(BOTH FROM service_name)));


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_role ON public.users USING btree (role);


--
-- Name: subscriptions subscriptions_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.subscription_templates(id) ON DELETE SET NULL;


--
-- Name: subscriptions subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict xofqGegNV31EQvdYS9HjqA6D4MZpKfWxvoJrUtMuH2CO7TRjGv7K4KVcD0Nmbaw

