// **Die Belegung sagt, woher sie kommt, und hält beim Schreiben an** —
// Befunde M15 und G13 des Zweitblicks über Phase 6.3
// (#1665):
// `kuerzel.ziffern` zeichnete weder Ebene noch Herkunft noch Wirkzeitpunkt,
// und während des Schreibens ließ sie einen zweiten Wunsch zu.
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Kuerzeltafel } from './Kuerzeltafel'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegungAus } from '../rahmen/ziffernbelegung'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

const eintrag: EinstellungDTO = {
  schluessel: 'kuerzel.ziffern',
  wert: '3=karten',
  wirksam: '3=karten',
  werte: [],
  bereich: 'anwendung',
  ebene: 'nutzer',
  herkunft: 'nutzer',
  wirkung: 'sofort',
  vorgabe: '',
}

function zeige(angehalten: boolean) {
  render(
    <Sprachraum sprache="de">
      <Kuerzeltafel
        belegung={belegungAus(eintrag.wert)}
        setze={vi.fn()}
        eintrag={eintrag}
        angehalten={angehalten}
      />
    </Sprachraum>,
  )
}

describe('die Herkunft der Belegung', () => {
  it('nennt die Herkunft als Marke wie jede Einstellung', () => {
    zeige(false)

    // **Ohne den rohen Wert** — die Tafel zeigt ihn als Auswahl, und
    // „3=karten" in einem Satz liest sich wie ein Fehler. Die Marke (AP-11,
    // #2368) trägt den ganzen Satz für das Vorleseprogramm.
    expect(screen.getByText('von dir')).toBeInTheDocument()
    expect(
      screen.getByText('Gesetzt von dir, Ebene Nutzer.'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Es gilt/)).toBeNull()
    expect(screen.queryByText(/Wirkt sofort/)).toBeNull()
    expect(screen.getByRole('combobox', { name: 'Seite für ⌘3' })).toBeEnabled()
  })

  // **Ein Vorgang je Datensatz** (§7, Regel 4): solange geschrieben wird,
  // nimmt die Tafel keinen zweiten Wunsch an.
  it('hält Auswahl und Knöpfe an, solange geschrieben wird', () => {
    zeige(true)

    for (const feld of screen.getAllByRole('combobox')) {
      expect(feld).toBeDisabled()
    }
    expect(
      screen.getByRole('button', { name: eigene.belegungVor06 }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: eigene.belegungVorgabe }),
    ).toBeDisabled()
  })
})
