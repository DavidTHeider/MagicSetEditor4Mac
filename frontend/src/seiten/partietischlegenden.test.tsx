// **Die Legendenwahl am Tisch** — Werkstück M12 der Regelmaschine
// (#1614).
//
// # Warum sie einen eigenen Lauf hat
//
// **Weil sie das Spiel anhält.** CR 704.5j lässt die Partie stehen, bis
// jemand gewählt hat; ein Tisch, der die Frage nicht zeigt, sieht aus wie
// ein Ausfall. Das findet kein Wächter, der Bindungen zählt — nur ein Lauf,
// der nachsieht, ob die Frage ankommt.
import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { einLauf, einObjekt, einePartie, element } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

// zweiLegenden gibt eine Partie mit zwei gleichnamigen Legenden und der
// offenen Frage dazu.
function zweiLegenden() {
  return einePartie({
    objekte: [
      einObjekt({
        kennung: 'sitz1/held/0',
        name: 'Held',
        zone: 'tisch',
        beherrscher: 'sitz1',
        geheim: false,
      }),
      einObjekt({
        kennung: 'sitz1/held/1',
        name: 'Held',
        zone: 'tisch',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
    legendenwahlen: [{ sitz: 'sitz1', name: 'Held' }],
  })
}

describe('die Legendenwahl am Tisch', () => {
  test('nennt, wer über welchen Namen wählen muss', () => {
    zeige(einLauf(), zweiLegenden())

    const meldung = screen.getByRole('status')

    expect(meldung).toHaveTextContent('Anna')
    expect(meldung).toHaveTextContent('Held')
    expect(meldung).toHaveTextContent('CR 704.5j')
  })

  test('und stellt je Kandidat einen Knopf', async () => {
    const befehlen = vi.fn()
    const nutzer = zeige(einLauf({ befehlen }), zweiLegenden())

    const knoepfe = screen.getAllByRole('button', {
      name: kataloge.de.seiten.spielen.legendeBehalten,
    })

    expect(knoepfe).toHaveLength(2)

    await nutzer.click(element(knoepfe[0]))

    expect(befehlen).toHaveBeenCalledWith('spieler.legende', {
      objekt: 'sitz1/held/0',
    })
  })

  test('und nennt einen ausgetragenen Sitz mit seiner Kennung', () => {
    // **Ein Befund und kein Fehler** (§8): die Frage hängt an einem Sitz,
    // und ein von Hand geänderter Bestand kann ihn ausgetragen haben. Die
    // Frage bleibt lesbar, statt mit einer Lücke dazustehen.
    zeige(
      einLauf(),
      einePartie({ legendenwahlen: [{ sitz: 'weg', name: 'Held' }] }),
    )

    expect(screen.getByRole('status')).toHaveTextContent('weg')
  })

  test('und schweigt, wo niemand gefragt ist', () => {
    zeige(einLauf(), einePartie())

    expect(
      screen.queryByRole('button', {
        name: kataloge.de.seiten.spielen.legendeBehalten,
      }),
    ).toBeNull()
  })

  test('und nennt den Beherrscher an jedem Objekt auf dem Tisch', () => {
    zeige(
      einLauf(),
      einePartie({
        objekte: [
          einObjekt({
            kennung: 'sitz2/baer/0',
            name: 'Bär',
            zone: 'tisch',
            beherrscher: 'sitz2',
            geheim: false,
          }),
        ],
      }),
    )

    // **Am Kartenknopf und nicht irgendwo** — der Name steht auch in der
    // Sitzliste, und ein `getByText` über beide sagte nichts darüber, dass
    // er **an der Karte** steht.
    expect(screen.getByRole('button', { name: /Bär/ })).toHaveTextContent(
      'Bert',
    )
  })
})

// **Und ein Sitz, den es nicht gibt, steht mit seiner Kennung da.**
//
// **Ein Befund und kein Fehler** (§8): ein Bestand von Hand geändert, ein
// Sitz ausgetragen — und der Tisch bleibt lesbar, statt eine leere Stelle
// zu zeigen.
test('ein unbekannter Beherrscher steht mit seiner Kennung da', () => {
  vi.stubGlobal('scrollTo', () => undefined)

  zeige(
    einLauf(),
    einePartie({
      objekte: [
        einObjekt({
          kennung: 'weg/baer/0',
          name: 'Bär',
          zone: 'tisch',
          beherrscher: 'weg',
          geheim: false,
        }),
      ],
    }),
  )

  expect(screen.getByRole('button', { name: /Bär/ })).toHaveTextContent('weg')
})
