# Staging in 5 Comandi

Prerequisiti:

- Render service già creata da `render.yaml` (o server locale avviato).
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

## Variante con server staging Render

Se il server è su Render, usa:

```bash
./scripts/admin-create-license.sh --server-url "https://<tuo-service>.onrender.com" --api-key "<NP_ADMIN_API_KEY>" --reference order_123 --email cliente@example.com
```

```bash
./scripts/test-stripe-webhook.sh --server-url "https://<tuo-service>.onrender.com" --webhook-secret "<NP_STRIPE_WEBHOOK_SECRET>" --reference order_123 --email cliente@example.com
```
