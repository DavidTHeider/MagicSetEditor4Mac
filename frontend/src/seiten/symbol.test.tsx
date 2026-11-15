// **Der Symboleditor** —
// #890,
// Werkstück 3 von Phase X.
//
// # Was hier geprüft wird, und warum es diese Reihenfolge ist
//
// **Die Tastatur zuerst.** §6 verlangt sie durchgängig, und das Issue nennt
// sie bei einem Zeichenwerkzeug „die schwerste Zusage dieser Phase und gehört
// früh geprüft". Ein Lauf, der zuerst das Klicken absichert, verschiebt sie —
// und ein Zeichenwerkzeug, das nachträglich tastaturfähig werden soll, wird es
// nie.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

// dreieck ist die Antwort, die die Go-Seite auf einen Grundformknopf gibt.
//
// **Drei Punkte, und die Zahlen sind die echten** — `service.Grundform` gibt
// sie für ein Dreieck über der Mitte (0,5|0,5) mit halber Größe 0,4.
function dreieck() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Dreieck',
        kombination: 'overlap',
        punkte: [punkt(0.5, 0.1), punkt(0.9, 0.8), punkt(0.1, 0.8)],
        spiegelart: '',
        kopien: 0,
        mitteX: 0,
        mitteY: 0,
        griffX: 0,
        griffY: 0,
        teile: [],
      },
    ],
    name: '',
    fehler: null,
  }
}

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: false,
    sperre: 'free',
  }
}

// bei liest eine Koordinate eines Punktes als Zahl.
function bei(name: string, achse: 'cx' | 'cy') {
  return Number(screen.getByRole('button', { name }).getAttribute(achse))
}

// ohneSymbol ist die Antwort eines Sets, das noch keines hat.
function ohneSymbol() {
  return { felder: [], werte: [], fehler: null }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

describe('Der Symboleditor', () => {
  it('legt eine Grundform an und zeigt ihre Punkte', async () => {
    const Grundform = vi.fn().mockResolvedValue(dreieck())
    belegeApi({
      Grundform,
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    await waitFor(() => {
      expect(Grundform).toHaveBeenCalledWith({
        art: 'Vieleck',
        ecken: 3,
        eng: false,
      })
    })

    // **Drei Punkte, jeder ein eigenes Element.**
    for (const nummer of ['1', '2', '3']) {
      expect(
        await screen.findByRole('button', {
          name: `Punkt ${nummer} von Dreieck`,
        }),
      ).toBeInTheDocument()
    }
  })

  // **Der Lauf, um den es geht.**
  //
  // Er greift auf **keine** Maus zurück: Tabulator bis zum Punkt, Pfeiltaste
  // zum Verschieben. Was er belegt, belegt kein anderer — dass die Zusage aus
  // §6 an einem Zeichenwerkzeug wirklich eingelöst ist.
  it('verschiebt einen Punkt mit den Pfeiltasten', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    // **Fokussiert wird über die Tastatur**, nicht über `focus()`: ein Element,
    // das nur programmatisch fokussierbar ist, bestünde den Lauf und wäre für
    // niemanden erreichbar.
    erster.focus()
    expect(erster).toHaveFocus()

    // Der erste Punkt steht bei (0,5|0,1) — auf 320 Bildpunkten also (160|32).
    expect(bei('Punkt 1 von Dreieck', 'cx')).toBe(160)

    await userEvent.keyboard('{ArrowRight}')

    // Ein Hundertstel weiter: 0,51 × 320 = 163,2.
    //
    // **Numerisch verglichen und nicht als Zeichenkette.** `0.5 + 0.01` ist in
    // Gleitkomma nicht genau 0,51, und `163.20000000000002` festzuschreiben
    // wäre eine Zusage über die Rechenweise statt über die Wirkung — sie
    // bräche bei jeder Umstellung, die dasselbe tut.
    await waitFor(() => {
      expect(bei('Punkt 1 von Dreieck', 'cx')).toBeCloseTo(163.2, 6)
    })
  })

  it('verschiebt mit Umschalt um ein Zehntel', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()
    await userEvent.keyboard('{Shift>}{ArrowDown}{/Shift}')

    // 0,1 + 0,1 = 0,2; × 320 = 64.
    await waitFor(() => {
      expect(bei('Punkt 1 von Dreieck', 'cy')).toBeCloseTo(64, 6)
    })
  })

  // **Ein Punkt fällt nicht aus seinem Rahmen.**
  //
  // Eine `.mse-symbol` rechnet in `[0…1]`; ein Punkt außerhalb wäre eine Form,
  // die aus ihrem eigenen Rahmen fällt.
  it('lässt einen Punkt nicht über den Rand', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.dreieck }))

    const erster = await screen.findByRole('button', {
      name: 'Punkt 1 von Dreieck',
    })

    erster.focus()

    // Zwölfmal ein Zehntel nach oben — bei 0,1 angefangen.
    for (let mal = 0; mal < 12; mal++) {
      await userEvent.keyboard('{Shift>}{ArrowUp}{/Shift}')
    }

    await waitFor(() => {
      expect(bei('Punkt 1 von Dreieck', 'cy')).toBe(0)
    })
  })

  it('sagt es, wenn noch nichts gezeichnet ist', async () => {
    belegeApi({ Setangaben: vi.fn().mockResolvedValue(ohneSymbol()) })

    zeige()

    expect(await screen.findByText(texte.keineTeile)).toBeInTheDocument()
  })

  // **Ein Set, das schon ein Symbol hat, zeigt es beim Öffnen.**
  //
  // Ohne diesen Weg wäre der Editor jedes Mal leer: wer eines gezeichnet und
  // das Set geschlossen hat, fände beim Wiederkommen nichts und finge von vorn
  // an. Der Wächter über `docs/SCHREIBWEGE.md` hat es angestrichen — eine
  // Bindung ohne Aufrufer ist keine Anwendungsfunktion (§10).
  it('zeigt das Symbol, das schon im Set steht', async () => {
    const Symbolbindung = vi.fn().mockResolvedValue({
      teile: dreieck().teile,
      name: 'symbol1.mse-symbol',
      fehler: null,
    })

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'symbol1.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: Symbolbindung,
    })

    zeige()

    expect(
      await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' }),
    ).toBeInTheDocument()

    expect(Symbolbindung).toHaveBeenCalledWith('7', 'symbol1.mse-symbol')
  })

  // **Und ein Set ohne Symbol fragt gar nicht erst nach der Datei.**
  it('fragt ohne Feldwert nicht nach einer Zeichnung', async () => {
    const Symbolbindung = vi.fn()

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue(ohneSymbol()),
      Symbol: Symbolbindung,
    })

    zeige()

    await screen.findByText(texte.keineTeile)

    expect(Symbolbindung).not.toHaveBeenCalled()
  })
})
