# NeuralPathTacticalGuard - Report Architetturale (Cursor)

## 1) Panoramica struttura

Il progetto e' una applicazione desktop Go/Fyne organizzata in un solo package `main`, con questi moduli logici:

- `main.go`: orchestrazione UI, ciclo di monitoraggio, notifiche, export report, gestione modalita' test/reale.
- `logic.go`: stato dominio (`State`) e regole di aggiornamento allarmi (`updateLogic`).
- `real_network.go`: probing rete reale con `net.DialTimeout` su TCP/53.
- `mock.go`: rete simulata thread-safe per demo/test manuali.
- `config.go`: configurazione applicativa (`config.json`), default, persistenza e directory operative.
- `licenze.go`: trial/pro, validazione licenza con HMAC-SHA256, attivazione/disattivazione.
- `logger.go`: logging su file locale con mutex.
- `packaging/*`, `main.manifest`, `wintun.dll`, `*.syso`: packaging Windows e asset runtime.

## 2) Collegamento componenti "neurali" e Tactical Guard

Nel sorgente analizzato non emergono componenti neurali/ML reali (nessun modello, inferenza o training).  
"NeuralPath" appare come naming/prodotto, mentre la parte Tactical Guard e' implementata come monitor operativo di connettivita':

1. `main()` carica config/licenza e sceglie backend rete (mock o reale).
2. Una goroutine esegue periodicamente `updateLogic(...)`.
3. `updateLogic` usa `DetectDevice()` + `Ping()` per calcolare stato link principale/back-up.
4. Lo stato aggiorna UI, grafico storico, allarmi sessione, log e notifiche desktop.
5. Report TXT/CSV vengono esportati da UI su richiesta utente.

In sintesi: connessione "neurale -> tactical" attualmente e' di branding/interfaccia, non di pipeline computazionale neurale.

## 3) Mappa dipendenze ad alto livello

- `main.go` dipende da: `logic.go`, `config.go`, `licenze.go`, `logger.go`, `mock.go`, `real_network.go`, `fyne.io/fyne/v2`.
- `logic.go` dipende dall'interfaccia `Network` implementata da `RealNetwork` e `MockNetwork`.
- `real_network.go` dipende da standard library (`net`, `time`).
- `licenze.go` dipende da standard library crypto (`hmac`, `sha256`) e filesystem locale.
- `logger.go`/`config.go` dipendono da filesystem locale.

## 4) Vulnerabilita' e rischi evidenti

### Alta priorita'

- **Segreto licenza placeholder in codice**: `licenseSecrets` contiene un valore di esempio; se non sostituito prima della build, la protezione commerciale risulta debole.
- **Possibili data race tra goroutine monitor e callback UI**: variabili condivise (`countAllarmi`, `lastState`, `starlinkHistory`, `deviceHistory`, `startTime`) vengono lette/scritte da contesti diversi senza sincronizzazione esplicita.

### Media priorita'

- **`cfg` globale non protetta**: il ciclo monitor legge parametri (`LagThresholdMs`, IP, intervallo) mentre altri handler possono modificare/salvare config/licenza.
- **Permessi file permissivi (`0644`)** su `config.json`, `license.key` e log/report: in ambienti multi-utente possono esporre metadati sensibili.
- **Manifest con `requireAdministrator`**: aumenta la superficie d'impatto in caso di abuso rispetto al principio del minimo privilegio.

### Bassa priorita'

- **Ping applicativo su TCP/53**: e' un check di raggiungibilita' socket, non un ICMP ping reale; possibili falsi positivi/negativi in presenza di firewall/DNS policy.
- **Gestione errori non sempre esplicita** in punti non critici (es. fallback silenzioso immagini embed).

## 5) Colli di bottiglia computazionali

- **Ridisegno completo del grafico ad ogni tick**: `refreshGraph` ricrea tutti gli oggetti canvas, con costo CPU/allocazioni proporzionale alla frequenza di refresh.
- **PiU' dial TCP per ciclo** (Starlink + probe device): con timeout elevati o rete degradata il ciclo puo' accumulare latenza percepita.
- **Loop continuo con `time.Sleep`**: semplice e robusto, ma senza backpressure/adaptive interval in presenza di timeout ripetuti.

## 6) Conclusione operativa

Il sistema e' coerente come monitor tactical desktop, ma oggi non include un layer neurale reale nel codice Go.  
Le priorita' tecniche immediate sono:

1. mettere in sicurezza licenze/segreti;
2. eliminare race condition sui dati condivisi;
3. ottimizzare refresh grafico e probing rete per ridurre overhead.

