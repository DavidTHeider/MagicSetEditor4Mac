// **Das Formular liest sich wie die Karte**
// (#674).
import { describe, expect, it } from 'vitest'

import type { FeldbeschreibungDTO, WertDTO } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { angeboteneFelder } from './angebotenefelder'

function flaeche(feld: string, oben: number, links = 0): FeldflaecheDTO {
  return { feld, links, oben, breite: 10, hoehe: 5 }
}

function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    aufschrift: name,
  }
}

function wert(feld: string, text = ''): WertDTO {
  return { feld, wert: text }
}

describe('Die Ordnung der angebotenen Felder', () => {
  // **Von oben nach unten, wie sie auf der Karte liegen.**
  //
  // `magic.mse-game` erklärt `border color`, `corners` und `card color` vor
  // `name` — das ist MSEs Reihenfolge, und sie stimmt: MSE baut die
  // Werteliste aus `game.card_fields`. Nur zeigt MSE gar kein Formular,
  // sondern die Felder **auf** der Karte. Die senkrechte Liste ist eine
  // Anpassung dieses Projekts; dann gehört auch ihre Ordnung hierher.
  it('ordnet nach der Lage auf der Karte', () => {
    const felder = angeboteneFelder(
      [],
      // **Die Erklärung des Spiels steht bewusst andersherum** — sonst
      // stimmte die Reihenfolge zufällig, und der Fall wäre nicht enthalten
      // (§10).
      [form('pt'), form('name'), form('border color')],
      [flaeche('pt', 480), flaeche('name', 30), flaeche('border color', 0)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['border color', 'name', 'pt'])
  })

  // **Bei gleicher Höhe entscheidet die Spalte** — links vor rechts, wie
  // jemand liest.
  it('ordnet gleich hohe Felder von links nach rechts', () => {
    const felder = angeboteneFelder(
      [],
      [form('rechts'), form('links')],
      [flaeche('rechts', 30, 200), flaeche('links', 30, 10)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['links', 'rechts'])
  })

  // **Was der Stil nicht zeichnet, steht hinten** — und untereinander in der
  // Reihenfolge des Spiels. Ein Set aus MTG JSON trägt Felder, für die die
  // gewählte Vorlage keine Ebene hat; sie zu verstecken hieße, sie
  // stillschweigend zu verlieren.
  it('hängt ungezeichnete Felder hinten an', () => {
    const felder = angeboteneFelder(
      [wert('notiz', 'etwas')],
      [form('name')],
      [flaeche('name', 30)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['name', 'notiz'])
  })

  // **Der Wert der Karte bleibt der Wert der Karte** — die Ordnung ändert
  // nichts daran, was in einem Feld steht.
  it('behält die Werte', () => {
    const felder = angeboteneFelder(
      [wert('name', 'Feuerball')],
      [form('name')],
      [flaeche('name', 30)],
    )

    expect(felder).toEqual([{ feld: 'name', wert: 'Feuerball' }])
  })

  // **Ein Feld, das die ganze Karte deckt, ist ihr Grund** — und steht am
  // Ende, nicht am Anfang.
  //
  // Bei `magic-m15` sind das `card color`, `border color`, `foil layer` und
  // `corners`: je 375 × 523 von einer Hülle von 375 × 536. Nach `oben`
  // sortiert stünden sie ganz vorn, denn sie beginnen bei null — und der Name
  // der Karte käme dahinter.
  it('stellt den Grund der Karte hinter die gesetzten Felder', () => {
    const felder = angeboteneFelder(
      [],
      [form('card color'), form('name'), form('text')],
      [
        { feld: 'card color', links: 0, oben: 0, breite: 375, hoehe: 523 },
        { feld: 'name', links: 32, oben: 30, breite: 309, hoehe: 23 },
        { feld: 'text', links: 29, oben: 327, breite: 314, hoehe: 154 },
      ],
    )

    expect(felder.map((f) => f.feld)).toEqual(['name', 'text', 'card color'])
  })

  // **Eine Zeile ist eine Zeile, auch wenn ein Feld einen Punkt höher sitzt.**
  //
  // Bei `magic-m15` beginnt `card symbol` bei 29, `casting cost` bei 29 und
  // `name` bei **30**. Nach `oben` allein stünde der Name hinter den Kosten —
  // auf der Karte stehen alle drei nebeneinander in der Titelzeile.
  it('fasst überlappende Felder zu einer Zeile zusammen', () => {
    const felder = angeboteneFelder(
      [],
      [form('casting cost'), form('name'), form('card symbol')],
      [
        { feld: 'casting cost', links: 311, oben: 29, breite: 35, hoehe: 23 },
        { feld: 'name', links: 32, oben: 30, breite: 309, hoehe: 23 },
        { feld: 'card symbol', links: 8, oben: 29, breite: 9, hoehe: 21 },
      ],
    )

    expect(felder.map((f) => f.feld)).toEqual([
      'card symbol',
      'name',
      'casting cost',
    ])
  })
})
