// **Ein Textfeld mit Zeilennummern, Farbe und Fehlermarke** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Warum die Farbe über der Eingabe liegt und nicht darin
//
// Ein `<textarea>` kann keine Farben. Der übliche Griff ist ein `<pre>`
// darunter, das denselben Text in denselben Maßen zeigt, und darüber ein
// durchsichtiges Feld, das die Eingaben nimmt. Beide tragen dieselbe Schrift
// und dieselben Abstände; läuft das auseinander, steht die Farbe neben dem
// Zeichen.
//
// **Deshalb prüft diese Datei die Maße mit.** Eine Zusicherung „es gibt einen
// Abschnitt mit der Farbe X" sagt nichts darüber, ob er an der richtigen
// Stelle liegt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Quelltextfeld } from './Quelltextfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

// merkmale belegt die Farbbindung mit einer festen Zerlegung.
function merkmale(
  ...stellen: { art: string; zeile: number; spalte: number; breite: number }[]
) {
  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: stellen }) })
}

describe('Das Quelltextfeld', () => {
  afterEach(raeumeApi)

  it('nummeriert jede Zeile', async () => {
    merkmale()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert={'eins\nzwei\ndrei'}
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Drei Zeilen, drei Nummern** — und die dritte heißt 3, nicht 2.
    for (const nummer of ['1', '2', '3']) {
      expect(await screen.findByText(nummer)).toBeInTheDocument()
    }

    expect(screen.queryByText('4')).not.toBeInTheDocument()
  })

  // **Die letzte Zeile zählt mit, auch wenn sie leer ist.**
  //
  // `'a\n'` sind zwei Zeilen: eine mit `a` und eine leere darunter, in der der
  // Schreibzeiger steht. Ein `split('\n').length - 1` zählte hier eine.
  it('zählt die leere letzte Zeile mit', async () => {
    merkmale()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert={'a\n'}
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(await screen.findByText('2')).toBeInTheDocument()
  })

  it('färbt ein Merkmal an seiner Stelle', async () => {
    merkmale({ art: 'Schlüsselwort', zeile: 1, spalte: 1, breite: 2 })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld bezeichnung="Ausdruck" wert="if x" setzeWert={vi.fn()} />
      </Sprachraum>,
    )

    const gefaerbt = await screen.findByText('if')

    expect(gefaerbt.className).toContain('marke')
  })

  // **Was zwischen den Merkmalen steht, bleibt stehen.**
  //
  // Ohne diese Zusicherung könnte die Anzeige die Zwischenräume verschlucken —
  // und dann läge jedes Zeichen hinter dem ersten Merkmal neben seiner Farbe.
  it('behält den Text zwischen den Merkmalen', async () => {
    merkmale(
      { art: 'Bezeichner', zeile: 1, spalte: 1, breite: 1 },
      { art: 'Zahl', zeile: 1, spalte: 5, breite: 1 },
    )

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert="a + 1"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    const anzeige = await screen.findByTestId('quelltextanzeige')

    expect(anzeige.textContent).toBe('a + 1')
  })

  // **Ein Merkmal außerhalb der Zeile wird übergangen.**
  //
  // # Wann das vorkommt
  //
  // Zwischen Anschlag und Antwort: die Farben gehören zum **vorigen** Text,
  // der neue ist schon kürzer. Ohne diesen Zweig griffe `slice` ins Leere und
  // die Zeile bekäme Abschnitte, die es nicht gibt.
  //
  // **Und bei einem Merkmal, das schon vergeben ist**: zwei Stellen auf
  // derselben Spalte gäbe es nur bei widersprüchlichen Farben, aber ein
  // ungeprüfter Rückwärtsschritt machte aus `slice(bei, anfang)` eine leere
  // Zeichenkette und verlöre den Text dazwischen.
  it('übergeht ein Merkmal außerhalb der Zeile', async () => {
    merkmale(
      { art: 'Bezeichner', zeile: 1, spalte: 1, breite: 1 },
      { art: 'Zahl', zeile: 1, spalte: 40, breite: 1 },
      { art: 'Zahl', zeile: 1, spalte: 1, breite: 1 },
    )

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert="a + 1"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    const anzeige = await screen.findByTestId('quelltextanzeige')

    expect(anzeige.textContent).toBe('a + 1')
  })

  // **Eine Art, die dieses Fenster nicht kennt, bleibt lesbar.**
  //
  // Sie wird nur nicht hervorgehoben. Das ist die richtige Seite, auf der man
  // sich irrt: eine unbekannte Art unsichtbar zu färben versteckte Quelltext.
  it('lässt eine unbekannte Art lesbar', async () => {
    merkmale({ art: 'Neuart', zeile: 1, spalte: 1, breite: 1 })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld bezeichnung="Ausdruck" wert="a" setzeWert={vi.fn()} />
      </Sprachraum>,
    )

    // **Gewartet wird auf den gefärbten Zug, nicht auf den Kasten.**
    //
    // Der `<pre>` steht ab der **ersten** Zeichnung da — mit dem rohen Text
    // und ohne einen einzigen `<span>`, weil `Skriptfarben` noch unterwegs
    // ist. `findByTestId` sagt zu, dass ein Knoten im DOM steht, nicht dass
    // eine Antwort da ist (§3), und danach fand `querySelector` unter Last
    // nichts. Beim vollen Lauf am 2026-08-28 ist genau das passiert.
    //
    // **`waitFor` ist hier richtig und nicht die andere Falle**: die
    // Zusicherung trifft vorher **nicht** zu — es gibt keinen `<span>` —, sie
    // kann also nicht schon vor der Antwort halten.
    //
    // # Was belegt ist und was nicht
    //
    // **Belegt ist die Ursache**, und zwar durch die Meldung des roten Laufs
    // selbst: `the given combination of arguments (undefined and string)` —
    // `gefaerbt?.className` war `undefined`, also gab `querySelector` nichts
    // zurück.
    //
    // **Nicht belegt ist die Häufigkeit.** Nachstellen ließ es sich nicht: mit
    // der alten Form zwölf einzelne Läufe, drei gleichzeitige volle Suiten und
    // sechs Läufe unter künstlicher Last — **alle grün**. Was hier steht, ist
    // deshalb eine Härtung gegen eine Gestalt, die §3 benennt, und keine
    // vermessene Fehlerrate (§10).
    await waitFor(() => {
      expect(
        screen.getByTestId('quelltextanzeige').querySelector('span')?.className,
      ).toContain('text-text')
    })
  })

  // **Die gemeldete Stelle wird angestrichen.**
  it('streicht die gemeldete Stelle an', async () => {
    merkmale()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert={'1 +\n  * 2'}
          setzeWert={vi.fn()}
          stelle={{ zeile: 2, spalte: 3 }}
        />
      </Sprachraum>,
    )

    const marke = await screen.findByTestId('fehlermarke')

    expect(marke.textContent).toBe('*')
  })

  // **Eine Stelle hinter dem Zeilenende bekommt trotzdem eine Marke.**
  //
  // # Wann das vorkommt
  //
  // Ständig: „hier fehlt ein Wert" zeigt auf die Stelle **hinter** dem letzten
  // Zeichen. `slice` gibt dort nichts, und ein Zeiger ohne Inhalt wäre null
  // Pixel breit — man sähe ihn nicht.
  it('streicht auch hinter dem Zeilenende an', async () => {
    merkmale()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert="1 +"
          setzeWert={vi.fn()}
          stelle={{ zeile: 1, spalte: 4 }}
        />
      </Sprachraum>,
    )

    const marke = await screen.findByTestId('fehlermarke')

    // **Ein geschütztes Leerzeichen (U+00A0) und nicht nichts** — ein Zeiger
    // ohne Inhalt wäre null Pixel breit.
    expect(marke.textContent.codePointAt(0)).toBe(0x00a0)
  })

  // **Ohne Stelle gibt es keine Marke.**
  //
  // Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn immer
  // etwas angestrichen wäre.
  it('streicht ohne Stelle nichts an', async () => {
    merkmale()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert="1 + 1"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    await screen.findByTestId('quelltextanzeige')

    expect(screen.queryByTestId('fehlermarke')).not.toBeInTheDocument()
  })

  // **Ohne Sprache wird nicht eingefärbt — und nicht einmal gefragt.**
  //
  // Eine Paketdatei ist kein MSE-Script (#1181). Ein Feld, das trotzdem fragte,
  // bekäme Merkmale einer Sprache, die dort nicht steht — und färbte danach.
  it('fragt ohne Sprache gar nicht erst nach Farben', async () => {
    const gefragt = vi.fn(() => Promise.resolve({ merkmale: [] }))

    belegeApi({ Skriptfarben: gefragt })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Inhalt"
          wert="game: magic"
          setzeWert={vi.fn()}
          sprache="keine"
        />
      </Sprachraum>,
    )

    const anzeige = await screen.findByTestId('quelltextanzeige')

    expect(anzeige.textContent).toBe('game: magic')
    expect(gefragt).not.toHaveBeenCalled()
  })

  it('reicht die Eingabe weiter', async () => {
    merkmale()

    const gesetzt = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld bezeichnung="Ausdruck" wert="" setzeWert={gesetzt} />
      </Sprachraum>,
    )

    await userEvent.type(screen.getByLabelText('Ausdruck'), 'x')

    expect(gesetzt).toHaveBeenCalledWith('x')
  })

  // **Die Farben werden neu geholt, wenn sich der Text ändert.**
  it('holt die Farben nach einer Änderung neu', async () => {
    const gefragt = vi.fn(() => Promise.resolve({ merkmale: [] }))

    belegeApi({ Skriptfarben: gefragt })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Quelltextfeld bezeichnung="Ausdruck" wert="a" setzeWert={vi.fn()} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalledWith('a')
    })

    rerender(
      <Sprachraum sprache="de">
        <Quelltextfeld bezeichnung="Ausdruck" wert="ab" setzeWert={vi.fn()} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalledWith('ab')
    })
  })

  // **Ein Fehler beim Einfärben lässt den Text stehen.**
  //
  // Die Farbe ist eine Zutat, kein Inhalt. Wer den Text verlöre, weil eine
  // Bindung nicht antwortet, hätte einen Editor, der beim ersten Aussetzer
  // leer ist.
  it('lässt den Text stehen, wenn das Einfärben scheitert', async () => {
    belegeApi({ Skriptfarben: () => Promise.reject(new Error('weg')) })

    render(
      <Sprachraum sprache="de">
        <Quelltextfeld
          bezeichnung="Ausdruck"
          wert="1 + 1"
          setzeWert={vi.fn()}
        />
      </Sprachraum>,
    )

    const anzeige = await screen.findByTestId('quelltextanzeige')

    await waitFor(() => {
      expect(anzeige.textContent).toBe('1 + 1')
    })
  })
})
