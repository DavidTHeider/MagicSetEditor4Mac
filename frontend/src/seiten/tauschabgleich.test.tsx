// **Der Tauschabgleich am Kontakt** — Werkstück T4 von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass ein unlesbarer Stand nicht hinausgeht.** „Dazu passt nichts" und
// „das war kein Stand" sind zwei Auskünfte; wer den Text ungeprüft schickte,
// bekäme für beides dieselbe.
//
// **Dass die Erklärung an jeder Zeile steht** — Rang und Art sagen, warum
// diese Zeile dasteht.
//
// **Und dass ein einseitiger Abgleich kein Vorschlag ist.**
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import {
  oeffneAbgleich,
  bestaetigt,
  einAbgleich,
  einKandidat,
  einStand,
  setzeEin,
  zeigeHandel,
} from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

describe('Was zwei Sammlungen einander zu bieten haben', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('steht einem unbestätigten Gegenüber gar nicht offen', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(eineListe()) })

    render(
      <Sprachraum sprache="de">
        <Handeln />
      </Sprachraum>,
    )

    // **Gewählt wird in der Schublade hinter „Neuer Tausch …"** (UXR-15).
    await userEvent.click(
      await screen.findByRole('button', { name: eigene.neuerTausch }),
    )

    // **Es steht nicht einmal zur Wahl.** Ein Eintrag, der aufschlägt und
    // dann nichts kann, wäre ein Weg ins Leere: an einem unbestätigten
    // Gegenüber wirkt keine Freigabe, und ein Abgleich bekäme nichts.
    expect(await screen.findByText(eigene.partnerLeer)).toBeInTheDocument()

    expect(screen.queryByText(eigene.tauschTitel)).not.toBeInTheDocument()
  })

  it('nennt beide Richtungen samt ihrer Erklärung', async () => {
    const abgeglichen = vi.fn(() => Promise.resolve(einAbgleich()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: abgeglichen,
    })

    await setzeEin(await zeigeHandel(), einStand)

    const gebe = await screen.findByRole('list', { name: eigene.tauschGebe })

    expect(within(gebe).getByText(/Shivan Dragon/)).toBeInTheDocument()

    const bekomme = screen.getByRole('list', { name: eigene.tauschBekomme })

    // **Der Rang steht dran** — er ist die halbe Erklärung.
    expect(within(bekomme).getByText(/hoch/)).toBeInTheDocument()

    expect(abgeglichen).toHaveBeenCalledWith(
      'bgegenueber',
      {
        angebot: [{ printing: 'bolt-m10', menge: 4 }],
        wuensche: [{ printing: 'shivan-lea', menge: 1 }],
      },
      1,
      0,
    )
  })

  // **Eine Richtung genügt.** Wer nur die Wünsche seines Gegenübers hat,
  // bekommt den halben Abgleich — und nicht eine Meldung, dass etwas fehle.
  it('nimmt einen Stand mit nur einer Liste', async () => {
    const abgeglichen = vi.fn(() => Promise.resolve(einAbgleich()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: abgeglichen,
    })

    const hand = await zeigeHandel()

    await setzeEin(
      hand,
      JSON.stringify({ wuensche: [{ printing: 'shivan-lea', menge: 1 }] }),
    )

    expect(abgeglichen).toHaveBeenCalledWith(
      'bgegenueber',
      { angebot: [], wuensche: [{ printing: 'shivan-lea', menge: 1 }] },
      1,
      0,
    )

    // **Und andersherum ebenso** — wer nur die Dubletten hat, bekommt die
    // andere Richtung.
    await setzeEin(
      hand,
      JSON.stringify({ angebot: [{ printing: 'bolt-m10', menge: 4 }] }),
    )

    expect(abgeglichen).toHaveBeenLastCalledWith(
      'bgegenueber',
      { angebot: [{ printing: 'bolt-m10', menge: 4 }], wuensche: [] },
      1,
      0,
    )
  })

  // **Ein unlesbarer Stand geht nicht hinaus.**
  it('schickt einen unlesbaren Stand nicht hinaus', async () => {
    const abgeglichen = vi.fn(() => Promise.resolve(einAbgleich()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: abgeglichen,
    })

    await setzeEin(await zeigeHandel(), '{kein json')

    expect(await screen.findByText(eigene.tauschUnlesbar)).toBeInTheDocument()

    expect(abgeglichen).not.toHaveBeenCalled()
  })

  // **Und ein leerer Text ebenso wenig** — er ist kein Stand, sondern keiner.
  it('schickt einen leeren Text nicht hinaus', async () => {
    const abgeglichen = vi.fn(() => Promise.resolve(einAbgleich()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: abgeglichen,
    })

    await setzeEin(await zeigeHandel(), '')

    expect(await screen.findByText(eigene.tauschUnlesbar)).toBeInTheDocument()

    expect(abgeglichen).not.toHaveBeenCalled()
  })

  // **Ein Stand ohne beide Listen ist keiner** — eine Zahl, ein Wort, ein
  // Verbund ohne `angebot` und ohne `wuensche`.
  it('weist einen Stand ohne beide Listen ab', async () => {
    const abgeglichen = vi.fn(() => Promise.resolve(einAbgleich()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: abgeglichen,
    })

    const hand = await zeigeHandel()

    for (const text of ['42', '{"zeilen":[]}']) {
      await setzeEin(hand, text)

      expect(await screen.findByText(eigene.tauschUnlesbar)).toBeInTheDocument()
    }

    expect(abgeglichen).not.toHaveBeenCalled()
  })

  // **Eine Seite ohne Treffer sagt es** — und der fehlende Rest steht dran.
  it('sagt, wo nichts passt, und nennt den Rest', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: () =>
        Promise.resolve(
          einAbgleich({
            gebe: [einKandidat({ fehlend: 3 })],
            bekomme: [],
            hatVorschlag: false,
            befunde: ['Zeile ohne Druck: Diese Zeile nennt keinen Druck.'],
          }),
        ),
    })

    await setzeEin(await zeigeHandel(), einStand)

    expect(await screen.findByText(eigene.tauschLeer)).toBeInTheDocument()

    const gebe = screen.getByRole('list', { name: eigene.tauschGebe })

    expect(within(gebe).getByText(/−3/)).toBeInTheDocument()

    expect(screen.getByText(eigene.tauschOhneVorschlag)).toBeInTheDocument()

    const befunde = screen.getByRole('list', { name: eigene.tauschBefunde })

    expect(within(befunde).getByText(/Zeile ohne Druck/)).toBeInTheDocument()
  })

  // **Ein gescheiterter Abgleich sagt es** — und zeigt kein Ergebnis von
  // vorhin (§7, Regel 1).
  it('meldet einen gescheiterten Abgleich', async () => {
    let erster = true

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Tauschabgleich: () => {
        if (erster) {
          erster = false

          return Promise.resolve(einAbgleich())
        }

        return Promise.resolve(
          einAbgleich({
            fehler: {
              code: 'validierung' as const,
              schluessel: 'gibtesnicht',
              meldung: 'so nicht',
            },
          }),
        )
      },
    })

    const hand = await zeigeHandel()

    await setzeEin(hand, einStand)

    expect(
      await screen.findByRole('list', { name: eigene.tauschGebe }),
    ).toBeInTheDocument()

    await setzeEin(hand, einStand)

    expect(await screen.findByText('so nicht')).toBeInTheDocument()

    expect(
      screen.queryByRole('list', { name: eigene.tauschGebe }),
    ).not.toBeInTheDocument()
  })
})

// **Der fremde Stand kann aus einem Umschlag kommen** — Werkstück T7 von
// Phase 5.6 (#1472).
//
// **Gelesen und nicht gebucht**: eine Sammlungssicht sagt, was ein Gegenüber
// hat — in den eigenen Bestand geht sie nie.
describe('Der fremde Stand aus einem Umschlag', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('setzt ein, was im Umschlag steht', async () => {
    const gelesen = vi.fn(() =>
      Promise.resolve({
        zeilen: [
          {
            referenz: 'r1',
            printing: 'bolt-m10',
            bezeichnung: 'Lightning Bolt',
            menge: 4,
            zustand: 'nearMint',
            sprache: 'de',
            finish: 'nonfoil',
          },
        ],
        fehler: null,
      }),
    )

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      UmschlagSicht: gelesen,
    })

    const hand = await zeigeHandel()
    await oeffneAbgleich(hand)

    await hand.click(
      screen.getByRole('button', { name: eigene.tauschAusUmschlag }),
    )

    const feld = await screen.findByLabelText(eigene.tauschStand)

    // **Der eingesetzte Text ist der Stand** — er geht danach durch dieselbe
    // Prüfung wie einer von Hand.
    expect((feld as HTMLTextAreaElement).value).toContain('bolt-m10')

    expect(gelesen).toHaveBeenCalled()
  })

  it('meldet einen Umschlag, der sich nicht lesen lässt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      UmschlagSicht: () =>
        Promise.resolve({
          zeilen: [],
          fehler: {
            code: 'validierung' as const,
            schluessel: 'umschlag',
            meldung: 'das ist kein Umschlag',
          },
        }),
    })

    const hand = await zeigeHandel()
    await oeffneAbgleich(hand)

    await hand.click(
      screen.getByRole('button', { name: eigene.tauschAusUmschlag }),
    )

    expect(await screen.findByText('das ist kein Umschlag')).toBeInTheDocument()
  })
})
