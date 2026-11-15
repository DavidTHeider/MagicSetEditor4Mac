// **Die Diagnose im Fenster** — Werkstück 4 von Phase 3.5
// (#1166).
//
// Geprüft wird die **Zustimmung**: der Text erscheint erst auf Wunsch, und
// weitergegeben wird er erst danach.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Diagnose } from './Diagnose'

const texte = kataloge.de.seiten.technik.diagnose

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Diagnose />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

describe('die Diagnose', () => {
  // **Ohne Klick geschieht nichts.**
  //
  // Das ist die Zusage des Werkstücks: sie wird gezeigt, bevor sie irgendwohin
  // geht. Ein Weg, der beim Öffnen der Seite schon läde, hätte diese Zusage
  // nicht — und den Unterschied sieht man nur hier.
  it('holt nichts, bevor jemand darum bittet', async () => {
    const geholt = vi
      .fn()
      .mockResolvedValue({ text: 'MSE4Mac 0.2.0', fehler: null })
    belegeApi({ Diagnose: geholt })

    zeige()

    expect(await screen.findByText(kurzform(texte.inhalt))).toBeInTheDocument()
    expect(geholt).not.toHaveBeenCalled()

    // **Der Beleg, dass jemand zuhört** (§3): derselbe Weg mit einem Klick
    // ruft. Ohne ihn wäre die Verneinung auch bei einem Knopf grün, der nichts
    // tut.
    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))

    expect(geholt).toHaveBeenCalledOnce()
  })

  // **Und danach steht der Text da.**
  it('zeigt den Text und bietet erst dann das Weitergeben an', async () => {
    belegeApi({
      Diagnose: vi.fn().mockResolvedValue({
        text: 'MagicSetEditor4Mac 0.2.0\ndarwin/arm64',
        fehler: null,
      }),
    })

    zeige()

    // **Vorher gibt es die beiden Knöpfe nicht.** Ein »Kopieren«, das nichts
    // zu kopieren hat, ist eine Einladung ins Leere.
    expect(
      screen.queryByRole('button', { name: texte.kopieren }),
    ).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))

    expect(await screen.findByText(/darwin\/arm64/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: texte.kopieren }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: texte.sichern }),
    ).toBeInTheDocument()
  })

  // **Kopieren legt den Text in die Zwischenablage.**
  it('kopiert den Text', async () => {
    const geschrieben = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: geschrieben },
    })

    belegeApi({
      Diagnose: vi
        .fn()
        .mockResolvedValue({ text: 'die Diagnose', fehler: null }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.kopieren }),
    )

    expect(geschrieben).toHaveBeenCalledWith('die Diagnose')
    expect(await screen.findByText(texte.kopiert)).toBeInTheDocument()
  })

  // **Sichern geht durch die Bindung.**
  it('sichert über den Dateidialog', async () => {
    const gesichert = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Diagnose: vi
        .fn()
        .mockResolvedValue({ text: 'die Diagnose', fehler: null }),
      DiagnoseSichern: gesichert,
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.sichern }),
    )

    expect(gesichert).toHaveBeenCalledOnce()
  })

  // **Ein Fehler erscheint und verdeckt nichts.**
  it('meldet einen Fehler', async () => {
    belegeApi({
      Diagnose: vi.fn().mockResolvedValue({
        text: '',
        fehler: {
          code: 'unbekannt',
          schluessel: '',
          meldung: 'nicht zu lesen',
        },
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))

    expect(await screen.findByRole('alert')).toHaveTextContent('nicht zu lesen')
  })

  // **Und ein Fehler beim Sichern ebenso.**
  it('meldet einen Fehler beim Sichern', async () => {
    belegeApi({
      Diagnose: vi
        .fn()
        .mockResolvedValue({ text: 'die Diagnose', fehler: null }),
      DiagnoseSichern: vi.fn().mockResolvedValue({
        fehler: { code: 'validierung', schluessel: '', meldung: 'kein Platz' },
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.sichern }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Platz')
  })

  // **Und sagt es, wenn die Zwischenablage nicht darf** — Befund 5 des
  // Nachlaufs zum Zweitblick über Phase 5.1: hier blieb eine abgewiesene
  // Zusage ohne Empfänger, und die Seite zeigte nichts.
  it('meldet, wenn das Kopieren nicht geht', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: {
        writeText: vi.fn().mockRejectedValue(new Error('keine Erlaubnis')),
      },
    })

    belegeApi({
      Diagnose: vi
        .fn()
        .mockResolvedValue({ text: 'die Diagnose', fehler: null }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.anzeigen }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.kopieren }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.kopierfehler,
    )

    // **Und „Kopiert" steht nicht da** — die Zusage gilt nicht.
    expect(screen.queryByText(texte.kopiert)).not.toBeInTheDocument()
  })
})
