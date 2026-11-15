// **Die Gestalt der Skriptkonsole** — UXR-08
// (#1916).
//
// # Der Befund
//
// Gemessen am 2026-10-02 (`docs/UIREVIEW.md`, Commit ddadb180): 20
// Bedienelemente, 14 im Bild, **kein** Hauptknopf, sieben Überschriften.
// *Auswerten* sah aus wie jedes der neun Beispiele daneben. Und das Ergebnis
// stand **unter** den Beispielen und der Namensliste — wer die Namen offen
// hatte, sah nach dem Auswerten sechzig Knöpfe und kein Ergebnis.
//
// # Was hier gilt
//
// Eine führende Handlung (R1), das Ergebnis gleich unter ihr, und die zwei
// Hilfen für den, der die Sprache noch nicht kennt, eingeklappt dahinter.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KonsolenAntwort } from '../antworten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { leere } from '../skriptverlauf'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { Skriptkonsole } from './Skriptkonsole'

const texte = kataloge.de.seiten.karten.konsole

const antwort: KonsolenAntwort = {
  quelltext: '"ab"',
  text: 'ab',
  art: 'text',
  befunde: [],
  stelle: null,
  dauerMs: 3,
  abgebrochen: false,
  eintraege: [],
  fehler: null,
}

function zeige(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({ Skript: () => Promise.resolve(antwort), ...teile })

  return render(
    <Sprachraum sprache="de">
      <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
    </Sprachraum>,
  )
}

// folgt sagt, ob `spaeter` im Dokument nach `frueher` steht.
function folgt(frueher: Element, spaeter: Element): boolean {
  return (
    (frueher.compareDocumentPosition(spaeter) &
      Node.DOCUMENT_POSITION_FOLLOWING) !==
    0
  )
}

afterEach(() => {
  raeumeApi()
  leere()
})

describe('Die Gestalt der Skriptkonsole', () => {
  // **R1: eine führende Handlung.** Vorher trug kein Knopf das Hauptgewicht.
  it('führt mit Auswerten und mit nichts sonst', () => {
    zeige()

    const auswerten = screen.getByRole('button', { name: texte.auswerten })

    expect(auswerten).toHaveClass('bg-marke')
    expect(
      screen
        .getAllByRole('button')
        .filter((knopf) => knopf.classList.contains('bg-marke')),
    ).toEqual([auswerten])
  })

  // **Das Ergebnis gehört unter den Knopf**, der es hervorbringt — nicht
  // unter zwei Hilfen, die man beim Auswerten nicht braucht.
  it('zeigt das Ergebnis vor Beispielen und Namen', async () => {
    zeige()

    await userEvent.type(screen.getByLabelText(texte.ausdruck), '"a" + "b"')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    const ergebnis = await screen.findByRole('heading', {
      name: texte.ergebnis,
    })

    expect(
      folgt(ergebnis, screen.getByRole('button', { name: texte.beispiele })),
    ).toBe(true)
    expect(
      folgt(ergebnis, screen.getByRole('button', { name: texte.namen })),
    ).toBe(true)
  })

  // **Die Beispiele sind eingeklappt** — neun Knöpfe in zwei Gruppen, für
  // den ersten Ausdruck. Aufgeklappt setzen sie ein wie vorher.
  it('hält die Beispiele eingeklappt, bis jemand sie öffnet', async () => {
    zeige()

    const knopf = screen.getByRole('button', { name: texte.beispiele })

    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: /1 \+ 1/ }),
    ).not.toBeInTheDocument()

    await userEvent.click(knopf)

    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    await userEvent.click(
      screen.getByRole('button', {
        name: `${texte.beispielListe.rechnung} — 1 + 1`,
      }),
    )

    expect(screen.getByLabelText(texte.ausdruck)).toHaveValue('1 + 1')
  })

  // **Die Namen ebenso** — und Zuklappen wirft sie nicht weg: wer sie wieder
  // öffnet, bekommt dieselbe Liste ohne zweiten Ruf.
  it('holt die Namen beim ersten Öffnen und behält sie', async () => {
    const Skriptnamen = vi.fn(() =>
      Promise.resolve({ namen: ['card', 'length'], fehler: null }),
    )

    zeige({ Skriptnamen })

    const knopf = screen.getByRole('button', { name: texte.namen })

    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(Skriptnamen).not.toHaveBeenCalled()

    await userEvent.click(knopf)

    expect(
      await screen.findByRole('button', { name: 'length' }),
    ).toBeInTheDocument()

    await userEvent.click(knopf)
    expect(
      screen.queryByRole('button', { name: 'length' }),
    ).not.toBeInTheDocument()

    await userEvent.click(knopf)
    expect(screen.getByRole('button', { name: 'length' })).toBeInTheDocument()

    await waitFor(() => {
      expect(Skriptnamen).toHaveBeenCalledTimes(1)
    })
  })

  // **axe mit allem offen** — Ergebnis, beide Kästen aufgeklappt (WCAG 2.2
  // AA; Kontraste sieht axe unter jsdom nicht, §3).
  it('hält WCAG mit offenen Hilfen und einem Ergebnis ein', async () => {
    const { container } = zeige({
      Skriptnamen: () =>
        Promise.resolve({ namen: ['card', 'length'], fehler: null }),
    })

    await userEvent.type(screen.getByLabelText(texte.ausdruck), '"a"')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))
    await screen.findByRole('heading', { name: texte.ergebnis })
    await userEvent.click(screen.getByRole('button', { name: texte.beispiele }))
    await userEvent.click(screen.getByRole('button', { name: texte.namen }))
    await screen.findByRole('button', { name: 'length' })

    expect(await verstoesse(container)).toEqual([])
  })
})
