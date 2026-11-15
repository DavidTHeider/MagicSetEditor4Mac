import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Feldeingabe } from './Feldeingabe'

// **Ersetzt wird beim Tippen — und nur dann.**
//
// So steht es in MSE2: `TextValueEditor::onChar` gibt als einziger Aufruf
// `allow_auto_replace = true` mit (`src/gui/value/text.cpp`, Zeilen 497 und
// 499, Fassung f9d9356 vom 2021-07-06). Einfügen, Löschen, Rückschritt und
// Zeilenumbruch ersetzen nicht
// (#897).

function stelle(feld: HTMLElement): number {
  return (feld as HTMLInputElement).selectionStart ?? -1
}

describe('die Autoersetzung im Feld', () => {
  function aufbau(
    ersetzen: (
      text: string,
      stelle: number,
    ) => Promise<{ text: string; stelle: number }>,
  ) {
    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
        autoErsetzen={ersetzen}
      />,
    )

    return { aendern, feld: screen.getByLabelText('name') }
  }

  it('ersetzt beim Tippen und setzt die Schreibmarke', async () => {
    const ersetzen = vi.fn(() => Promise.resolve({ text: '©', stelle: 1 }))
    const { aendern, feld } = aufbau(ersetzen)

    fireEvent.input(feld, {
      target: { value: '(C)' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalled()
    })

    expect((feld as HTMLInputElement).value).toBe('©')
    expect(stelle(feld)).toBe(1)
    expect(aendern).toHaveBeenLastCalledWith('©')
  })

  it('ersetzt nicht beim Einfügen', async () => {
    const ersetzen = vi.fn(() => Promise.resolve({ text: '©', stelle: 1 }))
    const { aendern, feld } = aufbau(ersetzen)

    fireEvent.input(feld, {
      target: { value: '(C)' },
      inputType: 'insertFromPaste',
    })

    // **Ein Beleg, dass jemand zuhört**, bevor die Verneinung kommt (§3): das
    // Tippen darunter feuert denselben Weg und ruft den Spion.
    //
    // **Mit einem anderen Wert**, denn React meldet eine Eingabe nur, wenn
    // sich der Wert geändert hat — derselbe Text zweimal käme nie an, und der
    // Beleg wäre keiner.
    fireEvent.input(feld, {
      target: { value: '(C)x' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledTimes(1)
    })

    expect(aendern).toHaveBeenCalledWith('(C)')
  })

  it('lässt den Text stehen, wenn nichts zu ersetzen war', async () => {
    const ersetzen = vi.fn(() => Promise.resolve({ text: 'abc', stelle: 3 }))
    const { feld } = aufbau(ersetzen)

    fireEvent.input(feld, {
      target: { value: 'abc' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalled()
    })

    expect((feld as HTMLInputElement).value).toBe('abc')
  })

  // **Ohne Schreibmarke wird ans Ende gesetzt.**
  //
  // `selectionStart` ist laut DOM-Vertrag `null` für Elemente, die keine
  // Auswahl führen. Für ein Textfeld kommt das nicht vor — der Rückfall steht
  // trotzdem da, weil der Typ ihn verlangt, und dieser Fall führt ihn aus,
  // statt ihn als unerreichbaren Zweig stehen zu lassen (§3).
  it('setzt ohne Schreibmarke ans Ende', async () => {
    const ersetzen = vi.fn(() => Promise.resolve({ text: '©', stelle: 1 }))
    const { feld } = aufbau(ersetzen)

    Object.defineProperty(feld, 'selectionStart', {
      value: null,
      configurable: true,
    })

    fireEvent.input(feld, { target: { value: '(C)' }, inputType: 'insertText' })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledWith('(C)', 3)
    })
  })

  it('kommt ohne Ersetzer aus', () => {
    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
      />,
    )

    fireEvent.input(screen.getByLabelText('name'), { target: { value: 'x' } })

    expect(aendern).toHaveBeenCalledWith('x')
  })
})

// **Eine überholte Antwort wird verworfen** — §7 Regel 2: eine Antwort gehört
// zu dem Stand, aus dem sie stammt.
//
// Die Ersetzung ist ein **asynchroner** Aufruf während des Tippens. Wer schnell
// tippt, hat zwei davon unterwegs; kommt die erste nach der zweiten zurück und
// schriebe ihr Ergebnis ins Feld, verlöre der Tippende sein zuletzt getipptes
// Zeichen.
describe('eine überholte Antwort', () => {
  it('überschreibt den neueren Text nicht', async () => {
    let ersteAufloesen: (wert: { text: string; stelle: number }) => void = () =>
      undefined

    const ersetzen = vi
      .fn<
        (
          text: string,
          stelle: number,
        ) => Promise<{ text: string; stelle: number }>
      >()
      .mockImplementationOnce(
        () =>
          new Promise((aufloesen) => {
            ersteAufloesen = aufloesen
          }),
      )
      .mockImplementation((text, stelle) => Promise.resolve({ text, stelle }))

    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
        autoErsetzen={ersetzen}
      />,
    )

    const feld = screen.getByLabelText('name')

    // Der erste Tastendruck — seine Antwort bleibt unterwegs.
    fireEvent.input(feld, { target: { value: '(C' }, inputType: 'insertText' })

    // Der zweite überholt ihn und kommt sofort zurück.
    fireEvent.input(feld, {
      target: { value: '(C)x' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledTimes(2)
    })

    // Und jetzt erst antwortet der erste — mit einem Stand von vorhin.
    ersteAufloesen({ text: '©', stelle: 1 })

    await vi.waitFor(() => {
      expect(aendern).toHaveBeenLastCalledWith('(C)x')
    })

    expect((feld as HTMLInputElement).value).toBe('(C)x')
  })
})

// **Eine abgelehnte Zusage hält das Tippen nicht auf** —
// #918.
//
// Der Weg hatte nur `.then`. Wird die Wails-Zusage **abgelehnt** — ein
// geschlossenes Fenster, ein Fehler in der Naht —, lief der Nachlauf gar nicht:
// das getippte Zeichen erreichte die Seite nicht, und die Vorschau blieb hinter
// dem sichtbaren Feld zurück, bis jemand das nächste tippte.
describe('eine abgelehnte Zusage', () => {
  it('lässt den getippten Text trotzdem ankommen', async () => {
    const ersetzen = vi.fn(() => Promise.reject(new Error('die Naht ist zu')))
    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
        autoErsetzen={ersetzen}
      />,
    )

    fireEvent.input(screen.getByLabelText('name'), {
      target: { value: 'abc' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(aendern).toHaveBeenCalledWith('abc')
    })
  })

  // **Auch die Ablehnung gehört zu ihrem Stand** (§7 Regel 2).
  //
  // Der Zweitblick über Phase R hat es eingefordert: geprüft wird vor Erfolg
  // **und** Fehler. Kommt die Ablehnung, nachdem jemand weitergetippt hat,
  // meldete ein `nachher(vorher)` den **alten** Text an die Seite zurück.
  it('meldet nach einem weiteren Tastendruck nichts mehr', async () => {
    let ersteAblehnen: (grund: Error) => void = () => undefined

    const ersetzen = vi
      .fn<
        (
          text: string,
          stelle: number,
        ) => Promise<{ text: string; stelle: number }>
      >()
      .mockImplementationOnce(
        () =>
          new Promise((_, ablehnen) => {
            ersteAblehnen = ablehnen
          }),
      )
      .mockImplementation((text, stelle) => Promise.resolve({ text, stelle }))

    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
        autoErsetzen={ersetzen}
      />,
    )

    const feld = screen.getByLabelText('name')

    fireEvent.input(feld, { target: { value: 'ab' }, inputType: 'insertText' })
    fireEvent.input(feld, { target: { value: 'abc' }, inputType: 'insertText' })

    await vi.waitFor(() => {
      expect(aendern).toHaveBeenCalledWith('abc')
    })

    aendern.mockClear()
    ersteAblehnen(new Error('zu spät'))

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledTimes(2)
    })

    expect(aendern).not.toHaveBeenCalled()
    expect((feld as HTMLInputElement).value).toBe('abc')
  })
})

// **Der Stand ist Text **und** Schreibmarke** — §7 Regel 2, ganz.
//
// #908 hat den
// Text verglichen und die Marke stehen lassen. Der Aufruf schickt aber beides,
// und die Antwort schreibt beides zurück: bewegt jemand die Marke, während die
// Ersetzung läuft, sprang sie danach an die alte Stelle
// (#933).
describe('eine nur bewegte Schreibmarke', () => {
  it('macht die Antwort ebenso ungültig wie neuer Text', async () => {
    let aufloesen: (wert: { text: string; stelle: number }) => void = () =>
      undefined

    const ersetzen = vi.fn(
      () =>
        new Promise<{ text: string; stelle: number }>((loesen) => {
          aufloesen = loesen
        }),
    )

    const aendern = vi.fn()

    render(
      <Feldeingabe
        feld="name"
        wert=""
        form={undefined}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon={undefined}
        aendern={aendern}
        verlassen={vi.fn()}
        autoErsetzen={ersetzen}
      />,
    )

    const feld: HTMLInputElement = screen.getByLabelText('name')

    fireEvent.input(feld, { target: { value: '(C)' }, inputType: 'insertText' })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledWith('(C)', 3)
    })

    // **Nur die Marke bewegt sich** — der Text bleibt derselbe.
    feld.setSelectionRange(0, 0)

    aufloesen({ text: '©', stelle: 1 })

    await vi.waitFor(() => {
      expect(ersetzen).toHaveBeenCalledTimes(1)
    })

    expect(feld.value).toBe('(C)')
    expect(feld.selectionStart).toBe(0)
  })
})
