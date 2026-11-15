// **Die Startseite sagt, dass sie lädt, und bietet nach einem Scheitern den
// nächsten Versuch an** — G11
// (#2277).
//
// B03 verzögerte die Bindungen der Startseite um vier Sekunden: „Weiterarbeiten"
// und „Was Aufmerksamkeit braucht" standen so lange gar nicht da, ohne Satz.
// Scheiterten sie, stand eine Meldung da — und um es noch einmal zu
// versuchen, blieb nur, die Seite zu verlassen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AufmerksamkeitAntwort } from '../apiaufmerksamkeit'
import type { WeiterarbeitenAntwort } from '../apiweiterarbeiten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { AufmerksamkeitProbe } from '../test/aufmerksamkeitprobe'
import { Weiterarbeiten } from './Weiterarbeiten'

const zustaende = kataloge.de.bausteine.zustaende

const fehlschlag = {
  code: 'unbekannt' as const,
  schluessel: '',
  meldung: 'die Ablage schweigt',
}

const leereArbeit: WeiterarbeitenAntwort = {
  zuletzt: null,
  geoeffnet: [],
  eingefuehrt: [],
  ausgegeben: [],
  angeheftet: [],
  fehler: null,
}

const leereAufmerksamkeit: AufmerksamkeitAntwort = {
  hinweise: [],
  durchgesehen: { sets: 0, decks: 0, pakete: 0 },
  befunde: [],
  fehler: null,
}

afterEach(raeumeApi)

const abschnitte = [
  {
    name: 'Weiterarbeiten',
    bindung: 'Weiterarbeiten',
    leer: leereArbeit,
    titel: kataloge.de.seiten.start.weiterarbeiten.titel,
    zeige: () => (
      <Weiterarbeiten wechseln={vi.fn()} waehleSet={vi.fn()} stand={0} />
    ),
  },
  {
    name: 'Aufmerksamkeit',
    bindung: 'Aufmerksamkeit',
    leer: leereAufmerksamkeit,
    titel: kataloge.de.seiten.start.aufmerksamkeit.titel,
    zeige: () => <AufmerksamkeitProbe waehleSet={vi.fn()} stand={0} />,
  },
] as const

describe.each(abschnitte)('$name auf der Startseite', (abschnitt) => {
  it('sagt, dass er lädt, solange die Antwort aussteht', async () => {
    belegeApi({ [abschnitt.bindung]: () => new Promise(() => undefined) })

    render(<Sprachraum sprache="de">{abschnitt.zeige()}</Sprachraum>)

    expect(await screen.findByRole('status')).toHaveTextContent(zustaende.laedt)
  })

  it('nennt ein Scheitern samt Grund und lädt erneut', async () => {
    let gefragt = 0
    belegeApi({
      [abschnitt.bindung]: () => {
        gefragt++

        return Promise.resolve(
          gefragt === 1
            ? { ...abschnitt.leer, fehler: fehlschlag }
            : abschnitt.leer,
        )
      },
    })

    render(<Sprachraum sprache="de">{abschnitt.zeige()}</Sprachraum>)

    // **Der Vorsatz bleibt**: auf der Startseite stehen zwei Abschnitte, und
    // „Das ließ sich nicht laden." sagt, dass es um das Laden ging.
    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(zustaende.fehler)
    expect(meldung).toHaveTextContent('die Ablage schweigt')

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )

    expect(
      await screen.findByRole('region', { name: abschnitt.titel }),
    ).toBeVisible()
    expect(screen.queryByRole('alert')).toBeNull()
    expect(gefragt).toBe(2)
  })
})
