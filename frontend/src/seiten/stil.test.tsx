// Die Lasche „Stil“ (#288).
//
// Sie ist die Hülle um [Stilwahl] — geprüft wird, was **sie** zusagt: ohne Set
// ein erklärter leerer Zustand, mit Set die Stilwahl, und die Meldung an das
// Fenster kommt durch.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stil } from './Stil'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

describe('Stil', () => {
  // **Der Fall „kein Set" steht hier nicht mehr.** Seit
  // #288
  // rendert der Arbeitsplatz diese Lasche nur mit offenem Set.

  it('zeigt mit Set die Stilwahl', async () => {
    belegeApi({
      Stile: () =>
        Promise.resolve({
          stile: [
            {
              paket: 'probe-schlicht',
              anzeigename: 'Probe-schlicht',
              spiel: 'probe',
              version: '2026-08-15',
              eigenesSpiel: true,
              gewaehlt: true,
            },
          ],
        }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
    })

    render(
      <Sprachraum sprache="de">
        <Stil setID="7" geaendert={vi.fn()} />
      </Sprachraum>,
    )

    // **Eine Auswahl und keine Gruppe von Optionsfeldern** — seit UXR-07
    // (#1915); die Gruppe zeichnete je installiertem Stil ein Feld.
    expect(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
    ).toBeInTheDocument()
  })
})

// **Die Vorschau steht neben den Stiloptionen, nicht darunter** — Werkstück C
// von Phase 4.10
// (#1325).
//
// # Was vorher war
//
// Sie stand ganz unten, nach allen Optionen. Gemessen am 2026-09-03: die
// Lasche ist **5,6 Bildschirme** hoch und trägt 98 Tabstellen — wer eine
// Option ändert, sah die Wirkung erst nach dem Scrollen.
//
// §6 sagt: *„Eine Änderung an einem Feld ist in der Vorschau zu sehen, ohne
// dass jemand etwas bestätigt."* Sie **war** zu sehen, nur nicht von dort, wo
// man tippt.
//
// # Warum die DOM-Reihenfolge geprüft wird und nicht die Optik
//
// **jsdom hat kein Layout** — eine Zusicherung über Spalten wäre hier eine
// Behauptung. Was prüfbar ist und zählt: die Vorschau steht **vor** den
// Optionen im Baum. Damit landet sie auch dann oben, wenn das Gitter bei
// einem schmalen Fenster auf eine Spalte umbricht — und ein
// Vorleseprogramm nennt sie zuerst.
//
// **Dass sie im breiten Fenster daneben steht**, entscheidet das Gitter in
// `Stilwahl.tsx`; geprüft ist es an der laufenden Anwendung.
it('nennt die Vorschau vor den Stiloptionen', async () => {
  belegeApi({
    Stile: () =>
      Promise.resolve({
        stile: [
          {
            paket: 'probe-schlicht',
            anzeigename: 'Probe-schlicht',
            spiel: 'probe',
            version: '2026-08-15',
            eigenesSpiel: true,
            gewaehlt: true,
          },
        ],
      }),
    // **Mit einer Option** — ohne sie zeichnet `Stiloptionen` nichts, und der
    // Test prüfte die Reihenfolge zweier Überschriften, von denen es eine
    // nicht gibt.
    Stiloptionen: () =>
      Promise.resolve({
        optionen: [
          {
            name: 'hintergrund',
            anzeigename: 'Hintergrund',
            art: 'choice',
            beschreibung: 'Der Grund hinter der Karte.',
            wert: 'hell',
            moeglichkeiten: [
              { wert: 'hell', beschriftung: 'hell' },
              { wert: 'dunkel', beschriftung: 'dunkel' },
            ],
            eigeneFarbe: false,
            gesetzt: false,
            hatVorgabe: false,
            hatWahlbilder: false,
            vorgabe: '',
          },
        ],
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Stil setID="7" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  const vorschau = await screen.findByRole('heading', {
    name: texte.seiten.stil.vorschauTitel,
  })
  // **Beide über `findBy`.** Die Optionen kommen über einen eigenen Ruf; ein
  // `getBy` griffe, bevor er beantwortet ist (§3).
  const optionen = await screen.findByRole('heading', {
    name: texte.seiten.karten.stiloptionen,
  })

  expect(
    vorschau.compareDocumentPosition(optionen) &
      Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy()
})
