// **Bei schmalem Fenster gibt die Seitenleiste nach** —
// #1630.
//
// # Was diese Läufe halten
//
// **Dass die Leiste unterhalb der Grenze nicht im Fluss steht.** Sie war mit
// `w-60 shrink-0` gesetzt und gab nie nach; bei 370 Bildpunkten Fensterbreite
// blieben dem Inhalt **89** — jede Zeile brach nach einem Wort.
//
// **Und dass ein schmales Fenster die Einstellung nicht überschreibt.** Ob
// die Leiste stehen soll, ist eine Entscheidung des Menschen und liegt im
// Bestand (`seitenleiste`, U14/#1490). Eine Fensterbreite darf sie
// verdecken, nicht ändern.
//
// # Was sie nicht halten
//
// **Die Zahl.** jsdom rechnet kein Layout; `getBoundingClientRect` gibt hier
// Nullen. Dass dem Inhalt bei 370 Bildpunkten wirklich die volle Breite
// bleibt, ist eine **Sichtprüfung** — gemessen im laufenden Fenster und im
// Issue festgehalten (§3).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'

const navigation = kataloge.de.allgemein.navigation

// **Die Frage wird gelesen und nicht geraten**: der Haken fragt nach einer
// Höchstbreite, und diese Antwort vergleicht sie mit der gesetzten Breite.
// Ein Stummel, der immer `matches: true` gäbe, prüfte den Haken nicht.
function setzeFensterbreite(breite: number): void {
  window.matchMedia = ((frage: string) => {
    const grenze = Number(/max-width:\s*(\d+)px/.exec(frage)?.[1] ?? '0')

    return {
      matches: breite <= grenze,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
    }
  }) as unknown as typeof window.matchMedia
}

function zeige(breite: number, umschalten = vi.fn(), wechseln = vi.fn()) {
  setzeFensterbreite(breite)
  belegeApi({})

  const ansicht = render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="start"
        wechseln={wechseln}
        seitenleiste={{ offen: true, umschalten }}
      >
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  return { ...ansicht, umschalten, wechseln }
}

afterEach(() => {
  raeumeApi()
  Reflect.deleteProperty(window, 'matchMedia')
})

describe('Ein schmales Fenster', () => {
  it('lässt die Leiste eingeklappt, obwohl die Einstellung sie will', () => {
    zeige(370)

    expect(
      screen.queryByRole('navigation', { name: navigation.bezeichnung }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: navigation.seitenleisteEinblenden }),
    ).toHaveAttribute('aria-expanded', 'false')
  })

  it('öffnet sie als Schublade über dem Inhalt — ohne die Einstellung zu schreiben', async () => {
    const { umschalten } = zeige(370)

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteEinblenden }),
    )

    // **Der Beleg, dass der Knopf trägt**, steht vor der Verneinung (§3):
    // die Leiste ist da, und sie steht außerhalb des Flusses.
    expect(
      screen.getByRole('navigation', { name: navigation.bezeichnung }),
    ).toBeInTheDocument()
    // **`fixed` und nicht `absolute`** — seit UX-02
    // (#1674):
    // `absolute` misst gegen den Rahmen, und der ist so hoch wie der Inhalt.
    // Auf der Startseite waren das 24.875 Bildpunkte.
    expect(document.getElementById('seitenleiste')).toHaveClass('fixed')

    expect(umschalten).not.toHaveBeenCalled()
  })

  it('schließt sie wieder, wenn derselbe Knopf noch einmal kommt', async () => {
    zeige(370)

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteEinblenden }),
    )
    expect(
      screen.getByRole('navigation', { name: navigation.bezeichnung }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteAusblenden }),
    )

    expect(
      screen.queryByRole('navigation', { name: navigation.bezeichnung }),
    ).not.toBeInTheDocument()
  })

  it('schließt die Schublade, sobald eine Fläche aufgeschlagen ist', async () => {
    const { rerender } = zeige(370)

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteEinblenden }),
    )
    expect(
      screen.getByRole('navigation', { name: navigation.bezeichnung }),
    ).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Rahmen
          aktuell="karten"
          wechseln={vi.fn()}
          seitenleiste={{ offen: true, umschalten: vi.fn() }}
        >
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    expect(
      screen.queryByRole('navigation', { name: navigation.bezeichnung }),
    ).not.toBeInTheDocument()
  })
})

describe('Ein breites Fenster', () => {
  it('lässt die Leiste im Fluss stehen und schreibt die Einstellung', async () => {
    const { umschalten } = zeige(1200)

    expect(
      screen.getByRole('navigation', { name: navigation.bezeichnung }),
    ).toBeInTheDocument()
    expect(document.getElementById('seitenleiste')).not.toHaveClass('absolute')

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteAusblenden }),
    )

    expect(umschalten).toHaveBeenCalledTimes(1)
  })
})

describe('Ein Fenster ohne matchMedia', () => {
  it('gilt als breit — die Leiste steht', () => {
    belegeApi({})
    Reflect.deleteProperty(window, 'matchMedia')

    render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    expect(
      screen.getByRole('navigation', { name: navigation.bezeichnung }),
    ).toBeInTheDocument()
  })
})
