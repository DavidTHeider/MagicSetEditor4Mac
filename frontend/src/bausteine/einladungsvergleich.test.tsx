// **Einen Code eingeben und vergleichen** — NET-01
// (#1705).
//
// # Wogegen hier geprüft wird
//
// **Die drei Ausgänge des Vergleichs** (AK-04): abbrechen schreibt nichts,
// später vergleichen nimmt unbestätigt auf, „stimmt überein" bestätigt — und
// ein Zweck nur, wenn jemand ihn ausdrücklich anhakt.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'

import {
  alsAntwort,
  derCode,
  eigene,
  einFehler,
  eineLeereListe,
  eineVorschau,
  pruefe,
  stelleAuf,
} from '../test/einladungsprobe'

afterEach(() => {
  raeumeApi()
})

it('bestätigt mit dem Code, den beide sehen', async () => {
  const annehmen = vi.fn(eineLeereListe)
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand, nachAufnahme } = stelleAuf()
  await pruefe(hand)

  // **Der Titel nennt, wen**, und der Code steht zum Vorlesen da.
  expect(
    await screen.findByRole('heading', { name: 'Anna bestätigen' }),
  ).toBeInTheDocument()
  expect(
    screen.getByText('49311 67386 01529 95503 63193 56021'),
  ).toBeInTheDocument()

  // **Der Zweck ist nicht vorgewählt** — eine Zustimmung gibt ein Mensch.
  const haken = screen.getByRole('checkbox')
  expect(haken).not.toBeChecked()

  await hand.clear(screen.getByLabelText(eigene.eigenerName))
  await hand.type(
    screen.getByLabelText(eigene.eigenerName),
    'Anna vom Stammtisch',
  )
  await hand.click(screen.getByRole('button', { name: eigene.stimmt }))

  await waitFor(() => {
    expect(nachAufnahme).toHaveBeenCalled()
  })
  expect(annehmen).toHaveBeenCalledWith(
    derCode,
    'Anna vom Stammtisch',
    '49311 67386 01529 95503 63193 56021',
    false,
    '',
  )
  expect(
    await screen.findByText('Anna vom Stammtisch ist bestätigt.'),
  ).toBeInTheDocument()

  // **Danach ist der Schritt geräumt**: kein Vergleich, leeres Feld.
  expect(screen.queryByRole('button', { name: eigene.stimmt })).toBeNull()
  expect(screen.getByLabelText(eigene.eingabe)).toHaveValue('')
})

it('gibt den Zweck nur mit dem Haken frei', async () => {
  const annehmen = vi.fn(eineLeereListe)
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('checkbox'))
  await hand.click(screen.getByRole('button', { name: eigene.stimmt }))

  await waitFor(() => {
    expect(annehmen).toHaveBeenCalledWith(
      derCode,
      'Anna',
      '49311 67386 01529 95503 63193 56021',
      true,
      '',
    )
  })
})

it('nimmt später auf, ohne zu vergleichen und ohne freizugeben', async () => {
  const annehmen = vi.fn(eineLeereListe)
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  // **Auch mit gesetztem Haken**: ohne Vergleich gibt es keine Freigabe.
  await hand.click(await screen.findByRole('checkbox'))
  await hand.clear(screen.getByLabelText(eigene.eigenerName))
  await hand.click(screen.getByRole('button', { name: eigene.spaeter }))

  await waitFor(() => {
    expect(annehmen).toHaveBeenCalledWith(derCode, '', '', false, '')
  })

  // **Ohne eigenen Namen heißt es, wie es sich nennt.**
  expect(
    await screen.findByText('Anna ist aufgenommen — noch unbestätigt.'),
  ).toBeInTheDocument()
})

it('bricht ab, ohne etwas zu schreiben', async () => {
  const annehmen = vi.fn(eineLeereListe)
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: annehmen,
  })

  const { hand } = stelleAuf()
  await pruefe(hand)

  await hand.click(
    await screen.findByRole('button', { name: eigene.abbrechen }),
  )

  expect(screen.queryByRole('heading', { name: 'Anna bestätigen' })).toBeNull()
  expect(screen.getByLabelText(eigene.eingabe)).toHaveValue('')
  expect(annehmen).not.toHaveBeenCalled()
})

it('zeigt einen Fehler der Annahme am Vergleich und lässt ihn stehen', async () => {
  belegeApi({
    KontakteinladungPruefen: () => alsAntwort(eineVorschau()),
    KontakteinladungAnnehmen: () =>
      Promise.resolve({
        ...einFehler('der verglichene Code stimmt nicht'),
        kontakte: [],
        zwecke: [],
      }),
  })

  const { hand, nachAufnahme } = stelleAuf()
  await pruefe(hand)

  await hand.click(await screen.findByRole('button', { name: eigene.stimmt }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'der verglichene Code stimmt nicht',
  )
  expect(
    screen.getByRole('button', { name: eigene.stimmt }),
  ).toBeInTheDocument()
  expect(nachAufnahme).not.toHaveBeenCalled()
})

it('sagt, wenn die Einladung schon angenommen oder das Gegenüber gesperrt ist', async () => {
  let vorschau = eineVorschau({ angenommen: true })
  belegeApi({ KontakteinladungPruefen: () => alsAntwort(vorschau) })

  const { hand } = stelleAuf()
  await pruefe(hand)

  expect(await screen.findByRole('alert')).toHaveTextContent(eigene.angenommen)
  expect(screen.queryByRole('button', { name: eigene.stimmt })).toBeNull()

  // **Nur der Name darüber** — „Anna bestätigen" widerspräche dem Satz.
  expect(screen.getByRole('heading', { name: 'Anna' })).toBeInTheDocument()

  vorschau = eineVorschau({ lage: 'widerrufen' })
  await hand.click(screen.getByRole('button', { name: eigene.abbrechen }))
  await hand.click(screen.getByLabelText(eigene.eingabe))
  await hand.paste(derCode)
  await hand.click(screen.getByRole('button', { name: eigene.pruefen }))

  expect(await screen.findByRole('alert')).toHaveTextContent(eigene.gesperrt)
  expect(screen.queryByRole('button', { name: eigene.spaeter })).toBeNull()
})
