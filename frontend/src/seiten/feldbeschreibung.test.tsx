// **Die Beschreibung erreicht eine Vorlesehilfe** —
// #1278,
// Befund 5.
//
// # Warum das ein eigener Lauf ist
//
// `aria-describedby` ist eine **Liste von Kennungen, an Leerraum getrennt**.
// Aus `hilfe-rule text` werden zwei Verweise, und keiner zeigt auf einen
// Knoten. Ein Feldname mit Leerzeichen ist dabei der Normalfall:
// `archenemy.mse-game` nennt sein Textfeld `rule text`, `magic.mse-game` führt
// `casting cost` und `card number`.
//
// **Seit #1274 hängt die Beschreibung allein daran**: außerhalb des Fokus ist
// sie `sr-only`, und die Begründung dafür war, eine Vorlesehilfe lese sie
// weiterhin über `aria-describedby`. Diese Zusage ging ins Leere.
//
// # Was er prüft und was nicht
//
// Er prüft, dass die Kennung am Feld **auf einen vorhandenen Knoten zeigt** —
// über `getElementById`, also so, wie ein Browser sie auflöst.
//
// **Was er nicht prüft**: was eine echte Vorlesehilfe daraus macht. Das sieht
// nur, wer eine benutzt.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { Feldzeile } from './Feldzeile'

function form(name: string, beschreibung: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
    beschreibung,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    vorgabe: '',
    editierbar: true,
    gespeichert: true,
    teilfelder: [],
    reihe: '',
    stelle: 0,
  }
}

function zeige(feld: string, fehler?: string) {
  render(
    <Sprachraum sprache="de">
      <Feldzeile
        feld={feld}
        wert=""
        form={form(feld, 'Was in dieses Feld gehört.')}
        fehler={fehler}
        laeuft={false}
        palette={undefined}
        wortlisten={[]}
        hergeleitet={undefined}
        gewaehlt={() => undefined}
        getippt={() => undefined}
        schreiben={() => undefined}
        einfuegen={() => undefined}
        bildWaehlen={() => undefined}
        wortWaehlen={() => undefined}
      />
    </Sprachraum>,
  )
}

// **Die echten Feldnamen aus dem Pack**, nicht ausgedachte.
const ausDemPack = ['name', 'rule text', 'casting cost', 'card number']

describe('die Beschreibung eines Feldes', () => {
  it.each(ausDemPack)('ist bei %p wirklich verknüpft', (feld) => {
    zeige(feld)

    const eingabe = screen.getByRole('textbox', { name: feld })
    const kennung = eingabe.getAttribute('aria-describedby')

    expect(kennung).not.toBeNull()

    // **Kein Leerraum.** Ein Browser zerlegt die Angabe daran, und aus einem
    // Verweis würden zwei, die ins Leere zeigen.
    expect(kennung).not.toMatch(/\s/)

    expect(document.getElementById(kennung ?? '')).not.toBeNull()
    expect(document.getElementById(kennung ?? '')).toHaveTextContent(
      'Was in dieses Feld gehört.',
    )
  })

  // **Und der Fehler ebenso** — er geht vor, und er hängt an derselben Stelle.
  it.each(ausDemPack)('verknüpft auch den Fehler bei %p', (feld) => {
    zeige(feld, 'Das geht so nicht.')

    const eingabe = screen.getByRole('textbox', { name: feld })
    const kennung = eingabe.getAttribute('aria-describedby')

    expect(kennung).not.toMatch(/\s/)
    expect(document.getElementById(kennung ?? '')).toHaveTextContent(
      'Das geht so nicht.',
    )
  })
})
