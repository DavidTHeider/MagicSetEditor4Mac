// **Der Bereich Handeln** — Werkstück T9 von Phase 5.6
// (#1474).
//
// # Was hier gemessen wird
//
// **Dass ein Wechsel des Gegenübers den offenen Stand mitnimmt** (§7, Regel 1
// und 2 in einer Gestalt). Das ist die Zusage des `key` am Handelsraum — ohne
// ihn stünde der halb ausgefüllte Stand des vorigen über dem neuen, und der
// nächste Klick schickte ihn dorthin.
//
// **Dass nur bestätigte Gegenüber zur Wahl stehen.** An einem unbestätigten
// wirkt keine Freigabe; ein Eintrag wäre ein Weg ins Leere.
//
// **Und dass das Marktgebiet sagt, was eine Beobachtung ist.** Eine Zahl ohne
// diesen Satz sähe aus wie ein Preis, den jemand zahlt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe } from './peerkontaktprobe'
import { oeffneAbgleich, waehleGegenueber } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// zweiGegenueber sind zwei bestätigte Kontakte.
function zweiGegenueber() {
  return eineListe([
    einKontakt({ stand: 'bestaetigt', rechte: ['handel'] }),
    einKontakt({
      kennung: 'czweiter',
      eigenerName: 'Zweiter Rechner',
      fingerabdruck: '9999-8888-7777-6666-5555-4444-3333-2222',
      stand: 'bestaetigt',
      rechte: ['handel'],
    }),
  ])
}

// aufschlagen wählt ein Gegenüber — **seit UXR-15 in der Schublade hinter
// „Neuer Tausch …"** (#1923).
const aufschlagen = waehleGegenueber

// oeffneSchublade drückt „Neuer Tausch …" und gibt die Schublade.
async function oeffneSchublade(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(
    await screen.findByRole('button', { name: eigene.neuerTausch }),
  )

  return screen.findByRole('dialog', { name: eigene.neuerTauschTitel })
}

// zeige stellt die Seite auf.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Der Bereich Handeln', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sagt es, solange kein Gegenüber aufgeschlagen ist', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(zweiGegenueber()) })

    zeige()

    expect(await screen.findByText(eigene.partnerKeiner)).toBeInTheDocument()

    // **Und der Kasten steht nicht da** — ein leeres Formular ohne Gegenüber
    // ließe offen, wohin sein Inhalt ginge.
    expect(screen.queryByText(eigene.tauschTitel)).not.toBeInTheDocument()
  })

  it('stellt nur bestätigte Gegenüber zur Wahl', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([
            einKontakt({ stand: 'bestaetigt', rechte: ['handel'] }),
            einKontakt({
              kennung: 'dgesperrt',
              eigenerName: 'Gesperrter Rechner',
              stand: 'gesperrt',
            }),
          ]),
        ),
    })

    const schublade = await oeffneSchublade(zeige())

    expect(
      await within(schublade).findByText('Fremder Rechner'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Gesperrter Rechner')).not.toBeInTheDocument()
  })

  it('räumt den eingesetzten Stand, wenn das Gegenüber wechselt', async () => {
    const abgeglichen = vi.fn()

    belegeApi({
      Peerkontakte: () => Promise.resolve(zweiGegenueber()),
      Handelspoststand: () =>
        Promise.resolve({
          stand: { kennung: 'bgegenueber', offen: 0, hoechsteVersuche: 0 },
          fehler: null,
        }),
      Tauschabgleich: abgeglichen,
    })

    const hand = zeige()

    await aufschlagen(hand, 'Fremder Rechner')
    await oeffneAbgleich(hand)

    const feld = await screen.findByLabelText(eigene.tauschStand)

    await hand.click(feld)
    await hand.paste('{"angebot":[],"wuensche":[]}')

    expect(feld).toHaveValue('{"angebot":[],"wuensche":[]}')

    await aufschlagen(hand, 'Zweiter Rechner')
    await oeffneAbgleich(hand)

    // **Der Stand des vorigen ist weg** — und zwar, weil der Kasten ein
    // anderer ist und nicht, weil jemand ihn gelöscht hat.
    expect(await screen.findByLabelText(eigene.tauschStand)).toHaveValue('')
    expect(abgeglichen).not.toHaveBeenCalled()
  })

  it('meldet, wenn die Liste der Gegenüber klagt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.reject(new Error('der Bestand ist zu')),
    })

    const schublade = await oeffneSchublade(zeige())

    expect(await within(schublade).findByRole('alert')).toHaveTextContent(
      'der Bestand ist zu',
    )
  })

  it('nennt im Markt, dass eine Beobachtung keine Zusage ist', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(zweiGegenueber()),
      Portfolio: () =>
        Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    })

    const hand = zeige()

    await hand.click(screen.getByRole('tab', { name: eigene.gebietMarkt }))

    const gebiet = await screen.findByRole('tabpanel')

    expect(
      within(gebiet).getByRole('heading', { name: eigene.marktTitel }),
    ).toBeInTheDocument()

    // **Der Satz steht über den Zahlen und nicht darunter** — wer ihn erst
    // nach der Summe läse, hätte sie schon für einen Preis gehalten.
    expect(within(gebiet).getByText(/keine Zusage/)).toBeInTheDocument()
  })
})
