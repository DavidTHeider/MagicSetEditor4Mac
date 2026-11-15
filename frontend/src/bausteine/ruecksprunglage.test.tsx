// **Was ein Zurückspielen kostet** — Werkstück O7 von Phase 5.2
// (#1431).
//
// **Am Bauteil und nicht an der Seite**: es stellt dar, was der Haken geholt
// hat. Dass die Lage überhaupt in der Rückfrage steht, prüft
// `appbestandspflege.test.tsx`.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Ruecksprunglage } from './Ruecksprunglage'
import { alsZeitpunkt } from '../zeitpunkt'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { RuecksprunglageAntwort } from '../apirueckkehr'

const worte = kataloge.de.seiten.ruecksprung

function lage(
  teil: Partial<RuecksprunglageAntwort> = {},
): RuecksprunglageAntwort {
  return {
    datenstand: '2026-09-01T08:30:00Z',
    zeilen: [
      {
        gegenstand: 'sets',
        inSicherung: 5,
        jetzt: 6,
        verlust: 1,
        unbekannt: false,
      },
      {
        gegenstand: 'karten',
        inSicherung: 300,
        jetzt: 312,
        verlust: 12,
        unbekannt: false,
      },
    ],
    unklar: [],
    handpruefungNoetig: false,
    fehler: null,
    ...teil,
  }
}

function zeige(gezeigt: RuecksprunglageAntwort) {
  render(
    <Sprachraum sprache="de">
      <Ruecksprunglage lage={gezeigt} />
    </Sprachraum>,
  )
}

describe('Rücksprunglage', () => {
  it('nennt den Datenstand und was verloren geht', () => {
    zeige(lage())

    // **Lesbar und nicht roh** — die Sichtprüfung zu O7 fand die ISO-Angabe
    // wörtlich in der Rückfrage. Geprüft wird die **gerechnete** Form, nicht
    // eine abgeschriebene: `alsZeitpunkt` kennt die Zone dieses Rechners, ein
    // Literal hier nicht.
    expect(
      screen.getByText(
        `${worte.datenstand}: ${alsZeitpunkt('2026-09-01T08:30:00Z', 'de')}`,
      ),
    ).toBeInTheDocument()

    // **Die Zeile nennt beide Seiten und die Differenz** — „es geht etwas
    // verloren" ist keine Auskunft, mit der jemand entscheiden kann.
    const karten = screen.getByRole('row', { name: /Karten/ })

    expect(karten).toHaveTextContent('300')
    expect(karten).toHaveTextContent('312')
    expect(karten).toHaveTextContent('12')
  })

  // **−1 heißt „diese Seite sagt nichts darüber"** und ist etwas anderes als
  // 0. Eine Zahl, die nicht stimmt, wäre schlimmer als ein Strich.
  it('zeigt für eine unbekannte Zahl einen Strich statt einer Null', () => {
    zeige(
      lage({
        zeilen: [
          {
            gegenstand: 'decks',
            inSicherung: -1,
            jetzt: 3,
            verlust: 0,
            unbekannt: true,
          },
        ],
      }),
    )

    const decks = screen.getByRole('row', { name: /Decks/ })

    expect(decks).toHaveTextContent(worte.unbekannt)
    expect(decks).not.toHaveTextContent('-1')
  })

  // **Ohne Datenstand steht keine Zeile da** — eine leere Zeitangabe wäre
  // schlimmer als keine.
  it('lässt den Datenstand weg, wenn keiner feststeht', () => {
    zeige(lage({ datenstand: '' }))

    expect(
      screen.queryByText(/Wiederhergestellter Stand/),
    ).not.toBeInTheDocument()
  })

  it('nennt die unklaren Mengen mit ihrem Namen', () => {
    zeige(
      lage({
        unklar: [
          {
            schluessel: 'scryfall:abc',
            zustand: 'nm',
            sprache: 'en',
            finish: 'nonfoil',
            menge: 3,
            name: 'Lightning Bolt',
            setcode: '',
            nummer: '',
          },
        ],
      }),
    )

    expect(screen.getByText(worte.unklarTitel)).toBeInTheDocument()
    expect(screen.getByText('3 × Lightning Bolt')).toBeInTheDocument()
  })

  // **Ohne Namen steht der Schlüssel da** — eine Kennung ist eine Auskunft,
  // eine leere Zeile keine.
  it('nennt den Schlüssel, wenn kein Name da ist', () => {
    zeige(
      lage({
        unklar: [
          {
            schluessel: 'druck:4',
            zustand: 'nm',
            sprache: 'en',
            finish: 'foil',
            menge: 2,
            name: '',
            setcode: '',
            nummer: '',
          },
        ],
      }),
    )

    expect(screen.getByText('2 × druck:4')).toBeInTheDocument()
  })

  // **Ein Gegenstand, den der Katalog nicht kennt, steht als Marke da** —
  // dieselbe Entscheidung wie bei den Rückkehrarten: eine Marke ist immer noch
  // eine Auskunft, ein leerer Platz keine.
  it('zeigt einen unbekannten Gegenstand, statt ihn zu verschlucken', () => {
    zeige(
      lage({
        zeilen: [
          {
            gegenstand: 'schallplatten',
            inSicherung: 1,
            jetzt: 2,
            verlust: 1,
            unbekannt: false,
          },
        ],
      }),
    )

    expect(
      screen.getByRole('row', { name: /schallplatten/ }),
    ).toBeInTheDocument()
  })

  it('sagt es, wenn von Hand nachzusehen ist', () => {
    zeige(lage({ handpruefungNoetig: true }))

    expect(screen.getByText(worte.handpruefung)).toBeInTheDocument()
  })

  // **Die Gegenprobe**: ohne Befund steht der Satz nicht da — sonst stünde er
  // immer und sagte nichts (§10).
  it('sagt nichts von Handprüfung, wenn keine nötig ist', () => {
    zeige(lage())

    expect(screen.queryByText(worte.handpruefung)).not.toBeInTheDocument()
    expect(screen.queryByText(worte.unklarTitel)).not.toBeInTheDocument()
  })
})
