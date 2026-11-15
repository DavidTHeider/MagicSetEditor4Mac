// **Die Einzeltasten an der Liste der Zonen** — P4
// (#2372):
// dieselben Tasten wie am gezeichneten Tisch, und eine angenommene greift
// nicht zugleich als Geste.
//
// **Und der Ziel-Sitz** — P5
// (#2373):
// er geht mit der Geste, statt beim Ablegen geraten zu werden.
import { fireEvent, render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import type { Partie } from '../apipartie'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { Partiezonen } from './Partiezonen'
import { einObjekt, einePartie, ruhenderLauf } from './partieprobe'

const mitWald = einePartie({
  objekte: [
    einObjekt({ kennung: 'k1', zone: 'tisch', name: 'Wald', geheim: false }),
  ],
})

function zeige(
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void,
  taste: (kennung: string, wort: string) => boolean,
  partie: Partie = mitWald,
) {
  render(
    <Sprachraum sprache="de">
      <Partiezonen
        partie={partie}
        texturen={ruhenderLauf()}
        gewaehlt={undefined}
        waehle={() => undefined}
        wirkungen={new Map()}
        erlaubt={['objekt.ziehen']}
        befehlen={befehlen}
        handlungen={[]}
        greife={() => undefined}
        tastenFuer={() => 'T'}
        taste={taste}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

const karte = () => screen.getByRole('button', { name: /Wald/ })
const ansage = () =>
  document.querySelector('[data-gestenansage]')?.textContent ?? ''

it('hält eine angenommene Einzeltaste an', () => {
  const taste = vi.fn(() => true)
  zeige(vi.fn(), taste)

  expect(karte()).toHaveAttribute('aria-keyshortcuts', 'T')
  expect(fireEvent.keyDown(karte(), { key: 't' })).toBe(false)
  expect(taste).toHaveBeenCalledWith('k1', 't')
})

// **Der Ziel-Sitz steht in der Absicht** — gesetzt mit dem Ziel: aufs
// Schlachtfeld der Beherrscher, sonst der Eigner (CR 110.2, 400.3).
it('nennt aufs Schlachtfeld den Beherrscher, sonst den Eigner', async () => {
  const befehlen = vi.fn()
  const geliehen = einePartie({
    objekte: [
      einObjekt({
        kennung: 'k1',
        zone: 'hand',
        name: 'Wald',
        geheim: false,
        eigner: 'sitz2',
        beherrscher: 'sitz1',
      }),
    ],
  })
  const nutzer = zeige(befehlen, () => false, geliehen)
  const { zoneTisch, zoneFriedhof } = kataloge.de.seiten.spielen
  // legeAb greift und wählt weiter, bis das Ziel die Zone `wort` nennt.
  const legeAb = async (wort: string) => {
    karte().focus()
    await nutzer.keyboard(' ')
    for (
      let schritt = 0;
      schritt < 8 && !ansage().includes(wort);
      schritt += 1
    ) {
      await nutzer.keyboard('{ArrowRight}')
    }
    await nutzer.keyboard('{Enter}')
  }

  await legeAb(zoneTisch)
  await legeAb(zoneFriedhof)

  const sitze = befehlen.mock.calls.map(
    ([, nutzlast]) => nutzlast as { zone: string; sitz: string },
  )
  expect(sitze).toEqual([
    expect.objectContaining({ zone: 'tisch', sitz: 'sitz1' }),
    expect.objectContaining({ zone: 'friedhof', sitz: 'sitz2' }),
  ])
})
