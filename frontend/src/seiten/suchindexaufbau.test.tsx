// **Der Aufbau des Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Der Satz, um den es geht**: nach dem Aufbau wird der Stand neu geholt —
// **auch nach einem Fehlschlag**. Ein gescheiterter Aufbau hinterlässt einen
// Fehlerzustand, und der ist die Auskunft, die sagt, was war (§7 Regel 3).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Suchindex } from './Suchindex'
import { Suchindexaufbau } from './Suchindexaufbau'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.technik

// einBericht ist ein geglückter Aufbau.
function einBericht() {
  return {
    generation: 3,
    geschrieben: 4711,
    dauer: 1234,
    fortgeschrieben: false,
    manifest: {
      modell: 'all-minilm',
      modellrevision: 'aaaa',
      laufzeit: 'ollama 0.33.3',
      abstandsmass: 'cosinus',
      dimension: 384,
      zerteilfassung: 'karten-seiten-1',
      suchkonfiguration: 'rrf-60',
      korpusstaende: [],
    },
    fehler: null,
  }
}

// einStand ist die Auskunft ohne Index.
function einStand() {
  return {
    aktiv: false,
    manifest: einBericht().manifest,
    aufbau: { laeuft: false, generation: 0, geschrieben: 0, fehler: '' },
    suchbar: false,
    grund: '',
    fehler: null,
  }
}

// fuelleUndBaue trägt die drei Felder ein und drückt den Knopf.
async function fuelleUndBaue() {
  await userEvent.type(
    screen.getByLabelText(texte.suchindexModellfeld),
    'all-minilm',
  )
  await userEvent.type(
    screen.getByLabelText(texte.suchindexQuellfeld),
    '/tmp/oracle.jsonl.gz',
  )
  await userEvent.type(
    screen.getByLabelText(texte.suchindexRevisionfeld),
    '2026-09-01',
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.suchindexAufbauen }),
  )
}

describe('Der Aufbau des Suchindex', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('reicht die drei Angaben durch und meldet die Generation', async () => {
    const aufbau = vi.fn(() => Promise.resolve(einBericht()))
    const danach = vi.fn()

    belegeApi({ IndexAufbauen: aufbau })

    render(
      <Sprachraum sprache="de">
        <Suchindexaufbau danach={danach} />
      </Sprachraum>,
    )

    await fuelleUndBaue()

    expect(aufbau).toHaveBeenCalledWith(
      'all-minilm',
      '/tmp/oracle.jsonl.gz',
      '2026-09-01',
      '',
      false,
    )

    expect(await screen.findByText(/4711/)).toBeInTheDocument()

    // **Und der Stand wird danach geholt.**
    expect(danach).toHaveBeenCalled()
  })

  it('meldet einen Fehlschlag und holt den Stand trotzdem', async () => {
    const danach = vi.fn()

    belegeApi({
      IndexAufbauen: () => Promise.reject(new Error('die Platte ist voll')),
    })

    render(
      <Sprachraum sprache="de">
        <Suchindexaufbau danach={danach} />
      </Sprachraum>,
    )

    await fuelleUndBaue()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Platte ist voll',
    )

    expect(danach).toHaveBeenCalled()
  })

  // **Und der frische Stand kommt wirklich an** — hier über die echte
  // Kopplung und nicht über einen Spion.
  it('zeigt nach dem Aufbau den frischen Stand', async () => {
    let gefragt = 0

    belegeApi({
      Indexstand: () => {
        gefragt += 1

        return Promise.resolve(
          gefragt === 1 ? einStand() : { ...einStand(), aktiv: true },
        )
      },
      IndexAufbauen: () => Promise.resolve(einBericht()),
    })

    render(
      <Sprachraum sprache="de">
        <Suchindex />
      </Sprachraum>,
    )

    await screen.findByText(texte.suchindexKeiner)

    await fuelleUndBaue()

    await waitFor(() => {
      expect(screen.queryByText(texte.suchindexKeiner)).not.toBeInTheDocument()
    })
  })

  // **Und ein Fehler beim Nachholen erreicht die Oberfläche** — sonst bliebe
  // ein veralteter Stand stehen, ohne dass jemand es merkt.
  it('meldet, wenn der Stand danach nicht lesbar ist', async () => {
    let gefragt = 0

    belegeApi({
      Indexstand: () => {
        gefragt += 1

        return gefragt === 1
          ? Promise.resolve(einStand())
          : Promise.reject(new Error('der Bestand ist zu'))
      },
      IndexAufbauen: () => Promise.resolve(einBericht()),
    })

    render(
      <Sprachraum sprache="de">
        <Suchindex />
      </Sprachraum>,
    )

    await screen.findByText(texte.suchindexKeiner)

    await fuelleUndBaue()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Bestand ist zu',
    )
  })
})

// **Der Schalter für den Deltaimport kommt an.**
//
// **Er ist die Ansage „baue auf dem Vorhandenen auf"**, und was das heißt,
// entscheidet der Fachkern: was die Quelle liefert, wird geschrieben, der
// Rest **kopiert** statt neu eingebettet.
describe('Der Deltaimport', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('reicht den Schalter durch', async () => {
    const aufbau = vi.fn(() => Promise.resolve(einBericht()))

    belegeApi({ IndexAufbauen: aufbau })

    render(
      <Sprachraum sprache="de">
        <Suchindexaufbau danach={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByLabelText(texte.suchindexDelta))

    await fuelleUndBaue()

    expect(aufbau).toHaveBeenCalledWith(
      'all-minilm',
      '/tmp/oracle.jsonl.gz',
      '2026-09-01',
      '',
      true,
    )
  })
})
