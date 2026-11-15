// **Wie tief ein Kartenfeld liegt** — UX-05.AK-01
// (#1677).
import { describe, expect, it } from 'vitest'

import {
  feldtiefen,
  stehtIn,
  tiefeVon,
  verdeckterBefund,
  wirksameTiefe,
} from './feldtiefe'

import type { FeldbeschreibungDTO } from '../api'

function beschreibung(teile: Partial<FeldbeschreibungDTO> = {}) {
  return {
    name: 'name',
    aufschrift: 'Name',
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
    ...teile,
  }
}

describe('tiefeVon', () => {
  it('stellt ein gewöhnliches Feld des Spielpakets auf normal', () => {
    expect(tiefeVon(beschreibung())).toBe('normal')
  })

  // **Ein Feld, das ein Skript füllt, gehört nicht zwischen Kartenname und
  // Regeltext** — `field.hpp` Zeilen 50–51, Fassung f9d9356 vom 2021-07-06:
  // `editable` und `save_value`. Es sagt, **warum** etwas auf der Karte
  // steht, und das ist eine zweite Frage.
  it('stellt ein nicht bearbeitbares Feld auf erweitert', () => {
    expect(tiefeVon(beschreibung({ editierbar: false }))).toBe('erweitert')
  })

  it('stellt ein nicht gespeichertes Feld auf erweitert', () => {
    expect(tiefeVon(beschreibung({ gespeichert: false }))).toBe('erweitert')
  })

  // **Was das Spielpaket nicht erklärt, ist keine Karteneigenschaft.**
  // `stylesheet`, `styling data` und ihresgleichen sind Mitglieder von
  // `Card` (`src/data/card.hpp`, Zeilen 44–50) und keine `card field:`.
  it('stellt ein Feld ohne Beschreibung auf experte', () => {
    expect(tiefeVon(undefined)).toBe('experte')
  })
})

describe('stehtIn', () => {
  // **Die Tiefen liegen ineinander.** Was in `normal` steht, steht auch in
  // `erweitert` und in `experte` — sonst wäre ein Moduswechsel ein Verlust,
  // und AK-02 verlangt das Gegenteil.
  it('zeigt jede flachere Tiefe mit', () => {
    for (const gewaehlt of feldtiefen) {
      for (const feld of feldtiefen) {
        const flacher = feldtiefen.indexOf(feld) <= feldtiefen.indexOf(gewaehlt)

        expect(stehtIn(gewaehlt, feld)).toBe(flacher)
      }
    }
  })

  it('zeigt in normal nur normal', () => {
    expect(stehtIn('normal', 'normal')).toBe(true)
    expect(stehtIn('normal', 'erweitert')).toBe(false)
    expect(stehtIn('normal', 'experte')).toBe(false)
  })

  it('zeigt in experte alles', () => {
    expect(stehtIn('experte', 'normal')).toBe(true)
    expect(stehtIn('experte', 'erweitert')).toBe(true)
    expect(stehtIn('experte', 'experte')).toBe(true)
  })
})

describe('verdeckterBefund', () => {
  const nachName = new Map([['name', beschreibung()]])

  // **Ein Befund in einem verdeckten Feld ist der Fall, für den AK-02 die
  // „direkte Öffnungsaktion" verlangt.** Ohne ihn stünde eine Karte da, die
  // sich nicht speichern lässt, und niemand sähe, woran es liegt.
  it('nennt die Tiefe, in der ein verdeckter Befund liegt', () => {
    expect(verdeckterBefund('normal', [{ feld: 'stylesheet' }], nachName)).toBe(
      'experte',
    )
  })

  it('nennt nichts, wo der Befund schon zu sehen ist', () => {
    expect(
      verdeckterBefund('normal', [{ feld: 'name' }], nachName),
    ).toBeUndefined()
  })

  it('nennt nichts ohne Befunde', () => {
    expect(verdeckterBefund('normal', [], nachName)).toBeUndefined()
  })

  // **Ein Befund ohne Feld gehört nirgendwohin.** Ein Rasterbefund nennt
  // nicht immer eines, und ein Sprung ins Blaue wäre schlechter als keiner.
  it('nennt nichts für einen Befund ohne Feld', () => {
    expect(verdeckterBefund('normal', [{}], nachName)).toBeUndefined()
  })

  // **Die flachste reicht.** Liegen zwei verdeckt, führt der Sprung dorthin,
  // wo weniger dazukommt — wer tiefer will, geht weiter.
  it('nennt die flachste von zweien', () => {
    expect(
      verdeckterBefund(
        'normal',
        [{ feld: 'stylesheet' }, { feld: 'skriptfeld' }],
        new Map([
          ['name', beschreibung()],
          ['skriptfeld', beschreibung({ editierbar: false })],
        ]),
      ),
    ).toBe('erweitert')
  })
})

describe('wirksameTiefe', () => {
  const nachName = new Map([['name', beschreibung()]])

  it('lässt die gewählte stehen, solange etwas in normal liegt', () => {
    expect(wirksameTiefe('normal', [{ feld: 'name' }], nachName)).toBe('normal')
  })

  // **Ein Set ohne sein Spielpaket** — dann erklärt niemand ein Feld, und
  // `normal` zeigte ein leeres Formular. Eine Karte, die sich nicht
  // ausfüllen lässt, ist schlechter als eine mit zu vielen Feldern.
  it('geht auf experte, wo nichts in normal liegt', () => {
    expect(wirksameTiefe('normal', [{ feld: 'stylesheet' }], new Map())).toBe(
      'experte',
    )
  })

  it('geht auch ohne Felder auf experte', () => {
    expect(wirksameTiefe('erweitert', [], new Map())).toBe('experte')
  })
})

// **Und die Gegenrichtung**: liegt der zweite Befund tiefer, bleibt der
// flachste stehen. Ohne diesen Lauf bliebe der Zweig unbetreten, und die
// Reihenfolge der Befunde entschiede, wohin der Sprung führt.
it('behält die flachste, wenn der zweite tiefer liegt', () => {
  expect(
    verdeckterBefund(
      'normal',
      [{ feld: 'skriptfeld' }, { feld: 'stylesheet' }],
      new Map([['skriptfeld', beschreibung({ editierbar: false })]]),
    ),
  ).toBe('erweitert')
})
