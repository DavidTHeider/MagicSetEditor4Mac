// **Eine Zeile der Tokenliste** — Werkstück O2 von Phase 5.2
// (#1426).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Tokenzeile } from './Tokenzeile'
import { kataloge } from '../sprache/kataloge'

import type { TokenDTO } from '../api'

const eigene = kataloge.de.seiten.nutzer

function einToken(teile: Partial<TokenDTO> = {}): TokenDTO {
  return {
    id: '1',
    profilID: '10',
    bezeichnung: 'Nachtlauf',
    principal: 'Rechner',
    scope: 'lesen',
    ablauf: '',
    angelegt: '2026-09-05T12:00:00Z',
    abgelaufen: false,
    letzteNutzung: '',
    vorgaengerBis: '',
    partie: '',
    sitz: '',
    fassung: '1',
    ...teile,
  }
}

function zeige(eines: TokenDTO, laeuft = false) {
  const erlaubnis = vi.fn()
  const rotieren = vi.fn()
  const zurueck = vi.fn()

  render(
    <Sprachraum sprache="de">
      <ul>
        <Tokenzeile
          eines={eines}
          laeuft={laeuft}
          aufErlaubnis={erlaubnis}
          aufRotieren={rotieren}
          aufZurueckziehen={zurueck}
        />
      </ul>
    </Sprachraum>,
  )

  return { erlaubnis, rotieren, zurueck }
}

describe('Tokenzeile', () => {
  it('nennt Principal, Gültigkeit und dass es nie benutzt wurde', () => {
    zeige(einToken())

    const zeile = screen.getByRole('listitem')
    expect(zeile).toHaveTextContent('Rechner')
    expect(zeile).toHaveTextContent(eigene.tokenUnbefristet)
    expect(zeile).toHaveTextContent(eigene.tokenNieBenutzt)
  })

  it('nennt die letzte Nutzung, den Ablauf und die Frist des alten Geheimnisses', () => {
    zeige(
      einToken({
        letzteNutzung: '2026-09-08T12:30:00Z',
        ablauf: '2027-01-01T00:00:00Z',
        vorgaengerBis: '2026-09-08T13:00:00Z',
      }),
    )

    const zeile = screen.getByRole('listitem')
    expect(zeile).toHaveTextContent('zuletzt benutzt 2026-09-08T12:30:00Z')
    expect(zeile).toHaveTextContent('2027-01-01T00:00:00Z')
    expect(zeile).toHaveTextContent(
      'altes Geheimnis gilt bis 2026-09-08T13:00:00Z',
    )
  })

  it('sagt, wenn es abgelaufen ist', () => {
    zeige(einToken({ ablauf: '2020-01-01T00:00:00Z', abgelaufen: true }))

    expect(screen.getByRole('listitem')).toHaveTextContent(
      eigene.tokenAbgelaufen,
    )
  })

  it('reicht Erlaubnis und Zurückziehen durch', async () => {
    const { erlaubnis, zurueck } = zeige(einToken())

    await userEvent.selectOptions(
      screen.getByRole('combobox', {
        name: `${eigene.tokenScope} — Nachtlauf`,
      }),
      'schreiben',
    )
    expect(erlaubnis).toHaveBeenCalledWith('schreiben')

    // **Zurückgezogen wird erst nach der Frage** (UX-10.AK-04, #1682) — und
    // Abbrechen zieht nichts zurück.
    const knopf = screen.getByRole('button', { name: eigene.tokenLoeschen })
    await userEvent.click(knopf)
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )
    expect(zurueck).not.toHaveBeenCalled()
    expect(knopf).toHaveFocus()

    await userEvent.click(knopf)
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: eigene.tokenLoeschenFrage.replace('{name}', 'Nachtlauf'),
        }),
      ).getByRole('button', { name: eigene.tokenLoeschenJa }),
    )
    expect(zurueck).toHaveBeenCalled()
  })

  it('ruht, solange ein Vorgang läuft', () => {
    zeige(einToken(), true)

    expect(
      screen.getByRole('button', { name: eigene.tokenRotieren }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: eigene.tokenLoeschen }),
    ).toBeDisabled()
  })
})
