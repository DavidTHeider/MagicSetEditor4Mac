// **Ein Handel in vier Phasen** — TRADE-02
// (#1716).
//
// # Wogegen hier geprüft wird
//
// **AK-01**: die dreizehn Zustände erscheinen als vier Phasen; **AK-02**: die
// Lage sagt, wer dran ist, und es stehen nur die Schritte da, die gehen;
// **AK-03**: Zurück kostet den Entwurf nicht.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { zaehle } from '../test/bedienzaehlung'
import { kataloge } from '../sprache/kataloge'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import {
  bestaetigt,
  kennungsfeld,
  oeffneZeile,
  zeigeHandel,
  zuPhase,
} from './tauschprobe'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

// mitHandel stellt einen Handel mit dem Gegenüber bereit.
function mitHandel(
  vorgang: Handelsvorgang,
  mehr: Parameters<typeof belegeApi>[0] = {},
) {
  belegeApi(
    mitStandnachlauf({
      ...mehr,
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
    }),
  )
}

// oeffne wählt den Handel in „Mit wem".
async function oeffne(hand: Awaited<ReturnType<typeof zeigeHandel>>) {
  await zuPhase(hand, 'gegenueber')
  await hand.click(await screen.findByRole('radio', { name: /Revision 1/ }))
}

it('steht in der Phase des Handels und zeigt nur, was geht', async () => {
  mitHandel(
    einVorgang({
      zustand: 'angenommen',
      phase: 'zusagen',
      naechster: 'vorbereiten',
      moegliche: ['vorbereiten', 'abbrechen'],
    }),
  )

  const hand = await zeigeHandel()
  await oeffne(hand)

  // **Die Ansicht springt zur Phase des gewählten Handels.**
  expect(
    await screen.findByRole('listitem', { current: 'step' }),
  ).toHaveTextContent(eigene.assistent.phasen.zusagen)
  expect(
    screen.getByRole('region', { name: eigene.zusagenTitel }),
  ).toBeInTheDocument()

  expect(screen.getByText('Du bist dran: vorbereiten')).toBeInTheDocument()

  const schritte = screen.getByRole('group', {
    name: eigene.assistent.schritte,
  })
  expect(
    within(schritte)
      .getAllByRole('button')
      .map((knopf) => knopf.textContent),
    // **Der Hauptknopf sagt, was sein Zug geht** (P13, Vermerk 0200).
  ).toEqual([eigene.zug.vorbereitenUebergeben, eigene.schrittAbbrechen])

  // **Eine erledigte Phase heißt so** — für ein Vorleseprogramm.
  expect(
    screen.getByRole('button', {
      name: new RegExp(
        `${eigene.assistent.phasen.angebot}.*${eigene.assistent.erledigt}`,
      ),
    }),
  ).toBeInTheDocument()
})

it('sagt, wenn nichts zu tun ist, und auf wen gewartet wird', async () => {
  mitHandel(einVorgang({ lage: 'wartet', naechster: '', moegliche: [] }))

  const hand = await zeigeHandel()
  await oeffne(hand)

  expect(
    await screen.findByText(eigene.assistent.keinSchritt),
  ).toBeInTheDocument()
  // **Beim Namen** — seit UXR-15 (#1923) und nicht mehr „dein Gegenüber".
  expect(screen.getByText('Wartet auf Fremder Rechner')).toBeInTheDocument()
})

it('zeigt die Schritte des Abschlusses nicht als Knopf', async () => {
  mitHandel(
    einVorgang({
      zustand: 'uebergabeOffen',
      phase: 'verbuchen',
      naechster: 'verbuchen',
      moegliche: ['verbuchen', 'klaeren'],
    }),
  )

  const hand = await zeigeHandel()
  await oeffne(hand)

  // **Verbuchen fragt nach einer Freigabe** und steht in seiner Tafel; hier
  // bleibt nur, was die Leitung geht.
  const schritte = await screen.findByRole('group', {
    name: eigene.assistent.schritte,
  })
  expect(
    within(schritte)
      .getAllByRole('button')
      .map((knopf) => knopf.textContent),
  ).toEqual([eigene.schrittKlaeren])
  expect(
    screen.getByRole('region', { name: eigene.buchungTitel }),
  ).toBeInTheDocument()
})

it('verliert den Entwurf nicht, wenn jemand zurückgeht', async () => {
  mitHandel(einVorgang())

  const hand = await zeigeHandel()

  // **Ein neuer Handel steht in „Was tauscht ihr"**, und dort wird
  // zusammengestellt.
  // **Die Spalte des Gegenübers** — ohne Ausweis bleibt die eigene zu.
  await oeffneZeile(hand, 'gegenueber')
  const feld = await kennungsfeld(hand)
  await hand.click(feld)
  await hand.paste('bolt-m10')

  await zuPhase(hand, 'gegenueber')
  // **Verborgen und nicht ausgehängt** — sonst wäre der Entwurf fort.
  expect(screen.getByLabelText(eigene.positionPrinting)).not.toBeVisible()

  await zuPhase(hand, 'angebot')
  expect(screen.getByLabelText(eigene.positionPrinting)).toHaveValue('bolt-m10')
})

it('zeigt einen beendeten Handel in „Was tauscht ihr"', async () => {
  mitHandel(
    einVorgang({
      zustand: 'abgebrochen',
      phase: 'ende',
      lage: 'abgebrochen',
      naechster: '',
      moegliche: [],
      ende: true,
    }),
  )

  const hand = await zeigeHandel()
  await oeffne(hand)

  // **Ein Ende ist keine fünfte Phase** — gezeigt wird, was hin und her
  // ging, und die Lage sagt, dass es vorbei ist.
  expect(
    await screen.findByRole('listitem', { current: 'step' }),
  ).toHaveTextContent(eigene.assistent.phasen.angebot)
  expect(
    screen.getByText(new RegExp(`^Beendet: ${eigene.buchungStandAbgebrochen}`)),
  ).toBeInTheDocument()
})

it('meldet einen gescheiterten Schritt an den Schritten', async () => {
  mitHandel(einVorgang(), {
    HandelSchritt: () => Promise.reject(new Error('geht nicht')),
  })

  const hand = await zeigeHandel()
  await oeffne(hand)

  const schritte = await screen.findByRole('group', {
    name: eigene.assistent.schritte,
  })
  await hand.click(
    within(schritte).getByRole('button', { name: eigene.schrittZustimmen }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent('geht nicht')
})

// **Ein Hauptknopf je Bereich** — R1 in `docs/UIREVIEW.md` (UXR-15, #1923).
// Die Schrittleiste zeigte die aufgeschlagene Phase als gedrückten Knopf,
// und ein gedrückter Knopf trägt das Hauptgewicht: neben „Zustimmen" standen
// zwei gefüllte Knöpfe im selben Bereich. Jetzt sieht die gewählte Phase aus
// wie eine gewählte Lasche.
it('führt mit dem nächsten Schritt und nicht mit der Schrittleiste', async () => {
  mitHandel(einVorgang())

  const hand = await zeigeHandel()
  await oeffne(hand)
  await zuPhase(hand, 'angebot')

  const leiste = await screen.findByRole('navigation', {
    name: eigene.assistent.schrittleiste,
  })
  const gewaehlt = within(leiste).getByRole('button', { pressed: true })
  expect(gewaehlt).toHaveTextContent(eigene.assistent.phasen.angebot)
  expect(gewaehlt).not.toHaveClass('bg-marke')

  // **Gezählt um den Assistenten herum** — er selbst ist der Bereich.
  const raum = leiste.closest('section')?.parentElement
  if (raum === null || raum === undefined) {
    throw new Error('der Raum fehlt')
  }
  expect(zaehle(raum).hauptJeRegion).toBe(1)
})
