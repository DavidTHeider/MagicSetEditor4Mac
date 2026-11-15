// **Mit diesem Handel: kopieren, Entwurf speichern und löschen, archivieren**
// — TRADE-01.AK-03 (#1715):
// was hinausgegangen ist, wird kopiert oder archiviert, nicht gelöscht.
//
// **Seit UXR-15 in einem Menü** (#1923):
// die Handgriffe standen als Knopfreihe unter den Schritten, gleich unter
// dem Hauptknopf — selten gebraucht und doch neben dem, was als Nächstes
// zu tun ist.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { zuPhase } from './tauschprobe'
import { eineUebersicht, eineZeile, einVorgang } from './handelslisteprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const handeln = kataloge.de.seiten.handeln
const eigene = handeln.liste

// oeffne zeigt die Seite mit einem Handel in der Liste und schlägt ihn auf.
async function oeffne(vorgang: Handelsvorgang, archiviert = false, mehr = {}) {
  const HandelUebersicht = vi.fn(() =>
    eineUebersicht([eineZeile({ archiviert }, vorgang)]),
  )

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    HandelUebersicht,
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
    ...mehr,
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  const hand = userEvent.setup()
  await hand.click(await screen.findByRole('button', { name: /Turfer/ }))

  return { hand, HandelUebersicht }
}

// imMenue öffnet „Mit diesem Handel" und wählt einen Eintrag.
async function imMenue(hand: ReturnType<typeof userEvent.setup>, name: string) {
  await hand.click(await screen.findByRole('button', { name: eigene.aktionen }))
  await hand.click(
    await screen.findByRole('menuitem', { name: new RegExp(name) }),
  )
}

// eintraege nennt, was das Menü anbietet.
async function eintraege(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(await screen.findByRole('button', { name: eigene.aktionen }))

  return within(await screen.findByRole('menu', { name: eigene.aktionen }))
    .getAllByRole('menuitem')
    .map((eintrag) => eintrag.textContent)
}

describe('Die Aktionen an einem Handel', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('kopiert einen Handel als neuen Entwurf und wählt ihn', async () => {
    const kopie = einVorgang({ vorgang: 'h-kopie', zustand: 'entwurf' })
    const HandelKopieren = vi.fn(() =>
      Promise.resolve({ vorgang: kopie, fehler: null }),
    )
    const { hand, HandelUebersicht } = await oeffne(einVorgang(), false, {
      HandelKopieren,
    })

    await imMenue(hand, eigene.kopieren)

    expect(HandelKopieren).toHaveBeenCalledWith('h-1')
    expect(await screen.findByText(eigene.kopiert)).toBeInTheDocument()
    await vi.waitFor(() => {
      expect(HandelUebersicht.mock.calls.length).toBeGreaterThan(1)
    })
  })

  it('meldet, wenn das Kopieren scheitert', async () => {
    const HandelKopieren = () =>
      Promise.resolve({
        vorgang: einVorgang(),
        fehler: { code: 'konflikt', schluessel: '', meldung: 'geht nicht' },
      })
    const { hand } = await oeffne(einVorgang(), false, { HandelKopieren })

    await imMenue(hand, eigene.kopieren)

    expect(await screen.findByText('geht nicht')).toBeInTheDocument()
  })

  it('löscht einen Entwurf erst nach der Rückfrage', async () => {
    const HandelEntwurfLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    const { hand } = await oeffne(einVorgang({ zustand: 'entwurf' }), false, {
      HandelEntwurfLoeschen,
    })

    await imMenue(hand, eigene.entwurfLoeschen)
    await hand.click(
      await within(
        screen.getByRole('group', { name: eigene.entwurfLoeschenFrage }),
      ).findByRole('button', { name: eigene.abbrechen }),
    )
    expect(HandelEntwurfLoeschen).not.toHaveBeenCalled()

    await imMenue(hand, eigene.entwurfLoeschen)
    await hand.click(
      await screen.findByRole('button', {
        name: new RegExp(eigene.entwurfLoeschenJa),
      }),
    )

    await vi.waitFor(() => {
      expect(HandelEntwurfLoeschen).toHaveBeenCalledWith('h-1')
    })
    // **Nach dem Verwerfen ist ein neuer Handel gewählt** — die Wahl steht
    // in „Mit wem" (TRADE-02, #1716).
    await zuPhase(hand, 'gegenueber')
    expect(
      await screen.findByRole('radio', { name: new RegExp(handeln.wahlNeu) }),
    ).toBeChecked()
  })

  it('speichert einen Entwurf, statt ihn vorzuschlagen', async () => {
    const entwurf = einVorgang({ zustand: 'entwurf' })
    const HandelEntwurfAendern = vi.fn(() =>
      Promise.resolve({ vorgang: entwurf, fehler: null }),
    )
    const { hand } = await oeffne(entwurf, false, { HandelEntwurfAendern })

    await hand.click(
      await screen.findByRole('button', { name: eigene.entwurfSpeichern }),
    )

    expect(HandelEntwurfAendern).toHaveBeenCalledWith(
      expect.objectContaining({
        vorgang: 'h-1',
        art: 'tausch',
        gegenueber: ['bgegenueber'],
      }),
    )
    expect(
      screen.queryByRole('button', { name: handeln.vorgangVorschlagen }),
    ).not.toBeInTheDocument()
  })

  it.each([
    [false, eigene.archivieren, true],
    [true, eigene.hervorholen, false],
  ])(
    'legt einen beendeten Handel ins Archiv und holt ihn heraus (archiviert: %s)',
    async (archiviert, knopf, an) => {
      const HandelArchivieren = vi.fn(() => Promise.resolve({ fehler: null }))
      const { hand, HandelUebersicht } = await oeffne(
        einVorgang({ zustand: 'beidseitigVerbucht', ende: true }),
        archiviert,
        { HandelArchivieren },
      )

      expect(
        await screen.findByText(/Ins Archiv geht nur ein Handel/),
      ).toBeInTheDocument()
      await imMenue(hand, knopf)

      expect(HandelArchivieren).toHaveBeenCalledWith('h-1', an)
      await vi.waitFor(() => {
        expect(HandelUebersicht.mock.calls.length).toBeGreaterThan(1)
      })
    },
  )

  it('bietet einem laufenden Handel weder Archiv noch Löschen an', async () => {
    const { hand } = await oeffne(einVorgang())

    // **Kein Handgriff steht offen** — alle drei im Menü.
    expect(
      screen.queryByRole('button', { name: new RegExp(eigene.kopieren) }),
    ).not.toBeInTheDocument()
    expect(await eintraege(hand)).toEqual([
      expect.stringContaining(eigene.kopieren),
    ])
  })

  it('bietet einem Entwurf das Löschen an', async () => {
    const { hand } = await oeffne(einVorgang({ zustand: 'entwurf' }))

    expect(await eintraege(hand)).toEqual([
      expect.stringContaining(eigene.kopieren),
      expect.stringContaining(eigene.entwurfLoeschen),
    ])
  })
})
