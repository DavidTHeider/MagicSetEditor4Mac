// **Ein Code ohne eigene Identität** — G01
// (#2248).
//
// Gefunden am nativen Fenster (`g01fan`): ein Profil ohne Identität bekam
// in der Schublade »Neuer Tausch« erst nach dem Einfügen und Prüfen zu
// lesen, es solle die Identität »im Kasten ‚Womit sich dieses Profil
// ausweist‘« anlegen — der steht nicht in der Schublade, und »Identität
// anlegen« auf der Seite lag hinter ihr. Jetzt steht der Knopf am Fehler,
// und nach dem Anlegen wird derselbe Code noch einmal geprüft.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import {
  alsAntwort,
  derCode,
  eigene,
  eineVorschau,
  pruefe,
  stelleAuf,
} from '../test/einladungsprobe'

import type {
  Einladungsvorschauantwort,
  Kontakteinladungsantwort,
} from '../apikontakteinladung'
import type { Peeridentitaetsantwort } from '../apipeer'

afterEach(() => {
  raeumeApi()
})

const handeln = kataloge.de.seiten.handeln.liste

// derFehler ist, woran ein Schritt ohne eigene Identität scheitert.
const derFehler = {
  code: 'validierung' as const,
  schluessel: 'ohnePeeridentitaet',
  meldung: 'roh',
}

// angelegt ist die Antwort, mit der die Identität entsteht.
const angelegt = (): Promise<Peeridentitaetsantwort> =>
  Promise.resolve({
    stand: 'bereit',
    kennung: '12D3KooWIch',
    fingerabdruck: '1a2b',
    seit: '2026-10-11T00:00:00Z',
    vorher: '',
    vorherFingerabdruck: '',
    befunde: [],
    fehler: null,
  })

it('bietet beim Prüfen die Identität an und prüft danach denselben Code', async () => {
  let da = false
  const pruefen = vi.fn((): Promise<Einladungsvorschauantwort> =>
    da
      ? alsAntwort(eineVorschau())
      : Promise.resolve({ vorschau: eineVorschau(), fehler: derFehler }),
  )
  const anlegen = vi.fn(() => {
    da = true

    return angelegt()
  })
  belegeApi({
    KontakteinladungPruefen: pruefen,
    PeeridentitaetSicherstellen: anlegen,
  })

  const { hand } = stelleAuf()
  await pruefe(hand, derCode)

  await hand.click(
    await screen.findByRole('button', { name: handeln.ohneIdentitaetAnlegen }),
  )

  expect(anlegen).toHaveBeenCalledOnce()
  expect(
    await screen.findByText(eineVorschau().vergleichscode),
  ).toBeInTheDocument()
  expect(pruefen).toHaveBeenCalledTimes(2)
  expect(pruefen).toHaveBeenLastCalledWith(derCode)
})

it('bietet beim Erzeugen die Identität an', async () => {
  belegeApi({
    KontakteinladungErstellen: (): Promise<Kontakteinladungsantwort> =>
      Promise.resolve({
        einladung: { code: '', ablauf: '', fingerabdruck: '' },
        fehler: derFehler,
      }),
    PeeridentitaetSicherstellen: angelegt,
  })

  const { hand } = stelleAuf()
  await hand.click(screen.getByRole('button', { name: eigene.zeigen }))
  await hand.type(screen.getByLabelText(eigene.name), 'Anna')
  await hand.click(screen.getByRole('button', { name: eigene.erzeugen }))

  await hand.click(
    await screen.findByRole('button', { name: handeln.ohneIdentitaetAnlegen }),
  )
  expect(screen.queryByText(kataloge.de.fehler.ohnePeeridentitaet)).toBeNull()

  // **Nach dem Anlegen ist der Fehler fort**, und geprüft wird nichts — zum
  // Erzeugen gibt es keinen Code, und Name und Zweck stehen noch da.
  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: handeln.ohneIdentitaetAnlegen }),
    ).toBeNull()
  })
  expect(screen.getByRole('button', { name: eigene.erzeugen })).toBeEnabled()
})

// **§7, Regel 2** — Befund 6 des Zweitblicks über Phase G, zweiter Teil:
// ändert sich die Eingabe, während die Identität entsteht, gehört das Ende
// des Anlegens zur alten Einladung. Es prüft die neue nicht und räumt ihren
// Fehler nicht.
it('prüft nach dem Anlegen keine Einladung, die inzwischen eine andere ist', async () => {
  let fertig: (antwort: Peeridentitaetsantwort) => void = () => undefined
  const pruefen = vi.fn((): Promise<Einladungsvorschauantwort> =>
    Promise.resolve({ vorschau: eineVorschau(), fehler: derFehler }),
  )
  belegeApi({
    KontakteinladungPruefen: pruefen,
    PeeridentitaetSicherstellen: () =>
      new Promise<Peeridentitaetsantwort>((ja) => {
        fertig = ja
      }),
  })

  const { hand } = stelleAuf()
  await pruefe(hand, derCode)
  await hand.click(
    await screen.findByRole('button', { name: handeln.ohneIdentitaetAnlegen }),
  )

  // **Die Eingabe ändert sich**, während die Identität noch entsteht.
  await hand.type(screen.getByLabelText(eigene.eingabe), 'x')
  fertig(await angelegt())
  await effekteNachholen()

  expect(pruefen).toHaveBeenCalledOnce()
  expect(screen.getByLabelText(eigene.eingabe)).toHaveValue(derCode + 'x')
})
