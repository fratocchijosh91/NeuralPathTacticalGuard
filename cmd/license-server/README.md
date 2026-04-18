# License Server (NeuralPath Tactical Guard)

Mini backend per attivazione licenze compatibile con il client in `licenze.go`.

## Endpoint

- `GET /healthz`
- `GET /v1/detected-devices` (lettura JSON dispositivi rilevati; file `NP_DETECTED_DEVICES_PATH`, formato `{"devices":[...]}` o array JSON — vedi `data/detected-devices.example.json`)
- `GET /v1/public-key`
- `POST /v1/licenses/activate`
- `POST /v1/webhooks/stripe`
- `POST /v1/admin/licenses/create`

Payload attivazione:

```json
{
  "license_key": "NP-PRO-XXXX-XXXX",
  "machine_id": "ABCDEF1234567890",
  "product": "neuralpath-tactical-guard",
  "version": "v2.1"
}
```

Risposta:

```json
{
  "token": "<payload_base64url>.<signature_base64url>"
}
```

## Variabili ambiente

- `NP_LICENSE_PRIVATE_KEY_B64` (obbligatoria): chiave privata ed25519 in base64 (seed 32 byte o private key 64 byte).
- `NP_LICENSE_ADDR` (default `:8080`)
- `NP_LICENSE_PRODUCT` (default `neuralpath-tactical-guard`)
- `NP_LICENSE_TIER` (default `PRO`)
- `NP_LICENSE_PREFIX` (default `NP`)
- `NP_LICENSE_TOKEN_TTL_HOURS` (default `720`, 30 giorni)
- `NP_LICENSE_ALLOW_ANY_KEY` (`true`/`false`, default `false`)
- `NP_LICENSE_KEYS` (lista chiavi separate da virgola, usata se `ALLOW_ANY_KEY=false`)
- `NP_LICENSE_KEYS_PATH` (default `data/allowed-keys.json`, file persistenza allowlist)
- `NP_DETECTED_DEVICES_PATH` (default `data/detected-devices.json`, JSON per app mobile / dashboard: lista ultimi dispositivi hotspot)
- `NP_LICENSE_RATE_LIMIT_PER_MIN` (default `10`, limite endpoint activate)
- `NP_STRIPE_WEBHOOK_SECRET` (se impostata abilita verifica firma webhook Stripe)
- `NP_STRIPE_WEBHOOK_TOLERANCE_SEC` (opzionale, default `300`): finestra anti-replay in secondi per l’header `Stripe-Signature` (allineato allo Stripe SDK, max effettivo 3600 → cap a 1h)
- `NP_ADMIN_API_KEY` (abilita endpoint admin per creazione chiavi)

## Webhook Stripe

Endpoint: `POST /v1/webhooks/stripe`

- Verifica header `Stripe-Signature` via HMAC-SHA256 con `NP_STRIPE_WEBHOOK_SECRET` (stringa completa incluso prefisso `whsec_`, come nel Dashboard / Stripe CLI).
- Controlla il timestamp dell’evento entro `NP_STRIPE_WEBHOOK_TOLERANCE_SEC` (default 5 minuti) per limitare replay.
- Valida tutte le firme `v1` presenti nell’header (rotazione segreti Stripe).
- Evento supportato: `checkout.session.completed` con `payment_status=paid`.
- Se `metadata.license_key` è presente, usa quella chiave.
- Altrimenti genera automaticamente una chiave `NP-PRO-<hash>`.
- La chiave viene aggiunta all'allowlist in memoria e persistita nel file `NP_LICENSE_KEYS_PATH`.

## Stripe Live (produzione)

1. **Dashboard Stripe**  
   Passa alla **modalità Live** (interruttore in alto a destra). Test e Live hanno **chiavi e webhook signing secret diversi**: non riusare `whsec_` di test in produzione.

2. **Endpoint webhook Live**  
   In [Developers → Webhooks](https://dashboard.stripe.com/webhooks) (contesto Live) aggiungi endpoint:  
   `https://<tuo-dominio-railway>/v1/webhooks/stripe`  
   Evento da inviare: **`checkout.session.completed`**.  
   Copia il **Signing secret** (`whsec_...`) e impostalo su Railway come **`NP_STRIPE_WEBHOOK_SECRET`** (sostituendo il valore di test).

3. **Checkout / Payment Link Live**  
   Crea il flusso di pagamento in **Live** (Payment Links, Checkout Session, ecc.).  
   Opzionale ma consigliato: in **metadata** imposta `license_key` con una chiave già nel formato `NP-PRO-...` se vuoi controllare tu la stringa; altrimenti il server la deriva da `client_reference_id`, email o `session.id`.

4. **Railway**  
   Ridistribuisci il servizio dopo aver aggiornato le variabili. Verifica nei log che compaia `webhookEnabled=true` e che dopo un acquisto reale compaia `AUDIT STRIPE_WEBHOOK_OK`.

5. **Staging vs produzione**  
   Per non mischiare acquisti test e live, conviene **due progetti Railway** (o due servizi) con coppie distinte di `NP_STRIPE_WEBHOOK_SECRET` e allowlist (`NP_LICENSE_KEYS_PATH` / volume se in futuro userai storage condiviso).

6. **Client desktop**  
   L’app non usa la secret Stripe: deve solo avere `license_server_url` e `license_public_key` del **server che firma i token** in produzione (solitamente lo stesso deploy Live del license server).

## Endpoint admin (manuale/supporto)

Endpoint: `POST /v1/admin/licenses/create` (protetto da API key)

Header:

- `Authorization: Bearer <NP_ADMIN_API_KEY>` oppure `X-API-Key: <NP_ADMIN_API_KEY>`

Payload esempio:

```json
{
  "reference": "order_12345",
  "email": "cliente@example.com"
}
```

Risposta:

```json
{
  "status": "ok",
  "license_key": "NP-PRO-ABCDEF123456"
}
```

Script helper locale:

```bash
./scripts/admin-create-license.sh --reference order_123 --email cliente@example.com
```

Script test webhook Stripe:

```bash
./scripts/test-stripe-webhook.sh --reference order_123 --email cliente@example.com
```

Script E2E completo (healthcheck + admin + activate + webhook):

```bash
./scripts/staging-e2e.sh
```

## Avvio locale

```bash
go run ./cmd/license-server
```

Gestione rapida processo locale:

```bash
./scripts/license-server-up.sh
./scripts/license-server-status.sh
./scripts/license-server-down.sh
```

## Setup client one-click (locale)

Per aggiornare automaticamente `config.json` del client con URL server e public key:

```bash
./scripts/dev-license-env.sh
./scripts/setup-local-license-client.sh
```

## Smoke test end-to-end

```bash
./scripts/dev-license-env.sh
./scripts/license-smoke-test.sh
```

Lo script:

1. avvia il server licenze locale;
2. chiama `POST /v1/licenses/activate`;
3. verifica il token firmato con `cmd/license-token-check`;
4. conferma compatibilità base con la validazione client.

## Nota deploy platform

- **Railway**: consigliata per `license-server` (servizio always-on + variabili + log).
- **Vercel**: consigliata per frontend/landing; non ideale come backend principale per questo servizio con persistenza allowlist locale.

Guida veloce in 5 comandi:

```bash
cat STAGING_5_COMMANDS.md
```

## Configurazione client app

All'avvio il server stampa nel log:

- `NP_LICENSE_PUBLIC_KEY_B64=<...>`

Quella stringa va impostata nel client (`config.json` o env):

- `license_public_key`
- `license_server_url` (esempio: `http://127.0.0.1:8080`)

oppure come env:

- `NP_LICENSE_PUBLIC_KEY_B64`
- `NP_LICENSE_SERVER_URL`
