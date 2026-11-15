// **Die Farben des Tisches** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import {
  ausDemDokument,
  dunklesBild,
  farbzahl,
  mitTischbild,
  Tischersatz,
  tischfarben,
} from './tischfarben'

describe('farbzahl', () => {
  it('liest ein Hexmerkmal, wie es in stil.css steht', () => {
    expect(farbzahl('#d4bb94', 0)).toBe(0xd4bb94)
    expect(farbzahl('  #221B13  ', 0)).toBe(0x221b13)
  })

  it('liest die Kurzform', () => {
    expect(farbzahl('#abc', 0)).toBe(0xaabbcc)
  })

  it('liest, was ein Browser aus einer Variablen macht', () => {
    expect(farbzahl('rgb(212, 187, 148)', 0)).toBe(0xd4bb94)
    expect(farbzahl('rgb(212 187 148 / 0.5)', 0)).toBe(0xd4bb94)
  })

  it('nimmt den Ersatz, wo nichts Lesbares steht', () => {
    for (const nichts of [
      '',
      '   ',
      'var(--fehlt)',
      'oklch(0.7 0.1 80)',
      '#12',
    ])
      expect(farbzahl(nichts, 0x123456)).toBe(0x123456)
  })
})

describe('tischfarben', () => {
  it('holt jede Farbe unter ihrem Namen', () => {
    const gefragt: string[] = []
    const farben = tischfarben((name) => {
      gefragt.push(name)
      return '#010203'
    })

    expect(gefragt).toEqual([
      '--mse-tisch-grund',
      '--mse-tisch-sitz',
      '--mse-tisch-eigen',
      '--mse-tisch-kante',
      '--mse-tisch-ruecken',
      '--mse-tisch-wartet',
      '--mse-tisch-auswahl',
      '--mse-tisch-schrift',
    ])
    expect(farben.grund).toBe(0x010203)
    expect(farben.wartet).toBe(0x010203)
  })

  it('fällt auf den hellen Satz zurück, wo nichts zu lesen ist', () => {
    // **Nicht auf Schwarz.** Ein Tisch, dessen Merkmale fehlen, wäre sonst
    // eine schwarze Fläche — und niemand sähe, dass etwas fehlt.
    expect(tischfarben(() => '')).toEqual(Tischersatz)
  })
})

describe('ausDemDokument', () => {
  it('fragt den Wurzelknoten nach dem Merkmal', () => {
    document.documentElement.style.setProperty('--mse-tisch-grund', '#0a0b0c')

    expect(farbzahl(ausDemDokument(document)('--mse-tisch-grund'), 0)).toBe(
      0x0a0b0c,
    )

    document.documentElement.style.removeProperty('--mse-tisch-grund')
  })

  it('gibt eine leere Zeichenkette, wo das Merkmal fehlt', () => {
    expect(ausDemDokument(document)('--gibt-es-nicht')).toBe('')
  })

  it('kommt mit einem Dokument ohne Fenster zurecht', () => {
    // **Ein Dokument ohne `defaultView` gibt es wirklich** — jedes, das nicht
    // in einem Fenster hängt. Ohne Riegel käme dort kein Ersatz, sondern ein
    // Wurf.
    const ohneFenster = document.implementation.createHTMLDocument()

    expect(ausDemDokument(ohneFenster)('--mse-tisch-grund')).toBe('')
  })
})

describe('dunklesBild', () => {
  it('erkennt einen dunklen Grund', () => {
    expect(dunklesBild(() => '#15120e')).toBe(true)
  })

  it('erkennt einen hellen Grund', () => {
    expect(dunklesBild(() => '#f2e8d5')).toBe(false)
  })

  it('nimmt hell an, wo nichts zu lesen ist', () => {
    // **Der Ersatz ist die Kartenfarbe des hellen Satzes** — sie ist hell,
    // und damit fällt die Antwort auf das Erscheinungsbild, das ohne jede
    // Angabe gilt.
    expect(dunklesBild(() => '')).toBe(false)
  })
})

describe('mitTischbild', () => {
  const einBild = {
    name: 'dungeon' as const,
    hell: { grund: 0x111111, sitz: 0x222222, eigen: 0x333333 },
    dunkel: { grund: 0x444444, sitz: 0x555555, eigen: 0x666666 },
  }

  it('tauscht die drei Flächenfarben im hellen Satz', () => {
    const tafel = mitTischbild(Tischersatz, einBild, false)

    expect(tafel.grund).toBe(0x111111)
    expect(tafel.sitz).toBe(0x222222)
    expect(tafel.eigenerSitz).toBe(0x333333)
  })

  it('nimmt den dunklen Satz, wo das Erscheinungsbild dunkel ist', () => {
    expect(mitTischbild(Tischersatz, einBild, true).grund).toBe(0x444444)
  })

  // **Kante, Rücken, Wartefarbe und Auswahl tragen die Zugänglichkeit** — ein
  // Tischbild darf sie nicht anfassen.
  it('lässt alles andere, wie es war', () => {
    const tafel = mitTischbild(Tischersatz, einBild, false)

    expect(tafel.kante).toBe(Tischersatz.kante)
    expect(tafel.ruecken).toBe(Tischersatz.ruecken)
    expect(tafel.wartet).toBe(Tischersatz.wartet)
    expect(tafel.auswahl).toBe(Tischersatz.auswahl)
  })
})
