// **Der Vergleichscode in der Liste** — NET-01
// (#1705).
//
// # Wogegen hier geprüft wird
//
// **Dass „später vergleichen" ein Später hat**: wer ohne Vergleich
// aufgenommen hat, bestätigt hier mit demselben Code. **Und dass ein neuer
// Schlüssel als eigener Zustand dasteht** (AK-03).
import { screen, waitFor, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { kataloge } from '../sprache/kataloge'

import { einKontakt, eineListe, zeige } from './peerkontaktprobe'

const technik = kataloge.de.seiten.technik
const eigene = technik.einladung

afterEach(() => {
  raeumeApi()
})

it('bestätigt mit dem Code, den beide sehen', async () => {
  const bestaetigen = vi.fn(() => Promise.resolve(eineListe()))
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            lage: 'schluesselGeaendert',
            vergleichscode: '11111 22222 33333 44444 55555 66666',
          }),
        ]),
      ),
    PeerkontaktBestaetigen: bestaetigen,
  })

  const hand = await zeige()
  const zeile = (
    await screen.findByText(eigene.lageSchluesselGeaendert)
  ).closest('li') as HTMLElement

  expect(
    within(zeile).getByText('11111 22222 33333 44444 55555 66666'),
  ).toBeInTheDocument()

  // **Das Feld für den Abdruck steht dahinter** — für Fachleute.
  expect(within(zeile).queryByLabelText(technik.kontakteVergleichen)).toBeNull()
  await hand.click(
    within(zeile).getByRole('button', { name: eigene.zeileVonHand }),
  )
  expect(
    within(zeile).getByLabelText(technik.kontakteVergleichen),
  ).toBeInTheDocument()

  await hand.click(
    within(zeile).getByRole('button', { name: eigene.zeileStimmt }),
  )

  await waitFor(() => {
    expect(bestaetigen).toHaveBeenCalledWith(
      'bgegenueber',
      '11111 22222 33333 44444 55555 66666',
    )
  })
})

it('lädt die Liste nach einer Annahme nach', async () => {
  const liste = vi.fn(() => Promise.resolve(eineListe([])))
  belegeApi({
    Peerkontakte: liste,
    KontakteinladungPruefen: () =>
      Promise.resolve({
        vorschau: {
          kennung: 'k',
          name: 'Anna',
          zweck: 'handel',
          adresse: '',
          ablauf: '2026-10-01T12:00:00Z',
          fingerabdruck: 'ab',
          vergleichscode: '1',
          lage: 'unbekannt',
          vorgaenger: '',
          angenommen: false,
          einloesbar: false,
        },
        fehler: null,
      }),
    KontakteinladungAnnehmen: () => Promise.resolve(eineListe()),
  })

  const hand = await zeige()
  await hand.click(await screen.findByRole('button', { name: eigene.eingeben }))
  await hand.click(screen.getByLabelText(eigene.eingabe))
  await hand.paste('mse4mac-kontakt:1:a.b')
  await hand.click(screen.getByRole('button', { name: eigene.pruefen }))
  await hand.click(await screen.findByRole('button', { name: eigene.spaeter }))

  await waitFor(() => {
    expect(liste).toHaveBeenCalledTimes(2)
  })

  // **Die Meldung übersteht das Nachladen** — gefunden an der Hüllenprobe
  // `net01`: ein Nachladen über eine geänderte Abhängigkeit räumte die
  // Liste als neues Bezugsobjekt, und mit ihr den ganzen Tausch.
  await effekteNachholen()
  expect(
    screen.getByText('Anna ist aufgenommen — noch unbestätigt.'),
  ).toBeInTheDocument()
  expect(screen.getByLabelText(eigene.eingabe)).toBeInTheDocument()
})
