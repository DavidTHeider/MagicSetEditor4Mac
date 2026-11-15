// **Einen Handel wählen, statt seine Kennung zu tippen** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind; Expertendetails bleiben mit
// verständlicher Erläuterung erreichbar."*
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { zeigeHandel, zuPhase } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

// einHandel ist ein Handel, wie die Liste ihn nennt.
function einHandel(anders: Partial<Handelsvorgang> = {}): Handelsvorgang {
  return {
    vorgang: 'h-1',
    revision: 2,
    eltern: 'ab',
    abdruck: 'abcdef0123456789',
    art: 'tausch',
    initiator: 'bich',
    parteien: ['bich', 'bgegenueber'],
    ich: 'bich',
    zustand: 'angeboten',
    positionen: [],
    bewertungGeteilt: false,
    zustimmungen: [],
    ende: false,
    beschluss: '',
    lage: 'dran',
    naechster: 'zustimmen',
    moegliche: ['zustimmen', 'abbrechen', 'ablaufen'],
    phase: 'angebot',
    ...anders,
  }
}

// mitHandeln belegt die Liste und das Gegenüber.
function mitHandeln(vorgaenge: Handelsvorgang[], mehr = {}) {
  const HandelVorgaenge = vi.fn(() =>
    Promise.resolve({ vorgaenge, fehler: null }),
  )

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    HandelVorgaenge,
    ...mehr,
  })

  return HandelVorgaenge
}

describe('Die Wahl eines Handels', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('nennt die Handel mit dem Gegenüber — und ein neuer ist vorgewählt', async () => {
    const HandelVorgaenge = mitHandeln([
      einHandel(),
      einHandel({ vorgang: 'h-2', revision: 1, zustand: 'entwurf' }),
    ])

    await zuPhase(await zeigeHandel(), 'gegenueber')

    const wahl = await screen.findByRole('group', { name: eigene.wahlTitel })
    expect(HandelVorgaenge).toHaveBeenCalledWith('bgegenueber')

    expect(
      await within(wahl).findByRole('radio', {
        name: /Angeboten · Revision 2/,
      }),
    ).toBeInTheDocument()
    expect(
      within(wahl).getByRole('radio', { name: /Entwurf · Revision 1/ }),
    ).toBeInTheDocument()
    expect(
      within(wahl).getByRole('radio', { name: new RegExp(eigene.wahlNeu) }),
    ).toBeChecked()
  })

  // **Gewählt und nicht getippt** — die Kästen darunter arbeiten mit dem
  // gewählten Handel, und sein Stand steht schon da.
  it('reicht die Wahl weiter, ohne dass jemand eine Kennung tippt', async () => {
    const HandelStand = vi.fn(() =>
      Promise.resolve({ vorgang: einHandel(), fehler: null }),
    )
    mitHandeln([einHandel()], { HandelStand })

    const hand = await zeigeHandel()
    await zuPhase(hand, 'gegenueber')

    await hand.click(
      await screen.findByRole('radio', { name: /Angeboten · Revision 2/ }),
    )

    const vorgang = screen.getByRole('region', { name: eigene.vorgangTitel })

    // **Revision und „Stand ansehen" stehen für Fachleute** — seit UXR-15
    // (#1923) in einem Kasten.
    await hand.click(
      within(vorgang).getByRole('button', { name: eigene.vorgangFachleute }),
    )
    expect(
      within(vorgang).getByText(new RegExp(`${eigene.vorgangRevision} 2`)),
    ).toBeInTheDocument()

    await hand.click(
      within(vorgang).getByRole('button', { name: eigene.vorgangAnsehen }),
    )
    expect(HandelStand).toHaveBeenCalledWith('h-1')
  })

  // **Zurück zum neuen Handel** — dann steht wieder das Eröffnen da, und der
  // Stand des gewählten ist geräumt (§7, Regel 1).
  it('wechselt zurück zu einem neuen Handel', async () => {
    mitHandeln([einHandel()])

    const hand = await zeigeHandel()
    await zuPhase(hand, 'gegenueber')

    await hand.click(
      await screen.findByRole('radio', { name: /Angeboten · Revision 2/ }),
    )

    // **Die Wahl springt zum Handel** — und zurück geht es über „Mit wem".
    await zuPhase(hand, 'gegenueber')
    await hand.click(
      screen.getByRole('radio', { name: new RegExp(eigene.wahlNeu) }),
    )

    const vorgang = await screen.findByRole('region', {
      name: eigene.vorgangTitel,
    })
    expect(
      within(vorgang).getByRole('button', { name: eigene.vorgangEroeffnen }),
    ).toBeInTheDocument()
    expect(
      within(vorgang).queryByText(new RegExp(`${eigene.vorgangRevision} 2`)),
    ).toBeNull()
  })

  it('sagt, wenn es mit dem Gegenüber noch keinen Handel gibt', async () => {
    mitHandeln([])

    await zuPhase(await zeigeHandel(), 'gegenueber')

    expect(await screen.findByText(eigene.wahlLeer)).toBeInTheDocument()
    expect(
      screen.getByRole('radio', { name: new RegExp(eigene.wahlNeu) }),
    ).toBeChecked()
  })

  // **Die Kennung bleibt erreichbar, mit einer Erklärung** — für einen
  // Handel, dessen erste Nachricht noch unterwegs ist.
  it('öffnet einen Handel über seine Kennung — für Fachleute', async () => {
    const HandelStand = vi.fn(() =>
      Promise.resolve({ vorgang: einHandel({ vorgang: 'h-9' }), fehler: null }),
    )
    mitHandeln([], { HandelStand })

    const hand = await zeigeHandel()
    await zuPhase(hand, 'gegenueber')

    await hand.click(
      await screen.findByRole('button', { name: eigene.wahlFachleute }),
    )
    await hand.click(screen.getByLabelText(eigene.wahlKennung))
    await hand.paste('h-9')
    await hand.click(screen.getByRole('button', { name: eigene.wahlOeffnen }))

    expect(
      screen.getByText(eigene.wahlGewaehlt.replace('{kennung}', 'h-9')),
    ).toBeInTheDocument()

    const vorgang = screen.getByRole('region', { name: eigene.vorgangTitel })
    await hand.click(
      within(vorgang).getByRole('button', { name: eigene.vorgangFachleute }),
    )
    await hand.click(
      within(vorgang).getByRole('button', { name: eigene.vorgangAnsehen }),
    )
    expect(HandelStand).toHaveBeenCalledWith('h-9')
  })

  it('bietet Wiederholen, wenn die Liste nicht kommt', async () => {
    const HandelVorgaenge = vi
      .fn()
      .mockRejectedValueOnce(new Error('Bestand gesperrt'))
      .mockResolvedValue({ vorgaenge: [einHandel()], fehler: null })
    mitHandeln([], { HandelVorgaenge })

    const hand = await zeigeHandel()
    await zuPhase(hand, 'gegenueber')

    await hand.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    )

    expect(
      await screen.findByRole('radio', { name: /Angeboten · Revision 2/ }),
    ).toBeInTheDocument()
  })
})
