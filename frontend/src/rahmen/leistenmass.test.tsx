// **Die Seitenleiste passt ins Mindestfenster** — UXR-02
// (#1910).
//
// # Der Befund, gerechnet aus den Klassen
//
// Die Leiste steht `sticky` und so hoch wie das Fenster; was länger ist,
// rollt in ihrer Navigation. Bis UXR-02 war sie, Zeile für Zeile aus den
// Klassen gerechnet: Name 53 (`py-4`, `text-sm`, Linie), Kontext 125 (Bild
// und drei Zeilen, „Profil wechseln …"), zwei Zonen 481 (zwei Überschriften
// zu 35, neun Einträge zu **40** — `py-2` um eine Zeile von 24 —, Trenner,
// Ränder) und der Fuß mit ⌘K 65 — zusammen **724 Punkte**. Das Mindestfenster
// ist 700 hoch, samt Titelleiste. **Die Leiste rollte also im engen Fenster**,
// und *Hilfe*, der letzte Eintrag, stand unter ihrem Rand.
//
// Das Artboard zeichnet einen Eintrag **32** hoch (`.nav-eintrag`: 7 Punkte
// Rand um eine Zeile von 13) und den Namen ohne eigene Kante. Hier: `py-1.5`
// um `text-sm` gibt 32, der Name `py-3` — zusammen **624** Punkte, auch mit
// der Zeile „Profil von …" unter 700.
//
// **Was jsdom nicht sieht**, ist die Höhe selbst — es rechnet kein Layout.
// Geprüft wird hier die Gestalt; ob *Hilfe* im Bild steht, misst die
// Hüllenprobe `uxr02` am nativen Fenster.
//
// **Und dort war `py-1.5` nicht 32, sondern 37 Punkte** (`uxr02`, 2026-10-02):
// die Zeile ist im WebKit des Fensters höher, als `text-sm` verspricht —
// vermutlich wegen der Sinnbilder aus einer anderen Schrift; gemessen ist nur
// die Höhe. Im engen Fenster rollte die Leiste damit weiter. **Deshalb eine feste Höhe** (`h-8`)
// statt eines Rands um eine Zeile, deren Höhe die Schrift bestimmt.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { leisteneintraege } from '../leistenordnung'
import { leisteneintrag } from '../test/leistenweg'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de

function zeige(waagerecht = false) {
  belegeApi({})

  render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="start"
        wechseln={vi.fn()}
        seitenleiste={{ offen: true, umschalten: vi.fn(), waagerecht }}
      >
        <p>{texte.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

describe('die Seitenleiste im Mindestfenster', () => {
  it('zeichnet jeden Eintrag 32 Punkte hoch, wie das Artboard', async () => {
    zeige()

    for (const eintrag of leisteneintraege) {
      const knopf = await leisteneintrag(eintrag.ziel)

      expect(knopf).toHaveClass('h-8', 'text-sm')
      expect(knopf).not.toHaveClass('py-2')
      expect(knopf).not.toHaveClass('py-1.5')
    }
  })

  it('hält den Fuß mit der Palette in derselben Höhe', () => {
    zeige()

    expect(
      screen.getByRole('button', { name: texte.allgemein.palette.oeffnen }),
    ).toHaveClass('h-8', 'text-sm')
  })

  it('gibt dem Namen der Anwendung einen schmaleren Rand', () => {
    zeige()

    const name = screen.getByText(texte.allgemein.anwendungsname, {
      selector: '#seitenleiste > p',
    })

    expect(name).toHaveClass('py-3')
    expect(name).not.toHaveClass('py-4')
  })

  // **Die Leiste oben bleibt, wie sie ist** (UX-14.AK-02): sie rollt nicht,
  // sie bricht um, und ihre Höhe ist gemessen (49 Punkte bei 1.440, 87 bei
  // 1.024). Der Befund gilt der senkrechten.
  it('lässt die Leiste oben unverändert', async () => {
    zeige(true)

    const knopf = await leisteneintrag('start')

    expect(knopf).toHaveClass('py-2')
  })
})
