import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Flaeche } from './Flaeche'
import { Knopf } from './Knopf'
import { Leerzustand } from './Leerzustand'
import { Seitenkopf } from './Seitenkopf'

// **Die Bausteine der Fläche** — die Stelle, an der das Aussehen der Anwendung
// **einmal** steht statt auf jeder Seite neu
// (#288).
//
// Sie tragen **keinen eigenen Text**: was daraufsteht, kommt vom Aufrufer aus
// dem Katalog. `keintextimjsx.test.ts` sähe es sonst — zu Recht.
describe('Flaeche', () => {
  it('zeigt ihren Inhalt', () => {
    render(
      <Flaeche>
        <p>Inhalt</p>
      </Flaeche>,
    )

    expect(screen.getByText('Inhalt')).toBeInTheDocument()
  })

  // **Eine Fläche mit Überschrift bekommt einen `section`-Bereich.** Ohne
  // Überschrift wäre er ein Bereich ohne Namen, und ein Vorleseprogramm zählte
  // ihn mit, ohne sagen zu können, was darin steht.
  it('wird mit Überschrift zu einem benannten Bereich', () => {
    render(
      <Flaeche ueberschrift="Angaben">
        <p>Inhalt</p>
      </Flaeche>,
    )

    expect(screen.getByRole('region', { name: 'Angaben' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Angaben' })).toBeInTheDocument()
  })

  // **Eine Marke steht neben der Überschrift und nicht in ihr** — „2 Sitze"
  // im Artboard *Eine Partie anfangen*. Der Name des Bereichs bleibt, was
  // er ist; eine Zahl darin änderte ihn bei jedem neuen Sitz.
  it('stellt eine Marke neben die Überschrift', () => {
    render(
      <Flaeche ueberschrift="Wer sitzt am Tisch" marke="2 Sitze">
        <p>Inhalt</p>
      </Flaeche>,
    )

    expect(
      screen.getByRole('region', { name: 'Wer sitzt am Tisch' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Wer sitzt am Tisch' }),
    ).toBeInTheDocument()
    expect(screen.getByText('2 Sitze')).toBeInTheDocument()
  })

  it('bleibt ohne Überschrift ein schlichter Rahmen', () => {
    render(
      <Flaeche>
        <p>Inhalt</p>
      </Flaeche>,
    )

    expect(screen.queryByRole('region')).not.toBeInTheDocument()
  })
})

describe('Seitenkopf', () => {
  it('zeigt Titel und Erklärung', () => {
    render(<Seitenkopf titel="Karten" erklaerung="Was hier passiert" />)

    expect(screen.getByRole('heading', { name: 'Karten' })).toBeInTheDocument()
    expect(screen.getByText('Was hier passiert')).toBeInTheDocument()
  })

  it('kommt ohne Erklärung aus', () => {
    render(<Seitenkopf titel="Karten" />)

    expect(screen.getByRole('heading', { name: 'Karten' })).toBeInTheDocument()
  })

  it('nimmt Handlungen auf', () => {
    render(
      <Seitenkopf titel="Karten">
        <button type="button">Neu</button>
      </Seitenkopf>,
    )

    expect(screen.getByRole('button', { name: 'Neu' })).toBeInTheDocument()
  })
})

describe('Knopf', () => {
  it.each(['haupt', 'neben', 'leise'] as const)(
    'löst in der Gewichtung %s aus',
    async (gewicht) => {
      const gedrueckt = vi.fn()
      render(
        <Knopf gewicht={gewicht} aufKlick={gedrueckt}>
          Los
        </Knopf>,
      )

      await userEvent.click(screen.getByRole('button', { name: 'Los' }))

      expect(gedrueckt).toHaveBeenCalledTimes(1)
    },
  )

  it('ist ohne Gewichtung ein Nebenknopf', async () => {
    const gedrueckt = vi.fn()
    render(<Knopf aufKlick={gedrueckt}>Los</Knopf>)

    await userEvent.click(screen.getByRole('button', { name: 'Los' }))

    expect(gedrueckt).toHaveBeenCalledTimes(1)
  })

  // **Ein angehaltener Knopf löst nicht aus** — und sagt es über `disabled`,
  // nicht nur über die Farbe.
  it('löst angehalten nicht aus', async () => {
    const gedrueckt = vi.fn()
    render(
      <Knopf aufKlick={gedrueckt} angehalten>
        Los
      </Knopf>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Los' }))

    expect(gedrueckt).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Los' })).toBeDisabled()
  })

  // **Klein ist kleiner in Schrift und Rand**, nicht in der Gewichtung — für
  // eine Reihe gleichrangiger Handlungen neben dem Tisch, wie das Artboard
  // *Tisch* sie zeigt (UX-08, #1680). Die Größe steht nur in den Klassen, und
  // deshalb wird sie dort gelesen.
  it('ist klein kleiner in Schrift und Rand', () => {
    render(
      <>
        <Knopf aufKlick={vi.fn()}>Groß</Knopf>
        <Knopf klein aufKlick={vi.fn()}>
          Klein
        </Knopf>
      </>,
    )

    expect(screen.getByRole('button', { name: 'Groß' })).toHaveClass(
      'px-3',
      'py-2',
    )
    expect(screen.getByRole('button', { name: 'Klein' })).toHaveClass(
      'px-2',
      'py-1',
      'text-sm',
    )
    expect(screen.getByRole('button', { name: 'Klein' })).not.toHaveClass(
      'py-2',
    )
  })
})

describe('Leerzustand', () => {
  // **Ein leerer Zustand erklärt den nächsten Schritt** (§6) — „Keine Daten"
  // ist keine Auskunft.
  it('nennt den nächsten Schritt', () => {
    render(<Leerzustand titel="Kein Set offen" erklaerung="Öffne eine Datei" />)

    expect(screen.getByText('Kein Set offen')).toBeInTheDocument()
    expect(screen.getByText('Öffne eine Datei')).toBeInTheDocument()
  })

  it('bietet die Handlung gleich mit an', async () => {
    const gedrueckt = vi.fn()
    render(
      <Leerzustand titel="Kein Set offen" erklaerung="Öffne eine Datei">
        <Knopf gewicht="haupt" aufKlick={gedrueckt}>
          Öffnen
        </Knopf>
      </Leerzustand>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Öffnen' }))

    expect(gedrueckt).toHaveBeenCalledTimes(1)
  })

  it('kommt ohne Handlung aus', () => {
    render(<Leerzustand titel="Kein Set offen" erklaerung="Öffne eine Datei" />)

    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })
  // **Eine Bezeichnung nennt, was die Beschriftung nicht sagt.** In einer Liste
  // steht „Sichern" fünfmal; wer nur die Beschriftung hört, weiß nicht, welches
  // Set gemeint ist.
  it('nimmt eine eigene Bezeichnung an', () => {
    render(
      <Knopf bezeichnung="Sichern — Vier Winde" aufKlick={vi.fn()}>
        Sichern
      </Knopf>,
    )

    expect(
      screen.getByRole('button', { name: 'Sichern — Vier Winde' }),
    ).toBeInTheDocument()
  })
})
