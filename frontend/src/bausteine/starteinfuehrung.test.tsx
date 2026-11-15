// **Wann die Einführung steht und was sie mitbekommt.**
//
// **Eine eigene Datei**, weil `einfuehrung.test.tsx` das **Fenster** prüft und
// hier steht, **wer es aufmacht** (§2).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Starteinfuehrung } from './Starteinfuehrung'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Einstellungsstand } from '../haken/useEinstellungen'

const texte = kataloge.de.seiten.start

afterEach(() => {
  raeumeApi()
})

// einStand baut einen Einstellungsstand mit dem gegebenen Wert.
function einStand(wert: string): Einstellungsstand {
  return {
    geladen: true,
    wert: (schluessel: string) => (schluessel === 'ersteschritte' ? wert : ''),
  } as unknown as Einstellungsstand
}

function zeige(wert: string, erneut = false) {
  belegeApi({})

  render(
    <Sprachraum sprache="de">
      <Starteinfuehrung
        einstellungen={einStand(wert)}
        zugeklappt={false}
        erneut={erneut}
        setID={undefined}
        wechseln={vi.fn()}
        schliessen={vi.fn()}
        gewaehlt={vi.fn()}
      />
    </Sprachraum>,
  )
}

describe('Die Einführung beim Start', () => {
  it('steht, solange sie nicht abgeschaltet ist', () => {
    zeige('an')

    expect(
      screen.getByRole('checkbox', { name: texte.einfuehrung.wiederzeigen }),
    ).toBeChecked()
  })

  it('bleibt fort, wenn sie abgeschaltet ist', () => {
    zeige('aus')

    expect(
      screen.queryByRole('checkbox', { name: texte.einfuehrung.wiederzeigen }),
    ).not.toBeInTheDocument()
  })

  // **Und ohne gespeicherten Wert gilt die Vorgabe der Domäne.**
  //
  // # Warum `wiederZeigen` hier fest steht und das richtig ist
  //
  // **Weil die Einführung gar nicht erst gezeichnet wird, wenn `aus`
  // gespeichert ist.** Der einzige Wert, mit dem sie überhaupt erscheint, ist
  // „an" — ein fester `true` ist damit kein Sonderfall, sondern der einzige
  // Fall. Wer ihn aus der Einstellung läse, schriebe dieselbe Zahl noch einmal
  // hin.
  //
  // **Das gehört danebengeschrieben**, weil es beim Lesen wie eine
  // Nachlässigkeit aussieht: am 2026-09-06 wurde es beim Suchen nach „der
  // Haken geht nicht weg" zunächst dafür gehalten. Der Fehler lag anderswo —
  // das Feld war **kontrolliert** über eine Eigenschaft, die sich nie ändert
  // (siehe `einfuehrung.test.tsx`).
  it('steht auch ohne gespeicherten Wert — mit Haken', () => {
    // **`leer` heißt: nichts gespeichert.** Die Vorgabe der Domäne ist „an",
    // und die Einführung steht — mit Haken.
    zeige('')

    expect(
      screen.getByRole('checkbox', { name: texte.einfuehrung.wiederzeigen }),
    ).toBeChecked()
  })

  // **Erneut verlangt zählt mehr als die Einstellung** — Werkstück U11 von
  // Phase 5.1 (#1422): wer die ersten Schritte aus den Führungen zurückholt,
  // bekommt sie auch bei „aus".
  it('steht erneut, obwohl die Einstellung aus sagt', () => {
    zeige('aus', true)

    expect(
      screen.getByRole('checkbox', { name: texte.einfuehrung.wiederzeigen }),
    ).toBeInTheDocument()
  })
})
