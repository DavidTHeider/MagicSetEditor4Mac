// **Scheitert die Vorschau beim Tippen, steht es da** — E01
// (#2238).
//
// Am nativen Fenster gefunden (`e01vorschau`): scheiterte die Probevorschau,
// behielt die Fläche das Bild des **vorigen** Entwurfs und sagte nichts — eine
// alte Vorschau, die aussah wie die zum neuen Text. `useKartenvorschau` reichte
// nur den Fehler der gespeicherten Vorschau weiter.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const karte: KarteDTO = {
  id: '5',
  nummer: 0,
  werte: [{ feld: 'name', wert: 'Feuerball' }],
}

function bild(name: string) {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: `/vorschau/${name}.png`,
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

it('meldet eine gescheiterte Probevorschau und erholt sich mit der nächsten Eingabe', async () => {
  let scheitern = false
  const Probevorschau = vi.fn((_karte: string, _feld: string, wert: string) =>
    scheitern
      ? Promise.reject(new Error('der Rasterer ist ausgefallen'))
      : Promise.resolve(bild(wert)),
  )

  belegeApi({
    Vorschau: () => Promise.resolve(bild('gespeichert')),
    Probevorschau,
    FeldSetzen: () => Promise.resolve({}),
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={[]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )

  const feld = await screen.findByLabelText('name')
  await userEvent.type(feld, 'X')
  await waitFor(() => {
    expect(screen.getByAltText('Vorschau der Karte')).toHaveAttribute(
      'src',
      expect.stringContaining('FeuerballX'),
    )
  })

  scheitern = true
  await userEvent.type(feld, 'Y')

  // **Die Fläche sagt, dass das Bild nicht zum Text gehört.**
  const meldung = await screen.findByRole('alert')
  expect(meldung).toHaveTextContent('der Rasterer ist ausgefallen')

  // **Und die nächste Eingabe, die gelingt, nimmt die Meldung weg.**
  scheitern = false
  await userEvent.type(feld, 'Z')
  await waitFor(() => {
    expect(screen.getByAltText('Vorschau der Karte')).toHaveAttribute(
      'src',
      expect.stringContaining('FeuerballXYZ'),
    )
  })
  expect(screen.queryByRole('alert')).toBeNull()
})
