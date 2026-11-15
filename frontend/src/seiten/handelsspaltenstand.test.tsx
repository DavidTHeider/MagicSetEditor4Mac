// **Was vereinbart ist, steht da** — TRADE-02b
// (#1951).
//
// # Der Befund
//
// **Ein gewählter Handel zeigte seine Zeilen nicht.** Der Stand nannte
// Zustand, Parteien und Art; was hin und her geht, stand nirgends, und unter
// „Was hin und her geht" wartete ein leeres Formular — auch nach der
// Einigung, wenn kein Vorschlag mehr geht (Sichtprüfung zu TRADE-03 am
// 2026-09-30). Ein Gegenangebot begann mit null Zeilen.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import {
  bestaetigt,
  keinAnhalt,
  kennungsfeld,
  oeffneZeile,
  zeigeHandel,
  zuPhase,
} from './tauschprobe'

import type { Handelsposition, Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

function zeile(von: string, printing: string): Handelsposition {
  return {
    von,
    printing,
    druckId: 0,
    menge: 1,
    guete: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    abbild: '',
    bewertung: 0,
  }
}

const ausweis = () =>
  Promise.resolve({
    stand: 'bereit',
    kennung: 'bich',
    fingerabdruck: 'AB-CD',
    seit: '2026-09-19T08:00:00Z',
    vorher: '',
    vorherFingerabdruck: '',
    befunde: [],
    fehler: null,
  })

// mitHandel stellt einen Handel bereit und wählt ihn in „Mit wem".
async function mitHandel(
  vorgang: Handelsvorgang,
  mehr: Parameters<typeof belegeApi>[0] = {},
) {
  belegeApi({
    ...keinAnhalt,
    ...mehr,
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    Peeridentitaet: ausweis,
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
  })

  const hand = await zeigeHandel()
  await zuPhase(hand, 'gegenueber')
  await hand.click(await screen.findByRole('radio', { name: /Revision 1/ }))
  await zuPhase(hand, 'angebot')

  return hand
}

it('zeigt nach der Einigung die vereinbarten Zeilen, nur zum Lesen', async () => {
  await mitHandel(
    einVorgang({
      zustand: 'angenommen',
      phase: 'zusagen',
      moegliche: ['vorbereiten', 'abbrechen'],
      positionen: [zeile('bich', 'ring'), zeile('bgegenueber', 'goyf')],
    }),
  )

  const raum = await screen.findByRole('region', { name: eigene.vorgangTitel })
  const seine = within(raum).getByRole('region', {
    name: eigene.positionVonGegenueber,
  })

  expect(within(seine).getByRole('listitem')).toHaveTextContent('goyf')
  expect(
    within(
      within(raum).getByRole('region', { name: eigene.positionVonIch }),
    ).getByRole('listitem'),
  ).toHaveTextContent('ring')

  // **Kein Formular und kein Entfernen** — nach der Einigung geht kein
  // Vorschlag mehr, und nichts hier verspricht einen.
  expect(within(raum).queryByLabelText(eigene.positionPrinting)).toBeNull()
  expect(
    within(raum).queryByRole('button', { name: /Zeile hinzufügen/ }),
  ).toBeNull()
  expect(within(raum).queryByText(eigene.positionLeer)).toBeNull()
})

it('beginnt ein Gegenangebot beim letzten Angebot', async () => {
  const vorgeschlagen = vi.fn(() =>
    Promise.resolve({
      vorgang: einVorgang({
        revision: 2,
        abdruck: 'neu',
        zustand: 'verhandlung',
        positionen: [
          zeile('bgegenueber', 'goyf'),
          zeile('bgegenueber', 'bolt'),
        ],
      }),
      fehler: null,
    }),
  )

  const hand = await mitHandel(
    einVorgang({
      initiator: 'bgegenueber',
      moegliche: ['zustimmen', 'gegenangebot', 'ablehnen'],
      positionen: [zeile('bich', 'ring'), zeile('bgegenueber', 'goyf')],
    }),
    { HandelVorschlagen: vorgeschlagen },
  )

  const raum = await screen.findByRole('region', { name: eigene.vorgangTitel })

  // **Die Zeilen stehen da und lassen sich ändern.**
  await hand.click(
    within(raum).getByRole('button', {
      name: eigene.positionEntfernenName.replace('{nummer}', '1'),
    }),
  )
  await hand.click(
    within(raum).getByRole('button', { name: eigene.vorgangVorschlagen }),
  )

  expect(vorgeschlagen).toHaveBeenCalledWith('h-1', [
    expect.objectContaining({ von: 'bgegenueber', printing: 'goyf' }),
  ])

  // **Nach dem Schreiben steht der neue Stand da** — auch eine Zeile, die
  // hier niemand angelegt hat.
  expect(
    await within(
      within(raum).getByRole('region', { name: eigene.positionVonGegenueber }),
    ).findByText(/bolt/),
  ).toBeInTheDocument()
})

it('stellt eine Zeile vom Gegenüber in seine Spalte', async () => {
  belegeApi({
    ...keinAnhalt,
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    Peeridentitaet: ausweis,
  })

  const hand = await zeigeHandel()

  await oeffneZeile(hand, 'gegenueber')

  const raum = screen.getByRole('region', { name: eigene.vorgangTitel })

  await hand.click(await kennungsfeld(hand, raum))
  await hand.paste('shivan-lea')
  await hand.click(
    within(raum).getByRole('button', { name: eigene.positionHinzufuegen }),
  )

  expect(
    within(
      within(raum).getByRole('region', { name: eigene.positionVonGegenueber }),
    ).getByText(/shivan-lea/),
  ).toBeInTheDocument()
})
