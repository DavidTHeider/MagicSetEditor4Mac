// **Was zu einem Gegenüber hinausgeht** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass der Stand in den Ausgang geht und seine Kennung zurückkommt** — an
// ihr räumt die Bestätigung die Zeile wieder weg.
//
// **Dass eine Klärung den Rest verdeckt.** Stehen zwei Verläufe da, liefert
// niemand nach; eine Liste daneben sähe aus, als ginge es weiter.
//
// **Und dass jedes Feld seinen eigenen Satz bekommt**, wenn es leer ist — der
// Fehler erscheint am verursachenden Feld und nicht als anonymer Hinweis
// (§6).
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { oeffneHandel, zeigeHandel, zuPhase } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

import type { Wiederaufnahmeantwort } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// imAusgang gibt den Kasten.
function imAusgang() {
  return screen.getByRole('region', { name: eigene.ausgangTitel })
}

// waehleUndDruecke wählt den Handel — leer heißt: keinen — und drückt einen
// Knopf des Kastens.
async function waehleUndDruecke(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  vorgang: string,
  knopf: string,
) {
  if (vorgang !== '') {
    await oeffneHandel(hand, vorgang)
  }

  // **Der Ausgang steht in „Verbuchen"** (TRADE-02, #1716).
  await zuPhase(hand, 'verbuchen')

  const kasten = await screen.findByRole('region', {
    name: eigene.ausgangTitel,
  })

  await hand.click(within(kasten).getByRole('button', { name: knopf }))
}

// raeumeVonHand räumt eine Nachricht über ihre Kennung weg — der Weg für
// Fachleute (UX-08.AK-04).
async function raeumeVonHand(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  nachricht: string,
) {
  await zuPhase(hand, 'verbuchen')

  const kasten = await screen.findByRole('region', {
    name: eigene.ausgangTitel,
  })
  const fachleute = within(kasten).getByRole('button', {
    name: eigene.ausgangFachleute,
  })

  if (fachleute.getAttribute('aria-expanded') !== 'true') {
    await hand.click(fachleute)
  }

  if (nachricht !== '') {
    await hand.click(within(kasten).getByLabelText(eigene.ausgangNachricht))
    await hand.paste(nachricht)
  }

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.ausgangBestaetigen }),
  )
}

// eineWiederaufnahme baut die Antwort des Ausgangs.
function eineWiederaufnahme(
  anders: Partial<Wiederaufnahmeantwort> = {},
): Wiederaufnahmeantwort {
  return { nachzuliefern: [], klaerung: false, fehler: null, ...anders }
}

describe('Was zu einem Gegenüber hinausgeht', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('legt den Stand in den Ausgang und nennt die Nachricht', async () => {
    const geschickt = vi.fn(() =>
      Promise.resolve({ nachrichtId: 'n-7', fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelSchicken: geschickt,
    })

    await waehleUndDruecke(await zeigeHandel(), 'v-1', eigene.ausgangSchicken)

    expect(geschickt).toHaveBeenCalledWith('v-1', 'bgegenueber')
    expect(await within(imAusgang()).findByText('n-7')).toBeInTheDocument()
  })

  // **Und es geht wirklich hinaus** —
  // #1634.
  //
  // **Bis dahin konnte dieses Fenster einen Handel annehmen und keinen
  // hinausbringen**: `HandelSchicken` legte in den Ausgang, und getragen hat
  // das niemand.
  it('trägt den Ausgang über die Leitung und nennt die Zahl', async () => {
    const getragen = vi.fn(() =>
      Promise.resolve({ gegenueber: 'bgegenueber', getragen: 2, fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelZustellen: getragen,
    })

    await waehleUndDruecke(await zeigeHandel(), '', eigene.ausgangZustellen)

    expect(getragen).toHaveBeenCalledWith('bgegenueber')

    expect(
      await within(imAusgang()).findByText(
        eigene.ausgangGetragen.replace('{zahl}', '2'),
      ),
    ).toBeInTheDocument()
  })

  it('räumt eine angekommene Nachricht weg', async () => {
    const bestaetigt_ = vi.fn(() =>
      Promise.resolve({ nachrichtId: 'n-7', fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelPostbestaetigung: bestaetigt_,
    })

    await raeumeVonHand(await zeigeHandel(), 'n-7')

    expect(bestaetigt_).toHaveBeenCalledWith('n-7')
  })

  // **Gewählt und nicht getippt** (UX-08.AK-04): was aussteht, hat seinen
  // eigenen Knopf, und die Kennung der Nachricht tippt niemand.
  it('räumt eine ausstehende Nachricht mit ihrem Knopf weg', async () => {
    const weggeraeumt = vi.fn(() =>
      Promise.resolve({ nachrichtId: 'n-7', fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelPostbestaetigung: weggeraeumt,
      HandelWiederaufnahme: () =>
        Promise.resolve(
          eineWiederaufnahme({
            nachzuliefern: [
              {
                nachrichtId: 'n-7',
                vorgang: 'v-1',
                empfaenger: 'bgegenueber',
                folge: 3,
                art: 'vorschlag',
              },
            ],
          }),
        ),
    })

    const hand = await zeigeHandel()

    await waehleUndDruecke(hand, 'v-1', eigene.ausgangWiederaufnahme)

    await hand.click(
      await within(imAusgang()).findByRole('button', {
        name: eigene.ausgangWegraeumenName.replace('{nachricht}', 'n-7'),
      }),
    )

    expect(weggeraeumt).toHaveBeenCalledWith('n-7')
    expect(within(imAusgang()).queryByText(/n-7 ·/)).not.toBeInTheDocument()
  })

  it('zeigt, was noch einmal hinausmuss', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelWiederaufnahme: () =>
        Promise.resolve(
          eineWiederaufnahme({
            nachzuliefern: [
              {
                nachrichtId: 'n-7',
                vorgang: 'v-1',
                empfaenger: 'bgegenueber',
                folge: 3,
                art: 'vorschlag',
              },
            ],
          }),
        ),
    })

    await waehleUndDruecke(
      await zeigeHandel(),
      'v-1',
      eigene.ausgangWiederaufnahme,
    )

    const kasten = imAusgang()

    expect(
      await within(kasten).findByText(eigene.ausgangNachzuliefern),
    ).toBeInTheDocument()
    expect(within(kasten).getByText(/n-7/)).toBeInTheDocument()
  })

  it('sagt, wenn nichts im Ausgang liegt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelWiederaufnahme: () => Promise.resolve(eineWiederaufnahme()),
    })

    await waehleUndDruecke(
      await zeigeHandel(),
      'v-1',
      eigene.ausgangWiederaufnahme,
    )

    expect(
      await within(imAusgang()).findByText(eigene.ausgangLeer),
    ).toBeInTheDocument()
  })

  it('verdeckt bei einer Klärung, was sonst nachzuliefern wäre', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelWiederaufnahme: () =>
        Promise.resolve(eineWiederaufnahme({ klaerung: true })),
    })

    await waehleUndDruecke(
      await zeigeHandel(),
      'v-1',
      eigene.ausgangWiederaufnahme,
    )

    const kasten = imAusgang()

    expect(
      await within(kasten).findByText(eigene.ausgangKlaerung),
    ).toBeInTheDocument()
    expect(
      within(kasten).queryByText(eigene.ausgangLeer),
    ).not.toBeInTheDocument()
  })

  it('nennt je Feld, was fehlt — und ruft nichts', async () => {
    const geschickt = vi.fn(() =>
      Promise.resolve({ nachrichtId: '', fehler: null }),
    )
    const weggeraeumt = vi.fn(() =>
      Promise.resolve({ nachrichtId: '', fehler: null }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelSchicken: geschickt,
      HandelPostbestaetigung: weggeraeumt,
    })

    const hand = await zeigeHandel()

    await waehleUndDruecke(hand, '', eigene.ausgangSchicken)

    expect(await within(imAusgang()).findByRole('alert')).toHaveTextContent(
      eigene.ausgangOhneVorgang,
    )

    await raeumeVonHand(hand, '')

    expect(await within(imAusgang()).findByRole('alert')).toHaveTextContent(
      eigene.ausgangOhneNachricht,
    )

    expect(geschickt).not.toHaveBeenCalled()
    expect(weggeraeumt).not.toHaveBeenCalled()
  })

  it('meldet eine Absage der Naht', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelSchicken: () => Promise.reject(new Error('geht nicht')),
    })

    await waehleUndDruecke(await zeigeHandel(), 'v-1', eigene.ausgangSchicken)

    expect(await within(imAusgang()).findByRole('alert')).toBeInTheDocument()
  })
})
