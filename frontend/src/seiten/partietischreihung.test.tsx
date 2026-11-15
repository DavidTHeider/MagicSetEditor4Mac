// **Die Reihung der Auslöser am Tisch** — Werkstück M9 der Regelmaschine
// (#1611).
//
// # Warum sie einen eigenen Lauf hat
//
// **Weil sie das Spiel anhält** — derselbe Grund wie bei der Legendenwahl
// nebenan. CR 603.3 legt die Fähigkeiten auf den Stapel, **bevor** das
// nächste Mal jemand Priorität bekäme; solange geordnet wird, hat **niemand**
// eine. Ein Tisch, der die Frage nicht zeigt, sieht aus wie ein Ausfall.
import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { einLauf, einePartie, element } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

// zweiAusloeser gibt eine Partie mit zwei wartenden Fähigkeiten desselben
// Sitzes — der Fall aus CR 603.3b.
function zweiAusloeser() {
  return einePartie({
    ausloeserreihung: 'sitz1',
    ausgeloest: [
      {
        kennung: 'sitz1/monarch.wechsel/0',
        art: 'monarch.wechsel',
        beherrscher: 'sitz1',
        wen: 'sitz2',
      },
      {
        kennung: 'sitz1/initiative.wechsel/0',
        art: 'initiative.wechsel',
        beherrscher: 'sitz1',
        wen: 'sitz2',
      },
    ],
  })
}

describe('die Reihung am Tisch', () => {
  test('nennt, wer ordnen muss, und die Regel dazu', () => {
    zeige(einLauf(), zweiAusloeser())

    const meldung = screen.getByRole('status', {
      name: kataloge.de.seiten.spielen.reihungTitel,
    })

    expect(meldung).toHaveTextContent('Anna')
    expect(meldung).toHaveTextContent('CR 603.3b')
  })

  test('und stellt je wartender Fähigkeit einen Knopf', async () => {
    const befehlen = vi.fn()
    const nutzer = zeige(einLauf({ befehlen }), zweiAusloeser())

    const knoepfe = screen.getAllByRole('button', {
      name: new RegExp(kataloge.de.seiten.spielen.reihungLegen),
    })

    expect(knoepfe).toHaveLength(2)
    expect(element(knoepfe[1])).toHaveTextContent(
      kataloge.de.seiten.handlungen.ausloesername['initiative.wechsel'],
    )

    await nutzer.click(element(knoepfe[1]))

    expect(befehlen).toHaveBeenCalledWith('spieler.reihung', {
      kennung: 'sitz1/initiative.wechsel/0',
    })
  })

  test('und zeigt nur die des Gefragten', () => {
    const partie = zweiAusloeser()

    zeige(
      einLauf(),
      einePartie({
        ausloeserreihung: 'sitz1',
        ausgeloest: [
          ...partie.ausgeloest,
          {
            kennung: 'sitz2/strahlung/0',
            art: 'strahlung',
            beherrscher: 'sitz2',
            wen: '',
          },
        ],
      }),
    )

    expect(
      screen.getAllByRole('button', {
        name: new RegExp(kataloge.de.seiten.spielen.reihungLegen),
      }),
    ).toHaveLength(2)
  })

  test('und nennt einen ausgetragenen Sitz mit seiner Kennung', () => {
    // **Ein Befund und kein Fehler** (§8): dieselbe Gestalt wie bei der
    // Legendenwahl — ein von Hand geänderter Bestand kann den Sitz
    // ausgetragen haben, und die Frage bleibt lesbar.
    zeige(
      einLauf(),
      einePartie({
        ausloeserreihung: 'weg',
        ausgeloest: [
          {
            kennung: 'weg/strahlung/0',
            art: 'strahlung',
            beherrscher: 'weg',
            wen: '',
          },
        ],
      }),
    )

    expect(
      screen.getByRole('status', {
        name: kataloge.de.seiten.spielen.reihungTitel,
      }),
    ).toHaveTextContent('weg')
  })

  test('und eine unbekannte Art steht mit ihrer Kennung da', () => {
    // **Derselbe Rückfall wie bei jedem Wort des Kerns** (§8): `M8` bringt
    // beliebig viele Auslöserarten, und bis das Fenster sie kennt, ist die
    // Kennung besser als eine leere Beschriftung.
    zeige(
      einLauf(),
      einePartie({
        ausloeserreihung: 'sitz1',
        ausgeloest: [
          {
            kennung: 'sitz1/neu/0',
            art: 'neu.und.unbekannt',
            beherrscher: 'sitz1',
            wen: '',
          },
        ],
      }),
    )

    expect(
      screen.getByRole('button', {
        name: new RegExp(kataloge.de.seiten.spielen.reihungLegen),
      }),
    ).toHaveTextContent('neu.und.unbekannt')
  })

  test('und schweigt, wo niemand gefragt ist', () => {
    zeige(einLauf(), einePartie())

    expect(
      screen.queryByRole('status', {
        name: kataloge.de.seiten.spielen.reihungTitel,
      }),
    ).toBeNull()
  })
})
