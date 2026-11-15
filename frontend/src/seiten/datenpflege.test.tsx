// **Nachsehen, wählen, aufräumen** — Werkstück O8 von Phase 5.2
// (#1432).
//
// **Am Bauteil und nicht an der Seite**: dass das Gebiet „Bestand" es zeigt,
// prüft `appdatenpflege.test.tsx`; hier steht, **was** es zeigt und was die
// Klasse an der Vorauswahl ändert.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Datenpflege } from './Datenpflege'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { stilleDatenpflege } from '../test/bestandspflegeprobe'

import type { PflegebefundDTO, PflegevorschauAntwort } from '../apipflege'
import type { Pflegelauf } from '../haken/useDatenpflege'

const worte = kataloge.de.seiten.datenpflege

// satz holt einen Eintrag aus einer Marken-Tafel — und scheitert laut, wenn
// es ihn nicht gibt (§10: ein `?? ''` träfe immer zu).
function satz(tafel: Record<string, string>, schluessel: string): string {
  const wert = tafel[schluessel]

  if (wert === undefined) {
    throw new Error(`der Katalog kennt ${schluessel} nicht`)
  }

  return wert
}

function befund(teil: Partial<PflegebefundDTO> = {}): PflegebefundDTO {
  return {
    art: 'kartenwerteOhneKarte',
    klasse: 'sicher',
    anzahl: 3,
    revision: 'abc',
    beispiele: ['card_values:1:name'],
    ...teil,
  }
}

function vorschau(
  teil: Partial<PflegevorschauAntwort> = {},
): PflegevorschauAntwort {
  return {
    erstellt: '2026-09-10T12:00:00Z',
    gueltigbis: '2026-09-10T12:15:00Z',
    befunde: [befund()],
    fehler: null,
    ...teil,
  }
}

function zeige(teile: Partial<Pflegelauf> = {}) {
  render(
    <Sprachraum sprache="de">
      <Datenpflege stand={stilleDatenpflege(teile)} />
    </Sprachraum>,
  )
}

describe('Datenpflege', () => {
  it('sieht auf Klick nach', async () => {
    const nachsehen = vi.fn()
    const nutzer = userEvent.setup()

    zeige({ nachsehen })

    await nutzer.click(screen.getByRole('button', { name: worte.nachsehen }))

    expect(nachsehen).toHaveBeenCalled()
  })

  it('sperrt den Knopf, solange etwas läuft', () => {
    zeige({ laeuft: true })

    expect(
      screen.getByRole('button', { name: worte.nachsehenLaeuft }),
    ).toBeDisabled()
  })

  it('erklärt einen sauberen Bestand, statt nichts zu zeigen', () => {
    zeige({ vorschau: vorschau({ befunde: [] }) })

    expect(screen.getByText(worte.sauber)).toBeInTheDocument()
  })

  // **Die Klasse entscheidet über die Vorauswahl.** `sicher` steht angehakt
  // da, `entscheidung` nicht — sonst nähme die Anwendung jemandem die
  // Entscheidung ab, während es aussieht, als hätte er sie getroffen.
  it('hakt nur die sicheren Befunde vor', () => {
    zeige({
      vorschau: vorschau({
        befunde: [
          befund(),
          befund({
            art: 'abgelaufeneToken',
            klasse: 'entscheidung',
            revision: 'def',
          }),
        ],
      }),
    })

    expect(
      screen.getByRole('checkbox', {
        name: satz(worte.arten, 'kartenwerteOhneKarte'),
      }),
    ).toBeChecked()

    expect(
      screen.getByRole('checkbox', {
        name: satz(worte.arten, 'abgelaufeneToken'),
      }),
    ).not.toBeChecked()
  })

  // **Was `nurInformieren` heißt, hat kein Kästchen** — es lässt sich gar
  // nicht wählen.
  it('gibt einem Befund, der nicht angefasst wird, kein Kästchen', () => {
    zeige({
      vorschau: vorschau({
        befunde: [
          befund({ art: 'offeneRueckkehren', klasse: 'nurInformieren' }),
        ],
      }),
    })

    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
    expect(
      screen.getByText(satz(worte.arten, 'offeneRueckkehren')),
    ).toBeInTheDocument()

    // **Und dann gibt es auch keinen Knopf** — er könnte nur nichts tun.
    expect(
      screen.queryByRole('button', { name: worte.aufraeumen }),
    ).not.toBeInTheDocument()
  })

  it('schickt die gewählten Maßnahmen samt ihrer Revision', async () => {
    const pflegen = vi.fn()
    const nutzer = userEvent.setup()

    zeige({
      pflegen,
      vorschau: vorschau({
        befunde: [
          befund(),
          befund({
            art: 'abgelaufeneToken',
            klasse: 'entscheidung',
            revision: 'def',
          }),
        ],
      }),
    })

    // **Die zweite dazuwählen** — sie stand nicht angehakt da.
    await nutzer.click(
      screen.getByRole('checkbox', {
        name: satz(worte.arten, 'abgelaufeneToken'),
      }),
    )

    await nutzer.click(screen.getByRole('button', { name: worte.aufraeumen }))

    expect(pflegen).toHaveBeenCalledWith('2026-09-10T12:00:00Z', [
      { art: 'kartenwerteOhneKarte', revision: 'abc' },
      { art: 'abgelaufeneToken', revision: 'def' },
    ])
  })

  // **Ohne Wahl kein Knopf, der etwas tut** — abgewählt heißt abgewählt.
  it('sperrt den Knopf, wenn nichts gewählt ist', async () => {
    const nutzer = userEvent.setup()

    zeige({ vorschau: vorschau() })

    await nutzer.click(
      screen.getByRole('checkbox', {
        name: satz(worte.arten, 'kartenwerteOhneKarte'),
      }),
    )

    expect(
      screen.getByRole('button', { name: worte.aufraeumen }),
    ).toBeDisabled()
  })

  it('nennt die Bilanz samt beider Stände', () => {
    zeige({
      bilanz: {
        gesichert: '/wo/es/liegt.db',
        zeilen: [
          {
            art: 'kartenwerteOhneKarte',
            ausgang: 'erledigt',
            anzahl: 3,
            grund: '',
          },
        ],
        vorher: [befund()],
        nachher: [],
        fehler: null,
      },
    })

    expect(screen.getByText(/\/wo\/es\/liegt\.db/)).toBeInTheDocument()
    expect(screen.getByText(/Weggeräumt \(3\)/)).toBeInTheDocument()
    expect(screen.getByText('Vorher 1 Befunde, nachher 0.')).toBeInTheDocument()
  })

  // **Ohne Sicherung steht keine Zeile darüber** — „gesichert nach nichts"
  // wäre schlimmer als keine Angabe.
  it('nennt keinen Ort, wenn nicht gesichert wurde', () => {
    zeige({
      bilanz: {
        gesichert: '',
        zeilen: [],
        vorher: [],
        nachher: [],
        fehler: null,
      },
    })

    expect(screen.queryByText(/Gesichert nach/)).not.toBeInTheDocument()
  })

  // **Auch in der Bilanz steht eine unbekannte Marke als Marke da** — und
  // zwar beide: die Art und der Ausgang.
  it('zeigt unbekannte Marken auch in der Bilanz', () => {
    zeige({
      bilanz: {
        gesichert: '',
        zeilen: [
          { art: 'schallplatten', ausgang: 'zerbrochen', anzahl: 1, grund: '' },
        ],
        vorher: [],
        nachher: [],
        fehler: null,
      },
    })

    expect(
      screen.getByText('schallplatten: zerbrochen (1)'),
    ).toBeInTheDocument()
  })

  // **Und ein `nurInformieren`-Befund ohne Katalogeintrag ebenso** — der Weg
  // ohne Kästchen hat seinen eigenen Rückfall.
  it('zeigt eine unbekannte Marke ohne Kästchen als Marke', () => {
    zeige({
      vorschau: vorschau({
        befunde: [befund({ art: 'schallplatten', klasse: 'nurInformieren' })],
      }),
    })

    expect(screen.getByText('schallplatten')).toBeInTheDocument()
  })

  it('zeigt einen Fehler als Meldung', () => {
    zeige({ fehler: 'die Datenbank ist zu' })

    expect(screen.getByRole('alert')).toHaveTextContent('die Datenbank ist zu')
  })

  // **Eine unbekannte Marke steht als Marke da** — sie ist immer noch eine
  // Auskunft.
  it('zeigt eine unbekannte Marke, statt sie zu verschlucken', () => {
    zeige({
      vorschau: vorschau({
        befunde: [befund({ art: 'schallplatten', klasse: 'seltsam' })],
      }),
    })

    expect(screen.getByText('schallplatten')).toBeInTheDocument()
    expect(screen.getByText(/seltsam/)).toBeInTheDocument()
  })
})
