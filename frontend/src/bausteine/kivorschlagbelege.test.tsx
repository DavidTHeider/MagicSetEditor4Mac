// **Ein Vorschlag mit Belegen** — Werkstück A13.4 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass die Belegfrage wahlfrei ist und leer nichts auslöst.** Der
// gewöhnliche Vorschlag fragt keinen Index — wer trotzdem suchte, zahlte
// Zeit für nichts.
//
// **Und dass das Kästchen nur dasteht, wo es wirkt**: ohne Belegfrage gibt es
// keine Suche, und eine Freigabe dafür bedeutete nichts.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  einKatalog,
  einVorschlag,
  vorschlagstexte as eigene,
  zeigeVorschlagskasten as zeige,
} from '../test/kivorschlag'
import { kataloge } from '../sprache/kataloge'

// **Die Freigabe steht im Recherchekatalog** — sie gehört zur Suche und nicht
// zum Vorschlag, und derselbe Satz steht an drei Stellen im Fenster.
const recherche = kataloge.de.seiten.recherche

afterEach(() => {
  raeumeApi()
})

// **Ohne Belegfrage geht sie leer hinaus — und das Kästchen fehlt.**
it('fragt ohne Belegfrage keinen Index', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: vorschlagen,
  })

  zeige()

  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiVorschlagTitel))

  await screen.findByRole('option', { name: 'ollama/gemma4:8b' })

  // **Erst der Beleg, dass der Kasten offen ist** — die Verneinung darunter
  // träfe sonst auch zu, wenn gar nichts dastünde (§3).
  expect(
    screen.getByLabelText(eigene.kiVorschlagBelegfrage),
  ).toBeInTheDocument()

  expect(
    screen.queryByLabelText(recherche.freigabeEigeneKarten),
  ).not.toBeInTheDocument()

  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiVorschlagAuftrag), 'Schreibe.')
  await hand.click(screen.getByRole('button', { name: eigene.kiVorschlagen }))

  await screen.findByText('Blitzschlag')

  expect(vorschlagen).toHaveBeenCalledWith('4', {
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    auftrag: 'Schreibe.',
    nachweis: '',
    belegfrage: '',
    freigegeben: [],
  })
})

// **Mit Belegfrage kommt das Kästchen — und beides reist mit.**
it('schickt Belegfrage und Freigabe mit', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: vorschlagen,
  })

  zeige()

  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiVorschlagTitel))

  await screen.findByRole('option', { name: 'ollama/gemma4:8b' })

  await hand.selectOptions(
    screen.getByLabelText(eigene.kiVorschlagAnbieter),
    'ollama/gemma4:8b',
  )
  await hand.type(screen.getByLabelText(eigene.kiVorschlagAuftrag), 'Schreibe.')
  await hand.type(
    screen.getByLabelText(eigene.kiVorschlagBelegfrage),
    'trampelschaden',
  )

  // **Erst jetzt steht das Kästchen da** — das ist die Aussage dieses Laufs.
  await hand.click(screen.getByLabelText(recherche.freigabeEigeneKarten))

  await hand.click(screen.getByRole('button', { name: eigene.kiVorschlagen }))

  await screen.findByText('Blitzschlag')

  expect(vorschlagen).toHaveBeenCalledWith('4', {
    anbieter: 'ollama',
    alias: 'gemma4:8b',
    auftrag: 'Schreibe.',
    nachweis: '',
    belegfrage: 'trampelschaden',
    freigegeben: ['eigenekarte'],
  })
})
