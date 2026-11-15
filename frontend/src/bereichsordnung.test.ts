// **Die Ordnung der fünf Bereiche** — UX-02
// (#1674).
//
// Was hier gehalten wird, ist die Zusage aus **UX-01.AK-04**: *„Kein
// bisheriger Anwendungsfall verliert beim Umbau seinen Zugang."* Eine
// Aufzählung, die eine Seite vergisst, ist grün — und niemand sieht es.
import { describe, expect, it } from 'vitest'

import {
  alleSeitenschluessel,
  ausserhalbDerBereiche,
  bereiche,
  bereichsseiten,
  bereichFuer,
  bereichVon,
  ersteSeiteVon,
} from './bereiche'
import { leisteneintraege } from './leistenordnung'
import { seitenschluessel } from './navigation'

describe('die Bereiche', () => {
  // **Fünf fachliche, ein Einstieg, eine Dienstzone** — UX-02.AK-01 verlangt
  // „ungefähr fünf", und AK-04, dass die Dienstzone keinen davon verbraucht.
  it('führt fünf fachliche Bereiche neben Start und Dienstzone', () => {
    const fachlich = bereiche.filter(
      (bereich) => !bereich.dienst && bereich.schluessel !== 'start',
    )

    expect(fachlich).toHaveLength(5)
    expect(bereiche.filter((bereich) => bereich.dienst)).toHaveLength(1)
  })

  // **Jeder fachliche Bereich hat eine Ziffer, die Dienstzone keine.**
  // Ihre drei Einträge tragen ⌘, ⌘' und ⌘? — die macOS-Eigenheiten aus §6.
  it('gibt jedem fachlichen Bereich genau eine Ziffer', () => {
    const ziffern = bereiche
      .filter((bereich) => !bereich.dienst)
      .map((bereich) => bereich.kuerzel)

    expect(ziffern).toEqual(['1', '2', '3', '4', '5', '6'])
    expect(bereiche.find((bereich) => bereich.dienst)?.kuerzel).toBe('')
  })

  // **Kein Bereich ist leer.** Ein Eintrag in der Leiste, der nirgendwohin
  // führt, ist eine Einladung, die man zurücknimmt.
  it('lässt keinen Bereich ohne Seite', () => {
    for (const bereich of bereiche) {
      expect(bereich.seiten.length).toBeGreaterThan(0)
    }
  })

  // **Keine Seite steht in zwei Bereichen.** Sonst stünde sie zweimal in der
  // Leiste, und die Inventurdifferenz zählte sie doppelt.
  it('führt jede Seite höchstens einmal', () => {
    expect(new Set(bereichsseiten).size).toBe(bereichsseiten.length)
  })

  // **Und jede registrierte Seite hat ihren Platz** — in einem Bereich oder
  // ausdrücklich daneben. Das ist UX-01.AK-04 in beide Richtungen.
  it('ordnet jede registrierte Seite zu', () => {
    const zugeordnet = new Set([...bereichsseiten, ...ausserhalbDerBereiche])

    for (const schluessel of seitenschluessel) {
      expect(zugeordnet.has(schluessel)).toBe(true)
    }

    expect(zugeordnet.size).toBe(seitenschluessel.length)
  })

  // **Und umgekehrt: keine erfundene Seite.**
  it('nennt keine Seite, die es nicht gibt', () => {
    for (const schluessel of bereichsseiten) {
      expect(alleSeitenschluessel).toContain(schluessel)
    }
  })
})

describe('bereichVon', () => {
  it('findet den Bereich einer Seite', () => {
    expect(bereichVon('sammlung')?.schluessel).toBe('sammeln')
    expect(bereichVon('hilfe')?.schluessel).toBe('programm')
  })

  // **Eine Lasche des Sets steht in keinem Bereich** — sie gehört in die
  // Laschenleiste, und ihr einen anzudichten hieße, sie zweimal zu führen.
  it('gibt für eine Lasche des Sets keinen', () => {
    expect(bereichVon('stil')).toBeUndefined()
  })
})

describe('ersteSeiteVon', () => {
  it('nennt die erste Seite eines Bereichs', () => {
    const spielen = bereiche.find((bereich) => bereich.schluessel === 'spielen')

    expect(spielen && ersteSeiteVon(spielen)).toBe('decks')
  })
})

describe('bereichFuer', () => {
  // **Eine Lasche des Sets hebt *Gestalten* hervor.** Der Entwurf zeichnet es
  // so: die Artboards *Karten*, *Eigenschaften*, *Werkzeuge* und *Ausgeben*
  // tragen alle `aktiv: 'gestalten'`. Wer ein Set offen hat, arbeitet in
  // diesem Bereich — auch wenn die Lasche selbst nicht in der Leiste steht.
  it('hebt für eine Lasche des Sets Gestalten hervor', () => {
    expect(bereichFuer('stil')?.schluessel).toBe('gestalten')
    expect(bereichFuer('karten')?.schluessel).toBe('gestalten')
  })

  it('gibt für eine Leistenseite ihren eigenen Bereich', () => {
    expect(bereichFuer('handeln')?.schluessel).toBe('handeln')
  })
})

describe('die Einträge der Leiste', () => {
  // **Neun Einträge: sechs Bereiche und drei Dienste.** Der Entwurf zeichnet
  // die Leiste so — ein Eintrag je Bereich, und darunter, hinter einem
  // Trenner, die drei Dienste mit ihren eigenen Kürzeln. Die vierzehn
  // Einträge von vorher sind damit neun.
  it('führt sechs Bereiche und drei Dienste', () => {
    expect(
      leisteneintraege.filter((eintrag) => eintrag.alsBereich),
    ).toHaveLength(6)
    expect(
      leisteneintraege.filter((eintrag) => !eintrag.alsBereich),
    ).toHaveLength(3)
  })

  // **Jeder Eintrag trägt ein Sinnbild**, und keine zwei dasselbe — sonst
  // sagte es nichts, was der Name nicht schon sagt.
  it('gibt jedem Eintrag ein eigenes Sinnbild', () => {
    const bilder = leisteneintraege.map((eintrag) => eintrag.sinnbild)

    for (const bild of bilder) {
      expect(bild).not.toBe('')
    }

    expect(new Set(bilder).size).toBe(bilder.length)
  })

  // **Jeder Eintrag führt irgendwohin, und jeder trägt ein Kürzel.** Ein
  // Eintrag ohne Ziel wäre eine Einladung, die man zurücknimmt; einer ohne
  // Kürzel stünde im Entwurf ohne die Spalte rechts.
  it('gibt jedem Eintrag ein Ziel und ein Kürzel', () => {
    for (const eintrag of leisteneintraege) {
      expect(alleSeitenschluessel).toContain(eintrag.ziel)
      expect(eintrag.kuerzel).not.toBe('')
    }
  })

  // **Die Ziffern stehen in der Ordnung des Entwurfs.**
  it('nennt die Ziffern eins bis sechs in dieser Reihenfolge', () => {
    const ziffern = leisteneintraege
      .filter((eintrag) => eintrag.alsBereich)
      .map((eintrag) => eintrag.kuerzel)

    expect(ziffern).toEqual(['1', '2', '3', '4', '5', '6'])
  })

  // **Und die Dienstzone trägt die macOS-Kürzel** — §6.
  it('lässt der Dienstzone ihre eigenen Kürzel', () => {
    const dienste = leisteneintraege
      .filter((eintrag) => !eintrag.alsBereich)
      .map((eintrag) => [eintrag.ziel, eintrag.kuerzel])

    expect(dienste).toEqual([
      ['einstellungen', ','],
      ['nutzer', "'"],
      ['hilfe', '?'],
    ])
  })
})
