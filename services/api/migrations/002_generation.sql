create table if not exists reference_images (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete cascade,
  created_by uuid not null references users(id),
  object_key text not null,
  mime_type text not null,
  size_bytes integer not null,
  analysis jsonb not null default '{}'::jsonb,
  status text not null default 'uploaded',
  created_at timestamptz not null default now()
);

create table if not exists prompts (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete cascade,
  reference_image_id uuid not null references reference_images(id) on delete cascade,
  created_by uuid not null references users(id),
  sections jsonb not null default '[]'::jsonb,
  raw_text text not null default '',
  created_at timestamptz not null default now()
);

create table if not exists generation_jobs (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete cascade,
  created_by uuid not null references users(id),
  reference_image_id uuid not null references reference_images(id) on delete cascade,
  prompt_id uuid not null references prompts(id) on delete cascade,
  source_image_key text not null,
  status text not null,
  progress integer not null default 0,
  attempts integer not null default 0,
  cost_credits integer not null default 1,
  provider_job_id text,
  failure_reason text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists result_assets (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete cascade,
  job_id uuid not null unique references generation_jobs(id) on delete cascade,
  object_key text not null,
  mime_type text not null,
  favorite boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists billing_audit_logs (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid references tenants(id) on delete set null,
  provider text not null,
  event_type text not null,
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

