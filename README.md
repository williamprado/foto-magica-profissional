# Foto Magica Profissional

Geração de retratos profissionais com IA,  implementado de forma original com Go, React, worker assíncrono, créditos, billing e deploy em Docker Swarm.



## Execução local

1. Copie `.env.example` para `.env`.
2. Suba um PostgreSQL local.
3. Exporte as variáveis.
4. Rode a API:

```bash
go run ./services/api/cmd/api
```

5. Rode o worker:

```bash
go run ./services/worker/cmd/worker
```

6. Rode o frontend:

```bash
npm install
npm run dev -w apps/web-client
```

## Produção em Swarm

O arquivo principal de deploy é:

- `infra/swarm/stack.prod.yml`

Ele foi preparado para uso com:

- rede externa `waianet`
- Traefik já existente
- Portainer stack
- domínio 

### Variáveis de stack

- `PUBLIC_HOST=fotomagica.wcom.br`
- `TRAEFIK_CERT_RESOLVER=letsencrypt`
- `POSTGRES_PASSWORD=...`
- `JWT_SECRET=...`
- `GOOGLE_API_KEY=...`
- `APP_GIT_REPO=https://github.com/SEU_USUARIO/SEU_REPO.git`
- `APP_GIT_REF=main`

### Observação importante

O stack de produção atual foi preparado para bootstrap direto do repositório Git no runtime dos serviços, o que permite deploy imediato pelo Portainer/Swarm sem depender de um registry privado já configurado. Os Dockerfiles completos também estão no repositório para evolução posterior para imagens versionadas.

## Endpoints principais

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `GET /v1/profile/me`
- `GET /v1/billing/plans`
- `POST /v1/billing/checkout`
- `GET /v1/credits/wallet`
- `GET /v1/generation/jobs`
- `POST /v1/generation/jobs`
- `GET /v1/gallery`

## Validação já coberta na base


- RBAC e JWT
- Migrations automáticas
- Créditos e transações
- Pipeline assíncrono com retries
- Providers desacoplados
- Apps React tenant/admin
- Stack Swarm + Traefik + Portainer

## Próximos passos recomendados

- Ativar provider real de billing
- Publicar imagens em registry
- Adicionar testes automatizados
- Adicionar observabilidade e métricas

