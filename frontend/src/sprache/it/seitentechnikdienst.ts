// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Certificato proprio di {aussteller}, valido fino a {bis}',
  identitaet:
    'Certificato sull’identità di questo profilo — chi si connette ne confronta l’impronta.',
}

export const fernwege = {
  titel: 'Come sono collegate le controparti',
  keine: 'Nessuno ha ancora usato questo invito.',
  direkt: 'Diretta, tramite {adresse}',
  relais: 'Tramite il relay {adresse} — ancora senza percorso diretto',
  aufbau: 'Nota, ma non ancora raggiungibile attraverso alcun percorso',
  unbestimmt: 'Collegati: {anzahl} — il percorso non è ancora noto.',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Configurare la riga di comando e il server degli strumenti',
  satz: 'Un pulsante inserisce ciò che serve a un programma su questo Mac per trovare la riga di comando o il server degli strumenti — un secondo toglie esattamente quello. Prima vedi quale file e quali righe vengono toccati; le voci altrui restano come sono.',
  namen: {
    befehle: 'I collegamenti a mse e mse-mcp in ~/.local/bin',
    befehlszeile: 'La riga di comando nel PATH della shell di login',
    codex: 'Il server degli strumenti per Codex',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: 'Il servizio per script come LaunchAgent',
  },
  staende: {
    offen: 'non configurato',
    eingerichtet: 'configurato',
    veraltet: 'obsoleto',
    fremd: 'voce altrui',
    nichtGefunden: 'non presente su questo Mac',
    unnoetig: 'non necessario',
  },
  staendeSaetze: {
    veraltet: 'Ora l’applicazione si trova altrove; Configura lo ripara.',
    fremd: 'C’è già una voce altrui — resta com’è.',
    unnoetig: 'È già nel PATH.',
  },
  zeilen: 'Queste righe vengono aggiunte',
  vorschau: {
    json: 'Questa voce entra nel file',
    befehl: 'Viene eseguito questo comando',
    agent: 'Viene creato questo file e questo comando lo carica',
  },
  nurEintrag:
    '{name} era cambiato nel frattempo: è stata tolta solo la voce propria; il resto resta come l’hai lasciato.',
  einrichten: 'Configura',
  widerrufen: 'Revoca',
  alle: {
    titel: 'Tutto in una volta',
    satz: 'Configura ogni destinazione presente su questo Mac e ancora mancante, oppure ritira tutto ciò che è configurato. Le voci altrui restano come sono.',
    vorschau: 'Cosa verrà scritto ({anzahl} destinazioni)',
    vorschauLeer: 'Non manca nulla: è configurato ciò che esiste qui.',
    einrichten: 'Configura tutto',
    widerrufen: 'Ritira tutto',
    ergebnis: 'Cosa è successo',
    ausgaenge: {
      eingerichtet: 'configurato',
      widerrufen: 'ritirato',
      fremd: 'voce altrui, invariata',
      unveraendert: 'invariato',
      gescheitert: 'non riuscito',
      abgebrochen: 'annullato',
    },
  },
  lauschwahl: {
    titel: 'Dove ascoltano i servizi',
    satz: 'Vale per il servizio per script — avviato nella finestra e come LaunchAgent — e per il server degli strumenti via HTTP, dal loro prossimo avvio.',
    wahl: 'Raggiungibile tramite',
    schleife: 'Solo questo Mac (127.0.0.1)',
    auswahl: 'Indirizzi selezionati',
    alle: 'Tutte le interfacce (0.0.0.0)',
    alleSatz: 'Ogni interfaccia, anche una aggiunta in seguito.',
    aussen:
      'Un indirizzo scelto esce da questo Mac: i servizi richiedono allora TLS e un accesso (token), e macOS potrebbe chiedere il permesso per la rete locale.',
    inaktiv: 'spenta',
    ohneAdresse: 'nessun indirizzo',
    liste: 'Interfacce di questo Mac',
    arten: {
      schleife: 'Loopback',
      netz: 'Rete',
      tunnel: 'Tunnel (VPN)',
      bruecke: 'Bridge',
      direkt: 'Wireless diretto',
    },
    speichern: 'Applica la scelta',
    gespeichert: 'Applicato: ha effetto al prossimo avvio di un servizio.',
    keineGewaehlt: 'Scegli almeno un indirizzo.',
    platzhalter: 'come in «Dove ascoltano i servizi»',
    weitere: 'Ascolta anche su',
    ohne: 'Senza {adresse}: {grund}',
  },
}
