// **Die Massenwahl: anhaken, abhaken, und was ein Setwechsel damit macht.**
//
// Von `kartenwahl.test.tsx` getrennt, als die Datei über 400 Zeilen wuchs
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort, welche Karte
// aufgeschlagen ist, hier, welche **angehakt** sind.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { type Api } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'
import { waehleMehrere } from '../test/kartenlasche'

// belegeSeite belegt zusätzlich die Stilauswahl.
//
// **Weil die Kartenseite sie zeigt** (#41). Ohne Beleg meldete die Fassade „die
// Methode ist nicht belegt", und ein Test über die Kartenliste fände zwei
// Warnungen statt einer — genau dafür ist die Vorrichtung streng (§3).
function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    ...teile,
  })
}

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeKarte(teile: Partial<KarteDTO> = {}): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Feuerball' }],
    ...teile,
  }
}

describe('Die Massenwahl', () => {
  // **Die Anhakung führt in die Massenänderung** — und ein Setwechsel räumt sie
  // mit (§7, Regel 1): eine Massenänderung über Karten eines Sets, das nicht
  // mehr offen ist, ginge an ein fremdes Set.
  it('hakt Karten an und räumt die Anhakung beim Setwechsel', async () => {
    belegeSeite({
      // **Je Set ein eigener Name**, nicht nur eine eigene Kennung: an ihm
      // ist unten zu sehen, dass die Liste des neuen Sets angekommen ist
      // (#661).
      ...ausKarten((setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Feuerball ${setID}` }],
            }),
          ],
        }),
      ),
      Karten: (setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Feuerball ${setID}` }],
            }),
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    // **Erst in den Auswahlmodus** — seit UXR-05 (#1913) trägt die Liste
    // ihre Kästchen nur dort.
    await waehleMehrere()

    const anhaken = await screen.findByRole('checkbox', {
      name: new RegExp(texte.seiten.karten.massen.auswaehlen),
    })

    await userEvent.click(anhaken)

    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: texte.seiten.karten.massen.anwenden,
        }),
      ).toBeInTheDocument()
    })

    // **Abhaken nimmt den Abschnitt weg.** Seit
    // #288
    // erscheint die Massenänderung erst, wenn etwas angehakt ist — ein
    // Abschnitt, der bei jedem Blick dasteht und meistens leer ist, kostet
    // Platz und Aufmerksamkeit.
    await userEvent.click(anhaken)

    expect(
      screen.queryByRole('heading', {
        name: texte.seiten.karten.massen.ueberschrift,
      }),
    ).not.toBeInTheDocument()

    await userEvent.click(anhaken)
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: texte.seiten.karten.massen.anwenden,
        }),
      ).toBeInTheDocument()
    })

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="2"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    // **Erst der Beleg, dass die Antwort da ist, dann die Verneinung** (§3,
    // #661).
    // Unmittelbar nach dem Wechsel steht noch die Liste des vorigen Sets —
    // `useLaden` lässt sie stehen, während der neue Schlüssel lädt. Die
    // Zusicherung traf damit schon zu, bevor irgendetwas geschehen war, und
    // `waitFor` gab beim ersten Versuch zurück.
    await screen.findByRole('button', {
      name: `Feuerball 2 — ${texte.seiten.karten.bearbeiten}`,
    })

    expect(
      screen.queryByRole('heading', {
        name: texte.seiten.karten.massen.ueberschrift,
      }),
    ).not.toBeInTheDocument()
  })

  // **Was der Fall darüber nicht belegt**: dass die Haken *geräumt* wurden.
  // Nach dem Wechsel steht eine andere Karte in der Liste, und ein Haken auf
  // einer Kennung, die es nicht mehr gibt, wählt ohnehin nichts aus. Der Fall
  // hier ist der, der das Räumen festhält — er wechselt zurück.
  //
  // **Der Rückwechsel auf dasselbe Set holt die alten Haken nicht zurück.**
  //
  // Genau diesen Weg ging der erste Test nicht: er wechselte hin und nie
  // zurück, und deshalb blieb er grün, als der Wächter entfernt wurde.
  it('räumt die Anhakung, auch beim Rückwechsel', async () => {
    belegeSeite({
      ...ausKarten((setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
            }),
          ],
        }),
      ),
      Karten: (setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
            }),
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await waehleMehrere()
    await userEvent.click(
      await screen.findByRole('checkbox', {
        name: new RegExp(texte.seiten.karten.massen.auswaehlen),
      }),
    )

    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: texte.seiten.karten.massen.anwenden,
        }),
      ).toBeInTheDocument()
    })

    for (const set of ['2', '1']) {
      rerender(
        <Sprachraum sprache="de">
          <Karten
            karteGewaehlt={vi.fn()}
            setID={set}
            karteID={undefined}
            verlaufsstand={0}
            verlaufsfehler={undefined}
            bestandGeaendert={() => undefined}
          />
        </Sprachraum>,
      )

      // **Erst warten, bis die neue Liste da ist — und dann neu suchen.**
      //
      // Zwei Anläufe gingen hier daneben: der erste prüfte „keine Auswahl"
      // sofort und traf das Ladefenster, in dem die Liste ohnehin leer ist; der
      // zweite hielt das Kästchen der **alten** Liste in der Hand, ein
      // abgehängtes Element, das seinen Haken behält. Beide waren grün, ohne
      // ihren Fall zu sehen.
      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: new RegExp(`Karte aus ${set}`) }),
        ).toBeInTheDocument()
      })

      // **Der Wechsel räumt den Auswahlmodus mit** (UXR-05, #1913) — die
      // Spalte wird dabei abgebaut. Dass kein alter Haken zurückkam, zeigt
      // erst das Kästchen, also wird wieder umgeschaltet.
      expect(
        screen.queryByRole('checkbox', {
          name: new RegExp(texte.seiten.karten.massen.auswaehlen),
        }),
      ).not.toBeInTheDocument()
      await waehleMehrere()
      expect(
        screen.getByRole('checkbox', {
          name: new RegExp(texte.seiten.karten.massen.auswaehlen),
        }),
      ).not.toBeChecked()
      expect(
        screen.queryByRole('heading', {
          name: texte.seiten.karten.massen.ueberschrift,
        }),
      ).not.toBeInTheDocument()
    }
  })
})
