// **Ein Set ohne Karte sagt es auf jeder Lasche, die Karten braucht** — G02
// (#2249).
//
// Gemessen an einem leeren Set (`belegeProbeset({ leer: true })`): Suche und
// Ersetzen boten ein Feld an, das nie etwas finden kann, die Statistik zählte
// Nullen, der Stil sagte „leg eine an" ohne Knopf, und der Zufallspack sagte
// es erst nach dem Ziehen. Ausgeben und Konsole führten schon zu den Karten.
//
// **Was er nicht hält**: dass der Satz gut erklärt. Das zeigt das Bild der
// Zustandsprobe (`scripts/huellenprobe/zustaende.mjs`, Zustand „leeres Set").
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { belegeProbeset, zeigeSeite } from '../test/seitenvorrichtung'

const arbeitsplatz = kataloge.de.allgemein.arbeitsplatz

afterEach(raeumeApi)

const laschen = ['suche', 'ersetzen', 'statistik', 'stil', 'pack'] as const

describe('Ein Set ohne Karte', () => {
  for (const lasche of laschen) {
    it(`führt auf »${lasche}« zu den Karten`, async () => {
      belegeProbeset({ leer: true })
      const wechseln = vi.fn()

      zeigeSeite(lasche, 'de', wechseln)

      const leer = (
        await screen.findByText(arbeitsplatz.ohneKartenTitel)
      ).closest('[data-leerzustand]')

      expect(leer).not.toBeNull()

      await userEvent.click(
        within(leer as HTMLElement).getByRole('button', {
          name: arbeitsplatz.zuDenKarten,
        }),
      )

      expect(wechseln).toHaveBeenCalledWith('karten')
    })

    it(`schweigt auf »${lasche}«, sobald eine Karte da ist`, async () => {
      belegeProbeset()

      zeigeSeite(lasche, 'de')

      // **Erst, wenn das Set gelesen ist** (§3, ein Zuhörer vor der
      // Verneinung): der Name steht dann in den Brotkrumen.
      await screen.findAllByText('Probeset')

      expect(screen.queryByText(arbeitsplatz.ohneKartenTitel)).toBeNull()
    })
  }
})
