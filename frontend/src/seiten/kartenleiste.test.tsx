// **Eine Zeile über der Karte** — UXR-06
// (#1914).
//
// # Der Befund
//
// Gemessen am nativen Fenster standen auf der Kartenseite nach UXR-05 noch 27
// Bedienelemente im Bild (R2 aus `docs/UIREVIEW.md`: zwanzig). Über der
// Vorschau standen fünf Zeilen: Blättern, das Menü „Diese Karte" mit
// „Teilen …" daneben, der Aufklapper „Ablage und Bezug", sechs Knöpfe für
// den Maßstab und drei Kästchen. Das Artboard *Karten* zeigt eine: ← und →,
// die Stelle und den Maßstab.
//
// # Was hier gehalten wird
//
// - **Vor der Karte stehen drei Bedienelemente**: ←, → und „Diese Karte".
//   Der Maßstab ist eine Auswahl (`massstab.test.tsx`) und zählt als Feld.
// - **Was seltener ist, steht im Menü**: die Zwischenablage, der Bezug zu
//   einem Druck und das Teilen — keine Fähigkeit geht verloren, sie liegen
//   eine Ebene tiefer, und ⌘C kündigt der Eintrag weiter an.
// - **Ein Kasten aus dem Menü gehört zur Karte** (§7, Regel 1): wer blättert,
//   schließt ihn.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { vorschau } from '../test/bildflaeche'
import { ausKarten } from '../test/kartenproben'
import { verstoesse } from '../test/axe'
import { oeffneKartenhandlungen } from '../test/kartenlasche'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.seiten.karten

const karten: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Mirri' }] },
]

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...ausKarten(() => Promise.resolve({ karten })),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Verlauf: () => Promise.resolve({ schritte: [] }),
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="1"
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

// vorDerKarte sind die Bedienelemente, die im Bereich der Karte **vor** dem
// Bild stehen — dieselbe Auswahl wie im Drehbuch `uirevue`.
async function vorDerKarte(): Promise<HTMLElement[]> {
  const bild = await screen.findByRole('img', { name: eigene.vorschau })
  const bereich = screen.getByRole('region', { name: eigene.bearbeiten })

  return [
    ...bereich.querySelectorAll<HTMLElement>(
      'button, [role=button], a[href], [role=tab], [role=menuitem], summary',
    ),
  ].filter(
    (element) =>
      !bild.parentElement?.contains(element) &&
      element.compareDocumentPosition(bild) & Node.DOCUMENT_POSITION_FOLLOWING,
  )
}

describe('Die Zeile über der Karte', () => {
  it('trägt vor der Karte nur ←, → und „Diese Karte"', async () => {
    zeige()

    const vorn = await vorDerKarte()

    expect(
      vorn.map(
        (element) => element.getAttribute('aria-label') ?? element.textContent,
      ),
    ).toEqual([
      `${eigene.blaettern.vorige} (${eigene.blaettern.kuerzelVorige})`,
      `${eigene.blaettern.naechste} (${eigene.blaettern.kuerzelNaechste})`,
      eigene.dieseKarte,
    ])

    // **Kein eigener Knopf für das Teilen und kein Aufklapper davor.**
    expect(
      screen.queryByRole('button', { name: texte.bausteine.teilen.knopf }),
    ).not.toBeInTheDocument()
  })

  it('führt Zwischenablage, Bezug und Teilen im Menü „Diese Karte"', async () => {
    zeige()

    await oeffneKartenhandlungen()

    const menue = screen.getByRole('menu', { name: eigene.dieseKarte })
    const namen = within(menue)
      .getAllByRole('menuitem')
      .map((eintrag) => eintrag.textContent)

    for (const name of [
      eigene.duplizieren,
      eigene.ablage.kopieren.replace('{zahl}', '1'),
      eigene.ablage.ausschneiden.replace('{zahl}', '1'),
      eigene.ablage.einfuegen,
      eigene.nachOben,
      eigene.nachUnten,
      eigene.bezugAufklapp,
      texte.bausteine.teilen.knopf,
      eigene.loeschen,
    ]) {
      expect(namen.some((text) => text.includes(name))).toBe(true)
    }

    // **⌘C steht weiter am Kopieren** — vorher im Tooltip des Knopfes.
    expect(
      within(menue).getByRole('menuitem', {
        name: new RegExp(
          eigene.ablage.kopieren
            .replace('{zahl}', '1')
            .replace(/[()]/g, '\\$&'),
        ),
      }),
    ).toHaveAttribute('aria-keyshortcuts', 'Meta+C')
  })

  it('teilt das Bild dieser Karte über den Menüeintrag', async () => {
    const TeilenVorbereiten = vi.fn(() =>
      Promise.resolve({
        teilgut: {
          kennung: 'd'.repeat(32),
          art: 'kartenbild' as const,
          titel: 'Feuerball',
          dateiname: 'Feuerball.png',
          groesse: 4096,
          text: '',
          mitgehendes: [],
          verluste: [],
          zuGrossFuerMail: false,
        },
        fehler: null,
      }),
    )
    zeige({
      TeilenVorbereiten,
      TeilenVerwerfen: () => Promise.resolve({ teilgut: null, fehler: null }),
    })

    await oeffneKartenhandlungen()
    await userEvent.click(
      screen.getByRole('menuitem', { name: texte.bausteine.teilen.knopf }),
    )

    expect(await screen.findByText(/Feuerball\.png/)).toBeInTheDocument()
    expect(TeilenVorbereiten).toHaveBeenCalledWith('kartenbild', '1')

    // **„Fertig" räumt die Kopie weg** — sonst räumte sie erst das Verlassen,
    // und das fiele hier hinter das Aufräumen der Fassade.
    await userEvent.click(
      screen.getByRole('button', { name: texte.bausteine.teilen.fertig }),
    )
    await waitFor(() => {
      expect(screen.queryByText(/Feuerball\.png/)).not.toBeInTheDocument()
    })
  })

  it('öffnet den Bezug als Kasten, und Blättern schließt ihn', async () => {
    const KarteVerweis = vi.fn(() =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    )
    zeige({ KarteVerweis })

    await oeffneKartenhandlungen()
    await userEvent.click(
      screen.getByRole('menuitem', { name: eigene.bezugAufklapp }),
    )

    const kasten = await screen.findByRole('region', {
      name: eigene.bezugAufklapp,
    })
    expect(
      await within(kasten).findByText(eigene.verweis.keiner),
    ).toBeInTheDocument()
    expect(KarteVerweis).toHaveBeenCalledWith('1')

    // **Der Fokus steht im Kasten** — wer mit der Tastatur kam, ist dort.
    expect(kasten).toHaveFocus()

    // **Schließen gibt den Fokus an das Menü zurück.**
    await userEvent.click(
      within(kasten).getByRole('button', {
        name: texte.bausteine.aktionen.schliessen,
      }),
    )
    expect(
      screen.queryByRole('region', { name: eigene.bezugAufklapp }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: eigene.dieseKarte }),
    ).toHaveFocus()

    // **Wieder geöffnet und dann geblättert**: der Kasten gehörte zur Karte.
    await oeffneKartenhandlungen()
    await userEvent.click(
      screen.getByRole('menuitem', { name: eigene.bezugAufklapp }),
    )
    await screen.findByRole('region', { name: eigene.bezugAufklapp })

    await userEvent.click(
      screen.getByRole('button', {
        name: `${eigene.blaettern.naechste} (${eigene.blaettern.kuerzelNaechste})`,
      }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('region', { name: eigene.bezugAufklapp }),
      ).not.toBeInTheDocument()
    })
  })

  // **axe gegen WCAG 2.2 AA, mit offenem Menü und offenem Kasten** — die
  // Seitenprüfung in `barrierefreiheit.test.tsx` belegt keine Vorschau und
  // sieht deshalb weder die Ansicht noch die Karte. Hier steht beides, und
  // die Vorlage nennt keine Auflösung, damit auch der Grund an der Auswahl
  // hängt.
  it('besteht axe mit Vorschau, offenem Menü und offenem Bezug', async () => {
    const ohneAufloesung = () => {
      const eine = vorschau().vorschau

      if (eine === null) {
        throw new Error('die Vorrichtung gibt eine Vorschau')
      }

      return Promise.resolve({ vorschau: { ...eine, dpi: 0 } })
    }
    zeige({
      Vorschau: ohneAufloesung,
      Probevorschau: ohneAufloesung,
      KarteVerweis: () =>
        Promise.resolve({ verweis: null, befunde: [], fehler: null }),
    })

    await screen.findByRole('img', { name: eigene.vorschau })
    const bereich = screen.getByRole('region', { name: eigene.bearbeiten })

    await oeffneKartenhandlungen()
    expect(await verstoesse(bereich)).toEqual([])

    await userEvent.click(
      screen.getByRole('menuitem', { name: eigene.bezugAufklapp }),
    )
    await screen.findByText(eigene.verweis.keiner)
    expect(await verstoesse(bereich)).toEqual([])
  })
})
