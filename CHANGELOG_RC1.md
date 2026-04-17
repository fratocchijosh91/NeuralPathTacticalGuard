# v1.1.0-rc1 - Release Candidate

## Testo pronto per GitHub Release (copia da qui)

Release candidate: license server su Railway, client con attivazione via token **Ed25519**, hardening concorrenza e rete.

**Novità principali:** Docker per `cmd/license-server`, endpoint attivazione/admin/webhook Stripe, rate limit e audit log, `go test -race` e script E2E locali/staging.

**Client:** in `config.json` impostare `license_server_url` (HTTPS del servizio Railway) e `license_public_key` (base64 della chiave pubblica usata dal server in produzione). Dev e produzione devono usare la **stessa** coppia URL + chiave pubblica del server che firma i token.

**Deploy server:** variabili in `## Variabili richieste (Railway)` più sotto; non impostare `NP_LICENSE_ADDR` su Railway.

**Verifica rapida:** `./scripts/staging-e2e.sh --server-url "https://…" --admin-api-key "…" --webhook-secret "…"` deve terminare con `4 passati, 0 falliti`.

---

## Dettaglio (stesso contenuto espanso)

Release candidate con stack licensing pronto per produzione e deploy staging funzionante su Railway.

## Highlights

- License server Go pronto al deploy (Docker, Railway)
- Attivazione licenza con token firmato ed25519
- Endpoint admin protetto da API key
- Webhook Stripe con verifica firma
- Persistenza allowlist su file
- Rate limit + audit log + security headers
- CI locale (vet, test, race, build)

## Endpoint disponibili

- `GET /healthz`
- `GET /v1/public-key`
- `POST /v1/licenses/activate`
- `POST /v1/webhooks/stripe`
- `POST /v1/admin/licenses/create`

## Variabili richieste (Railway)

- `NP_LICENSE_PRIVATE_KEY_B64` (obbligatoria)
- `NP_ADMIN_API_KEY`
- `NP_STRIPE_WEBHOOK_SECRET`
- `NP_LICENSE_PRODUCT`, `NP_LICENSE_TIER`, `NP_LICENSE_PREFIX`
- `NP_LICENSE_TOKEN_TTL_HOURS`
- `NP_LICENSE_RATE_LIMIT_PER_MIN`
- `NP_LICENSE_KEYS_PATH`
- `NP_LICENSE_ALLOW_ANY_KEY`

Non impostare `NP_LICENSE_ADDR` su Railway: il servizio usa automaticamente `PORT`.

## Test eseguiti

- `go test ./...`
- `go test -race ./...`
- CI pipeline completa (8/8 passati)
- Staging E2E contro Railway: 4/4 passati
  - healthcheck
  - admin create license
  - activate token firmato
  - webhook Stripe firmato

## Known limitations

- Persistenza allowlist su file (singolo nodo).
- Stripe signing secret deve essere aggiornato in Railway dopo configurazione reale webhook.
- Rotazione chiavi ed25519 ancora manuale.

## Come testare rapidamente

```bash
./scripts/staging-e2e.sh \
  --server-url "https://<tuo-servizio>.up.railway.app" \
  --admin-api-key "<NP_ADMIN_API_KEY>" \
  --webhook-secret "<NP_STRIPE_WEBHOOK_SECRET>"
```

Deve chiudere con `4 passati, 0 falliti`.

## Pubblicare il tag e la release

```bash
git tag -a v1.1.0-rc1 -m "NeuralPath Tactical Guard v1.1.0-rc1 (license server + client hardening)"
git push origin v1.1.0-rc1
```

Su GitHub: **Releases → Draft a new release**, seleziona il tag `v1.1.0-rc1`, titolo ad es. `v1.1.0-rc1`, incolla nella descrizione il blocco **Testo pronto per GitHub Release** (dalla prima sezione fino alla riga di separazione `---`).
