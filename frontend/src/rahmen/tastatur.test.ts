import { describe, expect, it } from 'vitest'

import { seiten } from '../navigation'
import { deuten, type Tastenlage } from './tastatur'

// lage baut eine Tastenlage aus dem, was ein Fall wirklich sagt.
//
// **Mit Vorgaben**, damit ein Fall über ⌘K nicht drei Wahrheitswerte nennen
// muss, die mit ⌘K nichts zu tun haben.
function lage(teile: Partial<Tastenlage> & { taste: string }): Tastenlage {
  return {
    code: '',
    meta: true,
    alt: false,
    umschalt: false,
    imEingabefeld: false,
    ...teile,
  }
}

describe('Tastenkürzel', () => {
  it('öffnet die Palette mit ⌘K', () => {
    expect(deuten(lage({ taste: 'k' }))).toEqual({ art: 'palette' })
    expect(deuten(lage({ taste: 'K' }))).toEqual({ art: 'palette' })
  })

  it('öffnet die Suche mit ⌘F', () => {
    expect(deuten(lage({ taste: 'f' }))).toEqual({ art: 'suche' })
  })

  // **⌃⌘F gehört dem Vollbild** —
  // #1994.
  // Gemessen am 2026-10-01 in der Abnahme-VM: die Seite deutete es als ⌘F
  // und schlug die Suche auf. Eine Kombination mit ⌃ ist nie eines unserer
  // Kürzel; sie gehört macOS — ⌃⌘F dem Vollbild, ⌃⌘Leertaste den Emoji.
  it.each(['f', 'k', '1', ' '])('lässt ⌃⌘%s dem System', (taste) => {
    expect(deuten(lage({ taste, ctrl: true }))).toBeNull()
  })

  // ⌘, ist unter macOS die Einstellungen — jede Anwendung hält sich daran.
  it('führt mit ⌘, zu den Einstellungen', () => {
    expect(deuten(lage({ taste: ',' }))).toEqual({
      art: 'seite',
      seite: 'einstellungen',
    })
  })

  it.each(seiten.filter((seite) => seite.kuerzel !== ''))(
    'führt mit ⌘$kuerzel zu $schluessel',
    (seite) => {
      expect(deuten(lage({ taste: seite.kuerzel }))).toEqual({
        art: 'seite',
        seite: seite.schluessel,
      })
    },
  )

  // **Ohne ⌘ ist eine Taste eine Taste.** Wer in ein Textfeld ein `k` tippt,
  // will kein Fenster öffnen.
  it('deutet ohne die Befehlstaste nichts', () => {
    expect(deuten(lage({ taste: 'k', meta: false }))).toBeNull()
    expect(deuten(lage({ taste: ',', meta: false }))).toBeNull()
  })

  // **Seit den Setangaben ist keine Ziffer mehr frei** (#567): zehn Seiten,
  // **⌘1 bis ⌘7 sind die Seitenleiste** — seit dem 2026-09-03 (Werkstück C
  // von Phase 4.10). ⌘8, ⌘9 und ⌘0 führen auf **nichts**: die Leiste hat
  // sieben Einträge mit Ziffer, und die Laschen des Sets sitzen auf ⌥⌘.
  //
  // **Und seit T9 (#1474) ist keine mehr frei**: ⌘0 führt auf *Handeln*,
  // und die Reihe ⌘1 bis ⌘0 ist vergeben. Was danach kommt, bekommt ein
  // Zeichen wie ⌘' — die Begründung steht bei `reservierteKuerzel`.
  // **⌘8 ist seit Werkstück I von Phase 4.10 belegt**
  // (#1325):
  // die Karten sind als ⌘3 in die Leiste gewandert, und die übrigen sind
  // aufgerückt.
  // **Die Ziffernreihe führt seit UX-02 in die Bereiche**
  // (#1674).
  //
  // # Was hier vorher stand
  //
  // ⌘8 auf *Turniere*, ⌘9 auf *Spielen*, ⌘0 auf *Handeln* und ⌘3 auf
  // *Karten* — jede Seite der Leiste hatte ihre Ziffer, und die Reihe war
  // „voll". Sie ist jetzt **kürzer und geordnet**: sechs Ziffern für Start
  // und fünf Bereiche, und jede führt auf die **erste** Seite ihres Bereichs.
  //
  // **⌘7 bis ⌘0 sind damit wieder frei**, und das ist kein Verlust, sondern
  // Platz: sie standen für vier Seiten, die jetzt unter drei Bereichen
  // stehen.
  it('führt die Ziffernreihe in die Bereiche', () => {
    expect(deuten(lage({ taste: 'q' }))).toBeNull()

    expect(deuten(lage({ taste: '1' }))).toEqual({
      art: 'seite',
      seite: 'start',
    })

    expect(deuten(lage({ taste: '2' }))).toEqual({
      art: 'seite',
      seite: 'sets',
    })

    // **⌘3 hieß *Karten* und heißt jetzt *Sammeln*.** Das ist die Migration,
    // und sie steht in `PROTOTYP-0.6.0.md`.
    expect(deuten(lage({ taste: '3' }))).toEqual({
      art: 'seite',
      seite: 'sammlung',
    })

    expect(deuten(lage({ taste: '4' }))).toEqual({
      art: 'seite',
      seite: 'decks',
    })

    expect(deuten(lage({ taste: '5' }))).toEqual({
      art: 'seite',
      seite: 'handeln',
    })

    expect(deuten(lage({ taste: '6' }))).toEqual({
      art: 'seite',
      seite: 'pakete',
    })
  })

  // **Und ⌘7 bis ⌘0 sind unbelegt** — die Gegenprobe zur Zeile darüber.
  // Ohne sie stünde nur da, was belegt ist, und nicht, dass der Rest frei
  // blieb.
  it('lässt ⌘7 bis ⌘0 unbelegt', () => {
    for (const taste of ['7', '8', '9', '0']) {
      expect(deuten(lage({ taste })), `⌘${taste}`).toBeNull()
    }
  })

  // ⌘Z und ⇧⌘Z sind die gewohnte Form (#52).
  it('nimmt mit ⌘Z zurück und stellt mit ⇧⌘Z wieder her', () => {
    expect(deuten(lage({ taste: 'z' }))).toEqual({ art: 'zurueck' })
    expect(deuten(lage({ taste: 'z', umschalt: true }))).toEqual({ art: 'vor' })
    expect(deuten(lage({ taste: 'Z', umschalt: true }))).toEqual({ art: 'vor' })
  })

  // **Im Textfeld gehört ⌘Z dem Textfeld.**
  //
  // Geschrieben wird beim Verlassen des Feldes, nicht bei jedem Tastendruck:
  // was jemand gerade tippt, steht noch in keinem Verlauf. Wer ⌘Z dort abfinge,
  // nähme eine **frühere** Änderung zurück, während der Cursor in einer
  // ungesicherten steht — und die Rücknahme träfe etwas anderes als das, was
  // die tippende Person vor Augen hat.
  it('überlässt ⌘Z im Eingabefeld dem Feld', () => {
    expect(deuten(lage({ taste: 'z', imEingabefeld: true }))).toBeNull()
    expect(
      deuten(lage({ taste: 'z', umschalt: true, imEingabefeld: true })),
    ).toBeNull()
  })

  // **Die übrigen Kürzel gelten auch im Feld.** ⌘K und ⌘1 haben dort keine
  // eigene Bedeutung, und wer zum Wechseln erst klicken müsste, hätte keine
  // durchgängige Tastaturbedienung (§6).
  it('lässt die übrigen Kürzel im Eingabefeld gelten', () => {
    expect(deuten(lage({ taste: 'k', imEingabefeld: true }))).toEqual({
      art: 'palette',
    })
    expect(deuten(lage({ taste: '2', imEingabefeld: true }))).toEqual({
      art: 'seite',
      seite: 'sets',
    })
  })

  // **⇧ ändert an den übrigen nichts.** ⇧⌘K ist ⌘K; eine Kombination, die je
  // nach Umschalttaste etwas anderes tut, ist keine Gewohnheit mehr.
  it('lässt ⇧ die übrigen Kürzel unberührt', () => {
    expect(deuten(lage({ taste: 'k', umschalt: true }))).toEqual({
      art: 'palette',
    })
  })
})

// **⌘S sichert, ⇧⌘S sichert unter** (#162).
//
// `AGENTS.md` §6 verlangt, dass macOS-Eigenheiten bedient werden, und ⌘S ist
// die erste, die jemand ausprobiert.
describe('⌘S', () => {
  it('deutet ⌘S als sichern', () => {
    expect(
      deuten({
        taste: 's',
        code: '',
        meta: true,
        alt: false,
        umschalt: false,
        imEingabefeld: false,
      }),
    ).toEqual({ art: 'sichern' })
  })

  it('deutet ⇧⌘S als sichern unter', () => {
    expect(
      deuten({
        taste: 's',
        code: '',
        meta: true,
        alt: false,
        umschalt: true,
        imEingabefeld: false,
      }),
    ).toEqual({ art: 'sichernUnter' })
  })

  // **Auch im Eingabefeld**, anders als ⌘Z.
  //
  // Der Grund für die Ausnahme bei ⌘Z ist, dass das Textfeld selbst eine
  // Rücknahme kennt und die gemeinte näher liegt. ⌘S kennt es nicht: dort
  // sichert niemand etwas, und wer im Kartentext steht und ⌘S drückt, meint das
  // Set.
  it('gilt auch im Eingabefeld', () => {
    expect(
      deuten({
        taste: 's',
        code: '',
        meta: true,
        alt: false,
        umschalt: false,
        imEingabefeld: true,
      }),
    ).toEqual({ art: 'sichern' })
  })
})

// **⌥⌘S klappt die Seitenleiste** — Werkstück U14 von Phase 5.1 (#1490), das
// Kürzel des Finders und von Mail. ⌘S und ⇧⌘S bleiben das Sichern.
describe('⌥⌘S', () => {
  it('schaltet die Seitenleiste und lässt das Sichern in Ruhe', () => {
    expect(deuten(lage({ taste: 's', alt: true, code: 'KeyS' }))).toEqual({
      art: 'seitenleiste',
    })
    expect(deuten(lage({ taste: 's' }))).toEqual({ art: 'sichern' })
  })
})

// **⌘[ und ⌘] gehen den Seitengang** — UX-03.AK-02 (#1675), mit den Kürzeln,
// die der Finder für den vorigen und den nächsten Ordner führt.
describe('⌘[ und ⌘]', () => {
  it('geht zurück und vor — über die Taste, nicht das Zeichen', () => {
    // **Auf einer deutschen Tastatur steht dort „ü"**: gedeutet wird wie bei
    // ⌥⌘[ über `code`, sonst wäre es nur mit ⌥5 erreichbar — und das ist
    // ⌥⌘5, eine Seite.
    expect(deuten(lage({ taste: 'ü', code: 'BracketLeft' }))).toEqual({
      art: 'seiteZurueck',
    })
    expect(deuten(lage({ taste: '+', code: 'BracketRight' }))).toEqual({
      art: 'seiteVor',
    })
  })

  it('gilt auch im Eingabefeld, wie ⌘S', () => {
    expect(
      deuten(lage({ taste: '[', code: 'BracketLeft', imEingabefeld: true })),
    ).toEqual({ art: 'seiteZurueck' })
  })

  it('lässt ⌥⌘[ den Regionen', () => {
    expect(
      deuten(lage({ taste: '“', code: 'BracketLeft', alt: true })),
    ).toEqual({ art: 'navigator' })
  })
})
