// **Neben jedem Feld steht höchstens ein kurzer Satz** — UX-15.AK-03
// (#1771).
//
// # Warum eine Zahl und kein Gefühl
//
// Das Kriterium sagt es wörtlich: „nicht nach Gefühl, sondern an einer
// Grenze, die im Lauf steht". **140 Zeichen** — gemessen am 2026-09-22 über
// `de/einstellungshinweise.ts`: der längste erste Satz hat **114**, der
// Median **25**. Die Grenze lässt Luft und greift trotzdem: die ganzen
// Hinweise reichen bis **371**.
//
// # Was er nicht hält
//
// **Ob der Satz gut ist.** Er misst Zeichen; ob der erste Satz der ist, der
// beim Entscheiden hilft, sieht nur ein Mensch. Und ob etwas hinter dem
// Fragezeichen **versteckt** wurde, das man wissen muss, steht im Kopf von
// `Feldhilfe` — auch das ist eine Frage ans Review (§10).
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'
import { kurzform, ohneAuszeichnung } from './kurzform'

// Grenze ist, wie lang der sichtbare Satz höchstens ist.
const Grenze = 140

// erklaerungen sammelt jede Kopferklärung einer Seite.
//
// **Über den Katalog und nicht über eine Aufzählung**: eine Liste der Seiten
// liefe bei der nächsten auseinander, und der Lauf prüfte still zu wenig
// (§10).
function erklaerungen(seiten: Record<string, unknown>): [string, string][] {
  const gefunden: [string, string][] = []

  for (const [name, seite] of Object.entries(seiten)) {
    if (typeof seite !== 'object' || seite === null) continue

    const text = (seite as Record<string, unknown>).erklaerung

    if (typeof text === 'string') gefunden.push([name, text])
  }

  return gefunden
}

describe.each(['de', 'en'] as const)('Die Kopferklärungen in %s', (sprache) => {
  const koepfe = erklaerungen(kataloge[sprache].seiten)

  // **Sonst wäre dieser Lauf grün, weil er nichts sieht** (§10).
  it('findet überhaupt welche', () => {
    expect(koepfe.length).toBeGreaterThan(15)
  })

  // **Derselbe Schnitt wie unter einem Feld** — UX-15.AK-01
  // (#1771).
  // Gemessen am 2026-09-22 reichten sie bis **340** Zeichen; ihre ersten
  // Sätze bleiben darunter, und `Seitenkopf` zeigt nur den.
  it('hält jeden sichtbaren Satz unter der Grenze', () => {
    const zuLang = koepfe
      .map(([name, text]) => ({
        name,
        laenge: kurzform(ohneAuszeichnung(text)).length,
      }))
      .filter((eines) => eines.laenge > Grenze)

    expect(zuLang).toEqual([])
  })
})

describe.each(['de', 'en'] as const)('Die Hinweise in %s', (sprache) => {
  const hinweise: Record<string, string> =
    kataloge[sprache].seiten.einstellungen.hinweise

  it('führt überhaupt welche', () => {
    // **Sonst wäre dieser Lauf grün, weil er nichts sieht** (§10): ein
    // leerer Katalog erfüllt jede Grenze.
    expect(Object.keys(hinweise).length).toBeGreaterThan(20)
  })

  it('hält jeden sichtbaren Satz unter der Grenze', () => {
    const zuLang = Object.entries(hinweise)
      .map(([schluessel, text]) => ({
        schluessel,
        laenge: kurzform(ohneAuszeichnung(text)).length,
      }))
      .filter((eines) => eines.laenge > Grenze)

    expect(zuLang).toEqual([])
  })

  // **Und keiner ist leer.** Ein Hinweis, dessen erster Satz nichts enthält,
  // stünde als leere Zeile da und schöbe das nächste Feld nach unten.
  it('lässt keinen sichtbaren Satz leer', () => {
    const leere = Object.entries(hinweise)
      .filter(([, text]) => kurzform(ohneAuszeichnung(text)).trim() === '')
      .map(([schluessel]) => schluessel)

    expect(leere).toEqual([])
  })
})

// **Jeder Text im Katalog, nicht nur zwei Sorten** — Sichtprüfung zu UX-15
// am 2026-09-24 (#1771).
//
// Die beiden Läufe darüber hielten Kopferklärungen und Hinweise unter einer
// Einstellung. Eine **Abschnittserläuterung** ging durch `Erklaerungsabsatz`
// und stand unter keinem Lauf: auf *Nutzer und Profile* zeigte einer seinen
// ersten Satz mit 167 Zeichen, gemessen in WebKit. Ein Wächter, der die halbe
// Zusage hält, macht die andere unsichtbar (§10). **Dieser misst jeden Text**
// — gleich, durch welches Bauteil er geht, und auch einen, der ganz dasteht:
// dessen erster Satz ist der, den man zuerst liest.
//
// **Ausgenommen ist die Hilfe**: dort ist der lange Text der Inhalt, und
// ihn hinter ein Hilfesymbol zu legen wäre die Umkehrung ihres Zwecks
// (`docs/FLAECHENTEXTE.md`).
//
// **Was er nicht hält**: ob ein langer Text, dessen erster Satz kurz ist,
// ganz dasteht, statt hinter dem Symbol zu stehen. Das sieht nur das
// laufende Fenster — die Probe `ux15-seitenlaengen.probe` zählt es je Seite.
function alleTexte(wert: unknown, pfad: string): [string, string][] {
  if (typeof wert === 'string') return [[pfad, wert]]
  if (typeof wert !== 'object' || wert === null) return []

  return Object.entries(wert).flatMap(([schluessel, kind]) =>
    alleTexte(kind, `${pfad}.${schluessel}`),
  )
}

// Ausnahmen sind Texte, die kein Satz sind — je mit ihrem Grund. **Eine
// Ausnahme, die nicht mehr nötig ist, macht den Lauf rot**: sonst bliebe sie
// stehen und ließe den nächsten langen Text unter ihrem Namen durch.
const Ausnahmen: Record<string, string> = {
  'seiten.nutzer.tokenOnboardingBeispiel':
    'ein Stück Konfiguration zum Abschreiben, gesetzt als Quelltext — kein Satz',
}

describe.each(['de', 'en'] as const)(
  'Jeder Text außerhalb der Hilfe in %s',
  (sprache) => {
    const ohneHilfe = Object.fromEntries(
      Object.entries(kataloge[sprache]).filter(
        ([katalog]) => katalog !== 'hilfe',
      ),
    )
    const texte = alleTexte(ohneHilfe, sprache).map(
      ([pfad, text]): [string, string] => [
        pfad.slice(sprache.length + 1),
        text,
      ],
    )

    it('findet überhaupt welche', () => {
      expect(texte.length).toBeGreaterThan(1000)
    })

    const laengen = texte.map(([pfad, text]) => ({
      pfad,
      laenge: kurzform(ohneAuszeichnung(text)).length,
    }))

    it('hält jeden ersten Satz unter der Grenze', () => {
      const zuLang = laengen.filter(
        (eines) => eines.laenge > Grenze && !(eines.pfad in Ausnahmen),
      )

      expect(zuLang).toEqual([])
    })

    it('führt keine Ausnahme, die nicht mehr nötig ist', () => {
      const unnoetig = Object.keys(Ausnahmen).filter(
        (pfad) =>
          !laengen.some(
            (eines) => eines.pfad === pfad && eines.laenge > Grenze,
          ),
      )

      expect(unnoetig).toEqual([])
    })
  },
)
