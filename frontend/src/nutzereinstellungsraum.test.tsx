// **Die Übersteuerungen liegen am Wurzelelement** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkt 11.
//
// **Am `documentElement` und nicht an einem Knoten im Baum**: beides gilt für
// alles, auch für den Rollbalken, den die Anwendung gar nicht zeichnet.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  Nutzereinstellungsraum,
  useNutzereinstellungsstand,
} from './nutzereinstellungsraum'
import { Sprachraum } from './sprache/kontext'
import { belegeApi, raeumeApi } from './test/api'

import type { NutzereinstellungDTO } from './api'
import { probeprofil } from './test/profil'

afterEach(() => {
  raeumeApi()
  document.documentElement.removeAttribute('data-bewegung')
  document.documentElement.removeAttribute('data-kontrast')
})

// Anzeige liest den Stand aus dem Raum — **jedes Stück davon**.
//
// **Alle vier und nicht nur die Kennung**: der Rückfall außerhalb des Raums
// besteht aus vier Angaben, und drei davon sind Funktionen. Eine Anzeige, die
// nur `nutzerID` läse, ließe sie ungeprüft — und ein Rückfall, den niemand
// ruft, ist eine Behauptung darüber, was geschieht (§3).
function Anzeige() {
  const stand = useNutzereinstellungsstand()

  // **Jedes Stück in einem eigenen Knoten.** `findByText` sucht **ein**
  // Element; nebeneinanderliegende Textknoten in demselben findet es nicht.
  return (
    <output>
      <span>{`kennung:${stand.nutzerID}`}</span>
      <span>{`bewegung:${stand.wert('barrierefreiheit.bewegung')}`}</span>
      <span>{`laeuft:${String(stand.laeuft('barrierefreiheit.bewegung'))}`}</span>
      <button
        type="button"
        onClick={() => {
          stand.setze('barrierefreiheit.bewegung', 'reduziert')
        }}
      >
        setzen
      </button>
    </output>
  )
}

function zeige(eintraege: NutzereinstellungDTO[]) {
  belegeApi({
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({
        profil: probeprofil({
          id: '10',
          nutzerID: '7',
          name: 'Mein Profil',
          fassung: '1',
        }),
      }),
    ),
    Nutzereinstellungen: vi.fn(() =>
      Promise.resolve({ einstellungen: eintraege }),
    ),
  })

  render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Anzeige />
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )
}

// wahl baut einen Eintrag.
const wahl = (schluessel: string, wert: string): NutzereinstellungDTO => ({
  schluessel,
  wert,
  werte: [wert],
})

describe('Der Nutzereinstellungsraum', () => {
  it('setzt die Übersteuerungen ans Wurzelelement', async () => {
    zeige([
      wahl('barrierefreiheit.bewegung', 'reduziert'),
      wahl('barrierefreiheit.kontrast', 'hoch'),
    ])

    // **Gewartet wird auf die Wirkung** (§3): der Effekt läuft nach dem
    // Zeichnen, und das Zeichnen braucht die Antwort.
    expect(await screen.findByText('bewegung:reduziert')).toBeInTheDocument()

    // **Und auf das Attribut eigens** — der Lauf daneben tat es längst, dieser
    // nicht.
    //
    // **`findBy…` sagt zu, dass ein Knoten im Baum steht, nicht dass ein
    // Effekt gelaufen ist** (§3, „und der Beleg selbst geht ins Leere, wenn er
    // zu früh kommt"): Testing Library gibt dafür **eine** Makroaufgabe Frist,
    // React legt seine passiven Effekte über den Scheduler in eine andere.
    // Unter Last reicht die Frist nicht.
    //
    // **Gemessen am 2026-09-06**: in einem vollen Pipelinelauf war
    // `data-bewegung` gleich `null`, in drei Läufen derselben Datei einzeln
    // und zwei vollen Läufen danach nicht. Ein Flattern, das nur unter Last
    // erscheint, ist genau dieses.
    await waitFor(() => {
      expect(document.documentElement).toHaveAttribute(
        'data-bewegung',
        'reduziert',
      )
    })

    expect(document.documentElement).toHaveAttribute('data-kontrast', 'hoch')
  })

  // **`system` entfernt das Attribut, statt es zu setzen.**
  //
  // **Der Unterschied ist die ganze Zusage**: der Selektor in `stil.css`
  // fragt auf den Wert, und ein Attribut mit „system" wäre eine Angabe, die
  // niemand liest — aber eine, die man beim Lesen für wirksam hält.
  it('räumt sie wieder weg, wenn das System entscheidet', async () => {
    document.documentElement.setAttribute('data-bewegung', 'reduziert')
    document.documentElement.setAttribute('data-kontrast', 'hoch')

    zeige([
      wahl('barrierefreiheit.bewegung', 'system'),
      wahl('barrierefreiheit.kontrast', 'system'),
    ])

    expect(await screen.findByText('bewegung:system')).toBeInTheDocument()

    // **Auch hier auf die Wirkung** — und diese Verneinung kann nicht zu früh
    // zutreffen: die beiden Attribute standen **vor** dem Zeichnen da, und
    // nur der Effekt nimmt sie weg.
    await waitFor(() => {
      expect(document.documentElement).not.toHaveAttribute('data-bewegung')
    })

    expect(document.documentElement).not.toHaveAttribute('data-kontrast')
  })

  // **Außerhalb des Raums gibt es einen leeren Stand und keinen Absturz.**
  //
  // Ein Bauteil, das ihn dort ruft, ist ein Fehler — aber einer, der eine
  // leere Auswahlliste zeigt und nicht das Fenster weiß macht.
  it('gibt außerhalb einen leeren Stand', async () => {
    render(
      <Sprachraum sprache="de">
        <Anzeige />
      </Sprachraum>,
    )

    expect(screen.getByText('kennung:')).toBeInTheDocument()
    expect(screen.getByText('bewegung:')).toBeInTheDocument()
    expect(screen.getByText('laeuft:false')).toBeInTheDocument()

    // **Und der Rückfall tut nichts, statt zu stürzen.**
    await userEvent.click(screen.getByRole('button', { name: 'setzen' }))

    expect(screen.getByText('bewegung:')).toBeInTheDocument()
  })
})
