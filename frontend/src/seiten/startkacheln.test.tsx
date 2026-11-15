// **Die Einführung zeigt immer alle drei Schritte**
// (#673).
//
// **Sie stand bis zum 2026-09-03 auf der Startseite** und ist seither ein
// Fenster beim Start (#1325). Die Zusage von #673 zieht mit um: sie gilt den
// **Schritten**, nicht dem Ort, an dem sie stehen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Einfuehrung } from '../bausteine/Einfuehrung'

const texte = kataloge.de.seiten.start

function zeige(setID: string | undefined) {
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Einfuehrung setID={setID} wechseln={wechseln} schliessen={vi.fn()} />
    </Sprachraum>,
  )

  return wechseln
}

describe('Die Einführung', () => {
  // **Die Zahl der Kacheln hängt nicht am Zustand.**
  //
  // Bis #673 verschwand die mittlere, solange kein Set offen war — übrig
  // blieben zwei in einem Raster für drei, und weil die **mittlere** fehlte,
  // sah es aus, als sei eine andere weggefallen.
  it('zeigt drei Schritte, ob ein Set offen ist oder nicht', () => {
    zeige(undefined)

    expect(screen.getAllByRole('listitem')).toHaveLength(3)
  })

  it('zeigt dieselben drei mit offenem Set', () => {
    zeige('7')

    expect(screen.getAllByRole('listitem')).toHaveLength(3)
  })

  // **Ein Schritt, den man noch nicht gehen kann, sagt warum** (§6) — und
  // führt dorthin, wo die Voraussetzung entsteht. Ihn auszublenden nähme dem
  // Anfänger die Landkarte.
  it('sagt beim Kartenschritt, dass zuerst ein Set fehlt', () => {
    zeige(undefined)

    expect(
      screen.getByText(texte.schritte.kartenBearbeiten.ersteinSet),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', {
        name: texte.schritte.kartenBearbeiten.stattdessen,
      }),
    ).toBeInTheDocument()
  })

  it('führt mit offenem Set unmittelbar zu den Karten', () => {
    zeige('7')

    expect(
      screen.queryByText(texte.schritte.kartenBearbeiten.ersteinSet),
    ).not.toBeInTheDocument()

    expect(
      screen.getByRole('button', {
        name: texte.schritte.kartenBearbeiten.handlung,
      }),
    ).toBeInTheDocument()
  })

  // **Und er führt dorthin, wo die Voraussetzung entsteht** — zu den Sets,
  // nicht zu den Karten, die es noch nicht gibt.
  it('führt vom Kartenschritt zu den Sets, solange keins offen ist', async () => {
    const wechseln = zeige(undefined)

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.schritte.kartenBearbeiten.stattdessen,
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('sets')
  })
})
