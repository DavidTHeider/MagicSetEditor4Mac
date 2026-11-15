// **Wann zuletzt gesichert wurde, auf der Seite** — OPS01
// (#2347).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { Sicherungsstandlauf } from '../haken/useSicherungsstand'
import { Sprachraum } from '../sprache/kontext'
import { standMit } from '../test/sicherungsstand'
import { Sicherungsalter } from './Sicherungsalter'

function zeige(teile: Partial<Sicherungsstandlauf> = {}) {
  const lauf: Sicherungsstandlauf = {
    stand: standMit(),
    planfehler: undefined,
    fehler: undefined,
    einrichten: vi.fn(),
    beenden: vi.fn(),
    laeuft: false,
    ...teile,
  }

  render(
    <Sprachraum sprache="de">
      <Sicherungsalter lauf={lauf} />
    </Sprachraum>,
  )

  return lauf
}

describe('Sicherungsalter', () => {
  it('sagt unbekannt, wo es keinen Vermerk gibt — und nie heute', () => {
    zeige()

    expect(screen.getByText('unbekannt')).toBeInTheDocument()
    expect(screen.getByText(/noch nie gesichert/)).toBeInTheDocument()
    expect(screen.queryByText('heute')).not.toBeInTheDocument()
  })

  it.each([
    [0, 'heute'],
    [1, 'vor einem Tag'],
    [9, 'vor 9 Tagen'],
  ])('nennt das Alter nach %i Tagen', (tage, erwartet) => {
    zeige({
      stand: standMit({
        stufe: 'frisch',
        tage,
        gesichert: '2026-10-15T09:30:00Z',
        flaeche: 'cli',
        anlass: 'hand',
        datei: 'a.msesicherung',
      }),
    })

    expect(screen.getByText(erwartet)).toBeInTheDocument()
    expect(
      screen.getByText(
        /über die Kommandozeile in a\.msesicherung, zurückgelesen und geprüft/,
      ),
    ).toBeInTheDocument()
  })

  it('hebt über 14 Tage hervor', () => {
    zeige({
      stand: standMit({
        stufe: 'alt',
        tage: 16,
        gesichert: '2026-09-29T09:30:00Z',
        flaeche: 'fenster',
        anlass: 'hand',
      }),
    })

    expect(screen.getByText('vor 16 Tagen · älter als 14 Tage')).toHaveClass(
      'text-fehler',
    )
  })

  it('nennt eine Zeit nach der Uhr und eine Sicherung nach Plan', () => {
    zeige({
      stand: standMit({
        stufe: 'unstimmig',
        gesichert: '2027-01-01T00:00:00Z',
      }),
    })

    expect(screen.getByText('Zeit nach der Uhr')).toBeInTheDocument()
    expect(screen.getByText(/nach der Uhr dieses Rechners/)).toBeInTheDocument()
  })

  it('nennt einen Lauf nach Plan und eine unbekannte Fläche mit ihrem Namen', () => {
    zeige({
      stand: standMit({
        stufe: 'frisch',
        tage: 2,
        anlass: 'plan',
        flaeche: 'neu',
        datei: 'p.msesicherung',
        gesichert: '2026-10-13T09:30:00Z',
        plan: {
          ordner: '/Volumes/Stick',
          letzter: '2026-10-13T09:30:00Z',
          faellig: false,
        },
      }),
    })

    expect(screen.getByText(/nach Plan in p\.msesicherung/)).toBeInTheDocument()
    expect(screen.getByText('Ordner: /Volumes/Stick')).toBeInTheDocument()
    expect(screen.getByText(/Zuletzt nach Plan am/)).toBeInTheDocument()
    expect(screen.getByRole('checkbox')).toBeChecked()
  })

  it('richtet den Plan ein und beendet ihn über das Kästchen', async () => {
    const ohne = zeige()

    await userEvent.click(
      screen.getByRole('checkbox', {
        name: /Wöchentlich in einen Ordner sichern/,
      }),
    )
    expect(ohne.einrichten).toHaveBeenCalledTimes(1)
  })

  it('beendet einen eingerichteten Plan und nennt seinen Fehler', async () => {
    const mit = zeige({
      stand: standMit({
        plan: { ordner: '/Stick', letzter: '', faellig: true },
      }),
      planfehler: 'Der Ordner fehlt.',
      fehler: 'Belegt.',
    })

    expect(screen.getByText(/Nach Plan noch nie gesichert/)).toBeInTheDocument()
    expect(
      screen.getByText(
        /Der Plan konnte zuletzt nicht sichern: Der Ordner fehlt\./,
      ),
    ).toBeInTheDocument()
    expect(screen.getByText('Belegt.')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('checkbox'))
    expect(mit.beenden).toHaveBeenCalledTimes(1)
  })

  it('zeigt ohne Stand nur einen Fehler — oder nichts', () => {
    zeige({ stand: undefined, fehler: 'Belegt.' })

    expect(screen.getByRole('alert')).toHaveTextContent('Belegt.')
  })

  it('zeigt ohne Stand und Fehler nichts', () => {
    zeige({ stand: undefined })

    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
  })
})
