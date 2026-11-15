// **Ein Beleg und ein Paket, wie der Dienst sie schickt** — die Probe für die
// Tests der Kontextfreigabe (A16,
// #1515).
//
// **Eine Stelle und nicht zwei Kopien**: die Gestalt des Pakets ist die des
// Dienstes, und zwei Tests mit je eigenem Nachbau liefen auseinander, sobald
// ein Feld hinzukäme.

// einBeleg ist ein Ausschnitt samt Herkunft.
export function einBeleg(anders: Record<string, unknown> = {}) {
  return {
    kennung: 'B1',
    quellkennung: 'regeln:702.19:0',
    quelle: 'regelwerk',
    kennungssystem: '',
    fachkennung: '702.19',
    semantik: '',
    titel: 'Trampelschaden',
    ausschnitt: '702.19a Trampelschaden ändert die Zuweisung.',
    inhaltshash: 'deadbeef',
    sprache: 'de',
    regelnummer: '702.19',
    urheber: '',
    bezugsadresse: 'https://beispiel.invalid/cr',
    quellrevision: '2026-07-26',
    quelldatum: '2026-07-26',
    abgerufen: '2026-09-11',
    veraltet: false,
    ...anders,
  }
}

// einPaket baut eine vollständige Antwort.
export function einPaket(anders: Record<string, unknown> = {}) {
  return {
    paketID: 'kp-1',
    aufgabe: 'Erkläre Trampelschaden',
    zielrevision: '',
    zaehlweise: 'zeichen/4',
    kontextgrenze: 4096,
    ausgabereserve: 1024,
    verbraucht: 120,
    kontextpolicy: 'kontext-1',
    korpusstaende: [],
    belege: [einBeleg()],
    freigabe: {
      zielanbieter: 'ollama',
      zielmodell: 'gemma3',
      ort: 'lokal',
      zieladresse: '',
      kategorien: ['aufgabe', 'oeffentlich'],
      belege: 1,
      zeichen: 400,
      tokens: 100,
      privat: false,
      satz: 'An ollama/gemma3 gingen 1 Beleg.',
    },
    laeuftab: '2026-09-12T18:00:00Z',
    ergebnis: {
      treffer: [],
      modus: 'hybrid',
      leere: '',
      grund: '',
      modell: '',
      modellrevision: '',
      suchkonfiguration: '',
      indexfingerabdruck: '',
      raumabdruck: '',
      ausgangsstand: '',
      korpusstaende: [],
      einschraenkungen: [],
      fortsetzung: '',
      fehler: null,
    },
    fehler: null,
    ...anders,
  }
}
