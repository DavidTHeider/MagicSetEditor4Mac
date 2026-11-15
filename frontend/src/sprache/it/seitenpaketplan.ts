// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: 'Area di lavoro dei pacchetti',
  navigator: 'Ricerca e filtri',
  inspector: 'Pacchetto',
  zentrum: 'Pacchetti',

  herkunft: 'Origine',
  herkunftAlle: 'Qualsiasi origine',
  herkunftsarten: {
    mitgeliefert: 'incluso',
    bezogen: 'scaricato',
    uebernommen: 'importato',
    eigen: 'proprio',
  },

  spalteStand: 'Stato',
  standEingebunden: 'installato',
  standGeaendert: 'modificato in locale',

  inspectorLeer:
    'Scegli un pacchetto nell’elenco — qui compaiono la sua origine, la versione, le dipendenze e cosa comporterebbe una modifica.',
  kopf: '{art} · versione {version}',
  kopfOhneFassung: '{art} · senza versione',
  eingebunden: 'Installato',
  keinSet: 'Nessun set di questo profilo lo usa.',
  einSet: 'Lo usa un set.',
  vieleSets: 'Lo usano {anzahl} set.',
  setsUnbekannt:
    'Quali set ne hanno bisogno non è indicato nell’intestazione di alcun set.',
  geaendert:
    'Modificato in locale dopo l’installazione. Prima di sostituirlo viene chiesta conferma; la modifica viene conservata come versione precedente.',
  woher: 'Provenienza',
  woherMitgeliefert: 'Incluso nell’applicazione.',
  woherBezogen: 'Scaricato da {herkunft}.',
  woherUebernommen: 'Importato da {herkunft}.',
  woherEigen: 'Creato qui o inserito a mano nella cartella dei pacchetti.',
  pruefsumme: 'Checksum',
  lizenz: 'Licenza',
  lizenzUnbekannt:
    'non indicata — un pacchetto MSE non ha un campo licenza; i contenuti appartengono ai rispettivi titolari dei diritti',
  ort: 'Posizione',
  haengtAb: 'Dipende da',
  gebrauchtVon: 'Necessario a',
  nichts: 'nulla',

  einbinden: 'Installa pacchetto …',
  beziehen: 'Scarica un pacchetto …',
  beziehenTitel: 'Scarica un pacchetto',
  loeschen: 'Elimina …',
  zurueck: 'Ripristina la versione precedente …',
  weitergeben: 'Condividi come installer …',
  weitergegeben: 'Salvato come installer: {datei}',

  titel: {
    loeschen: 'Elimina {paket}',
    einbinden: 'Installa {paket}',
    ersetzen: 'Sostituisci {paket}',
    zurueck: 'Riporta {paket} alla versione precedente',
  },
  fassung: 'Versione {von} → {zu}',
  ohneFassung: 'senza versione',
  sets: 'Set interessati',
  keineSets: 'Nessun set di questo profilo è interessato.',
  konflikte: 'Non più soddisfatte dopo',
  keineKonflikte: 'Nessuna dipendenza si rompe.',
  planGeaendert:
    'Il pacchetto è stato modificato in locale dopo l’installazione.',
  ueberschreiben:
    'Sostituisci la modifica locale — viene conservata come versione precedente',
  nichtMoeglich: {
    loeschen:
      'Questo pacchetto non si trova nella cartella gestita. Qui viene eliminato solo ciò che questa applicazione ha installato.',
    einbinden: 'Questo pacchetto non può essere installato qui.',
    ersetzen: 'Questo pacchetto non può essere sostituito qui.',
    zurueck:
      'Non esiste una versione precedente. Ne viene conservata una quando installi una versione più recente sopra questo pacchetto.',
  },
  bestaetigen: {
    loeschen: 'Elimina',
    einbinden: 'Installa',
    ersetzen: 'Sostituisci',
    zurueck: 'Ripristina',
  },
  verwerfen: 'Non modificare nulla',
  bedarf: {
    ueberschrift: 'Pacchetti che mancano a questo set',
    erklaerung:
      'Senza di essi l’anteprima disegna un ripiego o omette ciò che il pacchetto fornisce. Installa il pacchetto da una cartella — oppure cercalo tra i pacchetti per scaricarlo o importarlo da MSE2.',
    suchen: 'Cerca tra i pacchetti',
  },
  quelle: 'Da {quelle}',
  keineRechte:
    'Un pacchetto porta con sé modelli e MSE-Script. Lo script non ha accesso a file, rete o programmi, e ciò di cui il pacchetto ha bisogno non viene scaricato insieme — compare sopra come conflitto.',
}
