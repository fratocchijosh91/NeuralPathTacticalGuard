# Release Checklist Windows (NeuralPath Tactical Guard)

## 1) Sicurezza e licensing

- [ ] `license_public_key` e `license_server_url` configurati correttamente.
- [ ] Nessun segreto privato nel repository (`NP_LICENSE_PRIVATE_KEY_B64` solo lato server).
- [ ] Endpoint licensing in staging/prod raggiungibile (`/healthz`, `/v1/licenses/activate`).
- [ ] Verifica attivazione/disattivazione licenza su installazione pulita.

## 2) Qualità codice

- [ ] `go test ./...` verde.
- [ ] `go test -race ./...` verde.
- [ ] `go vet ./...` verde.
- [ ] Nessun linter error bloccante.

## 3) Build e packaging

- [ ] Build release con `packaging/build-release.ps1`.
- [ ] ZIP portabile generato in `artifacts/out`.
- [ ] SHA256 aggiornato e verificato.
- [ ] Installer Inno Setup generato.

## 4) Firma digitale

- [ ] Firma `.exe` con certificato code signing.
- [ ] Firma installer `.exe`.
- [ ] Verifica firma valida su macchina Windows target.

## 5) Test funzionali pre-rilascio

- [ ] Modalità test: online/lag/offline funzionanti.
- [ ] Modalità rete reale: rilevamento device coerente.
- [ ] Notifiche desktop e report TXT/CSV verificati.
- [ ] Avvio senza privilegi admin (se il manifest viene abbassato) o giustificazione esplicita se richiesto.

## 6) Documentazione e supporto

- [ ] `README` aggiornato con setup licenze.
- [ ] EULA/Privacy/Termini aggiornati nella release.
- [ ] Canale supporto pronto (email o helpdesk).
- [ ] Changelog versione pubblicato.
