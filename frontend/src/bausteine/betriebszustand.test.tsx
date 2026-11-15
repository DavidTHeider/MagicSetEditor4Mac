// **Der Betriebszustand im Fenster** — Werkstück O9 von Phase 5.2
// (#1433).
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Betriebszustand } from './Betriebszustand'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { GesundheitAntwort } from '../apigesundheit'

const texte = kataloge.de.seiten.technik

afterEach(raeumeApi)

// zeige zeichnet ihn über einer Antwort.
function zeige(teil: Partial<GesundheitAntwort> = {}) {
  belegeApi({
    Gesundheit: vi.fn(() =>
      Promise.resolve({
        lebt: true,
        bereit: true,
        grund: '',
        fremde: [],
        fehler: null,
        ...teil,
      }),
    ),
  })

  render(
    <Sprachraum sprache="de">
      <Betriebszustand />
    </Sprachraum>,
  )
}

describe('Der Betriebszustand', () => {
  it('sagt, dass er lebt und arbeiten kann', async () => {
    zeige()

    expect(await screen.findByText(texte.lebt)).toBeInTheDocument()
    expect(screen.getByText(texte.bereit)).toBeInTheDocument()

    // **Kein Grund ohne Grund** — „bereit, weil nichts dagegen spricht" wäre
    // eine Zeile ohne Aussage.
    expect(screen.queryByText(texte.grundBestandZu)).not.toBeInTheDocument()
  })

  it('nennt den Grund, wenn er nicht arbeiten kann', async () => {
    zeige({ bereit: false, grund: 'wanderungLaeuft' })

    expect(await screen.findByText(texte.nichtBereit)).toBeInTheDocument()
    expect(screen.getByText(texte.grundWanderungLaeuft)).toBeInTheDocument()
  })

  // **Ein Grund, den dieses Fenster nicht kennt, steht trotzdem da** — er ist
  // immer noch eine Auskunft.
  it('zeigt einen unbekannten Grund im Wortlaut', async () => {
    zeige({ bereit: false, grund: 'etwasNeues' })

    expect(await screen.findByText('etwasNeues')).toBeInTheDocument()
  })

  // **Die Regel, für die es diesen Baustein gibt** (§16, Satz 5).
  it('zeigt einen fremden Ausfall, ohne ihn gegen die Bereitschaft zu rechnen', async () => {
    zeige({
      fremde: [
        { dienst: 'scryfall', erreichbar: false, grund: 'kein Netz' },
        { dienst: 'mtgjson', erreichbar: true, grund: '' },
      ],
    })

    // **Bereit bleibt bereit.**
    expect(await screen.findByText(texte.bereit)).toBeInTheDocument()

    const liste = screen.getByRole('list', { name: texte.fremdeDienste })
    const zeilen = within(liste).getAllByRole('listitem')

    expect(zeilen).toHaveLength(2)
    expect(zeilen[0]).toHaveTextContent('scryfall')
    expect(zeilen[0]).toHaveTextContent(texte.nichtErreichbar)
    expect(zeilen[0]).toHaveTextContent('kein Netz')
    expect(zeilen[1]).toHaveTextContent(texte.erreichbar)
  })

  it('meldet, wenn die Auskunft selbst scheitert', async () => {
    belegeApi({
      Gesundheit: vi.fn(() =>
        Promise.resolve({
          lebt: false,
          bereit: false,
          grund: '',
          fremde: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'nichts zu holen',
          },
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Betriebszustand />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
