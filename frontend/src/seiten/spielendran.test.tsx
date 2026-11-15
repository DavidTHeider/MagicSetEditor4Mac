// **Die Liste ist neu, wenn man vom Tisch zurückkommt** — UXR-F16
// (#2031).
//
// # Wogegen hier geprüft wird
//
// **Gegen eine Marke, die älter ist als der Zug.** „dran" springt am Tisch
// um — wer passt, ist es nicht mehr —, und die Liste stammte bis dahin vom
// Laden der Seite. Der Lauf setzt sich an einen Tisch, lässt die Antwort der
// Liste beim zweiten Lesen eine andere sein und prüft, dass die Marke der
// zweiten folgt. **Zweimal gelesen und nicht öfter**: die Seite fragt beim
// Öffnen und beim Verlassen, nicht bei jedem Zeichnen.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { antwortAbwarten, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import { eineAntwort } from './partieprobe'
import { mitDecks, steigeEin, zeige } from './spielenprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// eineListe gibt die eine Partie, wie die Liste sie bei einem Lesen nennt.
function eineListe(dran: boolean) {
  return {
    partien: [
      {
        id: 'p1',
        fassung: '3',
        spiel: 'magic',
        stand: 'laeuft',
        sitze: [
          { kennung: 'sitz1', name: 'Anna' },
          { kennung: 'sitz2', name: 'Bert' },
        ],
        objekte: 12,
        zug: 2,
        eigenerSitz: 'sitz1',
        dran,
      },
    ],
    fehler: null,
  }
}

it('liest die Liste neu, wenn der Tisch verlassen wird', async () => {
  let gelesen = 0

  mitDecks({
    // **Das erste Lesen sagt „dran", das zweite nicht** — am Tisch wurde
    // gepasst.
    Partien: () => {
      gelesen += 1

      return Promise.resolve(eineListe(gelesen === 1))
    },
    Partie: () => Promise.resolve(eineAntwort()),
  })

  const hand = zeige()

  // **Erst ein Beleg, dass die Marke da ist** — eine Verneinung danach wäre
  // sonst auch grün, wenn sie nie dagewesen wäre (§3).
  expect(await screen.findByText(eigene.partienDran)).toBeInTheDocument()
  expect(gelesen).toBe(1)

  await steigeEin(hand, 'Anna · Bert', 'Anna')
  await screen.findByRole('heading', { level: 1, name: eigene.tischTitel })

  // **Am Tisch wird nicht gelesen** — die Liste ist nicht zu sehen.
  expect(gelesen).toBe(1)

  await hand.click(screen.getByRole('button', { name: eigene.verlassen }))
  await screen.findByRole('heading', { level: 1, name: eigene.anfangTitel })

  await waitFor(() => {
    expect(screen.queryByText(eigene.partienDran)).not.toBeInTheDocument()
  })

  expect(gelesen).toBe(2)

  await antwortAbwarten()
})
