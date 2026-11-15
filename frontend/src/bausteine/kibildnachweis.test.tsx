// **Der Nachweis eines übernommenen Bildes** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass dasteht, was feststeht — und sonst nichts.** Ein Bild von der
// Platte hat keine Angaben; „keine Angaben" wäre eine Auskunft über nichts.
//
// **Und dass zugeklappt nicht gefragt wird.** Der Nachweis liest den
// Bestand; das ist billig, aber es ist ein Ruf, den niemand ausgelöst hat.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  aufschlagenUndErzeugen,
  bildtexte as eigene,
  einKatalog,
  mitFreigabe,
  einVorschlag,
  ohneNachweis,
  zeigeBildkasten as zeige,
} from '../test/kibild'

afterEach(() => {
  raeumeApi()
})

// **Zugeklappt wird nicht nach dem Nachweis gefragt.**
it('zugeklappt wird nicht nach dem Nachweis gefragt', () => {
  const gefragt = vi.fn()

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: gefragt,
    }),
  )

  zeige()

  expect(gefragt).not.toHaveBeenCalled()
})

// **Aufgeschlagen steht da, was über das Bild im Feld feststeht.**
it('aufgeschlagen steht der Nachweis da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () =>
        Promise.resolve({
          angaben: [
            { name: 'guete', wert: 'high' },
            { name: 'seed', wert: '4711' },
          ],
          fehler: null,
        }),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  // **Über die Rolle und nicht über `closest`**: ein `section` mit
  // Überschrift ist eine `region`, und die Abfrage sagt, was gemeint ist.
  const kasten = within(
    await screen.findByRole('region', { name: eigene.kiBildNachweisTitel }),
  )

  // **Im Kasten gesucht und nicht auf der Seite**: »Güte« steht auch als
  // Beschriftung der Gütewahl da, und `getByText` fände beide.
  expect(kasten.getByText(eigene.kiBildAngabeGuete)).toBeInTheDocument()
  expect(kasten.getByText('high')).toBeInTheDocument()

  // **Und eine Angabe, die der Katalog nicht kennt, steht roh da.**
  expect(kasten.getByText('seed')).toBeInTheDocument()
  expect(kasten.getByText('4711')).toBeInTheDocument()
})

// **Ohne Angaben steht kein Kasten da.**
it('ohne Angaben steht kein Nachweis da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await screen.findByLabelText(eigene.kiVorschlagAnbieter)

  expect(screen.queryByText(eigene.kiBildNachweisTitel)).not.toBeInTheDocument()
})

// **Ein Fehler steht als Fehler da.**
it('ein Fehler beim Nachweis steht als Fehler da', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () =>
        Promise.reject(new Error('Der Bestand antwortet nicht.')),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Bestand antwortet nicht.',
  )
})

// **Nach einer Übernahme wird neu gelesen.**
//
// # Warum der Zähler von außen kommt
//
// **Weil nur der Editor weiß, wann eingesetzt wurde.** Der Kasten öffnet den
// Zuschnitt; was danach geschieht, entscheidet sich in [useBildfeld]. Ein
// Zähler in diesem Bauteil zeigte nach der Übernahme den Stand von vorhin.
it('ein neuer Stand liest den Nachweis neu', async () => {
  const gefragt = vi.fn(() => Promise.resolve(ohneNachweis()))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: gefragt,
      KIBildVorschlagen: () => Promise.resolve(einVorschlag()),
    }),
  )

  const { hand, neuZeichnen } = zeige()

  await aufschlagenUndErzeugen(hand)

  expect(gefragt).toHaveBeenCalledTimes(1)

  neuZeichnen(1)

  await screen.findByRole('figure')

  expect(gefragt).toHaveBeenCalledTimes(2)
})
