// **Eine abgeschaltete Seite erklärt sich** — OPS-10
// (#1908).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einstellungsprobe, zeigeSeite } from '../test/seitenvorrichtung'
import { Abgeschaltet } from './Abgeschaltet'

const eigene = kataloge.de.seiten.einstellungen.schalter.abgeschaltet

afterEach(() => {
  raeumeApi()
})

function mitSpielenAus() {
  return {
    ...einstellungsprobe,
    wert: (schluessel: string) =>
      schluessel === 'schalter.spielen'
        ? 'aus'
        : einstellungsprobe.wert(schluessel),
  }
}

describe('eine abgeschaltete Seite', () => {
  it('zeigt statt der Seite, warum sie fehlt', async () => {
    belegeApi({})
    zeigeSeite('decks', 'de', vi.fn(), mitSpielenAus())

    expect(
      await screen.findByText(
        eigene.titel.replace('{seite}', kataloge.de.seiten.decks.titel),
      ),
    ).toBeVisible()

    expect(screen.getByRole('button', { name: eigene.oeffnen })).toBeVisible()
  })

  // **Im Anschlussraum, wie im Fenster** — er trägt die Gruppe zur
  // Einstellungsseite, die sie aufschlägt.
  it('führt zu den Schaltern', async () => {
    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Anschlussraum aktuell="decks" wechseln={wechseln}>
          <Abgeschaltet seite="decks" />
        </Anschlussraum>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: eigene.oeffnen }))

    expect(wechseln).toHaveBeenCalledWith('einstellungen')
  })

  it('steht nicht da, wenn der Bereich an ist', async () => {
    belegeApi({})

    zeigeSeite('start', 'de', vi.fn(), mitSpielenAus())

    expect(await screen.findByText(kataloge.de.seiten.start.was)).toBeVisible()
    expect(screen.queryByRole('button', { name: eigene.oeffnen })).toBeNull()
  })
})
