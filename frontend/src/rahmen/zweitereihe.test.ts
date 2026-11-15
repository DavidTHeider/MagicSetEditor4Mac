// **Die zweite Kürzelreihe** —
// #1291,
// Werkstück E.
//
// # Der Befund
//
// Gemessen am 2026-08-30: von zwanzig Seiten haben **zwölf** ein Kürzel und
// **acht** keines — darunter *Vorlagen* und *Sammlung*, die beide ausdrücklich
// genannt wurden. Die Ziffernreihe ⌘1…⌘0 ist voll.
//
// **Erreichbar sind sie**: ⌘K listet alle zwanzig. Was fehlt, ist der eine
// Anschlag.
//
// # Warum ⌥⌘ und nicht ⇧⌘
//
// **Weil macOS ⇧⌘3, ⇧⌘4 und ⇧⌘5 selbst nimmt** — das sind die
// Bildschirmfotos, und sie erreichen keine Anwendung. Eine Reihe mit drei
// Löchern ist keine Reihe.
//
// # Und warum `code` statt `key`
//
// **Weil ⌥ die Taste umschreibt.** Auf einer Mac-Tastatur gibt ⌥1 nicht `"1"`,
// sondern `"¡"`; ⌥2 gibt `"™"`. Wer auf `event.key` prüft, findet die Ziffer
// nie. `event.code` bleibt `"Digit1"`.
import { describe, expect, it } from 'vitest'

import { seiten } from '../navigation'
import { deuten, zweiteReiheAus } from './tastatur'

describe('die zweite Kürzelreihe', () => {
  // **Der Satz lautet seit dem 2026-09-03 anders.**
  //
  // Hier stand „jede Seite ohne Ziffer hat ein ⌥⌘-Kürzel", und das ging auf,
  // solange zehn Plätze für zehn Seiten reichten. Seit die **ganze** erste
  // Reihe der Seitenleiste gehört (Werkstück C von Phase 4.10), stehen
  // dreizehn Laschen zehn Ziffern gegenüber — **drei tragen keine**.
  //
  // Die Zusage ist deshalb: **wer keine trägt, steht hier namentlich.** Eine
  // vierte Lasche ohne Taste ist damit eine Entscheidung und kein Versehen.
  // **Zwei seit Werkstück I von Phase 4.10**
  // (#1325):
  // die Karten sind in die Seitenleiste gewandert und haben ⌥⌘1 frei gemacht;
  // die Schlüsselwörter haben ihre Taste damit zurück. Übrig bleiben die zwei
  // Werkzeuge, an denen niemand im Minutentakt arbeitet.
  // **Sechs Seiten tragen keine Taste** — seit UX-02
  // (#1674).
  //
  // # Was hier vorher stand
  //
  // *„Lässt genau zwei Laschen ohne Taste"* — `pack` und `konsole`, und die
  // Begründung stand bei ihnen: niemand schlägt sie im Minutentakt auf.
  //
  // **Seit die Ziffer dem Bereich gehört, sind es sechs.** *Karten*,
  // *Vorlagen*, *Turniere* und *Spielen* stehen unter *Gestalten*,
  // *Ressourcen* und *Spielen*; die Ziffer führt in den Bereich, also auf
  // dessen **erste** Seite. Die ⌥⌘-Reihe können sie nicht bekommen — die
  // gehört den Laschen des offenen Sets.
  //
  // # Was das kostet, und es gehört ausgesprochen
  //
  // **Vier Seiten verlieren ihr Kürzel.** Erreichbar bleiben sie über die
  // Seitenleiste, über ⌘K und über die Menüleiste — aber nicht mehr mit einer
  // Taste. **UX-02.AK-03 verlangt dafür eine „konfigurierbare Belegung", und
  // die gibt es noch nicht**; solange sie fehlt, ist dieser Punkt offen und
  // nicht erfüllt.
  it('lässt sechs Seiten ohne Taste', () => {
    const ohneJede = seiten
      .filter((seite) => seite.kuerzel === '' && seite.zweitkuerzel === '')
      .map((seite) => seite.schluessel)

    expect(ohneJede).toEqual([
      'vorlagen',
      'turniere',
      'spielen',
      'karten',
      'pack',
      'konsole',
    ])
  })

  it('gibt jeder übrigen Seite ohne Ziffer ein ⌥⌘-Kürzel', () => {
    const ohneTaste = [
      'vorlagen',
      'turniere',
      'spielen',
      'karten',
      'pack',
      'konsole',
    ]

    const ohneZiffer = seiten.filter(
      (seite) => seite.kuerzel === '' && !ohneTaste.includes(seite.schluessel),
    )

    expect(ohneZiffer.length).toBe(10)

    for (const seite of ohneZiffer) {
      expect(typeof seite.zweitkuerzel).toBe('string')
      expect(seite.zweitkuerzel).not.toHaveLength(0)
    }
  })

  it('deutet ⌥⌘ über den Code und nicht über das Zeichen', () => {
    const erste = seiten.find((seite) => seite.zweitkuerzel === '1')

    expect(erste).toBeDefined()

    // **`¡` ist, was ⌥1 auf einer Mac-Tastatur wirklich schickt.**
    const treffer = deuten({
      taste: '¡',
      code: 'Digit1',
      meta: true,
      alt: true,
      umschalt: false,
      imEingabefeld: false,
    })

    expect(treffer).toEqual({ art: 'seite', seite: erste?.schluessel })
  })

  // **Auf AZERTY wirkt dieselbe Taste — und sie ist anders beschriftet.**
  //
  // `event.code` benennt die **Position**, nicht den Aufdruck: an `Digit1`
  // liegt dort kein `1`, sondern `&`. Die Navigation greift trotzdem, und
  // genau das ist der Befund
  // (#1345):
  // die Tafel schrieb „⌥⌘1", ohne zu sagen, welche Taste das ist.
  //
  // **Das Verhalten bleibt, die Beschriftung sagt es jetzt aus.** Eine
  // Beschriftung je Belegung ginge nur über `navigator.keyboard.getLayoutMap()`
  // — Chromium hat sie, das WKWebView dieses Fensters nicht.
  it('greift auch, wenn auf der Taste kein Zeichen einer Ziffer steht', () => {
    const erste = seiten.find((seite) => seite.zweitkuerzel === '1')

    expect(
      deuten({
        taste: '&',
        code: 'Digit1',
        meta: true,
        alt: true,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toEqual({ art: 'seite', seite: erste?.schluessel })
  })

  it('lässt ⌥⌘ ohne bekannte Ziffer in Ruhe', () => {
    expect(
      deuten({
        taste: '≠',
        code: 'Equal',
        meta: true,
        alt: true,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toBeNull()
  })

  // **Eine Ziffer, die keiner Seite gehört** — ⌥⌘9 etwa: die zweite Reihe
  // reicht bis 8. Sie darf nicht auf die Seite mit ⌘9 führen; das wäre eine
  // zweite Bedeutung für dieselbe Seite.
  // **Und seit dem 2026-09-01 ist es keine Ziffer mehr.** Erst wanderte der
  // Fall von `Digit9` auf `Digit0` (#1293), dann nahm die Turnierseite auch
  // die Null (#1294) — die Reihe ist voll, und eine Ziffer ohne Zweitseite
  // gibt es nicht mehr.
  //
  // **Was der Lauf wirklich zusagt**, ist allgemeiner als sein erster Fall:
  // eine Tastenkombination, die **keine Seite** beansprucht, bleibt
  // wirkungslos. Ein Buchstabe ist so eine und bleibt es — die zweite Reihe
  // führt Ziffern. Eine dritte Wanderung von Fall zu Fall wäre eine
  // Behauptung, die alle paar Monate knapp noch stimmt.
  it('lässt eine Kombination ohne Seite in Ruhe', () => {
    expect(
      deuten({
        taste: 'œ',
        code: 'KeyQ',
        meta: true,
        alt: true,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toBeNull()
  })

  it('deutet eine Ziffer ohne ⌥ weiter über das Zeichen', () => {
    const mitZiffer = seiten.find((seite) => seite.kuerzel === '2')

    expect(
      deuten({
        taste: '2',
        code: 'Digit2',
        meta: true,
        alt: false,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toEqual({ art: 'seite', seite: mitZiffer?.schluessel })
  })

  it('nimmt ⌥ ohne ⌘ nicht für sich', () => {
    expect(
      deuten({
        taste: '¡',
        code: 'Digit1',
        meta: false,
        alt: true,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toBeNull()
  })

  it('vergibt jedes Zweitkürzel nur einmal', () => {
    const vergeben = seiten
      .map((seite) => seite.zweitkuerzel)
      .filter((kuerzel) => kuerzel !== '')

    expect(new Set(vergeben).size).toBe(vergeben.length)
  })
})

// **Eine Ziffer, für die es keine Seite gibt, bleibt wirkungslos.**
//
// # Warum über eine eigene Liste und nicht über eine Taste
//
// Weil die zweite Reihe seit dem 2026-09-01 **voll** ist: zehn Seiten ohne
// Erstkürzel, zehn Ziffern. Über eine echte Taste lässt sich der Fall nicht
// mehr herstellen — und ein Zweig, den kein Test betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3).
//
// **Der Ernstfall ist hier eine Konfiguration, keine Eingabe**: genau so war
// `navigation.ts`, bevor die Turnierseite die Null bekam, und genau so wäre
// sie wieder, wenn jemand eine Seite entfernt.
describe('eine Ziffer ohne Seite', () => {
  it('bleibt wirkungslos', () => {
    expect(zweiteReiheAus('Digit7', [])).toBeNull()
  })

  it('führt auf ihre Seite, sobald es eine gibt', () => {
    expect(
      zweiteReiheAus('Digit7', [{ zweitkuerzel: '7', schluessel: 'turniere' }]),
    ).toEqual({ art: 'seite', seite: 'turniere' })
  })

  // **Und was keine Ziffer ist, ebenso.** Ohne diesen Fall bliebe der erste
  // Zweig der Funktion unbetreten.
  it('lässt eine Taste ohne Ziffer in Ruhe', () => {
    expect(zweiteReiheAus('KeyQ', [])).toBeNull()
  })
})
