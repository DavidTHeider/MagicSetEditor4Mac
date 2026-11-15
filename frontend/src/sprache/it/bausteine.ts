import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert:
      'Nel frattempo questo campo riporta «{wert}», modificato altrove. La tua bozza non è ancora salvata.',
    behalten: 'Salva la mia bozza',
    nehmen: 'Prendi il nuovo valore',
  },
  listenansicht: {
    suche: 'Cerca',
    suchePlatzhalter: 'Nome o gioco',
    sortierung: 'Ordina',
    ansicht: 'Vista',
    liste: 'Elenco',
    kacheln: 'Riquadri',
    zahl: '{gezeigt} di {gesamt}',
    nichtsPasst: 'Nessun risultato per «{suche}».',
    sucheLeeren: 'Cancella la ricerca',
    filterZuruecksetzen: 'Reimposta ricerca e filtri',
    weitere: 'Mostra altri {anzahl}',
    wieImBestand: 'Come salvato',
    zuletztBenutzt: 'Prima gli usati di recente',
    nachName: 'Per nome',
    nachKarten: 'Per numero di carte',
  },
  speicherstand: {
    schreibt: 'Salvataggio …',
    geschrieben: 'Ogni modifica è salvata nella libreria.',
    fehler: '«{feld}» non è salvato — il motivo è indicato nel campo.',
  },
  gegenueberTitel: 'Aggiungi o conferma una controparte',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Cerca una controparte …',
  feldhilfe: 'Spiegazione di {feld}',
  mehrZu: 'Altro su «{satz}»',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'Ripreso dal profilo {quelle}.',
  herkunftAusEinfuhr: 'Importato da {quelle}.',

  pflichtfeldLeer: '{feld} richiede un valore.',

  waehrung: {
    beispiel: 'ad es. EUR',
    keineGestalt:
      '«{wert}» non è un codice valuta — è composto da tre lettere, come EUR.',
    grossSchreiben:
      '«{wert}» — un codice valuta si scrive in maiuscolo: {gross}.',
  },

  drucksprachen: {
    beispiel: 'ad es. en',
    namen: {
      en: 'Inglese',
      es: 'Spagnolo',
      fr: 'Francese',
      de: 'Tedesco',
      it: 'Italiano',
      pt: 'Portoghese',
      ja: 'Giapponese',
      ko: 'Coreano',
      ru: 'Russo',
      zhs: 'Cinese semplificato',
      zht: 'Cinese tradizionale',
      he: 'Ebraico',
      la: 'Latino',
      grc: 'Greco antico',
      ar: 'Arabo',
      sa: 'Sanscrito',
      ph: 'Phyrexiano',
      qya: 'Quenya',
      dw: 'Nanico',
      tlh: 'Klingon',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'Mostra la password in chiaro',
    zeigenKurz: 'Mostra',
    verbergen: 'Nascondi di nuovo la password',
    verbergenKurz: 'Nascondi',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'Codice',
    abschicken: 'Conferma',
  },

  aktionen: {
    abbrechen: 'Annulla',
    oeffnen: 'Apri',
    schliessen: 'Chiudi',
    speichern: 'Salva',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Scegli uno stile —',
  },
  druckwahl: {
    drucke: 'Stampe',
    hinweis:
      'Questa controparte vede solo queste stampe, con la quantità che possiedi.',
    suchen: 'Cerca stampe',
    gewaehlt: 'Selezione: {anzahl}',
    nurSuche: 'La tua collezione ha {anzahl}. Cerca un nome per sceglierne.',
    keiner: 'Nessuna stampa corrisponde a «{suche}».',
    genauer:
      'Corrispondono più stampe di quelle mostrate. Restringi la ricerca.',
    leer: 'La tua collezione non ha ancora stampe che potresti condividere. Aggiungi prima degli esemplari.',
    fehlt:
      'Scegli almeno una stampa: una selezione senza stampe non condivide nulla.',
    behalten: 'Esemplari di ogni stampa che tengo',
    behaltenHinweis:
      'Questa controparte vede ciò che supera questo numero: con 1 e tre esemplari, due.',
    detailBehalten: 'tieni {behalten}',
  },
  suchauswahl: {
    keiner: 'Nessun risultato per «{suche}».',
    treffer: 'Risultati: {anzahl}',
    begrenzt:
      'Mostrati {gezeigt} risultati su {anzahl} — restringi la ricerca …',
  },
  zeitzonenwahl: {
    system: 'Come il sistema',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Crea', heisst: 'Crea qualcosa di nuovo.' },
      bearbeiten: { kurz: 'Modifica', heisst: 'Lo modifica sul posto.' },
      sichern: { kurz: 'Salva', heisst: 'Conserva lo stato attuale.' },
      zuruecknehmen: {
        kurz: 'Annulla',
        heisst: 'Annulla l’ultimo passaggio.',
      },
      duplizieren: {
        kurz: 'Duplica',
        heisst: 'Ne mette una seconda accanto.',
      },
      kopieren: { kurz: 'Copia', heisst: 'La mette negli appunti.' },
      entfernen: {
        kurz: 'Rimuovi',
        heisst: 'La toglie da questa vista; altrove resta.',
      },
      loeschen: {
        kurz: 'Elimina',
        heisst: 'La elimina per sempre, ovunque.',
      },
    },
  },
  nurlesen: {
    hinweis:
      'Stai vedendo questo profilo come lettore: puoi guardare tutto e non modificare nulla. Passa al tuo profilo per lavorare — oppure chiedi a chi gestisce questo profilo il ruolo «collaboratore».',
    knopf: 'Sola lettura: in questo profilo non puoi modificare nulla.',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: 'Contesto di lavoro',
    profilVon: 'di {eigner}',
    rollen: {
      leser: 'Lettore',
      mitarbeiter: 'Collaboratore',
      verwalter: 'Gestore',
    },
    nurLesen: 'sola lettura',
    zuDenNutzern: 'Cambia profilo …',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Caricamento …',
    leer: 'Niente qui',
    fehler: 'Impossibile caricare.',
    fehlerMitGrund: 'Impossibile caricare. {grund}',
    wiederholen: 'Carica di nuovo',
    neuLaden: 'Ricarica',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'Cerca nel testo',
    schliessen: 'Chiudi la ricerca',
    suchen: 'Cerca',
    ersetzen: 'Sostituisci con',
    treffer: '{bei} di {ganz}',
    keinTreffer: 'Nessuna corrispondenza',
    weiter: 'Corrispondenza successiva',
    zurueck: 'Corrispondenza precedente',
    einsetzen: 'Sostituisci',
    alleEinsetzen: 'Sostituisci tutto',
    grossKlein:
      'Maiuscole e minuscole contano — in MSE-Script «card» e «Card» sono due cose diverse.',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Quali carte',
    ganzesSet: 'L’intero set ({zahl} carte)',
    ganzesSetOhneZahl: 'L’intero set',
    teilmenge: 'Solo le selezionate ({zahl})',
    ohneNamen: 'senza nome',
    eingrenzen: 'Filtra le carte',
  },

  vorschlaege: {
    bezeichnung: 'Suggerimenti',
    uebernehmen: 'Accetta con ↹',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: 'Larghezza di “{region}”',
    breiteWert: '{punkte} punti',
    navigatorAuf: 'Mostra il navigatore',
    navigatorZu: 'Nascondi il navigatore',
    inspectorAuf: 'Mostra l’ispettore',
    inspectorZu: 'Nascondi l’ispettore',
    navigatorEng:
      'Troppo stretto per tutte e tre le colonne — il navigatore si apre sopra la parte centrale',
    inspectorEng:
      'Troppo stretto per tutte e tre le colonne — l’ispettore si apre sopra la parte centrale',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: 'provvisoria',
    vorlaeufigErklaerung:
      'Questa stampa non ha ancora un identificativo Scryfall. Prezzi e immagini restano assenti finché non viene risolta.',
  },
  kommandozentrale: {
    titel: 'Centro comandi',
    erklaerung:
      'Un comando è uno strumento del server MCP, richiamato qui nella finestra — la stessa risposta che si ottiene tramite MCP e dalla riga di comando. Niente passa da una shell: ciò che digiti resta un dato.',
    lesend: 'legge',
    schreibend: 'modifica',
    gefaehrlich: 'difficile da annullare',
    faehigkeit: 'Funzionalità',
    cli: 'Riga di comando',
    rest: 'Servizio',
    parameter: 'Parametri',
    keineParameter: 'Questo comando non richiede parametri.',
    pflicht: 'Obbligatorio',
    jsonHinweis:
      'Come JSON — un elenco tra parentesi quadre, un oggetto tra parentesi graffe.',
    befundPflicht: 'Questo campo richiede un valore.',
    befundZahl: 'Non è un numero.',
    befundJson: 'Non è JSON.',
    vorschau: 'Cosa verrà inviato',
    bestaetigen: 'Ho letto l’anteprima e voglio eseguirlo così com’è.',
    ausfuehren: 'Esegui',
    abbrechen: 'Annulla',
    laeuft: 'Il comando è in esecuzione …',
    ergebnis: 'Risultato',
    ausgabe: 'Output del comando',
    alsText: 'Testo',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      'Ancora niente eseguito. I dati riservati — token, password — non compaiono mai qui in chiaro.',
    verlauf: 'Cronologia di questa sessione',
    wiederholen: 'Ricarica {name}',
    gelungen: 'riuscito',
    gescheitert: 'non riuscito',
    abgebrochen: 'annullato',
    schliessen: 'Chiudi',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Apri diagnostica',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'Configura l’IA',
  },
  kikasten: {
    titel: 'IA per questa carta',
  },
} as const
