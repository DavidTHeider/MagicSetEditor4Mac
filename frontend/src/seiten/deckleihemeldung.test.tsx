// **Eine angenommene Leihe meldet sich bei der Seite** — P1b
// (#2370).
//
// Aus `deckleihe.test.tsx` herausgelöst, als die Datei über die Zeilengrenze
// aus §2 ging. Der Schnitt folgt der Frage: dort, was die Leihe ist; hier, wem
// eine Annahme Bescheid gibt.
import { screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { antwortAbwarten } from '../test/api'
import {
  eigene,
  leer,
  leiheApi,
  oeffneAnnahme,
  umschlagGewaehlt,
  waehleUmschlag,
  zeige,
} from '../test/leiheprobe'

// **Die Deckwahl am Sitz liest die Leihen selbst** und weiß sonst nichts von
// der Annahme. Gefunden am nativen Fenster (`p1b09`): die Leihe fehlte am Sitz,
// bis die Seite neu aufging.
it('meldet ein geglücktes Annehmen an die Seite, ein gescheitertes nicht', async () => {
  const geaendert = vi.fn()
  const genommen = vi
    .fn()
    .mockResolvedValueOnce(leer)
    .mockResolvedValueOnce({
      ...leer,
      fehler: {
        code: 'unbekannt' as const,
        schluessel: 'unbekannt',
        meldung: 'Der Umschlag ist beschädigt.',
      },
    })
  leiheApi({
    DeckleiheAnnehmen: genommen,
    UmschlagWaehlen: vi
      .fn()
      .mockResolvedValue(umschlagGewaehlt('/tmp/fremd.msepackage')),
  })

  const hand = zeige(geaendert)

  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/fremd.msepackage')
  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  await waitFor(() => {
    expect(geaendert).toHaveBeenCalledTimes(1)
  })
  await antwortAbwarten()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))
  await waitFor(() => {
    expect(genommen).toHaveBeenCalledTimes(2)
  })
  await antwortAbwarten()
  expect(geaendert).toHaveBeenCalledTimes(1)
})
