// **Ein Syntaxfehler steht im englischen Fenster auf Englisch** —
// #1178,
// die erste Abnahme.
//
// # Der Befund
//
// Der Satz des Zerteilers kam als fertiger deutscher Text aus
// `internal/msescript` und stand im Platzhalter `{meldung}` — auch im
// englischen Katalog.
import { describe, expect, it } from 'vitest'

import { Bindungsfehler } from './bindungsfehler'
import { kataloge } from './sprache/kataloge'
import { meldungAus } from './meldung'

// syntaxfehler baut den Fehler, wie ihn `internal/app` schickt.
function syntaxfehler(): Bindungsfehler {
  return new Bindungsfehler({
    code: 'validierung',
    schluessel: 'skriptFehler',
    meldung: 'Zeile 1, Spalte 5: nach `if` fehlt `then`',
    werte: {
      meldung: 'nach `if` fehlt `then`',
      skriptschluessel: 'fehltNach',
      skriptwert0: '`if`',
      skriptwert1: '`then`',
    },
  })
}

describe('ein Syntaxfehler', () => {
  it('steht im deutschen Fenster auf Deutsch', () => {
    expect(meldungAus(syntaxfehler(), kataloge.de)).toBe(
      'Der Ausdruck ist nicht auswertbar: nach `if` fehlt `then` ' +
        'Berichtige ihn an der genannten Stelle.',
    )
  })

  it('steht im englischen Fenster auf Englisch', () => {
    expect(meldungAus(syntaxfehler(), kataloge.en)).toBe(
      'The expression cannot be evaluated: after `if`, `then` is missing ' +
        'Correct it at the place named.',
    )
  })

  // **Ein Satz der Standardbibliothek bleibt deutsch** — mit Grund: die 48
  // Meldungen nennen Funktions- und Argumentnamen, und ein Katalogeintrag je
  // Funktion wäre eine zweite Abschrift der Bibliothek in zwei Sprachen.
  it('behält den durchgereichten Satz', () => {
    const fehler = new Bindungsfehler({
      code: 'validierung',
      schluessel: 'skriptFehler',
      meldung: 'Zeile 1, Spalte 1: contains braucht ein Argument `match:`',
      werte: {
        meldung: 'contains braucht ein Argument `match:`',
        skriptschluessel: 'durchgereicht',
        skriptwert0: 'contains braucht ein Argument `match:`',
      },
    })

    expect(meldungAus(fehler, kataloge.en)).toBe(
      'The expression cannot be evaluated: contains braucht ein Argument `match:` ' +
        'Correct it at the place named.',
    )
  })

  // **Und ein Schlüssel, den der Katalog nicht kennt, fällt auf den deutschen
  // Satz zurück** (§9): eine unübersetzte Meldung ist besser als eine leere.
  it('fällt auf den deutschen Satz zurück', () => {
    const fehler = new Bindungsfehler({
      code: 'validierung',
      schluessel: 'skriptFehler',
      meldung: 'kaputt',
      werte: { meldung: 'etwas Neues', skriptschluessel: 'gibtsnochnicht' },
    })

    expect(meldungAus(fehler, kataloge.en)).toContain('etwas Neues')
  })

  // **Ein Schlüssel ohne `meldung` gibt den leeren Satz** — er kommt aus Go
  // nie so, und ein `undefined` in der Anzeige wäre die schlechtere Antwort.
  it('kommt ohne den deutschen Satz aus', () => {
    const fehler = new Bindungsfehler({
      code: 'validierung',
      schluessel: 'skriptFehler',
      meldung: 'kaputt',
      werte: { skriptschluessel: 'gibtsnochnicht' },
    })

    expect(meldungAus(fehler, kataloge.en)).toBe(
      'The expression cannot be evaluated:  Correct it at the place named.',
    )
  })

  // **Ein Platzhalter ohne Wert bleibt stehen.** Er ist sichtbar falsch; ihn
  // durch nichts zu ersetzen versteckte den Fehler.
  it('lässt einen Platzhalter ohne Wert stehen', () => {
    const fehler = new Bindungsfehler({
      code: 'validierung',
      schluessel: 'skriptFehler',
      meldung: 'kaputt',
      werte: { meldung: 'roh', skriptschluessel: 'fehltNach' },
    })

    expect(meldungAus(fehler, kataloge.en)).toContain('{0}')
  })
})
