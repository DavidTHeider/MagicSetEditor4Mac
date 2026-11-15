// **Wie ein Handel dorthin kam** — TRADE-03.AK-03
// (#1717).
//
// # Wogegen hier geprüft wird
//
// Der Kasten „Verlauf" in *Verbuchen* zeigt das Protokoll des gewählten
// Handels — jeden Wechsel mit Revision und Zeit —, sagt, wenn es noch keinen
// gibt oder kein Handel gewählt ist, und meldet einen Fehler an sich selbst.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import { bestaetigt, zeigeHandel, zuPhase } from './tauschprobe'

import type { Handelsprotokollantwort } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

// mitProtokoll stellt einen Handel bereit, dessen Protokoll `antwort` gibt.
function mitProtokoll(antwort: () => Promise<Handelsprotokollantwort>) {
  const gefragt = vi.fn(antwort)

  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [einVorgang()], fehler: null }),
    HandelProtokoll: gefragt,
  })

  return gefragt
}

// oeffneVerlauf wählt den Handel und klappt in „Verbuchen" den Verlauf auf.
async function oeffneVerlauf() {
  const hand = await zeigeHandel()

  await zuPhase(hand, 'gegenueber')
  await hand.click(await screen.findByRole('radio', { name: /Revision 1/ }))
  await zuPhase(hand, 'verbuchen')
  await hand.click(
    await screen.findByRole('button', { name: eigene.assistent.verlauf }),
  )
}

it('zeigt jeden Wechsel mit seiner Revision', async () => {
  const gefragt = mitProtokoll(() =>
    Promise.resolve({
      zeilen: [
        {
          revision: 1,
          von: '',
          nach: 'entwurf',
          zeit: '2026-09-30T08:00:00.000Z',
        },
        {
          revision: 1,
          von: 'entwurf',
          nach: 'angeboten',
          zeit: '2026-09-30T08:05:00.000Z',
        },
      ],
      fehler: null,
    }),
  )

  await oeffneVerlauf()

  const zeilen = within(
    await screen.findByRole('list', { name: eigene.assistent.verlauf }),
  ).getAllByRole('listitem')
  expect(zeilen[0]).toHaveTextContent(
    new RegExp(`^Revision 1: ${eigene.buchungStandEntwurf} · `),
  )
  expect(zeilen[1]).toHaveTextContent(
    new RegExp(
      `^Revision 1: ${eigene.buchungStandEntwurf} → ${eigene.buchungStandAngeboten} · `,
    ),
  )
  // **Die Zeit kommt lesbar an**, nicht als ISO-Angabe.
  expect(zeilen[1]).not.toHaveTextContent('2026-09-30T08:05')
  expect(gefragt).toHaveBeenCalledWith('h-1')
})

it('sagt, wenn es noch keinen Wechsel gibt', async () => {
  mitProtokoll(() => Promise.resolve({ zeilen: [], fehler: null }))

  await oeffneVerlauf()

  expect(
    await screen.findByText(eigene.assistent.verlaufLeer),
  ).toBeInTheDocument()
})

it('meldet einen Fehler am Verlauf', async () => {
  mitProtokoll(() =>
    Promise.resolve({
      zeilen: [],
      fehler: {
        code: 'nichtGefunden',
        schluessel: '',
        meldung: 'Den Handel gibt es nicht.',
      },
    }),
  )

  await oeffneVerlauf()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Den Handel gibt es nicht.',
  )
})

it('fragt nicht, solange kein Handel gewählt ist', async () => {
  const gefragt = mitProtokoll(() =>
    Promise.resolve({ zeilen: [], fehler: null }),
  )

  const hand = await zeigeHandel()
  await zuPhase(hand, 'verbuchen')
  await hand.click(
    await screen.findByRole('button', { name: eigene.assistent.verlauf }),
  )

  expect(
    await screen.findByText(eigene.assistent.verlaufOhneHandel),
  ).toBeInTheDocument()
  expect(gefragt).not.toHaveBeenCalled()
})
