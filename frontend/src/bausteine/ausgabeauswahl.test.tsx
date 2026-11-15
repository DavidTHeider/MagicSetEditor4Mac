// **Welche Karten hinausgehen** — Werkstück 4 von Phase 3.6
// (#1167).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgabeauswahl } from './Ausgabeauswahl'

const texte = kataloge.de.bausteine.ausgabeauswahl

const zeilen = [
  { id: '1', nummer: 0, name: 'Erste', felder: 2 },
  { id: '2', nummer: 1, name: 'Zweite', felder: 2 },
  { id: '3', nummer: 2, name: '', felder: 0 },
]

function zeige(gewaehlt: readonly string[] = [], angehalten = false) {
  const setzeGewaehlt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Ausgabeauswahl
        setID="1"
        gewaehlt={gewaehlt}
        setzeGewaehlt={setzeGewaehlt}
        angehalten={angehalten}
      />
    </Sprachraum>,
  )

  return setzeGewaehlt
}

afterEach(() => {
  raeumeApi()
})

describe('die Ausgabeauswahl', () => {
  // **Ohne Auswahl steht „das ganze Set" da — samt Zahl.**
  //
  // MSE2 zeigt die Zahl ebenso (`ExportWindowBase::update` setzt
  // `card_count`); ohne sie wählte man zwischen zwei Wörtern.
  it('nennt das ganze Set mit seiner Kartenzahl', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    zeige()

    expect(
      await screen.findByRole('radio', { name: /Das ganze Set \(3 Karten\)/ }),
    ).toBeChecked()
  })

  // **Die Teilmenge wählt die erste Karte vor.** Eine Teilmenge ohne Karte
  // wäre eine Wahl, die keine Ausgabe zulässt.
  it('wählt beim Umschalten die erste Karte vor', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    const setzeGewaehlt = zeige()

    await userEvent.click(
      await screen.findByRole('radio', { name: /Nur ausgewählte/ }),
    )

    expect(setzeGewaehlt).toHaveBeenCalledWith(['1'])
  })

  // **Mit Auswahl steht die Liste da**, und ein Haken nimmt eine Karte dazu.
  it('nimmt eine angehakte Karte dazu und wieder heraus', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    const setzeGewaehlt = zeige(['1'])

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Zweite' }),
    )

    expect(setzeGewaehlt).toHaveBeenCalledWith(['1', '2'])

    await userEvent.click(screen.getByRole('checkbox', { name: 'Erste' }))

    expect(setzeGewaehlt).toHaveBeenCalledWith([])
  })

  // **Eine Karte ohne Namen bekommt einen Ersatz** — sonst stünde ein Haken
  // ohne Beschriftung da, und niemand könnte ihn benennen (§6).
  it('nennt eine Karte ohne Namen', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    zeige(['1'])

    expect(
      await screen.findByRole('checkbox', { name: texte.ohneNamen }),
    ).toBeInTheDocument()
  })

  // **Während die Ausgabe läuft, ist die Wahl gesperrt** (§7, Regel 4).
  it('sperrt die Wahl, solange die Ausgabe läuft', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    zeige(['1'], true)

    expect(
      await screen.findByRole('radio', { name: /Das ganze Set/ }),
    ).toBeDisabled()

    // **Gewartet, nicht gegriffen** (§3, #1321): der Radioknopf steht sofort
    // da, das Kästchen entsteht erst mit den Kartenzeilen. Unter Last war
    // dieser Griff einer von 2.186, der ins Leere ging.
    expect(
      await screen.findByRole('checkbox', { name: 'Erste' }),
    ).toBeDisabled()
  })

  // **Und zurück auf das ganze Set.** Der Weg heraus aus der Teilmenge ist
  // derselbe Knopf, mit dem man hineingekommen ist — ohne ihn bliebe eine
  // einmal getroffene Auswahl bestehen, bis jemand jeden Haken einzeln
  // entfernt.
  it('schaltet zurück auf das ganze Set', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen, fehler: null }),
    })

    const setzeGewaehlt = zeige(['1'])

    await userEvent.click(
      await screen.findByRole('radio', { name: /Das ganze Set/ }),
    )

    expect(setzeGewaehlt).toHaveBeenCalledWith([])
  })

  // **Ein leeres Set lässt keine Teilmenge zu.** Eine Wahl, aus der nichts
  // folgen kann, wäre ein Knopf, der nur einen Fehler erzeugt.
  it('lässt bei einem leeren Set keine Teilmenge zu', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    zeige()

    expect(
      await screen.findByRole('radio', { name: /Nur ausgewählte/ }),
    ).toBeDisabled()
  })

  // **Und ein Lesefehler steht am Kasten.**
  it('zeigt einen Lesefehler', async () => {
    belegeApi({
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'der Bestand ist belegt',
          },
        }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })
})
