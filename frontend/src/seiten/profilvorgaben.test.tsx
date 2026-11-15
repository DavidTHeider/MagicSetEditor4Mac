// **Womit ein neuer Datensatz anfängt** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 2, 3 und 8.
//
// # Die Zusage
//
// **Eine Vorgabe ist erst eine, wenn sie im Formular steht.** Ein Wert im
// Bestand, den kein Feld zeigt, ist eine Einstellung ohne Wirkung — und
// genau das prüfen diese Läufe: den Weg von der Einstellung ins Formular.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Einstellungen } from './Einstellungen'
import { Sammlungsformular } from '../bausteine/Sammlungsformular'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { probeeinstellung } from '../test/einstellung'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from '../test/bestandspflegeprobe'

import type { EinstellungDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// vorgabezeilen sind die drei, so wie `internal/domain` sie ausgibt.
const vorgabezeilen: EinstellungDTO[] = [
  probeeinstellung({
    schluessel: 'vorgabe.zustand',
    wert: 'ohne',
    // **Alle acht, wie Go sie schickt** (`docs/EINSTELLUNGEN.md`): mit acht
    // Werten bleibt der Zustand eine Auswahlliste. Drei wären seit UXR-F2
    // (#2008) drei Segmente — eine Vorrichtung, die großzügiger ist als der
    // Betrieb.
    werte: [
      'ohne',
      'mint',
      'nearMint',
      'excellent',
      'good',
      'lightPlayed',
      'played',
      'poor',
    ],
    bereich: 'vorgaben',
  }),
  probeeinstellung({
    schluessel: 'vorgabe.einzelspiel',
    wert: 'aus',
    werte: ['aus', 'an'],
    bereich: 'vorgaben',
  }),

  // **Ohne Werteliste** — ein Verzeichnis lässt sich nicht aufzählen.
  probeeinstellung({
    schluessel: 'vorgabe.hochladepfad',
    wert: '/Users/wer/Bilder',
    werte: [],
    bereich: 'vorgaben',
  }),
]

describe('Die Vorgaben eines Profils', () => {
  it('stehen in einem eigenen Gebiet', async () => {
    const setze = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Einstellungen
          eintraege={[]}
          dienste={[]}
          fragen={[]}
          ki={[]}
          vorgaben={vorgabezeilen}
          anmeldung={[]}
          ursache={undefined}
          meldefehler={undefined}
          setze={setze}
          laeuft={() => false}
          bestandspflege={stillePflege()}
          ruecksprung={stillerRuecksprung()}
          pflege={stilleDatenpflege()}
          setID={undefined}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('tab', { name: eigene.gruppen.fach }),
    )

    const zustand = screen.getByLabelText(eigene.namen['vorgabe.zustand'])

    expect(zustand).toHaveValue('ohne')
    expect(within(zustand).getAllByRole('option')).toHaveLength(8)
  })

  // **Der Pfad ist ein Textfeld und keine leere Auswahl.**
  //
  // **Das ist die Zusage aus Punkt 8**: ein Verzeichnis lässt sich nicht
  // aufzählen, und eine leere Auswahlliste wäre ein Bedienelement, das etwas
  // verspricht, was nicht kommt.
  it('zeigen einen Pfad als Textfeld', async () => {
    const setze = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Einstellungen
          eintraege={[]}
          dienste={[]}
          fragen={[]}
          ki={[]}
          vorgaben={vorgabezeilen}
          anmeldung={[]}
          ursache={undefined}
          meldefehler={undefined}
          setze={setze}
          laeuft={() => false}
          bestandspflege={stillePflege()}
          ruecksprung={stillerRuecksprung()}
          pflege={stilleDatenpflege()}
          setID={undefined}
        />
      </Sprachraum>,
    )

    // **Der Hochladepfad steht unter *Daten und Sicherung*** — UX-04.AK-01
    // (#1676):
    // der Entwurf schreibt „Bestand, Pfade, Sicherung" in diesen Untertitel,
    // und ein Pfad ist eine Frage an die Ablage und keine an die Karte.
    await userEvent.click(
      await screen.findByRole('tab', { name: eigene.gruppen.daten }),
    )

    // **Seit UXR-19 im Abschnitt „Ordner und Dateinamen"**
    // (#1927):
    // einen Pfad legt man einmal fest, der Bestand steht offen darüber.
    await userEvent.click(
      screen.getByRole('button', { name: eigene.abschnitte.ordner }),
    )

    const pfad = screen.getByLabelText(eigene.namen['vorgabe.hochladepfad'])

    expect(pfad).toHaveValue('/Users/wer/Bilder')
    expect(pfad.tagName).toBe('INPUT')

    // **Geschrieben wird der ganze Pfad, nicht jeder Buchstabe** — OPS-02
    // (#1719): seit das Setzen die Platte fragt, wiese sie jeden halben
    // Pfad ab, und das Feld ließe sich nicht mehr füllen. Der Entwurf gilt
    // beim Verlassen des Feldes oder mit ⏎.
    await userEvent.type(pfad, '/x')

    expect(pfad).toHaveValue('/Users/wer/Bilder/x')
    expect(setze).not.toHaveBeenCalled()

    await userEvent.tab()

    expect(setze).toHaveBeenCalledTimes(1)
    expect(setze).toHaveBeenCalledWith(
      'vorgabe.hochladepfad',
      '/Users/wer/Bilder/x',
    )

    // **Mit ⏎ genauso** — und ein unveränderter Entwurf schreibt nichts.
    await userEvent.type(pfad, '/y{Enter}')
    await userEvent.tab()

    expect(setze).toHaveBeenCalledTimes(2)
    expect(setze).toHaveBeenLastCalledWith(
      'vorgabe.hochladepfad',
      '/Users/wer/Bilder/x/y',
    )
  })

  // **Und der Zustand steht wirklich im Sammlungsformular.**
  //
  // **Ohne diesen Lauf wäre die Einstellung eine Zahl im Bestand**: dass sie
  // gesetzt werden kann, sagt noch nicht, dass sie irgendwo ankommt (§10).
  it('füllen den Zustand einer neuen Sammlungszeile', async () => {
    belegeApi({
      Einstellungen: vi.fn(() =>
        Promise.resolve({
          einstellungen: [
            probeeinstellung({
              schluessel: 'vorgabe.zustand',
              wert: 'mint',
              // **Alle acht, wie Go sie schickt** (`docs/EINSTELLUNGEN.md`): mit acht
              // Werten bleibt der Zustand eine Auswahlliste. Drei wären seit UXR-F2
              // (#2008) drei Segmente — eine Vorrichtung, die großzügiger ist als der
              // Betrieb.
              werte: [
                'ohne',
                'mint',
                'nearMint',
                'excellent',
                'good',
                'lightPlayed',
                'played',
                'poor',
              ],
              bereich: 'vorgaben',
            }),
          ],
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular
          aufAufnehmen={() => Promise.resolve(true)}
          laeuft={false}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByDisplayValue(
        kataloge.de.seiten.sammlung.zustaende.mint,
      ),
    ).toBeInTheDocument()
  })

  // **Und »ohne« lässt es leer, wie es immer war.**
  //
  // **Die Gegenprobe zum Lauf darüber**: ohne sie hielte er auch, wenn der
  // Wert roh ins Feld ginge — dann stünde dort „ohne" als Kartenzustand.
  it('lassen den Zustand leer, wenn keine Vorgabe gewählt ist', async () => {
    belegeApi({
      Einstellungen: vi.fn(() =>
        Promise.resolve({
          einstellungen: [
            probeeinstellung({
              schluessel: 'vorgabe.zustand',
              wert: 'ohne',
              // **Alle acht, wie Go sie schickt** (`docs/EINSTELLUNGEN.md`): mit acht
              // Werten bleibt der Zustand eine Auswahlliste. Drei wären seit UXR-F2
              // (#2008) drei Segmente — eine Vorrichtung, die großzügiger ist als der
              // Betrieb.
              werte: [
                'ohne',
                'mint',
                'nearMint',
                'excellent',
                'good',
                'lightPlayed',
                'played',
                'poor',
              ],
              bereich: 'vorgaben',
            }),
          ],
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular
          aufAufnehmen={() => Promise.resolve(true)}
          laeuft={false}
        />
      </Sprachraum>,
    )

    const zustand = await screen.findByLabelText(
      kataloge.de.seiten.sammlung.zustand,
    )

    expect(zustand).toHaveValue('')
  })
})
