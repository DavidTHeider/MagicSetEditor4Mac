// **Was zu einem Gegenüber noch unterwegs ist** — Werkstück T5 von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass „nichts steht aus" dasteht.** Ein Kasten, der nur bei einem Stau
// erscheint, lässt im ruhigen Fall offen, ob gerade alles angekommen ist oder
// ob niemand nachgesehen hat.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { oeffneUnterwegs, zeigeHandel } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// einStand baut die Antwort.
function einStand(offen: number, hoechsteVersuche: number) {
  return Promise.resolve({
    stand: { kennung: 'bgegenueber', offen, hoechsteVersuche },
    fehler: null,
  })
}

describe('Was zu einem Gegenüber noch unterwegs ist', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sagt es, wenn nichts unterwegs ist', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelspoststand: () => einStand(0, 0),
    })

    await oeffneUnterwegs(await zeigeHandel())

    expect(await screen.findByText(eigene.postLeer)).toBeInTheDocument()
  })

  it('nennt die offenen und die hartnäckigste', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelspoststand: () => einStand(2, 3),
    })

    await oeffneUnterwegs(await zeigeHandel())

    // **Die höchste Versuchszahl und nicht die Summe** — sie sagt, ob eine
    // klemmt.
    expect(
      await screen.findByText(/Unterwegs: 2 · Versuche: 3/),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler statt einer leeren Zahl', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelspoststand: () =>
        Promise.resolve({
          stand: { kennung: 'bgegenueber', offen: 0, hoechsteVersuche: 0 },
          fehler: {
            code: 'validierung' as const,
            schluessel: 'gibtesnicht',
            meldung: 'so nicht',
          },
        }),
    })

    await oeffneUnterwegs(await zeigeHandel())

    expect(await screen.findByText('so nicht')).toBeInTheDocument()

    expect(screen.queryByText(eigene.postLeer)).not.toBeInTheDocument()
  })
})
