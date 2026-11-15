// **Die Sammlung am Fenster** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Sammlung } from './Sammlung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { kataloge } from '../sprache/kataloge'

import type { ExemplarDTO } from '../api'

// geladen wartet, bis die Seite steht.
async function geladen() {
  await screen.findByRole('heading', { name: texte.titel })
}

// formularOeffnen öffnet die Schublade mit dem Formular.
//
// **Seit UXR-10 (#1918) steht es hinter dem führenden Knopf im Kopf** —
// nach dem Artboard `Sammlung`. Vorher stand es in einem Kasten im
// Navigator, bei leerer Sammlung offen: achtzehn Felder in 240 Punkten,
// gemessen am 2026-10-02 (`docs/UIREVIEW.md`).
async function formularOeffnen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.aufnehmenOeffnen }),
  )

  return screen.findByRole('dialog', { name: texte.neuesTitel })
}

const texte = kataloge.de.seiten.sammlung

// exemplar baut ein Exemplar, wie die Bindung es gibt.
function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 3,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
    ...teil,
  }
}

// preisstand baut einen Preisstand.
function preisstand() {
  return {
    printing: 'probe-druck',
    druckID: '',
    quelle: 'scryfall',
    markt: 'paper',
    art: 'retail',
    zustand: '',
    sprache: '',
    lage: 'genannt',
    waehrung: 'EUR',
    finish: 'nonfoil',
    wert: '4.17',
    beobachtet: '',
    erhoben: '2026-08-27T12:00:00Z',
    alterStunden: 3,
  }
}

// belege setzt die Bindungen dieser Seite.
function belege(teil: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    Preise: () =>
      Promise.resolve({ staende: [preisstand()], befunde: [], fehler: null }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    ...teil,
  })
}

// zeige zeichnet die Seite.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// inspector gibt den Bereich rechts — dort wird geändert und gelöscht (U8).
function inspector() {
  return within(screen.getByRole('complementary', { name: texte.inspector }))
}

// zeileWaehlen klickt die Zeile eines Exemplars an — die Wahl für den Inspector.
async function zeileWaehlen(bezeichnung: string) {
  await userEvent.click(await screen.findByText(bezeichnung))
}

// **jsdom rechnet kein Layout**: die Fensterliste liest ihre Höhe aus
// `clientHeight`, und die stellen wir.
beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

describe('Die Sammlung', () => {
  afterEach(raeumeApi)

  it('zeigt die Exemplare mit Menge, Zustand und Ausführung in der Liste', async () => {
    belege()
    zeige()
    await geladen()

    const liste = within(
      await screen.findByRole('grid', { name: texte.listeBezeichnung }),
    )
    expect(liste.getByText('probe-druck')).toBeInTheDocument()
    expect(
      liste.getByText((_, knoten) => knoten?.textContent === '3 × Near Mint'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('region', { name: texte.arbeitsplatz }),
    ).toBeInTheDocument()
  })

  it('erklärt bei leerer Sammlung, was zu tun ist', async () => {
    belege({ Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }) })
    zeige()
    await geladen()
    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
  })

  it('nimmt ein Exemplar auf und lädt danach neu', async () => {
    const angelegt = vi.fn(() => Promise.resolve({ id: 2, fehler: null }))
    const gelesen = vi.fn(() =>
      Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    )
    belege({ ExemplarAnlegen: angelegt, Sammlung: gelesen })
    zeige()
    await formularOeffnen()
    await userEvent.type(await druckfeld(), 'neuer-druck')
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))
    await waitFor(() => {
      expect(angelegt).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(gelesen).toHaveBeenCalledTimes(2)
    })
  })

  it('lässt den Knopf ruhen, solange geschrieben wird', async () => {
    let weiter: (() => void) | undefined
    belege({
      ExemplarAnlegen: () =>
        new Promise((los) => {
          weiter = () => {
            los({ id: 2, fehler: null })
          }
        }),
    })
    zeige()
    await formularOeffnen()
    const druck = await druckfeld()
    await userEvent.type(druck, 'abc')
    const knopf = screen.getByRole('button', { name: texte.aufnehmen })
    await userEvent.click(knopf)
    expect(knopf).toBeDisabled()
    expect(druck).toHaveValue('abc')
    weiter?.()
    await waitFor(() => {
      expect(druck).toHaveValue('')
    })
  })

  it('meldet einen Fehler beim Aufnehmen und lädt nicht nach', async () => {
    const gelesen = vi.fn(() =>
      Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    )
    belege({
      Sammlung: gelesen,
      ExemplarAnlegen: () =>
        Promise.resolve({
          id: 0,
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'so nicht',
          },
        }),
    })
    zeige()
    await formularOeffnen()
    await userEvent.type(await druckfeld(), 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
    expect(gelesen).toHaveBeenCalledTimes(1)
  })

  // **Geändert wird im Inspector** (U8): die Zeile ist die Wahl.
  it('ändert ein gewähltes Exemplar im Inspector', async () => {
    const geaendert = vi.fn(() => Promise.resolve(null))
    belege({ ExemplarAendern: geaendert })
    zeige()
    await geladen()

    expect(inspector().getByText(texte.inspectorLeer)).toBeInTheDocument()
    await zeileWaehlen('probe-druck')
    await userEvent.click(
      inspector().getByRole('button', { name: texte.bearbeiten }),
    )
    const menge = inspector().getByLabelText(texte.menge)
    await userEvent.clear(menge)
    await userEvent.type(menge, '9')
    await userEvent.click(
      inspector().getByRole('button', { name: texte.aenderungSichern }),
    )

    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(
        expect.objectContaining({ id: 1, menge: 9 }),
      )
    })
  })

  it('löscht ein gewähltes Exemplar', async () => {
    const geloescht = vi.fn(() => Promise.resolve(null))
    belege({ ExemplarLoeschen: geloescht })
    zeige()
    await geladen()

    await zeileWaehlen('probe-druck')

    // **Seit UX-10.AK-04 liegt eine Rückfrage dazwischen** — hier gibt es
    // keine Rücknahme, und was weg ist, tippt jemand neu.
    await userEvent.click(
      inspector().getByRole('button', { name: texte.loeschen }),
    )
    await userEvent.click(
      await inspector().findByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(geloescht).toHaveBeenCalledWith(1, 1)
    })
  })

  // **Bearbeitung nach Filterwechsel trifft denselben Datensatz** (#1419).
  it('hält den Inspector nach einem Filterwechsel am selben Exemplar', async () => {
    const geaendert = vi.fn(() => Promise.resolve(null))
    const eines = exemplar()
    const anderes = exemplar({
      id: 2,
      printing: 'anderer',
      bezeichnung: 'anderer',
    })
    const beide = [eines, anderes]

    // **Die Sicht wirkt in Go** (#1500) — das Fenster holt, was zu ihr gehört.
    // Der Beleg prüft damit zugleich, dass die Suche wirklich hinausgeht.
    belege({
      ExemplarAendern: geaendert,
      Sammlung: () => Promise.resolve({ exemplare: beide, fehler: null }),
      Sammlungsfenster: (suche: string) =>
        Promise.resolve({
          exemplare: suche === '' ? beide : [anderes],
          gesamt: suche === '' ? 2 : 1,
          insgesamt: 2,
          stueck: 2,
          fehler: null,
        }),
    })
    zeige()
    await geladen()

    await zeileWaehlen('probe-druck')
    await userEvent.type(screen.getByLabelText(texte.suche), 'anderer')
    expect(
      screen.queryByText('probe-druck', { selector: '[role="gridcell"] span' }),
    ).not.toBeInTheDocument()
    expect(inspector().getByText(texte.inspectorAusserhalb)).toBeInTheDocument()

    await userEvent.click(
      inspector().getByRole('button', { name: texte.bearbeiten }),
    )
    await userEvent.click(
      inspector().getByRole('button', { name: texte.aenderungSichern }),
    )
    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }))
    })
  })

  it('wechselt in die Gruppenansicht und behält die Haken', async () => {
    belege({
      Sammlung: () =>
        Promise.resolve({
          exemplare: [
            exemplar({ setcode: 'm15' }),
            exemplar({ id: 2, setcode: 'khm', bezeichnung: 'zweiter' }),
          ],
          fehler: null,
        }),
    })
    zeige()
    await geladen()

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'zweiter' }),
    )
    expect(screen.getByRole('checkbox', { name: 'zweiter' })).toBeChecked()

    // **Die Ansicht steht seit UXR-10 (#1918) über der Liste**, als
    // Segmente wie bei Sets und Decks — nicht mehr als Auswahlliste im
    // Navigator.
    const ansicht = screen.getByRole('group', { name: texte.ansicht })
    expect(
      screen
        .getByRole('complementary', { name: texte.navigator })
        .contains(ansicht),
    ).toBe(false)
    await userEvent.click(
      within(ansicht).getByRole('button', { name: texte.ansichtGruppen }),
    )
    await userEvent.selectOptions(
      screen.getByLabelText(texte.gruppierung),
      'set',
    )
    expect(screen.getAllByRole('rowheader').map((k) => k.textContent)).toEqual([
      'm151 Exemplare · 3 Stück',
      'khm1 Exemplare · 3 Stück',
    ])
    // **Seit #2016 bleibt der Haken**: Ansicht und Gruppierung zeigen
    // dieselben Zeilen anders, wie es das Artboard *Sammlung* verlangt; bis
    // dahin räumte jeder Wechsel der Sicht. Ein Filter räumt weiter
    // (`haken/sammlungsauswahl.test.tsx`).
    expect(screen.getByRole('checkbox', { name: 'zweiter' })).toBeChecked()
  })

  it('nennt den Vorbehalt zu den Preisen', async () => {
    belege()
    zeige()
    await geladen()
    expect(await screen.findByText(texte.vorbehalt)).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Lesen der Sammlung', async () => {
    belege({
      Sammlung: () =>
        Promise.resolve({
          exemplare: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'Der Bestand schweigt.',
          },
        }),
    })
    zeige()
    await geladen()
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
