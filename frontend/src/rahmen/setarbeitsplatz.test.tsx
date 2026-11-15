// Der Set-Arbeitsplatz.
//
// **Fünf der neun Seiten gehören zu einem Set**, vier zur Anwendung — und bis
// #288 standen
// alle neun in derselben Spalte. „Statistik" neben „Einstellungen" sagt nicht,
// dass die eine über das offene Set spricht und die andere über das Programm.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Seitenschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { MitOffenemSet } from '../test/offenesset'
import { Setarbeitsplatz } from './Setarbeitsplatz'

const texte = kataloge.de

const sets = {
  sets: [
    {
      id: '7',
      name: 'Vier Winde',
      spiel: 'magic',
      stil: 'magic-m15',
      karten: 312,
      herkunft: '/a.mse-set',
      herkunftFehlt: false,
    },
  ],
  fehler: undefined,
}

function zeige(
  setID: string | undefined,
  aktuell: Seitenschluessel = 'karten',
  wechseln: (seite: Seitenschluessel) => void = vi.fn(),
) {
  return render(
    <Sprachraum sprache="de">
      <MitOffenemSet setID={setID}>
        <Setarbeitsplatz
          setID={setID}
          aktuell={aktuell}
          wechseln={wechseln}
          verlaufsstand={0}
          oeffneKarte={vi.fn()}
        >
          {() => <p>Inhalt</p>}
        </Setarbeitsplatz>
      </MitOffenemSet>
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

describe('Setarbeitsplatz', () => {
  it('nennt das offene Set über den Laschen', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    zeige('7')

    expect(await screen.findByText('Vier Winde')).toBeInTheDocument()
  })

  // **Die Laschen sind eine Laschenleiste**, keine Knopfreihe: ein
  // Vorleseprogramm sagt damit „Lasche 2 von 5", und die Pfeiltasten wandern.
  it('führt die Setseiten als Laschen', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    zeige('7')

    const leiste = await screen.findByRole('tablist')
    expect(leiste).toBeInTheDocument()

    // **Drei seit dem Ausgabeassistenten** (UXR-09, #1917) — *Bilder*,
    // *Ausgeben* und *Druckbogen* öffnet der Knopf *Ausgeben* im Kopf. Sechs
    // waren es seit den Schubladen (UX-08, #1680), dreizehn seit dem
    // Ersetzen (#1168), zwölf seit dem Symboleditor (#890), elf seit dem
    // Zufallspack (#1072), zehn seit der Konsole (#1070), neun seit den
    // Schlüsselwörtern (#1062). Die Zahl steht hier und nicht als
    // `setlaschen.length`: eine Erwartung, die sich mit der Quelle
    // mitbewegt, prüft nichts.
    const laschen = screen.getAllByRole('tab')
    expect(laschen).toHaveLength(3)
    expect(
      screen.getByRole('tab', { name: texte.seiten.karten.titel }),
    ).toHaveAttribute('aria-selected', 'true')
  })

  it('wechselt über eine Lasche', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    render(
      <Sprachraum sprache="de">
        <MitOffenemSet setID="7">
          <Setarbeitsplatz
            setID="7"
            aktuell="karten"
            wechseln={gewechselt}
            verlaufsstand={0}
            oeffneKarte={vi.fn()}
          >
            {() => <p>Inhalt</p>}
          </Setarbeitsplatz>
        </MitOffenemSet>
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('tab', { name: texte.seiten.suche.titel }),
    )

    expect(gewechselt).toHaveBeenCalledWith('suche')
  })

  // **Jede Schublade geht bei ihrer ersten Seite auf** (UX-08, #1680) — der
  // Knopf ist ein Seitenwechsel wie eine Lasche, und die Seite entscheidet,
  // welche Schublade steht.
  it('öffnet die Schubladen bei ihrer ersten Seite', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    zeige('7', 'karten', gewechselt)

    const worte = texte.allgemein.arbeitsplatz.schubladen

    await userEvent.click(
      await screen.findByRole('button', { name: worte.eigenschaften.knopf }),
    )
    expect(gewechselt).toHaveBeenLastCalledWith('stil')

    await userEvent.click(
      screen.getByRole('button', { name: worte.werkzeuge.knopf }),
    )
    expect(gewechselt).toHaveBeenLastCalledWith('statistik')
  })

  // **Unter einer Schublade liegt die Kartenfläche** — und die Leiste sagt
  // es: *Karten* ist gewählt, das Feld nennt sie, und die Pfeile zählen von
  // dort. Ohne diesen Rückfall stünde auf *Stil* keine Lasche gewählt, und
  // die Leiste hätte keinen Eintrag im Tabweg.
  it('wählt unter einer Schublade die Karten', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    zeige('7', 'stil', gewechselt)

    const karten = await screen.findByRole('tab', {
      name: texte.seiten.karten.titel,
    })
    expect(karten).toHaveAttribute('aria-selected', 'true')
    expect(karten).toHaveAttribute('tabindex', '0')
    expect(screen.getByRole('tabpanel')).toHaveAttribute(
      'aria-labelledby',
      karten.id,
    )

    karten.focus()
    await userEvent.keyboard('{ArrowRight}')

    // **Neben den Karten steht seit UXR-09 die Suche** — die Bilder sind
    // ein Ziel des Ausgabeassistenten (#1917).
    expect(gewechselt).toHaveBeenLastCalledWith('suche')
  })

  // **Ohne offenes Set gibt es keinen Arbeitsplatz** — und der leere Zustand
  // sagt, was zu tun ist (§6), statt eine Laschenleiste über nichts zu zeigen.
  //
  // **Und der Schritt ist ein Knopf** — UX-15 (#1771). In der Sichtprüfung
  // vom 2026-09-24 stand hier „Öffne unter „Sets“ eine .mse-set-Datei", über
  // einem Bestand mit Sets: wer eines hatte, las, er müsse eine Datei holen.
  it('erklärt ohne Set den nächsten Schritt und führt zu den Sets', async () => {
    const wechseln = vi.fn()
    zeige(undefined, 'karten', wechseln)

    expect(
      screen.getByText(texte.allgemein.arbeitsplatz.keinSet),
    ).toBeInTheDocument()
    expect(screen.queryByRole('tablist')).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.allgemein.arbeitsplatz.zuDenSets,
      }),
    )
    expect(wechseln).toHaveBeenCalledWith('sets')
  })

  // **Ein Ladefehler nimmt den Arbeitsplatz nicht weg.** Der Name ist eine
  // Zugabe; die Karten sind der Grund, warum jemand hier ist.
  it('zeigt die Laschen auch ohne Namen', async () => {
    belegeApi({ Sets: vi.fn().mockRejectedValue(new Error('weg')) })

    zeige('7')

    await waitFor(() => {
      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })
  })

  it('zeigt den Inhalt der gewählten Lasche', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    zeige('7')

    expect(await screen.findByRole('tabpanel')).toHaveTextContent('Inhalt')
  })

  // **Lasche und Feld zeigen aufeinander** — #1553.
  //
  // Seit die Kennungen aus `useId` kommen, steht der Bezug nicht mehr als
  // lesbares Wort im Quelltext: `aria-controls="laschenfeld"` war falsch zu
  // schreiben schwer, `aria-controls={feldkennung}` ist es leicht. Geprüft
  // wird deshalb der Bezug selbst und nicht sein Name.
  it('verweist von der gewählten Lasche auf das Feld und zurück', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    zeige('7')

    const feld = await screen.findByRole('tabpanel')
    const gewaehlt = screen.getByRole('tab', { selected: true })

    expect(gewaehlt).toHaveAttribute('aria-controls', feld.id)
    expect(feld).toHaveAttribute('aria-labelledby', gewaehlt.id)
  })

  // **Die Pfeiltasten wandern zwischen den Laschen** — das Muster, das ein
  // Vorleseprogramm bei `role="tablist"` erwartet, und §6 verlangt jede
  // Handlung auch über die Tastatur.
  it('wandert mit den Pfeiltasten', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    render(
      <Sprachraum sprache="de">
        <MitOffenemSet setID="7">
          <Setarbeitsplatz
            setID="7"
            aktuell="karten"
            wechseln={gewechselt}
            verlaufsstand={0}
            oeffneKarte={vi.fn()}
          >
            {() => <p>Inhalt</p>}
          </Setarbeitsplatz>
        </MitOffenemSet>
      </Sprachraum>,
    )

    const lasche = await screen.findByRole('tab', {
      name: texte.seiten.karten.titel,
    })
    lasche.focus()

    // **Der Nachbar ist der in der Leiste, nicht der in der Kürzelreihe.**
    // Neben *Karten* steht seit UXR-09 die *Suche* (#1917); bis dahin waren
    // es die *Bilder* (#1166), davor *Setangaben*, weil die Reihe die
    // Kürzelreihe war.
    await userEvent.keyboard('{ArrowRight}')
    expect(gewechselt).toHaveBeenLastCalledWith('suche')

    // **Gezählt wird von der Lasche, auf der man steht** (UXR-09, #1917) —
    // der Fokus ist mit dem Pfeil zur Suche gewandert, und zurück geht es
    // von dort. Hier stand bis dahin, dass der zweite Pfeil von der
    // **gewählten** zählt; mit einem Wechsel, den niemand ausführt, ist das
    // eine Lasche, auf der niemand mehr steht.
    await userEvent.keyboard('{ArrowLeft}')
    expect(gewechselt).toHaveBeenLastCalledWith('karten')

    // **Zyklisch, also von der ersten zur letzten** — und die letzte ist seit
    // UXR-09 *Ersetzen*. Von UX-08 bis dahin war es der Druckbogen am Ende
    // der Gruppe „Ausgabe" (#1680), davor die Konsole (#1166).
    await userEvent.keyboard('{ArrowLeft}')
    expect(gewechselt).toHaveBeenLastCalledWith('ersetzen')
  })

  // **Eine andere Taste tut nichts.** Ohne diesen Fall stünde nicht fest, dass
  // die Leiste nur die Pfeile abfängt — ⌘K muss sie durchlassen.
  it('lässt andere Tasten durch', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    render(
      <Sprachraum sprache="de">
        <MitOffenemSet setID="7">
          <Setarbeitsplatz
            setID="7"
            aktuell="karten"
            wechseln={gewechselt}
            verlaufsstand={0}
            oeffneKarte={vi.fn()}
          >
            {() => <p>Inhalt</p>}
          </Setarbeitsplatz>
        </MitOffenemSet>
      </Sprachraum>,
    )

    ;(
      await screen.findByRole('tab', { name: texte.seiten.karten.titel })
    ).focus()

    await userEvent.keyboard('{ArrowDown}')

    expect(gewechselt).not.toHaveBeenCalled()
  })

  // **Der Ausgabeassistent ist keine Lasche** (UXR-09, #1917) — er steht
  // im Arbeitsplatz, aber keine Lasche ist gewählt, und das Feld darunter
  // nennt sich nicht nach einer. **Die Leiste bleibt erreichbar**: ohne
  // gewählte Lasche steht die erste im Tabweg, und die Pfeile wandern von
  // der Lasche aus, auf der man steht.
  //
  // Hier stand bis dahin „tut nichts, wenn die aktuelle Seite keine Lasche
  // ist" — mit einer Anwendungsseite, die der Arbeitsplatz nie umgibt. Seit
  // es eine Setseite ohne Lasche gibt, wäre „nichts" eine Leiste, aus der
  // die Tastatur nicht mehr herausführt.
  it('wählt auf dem Ausgabeassistenten keine Lasche und hält die erste im Tabweg', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    zeige('7', 'druckbogen', gewechselt)

    const karten = await screen.findByRole('tab', {
      name: texte.seiten.karten.titel,
    })
    expect(screen.queryByRole('tab', { selected: true })).toBeNull()
    expect(karten).toHaveAttribute('tabindex', '0')
    expect(screen.queryByRole('tabpanel')).toBeNull()
    expect(screen.getByText('Inhalt')).toBeInTheDocument()

    karten.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(gewechselt).toHaveBeenLastCalledWith('suche')
  })

  // **Ausgeben steht im Kopf** — wie im Artboard *Karten*: neben
  // *Eigenschaften* und *Werkzeuge* (UXR-09, #1917). Er schlägt den
  // Assistenten auf, ohne ein Ziel vorzuwählen. **Im Nebengewicht** seit AP-3
  // (#2360): der Kopf steht über jeder Seite des Sets, und die Seite hat ihre
  // eigene führende Handlung — „+ Karte anlegen“, „Nachschlagen“; je Blick
  // führt einer (`Zustaende09`). Das Ausgeben selbst führt auf seiner Seite
  // (`Wege09`).
  it('trägt Ausgeben im Kopf im Nebengewicht', async () => {
    belegeApi({ Sets: vi.fn().mockResolvedValue(sets) })

    const gewechselt = vi.fn()
    zeige('7', 'karten', gewechselt)

    const knopf = await screen.findByRole('button', {
      name: texte.seiten.ausgeben.titel,
    })
    expect(knopf).not.toHaveClass('bg-marke')
    expect(knopf).toHaveClass('border-kante')

    await userEvent.click(knopf)
    expect(gewechselt).toHaveBeenLastCalledWith('ausgeben')
  })
})
