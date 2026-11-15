// **Zwei Symbolhandlungen zugleich, an der Seite** — Befund 2 des Zweitblicks
// über Phase 4.11.
//
// # Warum dieser Lauf neben `useZeichenschutz.test.tsx` steht
//
// Dort steht, was der Wächter tut. **Hier steht, dass die Seite ihn
// benutzt** — an jeder der Stellen, die den ganzen Zeichenstand hinausschicken
// und die Antwort ganz einsetzen. Zwei grüne Seiten einer Grenze sind kein
// Nachweis für die Grenze.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: false,
    sperre: 'free',
  }
}

function teil(name: string) {
  return {
    art: 'Form',
    name,
    kombination: 'overlap',
    punkte: [punkt(0.5, 0.1), punkt(0.9, 0.8), punkt(0.1, 0.8)],
    spiegelart: '',
    kopien: 0,
    mitteX: 0,
    mitteY: 0,
    griffX: 0,
    griffY: 0,
    teile: [],
  }
}

describe('Zwei Handlungen zugleich', () => {
  // **Regel 4 aus §7**: ein Vorgang je Datensatz.
  //
  // Beide Rufe gingen gegen den Stand von **vorher**; die zweite Antwort
  // verlöre, was die erste wollte. Der Wächter lässt die zweite gar nicht
  // erst los.
  it('lässt nur eine Grundform zugleich anlegen', async () => {
    let ersteAufloesen: (wert: unknown) => void = () => undefined

    const Grundform = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((aufloesen) => {
            ersteAufloesen = aufloesen
          }),
      )
      .mockResolvedValue({ teile: [teil('Zweites')], name: '', fehler: null })

    belegeApi({
      Grundform,
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    const dreieck = await screen.findByRole('button', { name: texte.dreieck })
    const stern = screen.getByRole('button', { name: texte.stern })

    await userEvent.click(dreieck)
    await userEvent.click(stern)

    // **Der zweite Ruf ging gar nicht hinaus.**
    expect(Grundform).toHaveBeenCalledTimes(1)

    // **Und danach geht wieder etwas** — der Wächter hält an, er sperrt nicht.
    ersteAufloesen({ teile: [teil('Erstes')], name: '', fehler: null })

    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Erstes' }),
    ).toBeInTheDocument()

    await userEvent.click(stern)

    expect(Grundform).toHaveBeenCalledTimes(2)
  })

  // **Und dasselbe für das Nachzeichnen** — dieselbe Gestalt, andere Bindung.
  it('lässt Nachzeichnen nicht neben einer Grundform laufen', async () => {
    const Grundform = vi
      .fn()
      .mockImplementation(() => new Promise(() => undefined))
    const SymbolNachzeichnen = vi.fn().mockResolvedValue({
      teile: [teil('Nachgezeichnet')],
      name: '',
      fehler: null,
    })

    belegeApi({
      Grundform,
      SymbolNachzeichnen,
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.nachzeichnen }),
    )

    expect(Grundform).toHaveBeenCalledTimes(1)
    expect(SymbolNachzeichnen).not.toHaveBeenCalled()
  })

  // **Eine Antwort, die nach dem Wechsel eintrifft, trägt sich nicht ein.**
  //
  // §7 Regel 2 an den beiden Wegen der Seite selbst. Der Wechsel ist hier ein
  // anderes Set: die Seite wird mit einem neuen `key` neu gebaut, und die
  // Antwort von vorhin gehört zu einem Stand, den niemand mehr sieht.
  // **Ein Wechsel des Bezugsobjekts verwirft, was unterwegs ist** (§7, Regel
  // 1 und 2). Der Editor bekommt ein anderes Ziel — ein Kartensymbol statt des
  // Setsymbols —, und die Antwort von vorhin gehört zu einem Stand, den
  // niemand mehr sieht.
  it('verwirft eine Grundform nach einem Zielwechsel', async () => {
    let aufloesen: (wert: unknown) => void = () => undefined

    const Grundform = vi.fn().mockImplementation(
      () =>
        new Promise((fertig) => {
          aufloesen = fertig
        }),
    )

    belegeApi({
      Grundform,
      Symbol: vi.fn().mockResolvedValue({ teile: [], name: '', fehler: null }),
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
      Karte: vi.fn().mockResolvedValue({
        karte: { id: '3', felder: [], werte: [] },
        fehler: null,
      }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    // **Der Editor bekommt ein anderes Ziel**, während der Ruf läuft.
    rerender(
      <Sprachraum sprache="de">
        <Symbol setID="7" ziel={{ karteID: '3', feld: 'symbol' }} />
      </Sprachraum>,
    )

    aufloesen({ teile: [teil('Zu spät')], name: '', fehler: null })

    await screen.findByRole('button', { name: texte.dreieck })

    expect(
      screen.queryByRole('button', { name: 'Punkt 1 von Zu spät' }),
    ).not.toBeInTheDocument()
  })

  it('verwirft eine Grundform, deren Stand vorbei ist', async () => {
    let aufloesen: (wert: unknown) => void = () => undefined

    const Grundform = vi.fn().mockImplementation(
      () =>
        new Promise((fertig) => {
          aufloesen = fertig
        }),
    )

    belegeApi({
      Grundform,
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    })

    const { unmount } = render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )

    // **Die Seite geht, während der Ruf läuft.**
    unmount()

    aufloesen({ teile: [teil('Zu spät')], name: '', fehler: null })

    // **Nichts stürzt ab, und nichts trägt sich ein** — die Seite ist weg.
    expect(
      screen.queryByRole('button', { name: 'Punkt 1 von Zu spät' }),
    ).not.toBeInTheDocument()
  })

  // **Und dasselbe für das Nachzeichnen** — derselbe Zweig, andere Bindung.
  it('verwirft ein Nachzeichnen nach einem Zielwechsel', async () => {
    let aufloesen: (wert: unknown) => void = () => undefined

    const SymbolNachzeichnen = vi.fn().mockImplementation(
      () =>
        new Promise((fertig) => {
          aufloesen = fertig
        }),
    )

    belegeApi({
      SymbolNachzeichnen,
      Symbol: vi.fn().mockResolvedValue({ teile: [], name: '', fehler: null }),
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
      Karte: vi.fn().mockResolvedValue({
        karte: { id: '3', felder: [], werte: [] },
        fehler: null,
      }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.nachzeichnen }),
    )

    rerender(
      <Sprachraum sprache="de">
        <Symbol setID="7" ziel={{ karteID: '3', feld: 'symbol' }} />
      </Sprachraum>,
    )

    aufloesen({ teile: [teil('Zu spät')], name: '', fehler: null })

    await screen.findByRole('button', { name: texte.dreieck })

    expect(
      screen.queryByRole('button', { name: 'Punkt 1 von Zu spät' }),
    ).not.toBeInTheDocument()
  })
})
