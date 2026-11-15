// **Was gerade läuft** — TRADE-01.AK-01/02/04
// (#1715):
// alle Handel dieses Profils, mit Lage und nächstem Schritt; Suche und
// Auswahl gehen an den Dienst, und ein Klick schlägt Gegenüber und Vorgang
// auf.
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
import { handelsuebersicht } from '../test/handelsbestand'

import type { Handelsvorgangzeile } from '../apihandelsverzeichnis'

const eigene = kataloge.de.seiten.handeln.liste

function mitListe(zeilen: Handelsvorgangzeile[], mehr = {}) {
  const HandelUebersicht = vi.fn(() => eineUebersicht(zeilen))

  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    HandelUebersicht,
    ...mehr,
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  return HandelUebersicht
}

describe('Was gerade läuft', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('nennt jeden Handel mit Gegenüber, Lage und Karten', async () => {
    const HandelUebersicht = mitListe([
      eineZeile({ unbestaetigt: true }),
      eineZeile(
        {
          lage: 'wartet',
          naechster: '',
          gegenueber: [{ kennung: 'bunbekannt-und-lang', name: '' }],
          karten: [],
        },
        { vorgang: 'h-2', zustand: 'entwurf' },
      ),
    ])

    const liste = await screen.findByRole('region', {
      name: new RegExp(eigene.titel),
    })

    expect(
      await within(liste).findByText('Du bist dran: zustimmen'),
    ).toBeInTheDocument()
    expect(
      within(liste).getByText('Wartet auf bunbekannt-u'),
    ).toBeInTheDocument()
    expect(within(liste).getByText(eigene.unbestaetigt)).toBeInTheDocument()
    expect(within(liste).getByText(eigene.entwurf)).toBeInTheDocument()
    expect(
      within(liste).getByText(new RegExp(eigene.ohneKarten)),
    ).toBeInTheDocument()
    expect(within(liste).getByText('2')).toBeInTheDocument()

    expect(HandelUebersicht).toHaveBeenCalledWith(
      expect.objectContaining({
        lagen: ['dran', 'wartet'],
        archiv: false,
        suche: '',
        sprache: 'de',
      }),
    )
  })

  it('sagt die drei Enden in Worten', async () => {
    mitListe([
      eineZeile(
        { lage: 'abgeschlossen', naechster: '' },
        { zustand: 'beidseitigVerbucht', ende: true },
      ),
      eineZeile(
        { lage: 'abgebrochen', naechster: '', archiviert: true },
        { vorgang: 'h-2', zustand: 'abgelehnt', ende: true },
      ),
      eineZeile(
        { naechster: 'unbekannt' },
        { vorgang: 'h-3', zustand: 'klaerung' },
      ),
      // **Ohne Gegenüber gibt es nichts aufzuschlagen** — nur in einem von
      // Hand geänderten Bestand.
      eineZeile(
        { lage: 'wartet', naechster: '', gegenueber: [] },
        { vorgang: 'h-4' },
      ),
    ])

    const liste = await within(
      await screen.findByRole('region', { name: new RegExp(eigene.titel) }),
    ).findByRole('list')

    expect(
      await within(liste).findByText(eigene.lageAbgeschlossen),
    ).toBeInTheDocument()
    expect(within(liste).getByText('Beendet: Abgelehnt')).toBeInTheDocument()
    expect(within(liste).getByText(eigene.archiviert)).toBeInTheDocument()
    expect(
      within(liste).getByText(eigene.lageDranOhneSchritt),
    ).toBeInTheDocument()
    expect(
      within(liste).getByRole('button', { name: /Wartet auf/ }),
    ).toBeDisabled()
  })

  it('gibt Suche, Auswahl, Ordnung und Archiv an den Dienst', async () => {
    const HandelUebersicht = mitListe([eineZeile()])
    const hand = userEvent.setup()

    await hand.type(await screen.findByLabelText(eigene.suche), 'bolt')

    // **Fünf Werte sind ein Auswahlfeld, das Archiv ein Schalter** (R7 in
    // `docs/UIREVIEW.md`, UXR-15) — bis dahin fünf Umschalter und ein
    // Kästchen.
    await hand.selectOptions(
      screen.getByRole('combobox', { name: eigene.auswahl }),
      'abgeschlossen',
    )
    await hand.selectOptions(screen.getByLabelText(eigene.ordnung), 'partner')
    await hand.click(screen.getByRole('switch', { name: eigene.archiv }))

    expect(screen.getByLabelText(eigene.auswahl)).toHaveValue('abgeschlossen')
    expect(
      screen.queryByRole('button', { name: eigene.auswahlAbgeschlossen }),
    ).not.toBeInTheDocument()
    await vi.waitFor(() => {
      expect(HandelUebersicht).toHaveBeenLastCalledWith(
        expect.objectContaining({
          suche: 'bolt',
          lagen: ['abgeschlossen'],
          ordnung: 'partner',
          archiv: true,
        }),
      )
    })
  })

  it('erklärt eine leere Liste und setzt eine leere Auswahl zurück', async () => {
    const HandelUebersicht = mitListe([])
    const hand = userEvent.setup()

    expect(await screen.findByText(eigene.leer)).toBeInTheDocument()

    await hand.selectOptions(screen.getByLabelText(eigene.auswahl), 'wartet')
    await hand.click(
      await screen.findByRole('button', { name: eigene.zuruecksetzen }),
    )

    await vi.waitFor(() => {
      expect(HandelUebersicht).toHaveBeenLastCalledWith(
        expect.objectContaining({ lagen: ['dran', 'wartet'] }),
      )
    })
    expect(await screen.findByText(eigene.leer)).toBeInTheDocument()
  })

  // **Die leere Vorgabe ist „Offen" und nicht „gar keiner"** — gefunden mit
  // zwei Fenstern (Hüllenprobe `nat04handel`, #1672): nach einem
  // abgeschlossenen Tausch sagte die Liste „Noch kein Handel", und der eben
  // abgeschlossene stand daneben. Der Satz nennt jetzt, wo er steht — in
  // jeder Sprache mit dem Namen der Auswahl, den das Fenster dort zeigt.
  it('sagt bei leerer Vorgabe, wo die beendeten Handel stehen', () => {
    for (const [sprache, katalog] of Object.entries(kataloge)) {
      const liste = katalog.seiten.handeln.liste

      expect(liste.leer, sprache).toContain(liste.auswahl)
    }
  })

  // **Acht und der Rest stückweise** (R2, R6; UXR-15, #1923):
  // fünfundzwanzig abgeschlossene Handel waren fünfundzwanzig Knöpfe.
  it('zeigt acht Handel und den Rest stückweise', async () => {
    mitListe([], { HandelUebersicht: handelsuebersicht })
    const hand = userEvent.setup()

    await hand.selectOptions(
      await screen.findByLabelText(eigene.auswahl),
      'abgeschlossen',
    )

    const liste = await screen.findByRole('region', {
      name: new RegExp(eigene.titel),
    })
    await within(liste).findByText('8 von 25')
    expect(within(liste).getAllByRole('listitem')).toHaveLength(8)

    await hand.click(
      within(liste).getByRole('button', { name: 'Weitere 17 zeigen' }),
    )

    expect(within(liste).getAllByRole('listitem')).toHaveLength(25)
  })

  it('meldet einen Ladefehler und lädt auf Wunsch nach', async () => {
    const HandelUebersicht = vi
      .fn()
      .mockResolvedValueOnce({
        zeilen: [],
        fehler: { code: 'unbekannt', schluessel: '', meldung: 'klemmt' },
      })
      .mockResolvedValue({ zeilen: [], fehler: null })

    mitListe([], { HandelUebersicht })
    const hand = userEvent.setup()

    await hand.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    )

    expect(await screen.findByText(eigene.leer)).toBeInTheDocument()
  })

  it('schlägt Gegenüber und Vorgang auf einen Klick auf', async () => {
    mitListe([eineZeile()], {
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [einVorgang()], fehler: null }),
    })
    const hand = userEvent.setup()

    const zeile = await screen.findByRole('button', {
      name: /Turfer.*Du bist dran/,
    })
    await hand.click(zeile)

    expect(zeile).toHaveAttribute('aria-current', 'true')

    // **Der Assistent steht in der Phase des Handels**, und in „Mit wem" ist
    // er gewählt (TRADE-02, #1716).
    expect(
      await screen.findByRole('listitem', { current: 'step' }),
    ).toHaveTextContent(kataloge.de.seiten.handeln.assistent.phasen.angebot)
    await zuPhase(hand, 'gegenueber')
    expect(
      await screen.findByRole('radio', { name: /Angeboten · Revision 1/ }),
    ).toBeChecked()
  })

  // **Ein anderer Prozess schreibt — die Liste lädt nach** (AK-04): der
  // Bestandsstand der Anwendung zählt fremde Schreibvorgänge, und der
  // Bereich reicht ihn an die Liste.
  it('lädt nach, wenn ein anderer Prozess geschrieben hat', async () => {
    const HandelUebersicht = vi.fn(() => eineUebersicht([eineZeile()]))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([])),
      HandelUebersicht,
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Handeln bestandsstand={0} />
      </Sprachraum>,
    )

    await screen.findByRole('button', { name: /Turfer/ })
    const vorher = HandelUebersicht.mock.calls.length

    rerender(
      <Sprachraum sprache="de">
        <Handeln bestandsstand={1} />
      </Sprachraum>,
    )

    await vi.waitFor(() => {
      expect(HandelUebersicht.mock.calls.length).toBeGreaterThan(vorher)
    })
  })
})
