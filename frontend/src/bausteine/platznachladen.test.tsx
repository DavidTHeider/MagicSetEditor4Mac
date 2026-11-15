// **Weggeräumt ist weggeräumt** — §7, Regel 3
// (#1970).
//
// # Wogegen hier geprüft wird
//
// Go meldet nach dem Wegräumen, dass die Zählung danach nicht zu machen war
// — `nachladefehler` neben einer leeren Liste. Bis dahin blieb die **alte**
// Zahl stehen, als gälte sie noch. Jetzt zählt die Anzeige selbst neu, und
// was dabei scheitert, steht an ihr.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Platz } from './Platz'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

it('zählt nach einem Wegräumen mit Nachladefehler selbst neu', async () => {
  const zaehlen = vi
    .fn()
    .mockResolvedValueOnce({
      posten: [
        {
          posten: 'vorsicherungen',
          dateien: 3,
          bytes: 348_000_000,
          raeumbar: true,
        },
      ],
      fehler: null,
    })
    .mockResolvedValue({
      posten: [],
      fehler: {
        code: 'unbekannt',
        schluessel: 'unbekannt',
        meldung: 'nichts geht',
        feld: '',
        werte: {},
      },
    })

  belegeApi({
    Platz: zaehlen,
    PlatzFreigeben: vi.fn().mockResolvedValue({
      posten: [],
      nachladefehler: {
        code: 'unbekannt',
        schluessel: 'unbekannt',
        meldung: 'nichts geht',
      },
      fehler: null,
    }),
  })

  render(
    <Sprachraum sprache="de">
      <Platz />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: /Sicherungen vor Migrationen/ }),
  )
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.bestandspflege.platzLoeschenJa,
    }),
  )

  await waitFor(() => {
    expect(zaehlen).toHaveBeenCalledTimes(2)
  })

  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.unbekannt,
  )
  expect(screen.queryByText(/348/)).toBeNull()
})
