# Arquitetura

## Monorepo

- `apps/web-client`: app tenant com React, TypeScript e Tailwind
- `apps/web-admin`: painel de superadmin em React
- `services/api`: API principal em Go + Gin
- `services/worker`: worker para fila de geração e retries
- `internal/*`: domínios compartilhados do backend
- `packages/*`: UI, hooks, client HTTP e tipos compartilhados
- `infra/*`: Docker, Swarm e Traefik

## Domínios

- Multi-tenant: `tenants`, `users`, `memberships`, RBAC e middleware de escopo
- Billing: `plans`, `subscriptions`, `credit_wallets`, `credit_transactions`, `billing_audit_logs`
- Geração: `reference_images`, `prompts`, `generation_jobs`, `result_assets`
- Providers:
  - `internal/providers/ai`: interface abstrata + mock + Google GenAI
  - `internal/providers/storage`: local + S3/MinIO
  - `internal/providers/payment`: abstraction layer + mock

## Pipeline

1. Usuário envia referência
2. Provider AI analisa a imagem
3. Sistema gera prompt estruturado
4. Usuário envia a própria foto
5. API debita crédito e cria job
6. Worker processa fila e produz ativo final
7. Resultado entra na galeria do tenant

