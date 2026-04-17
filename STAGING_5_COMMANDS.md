# Staging in 5 Comandi (Railway)

Prerequisiti:

- Railway service già creata da repository GitHub (o server locale avviato).
- `NP_LICENSE_SERVER_URL`, `NP_ADMIN_API_KEY`, `NP_STRIPE_WEBHOOK_SECRET` disponibili.

## Flusso rapido (incolla questi 5 comandi)

```bash
./scripts/dev-license-env.sh
```

```bash
./scripts/setup-local-license-client.sh
```

```bash
source ./.license-dev.env && go run ./cmd/license-server
```

Oppure con gestione automatica PID/log:

```bash
./scripts/license-server-up.sh
./scripts/license-server-status.sh
```

```bash
./scripts/admin-create-license.sh --reference order_123 --email cliente@example.com
```

```bash
./scripts/test-stripe-webhook.sh --reference order_123 --email cliente@example.com
```

## One-shot E2E

Per validare tutto in un colpo (server già attivo):

```bash
./scripts/staging-e2e.sh
```

## Variante con server staging Railway

Se il server è su Railway, usa:

```bash
./scripts/admin-create-license.sh --server-url "https://<tuo-service>.up.railway.app" --api-key "<NP_ADMIN_API_KEY>" --reference order_123 --email cliente@example.com
```

```bash
./scripts/test-stripe-webhook.sh --server-url "https://<tuo-service>.up.railway.app" --webhook-secret "<NP_STRIPE_WEBHOOK_SECRET>" --reference order_123 --email cliente@example.com
```
