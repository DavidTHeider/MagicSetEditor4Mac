// **Die Notizen einer Karte** — Werkstück 3 von Phase 3.5
// (#1166).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { angeboteneFelder } from './angebotenefelder'
import { Kartennotizen } from './Kartennotizen'
import { lageAus, vergessen } from '../speicherstand'

const texte = kataloge.de.seiten.karten.notizen

// gemeldet sammelt, was der Kasten nach oben gibt.
//
// **Der Fehler steht seit #1215 nicht mehr im Kasten**, sondern bei der Spalte
// darüber: der Kasten wird beim Kartenwechsel abgebaut, sie nicht.
let gemeldet: (string | undefined)[] = []

function zeige(werte: { feld: string; wert: string }[] = []) {
  gemeldet = []

  render(
    <Sprachraum sprache="de">
      <Kartennotizen
        setID="1"
        karteID="7"
        werte={werte}
        geaendert={vi.fn()}
        aufFehler={(meldung) => {
          gemeldet.push(meldung)
        }}
      />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
  vergessen()
})

describe('die Kartennotizen', () => {
  // **Sie stehen an jeder Karte, auch an einer ohne.**
  //
  //	notes = new TextCtrl(nodes_panel, ID_NOTES, true);
  //
  // `src/gui/set/cards_panel.cpp`, Zeile 39, Fassung `a7336d6` vom
  // 2020-06-14 — ein Kasten, der immer dasteht, und `true` heißt
  // **mehrzeilig** (`TextCtrl::TextCtrl(…, bool multi_line, …)`).
  it('steht auch an einer Karte ohne Notizen', () => {
    zeige()

    const kasten = screen.getByLabelText(texte.ueberschrift)

    expect(kasten.tagName).toBe('TEXTAREA')
    expect(kasten).toHaveValue('')
  })

  // **Und zeigen, was die Karte trägt.**
  it('zeigt die vorhandenen Notizen', () => {
    zeige([{ feld: 'notes', wert: 'Sets: [pmoa];\nFormats: ;' }])

    expect(screen.getByLabelText(texte.ueberschrift)).toHaveValue(
      'Sets: [pmoa];\nFormats: ;',
    )
  })

  // **Geschrieben wird beim Verlassen und nicht bei jedem Tastendruck.**
  //
  // Eine Notiz ist Fließtext; ein Schreibvorgang je Zeichen wären bei drei
  // Zeilen hundert Aufrufe, und jeder ginge gegen einen Stand, den der
  // Bestand gerade erst bekommen hat (§7, Regel 4).
  it('schreibt beim Verlassen', async () => {
    const gesetzt = vi.fn().mockResolvedValue({ fehler: null })
    belegeApi({ FeldSetzen: gesetzt })

    zeige()

    const kasten = screen.getByLabelText(texte.ueberschrift)

    await userEvent.type(kasten, 'meine Notiz')

    // **Noch nichts** — und der Beleg dafür steht darunter: derselbe Kasten
    // schreibt, sobald er verlassen wird (§3).
    expect(gesetzt).not.toHaveBeenCalled()

    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledWith(
        '1',
        '7',
        'notes',
        'meine Notiz',
        expect.any(Array),
      )
    })
  })

  // **Der Speicherstand hört mit** — angestrichen vom Zweitblick über Phase
  // 6.3 (M8): die Notizen schrieben am Stand vorbei, und über der Karte
  // stand „Jede Änderung steht im Bestand.", während die Notiz fehlte.
  it('meldet dem Speicherstand, wie das Schreiben ausging', async () => {
    const gesetzt = vi
      .fn()
      .mockResolvedValueOnce({
        fehler: {
          code: 'intern',
          meldung: 'kaputt',
          feld: '',
          schluessel: 'intern',
        },
      })
      .mockResolvedValueOnce({ fehler: null })
    belegeApi({ FeldSetzen: gesetzt })

    zeige()
    const kasten = screen.getByLabelText(texte.ueberschrift)

    await userEvent.type(kasten, 'erste')
    await userEvent.tab()
    await waitFor(() => {
      expect(lageAus('1')).toEqual({ art: 'fehler', feld: 'notes' })
    })

    await userEvent.type(kasten, ' und zweite')
    await userEvent.tab()
    await waitFor(() => {
      expect(lageAus('1')).toEqual({ art: 'geschrieben' })
    })
  })

  // **Ein unveränderter Kasten schreibt gar nicht.**
  //
  // Wer hineinklickt und wieder hinaus, hat nichts geändert. Ein Aufruf
  // trüge einen Verlaufsschritt ein, den niemand veranlasst hat — und ⌘Z
  // nähme danach etwas zurück, das nie geschehen ist.
  it('schreibt nicht, wenn sich nichts geändert hat', async () => {
    const gesetzt = vi.fn().mockResolvedValue({ fehler: null })
    belegeApi({ FeldSetzen: gesetzt })

    zeige([{ feld: 'notes', wert: 'unverändert' }])

    await userEvent.click(screen.getByLabelText(texte.ueberschrift))
    await userEvent.tab()

    expect(gesetzt).not.toHaveBeenCalled()

    // **Der Beleg, dass der Weg überhaupt schreibt.**
    await userEvent.type(screen.getByLabelText(texte.ueberschrift), '!')
    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledOnce()
    })
  })

  // **Eine zweite Änderung während des ersten Schreibens geht nicht verloren.**
  //
  // # Der Befund
  //
  // `useVorgang.starte` gibt für einen laufenden Schlüssel still `undefined`
  // zurück (§7, Regel 4) — richtig für einen Knopf, falsch für einen Kasten,
  // in dem jemand weitertippt. Wer den Kasten ein zweites Mal verließ, während
  // der erste Ruf unterwegs war, verlor seine Änderung, und niemand sagte es.
  // Angestrichen vom Zweitblick über Phase 3.5 als Befund 9.
  //
  // **Der erste Ruf bleibt hier absichtlich offen**, denn genau das war der
  // Fall, den kein Lauf hielt.
  it('holt eine Änderung nach, die während des Schreibens entstand', async () => {
    const antworten: ((wert: { fehler: null }) => void)[] = []

    const gesetzt = vi.fn(
      () =>
        new Promise<{ fehler: null }>((aufloesen) => {
          antworten.push(aufloesen)
        }),
    )

    belegeApi({ FeldSetzen: gesetzt })

    zeige()

    const kasten = screen.getByLabelText(texte.ueberschrift)

    await userEvent.type(kasten, 'erste')
    await userEvent.tab()

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledTimes(1)
    })

    // **Der erste Ruf hängt noch.** Jetzt kommt die zweite Änderung.
    await userEvent.type(kasten, ' zweite')
    await userEvent.tab()

    expect(gesetzt).toHaveBeenCalledTimes(1)

    // **Und jetzt antwortet der erste.**
    antworten[0]?.({ fehler: null })

    await waitFor(() => {
      expect(gesetzt).toHaveBeenCalledTimes(2)
    })

    expect(gesetzt).toHaveBeenLastCalledWith(
      '1',
      '7',
      'notes',
      'erste zweite',
      expect.any(Array),
    )
  })

  // **Ein Fehler erscheint am Kasten.**
  it('meldet einen Fehler', async () => {
    belegeApi({
      FeldSetzen: vi.fn().mockResolvedValue({
        fehler: { code: 'validierung', schluessel: '', meldung: 'geht nicht' },
      }),
    })

    zeige()

    await userEvent.type(screen.getByLabelText(texte.ueberschrift), 'x')
    await userEvent.tab()

    await waitFor(() => {
      expect(gemeldet).toContain('geht nicht')
    })
  })
})

// **Und im Formular steht `notes` nicht noch einmal.**
//
// Es ist **kein Kartenfeld des Spiels** — MSE2 gibt ihm einen eigenen Kasten
// unter der Kartenliste (`cards_panel.cpp`, Zeilen 39–50) und nicht einen
// Platz zwischen den Feldwerten. Bei einer Karte, die eines trägt, bot der
// Editor bis zum 2026-08-27 beides an: ein Kästchen mit der Aufschrift
// »notes«, einzeilig, am Ende der Liste.
describe('das Formular', () => {
  it('bietet notes nicht als Feld an', () => {
    const felder = angeboteneFelder(
      [
        { feld: 'name', wert: 'Feuerball' },
        { feld: 'notes', wert: 'Sets: [pmoa];' },
      ],
      [],
      [],
    )

    expect(felder.map((eines) => eines.feld)).toEqual(['name'])
  })
})
