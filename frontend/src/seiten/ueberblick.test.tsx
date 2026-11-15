// **Der Überblick auf der Startseite** — Werkstück 3 von Phase 3.5
// (#1166).
//
// # Was der Befund war
//
// „Erste Schritte" stand **für immer** da. Bei einem leeren Bestand ist das
// genau richtig — §6 verlangt vom leeren Zustand, dass er den nächsten Schritt
// erklärt. Bei zwölf Sets und dreitausend Karten ist es eine Anleitung für
// jemanden, der längst weiter ist.
//
// # Was UXR-03 daran geändert hat
//
// #1911,
// gemessen am 2026-10-02 am nativen Fenster: **13 Überschriften** auf der
// Startseite, fünf davon für vier Zahlen — eine Überschrift über dem Raster,
// eine je Kachel, dazu ein Erklärsatz und ein Knopf „Zu …" in jeder. Das
// Artboard *Main* zeigt dieselben vier als **eine Zeile**: Zahl groß, Wort
// klein, keine Überschrift. **Der Weg bleibt**: jede Kennzahl ist selbst der
// Knopf zu der Seite, auf der ihre Zahl entsteht.
//
// **Und Decks statt Vorlagenpakete**, wie im Artboard: die Startseite fragt
// „woran arbeite ich weiter?", und ein Deck ist Arbeit, ein Paket Material.
// Die Paketzahl eines echten Bestands geht in die Hunderte und sagt darüber
// nichts.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api } from '../api'

import type { ExemplarDTO } from '../api'
import type { DeckDTO } from '../apidecks'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

afterEach(() => {
  raeumeApi()
})

function set(id: string, karten: number) {
  return {
    id,
    name: 'Vier Winde',
    spiel: 'magic',
    stil: 'magic-m15',
    karten,
    herkunft: '/a.mse-set',
    herkunftFehlt: false,
  }
}

// deck und exemplar bauen die Gestalten, die die Bindungen liefern.
function deck(id: string): DeckDTO {
  return {
    id,
    fassung: '1',
    name: `Deck ${id}`,
    spiel: 'magic',
    notiz: '',
    formatID: '',
    plaetze: [],
    hauptdeck: 60,
    sideboard: 15,
    herkunft: keineHerkunft,
  }
}

function exemplar(id: number, printing: string, menge: number): ExemplarDTO {
  return {
    id,
    fassung: 1,
    printing,
    druckID: '',
    vorlaeufig: false,
    menge,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: '',
    herkunft: keineHerkunft,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
    </Sprachraum>,
  )
}

describe('der Überblick', () => {
  // **Der Überblick steht auch bei leerem Bestand** — seit dem 2026-09-03
  // (#1325).
  //
  // Hier stand das Gegenteil: bei leerem Bestand zeigte die Seite die ersten
  // Schritte **statt** des Überblicks. Weil „noch nicht geladen" wie „kein
  // Set" aussah, sprang sie um, sobald die Antwort eintraf.
  //
  // **Vier Zahlen, von denen alle null sind, sagen zutreffend, dass hier noch
  // nichts liegt.** Was zu tun ist, sagt die Einführung — als Fenster.
  //
  // **Gegriffen über die Liste und ihren Namen** (UXR-03): „Dein Überblick"
  // ist keine Überschrift mehr, sondern der Name der Zeile.
  it('steht auch bei leerem Bestand', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [], fehler: null }),
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [],
          gesamt: 0,
          insgesamt: 0,
          stueck: 0,
          fehler: null,
        }),
    })

    zeige()

    // **Auf die Knöpfe warten, nicht auf die Liste** — seit AP-4 steht sie
    // schon beim Laden da, mit Platzhaltern; wer auf sie wartete, fragte die
    // Knöpfe ab, bevor die Antwort kam (rot in `make ci` am 2026-10-17).
    const zeile = screen.getByRole('list', { name: texte.ueberblick.titel })
    expect(await within(zeile).findAllByRole('button')).toHaveLength(4)
    expect(screen.queryByText(texte.ersteSchritte)).toBeNull()
  })

  // **Und schon, solange er lädt — in seiner Endhöhe** — AP-4
  // (#2361),
  // Befund B-12. Hier stand ein Kommentar „der Überblick steht immer“ über
  // einer Bedingung, die ihn bis zur Antwort wegließ: kam sie, rückte alles
  // darunter um eine Kachelzeile nach unten. Jetzt stehen vier Platzhalter da,
  // `aria-busy`, ohne Knöpfe, die noch nichts zu sagen haben.
  it('steht als Platzhalter da, solange die Zahlen ausstehen', async () => {
    let antworten: (sets: { sets: []; fehler: null }) => void = () => undefined
    belegeApi({
      Sets: () =>
        new Promise((los) => {
          antworten = los
        }),
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [],
          gesamt: 0,
          insgesamt: 0,
          stueck: 0,
          fehler: null,
        }),
    })

    zeige()

    const zeile = await screen.findByRole('list', {
      name: texte.ueberblick.titel,
    })
    expect(zeile).toHaveAttribute('aria-busy', 'true')
    expect(within(zeile).getAllByRole('listitem')).toHaveLength(4)
    expect(within(zeile).queryAllByRole('button')).toHaveLength(0)

    antworten({ sets: [], fehler: null })

    await waitFor(() => {
      expect(within(zeile).getAllByRole('button')).toHaveLength(4)
    })
    expect(zeile).not.toHaveAttribute('aria-busy')
  })

  // **Mit Sets steht der Überblick da — und die Zahlen sind echt.**
  //
  // **Das Wort zählt mit** (UXR-03): unter einer 1 steht „Set" und nicht
  // „Sets" — dieselbe Regel wie überall, wo eine Zahl ein Wort hat (#1218).
  it('zählt Sets, Karten, Exemplare und Decks', async () => {
    const fenster = vi.fn<Api['Sammlungsfenster']>(() =>
      Promise.resolve({
        exemplare: [exemplar(1, 'x', 3)],
        gesamt: 2,
        insgesamt: 2,
        stueck: 7,
        fehler: null,
      }),
    )

    belegeApi({
      Sets: () =>
        Promise.resolve({ sets: [set('1', 107), set('2', 250)], fehler: null }),
      Decks: () =>
        Promise.resolve({
          decks: [deck('1'), deck('2'), deck('3')],
          fehler: null,
        }),
      // **Die Summe kommt aus dem Fenster, nicht aus der Liste** (U15,
      // #2160): die ganze Sammlung wird dafür nicht mehr übertragen.
      Sammlung: () =>
        Promise.reject(new Error('die ganze Liste wird nicht gebraucht')),
      Sammlungsfenster: fenster,
    })

    zeige()

    const zeile = await screen.findByRole('list', {
      name: texte.ueberblick.titel,
    })

    // 2 Sets, 357 Karten, 7 Exemplare in 2 Zeilen, 3 Decks — **in der
    // Reihenfolge des Artboards**.
    await waitFor(() => {
      expect(
        within(zeile)
          .getAllByRole('button')
          .map((knopf) => knopf.textContent),
      ).toEqual(['2 Sets', '357 Karten', '7 Exemplare', '3 Decks'])
    })
    expect(screen.queryByText(texte.ersteSchritte)).not.toBeInTheDocument()
    expect(fenster.mock.calls[0]?.[4]).toBeLessThanOrEqual(1)
  })

  // **Keine Überschrift für vier Zahlen** (UXR-03, R4 aus
  // `docs/UIREVIEW.md`): vorher fünf — eine über dem Raster und eine je
  // Kachel.
  it('trägt keine eigene Überschrift', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [set('1', 1)], fehler: null }),
      Decks: () => Promise.resolve({ decks: [deck('1')], fehler: null }),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [exemplar(1, 'x', 1)],
          gesamt: 1,
          insgesamt: 1,
          stueck: 1,
          fehler: null,
        }),
    })

    zeige()

    const zeile = await screen.findByRole('list', {
      name: texte.ueberblick.titel,
    })
    await within(zeile).findByRole('button', { name: '1 Set' })

    expect(within(zeile).queryAllByRole('heading')).toHaveLength(0)
    expect(within(zeile).getByRole('button', { name: '1 Karte' })).toBeVisible()
    expect(
      within(zeile).getByRole('button', { name: '1 Exemplar' }),
    ).toBeVisible()
    expect(within(zeile).getByRole('button', { name: '1 Deck' })).toBeVisible()
  })

  // **Ein Lesefehler hält nichts auf** (§8).
  //
  // Die Zahlen sind eine Zugabe; wer die Anwendung öffnet, will damit
  // arbeiten. Eine Startseite, die wegen einer unlesbaren Deckliste gar
  // nichts zeigt, wäre der schlechtere Tausch.
  //
  // # Eine **abgelehnte** Zusage und kein fachlicher Fehlerwert
  //
  // Der erste Anlauf gab hier einen **erfüllten** Umschlag mit `fehler` zurück.
  // Der Produktionsweg packt ihn gar nicht aus — der Test war grün, und der
  // behauptete Weg blieb unbetreten. Genau so bricht ein Abbruch über die
  // Wails-Naht: der Promise **lehnt ab**. Angestrichen vom Zweitblick über
  // Phase 3.5 als Befund 6.
  it('bleibt stehen, wenn eine Zahl nicht zu holen ist', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [set('1', 107)], fehler: null }),
      Decks: () => Promise.reject(new Error('die Naht ist abgebrochen')),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [],
          gesamt: 0,
          insgesamt: 0,
          stueck: 0,
          fehler: null,
        }),
    })

    zeige()

    // **Und die Zahl, die zu holen war, steht trotzdem da** — die andere
    // zählt null.
    expect(
      await screen.findByRole('button', { name: '107 Karten' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '0 Decks' })).toBeInTheDocument()
  })

  // **Jede Kennzahl führt dorthin, wo ihre Zahl entsteht.**
  //
  // Eine Zahl ohne Weg ist eine Auskunft und kein Überblick — wer sieht, dass
  // er 357 Karten hat, will zu ihnen. **Ohne diesen Lauf wäre auch eine
  // Kachel grün, deren Knopf nichts tut.**
  it('führt von jeder Kennzahl zu ihrer Seite', async () => {
    const gewechselt = vi.fn()

    belegeApi({
      Sets: () => Promise.resolve({ sets: [set('1', 107)], fehler: null }),
      Decks: () => Promise.resolve({ decks: [deck('1')], fehler: null }),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [exemplar(1, 'x', 3)],
          gesamt: 1,
          insgesamt: 1,
          stueck: 3,
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Start wechseln={gewechselt} waehleSet={vi.fn()} bestandsstand={0} />
      </Sprachraum>,
    )

    // **Erst steht der Überblick, dann wird geklickt.**
    //
    // `findByRole` sagt zu, dass ein Knoten dasteht — nicht, dass React mit
    // seinen Effekten fertig ist (§3, `effekteNachholen`). Wer den Knopf davor
    // greift, klickt unter Umständen auf einen Knoten, den das nächste
    // Zeichnen ersetzt hat, und der Klick geht ins Leere.
    await screen.findByRole('button', { name: '1 Set' })
    await effekteNachholen()

    for (const [name, ziel] of [
      ['1 Set', 'sets'],
      // **Die Kartenzahl führt zu den Sets** (Befund 3): sie zählt über
      // alle, die Kartenlasche zeigt eines.
      ['107 Karten', 'sets'],
      ['3 Exemplare', 'sammlung'],
      ['1 Deck', 'decks'],
    ] as const) {
      await userEvent.click(screen.getByRole('button', { name }))

      expect(gewechselt).toHaveBeenLastCalledWith(ziel)
    }
  })
})
