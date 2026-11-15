// **Ein Set mit Befunden in der Aufmerksamkeit** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`: „Set ‚Tischprobe‘: 12 Karten mit Befund — zeigt die Liste
// mit dem Filter ‚mit Befund‘“. Die Zeile kommt aus denselben Prüfständen wie
// die Kacheln; ohne sie sagte die Aufmerksamkeit „Nichts“, während die Kachel
// ein Set mit Problemen nannte.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { AufmerksamkeitProbe } from '../test/aufmerksamkeitprobe'
import type { Setbefund } from './problemlage'

const texte = kataloge.de.seiten.start.aufmerksamkeit

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

function zeige(setbefunde: readonly Setbefund[] | undefined) {
  belegeApi({
    Aufmerksamkeit: vi.fn().mockResolvedValue({
      hinweise: [],
      durchgesehen: { sets: 1, decks: 0, pakete: 0 },
      befunde: [],
      fehler: null,
    }),
  })
  const waehleSet = vi.fn()
  render(
    <Sprachraum sprache="de">
      <AufmerksamkeitProbe
        waehleSet={waehleSet}
        stand={0}
        setbefunde={setbefunde}
      />
    </Sprachraum>,
  )

  return waehleSet
}

describe('ein Set mit Befunden in der Aufmerksamkeit', () => {
  it('steht als Zeile da und öffnet die Karten mit dem Filter „mit Befund“', async () => {
    const waehleSet = zeige([
      { setID: '7', name: 'Tischprobe', zahl: 2, karteIDs: ['70', '71'] },
    ])
    const bereich = await screen.findByRole('region', { name: texte.titel })

    expect(
      within(bereich).getByRole('heading', { name: `${texte.titel} 1` }),
    ).toBeInTheDocument()
    await userEvent.click(
      within(bereich).getByRole('button', {
        name: /Tischprobe Set · 2 Karten mit Befund/,
      }),
    )

    // **Mit der ersten Karte mit Befund** — die Kartenseite schlägt sie
    // auf, statt der ersten der Liste (gesehen am Fenster, `ap8befunde09`).
    expect(waehleSet).toHaveBeenCalledWith('7', '70')
    expect(uebergeben.at(-1)).toEqual({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '7',
      karteIDs: ['70', '71'],
    })
  })

  // **Ohne geprüfte Befunde bleibt es bei „Nichts“** — ein Beleg, dass die
  // Zeile nur aus sicheren Ständen kommt.
  it('sagt ohne Befunde weiter „Nichts“', async () => {
    zeige(undefined)

    expect(
      await screen.findByText(
        /^Nichts — 1 Set, 0 Decks und 0 eigene Pakete durchgesehen\.$/,
      ),
    ).toBeInTheDocument()
  })
})
