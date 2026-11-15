// **Welche Meldung im Band steht, und wo** — S08.NAT-02 aus UX-08
// (#1680).
//
// Der Lauf durch das ganze Fenster steht in `appmeldungvorgang.test.tsx`;
// hier stehen die Fälle, die er nicht alle gehen kann.
import { describe, expect, it } from 'vitest'

import { bandmeldung, type Meldungslage } from './bandmeldung'

function lage(teile: Partial<Meldungslage>): Meldungslage {
  return {
    sprache: 'de',
    rueckspielrest: '',
    dateimeldung: undefined,
    verlaufsfehler: undefined,
    darstellungsfehler: undefined,
    bestandsfehler: undefined,
    bestandspfad: undefined,
    setID: '7',
    setname: 'Probe',
    seite: 'karten',
    speicherungFuer: '7',
    speicherfehler: undefined,
    speicherpfad: undefined,
    ...teile,
  }
}

describe('Die Meldung des Fensters', () => {
  it.each([
    [
      'nennt Set und Vorgang',
      lage({ speicherfehler: 'Kein Schreibrecht.' }),
      'Das Set „Probe“ wurde nicht gesichert — Kein Schreibrecht.',
    ],
    [
      'nennt den Vorgang, solange der Name fehlt',
      lage({ setname: undefined, speicherfehler: 'Kein Schreibrecht.' }),
      'Das Set wurde nicht gesichert — Kein Schreibrecht.',
    ],
    [
      'steht auf der Setliste',
      lage({ seite: 'sets', speicherfehler: 'Kein Schreibrecht.' }),
      'Das Set „Probe“ wurde nicht gesichert — Kein Schreibrecht.',
    ],
    [
      'nennt beim Bestand den Bestand',
      lage({ seite: 'hilfe', bestandsfehler: 'Die Platte ist voll.' }),
      'Der Bestand wurde nicht gesichert — Die Platte ist voll.',
    ],
    [
      'nennt beim Zurücknehmen das Zurücknehmen',
      lage({ seite: 'hilfe', verlaufsfehler: 'Nichts mehr da.' }),
      'Zurücknehmen oder Wiederherstellen ging nicht — Nichts mehr da.',
    ],
    // **Die Darstellung aus der Menüleiste gehört ins Band** — Befund G5 des
    // Zweitblicks über Phase 6.3 (#1665): sie stand nur unter *Einstellungen
    // › Erscheinungsbild*, wo niemand ist, der gerade die Menüleiste nahm.
    [
      'nennt beim Zurücksetzen der Darstellung die Darstellung',
      lage({
        seite: 'hilfe',
        darstellungsfehler: 'Der Bestand antwortet nicht.',
      }),
      'Die Darstellung ließ sich nicht zurücksetzen — Der Bestand antwortet nicht.',
    ],
    [
      'lässt das Zurücknehmen vor der Darstellung stehen',
      lage({
        verlaufsfehler: 'Nichts mehr da.',
        darstellungsfehler: 'Kaputt.',
      }),
      'Zurücknehmen oder Wiederherstellen ging nicht — Nichts mehr da.',
    ],
  ])('%s', (_, eingabe, erwartet) => {
    expect(bandmeldung(eingabe)).toEqual({ text: erwartet, art: 'schlecht' })
  })

  // **Die Meldung eines Sets steht nicht über einer fremden Seite** — und
  // die gelungene ebenso wenig: auch „Gesichert nach …" gehört zum Set.
  it.each(['hilfe', 'einstellungen', 'turniere', 'decks'] as const)(
    'steht nicht über %s',
    (seite) => {
      expect(
        bandmeldung(lage({ seite, speicherfehler: 'Kein Schreibrecht.' })),
      ).toBeUndefined()
      expect(
        bandmeldung(lage({ seite, speicherpfad: '/tmp/probe.mse-set' })),
      ).toBeUndefined()
    },
  )
})
