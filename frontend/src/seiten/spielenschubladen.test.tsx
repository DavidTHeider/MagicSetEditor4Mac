// **Der fremde Tisch und die Leihe hinter Schubladen** — UXR-13
// (#1921).
//
// # Der Anlass
//
// **Unter dem Anfang standen bis hierher drei Abschnitte offen**: „An einem
// fremden Tisch mitspielen" mit der Wahl des Wirts, der Kasten für ein
// Gegenüber und die Leihe mit Entleiher, Deck, Stunden, Ziel und Umschlag.
// Mit echten Mengen (`test/spielbestand.ts`) zeichnete die Seite damit 19
// Felder (R3 lässt zwölf zu) und keinen Knopf, der in diesen Abschnitten
// führte (R1). Das Artboard *Eine Partie anfangen* führt beide nicht; sie
// gehören auf diese Seite, aber nicht neben jeden Anfang.
//
// # Wogegen hier geprüft wird
//
//   - **Ohne Klick steht keines der beiden Formulare da** — und die Seite
//     zeichnet mit echten Mengen höchstens zwölf Felder.
//   - **Jede Schublade trägt ihren Namen, führt mit einem Hauptknopf und
//     bringt den Kasten für ein Gegenüber mit**: unter ihr ist die Seite
//     gesperrt, und ohne ihn käme niemand an ein bestätigtes Gegenüber.
//   - **Wer sich an einen fremden Tisch setzt, sieht ihn**: die Schublade
//     geht zu, und der Tisch steht auf der Seite.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { zaehle } from '../test/bedienzaehlung'
import { kataloge } from '../sprache/kataloge'
import { spielbestand } from '../test/spielbestand'
import { derWirt } from './partiegastprobe'
import { eineListe } from './peerkontaktprobe'
import { mitDecks, zeige } from './spielenprobe'

const texte = kataloge.de
const eigene = texte.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// grenze ist R3 aus `docs/UIREVIEW.md`.
const grenze = 12

it('zeigt ohne Klick weder die Leihe noch den Beitritt', async () => {
  mitDecks(spielbestand())

  zeige()

  await screen.findByLabelText(`${eigene.sitzName} 1`)
  await screen.findAllByText(/Ada · Bo/)

  expect(screen.queryByLabelText(eigene.leiheStunden)).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.gastBeitreten }),
  ).not.toBeInTheDocument()
  expect(zaehle(document.body).felder).toBeLessThanOrEqual(grenze)
})

it('öffnet die Leihe in ihrer Schublade', async () => {
  mitDecks()

  const hand = zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.leiheOeffnen }),
  )

  const schublade = await screen.findByRole('dialog', { name: eigene.leihe })

  expect(
    within(schublade).getByRole('button', { name: eigene.leiheAnbieten }),
  ).toHaveClass('bg-marke')
  expect(
    within(schublade).getByRole('button', {
      name: texte.bausteine.gegenueberTitel,
    }),
  ).toBeInTheDocument()
  expect(await verstoesse(schublade)).toEqual([])

  await userEvent.keyboard('{Escape}')
  await waitFor(() => {
    expect(screen.queryByRole('dialog')).toBeNull()
  })
})

it('öffnet den Beitritt in seiner Schublade', async () => {
  mitDecks({ Peerkontakte: () => Promise.resolve(eineListe([derWirt])) })

  const hand = zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.gastOeffnen }),
  )

  const schublade = await screen.findByRole('dialog', { name: eigene.gast })

  expect(
    await within(schublade).findByRole('radio', { name: /^Wirt · / }),
  ).toBeInTheDocument()
  expect(
    within(schublade).getByRole('button', { name: eigene.gastBeitreten }),
  ).toHaveClass('bg-marke')
  expect(
    within(schublade).getByRole('button', {
      name: texte.bausteine.gegenueberTitel,
    }),
  ).toBeInTheDocument()
  expect(await verstoesse(schublade)).toEqual([])
})

it('schließt den Beitritt, sobald man sitzt, und zeigt den Platz auf der Seite', async () => {
  const beigetreten = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: '',
    gegenueber: 'abcd-ef01',
    fehler: null,
  })
  mitDecks({
    Peerkontakte: () => Promise.resolve(eineListe([derWirt])),
    PartieBeitreten: beigetreten,
    PartieBeitrittBeenden: () =>
      Promise.resolve({
        laeuft: false,
        partie: '',
        gegenueber: '',
        fehler: null,
      }),
  })

  const hand = zeige()

  await hand.click(
    await screen.findByRole('button', { name: eigene.gastOeffnen }),
  )
  const schublade = await screen.findByRole('dialog', { name: eigene.gast })
  await hand.click(
    await within(schublade).findByRole('radio', { name: /^Wirt · / }),
  )
  await hand.click(
    within(schublade).getByRole('button', { name: eigene.gastBeitreten }),
  )

  await waitFor(() => {
    expect(screen.queryByRole('dialog')).toBeNull()
  })
  // **Mit dem vorgewählten eigenen Deck** (P9c, #2377) — hier das erste.
  expect(beigetreten).toHaveBeenCalledWith('bwirt', '', {
    deckID: '1',
    leiheID: '',
  })

  const platz = screen.getByRole('region', { name: eigene.gastPlatz })

  expect(
    within(platz).getByText(eigene.gastSitzt, { exact: false }),
  ).toBeInTheDocument()
  expect(
    within(platz).getByRole('button', { name: eigene.gastAufstehen }),
  ).toBeInTheDocument()

  // **Wer sitzt, setzt sich nicht noch einmal** — der Knopf im Kopf ist
  // fort, bis er aufsteht.
  expect(
    screen.queryByRole('button', { name: eigene.gastOeffnen }),
  ).not.toBeInTheDocument()

  // **Und wer aufsteht, kann sich wieder setzen** — der Platz geht, der
  // Knopf kommt, und die Schublade bleibt zu.
  await hand.click(
    within(platz).getByRole('button', { name: eigene.gastAufstehen }),
  )

  expect(
    await screen.findByRole('button', { name: eigene.gastOeffnen }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('region', { name: eigene.gastPlatz }),
  ).not.toBeInTheDocument()
  expect(screen.queryByRole('dialog')).toBeNull()
})
