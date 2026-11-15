// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'La tua panoramica',
    mitProblemen: '{anzahl} con problemi',
    geprueftVon: '{geprueft} di {gesamt} verificati',
    mitBefund: '{anzahl} con rilievi',
    nichtGeprueft: 'non verificato del tutto',
    problemeGeprueft:
      'Problemi verificati {wann} · {geprueft} di {gesamt} carte',
    problemeOffen: 'Problemi: {offen} set su {gesamt} non verificati',
    pruefen: 'Verifica ora',
    setsMitProblemen: '{sets} con problemi',
    kartenMitBefund: '{karten} con rilievi in {sets}',
    aeltesterPreis: 'prezzo più vecchio {wann}',
    ohnePreis: '{anzahl} senza prezzo',
    mitFehlenden: '{anzahl} con carte mancanti',
    decksMitFehlenden: '{decks} con carte mancanti',
    alter: {
      gerade: 'proprio ora',
      minuten: '{minuten} min fa',
      stunden: '{stunden} h fa',
      tage: '{tage} giorni fa',
      tag: '1 giorno fa',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: 'Richiede attenzione',
    nichts: 'Nulla — {sets}, {decks} e {pakete} controllati.',
    durchgesehen: '{sets}, {decks} e {pakete} controllati.',
    arten: {
      setVorbehalte: '{anzahl} non interpretati',
      deckFehlend: '{anzahl} mancanti nella collezione',
      paketBefund: '{anzahl} negli script',
    },
    paketeMitBefund: '{pakete} con rilievi nello script',
    gruendeZeigen: 'Mostra i motivi',
    gruendeVerbergen: 'Nascondi i motivi',
    setBefunde: '{karten} con rilievi',
    befunde: 'Non controllati:',
    alle: 'Mostra tutti i {anzahl}',
    weniger: 'Mostra meno',
    gruende: {
      vorbehalt: '{0}, riga {1}',
    },
    amOrt: '{ort}: {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'Restano solo {frei} liberi accanto alla raccolta',
        sicherungAlt: 'Ultimo backup {tage} giorni fa',
        dienstBeendet:
          'Il servizio all’accesso è terminato con il codice {ausgang}',
        webprozessNeu: 'La vista web è stata ricaricata {anzahl} volte',
      },
      rat: {
        platzKnapp:
          'Sotto circa 2 GB o il 5 % liberi ogni backup diventa stretto — libera spazio in «Dati e backup».',
        sicherungAlt:
          'Più vecchio di 14 giorni — esegui il backup della raccolta in «Dati e backup».',
        dienstBeendet:
          'Dopo un errore launchd lo riavvia — la diagnostica dice perché è terminato.',
        webprozessNeu:
          'WebKit aveva terminato il suo processo; ciò che non era salvato in un modulo aperto può mancare — controllalo nella diagnostica.',
      },
    },
  },
  auftraege: {
    titel: 'Attività in background',
    staende: {
      laeuft: 'in corso',
      fertig: 'completata',
      gescheitert: 'non riuscita',
      abgebrochen: 'annullata',
    },
    arten: {
      preise: 'Scarica i prezzi per la collezione',
      sammeldaten: 'Carica i dati in blocco da Scryfall',
      fassung: 'Scarica una nuova versione',
      bogen: 'Componi un foglio di stampa',
      set: 'Importa un set',
      mtgjson: 'Importa MTG JSON',
      vorlage: 'Esporta tramite un modello',
      installer: 'Leggi un installer',
      weitergabe: 'Condividi come installer',
      probleme: 'Verificare i set',
    },
    abbrechen: 'Annulla {art}',
    dauer: '{sekunden} s',
    laufend: { eine: 'attività in corso', viele: 'attività in corso' },
  },
  weiterarbeiten: {
    titel: 'Riprendi',
    weiterAn: 'Continua con {art} “{name}”',
    zuletztWann: 'Ultima volta {vorgang} il {zeit}',
    angeheftet: 'fissato',
    angeheftetLeer:
      'Nulla di fissato — fissa una riga per tenerla qui in alto.',
    leer: 'Qui compare ciò a cui hai lavorato di recente. Per ora è vuoto — il primo passo è un set o un pacchetto di modelli.',
    handlungSets: 'Crea o apri un set',
    handlungPakete: 'Scarica un pacchetto',
    anheften: 'Fissa',
    loesen: 'Sblocca',
    anheftenVon: 'Fissa “{name}”',
    loesenVon: 'Sblocca “{name}”',
    arten: { set: 'Set', deck: 'Mazzo', paket: 'Pacchetto' },
    vorgaenge: {
      geoeffnet: 'aperto',
      eingefuehrt: 'importato',
      ausgegeben: 'esportato',
    },
  },
  titel: 'Home',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Progetta, controlla e stampa carte per Magic e altri giochi — con i modelli che usa Magic Set Editor 2.',
  ersteSchritte: 'Primi passi',
  einfuehrung: {
    willkommen:
      'Tre passi e la prima carta è pronta. Puoi riaprire questa finestra in qualsiasi momento dalle impostazioni.',
    wiederzeigen: 'Mostra di nuovo all’avvio',
    schliessen: 'Cominciamo',
    bezeichnung: 'Introduzione',
  },
  schritte: {
    setOeffnen: {
      titel: 'Apri un set',
      erklaerung:
        'Apri un file .mse-set esistente — oppure importa un set da MTG JSON.',
      handlung: 'Vai ai set',
    },
    kartenBearbeiten: {
      titel: 'Modifica le carte',
      erklaerung:
        'Guarda e modifica le carte del set aperto; l’anteprima mostra subito ogni modifica.',
      handlung: 'Vai alle carte',
      ersteinSet: 'Per questo serve prima un set aperto.',
      stattdessen: 'Prima apri un set',
    },
    paketeHolen: {
      titel: 'Scarica i modelli',
      erklaerung:
        'Otto pacchetti di modelli sono già presenti. Vuoi più giochi o stili? Scaricali qui.',
      handlung: 'Vai ai pacchetti',
    },
  },
}
