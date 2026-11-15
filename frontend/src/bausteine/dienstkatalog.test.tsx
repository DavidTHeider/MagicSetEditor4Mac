// **Der Dienstkatalog im Fenster** — Werkstück A1 von Phase 5.3
// (#1440).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Dienstkatalog } from './Dienstkatalog'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import { kurzform } from '../sprache/kurzform'
import type { DienstfaehigkeitDTO } from '../dienstkatalog'

const texte = kataloge.de.seiten.einstellungen

const vertrag = {
  basisadresse: 'https://api.scryfall.com',
  endpunkte: ['/cards/search'],
  zugang: 'keiner',
  tarif: 'kostenlos',
  grenze: '10 je Sekunde',
  frist: '24 Stunden',
  bedingungen: 'https://scryfall.com/docs/api',
  abgerufen: '2026-08-23',
  beleg: 'https://scryfall.com/docs/api/cards/search',
}

function eintrag(teile: Partial<DienstfaehigkeitDTO>): DienstfaehigkeitDTO {
  return {
    dienst: 'scryfall',
    frage: 'suche',
    status: 'bereit',
    grund: '',
    angabe: '',
    hatWahl: true,
    gewaehlt: true,
    wirkung: 'externSenden',
    vertrag,
    nachweis: null,
    ...teile,
  }
}

const katalog: DienstfaehigkeitDTO[] = [
  eintrag({
    nachweis: {
      lauf: 'internal/scryfall/netzlauf_test.go',
      datum: '2026-09-01',
      fassung: '0.3.0',
      endpunkte: ['/cards/search'],
      giltFuerDieseFassung: false,
    },
  }),
  eintrag({ frage: 'bild', hatWahl: false }),
  eintrag({
    dienst: 'scrydex',
    status: 'zugangFehlt',
    grund: 'die Angabe "team" fehlt im Schlüsselbund',
    angabe: 'team',
    gewaehlt: false,
  }),
  eintrag({
    dienst: 'mtgjson',
    status: 'gestoert',
    grund: 'der Schlüsselbund antwortet nicht',
    gewaehlt: false,
  }),
]

async function hinstellen(
  eintraege: DienstfaehigkeitDTO[] = katalog,
  stand = 'a',
) {
  const Dienstkatalogruf = vi
    .fn()
    .mockResolvedValue({ eintraege, fehler: null })
  belegeApi({ Dienstkatalog: Dienstkatalogruf })

  const sicht = render(
    <Sprachraum sprache="de">
      <Dienstkatalog stand={stand} />
    </Sprachraum>,
  )
  // **Kein Bereich mit Namen mehr** (#2465): den Namen trägt der Knopf des
  // Abschnitts, in dem der Katalog steht — gewartet wird auf die Dienste.
  await screen.findAllByRole('article')

  return { ...sicht, Dienstkatalogruf }
}

describe('der Dienstkatalog', () => {
  it('nennt je Dienst die Fähigkeiten mit ihrem Status', async () => {
    await hinstellen()

    const scryfall = screen.getByRole('article', { name: 'Scryfall' })
    expect(
      within(scryfall).getByText(texte.katalogFragen.suche),
    ).toBeInTheDocument()
    expect(
      within(scryfall).getByText(texte.katalogFragen.bild),
    ).toBeInTheDocument()
    expect(
      within(scryfall).getAllByText(texte.katalogStatus.bereit),
    ).toHaveLength(2)

    const scrydex = screen.getByRole('article', { name: 'Scrydex' })
    expect(
      within(scrydex).getByText(texte.katalogStatus.zugangFehlt),
    ).toBeInTheDocument()
    // **Die fehlende Angabe in der Sprache des Fensters**, nicht der Satz aus Go.
    expect(
      within(scrydex).getByText(`${texte.angaben.team} ist nicht hinterlegt`),
    ).toBeInTheDocument()
    expect(within(scrydex).queryByText(/Schlüsselbund/)).not.toBeInTheDocument()

    // **Eine Störung nennt ihren Grund** — der kommt aus Go.
    const mtgjson = screen.getByRole('article', { name: 'MTGJSON' })
    expect(
      within(mtgjson).getByText('der Schlüsselbund antwortet nicht'),
    ).toBeInTheDocument()
  })

  it('sagt beim Nachweis, wann und mit welcher Fassung — und dass es nicht diese ist', async () => {
    await hinstellen()

    const scryfall = screen.getByRole('article', { name: 'Scryfall' })
    expect(
      within(scryfall).getByText(/gelaufen am 2026-09-01 mit Fassung 0\.3\.0/),
    ).toBeInTheDocument()
    expect(
      within(scryfall).getByText(/nicht mit dieser Fassung/),
    ).toBeInTheDocument()
    // **Der Weg zum Lauf steht unter „Diagnose"**, zu — UX-08 S11.NAT-02
    // (#1680): lange Testdateipfade gehören nicht zwischen die Schalter.
    expect(within(scryfall).queryByText(/netzlauf_test\.go/)).toBeNull()
    await userEvent.click(
      within(scryfall).getByRole('button', { name: texte.katalogDiagnose }),
    )
    expect(within(scryfall).getByText(/netzlauf_test\.go/)).toBeVisible()
    // **Je Endpunkt**: die Suche war im Lauf, das Bild nicht.
    expect(within(scryfall).getByText(texte.katalogImLauf)).toBeInTheDocument()
    expect(
      within(scryfall).getByText(texte.katalogNichtImLauf),
    ).toBeInTheDocument()
  })

  it('nennt einen Lauf mit dieser Fassung als solchen', async () => {
    await hinstellen([
      eintrag({
        dienst: 'mtgapi',
        nachweis: {
          lauf: 'internal/mtgapi/netzlauf_test.go',
          datum: '2026-09-09',
          fassung: '0.4.0',
          endpunkte: ['/cards'],
          giltFuerDieseFassung: true,
        },
      }),
    ])

    const block = screen.getByRole('article', { name: 'magicthegathering.io' })
    expect(
      within(block).getByText(texte.katalogNachweisAktuell),
    ).toBeInTheDocument()
    expect(
      within(block).queryByText(/nicht mit dieser Fassung/),
    ).not.toBeInTheDocument()
    // **Nichts zu diagnostizieren** — der Lauf gilt für diese Fassung.
    expect(
      within(block).queryByRole('button', { name: texte.katalogDiagnose }),
    ).toBeNull()
  })

  it('sagt bei einem Dienst ohne Lauf, dass er ungeprüft ist', async () => {
    await hinstellen()

    const scrydex = screen.getByRole('article', { name: 'Scrydex' })
    expect(
      within(scrydex).getByText(`${texte.dienstUngeprueft}.`),
    ).toBeInTheDocument()
    expect(
      within(scrydex).getByText(kurzform(texte.dienstUngeprueftHinweis)),
    ).toBeInTheDocument()
    expect(
      within(scrydex).queryByText(texte.katalogNichtImLauf),
    ).not.toBeInTheDocument()

    const scryfall = screen.getByRole('article', { name: 'Scryfall' })
    expect(
      within(scryfall).queryByText(`${texte.dienstUngeprueft}.`),
    ).not.toBeInTheDocument()
  })

  it('klappt den Vertrag mit prüfbaren Adressen auf', async () => {
    await hinstellen()

    const scryfall = screen.getByRole('article', { name: 'Scryfall' })
    const [details] = within(scryfall).getAllByText(texte.katalogVertrag)
    expect(details).toBeInTheDocument()

    const bedingungen = within(scryfall).getAllByRole('link', {
      name: 'https://scryfall.com/docs/api',
    })
    expect(bedingungen[0]).toHaveAttribute(
      'href',
      'https://scryfall.com/docs/api',
    )
    expect(bedingungen[0]).toHaveAttribute('target', '_blank')
    expect(within(scryfall).getAllByText('2026-08-23')[0]).toBeInTheDocument()
    expect(
      within(scryfall).getAllByText('/cards/search')[0],
    ).toBeInTheDocument()
    expect(
      within(scryfall).getAllByText(texte.katalogWirkung.externSenden)[0],
    ).toBeInTheDocument()
  })

  it('nennt unbekannte Schlüssel roh statt sie zu verschlucken', async () => {
    // **Ein Dienst, eine Frage, ein Status oder eine Angabe, die der Katalog
    // des Fensters nicht kennt** — ein neuer Wert aus Go steht dann roh da,
    // statt als leere Stelle zu verschwinden (§8: Drift wird gezeigt).
    await hinstellen([
      eintrag({
        dienst: 'fremd',
        frage: 'orakel',
        status: 'sonst' as never,
        angabe: '',
        wirkung: 'sonst',
      }),
      eintrag({
        dienst: 'fremd',
        frage: 'suche',
        status: 'zugangFehlt',
        angabe: 'pin',
      }),
    ])

    const block = screen.getByRole('article', { name: 'fremd' })
    expect(within(block).getByText('orakel')).toBeInTheDocument()
    expect(
      within(block).getByText('pin ist nicht hinterlegt'),
    ).toBeInTheDocument()
    expect(within(block).getAllByText('sonst')).toHaveLength(2)
  })

  it('zeigt ohne Einträge nichts', async () => {
    const Dienstkatalogruf = vi
      .fn()
      .mockResolvedValue({ eintraege: [], fehler: null })
    belegeApi({ Dienstkatalog: Dienstkatalogruf })

    const { container } = render(
      <Sprachraum sprache="de">
        <Dienstkatalog stand="a" />
      </Sprachraum>,
    )
    await vi.waitFor(() => {
      expect(Dienstkatalogruf).toHaveBeenCalled()
    })
    await Promise.resolve()

    // **Kein Katalog, nicht nur kein Bereich** — einen Namen trägt er seit
    // #2465 nicht mehr; eine Abfrage danach wäre immer leer.
    expect(container.querySelector('section')).toBeNull()
  })

  it('meldet einen Fehler der Bindung', async () => {
    belegeApi({
      Dienstkatalog: () =>
        Promise.resolve({
          eintraege: [],
          fehler: {
            code: 'unbekannt',
            schluessel: '',
            meldung: 'der Bestand ist nicht lesbar',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Dienstkatalog stand="a" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('nicht lesbar')
  })

  it('lädt nach, wenn sich der Stand der Schalter ändert', async () => {
    const { rerender, Dienstkatalogruf } = await hinstellen(katalog, 'a')
    expect(Dienstkatalogruf).toHaveBeenCalledTimes(1)

    rerender(
      <Sprachraum sprache="de">
        <Dienstkatalog stand="b" />
      </Sprachraum>,
    )

    await vi.waitFor(() => {
      expect(Dienstkatalogruf).toHaveBeenCalledTimes(2)
    })
  })
})
