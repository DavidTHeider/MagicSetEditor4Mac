// **Die Symmetrie im Symboleditor**
// (#1092).
//
// **Der Befund war, dass es keinen Aufrufer gab**: `service.SymmetrieAbbilden`
// war portiert und geprüft, und der Editor bot weder Art noch Kopienzahl an.
// Diese Läufe halten das Gegenteil fest.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { mitKurve, mitSymmetrie, zweiKopien } from '../test/symbolproben'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

// leeresSet ist die Antwort auf `Setangaben`, wenn kein Symbol abgelegt ist.
const leeresSet = { felder: [], werte: [], fehler: null }

describe('Die Symmetrie', () => {
  it('lässt sich aus einem gewählten Teil anlegen', async () => {
    const SymmetrieAnlegen = vi.fn().mockResolvedValue(mitSymmetrie())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue(leeresSet),
      Grundform: vi.fn().mockResolvedValue(mitKurve()),
      SymmetrieAnlegen,
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.kreis }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.symmetrieAnlegen }),
    )

    await waitFor(() => {
      expect(SymmetrieAnlegen).toHaveBeenCalled()
    })

    // **Und danach stehen die Bedienelemente da** — Art und Kopienzahl.
    expect(
      await screen.findByRole('button', { name: texte.drehung }),
    ).toBeInTheDocument()

    expect(screen.getByLabelText(texte.kopien)).toHaveValue(2)
  })

  it('zeigt ihre Kopien und ihre beiden Anfasser', async () => {
    const Symmetriekopien = vi.fn().mockResolvedValue(zweiKopien())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    // **Die Kopien werden geholt** — der produktive Aufrufer von
    // `service.SymmetrieAbbilden`.
    await waitFor(() => {
      expect(Symmetriekopien).toHaveBeenCalled()
    })

    // **Mitte und Griff sind eigene Anfasser** — und beide erreichbar.
    expect(
      await screen.findByRole('button', { name: texte.mitte }),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: texte.griff }),
    ).toBeInTheDocument()
  })

  it('lässt Art und Kopienzahl mit der Tastatur ändern', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
      SymmetrieSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    // **Über die Tastatur und nicht über einen Klick** (§6): der Knopf wird
    // angesteuert und mit der Eingabetaste ausgelöst.
    const achse = await screen.findByRole('button', { name: texte.achse })
    achse.focus()
    await userEvent.keyboard('{Enter}')

    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ spiegelart: 'reflection', teil: 0 }),
      )
    })

    // **Und die Kopienzahl über das Zahlenfeld.**
    //
    // **`fireEvent.change` und nicht `userEvent.type`.** Ein Drehfeld setzt
    // seinen Wert als Ganzes; `type` hängt an, und der Wert, an den es anhängt,
    // kommt von der Go-Seite zurück — aus einer 2 und einer getippten 4 würde
    // 24. Derselbe Weg steht in `autoersetzung.test.tsx`.
    fireEvent.change(screen.getByLabelText(texte.kopien), {
      target: { value: '4' },
    })

    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ kopien: 4 }),
      )
    })
  })

  it('bewegt die Mitte mit den Pfeiltasten', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
      SymmetrieSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    const mitte = await screen.findByRole('button', { name: texte.mitte })
    mitte.focus()
    await userEvent.keyboard('{ArrowRight}')

    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ mitteSetzen: true, mitteX: 0.51 }),
      )
    })

    // **Und die Mitte allein** — der Griff bleibt unberührt.
    expect(SymmetrieSetzen).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ griffSetzen: false }),
    )
  })

  it('löst sich auf und lässt ihre Kinder stehen', async () => {
    const SymmetrieAufloesen = vi.fn().mockResolvedValue(mitKurve())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
      SymmetrieAufloesen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.symmetrieAufloesen }),
    )

    await waitFor(() => {
      expect(SymmetrieAufloesen).toHaveBeenCalled()
    })
  })

  it('meldet einen Fehler und behält die Zeichnung', async () => {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
      SymmetrieSetzen: vi.fn().mockResolvedValue({
        teile: [],
        name: '',
        fehler: { code: 'validierung', meldung: '99 Kopien gehen nicht' },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.achse }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '99 Kopien gehen nicht',
    )

    // **Und die Symmetrie steht noch da** (§7, Regel 3).
    expect(screen.getByLabelText(texte.kopien)).toHaveValue(2)
  })
})

describe('Wenn die Kopien nicht zu holen sind', () => {
  it('meldet den Fehler und lässt die Zeichnung stehen', async () => {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue({
        kopien: [],
        fehler: { code: 'validierung', meldung: 'das ist keine Symmetrie' },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'das ist keine Symmetrie',
    )

    // **Und die Symmetrie steht noch da** — ein Fehler beim Zeichnen der
    // Kopien nimmt niemandem seine Arbeit (§7, Regel 3).
    expect(screen.getByLabelText(texte.kopien)).toHaveValue(2)
  })
})

describe('Der Griff der Symmetrie', () => {
  it('lässt sich mit den Pfeiltasten bewegen — und die Mitte bleibt', async () => {
    const SymmetrieSetzen = vi.fn().mockResolvedValue(mitSymmetrie())

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'probe.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue(mitSymmetrie()),
      Symmetriekopien: vi.fn().mockResolvedValue(zweiKopien()),
      SymmetrieSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="1" />
      </Sprachraum>,
    )

    const griff = await screen.findByRole('button', { name: texte.griff })
    griff.focus()
    await userEvent.keyboard('{ArrowDown}')

    // **Der Griff steht relativ zur Mitte**: er sitzt bei (0,75 / 0,5), eine
    // Taste tiefer bei (0,75 / 0,51) — und relativ macht das (0,25 / 0,01).
    await waitFor(() => {
      expect(SymmetrieSetzen).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({ griffSetzen: true, mitteSetzen: false }),
      )
    })
  })
})
