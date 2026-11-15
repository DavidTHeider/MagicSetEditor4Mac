// **Die Zeile über den Spalten** — UXR-02
// (#1910).
//
// # Der Befund
//
// Jede Arbeitsplatzseite — Karten, Vorlagen, Sammlung, Decks, Pakete — trug
// über den Spalten eine eigene Zeile mit zwei Knöpfen in voller Größe
// („Navigator ausblenden", „Inspector ausblenden"), gerechnet aus den Klassen
// rund 40 Punkte hoch. **Im Mindestfenster kam ein Satz dazu**: „Zu schmal
// für alle drei Spalten — der Navigator geht über dem Zentrum auf", 73
// Zeichen neben zwei Knöpfen in einer Breite von rund 680 Punkten — die
// Zeile brach in zwei um und nahm rund 88 Punkte, auf der Kartenseite einem
// Zentrum von gerechnet 265 Punkten Höhe.
//
// **Das Artboard zeichnet die Zeile nicht**: jede Spalte trägt einen kleinen
// Kopf mit ihrem Kürzel (`region-kopf`, *Karten*), und im Artboard *Enge*
// „wartet der Inspector an der Kante". Deshalb hier: klein, jeder Knopf über
// seiner Spalte, und der Satz beschreibt den Knopf — für die Vorlesehilfe als
// Beschreibung, für den Zeiger als Titel, wie am Knopf der Seitenleiste.
//
// **Kein `Tooltip`**, und das ist gemessen: er verbraucht Escape, solange er
// steht — und wer die Spalte über dem Zentrum mit der Maus geöffnet hat, dem
// stand er unter dem ruhenden Zeiger. Das erste Escape schloss dann den
// Hinweis und nicht die Spalte (`arbeitsplatzenge.test.tsx` wurde rot).
//
// **Was jsdom nicht sieht**, ist die Höhe: es rechnet kein Layout. Geprüft
// wird hier die Gestalt; die Punkte misst die Hüllenprobe `uxr02`.
import { render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { regionskuerzel } from './Regionsknopf'
import { deuten } from '../rahmen/tastatur'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

const texte = kataloge.de.bausteine.arbeitsplatz

let raeumeBreite: () => void = () => undefined

afterEach(() => {
  raeumeBreite()
})

function Probe({ beiEnge }: { beiEnge?: 'navigator' | 'inspector' }) {
  return (
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{ bezeichnung: 'Karten', inhalt: <p>Die Liste</p> }}
        zentrum={<p>Die Vorschau</p>}
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
        beiEnge={beiEnge}
      />
    </Sprachraum>
  )
}

describe('die Zeile über den Spalten, mit Platz', () => {
  beforeEach(() => {
    raeumeBreite = gibJsdomEineBreite(1400)
  })

  // **Klein wie der Kopf einer Spalte** — dieselbe Größe, die `Knopf` für
  // eine Reihe gleichrangiger Handlungen hat (`klein`).
  it('zeichnet beide Regionsknöpfe klein', () => {
    render(<Probe />)

    for (const name of [texte.navigatorZu, texte.inspectorZu]) {
      expect(screen.getByRole('button', { name })).toHaveClass('text-sm')
    }
  })

  // **Jeder Knopf steht über seiner Spalte**: der des Navigators links, der
  // des Inspectors rechts — wie die Spaltenköpfe im Artboard.
  it('stellt den Knopf des Inspectors an den rechten Rand', () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: texte.inspectorZu })

    expect(knopf.closest('.ml-auto')).not.toBeNull()
    expect(
      screen
        .getByRole('button', { name: texte.navigatorZu })
        .closest('.ml-auto'),
    ).toBeNull()
  })

  // **Das Kürzel steht am Knopf** — das Artboard nennt es im Kopf jeder
  // Spalte (`⌥⌘[`, `⌥⌘]`). Hier sagt es der Knopf der Vorlesehilfe
  // (`aria-keyshortcuts`) und dem Zeiger im Titel, wie der Knopf der
  // Seitenleiste „(⌥⌘S)".
  it('nennt sein Kürzel', () => {
    render(<Probe />)

    const navigator = screen.getByRole('button', { name: texte.navigatorZu })
    const inspector = screen.getByRole('button', { name: texte.inspectorZu })

    expect(navigator).toHaveAttribute('aria-keyshortcuts', 'Alt+Meta+[')
    expect(inspector).toHaveAttribute('aria-keyshortcuts', 'Alt+Meta+]')
    expect(navigator).toHaveAttribute('title', `${texte.navigatorZu} (⌥⌘[)`)
    expect(inspector).toHaveAttribute('title', `${texte.inspectorZu} (⌥⌘])`)

    // **Ohne Enge gibt es nichts zu erklären.**
    expect(navigator).not.toHaveAttribute('aria-describedby')
  })

  // **Keine Kopie der Tasten**: was der Knopf ansagt, ist, was die Tastatur
  // deutet — sonst nennte er ein Kürzel, das nichts tut.
  it('nennt die Tasten, die die Tastatur auch deutet', () => {
    for (const [region, { tasten, zeichen }] of Object.entries(
      regionskuerzel,
    )) {
      const taste = tasten.slice(-1)

      expect(tasten).toBe(`Alt+Meta+${taste}`)
      expect(zeichen).toBe(`⌥⌘${taste}`)
      expect(
        deuten({
          taste,
          code: taste === '[' ? 'BracketLeft' : 'BracketRight',
          meta: true,
          ctrl: false,
          alt: true,
          umschalt: false,
          imEingabefeld: false,
        }),
      ).toEqual({ art: region })
    }
  })
})

describe('die Zeile über den Spalten, bei Enge', () => {
  beforeEach(() => {
    raeumeBreite = gibJsdomEineBreite(1000)
  })

  // **Der Satz steht nicht mehr in der Zeile** — er brach sie im
  // Mindestfenster in zwei. Er beschreibt den Knopf.
  it('trägt den Satz als Beschreibung am Knopf statt in der Zeile', async () => {
    render(<Probe />)
    await effekteNachholen()

    const knopf = screen.getByRole('button', { name: texte.navigatorAuf })

    expect(knopf).toHaveAccessibleDescription(texte.navigatorEng)
    expect(knopf).toHaveAttribute('title', `${texte.navigatorEng} (⌥⌘[)`)
    expect(screen.getByText(texte.navigatorEng)).not.toBeVisible()
  })

  it('sagt es am Inspector, wenn er weicht', async () => {
    render(<Probe beiEnge="inspector" />)
    await effekteNachholen()

    expect(
      screen.getByRole('button', { name: texte.inspectorAuf }),
    ).toHaveAccessibleDescription(texte.inspectorEng)
    expect(
      screen.getByRole('button', { name: texte.navigatorZu }),
    ).not.toHaveAttribute('aria-describedby')
  })
})
