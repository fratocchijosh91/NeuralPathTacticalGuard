# Staging Checklist (License Server)

## Pre-deploy (Railway)

- [ ] Keypair generato: `go run ./cmd/license-keygen`
- [ ] Private key salvata come secret (non nel repo)
- [ ] Public key copiata nella config client (`license_public_key`)
- [ ] Chiavi licenza autorizzate configurate (`NP_LICENSE_KEYS`)
- [ ] `railway.json` commitato e pushato su GitHub
- [ ] Rate limit configurato (`NP_LICENSE_RATE_LIMIT_PER_MIN`, default 10)
- [ ] Secret webhook Stripe configurato (`NP_STRIPE_WEBHOOK_SECRET`) se usi checkout live
- [ ] Persistenza chiavi configurata (`NP_LICENSE_KEYS_PATH`)
- [ ] API key admin configurata (`NP_ADMIN_API_KEY`)

## Deploy (Railway)

- [ ] Servizio Railway creato dal repository GitHub
- [ ] Variables compilate in Railway (`NP_LICENSE_PRIVATE_KEY_B64`, `NP_LICENSE_KEYS`)
- [ ] Deploy completato senza errori

## Verifica post-deploy

- [ ] `GET /healthz` ritorna `{"status":"ok"}`
- [ ] `GET /v1/public-key` ritorna la public key corretta
- [ ] `POST /v1/licenses/activate` con chiave valida ritorna token
- [ ] `POST /v1/licenses/activate` con chiave invalida ritorna 401
- [ ] Rate limit funzionante (11 richieste in 1 minuto -> 429)
- [ ] Audit log visibile nei log Railway (AUDIT ACTIVATE_OK / ACTIVATE_FAIL)
- [ ] Webhook Stripe testato con firma valida (`/v1/webhooks/stripe`) e chiave aggiunta in allowlist
- [ ] Endpoint admin testato (`/v1/admin/licenses/create`) con API key valida e invalida

## Collegamento client app

- [ ] `license_server_url` punta a URL staging Railway
- [ ] `license_public_key` corrisponde alla key del server
- [ ] Attivazione PRO dall'app funziona end-to-end
- [ ] Disattivazione licenza dall'app funziona
- [ ] Trial senza licenza funziona normalmente

## Produzione (Stripe Live)

- [ ] Dashboard Stripe in **modalità Live** (non Test)
- [ ] Webhook Live creato verso `https://<railway-prod>/v1/webhooks/stripe` con evento `checkout.session.completed`
- [ ] `NP_STRIPE_WEBHOOK_SECRET` su Railway = signing secret **Live** (`whsec_...` dalla pagina del webhook Live)
- [ ] (Opzionale) `NP_STRIPE_WEBHOOK_TOLERANCE_SEC` impostato se serve allungare la finestra (default 300 s)
- [ ] Payment Link / Checkout configurati in **Live** e pagamento di prova completato
- [ ] Log Railway: `AUDIT STRIPE_WEBHOOK_OK` dopo pagamento; allowlist aggiornata (`data/allowed-keys.json` nel volume)
- [ ] Client: `license_server_url` e `license_public_key` puntano al **deploy di produzione** che firma i token
