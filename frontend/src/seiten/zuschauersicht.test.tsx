// **Der Tisch lässt sich als Zuschauer ansehen** — PLAY-05.AK-02
// (#1713).
//
// # Der Anlass
//
// **Die Sichtwahl führte nur Sitze.** Die Domäne kennt den Zuschauer seit
// G1: `Partie.Sicht("")` gibt die Sicht eines Betrachters **ohne Sitz** —
// offene Zonen und sonst nichts —, MCP sagt es ausdrücklich („eine
// unbekannte sieht so viel wie ein Zuschauer"), und `partie.zuschauer` steht
// im DTO. **Nur ansehen konnte sie niemand**: `Partietisch.tsx` baute die
// Auswahlliste aus `partie.sitze`.
//
// Das Artboard `TischSicht` nennt den Zweck: *„sonst verriete ein
// Zuschauerexport, was er nicht darf."* Wer die Sicht nicht aufrufen kann,
// kann sie vor einem Export auch nicht ansehen.
//
// # Was er nicht prüft
//
// **Dass die Sicht nichts verrät.** Das hält
// [`internal/domain/zuschauersicht_test.go`](../../../internal/domain/zuschauersicht_test.go)
// an der Regel selbst — hier geht es darum, dass die Fläche sie **anbietet**.
import { screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { einLauf, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen

describe('die Sichtwahl', () => {
  it('bietet den Zuschauer an', () => {
    zeige(einLauf())

    expect(
      screen.getByRole('option', { name: eigene.sichtZuschauer }),
    ).toBeInTheDocument()
  })

  // **Und die Wahl fragt mit leerer Sitzkennung** — das ist der Zuschauer.
  it('sieht mit leerer Sitzkennung nach', async () => {
    const ansehen = vi.fn()
    const partie = einePartie()
    const hand = zeige(einLauf({ ansehen }), partie)

    await hand.selectOptions(
      screen.getByLabelText(eigene.sicht),
      eigene.sichtZuschauer,
    )

    expect(ansehen).toHaveBeenCalledWith(partie.id, '')
  })

  // **Der Steuerlauf** (§10): ohne ihn sagte der Lauf darüber nicht, ob die
  // Sitze noch wählbar sind.
  it('bietet die Sitze weiterhin an', async () => {
    const ansehen = vi.fn()
    const partie = einePartie()
    const hand = zeige(einLauf({ ansehen }), partie)

    const zweiter = partie.sitze[1]

    await hand.selectOptions(
      screen.getByLabelText(eigene.sicht),
      zweiter?.kennung ?? '',
    )

    expect(ansehen).toHaveBeenCalledWith(partie.id, zweiter?.kennung)
  })
})
