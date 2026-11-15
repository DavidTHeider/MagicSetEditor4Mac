// **Ein Bild kommt auch durch einen Klick in die Fläche** —
// #246.
//
// # Wofür
//
// AGENTS.md §6: die Karte ist das Hauptobjekt. Wer eine Illustration einsetzen
// will, klickt dorthin, wo sie hingehört — das Kunstfenster —, nicht auf einen
// Knopf in einer Liste von 24.
//
// # Abweichung von MSE, benannt
//
// MSE öffnet den Dialog beim **Doppelklick**
// (`ImageValueEditor::onLeftDClick`, `src/gui/value/image.cpp`, Zeilen 22–35,
// Fassung f9d9356 vom 2021-07-06) und wählt beim einfachen. Hier tut der
// einfache Klick es: „wählen" heißt bei uns, den Fokus ins Eingabefeld zu
// setzen, und ein Bildfeld hat keines, in das sich tippen ließe. Ein Klick,
// der nichts tut, ist die schlechtere Antwort.
import { fireEvent, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { effekteNachholen } from '../test/effekte'
import { raeumeApi } from '../test/api'
import {
  type MitLaufzeit,
  zeigeBildflaeche as zeige,
} from '../test/bildflaeche'

afterEach(() => {
  raeumeApi()
  delete (globalThis as MitLaufzeit).runtime
})

const texte = kataloge.de

describe('Ein Klick in die Fläche', () => {
  it('öffnet bei einem Bildfeld den Dateidialog', async () => {
    const { BildSetzen } = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
      }),
    )

    await waitFor(() => {
      expect(BildSetzen).toHaveBeenCalledWith('5', 'kunst')
    })
  })

  // **Ein Bildfeld ist ein Ablageziel, ein Textfeld nicht** — AP-4 (#2361),
  // B-24: Wails markiert beim Ziehen das nächste `data-file-drop-target`,
  // und `Ablagehinweis` nennt das Feld.
  it('macht nur ein Bildfeld zum Ablageziel', async () => {
    zeige()

    const bildfeld = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })
    expect(bildfeld).toHaveAttribute('data-file-drop-target', '')
    expect(
      document.querySelectorAll('[data-file-drop-target]:not([data-bildfeld])'),
    ).toHaveLength(0)
  })

  // **Und bei jedem anderen Feld bleibt es beim Wählen.** Ein Textfeld hat
  // eine Eingabe, in die sich der Fokus setzen lässt; ein Dateidialog wäre
  // dort das Falsche.
  it('setzt bei einem Textfeld nur den Fokus', async () => {
    const { BildSetzen } = zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `name — ${texte.seiten.karten.feldWaehlen}`,
      }),
    )

    expect(document.activeElement).toBe(screen.getByLabelText('name'))
    expect(BildSetzen).not.toHaveBeenCalled()
  })
})

// **Und eine gezogene Datei landet im Feld darunter.**
//
// # Zwei Hälften
//
// Der Browser sieht bei einem Ablegen nur einen **Namen**, keinen Pfad — die
// Sicherheitszusage jeder Webansicht. Wails hebt sie für die eigene Anwendung
// auf und meldet die absoluten Pfade (`runtime.OnFileDrop`,
// `EnableFileDrop`). **Über welchem Feld** losgelassen wurde, weiß dagegen nur
// das Fenster: Go bekommt Koordinaten, und die in ein Feld umzurechnen hieße,
// den Aufbau der Seite ein zweites Mal zu kennen.
//
// Hier trifft sich beides: das Überziehen merkt das Feld, Wails liefert den
// Pfad.
describe('Eine gezogene Datei', () => {
  it('landet im Bildfeld, über dem sie losgelassen wurde', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeige()

    const flaeche = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    fireEvent.dragOver(flaeche)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).toHaveBeenCalledWith(
        '5',
        'kunst',
        '/Users/david/urlaub.jpg',
      )
    })
  })

  // **Wer das Feld wieder verlässt, meint es nicht mehr** — Befund 2 des
  // Zweitblicks über Phase Y.
  //
  // `onDragOver` merkt das Feld, `nimmFeld` vergisst es beim **Ablegen**. Ein
  // Zug, der über das Feld geht und weiterzieht, ließ es bis zum 2026-08-25
  // stehen: das nächste Ablegen **irgendwo** traf dann `kunst`.
  it('vergisst das Feld, wenn der Zug es wieder verlässt', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeige()

    const flaeche = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    // **Erst die Effekte, dann die Ereignisse** (§3).
    //
    // `findByRole` sagt zu, dass der Knopf im DOM steht — **nicht**, dass
    // `useAbgelegteDateien` seinen Horcher schon gesetzt hat. Ohne dieses
    // Nachholen kann `ablegen` ins Leere gehen, und die Verneinung darunter
    // wäre grün, ohne etwas geprüft zu haben. Angestrichen vom Nachlauf des
    // Zweitblicks über Phase Y.
    await effekteNachholen()

    fireEvent.dragOver(flaeche)
    fireEvent.dragLeave(flaeche)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    // **Die Verneinung steht vor einem Beleg, dass jemand zuhört** (§3) — der
    // zweite Zug unten geht denselben Weg und kommt an.
    await waitFor(() => {
      expect(BildAusDatei).not.toHaveBeenCalled()
    })

    fireEvent.dragOver(flaeche)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).toHaveBeenCalledWith(
        '5',
        'kunst',
        '/Users/david/urlaub.jpg',
      )
    })
  })

  // **Und ein Textfeld räumt das gemerkte Bildfeld nicht weg.**
  //
  // Der Zug geht über die ganze Karte; `name` liegt unter `kunst` und meldet
  // sein eigenes Weggehen. Ohne den Wächter auf `istBildfeld` löschte es die
  // Erinnerung an `kunst`, und das Ablegen träfe nichts mehr.
  it('lässt ein Textfeld das gemerkte Bildfeld in Ruhe', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeige()

    const bildfeld = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    await effekteNachholen()

    fireEvent.dragOver(bildfeld)

    // **Das Textfeld daneben** — es ist kein Bildfeld und trägt deshalb keine
    // Aufschrift „Bild einsetzen".
    const textfeld = await screen.findByRole('button', {
      name: `name — ${texte.seiten.karten.feldWaehlen}`,
    })

    fireEvent.dragLeave(textfeld)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).toHaveBeenCalledWith(
        '5',
        'kunst',
        '/Users/david/urlaub.jpg',
      )
    })
  })

  // **Ein Set ist kein Bild** (#891).
  //
  // Bis zum 2026-08-25 ging **jeder** Pfad ins Feld darunter: wer eine
  // `.mse-set` auf eine Karte zog, bekam sie als Bild eingesetzt. Die Endung
  // entscheidet jetzt zuerst — und zwar **ohne Rücksicht** darauf, über
  // welchem Feld losgelassen wurde.
  it('öffnet ein Setpaket, statt es einzusetzen', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei, aufSet } = zeige()

    const flaeche = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    fireEvent.dragOver(flaeche)
    ablegen?.(0, 0, ['/Users/david/Vanguard.MSE-SET'])

    await waitFor(() => {
      expect(aufSet).toHaveBeenCalledWith('/Users/david/Vanguard.MSE-SET')
    })

    // **Und es landet nicht im Feld.** Der Beleg, dass jemand zuhört, steht
    // im Lauf darüber: derselbe Weg setzt ein Bild, wenn es eines ist.
    expect(BildAusDatei).not.toHaveBeenCalled()
  })

  // **Über einem Textfeld nicht.** Ein Regeltext nimmt kein Bild entgegen, und
  // ein Ablegen dort soll auch nichts merken: sonst landete die nächste Datei
  // in einem Feld, über dem niemand mehr war.
  it('wird über einem Textfeld nicht gemerkt', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeige()

    fireEvent.dragOver(
      await screen.findByRole('button', {
        name: `name — ${texte.seiten.karten.feldWaehlen}`,
      }),
    )
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    expect(BildAusDatei).not.toHaveBeenCalled()
  })
})
