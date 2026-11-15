// Die Statistikseite.
//
// Geprüft wird, was sie **zusagt**: die Zahlen kommen aus je einer Abfrage, die
// Darstellung ist ohne Farbe lesbar, und ein Wechsel der Wahl fragt neu.
//
// **Seit UXR-08 gibt es eine Wahl und nicht drei Abschnitte**
// (#1916).
// Belegung, Verteilung eines Felds und die Dimensionen des Spielpakets standen
// untereinander, mit je eigener Überschrift und zwei Auswahlfeldern — und die
// Verteilung erschien **unter** dem Abschnitt der Dimensionen, getrennt von
// ihrer eigenen Auswahl. Das Artboard *Werkzeuge* zeigt eine Wahl „Worüber"
// und darunter **ein** Bild. Die Fälle, die vorher das Feld „Feld" wählten,
// wählen deshalb hier den Eintrag `feld:<name>` derselben Wahl.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { Statistik } from './Statistik'

const texte = kataloge.de.seiten.statistik

const zahlen = {
  karten: 12,
  felder: 3,
  leere: 2,
  // **Der größte Wert ist nicht die Kartenzahl.** Sonst prüfte der Test nicht,
  // worauf sich der Anteil bezieht: 8 von 12 sind 67 %, gemessen am größten
  // Balken wären es 100 %.
  belegung: [
    { name: 'name', karten: 8 },
    { name: 'typ', karten: 6 },
  ],
}

// keineDimensionen ist ein Spielpaket, das nichts zum Auszählen erklärt.
const keineDimensionen = () =>
  Promise.resolve({ dimensionen: [], fehler: null })

function zeige(setID: string) {
  return render(
    <Sprachraum sprache="de">
      <Statistik setID={setID} />
    </Sprachraum>,
  )
}

// worueber ist die eine Wahl der Seite.
function worueber() {
  return screen.getByRole('combobox', { name: texte.worueber })
}

describe('Statistik', () => {
  afterEach(() => {
    raeumeApi()
  })

  // **Der Fall „kein Set" steht hier nicht mehr** — siehe `bilder.test.tsx`
  // und #288.

  // **Eine Abfrage, kein Durchlauf**: die Seite holt die Zahlen fertig ab.
  it('zeigt die Kennzahlen, wie sie kommen', async () => {
    const Statistik = vi.fn(() => Promise.resolve({ statistik: zahlen }))

    belegeApi({
      Statistik,
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: keineDimensionen,
    })
    zeige('7')

    expect(await screen.findByText('12')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(Statistik).toHaveBeenCalledExactlyOnceWith('7')
  })

  // **Ohne Farbe lesbar**: jeder Balken trägt Zahl und Anteil als Text, und die
  // Fläche daneben ist für Vorleser ausgeblendet.
  it('nennt jede Zahl im Text, nicht nur als Fläche', async () => {
    belegeApi({
      Statistik: () => Promise.resolve({ statistik: zahlen }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: keineDimensionen,
    })

    const { container } = zeige('7')

    expect(await screen.findByText('8 (67 %)')).toBeInTheDocument()
    expect(screen.getByText('6 (50 %)')).toBeInTheDocument()

    // Die Fläche sagt nichts, was der Text nicht sagt — und wird deshalb nicht
    // ein zweites Mal vorgelesen.
    expect(container.querySelectorAll('[aria-hidden="true"]')).toHaveLength(2)
  })

  // **Eine Wahl für dieselbe Frage** — „wonach zählen?". Vorher waren es zwei
  // Auswahlfelder mit eigener Überschrift und die Belegung als dritter
  // Abschnitt; das Artboard zeigt ein Feld „Worüber".
  it('stellt Belegung, Felder und Dimensionen in eine Wahl', async () => {
    belegeApi({
      Statistik: () => Promise.resolve({ statistik: zahlen }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: () =>
        Promise.resolve({
          dimensionen: [{ name: 'rarity', beschreibung: '', numerisch: false }],
          fehler: null,
        }),
    })
    zeige('7')

    await screen.findByText('8 (67 %)')

    expect(screen.getAllByRole('combobox')).toHaveLength(1)
    expect(worueber()).toHaveValue('belegung')

    const felder = await screen.findByRole('group', {
      name: texte.gruppeFelder,
    })
    expect(
      within(felder)
        .getAllByRole('option')
        .map((eine) => eine.textContent),
    ).toEqual(['name', 'typ'])

    const spielpaket = await screen.findByRole('group', {
      name: texte.gruppeSpielpaket,
    })
    expect(within(spielpaket).getByRole('option')).toHaveTextContent('rarity')
  })

  // Die Verteilung wird erst geholt, wenn ein Feld gewählt ist — und dann für
  // genau dieses. **Und sie steht an der Stelle der Belegung**, nicht unter
  // einem anderen Abschnitt.
  it('holt die Verteilung des gewählten Felds und zeigt nur sie', async () => {
    const Verteilung = vi.fn(() =>
      Promise.resolve({
        verteilung: [{ name: 'Zauber', karten: 6, kartenIDs: [] }],
      }),
    )

    belegeApi({
      Statistik: () => Promise.resolve({ statistik: zahlen }),
      Verteilung,
      Auszaehldimensionen: keineDimensionen,
    })
    zeige('7')

    await screen.findByText('12')
    expect(Verteilung).not.toHaveBeenCalled()
    expect(
      screen.getByRole('heading', { name: texte.belegung }),
    ).toBeInTheDocument()

    await userEvent.selectOptions(worueber(), 'feld:typ')

    await waitFor(() => {
      expect(Verteilung).toHaveBeenCalledExactlyOnceWith('7', 'typ')
    })

    expect(await screen.findByText('Zauber')).toBeInTheDocument()
    expect(
      screen.getByRole('heading', {
        name: einsetzen(texte.verteilungVon, { feld: 'typ' }),
      }),
    ).toBeInTheDocument()

    // **Ein Bild und nicht zwei**: die Belegung ist fort, solange das Feld
    // gewählt ist.
    expect(
      screen.queryByRole('heading', { name: texte.belegung }),
    ).not.toBeInTheDocument()
    expect(screen.queryByText('8 (67 %)')).not.toBeInTheDocument()
  })

  it('erklärt eine leere Verteilung', async () => {
    belegeApi({
      Statistik: () => Promise.resolve({ statistik: zahlen }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: keineDimensionen,
    })
    zeige('7')

    await screen.findByText('12')
    await userEvent.selectOptions(worueber(), 'feld:typ')

    expect(await screen.findByText(texte.verteilungLeer)).toBeInTheDocument()
  })

  it('erklärt ein Set ohne einen einzigen Feldwert', async () => {
    belegeApi({
      Statistik: () =>
        Promise.resolve({
          statistik: { karten: 0, felder: 0, leere: 0, belegung: [] },
        }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: keineDimensionen,
    })
    zeige('7')

    expect(await screen.findByText(texte.belegungLeer)).toBeInTheDocument()

    // **Ohne Felder gibt es auch keine Gruppe dafür** — eine leere Gruppe in
    // der Wahl wäre ein Eintrag, der nichts anbietet.
    expect(
      screen.queryByRole('group', { name: texte.gruppeFelder }),
    ).not.toBeInTheDocument()
  })

  it('meldet einen Fehler aus der Bindung', async () => {
    belegeApi({
      Statistik: () =>
        Promise.resolve({
          statistik: { karten: 0, felder: 0, leere: 0, belegung: [] },
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Die Zahlen kamen nicht.',
          },
        }),
      Verteilung: () => Promise.resolve({ verteilung: [] }),
      Auszaehldimensionen: keineDimensionen,
    })
    zeige('7')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      // **Der Satz aus dem Katalog** (#133): übersetzt wird über den
      // Schlüssel, nicht über den Satz aus Go.
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Ein Wechsel des Sets räumt die Wahl mit** (§7, Regel 1).
  //
  // Ohne das ginge die nächste Frage als `Verteilung(neuesSet, altesFeld)`
  // hinaus — ein Feld, das das neue Set vielleicht gar nicht führt.
  it('räumt die Wahl, wenn das Set wechselt', async () => {
    const Verteilung = vi.fn(() => Promise.resolve({ verteilung: [] }))

    belegeApi({
      Statistik: () => Promise.resolve({ statistik: zahlen }),
      Verteilung,
      Auszaehldimensionen: keineDimensionen,
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Statistik setID="7" />
      </Sprachraum>,
    )

    await screen.findByText('8 (67 %)')
    await userEvent.selectOptions(worueber(), 'feld:typ')

    await waitFor(() => {
      expect(Verteilung).toHaveBeenCalledWith('7', 'typ')
    })

    rerender(
      <Sprachraum sprache="de">
        <Statistik setID="8" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(worueber()).toHaveValue('belegung')
    })

    // Und keine Frage nach dem alten Feld im neuen Set.
    expect(Verteilung).not.toHaveBeenCalledWith('8', 'typ')
  })
})
