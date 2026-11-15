// **Der Rechtsklick an einer Partie zeigt alle ihre Handlungen** — AP-5
// (#2362),
// Befund B-01: weiterspielen, zuschauen und die Verwaltung, dieselben wie in
// ihren zwei Menüs. Löschen fragt hier so wenig wie dort — warum, steht in
// `Partienverwaltung.tsx`.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { Partiezeile } from '../apipartie'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { stilleVerwaltung } from '../test/verwaltung'
import { Partienzeile } from './Partienzeile'

const eigene = kataloge.de.seiten.spielen

const partie: Partiezeile = {
  id: 'p7',
  fassung: '1',
  spiel: 'magic',
  stand: 'laeuft',
  sitze: [
    { kennung: 's1', name: 'Anna' },
    { kennung: 's2', name: 'Bert' },
  ],
  objekte: 3,
  zug: 2,
  eigenerSitz: 's1',
  dran: false,
}

function zeige(einsteigen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <ul>
        <Partienzeile
          partie={partie}
          einsteigen={einsteigen}
          verwaltung={stilleVerwaltung()}
          marken={[]}
          nachladen={() => undefined}
        />
      </ul>
    </Sprachraum>,
  )

  return { einsteigen }
}

describe('der Rechtsklick an einer Partie', () => {
  it('zeigt alle ihre Handlungen und steigt über sie ein', async () => {
    const { einsteigen } = zeige()

    const ereignis = fireEvent.contextMenu(screen.getByText('Anna · Bert'), {
      clientX: 5,
      clientY: 5,
    })

    expect(ereignis).toBe(false)
    for (const name of [
      einsetzen(eigene.partienEinsteigenAls, { name: 'Anna' }),
      eigene.partienZuschauen,
      eigene.partienDuplizieren,
      eigene.partienAusfuhr,
      eigene.partienLoeschen,
      kataloge.de.bausteine.objektmarken.ueberschrift,
    ]) {
      expect(screen.getByRole('menuitem', { name })).toBeInTheDocument()
    }

    await userEvent.click(
      screen.getByRole('menuitem', {
        name: einsetzen(eigene.partienEinsteigenAls, { name: 'Bert' }),
      }),
    )
    expect(einsteigen).toHaveBeenCalledWith('p7', 's2')
  })
})
