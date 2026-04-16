--
-- PostgreSQL database dump
--

\restrict 8Zo8jsxo1tBaznfUAlISdJ16I8xkamhGbZe4uec2BfzR3vsNMUeIInGehN2HbGh

-- Dumped from database version 16.11
-- Dumped by pg_dump version 16.11

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
-- Name: media_ingest_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.media_ingest_bindings (
    id bigint NOT NULL,
    publish_session_id bigint NOT NULL,
    provider_publish_id text,
    provider_vhost text,
    provider_app text,
    ingest_query_param text,
    record_local_uri text,
    last_callback_action text,
    last_callback_at timestamp with time zone,
    provider_context jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: media_ingest_bindings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.media_ingest_bindings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: media_ingest_bindings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.media_ingest_bindings_id_seq OWNED BY public.media_ingest_bindings.id;


--
-- Name: media_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.media_providers (
    id bigint NOT NULL,
    code text NOT NULL,
    display_name text DEFAULT ''::text NOT NULL,
    api_base_url text,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: media_providers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.media_providers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: media_providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.media_providers_id_seq OWNED BY public.media_providers.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: stream_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stream_keys (
    id bigint NOT NULL,
    owner_user_id bigint NOT NULL,
    stream_key_secret text NOT NULL,
    media_provider_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone
);


--
-- Name: stream_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stream_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stream_keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stream_keys_id_seq OWNED BY public.stream_keys.id;


--
-- Name: stream_publish_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stream_publish_sessions (
    id bigint NOT NULL,
    streamer_user_id bigint NOT NULL,
    playback_id text NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'created'::text NOT NULL,
    playback_url_cdn text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    media_provider_id bigint NOT NULL,
    stream_key_id bigint NOT NULL,
    CONSTRAINT stream_publish_sessions_status_check CHECK ((status = ANY (ARRAY['created'::text, 'live'::text, 'ended'::text, 'failed'::text])))
);


--
-- Name: stream_publish_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stream_publish_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stream_publish_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stream_publish_sessions_id_seq OWNED BY public.stream_publish_sessions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    display_name text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: view_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.view_sessions (
    id bigint NOT NULL,
    publish_session_id bigint NOT NULL,
    viewer_user_id bigint,
    viewer_ref text DEFAULT ''::text NOT NULL,
    client_type text DEFAULT 'web'::text NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    left_at timestamp with time zone,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT view_sessions_viewer_identity_chk CHECK (((viewer_user_id IS NOT NULL) OR (length(TRIM(BOTH FROM viewer_ref)) > 0)))
);


--
-- Name: view_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.view_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: view_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.view_sessions_id_seq OWNED BY public.view_sessions.id;


--
-- Name: media_ingest_bindings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_ingest_bindings ALTER COLUMN id SET DEFAULT nextval('public.media_ingest_bindings_id_seq'::regclass);


--
-- Name: media_providers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_providers ALTER COLUMN id SET DEFAULT nextval('public.media_providers_id_seq'::regclass);


--
-- Name: stream_keys id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_keys ALTER COLUMN id SET DEFAULT nextval('public.stream_keys_id_seq'::regclass);


--
-- Name: stream_publish_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_publish_sessions ALTER COLUMN id SET DEFAULT nextval('public.stream_publish_sessions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: view_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.view_sessions ALTER COLUMN id SET DEFAULT nextval('public.view_sessions_id_seq'::regclass);


--
-- Name: media_ingest_bindings media_ingest_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_ingest_bindings
    ADD CONSTRAINT media_ingest_bindings_pkey PRIMARY KEY (id);


--
-- Name: media_ingest_bindings media_ingest_bindings_publish_session_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_ingest_bindings
    ADD CONSTRAINT media_ingest_bindings_publish_session_id_key UNIQUE (publish_session_id);


--
-- Name: media_providers media_providers_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_providers
    ADD CONSTRAINT media_providers_code_key UNIQUE (code);


--
-- Name: media_providers media_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_providers
    ADD CONSTRAINT media_providers_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: stream_keys stream_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_keys
    ADD CONSTRAINT stream_keys_pkey PRIMARY KEY (id);


--
-- Name: stream_keys stream_keys_stream_key_secret_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_keys
    ADD CONSTRAINT stream_keys_stream_key_secret_key UNIQUE (stream_key_secret);


--
-- Name: stream_publish_sessions stream_publish_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_publish_sessions
    ADD CONSTRAINT stream_publish_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.stream_publish_sessions
    ADD CONSTRAINT stream_publish_sessions_playback_id_key UNIQUE (playback_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: view_sessions view_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.view_sessions
    ADD CONSTRAINT view_sessions_pkey PRIMARY KEY (id);


--
-- Name: idx_media_ingest_bindings_provider_publish_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_media_ingest_bindings_provider_publish_id ON public.media_ingest_bindings USING btree (provider_publish_id);


--
-- Name: idx_stream_keys_owner_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_keys_owner_user_id ON public.stream_keys USING btree (owner_user_id);


--
-- Name: idx_stream_publish_sessions_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_publish_sessions_created_at ON public.stream_publish_sessions USING btree (created_at DESC);


--
-- Name: idx_stream_publish_sessions_media_provider_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_publish_sessions_media_provider_id ON public.stream_publish_sessions USING btree (media_provider_id);


--
-- Name: idx_stream_publish_sessions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_publish_sessions_status ON public.stream_publish_sessions USING btree (status);


--
-- Name: idx_stream_publish_sessions_stream_key_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_publish_sessions_stream_key_id ON public.stream_publish_sessions USING btree (stream_key_id);


--
-- Name: idx_stream_publish_sessions_streamer_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stream_publish_sessions_streamer_user_id ON public.stream_publish_sessions USING btree (streamer_user_id);


--
-- Name: idx_view_sessions_joined_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_view_sessions_joined_at ON public.view_sessions USING btree (joined_at DESC);


--
-- Name: idx_view_sessions_publish_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_view_sessions_publish_session_id ON public.view_sessions USING btree (publish_session_id);


--
-- Name: idx_view_sessions_viewer_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_view_sessions_viewer_user_id ON public.view_sessions USING btree (viewer_user_id);


--
-- Name: media_ingest_bindings media_ingest_bindings_publish_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_ingest_bindings
    ADD CONSTRAINT media_ingest_bindings_publish_session_id_fkey FOREIGN KEY (publish_session_id) REFERENCES public.stream_publish_sessions(id) ON DELETE CASCADE;


--
-- Name: stream_keys stream_keys_media_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_keys
    ADD CONSTRAINT stream_keys_media_provider_id_fkey FOREIGN KEY (media_provider_id) REFERENCES public.media_providers(id) ON DELETE RESTRICT;


--
-- Name: stream_keys stream_keys_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_keys
    ADD CONSTRAINT stream_keys_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;


--
-- Name: stream_publish_sessions stream_publish_sessions_media_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_publish_sessions
    ADD CONSTRAINT stream_publish_sessions_media_provider_id_fkey FOREIGN KEY (media_provider_id) REFERENCES public.media_providers(id) ON DELETE RESTRICT;


--
-- Name: stream_publish_sessions stream_publish_sessions_stream_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_publish_sessions
    ADD CONSTRAINT stream_publish_sessions_stream_key_id_fkey FOREIGN KEY (stream_key_id) REFERENCES public.stream_keys(id) ON DELETE RESTRICT;


--
-- Name: stream_publish_sessions stream_publish_sessions_streamer_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stream_publish_sessions
    ADD CONSTRAINT stream_publish_sessions_streamer_user_id_fkey FOREIGN KEY (streamer_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;


--
-- Name: view_sessions view_sessions_publish_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.view_sessions
    ADD CONSTRAINT view_sessions_publish_session_id_fkey FOREIGN KEY (publish_session_id) REFERENCES public.stream_publish_sessions(id) ON DELETE CASCADE;


--
-- Name: view_sessions view_sessions_viewer_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.view_sessions
    ADD CONSTRAINT view_sessions_viewer_user_id_fkey FOREIGN KEY (viewer_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

\unrestrict 8Zo8jsxo1tBaznfUAlISdJ16I8xkamhGbZe4uec2BfzR3vsNMUeIInGehN2HbGh

