// **Die Liste hat immer genau eine Tabstelle** — Befund 5 des Zweitblicks
// über Phase 4.3
// (#1290).
//
// # Der Befund
//
// Die Tabstelle lag auf der **gewählten** Karte. Die Liste zeichnet aber nur
// ihr Sichtfenster (#891) — wer von der gewählten Karte wegrollt, hat sie
// nicht mehr im Baum, und **keine** gezeichnete Zeile nimmt den Tabulator an.
//
// Der Zweitblick hat es nachgestellt: 200 Karten, Auswahl 0, zu Karte 92
// gerollt — **null** Elemente mit `tabindex="0"`.
//
// **Auswirkung**: das Gitter ist per Tabulator nicht mehr erreichbar. Nach
// einem Rollvorgang braucht es die Maus, um wieder einzusteigen — und genau
// das sollte diese Änderung abschaffen.
//
// # Die Regel danach
//
// **Die gewählte Karte, wenn sie gezeichnet ist; sonst die erste gezeichnete.**
// Das ist die Gestalt, die jedes Gitter hat: der Tabulator führt hinein, und
// wohin genau, entscheidet die Liste.

import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { Kartenliste } from './Kartenliste'

function karte(id: string, nummer: number, name: string): KartenzeileDTO {
  return { id, nummer, name, felder: 1 }
}

function zeige(
  karten: readonly KartenzeileDTO[],
  gewaehlt: string | undefined,
) {
  render(
    <Sprachraum sprache="de">
      <Kartenliste
        spalten={[]}
        marken={{}}
        setID="7"
        stand={0}
        karten={karten}
        gewaehlt={gewaehlt}
        waehlen={vi.fn()}
        angehakt={[]}
        anhaken={vi.fn()}
        mitKaestchen={false}
      />
    </Sprachraum>,
  )
}

function tabstellen() {
  return [...screen.getByRole('grid').querySelectorAll('[tabindex="0"]')]
}

const drei = [
  karte('1', 0, 'Erste'),
  karte('2', 1, 'Zweite'),
  karte('3', 2, 'Dritte'),
]

describe('Die Tabstelle der Kartenliste', () => {
  it('liegt auf der gewählten Karte, wenn sie gezeichnet ist', () => {
    zeige(drei, '2')

    expect(tabstellen()).toHaveLength(1)
    expect(tabstellen()[0]).toHaveAccessibleName(/Zweite/)
  })

  it('fällt auf die erste gezeichnete Zeile, wenn die gewählte fehlt', () => {
    // **Der Fall des Zweitblicks**, in klein: die gewählte Karte steht nicht
    // in der Liste — so wie sie beim Rollen nicht im Sichtfenster steht.
    zeige(drei, 'nicht gezeichnet')

    expect(tabstellen()).toHaveLength(1)
    expect(tabstellen()[0]).toHaveAccessibleName(/Erste/)
  })

  it('hat auch ohne Auswahl eine Tabstelle', () => {
    zeige(drei, undefined)

    expect(tabstellen()).toHaveLength(1)
  })

  it('hat in einer leeren Liste keine — und das ist kein Mangel', () => {
    // **Ein Gitter ohne Zeilen ist nichts, worin man sich bewegt.** Der
    // Leerzustand daneben sagt, was zu tun ist (§6).
    zeige([], undefined)

    expect(tabstellen()).toHaveLength(0)
  })
})
