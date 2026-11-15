// **Die Bedienfläche der Massenänderung** — Werkstück 5 von Phase 3.5
// (#1166).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `massenaenderung.test.tsx` prüft, was die Änderung **tut**; hier steht, ob
// ein Mensch sie überhaupt auslösen kann, ohne die Vorlage auswendig zu
// kennen.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Massenaenderung } from './Massenaenderung'

const texte = kataloge.de.seiten.karten.massen

const karten: KartenzeileDTO[] = [
  { id: '1', nummer: 0, name: 'Feuerball', felder: 1 },
]

const feldwerte = {
  Kartenfeldwerte: () => Promise.resolve({ werte: [] }),
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

// **Das Feld wird gewählt, nicht getippt** — Werkstück 5 von Phase 3.5
// (#1166).
//
// # Warum das eine Zusage über die Oberfläche ist und nicht über eine Bindung
//
// Die Bindung `FelderSetzen` gab es die ganze Zeit, und `FAEHIGKEITEN.md` hat
// die Fähigkeit als „am Fenster vorhanden" gezählt. **Vorhanden war ein
// Textfeld**, in das man den rohen Feldbezeichner tippen musste — `super
// type`, `casting cost`, `card color`. Wer die Vorlage nicht kennt, kennt
// diese Namen nicht, und ein Tippfehler schrieb still nichts.
//
// Angestrichen als Z3-01 im Audit: **Bindungsreichweite ist nicht
// Oberflächenreife.**
// warteAufDieFelder gibt die Feldauswahl, **nachdem** die Vorlage geantwortet
// hat.
//
// `findByLabelText(texte.feld)` findet sie sofort — mit ihrem Platzhalter und
// sonst nichts. Bis zum 2026-09-02 war die Antwort trotzdem da, weil
// `test/api.ts` synchron antwortete; seither kommt sie eine Makroaufgabe
// später, so wie über die IPC (§3).
async function warteAufDieFelder(): Promise<HTMLElement> {
  const feld = await screen.findByLabelText(texte.feld)

  await waitFor(() => {
    expect(within(feld).getAllByRole('option').length).toBeGreaterThan(1)
  })

  return feld
}

describe('das Feld der Massenänderung', () => {
  it('bietet die Felder der Vorlage an, statt sie tippen zu lassen', async () => {
    belegeApi({
      ...feldwerte,
      Kartenfelder: vi.fn().mockResolvedValue({
        felder: [
          {
            name: 'super type',
            aufschrift: 'Kartentyp',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
          {
            name: 'rarity',
            aufschrift: 'Seltenheit',
            art: 'choice',
            mehrzeilig: false,
            wahlen: ['common', 'rare'],
            beschreibung: '',
            vorgabe: 'common',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    const feld = await warteAufDieFelder()

    // **Eine Auswahl und kein Textfeld.** `getByRole('combobox')` trifft ein
    // `<select>`; ein `<input type="text">` hat die Rolle `textbox`.
    expect(feld.tagName).toBe('SELECT')

    // **Und die Aufschrift steht darin, nicht der Bezeichner** (§9): »Kartentyp«
    // ist, was in der Vorlage als `caption:` steht, `super type` der Name, an
    // dem geschrieben wird.
    expect(screen.getByRole('option', { name: 'Kartentyp' })).toHaveValue(
      'super type',
    )
  })

  // **Und ein Wahlfeld bekommt seine Wahlmöglichkeiten.**
  //
  // Der zweite Halbsatz des Befunds: „und einen untypisierten Wert". Wer
  // `rarity` auf `mythic rare` setzen will, soll es auswählen und nicht raten,
  // wie die Vorlage es schreibt.
  it('bietet für ein Wahlfeld seine Wahlmöglichkeiten an', async () => {
    const vorgeschaut = vi.fn().mockResolvedValue({
      feld: 'rarity',
      wert: 'rare',
      zeilen: [],
    })

    belegeApi({
      ...feldwerte,
      FelderVorschau: vorgeschaut,
      Kartenfelder: vi.fn().mockResolvedValue({
        felder: [
          {
            name: 'rarity',
            aufschrift: 'Seltenheit',
            art: 'choice',
            mehrzeilig: false,
            wahlen: ['common', 'rare'],
            beschreibung: '',
            vorgabe: 'common',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.selectOptions(await warteAufDieFelder(), 'rarity')

    const wert = screen.getByLabelText(texte.wert)

    expect(wert.tagName).toBe('SELECT')
    expect(screen.getByRole('option', { name: 'rare' })).toBeInTheDocument()

    // **Und die Wahl geht wirklich hinaus.**
    //
    // Ohne diese Hälfte prüfte der Lauf nur, dass die Möglichkeiten
    // **dastehen** — eine Auswahl, die nichts setzt, sähe genauso aus. §10:
    // ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar.
    await userEvent.selectOptions(wert, 'rare')
    await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))

    await waitFor(() => {
      expect(vorgeschaut).toHaveBeenCalledWith('1', ['1'], 'rarity', 'rare')
    })
  })
})

// **Ein Feld ohne Aufschrift zeigt seinen Bezeichner.**
//
// `caption:` ist in einer Vorlage nicht Pflicht, und ein leerer Eintrag in der
// Auswahl wäre schlimmer als ein technischer Name: er sagt gar nichts. Der
// Fall ist selten und die Zeile billig — ohne ihn stünde ein Ausweich da, das
// nie greift, und niemand wüsste, ob es stimmt (§3).
describe('ein Feld ohne Aufschrift', () => {
  it('zeigt seinen Bezeichner', async () => {
    belegeApi({
      ...feldwerte,
      Kartenfelder: vi.fn().mockResolvedValue({
        felder: [
          {
            name: 'card color',
            aufschrift: '',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    expect(
      await screen.findByRole('option', { name: 'card color' }),
    ).toHaveValue('card color')
  })
})

// **Was der Editor nicht anbietet, bietet die Massenänderung auch nicht an.**
//
// # Der Befund
//
// Die Auswahl aus #1204 nahm **jedes** Feld der Vorlage. Zwei Arten gehören
// nicht hinein, und der Editor lässt beide seit #708 und #696 weg:
//
//   - **`editable: false`** — MSE lässt ein solches Feld nicht einmal
//     auswählen (`is_enabled`, `src/gui/control/card_editor.cpp`, Zeilen
//     155–157). 24 der 249 Kartenfelder von `magic.mse-game` tragen es.
//   - **`save value: false`** — ein Feld, das aus anderen entsteht. Wer
//     hineinschreibt, schreibt einen Wert, den die Vorlage nicht liest; 46 der
//     249 tragen es. Das **erste** angebotene Feld, `shape`, trägt beides.
//
// Angestrichen vom Zweitblick über Phase 3.5 als Befund 1. Die Zahlen sind
// seine, an dem einen installierten `M15-Magic-Pack` gezählt.
describe('die Felder, die die Massenänderung anbietet', () => {
  it('lässt nicht editierbare und berechnete Felder weg', async () => {
    belegeApi({
      ...feldwerte,
      Kartenfelder: vi.fn().mockResolvedValue({
        felder: [
          {
            name: 'shape',
            aufschrift: 'Form',
            art: 'choice',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: false,
            gespeichert: false,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
          {
            name: 'text',
            aufschrift: 'Text',
            art: 'text',
            mehrzeilig: true,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: false,
            teilfelder: ['rule text', 'flavor text'],
            reihe: '',
            stelle: 0,
          },

          // **Und der Gegenfall, den der erste Anlauf durchließ**: ein
          // berechnetes Feld **ohne** ableitbare Teilfelder. `copyright line`
          // in `magic.mse-game` ist so eines (Zeilen 3234–3243: `save value:
          // false`, kein `editable: false`, nur ein Skript). Wer
          // `teilfelder.length === 0` als „gespeichert" liest, bietet es an —
          // und der Ruf dahinter endet zuverlässig im Fehler.
          //
          // Angestrichen vom Nachlauf über Phase 3.5 als Befund 1.
          {
            name: 'copyright line',
            aufschrift: 'Copyright',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: false,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
          {
            name: 'name',
            aufschrift: 'Kartenname',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    const feld = await warteAufDieFelder()

    const namen = [...feld.querySelectorAll('option')].map((eine) => eine.value)

    // **Nur das eine bleibt** — und die leere Wahl davor.
    expect(namen).toEqual(['', 'name'])
  })
})

// **Die Glieder einer Feldreihe stehen unter ihrem Grundfeld** — Werkstück 3
// von Phase 3.6
// (#1167).
//
// # Warum das an der Oberfläche geprüft wird und nicht nur an `gruppiereFelder`
//
// Weil dort die Zusage steht, die jemand liest: `Name`, `Name 2`, `Name 3` und
// `Name 4` als vier gleichrangige Einträge ließen niemanden sehen, dass sie
// dasselbe Feld an vier Stellen sind — und bei einer zweiseitigen Vorlage ist
// die zweite die Rückseite.
