// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// **Am Bauteil und nicht an der Seite**: dass das Gebiet „Bestand" es zeigt,
// prüft `appbestandspflege.test.tsx`; hier steht, **was** es zeigt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Ruecksprung } from './Ruecksprung'
import { alsZeitpunkt } from '../zeitpunkt'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { stillerRuecksprung } from '../test/bestandspflegeprobe'

import type { RueckkehrDTO } from '../apirueckkehr'
import type { Ruecksprungstand } from '../haken/useRuecksprung'

const worte = kataloge.de.seiten.ruecksprung

// satz holt einen Eintrag aus einer Marken-Tafel.
//
// **Und scheitert laut, wenn es ihn nicht gibt.** Ein `?? ''` daneben machte
// aus einem fehlenden Katalogeintrag eine leere Erwartung — und die trifft
// immer zu (§10).
function satz(tafel: Record<string, string>, schluessel: string): string {
  const wert = tafel[schluessel]

  if (wert === undefined) {
    throw new Error(`der Katalog kennt ${schluessel} nicht`)
  }

  return wert
}

function eineRueckkehr(teil: Partial<RueckkehrDTO> = {}): RueckkehrDTO {
  return {
    id: 7,
    zurueckgespielt: '2026-09-10T12:00:00Z',
    datenstand: '2026-09-01T08:30:00Z',
    quelle: 'bestand-2026-09-01.msesicherung',
    klaerung: 'offen',
    geklaertAm: '',
    mengen: [
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
    ...teil,
  }
}

function zeige(teile: Partial<Ruecksprungstand> = {}) {
  render(
    <Sprachraum sprache="de">
      <Ruecksprung stand={stillerRuecksprung(teile)} />
    </Sprachraum>,
  )
}

describe('Rücksprung', () => {
  // **Zwei der vier kann diese Anwendung nicht selbst** — und das ist kein
  // Mangel, sondern eine Aussage, die dasteht.
  it('nennt zu jeder Art, was ihr im Weg steht', () => {
    zeige({
      arten: [
        { art: 'appfassung', hindernis: 'nachProzessende' },
        { art: 'daten', hindernis: '' },
      ],
    })

    expect(
      screen.getByText(satz(worte.arten, 'appfassung')),
    ).toBeInTheDocument()
    expect(
      screen.getByText(satz(worte.hindernisse, 'nachProzessende')),
    ).toBeInTheDocument()

    expect(screen.getByText(satz(worte.arten, 'daten'))).toBeInTheDocument()
    expect(screen.getByText(worte.gehtHier)).toBeInTheDocument()
  })

  // **Eine Marke, die der Katalog nicht kennt, steht als Marke da** — sie ist
  // immer noch eine Auskunft; ein leerer Platz wäre keine.
  it('zeigt eine unbekannte Marke, statt sie zu verschlucken', () => {
    zeige({ arten: [{ art: 'zeitreise', hindernis: 'geht nicht' }] })

    expect(screen.getByText('zeitreise')).toBeInTheDocument()
    expect(screen.getByText('geht nicht')).toBeInTheDocument()
  })

  it('erklärt den leeren Stand, statt ihn leer zu lassen', () => {
    zeige()

    expect(screen.getByText(worte.keineRueckkehren)).toBeInTheDocument()
  })

  it('nennt eine offene Rückkehr samt ihrer Menge', () => {
    zeige({ rueckkehren: [eineRueckkehr()] })

    // **Lesbar und nicht roh** (Sichtprüfung zu O7): geprüft wird die
    // gerechnete Form, denn ein Literal kennt die Zone dieses Rechners nicht.
    expect(
      screen.getByText(
        new RegExp(`Am ${alsZeitpunkt('2026-09-10T12:00:00Z', 'de')} zurück`),
      ),
    ).toBeInTheDocument()

    expect(
      screen.getByText(/bestand-2026-09-01\.msesicherung/),
    ).toBeInTheDocument()

    expect(screen.getByText(/betrifft 3 Stück/)).toBeInTheDocument()
  })

  it('quittiert eine offene Rückkehr auf Klick', async () => {
    const quittieren = vi.fn()
    const nutzer = userEvent.setup()

    zeige({ rueckkehren: [eineRueckkehr()], quittieren })

    await nutzer.click(screen.getByRole('button', { name: worte.quittieren }))

    expect(quittieren).toHaveBeenCalledWith(7)
  })

  it('sperrt den Knopf, solange die Quittung läuft', () => {
    zeige({
      rueckkehren: [eineRueckkehr()],
      laeuft: (kennung) => kennung === 7,
    })

    expect(
      screen.getByRole('button', { name: worte.quittierenLaeuft }),
    ).toBeDisabled()
  })

  // **Kein Knopf bei einer quittierten** — sonst wäre es einer, der sicher
  // nichts tut.
  it('lässt den Knopf bei einer quittierten Rückkehr weg', () => {
    zeige({
      rueckkehren: [
        eineRueckkehr({
          klaerung: 'quittiert',
          geklaertAm: '2026-09-10T13:00:00Z',
        }),
      ],
    })

    expect(
      screen.queryByRole('button', { name: worte.quittieren }),
    ).not.toBeInTheDocument()

    expect(
      screen.getByText(
        new RegExp(
          `Am ${alsZeitpunkt('2026-09-10T13:00:00Z', 'de')} nachgesehen`,
        ),
      ),
    ).toBeInTheDocument()
  })

  // **Eine Rückkehr ohne Mengen nennt keine Zahl** — „betrifft 0 Stück" wäre
  // eine Messung, die nicht stattgefunden hat.
  it('nennt keine Menge, wenn keine unklar ist', () => {
    zeige({ rueckkehren: [eineRueckkehr({ mengen: [] })] })

    expect(screen.queryByText(/betrifft/)).not.toBeInTheDocument()
    expect(screen.getByText(worte.offen)).toBeInTheDocument()
  })

  it('zeigt einen Fehler als Meldung', () => {
    zeige({ fehler: 'das Buch ist zu' })

    expect(screen.getByRole('alert')).toHaveTextContent('das Buch ist zu')
  })
})
