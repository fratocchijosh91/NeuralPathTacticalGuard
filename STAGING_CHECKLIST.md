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
