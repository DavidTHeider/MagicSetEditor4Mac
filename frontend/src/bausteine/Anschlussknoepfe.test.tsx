// **Was der Inspector der Sammlung weitergibt** — UX-03.AK-01
// (#1675).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Anschlussknoepfe } from './Anschlussknoepfe'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { platzhalter } from '../haken/sammlungszeilen'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'

const texte = kataloge.de.seiten.sammlung

// **Der Raum wird hier nicht gestellt, sondern die Stelle eingesetzt.**
// Geprüft wird, *was* der Knopf übergibt — dass der Raum es dann mitnimmt,
// steht in `anschlussraum.test.tsx`.
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => eingesetzteStelle,
}))

let uebergeben: Anschluss | undefined

const eingesetzteStelle: Anschlussstelle = {
  offen: undefined,
  uebergib: (anschluss) => {
    uebergeben = anschluss
  },
  verbrauche: () => undefined,
}

const dasExemplar = {
  ...platzhalter(0),
  id: 4,
  name: 'Aschenwächter',
  printing: 'p-4',
  zustand: 'nearMint',
  sprache: 'de',
  finish: 'foil' as const,
}

function stelleDar() {
  render(
    <Sprachraum sprache="de">
      <Anschlussknoepfe exemplar={dasExemplar} />
    </Sprachraum>,
  )
}

describe('die Anschlussknöpfe der Sammlung', () => {
  it('schickt den Druck ins Deck', async () => {
    stelleDar()

    await userEvent.click(screen.getByRole('button', { name: texte.insDeck }))

    expect(uebergeben).toEqual({
      art: 'deck',
      name: 'Aschenwächter',
      printing: 'p-4',
    })
  })

  // **Güte, Sprache und Ausführung reisen mit.** Ohne sie füllte ein Mensch
  // dieselben vier Felder noch einmal aus, die er in der Sammlung längst
  // eingetragen hat.
  it('schickt Druck, Güte, Sprache und Ausführung zum Tausch', async () => {
    stelleDar()

    await userEvent.click(screen.getByRole('button', { name: texte.zumTausch }))

    expect(uebergeben).toEqual({
      art: 'handel',
      name: 'Aschenwächter',
      printing: 'p-4',
      guete: 'nearMint',
      sprache: 'de',
      finish: 'foil',
    })
  })
})
