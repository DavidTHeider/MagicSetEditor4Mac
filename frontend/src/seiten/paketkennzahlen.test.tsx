// **Die Kennzahlen der Pakete** — AP-9, Teil 2
// (#2366),
// Entwurf `Bereiche09`: installiert und mit Befund, jede Zahl mit Ziel.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import type { AbhaengigkeitDTO, PaketDTO, PaketeAntwort } from '../pakete'
import type { Ladezustand } from '../haken/useLaden'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { katalog } from '../test/paketliste'
import { Paketkennzahlen } from './Paketkennzahlen'

const texte = kataloge.de.seiten.pakete.leiste

const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

afterEach(() => {
  uebergeben.length = 0
})

function paket(name: string, art = 'mse-style'): PaketDTO {
  const [vorlage] = katalog().pakete
  if (vorlage === undefined) {
    throw new Error('die Vorrichtung trägt kein Paket')
  }

  return { ...vorlage, name, art, anzeigename: name }
}

function befund(paket: string): AbhaengigkeitDTO {
  return {
    paket,
    braucht: 'magic.mse-game',
    art: 'fehlt',
    verlangt: '',
    da: '',
  }
}

function zeige(
  stand: Partial<Ladezustand<PaketeAntwort>>,
  daten: AbhaengigkeitDTO[] | undefined,
  fehler?: string,
) {
  render(
    <Sprachraum sprache="de">
      <Paketkennzahlen
        pakete={{
          daten: undefined,
          laedt: false,
          fehler: undefined,
          nachladen: vi.fn(),
          ...stand,
        }}
        befunde={{ daten, fehler }}
      />
    </Sprachraum>,
  )
}

const leiste = () =>
  within(screen.getByRole('toolbar', { name: texte.bezeichnung }))

describe('die Leiste der Pakete', () => {
  it('führt zu allen und zu denen mit Befund', async () => {
    zeige(
      {
        daten: {
          pakete: [paket('eins'), paket('zwei'), paket('drei', 'mse-game')],
          befunde: [],
          fehler: null,
        },
      },
      [
        befund('eins.mse-style'),
        befund('eins.mse-style'),
        befund('drei.mse-game'),
      ],
    )

    await userEvent.click(
      leiste().getByRole('button', { name: `${texte.installiert} 3` }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'paketsuche',
      name: texte.installiert,
      paket: '',
    })

    await userEvent.click(
      leiste().getByRole('button', { name: `${texte.befund} 2` }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'paketauswahl',
      name: '2 Pakete mit Befund',
      namen: ['eins', 'drei'],
    })
  })

  it('führt mit null Befunden nirgends hin', () => {
    zeige({ daten: { pakete: [paket('eins')], befunde: [], fehler: null } }, [])
    expect(leiste().getByText('0')).toBeInTheDocument()
    expect(leiste().getAllByRole('button')).toHaveLength(1)
  })

  it('steht beim Laden in Endhöhe', () => {
    zeige({ laedt: true }, undefined)
    expect(
      screen.getByRole('toolbar', { name: texte.bezeichnung }),
    ).toHaveAttribute('aria-busy', 'true')
  })

  it('sagt „nicht gelesen“, wenn die Liste nicht zu lesen war', () => {
    zeige({ fehler: 'kaputt' }, [])
    expect(leiste().getAllByText(texte.nichtGelesen)).toHaveLength(1)
  })

  // **Ein Lesefehler ist kein Laden** — sonst stünde die Leiste für immer
  // mit `aria-busy` da (gefunden von `barrierefreiheit.test.tsx`).
  it('sagt „nicht gelesen“, wenn die Befunde nicht zu lesen waren', () => {
    zeige(
      { daten: { pakete: [paket('eins')], befunde: [], fehler: null } },
      undefined,
      'kaputt',
    )
    expect(leiste().getByText(texte.nichtGelesen)).toBeInTheDocument()
    expect(
      screen.getByRole('toolbar', { name: texte.bezeichnung }),
    ).not.toHaveAttribute('aria-busy')
  })
})
