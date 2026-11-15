// **Die Dienstschalter auf der Einstellungsseite** — Werkstück 2 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Einstellungen } from './Einstellungen'
import { Sprachraum } from '../sprache/kontext'
import { zuGebiet } from '../test/gebiete'
import { kataloge } from '../sprache/kataloge'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from '../test/bestandspflegeprobe'
import type { EinstellungDTO } from '../api'

const texte = kataloge.de.seiten.einstellungen

const dienste: EinstellungDTO[] = [
  {
    schluessel: 'dienst.scryfall',
    wert: 'an',
    wirksam: 'an',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
  {
    schluessel: 'dienst.scrydex',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
]

async function hinstellen(
  eintraege: EinstellungDTO[] = dienste,
  setze: (schluessel: string, wert: string) => void = vi.fn(),
) {
  render(
    <Sprachraum sprache="de">
      <Einstellungen
        eintraege={[]}
        dienste={eintraege}
        fragen={[]}
        ki={[]}
        vorgaben={[]}
        anmeldung={[]}
        ursache={undefined}
        meldefehler={undefined}
        setze={setze}
        laeuft={() => false}
        bestandspflege={stillePflege()}
        ruecksprung={stillerRuecksprung()}
        pflege={stilleDatenpflege()}
        setID={undefined}
      />
    </Sprachraum>,
  )

  // **Die Dienste liegen in ihrem Gebiet** — Werkstück I von Phase 4.10
  // (#1325).
  await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.dienste)
}

describe('die Dienstschalter', () => {
  it('stehen in einem eigenen Bereich', async () => {
    await hinstellen()

    const bereich = screen.getByRole('region', { name: texte.dienste })

    expect(within(bereich).getByLabelText('Scryfall')).toBeInTheDocument()
    expect(within(bereich).getByLabelText('Scrydex')).toBeInTheDocument()
  })

  // **Der Name des Dienstes wird nicht übersetzt** (0063, Entscheidung 5) —
  // „Scryfall" und „Scrydex" sind Eigennamen. Übersetzt werden die Zustände.
  it('nennt die Dienste bei ihrem Eigennamen', async () => {
    await hinstellen()

    const bereich = screen.getByRole('region', { name: texte.dienste })

    expect(within(bereich).getByText('Scryfall')).toBeInTheDocument()
  })

  // **Ein Dienst, der nie gegen sein Original gelaufen ist, sagt es dort, wo
  // man ihn einschaltet** (Vermerk 0056, Entscheidung 4; Abnahme 5 des
  // Issues). Nicht im Kleingedruckten — an der Stelle.
  // **Und das Umlegen kommt an.**
  //
  // Ohne diesen Lauf blieben Auswahl und Rückruf ungeprüft: die Schalter
  // stünden da und täten nichts, und die Tests darüber wären grün.
  it('reicht das Umlegen durch', async () => {
    const setze = vi.fn()

    await hinstellen(dienste, setze)

    const bereich = screen.getByRole('region', { name: texte.dienste })

    // **Ein Schalter und keine Auswahl** — UXR-F2
    // (#2008).
    await userEvent.click(
      within(bereich).getByRole('switch', { name: 'Scrydex' }),
    )

    expect(setze).toHaveBeenCalledWith('dienst.scrydex', 'an')

    await userEvent.click(
      within(bereich).getByRole('switch', { name: 'Scryfall' }),
    )

    expect(setze).toHaveBeenLastCalledWith('dienst.scryfall', 'aus')
  })

  // **Ein Dienst ohne Eigennamen fällt auf seinen Schlüssel zurück.**
  //
  // Er kann nur entstehen, wenn Go einen fünften Dienst kennt und das Fenster
  // noch nicht — dann steht dort `dienst.neuer` statt einer leeren Zeile.
  // **Und magicthegathering.io gilt als geprüft.**
  //
  // Es ist der **einzige** der drei neuen, für den es einen Lauf gegen das
  // Original gibt (`internal/mtgapi/netzlauf_test.go`) — und er hat beim
  // ersten Mal die Drossel angestrichen.
  //
  // **Gefunden hat den Fehler die Sichtprüfung am laufenden Fenster**, nicht
  // ein Test: dort stand „nie gegen den echten Dienst gelaufen" ausgerechnet
  // an dem Dienst, für den die Zusage am stärksten ist.
  it('nennt einen unbekannten Dienst bei seinem Schlüssel', async () => {
    await hinstellen([
      {
        schluessel: 'dienst.neuer',
        wert: 'aus',
        wirksam: 'aus',
        werte: ['aus', 'an'],
        bereich: 'dienste',
      },
    ])

    expect(
      within(
        screen.getByRole('region', { name: texte.dienste }),
      ).getByLabelText('dienst.neuer'),
    ).toBeInTheDocument()
  })

  // **Ohne Dienste gibt es die Lasche gar nicht** — Werkstück I von Phase
  // 4.10 (#1325).
  // Vorher gab der Schalter `null` zurück und der Block fehlte; als Gebiet
  // hinter einer Lasche wäre das ein Bedienelement, das etwas verspricht, was
  // nicht kommt — dieselbe Regel wie beim Kasten „Nicht gedeutet".
  it('führt kein Gebiet, wo keine Dienste sind', () => {
    render(
      <Sprachraum sprache="de">
        <Einstellungen
          eintraege={[]}
          dienste={[]}
          fragen={[]}
          ki={[]}
          vorgaben={[]}
          anmeldung={[]}
          ursache={undefined}
          meldefehler={undefined}
          setze={vi.fn()}
          laeuft={() => false}
          bestandspflege={stillePflege()}
          ruecksprung={stillerRuecksprung()}
          pflege={stilleDatenpflege()}
          setID={undefined}
        />
      </Sprachraum>,
    )

    expect(screen.queryByRole('tab', { name: texte.dienste })).toBeNull()
    expect(screen.queryByRole('region', { name: texte.dienste })).toBeNull()
  })
})
