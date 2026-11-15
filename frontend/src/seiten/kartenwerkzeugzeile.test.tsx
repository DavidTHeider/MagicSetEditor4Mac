// **Über dem Arbeitsplatz führt eine Handlung, das Seltene steht im Menü** —
// UXR-05 (#1913).
//
// # Was hier gemessen wird
//
// **Dass die Zeile über den drei Regionen zwei Bedienelemente trägt**: „Karte
// anlegen" und das Menü „Mehr". Bis UXR-05 standen dort mit dem Spielpaket
// `magic` fünf gleichrangige — Scryfall, Skripte, Umschlag und Autoersetzung
// neben dem Anlegen —, und das Artboard *Karten* zeigt keinen davon.
//
// **Dass keiner der vier Wege verloren geht**: jeder steht im Menü, per Maus
// und per Tastatur, und öffnet, was vorher sein Aufklapper öffnete — oder tut,
// was sein Knopf tat.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

const texte = kataloge.de
const eigene = texte.seiten.karten

afterEach(() => {
  raeumeApi()
})

const zweiSkripte = [
  { name: 'Zwei Farben', kuerzel: 'Z', beschreibung: '' },
  { name: 'Leerer Zug', kuerzel: '', beschreibung: '' },
]

function zeige(zusatz: Partial<Api> = {}, bestandGeaendert = vi.fn()) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...ausKarten(() =>
      Promise.resolve({
        karten: [
          { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Erste' }] },
        ],
      }),
    ),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Kartenskripte: () =>
      Promise.resolve({ skripte: zweiSkripte, fehler: null }),
    Autoersetzungen: () =>
      Promise.resolve({ regeln: [], an: true, fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        setID="7"
        karteID={undefined}
        karteGewaehlt={vi.fn()}
        bestandGeaendert={bestandGeaendert}
        verlaufsstand={0}
        verlaufsfehler={undefined}
      />
    </Sprachraum>,
  )

  return bestandGeaendert
}

// waehle öffnet das Menü und wählt einen Eintrag — mit der Maus.
async function waehle(name: string) {
  await userEvent.click(
    await screen.findByRole('button', { name: eigene.mehr }),
  )
  await userEvent.click(await screen.findByRole('menuitem', { name }))
}

describe('Die Zeile über dem Kartenarbeitsplatz', () => {
  it('trägt eine Handlung und ein Menü — die seltenen Wege stehen nicht offen da', async () => {
    zeige()

    expect(
      await screen.findByRole('button', { name: eigene.anlegen }),
    ).toBeInTheDocument()
    const mehr = await screen.findByRole('button', { name: eigene.mehr })
    expect(mehr).toHaveAttribute('aria-haspopup', 'menu')

    // **Erst der Beleg, dass die Skripte geladen sind** (§3) — sonst träfe
    // die Verneinung darunter auch eine Zeile, deren Skriptknopf nur noch
    // nicht da ist. Ein Menüeintrag ist kein Knopf; gezählt wird darunter,
    // was **offen** in der Zeile steht.
    await userEvent.click(mehr)
    await screen.findByRole('menuitem', { name: eigene.kartenskripteTitel })

    for (const selten of [
      eigene.ausScryfallAufklapp,
      eigene.kartenskripteTitel,
      eigene.umschlagAnnehmen,
      eigene.autoersetzung.ueberschrift,
    ]) {
      expect(
        screen.queryByRole('button', { name: selten }),
        selten,
      ).not.toBeInTheDocument()
    }
  })

  it('nennt im Menü alle vier Wege, wenn das Spielpaket Skripte hat', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.mehr }),
    )

    const menue = await screen.findByRole('menu')
    await waitFor(() => {
      expect(
        within(menue)
          .getAllByRole('menuitem')
          .map((eintrag) => eintrag.textContent),
      ).toEqual([
        eigene.ausScryfallAufklapp,
        eigene.kartenskripteTitel,
        eigene.umschlagAnnehmen,
        eigene.autoersetzung.ueberschrift,
      ])
    })
  })

  it('lässt die Skripte weg, wenn das Spielpaket keine hat', async () => {
    zeige({
      Kartenskripte: () => Promise.resolve({ skripte: [], fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.mehr }),
    )

    expect(
      within(await screen.findByRole('menu')).queryByRole('menuitem', {
        name: eigene.kartenskripteTitel,
      }),
    ).not.toBeInTheDocument()
  })

  it('öffnet den Scryfall-Kasten aus dem Menü und schließt ihn wieder', async () => {
    zeige()

    await waehle(eigene.ausScryfallAufklapp)

    const kasten = await screen.findByRole('region', {
      name: eigene.ausScryfallAufklapp,
    })
    expect(
      within(kasten).getByRole('button', {
        name: eigene.scryfallkarte.anlegen,
      }),
    ).toBeInTheDocument()

    await userEvent.click(
      within(kasten).getByRole('button', {
        name: texte.bausteine.aktionen.schliessen,
      }),
    )

    expect(
      screen.queryByRole('region', { name: eigene.ausScryfallAufklapp }),
    ).not.toBeInTheDocument()

    // **Und der Fokus steht wieder am Menü** — der Knopf, der schloss, ist
    // mit dem Kasten aus dem Bild.
    expect(screen.getByRole('button', { name: eigene.mehr })).toHaveFocus()
  })

  it('setzt den Fokus in den Kasten, den das Menü öffnet', async () => {
    zeige()

    await waehle(eigene.ausScryfallAufklapp)

    expect(
      await screen.findByRole('region', { name: eigene.ausScryfallAufklapp }),
    ).toHaveFocus()
  })

  it('führt ein Skript aus dem Kasten aus, den das Menü öffnet', async () => {
    const KartenskriptAusfuehren = vi.fn(() =>
      Promise.resolve({ karten: ['2'], befunde: [], fehler: null }),
    )
    zeige({ KartenskriptAusfuehren })

    await waehle(eigene.kartenskripteTitel)

    const kasten = await screen.findByRole('region', {
      name: eigene.kartenskripteTitel,
    })
    await userEvent.click(
      within(kasten).getByRole('button', { name: 'Zwei Farben' }),
    )

    expect(KartenskriptAusfuehren).toHaveBeenCalledWith('7', 'Zwei Farben')
  })

  it('zeigt die Autoersetzung im Kasten — ein Kasten zur Zeit', async () => {
    zeige()

    await waehle(eigene.ausScryfallAufklapp)
    await waehle(eigene.autoersetzung.ueberschrift)

    const kasten = await screen.findByRole('region', {
      name: eigene.autoersetzung.ueberschrift,
    })
    expect(
      await within(kasten).findByRole('checkbox', {
        name: eigene.autoersetzung.schalter,
      }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('region', { name: eigene.ausScryfallAufklapp }),
    ).not.toBeInTheDocument()
  })

  // **Ein Set ohne Skripte schließt den Skriptkasten** — der Eintrag fehlt
  // im Menü, und ein offener Kasten ohne ihn wäre einer, den niemand mehr
  // öffnen könnte.
  it('schließt den Skriptkasten, wenn das neue Set keine Skripte kennt', async () => {
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      ...ausKarten((setID) =>
        Promise.resolve({
          karten: [
            {
              id: `k${setID}`,
              nummer: 0,
              werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
            },
          ],
        }),
      ),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
      Kartenskripte: (setID: string) =>
        Promise.resolve({
          skripte: setID === '7' ? zweiSkripte : [],
          fehler: null,
        }),
    })

    const seite = (setID: string) => (
      <Sprachraum sprache="de">
        <Karten
          setID={setID}
          karteID={undefined}
          karteGewaehlt={vi.fn()}
          bestandGeaendert={vi.fn()}
          verlaufsstand={0}
          verlaufsfehler={undefined}
        />
      </Sprachraum>
    )
    const { rerender } = render(seite('7'))

    await waehle(eigene.kartenskripteTitel)
    expect(
      await screen.findByRole('region', { name: eigene.kartenskripteTitel }),
    ).toBeInTheDocument()

    rerender(seite('8'))

    await screen.findByRole('button', {
      name: `Karte aus 8 — ${eigene.bearbeiten}`,
    })
    await waitFor(() => {
      expect(
        screen.queryByRole('region', { name: eigene.kartenskripteTitel }),
      ).not.toBeInTheDocument()
    })
  })

  // **Zu heißt verborgen, nicht fort** — die Regel des Aufklappens aus
  // `docs/UX-ARCHITEKTUR.md` (S11.NAT-03): wer zwischen zwei Kästen
  // wechselt, findet im ersten wieder, was er getippt hatte.
  it('behält, was im Kasten getippt war, über einen Wechsel hinweg', async () => {
    zeige()

    await waehle(eigene.ausScryfallAufklapp)
    await userEvent.type(
      screen.getByLabelText(eigene.scryfallkarte.name),
      'Lightning Bolt',
    )

    await waehle(eigene.autoersetzung.ueberschrift)
    await waehle(eigene.ausScryfallAufklapp)

    expect(screen.getByLabelText(eigene.scryfallkarte.name)).toHaveValue(
      'Lightning Bolt',
    )
  })

  it('nimmt eine Karte aus dem Umschlag an, sagt es und lädt nach', async () => {
    const UmschlagKarte = vi.fn(() =>
      Promise.resolve({ karteID: '42', fehler: null }),
    )
    const geaendert = zeige({ UmschlagKarte })

    await waehle(eigene.umschlagAnnehmen)

    expect(UmschlagKarte).toHaveBeenCalledWith('7')
    expect(await screen.findByText(eigene.umschlagAngenommen)).toHaveAttribute(
      'role',
      'status',
    )
    expect(geaendert).toHaveBeenCalled()
  })

  it('meldet eine Absage des Umschlags an der Zeile', async () => {
    zeige({
      UmschlagKarte: () => Promise.reject(new Error('geht nicht')),
    })

    await waehle(eigene.umschlagAnnehmen)

    // **Am Satz gegriffen und nicht an der Rolle**: die Kartenseite meldet
    // daneben, was diese Vorrichtung nicht belegt (Verlauf, Vorschau).
    expect(await screen.findByText('geht nicht')).toHaveAttribute(
      'role',
      'alert',
    )
  })

  it('erreicht jeden Weg auch mit der Tastatur', async () => {
    zeige()

    const mehr = await screen.findByRole('button', { name: eigene.mehr })
    mehr.focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(
      await screen.findByRole('menuitem', { name: eigene.ausScryfallAufklapp }),
    ).toHaveFocus()

    await userEvent.keyboard('{Enter}')

    expect(
      await screen.findByRole('region', { name: eigene.ausScryfallAufklapp }),
    ).toBeInTheDocument()
  })
})
