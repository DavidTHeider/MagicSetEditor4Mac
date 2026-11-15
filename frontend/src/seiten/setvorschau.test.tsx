// **Die Vorschau vor dem Einlesen eines Sets** — FMT-03.AK-01
// (#1692).
//
// Sie erscheint nur, wenn es etwas zu sagen gibt — Stellen, die sich nicht
// lesen lassen, oder Pakete, die fehlen. Eine saubere Datei wird gleich
// eingelesen; das prüft `sets.test.tsx`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Setvorschau } from '../antwortenbestand'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const t = texte.seiten.sets.vorschau

function vorschauMit(teile: Partial<Setvorschau>): Setvorschau {
  return {
    pfad: '/tmp/mitbedarf.mse-set',
    name: 'Mit Bedarf',
    spiel: 'magic',
    stil: 'm15',
    karten: 12,
    befunde: [],
    schonOffen: '',
    bedarf: [
      {
        paket: 'Mit Bedarf',
        braucht: 'magic.mse-game',
        art: 'fehlt',
        verlangt: '',
        da: '',
      },
    ],
    ...teile,
  }
}

const eingelesen = {
  setID: '4',
  karten: 12,
  befunde: [],
  schonOffen: false,
  angelegt: ['4'],
  unvollstaendig: false,
  abgebrochen: false,
}

function zeige(vorschau: Setvorschau) {
  const SetAusDatei = vi.fn().mockResolvedValue(eingelesen)

  belegeApi({
    Sets: () => Promise.resolve({ sets: [] }),
    SetVorschauWaehlen: () => Promise.resolve({ abgebrochen: false, vorschau }),
    SetAusDatei,
  })

  const { container } = render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return { SetAusDatei, container }
}

async function oeffnen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
  )
}

describe('Die Vorschau vor dem Einlesen', () => {
  it('nennt fehlende Pakete und liest erst auf Wunsch ein', async () => {
    const { SetAusDatei, container } = zeige(vorschauMit({}))

    await oeffnen()

    const flaeche = await screen.findByRole('region', {
      name: einsetzen(t.ueberschrift, { name: 'Mit Bedarf' }),
    })

    expect(flaeche).toHaveTextContent(
      einsetzen(t.umfang, { karten: '12 Karten', spiel: 'magic', stil: 'm15' }),
    )
    expect(flaeche).toHaveTextContent(t.bedarf)
    expect(flaeche).toHaveTextContent('magic.mse-game')
    expect(SetAusDatei).not.toHaveBeenCalled()
    expect(await verstoesse(container)).toEqual([])

    await userEvent.click(screen.getByRole('button', { name: t.einlesen }))

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalledWith('/tmp/mitbedarf.mse-set')
    })
    expect(screen.queryByRole('region', { name: /Mit Bedarf/ })).toBeNull()
  })

  it('lässt sich verwerfen, ohne einzulesen', async () => {
    const { SetAusDatei } = zeige(vorschauMit({}))

    await oeffnen()
    await userEvent.click(
      await screen.findByRole('button', { name: t.verwerfen }),
    )

    expect(screen.queryByRole('region', { name: /Mit Bedarf/ })).toBeNull()
    expect(SetAusDatei).not.toHaveBeenCalled()
  })

  // **Eine schon offene Datei wird aufgeschlagen, nicht ein zweites Mal
  // angelegt** — und die Stellen, die sich nicht lesen lassen, stehen da.
  it('sagt, dass die Datei schon im Bestand liegt', async () => {
    zeige(
      vorschauMit({
        stil: '',
        schonOffen: '4',
        bedarf: [],
        befunde: ['set:12: unlesbare Karte'],
      }),
    )

    await oeffnen()

    expect(await screen.findByText(t.schonOffen)).toBeInTheDocument()
    expect(screen.getByText('set:12: unlesbare Karte')).toBeInTheDocument()
    expect(
      screen.getByText(
        einsetzen(t.umfangOhneStil, { karten: '12 Karten', spiel: 'magic' }),
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: t.aufschlagen }),
    ).toBeInTheDocument()
  })
})
