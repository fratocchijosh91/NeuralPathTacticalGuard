===============================================================
  🛰️  NEURALPATH TACTICAL GUARD v2.2
  Sviluppato da: Josh Fratocchi
===============================================================

DESCRIZIONE
-----------
NeuralPath Tactical Guard e' un monitor di rete avanzato progettato
per il controllo della stabilita' del link primario (Starlink / Fibra)
e del backup via connessione cellulare.

Il software gestisce automaticamente il failover tra le connessioni,
garantendo la continuita' del servizio in caso di degradazione o
perdita del collegamento principale.

---------------------------------------------------------------
REQUISITI DI SISTEMA
---------------------------------------------------------------
  - Sistema operativo : Windows 10 / Windows 11
  - Architettura      : x86 (32-bit) o x64 (64-bit)
  - Privilegi         : Amministratore (richiesti per il ping
                        di precisione e la gestione di rete)

---------------------------------------------------------------
INSTALLAZIONE
---------------------------------------------------------------
1. Esegui il file Setup.exe come Amministratore.
2. Segui le istruzioni guidate dell'installatore.
3. Al termine, avvia NeuralPath Tactical Guard dal menu Start
   o dal collegamento sul desktop (se selezionato).

---------------------------------------------------------------
PRIMO AVVIO
---------------------------------------------------------------
Al primo avvio, il programma rileva automaticamente la tua rete.

Se desideri personalizzare i nomi visualizzati o gli indirizzi IP
monitorati, apri il file config.json presente nella cartella di
installazione con un editor di testo (es. Notepad).

  Esempio config.json:
  {
    "primary_name": "AUTO",        <- lascia AUTO per il rilevamento
    "backup_name":  "SIM_4G",
    ...
  }

Per forzare il rilevamento automatico della rete primaria,
imposta il campo "primary_name" su "AUTO".

---------------------------------------------------------------
FILE DI LOG
---------------------------------------------------------------
Il software genera automaticamente i seguenti file di log nella
cartella di installazione:

  - neuralpath.log              Log generale dell'applicazione
  - neuralpath_tactical.log     Log eventi di failover e rete
  - neuralpath_log_AAAA-MM-GG.txt  Log giornaliero dettagliato

I log rimangono esclusivamente in locale e non vengono mai
trasmessi a server esterni.

---------------------------------------------------------------
DISINSTALLAZIONE
---------------------------------------------------------------
Per rimuovere il software:
  Pannello di controllo → Programmi → Disinstalla un programma
  → NeuralPath Tactical Guard → Disinstalla

---------------------------------------------------------------
NOTE LEGALI
---------------------------------------------------------------
Consultare i file EULA.txt e LICENSE.txt inclusi in questa
distribuzione per i termini di licenza e le informazioni sui
componenti di terze parti.

---------------------------------------------------------------
CONTATTI
---------------------------------------------------------------
  Sviluppatore: Josh Fratocchi
  Prodotto    : NeuralPath Tactical Guard v2.2

===============================================================
Copyright (c) 2026 Josh Fratocchi. Tutti i diritti riservati.
===============================================================
