// **Der Speicher der Spieltexturen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Die Abnahme des Issues, wörtlich**: „Sieben gleiche Karten nutzen eine
// Textur, bleiben sieben Instanzen" — und „ein noch verwendetes Shared Asset
// wird nicht zerstört".
import { expect, it } from 'vitest'

import { Texturspeicher, texturschluessel } from './texturspeicher'

import type { Textur } from '../apitextur'

// eineTextur gibt eine Antwort, wie die Naht sie gibt.
function eineTextur(adresse: string): Textur {
  return {
    schluessel: 'abc',
    adresse,
    guete: 'tisch',
    seite: 'vorne',
    breite: 512,
    hoehe: 714,
    befunde: [],
    zweiseitig: false,
  }
}

// **Sieben gleiche Karten holen einmal und bleiben sieben.**
it('holt für sieben gleiche Karten einmal', () => {
  const speicher = new Texturspeicher()
  const schluessel = texturschluessel({
    art: 'karte',
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch',
  })

  const neu = [...Array<number>(7)].map(() => speicher.belegen(schluessel))

  // **Einmal neu, sechsmal nicht** — das ist die ganze Ersparnis.
  expect(neu.filter(Boolean)).toHaveLength(1)
  expect(speicher.zahl(schluessel)).toBe(7)
  expect(speicher.groesse()).toBe(1)
})

// **Und eine noch verwendete Textur wird nicht weggeworfen.**
it('wirft eine noch verwendete Textur nicht weg', () => {
  const speicher = new Texturspeicher()
  const schluessel = texturschluessel({
    art: 'karte',
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch',
  })

  speicher.belegen(schluessel)
  speicher.belegen(schluessel)
  speicher.setze(schluessel, eineTextur('/spielbild/abc.png'), speicher.runde())

  speicher.gib(schluessel)

  // **Eines ist weg, eines zeigt sie noch.**
  expect(speicher.zahl(schluessel)).toBe(1)
  expect(speicher.hole(schluessel)?.adresse).toBe('/spielbild/abc.png')

  speicher.gib(schluessel)

  expect(speicher.zahl(schluessel)).toBe(0)
  expect(speicher.hole(schluessel)).toBeUndefined()
  expect(speicher.groesse()).toBe(0)
})

// **Eine späte Antwort trägt sich nicht in einen weggeräumten Beleg ein.**
//
// **§7, Regel 2 an der Stelle, an der der Beleg liegt.** Sonst hielte der
// Speicher eine Textur, die niemand zeigt — und niemand gäbe sie je zurück.
it('trägt eine späte Antwort nicht mehr ein', () => {
  const speicher = new Texturspeicher()
  const schluessel = texturschluessel({
    art: 'karte',
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch',
  })

  speicher.belegen(schluessel)
  speicher.gib(schluessel)

  speicher.setze(schluessel, eineTextur('/spielbild/abc.png'), speicher.runde())

  expect(speicher.hole(schluessel)).toBeUndefined()
  expect(speicher.groesse()).toBe(0)
})

// **Abmelden, was nie angemeldet war, ist kein Fehler.**
//
// **Es kommt wirklich vor**: ein Bauteil ohne Kartenkennung meldet sich nicht
// an und beim Verschwinden trotzdem ab.
it('nimmt ein Abmelden ohne Anmeldung hin', () => {
  const speicher = new Texturspeicher()

  speicher.gib(
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    }),
  )

  expect(speicher.groesse()).toBe(0)
})

// **Drei Angaben, drei Schlüssel.**
//
// **Sonst zeigte eine Miniatur das Bild der Betrachtungsgröße** — und
// umgekehrt.
it('trennt Karte, Seite und Güte', () => {
  const schluessel = [
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    }),
    texturschluessel({
      art: 'karte',
      karteID: '8',
      seite: 'vorne',
      guete: 'tisch',
    }),
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'ruecken',
      guete: 'tisch',
    }),
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'miniatur',
    }),
  ]

  expect(new Set(schluessel).size).toBe(4)
})

// **Und das Räumen wirft alles weg.**
it('räumt alles weg', () => {
  const speicher = new Texturspeicher()

  speicher.belegen(
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    }),
  )
  speicher.belegen(
    texturschluessel({
      art: 'karte',
      karteID: '8',
      seite: 'vorne',
      guete: 'tisch',
    }),
  )

  speicher.raeume()

  expect(speicher.groesse()).toBe(0)
})

// **Eine Antwort aus einer alten Runde trägt sich nicht ein.**
//
// # Der Fall, den der Beleg allein nicht fängt
//
// **Räumen, dieselbe Karte wieder anzeigen, und die Antwort von vorhin
// kommt an.** Der Beleg steht dann wieder da — er ist nur ein anderer. Nach
// einem Profilwechsel wäre das die Karte eines **anderen Profils** unter
// derselben Kennung, und der Speicher trüge sie ein, ohne dass jemand es
// merkt (§7, Regel 2).
it('verwirft eine Antwort aus einer alten Runde', () => {
  const speicher = new Texturspeicher()
  const schluessel = texturschluessel({
    art: 'karte',
    karteID: '7',
    seite: 'vorne',
    guete: 'tisch',
  })

  speicher.belegen(schluessel)

  const alte = speicher.runde()

  speicher.raeume()

  // **Derselbe Schlüssel, neu belegt** — der Beleg steht wieder da.
  speicher.belegen(schluessel)

  speicher.setze(schluessel, eineTextur('/spielbild/alt.png'), alte)

  expect(speicher.hole(schluessel)).toBeUndefined()

  // **Und die Antwort aus der neuen Runde kommt an.**
  speicher.setze(schluessel, eineTextur('/spielbild/neu.png'), speicher.runde())

  expect(speicher.hole(schluessel)?.adresse).toBe('/spielbild/neu.png')
})

// **Ein vorgeladener Schlüssel wird nicht zweimal geholt.**
//
// **Er trägt die Zahl null und trotzdem einen Eintrag**: das Bild ist
// unterwegs oder liegt schon da. Wer ihn danach anmeldet, bekommt `false` —
// also keinen zweiten Ruf.
it('holt einen vorgeladenen Schlüssel nicht noch einmal', () => {
  const speicher = new Texturspeicher()
  const schluessel = texturschluessel({
    art: 'ruecken',
    spiel: 'magic',
    guete: 'tisch-niedrig',
  })

  speicher.merke(schluessel)

  expect(speicher.kennt(schluessel)).toBe(true)
  expect(speicher.zahl(schluessel)).toBe(0)

  expect(speicher.belegen(schluessel)).toBe(false)
  expect(speicher.zahl(schluessel)).toBe(1)

  // **Und zweimal merken legt nichts zweites an.**
  speicher.merke(schluessel)

  expect(speicher.zahl(schluessel)).toBe(1)
})

// **Ein Rücken und eine Karte mit demselben Namen sind zwei Schlüssel.**
//
// **Die Art steht vorn**, sonst hätte ein Spiel `7` denselben Griff wie eine
// Karte `7`.
it('trennt den Rücken von der Karte', () => {
  expect(
    texturschluessel({ art: 'ruecken', spiel: '7', guete: 'tisch' }),
  ).not.toBe(
    texturschluessel({
      art: 'karte',
      karteID: '7',
      seite: 'vorne',
      guete: 'tisch',
    }),
  )
})
