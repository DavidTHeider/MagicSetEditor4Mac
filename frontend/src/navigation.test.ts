import { describe, expect, it } from 'vitest'

import {
  anwendungsschluessel,
  beschriftung,
  inhaltsflaeche,
  istSetseite,
  nachbarlasche,
  seiten,
  seitenschluessel,
  setschluessel,
} from './navigation'
import {
  anwendungsseiten,
  leisteneintraege,
  reservierteKuerzel,
} from './leistenordnung'
import { ausgabeseiten } from './ausgabe'
import { bereiche, ersteSeiteVon } from './bereiche'
import { setlaschen } from './laschengruppen'
import { schubladen } from './schubladen'
import { kataloge, sprachen } from './sprache/kataloge'

describe('die Seitenliste', () => {
  // **Jede genau einmal — die Reihenfolge ist eine andere Frage.**
  //
  // Hier stand `toEqual([...seitenschluessel])`, und damit war die Ordnung der
  // Leiste an die Reihenfolge der Typliste gebunden. Seit
  // #288
  // stehen erst die vier Anwendungsseiten und dann die fünf Laschen des Sets;
  // das ist eine Entscheidung über die Oberfläche und keine über den Typ.
  it('führt jede Seite genau einmal', () => {
    const schluessel = seiten.map((seite) => seite.schluessel)

    expect(new Set(schluessel).size).toBe(schluessel.length)
    expect([...schluessel].sort()).toEqual([...seitenschluessel].sort())
  })

  // **Alle Bereiche zusammen sind alle Seiten.** Eine Seite ohne Bereich
  // stünde in keiner Leiste — sichtbar wäre das erst im Betrieb, als etwas,
  // das man nicht mehr erreicht.
  //
  // **Seit UX-08 sind es drei** (#1680): die Seitenleiste, die Laschenleiste
  // des Sets und seine Schubladen; **seit UXR-09 vier** (#1917) — der
  // Ausgabeassistent, den der Knopf *Ausgeben* im Kopf öffnet. Verglichen
  // werden die Schlüssel und nicht nur ihre Zahl — eine doppelte Seite und
  // eine fehlende höben sich sonst auf.
  it('ordnet jede Seite einem Bereich zu', () => {
    const zugeordnet = [
      ...anwendungsseiten.map((seite) => seite.schluessel),
      ...setlaschen.map((seite) => seite.schluessel),
      ...schubladen.flatMap((schublade) => schublade.laschen),
      ...ausgabeseiten,
    ]

    expect(zugeordnet.sort()).toEqual([...seitenschluessel].sort())
  })

  it('gibt jeder Seite einen eigenen Pfad', () => {
    const pfade = seiten.map((seite) => seite.pfad)

    expect(new Set(pfade).size).toBe(pfade.length)
    expect(pfade.every((pfad) => pfad.startsWith('/'))).toBe(true)
  })

  it('vergibt jedes Kürzel höchstens einmal', () => {
    const kuerzel = seiten.map((seite) => seite.kuerzel).filter(Boolean)

    expect(new Set(kuerzel).size).toBe(kuerzel.length)
  })

  // **Die erste Reihe gehört der Seitenleiste, die zweite den Laschen** —
  // Werkstück C von Phase 4.10
  // (#1325).
  //
  // # Was hier vorher stand
  //
  // Eine Mischung: ⌘1 bis ⌘3 an der Leiste, ⌘4 bis ⌘0 an den Laschen eines
  // Sets, und *Vorlagen*, *Sammlung*, *Decks* und *Turniere* in einer zweiten
  // Reihe (⌥⌘1, ⌥⌘2, ⌥⌘9, ⌥⌘0), die die Leiste nicht einmal anzeigte.
  //
  // Die Begründung stand an jeder der vier Zeilen — „⌘4 bis ⌘0 gehören den
  // Laschen des Sets" — und sie war in sich schlüssig. **Nur ergab die
  // Aufteilung keinen Sinn, den man sich merken kann**: dieselbe Tastenreihe
  // führte an zwei ganz verschiedene Orte, und welcher, hing davon ab, wo man
  // gerade war.
  //
  // # Die Regel jetzt
  //
  // **⌘ + Ziffer ist die Seitenleiste. ⌥⌘ + Ziffer sind die Laschen des
  // offenen Sets.** Ein Satz, zwei Reihen, kein Rand.
  //
  // Sie kostet nichts und bringt zwei: die Laschen bekommen **zehn** Plätze
  // statt sieben, und *Symbol* und *Ausgeben* haben damit erstmals eine Taste.
  // **Die Karten sind die eine Ausnahme** — Werkstück I von Phase 4.10
  // (#1325):
  // eine Setseite, die in der Leiste steht, weil sie die Fläche ist, auf der
  // gearbeitet wird. Die Regel gilt weiter für alle anderen — und sie hier
  // auszunehmen ist eine Zeile, während sie stillschweigend zu brechen ein
  // Wächter wäre, der die halbe Zusage hält (§10).
  it('legt die erste Reihe auf die Leiste und die zweite auf die Laschen', () => {
    for (const seite of seiten) {
      const inDerLeiste =
        !istSetseite(seite.schluessel) || seite.schluessel === 'karten'

      if (/^\d$/.test(seite.kuerzel)) {
        expect(
          inDerLeiste,
          `⌘${seite.kuerzel} führt auf ${seite.schluessel} — eine Lasche`,
        ).toBe(true)
      }

      if (seite.zweitkuerzel !== '') {
        expect(
          inDerLeiste,
          `⌥⌘${seite.zweitkuerzel} führt auf ${seite.schluessel} — eine Seite der Leiste`,
        ).toBe(false)
      }
    }
  })

  // **Und die zehn Seiten mit Ziffer sind genau die zehn der Leiste.**
  //
  // **Die Null steht am Ende** (T9, #1474): auf der Tastatur liegt sie rechts
  // neben der Neun, und eine Leiste, die mit ⌘0 begänne, folgte der Zahl und
  // nicht der Hand.
  //
  // *Hilfe* und *Einstellungen* tragen keine Ziffer, sondern `?` und `,` —
  // die macOS-Eigenheiten aus §6.
  // **Über `leisteneintraege` und nicht über `anwendungsseiten`**: die Leiste
  // führt seit UX-02 Bereiche, und die Ziffer steht am Eintrag.
  // **Die Ziffer gehört seit UX-02 dem Bereich, nicht der Seite**
  // (#1674).
  //
  // # Was hier vorher stand
  //
  // Zehn Zeilen, `start=1` bis `handeln=0` — **jede** Seite der Leiste hatte
  // ihre Ziffer, und die Regel dazu lautete „⌘ + Ziffer ist die Leiste".
  //
  // Der Entwurf der Zieloberfläche ordnet die dreizehn Einträge zu **fünf
  // Bereichen**, und UX-02.AK-01 verlangt genau das. Eine Ziffer je Eintrag
  // gäbe es dann dreizehnmal für sechs Orte; die Ziffer führt jetzt in den
  // **Bereich**, und zwar auf seine erste Seite.
  //
  // **Das ist eine Migration und keine Kleinigkeit** — ⌘3 hieß *Karten* und
  // heißt jetzt *Sammeln*. Sie steht in `PROTOTYP-0.6.0.md` und damit dort,
  // wo AK-03 sie verlangt: dokumentiert, nicht stillschweigend.
  it('gibt jedem Bereich seine Ziffer und der ersten Seite darin dieselbe', () => {
    const mitZiffer = leisteneintraege
      .filter((eintrag) => /^\d$/.test(eintrag.kuerzel))
      .map((eintrag) => `${eintrag.ziel}=${eintrag.kuerzel}`)

    expect(mitZiffer).toEqual([
      'start=1',
      'sets=2',
      'sammlung=3',
      'decks=4',
      'handeln=5',
      'pakete=6',
    ])
  })

  // **Und jede Ziffer führt wirklich in ihren Bereich.** Ohne diesen Lauf
  // stünden zwei Angaben über dieselbe Taste nebeneinander — die am Bereich
  // und die an der Seite —, und sie liefen auseinander.
  it('hält die Ziffer des Bereichs und die seiner ersten Seite gleich', () => {
    for (const bereich of bereiche.filter((einer) => einer.kuerzel !== '')) {
      const erste = seiten.find(
        (seite) => seite.schluessel === ersteSeiteVon(bereich),
      )

      expect(erste?.kuerzel).toBe(bereich.kuerzel)
    }
  })

  // **⌘, führt zu den Einstellungen** — die macOS-Eigenheit aus §6.
  it('legt das Komma auf die Einstellungen', () => {
    const einstellungen = seiten.find(
      (seite) => seite.schluessel === 'einstellungen',
    )

    expect(einstellungen?.kuerzel).toBe(',')
  })

  // Eine Seite ohne Namen im Katalog fiele erst im Betrieb auf — als leerer
  // Eintrag in der Navigation, den niemand anklickt.
  it.each(sprachen)('nennt jede Seite in %s beim Namen', (sprache) => {
    for (const seite of seiten) {
      const name = beschriftung(kataloge[sprache], seite.schluessel)

      expect(name.trim().length).toBeGreaterThan(0)
    }
  })

  // Und die Namen unterscheiden sich — zwei Seiten desselben Namens sind eine
  // Navigation, in der man sich nicht zurechtfindet.
  it.each(sprachen)(
    'gibt in %s keiner Seite den Namen einer anderen',
    (sprache) => {
      const namen = seiten.map((seite) =>
        beschriftung(kataloge[sprache], seite.schluessel),
      )

      expect(new Set(namen).size).toBe(namen.length)
    },
  )
})

// **Der Nachbar einer Lasche** — die Rechnung hinter den Pfeiltasten.
//
// Sie steht als eigene Funktion, weil sie zwei Antworten hat, die am Bildschirm
// niemand auseinanderhält: „die nächste" und „gar keine".
describe('nachbarlasche', () => {
  const reihe = ['karten', 'suche', 'statistik'] as const

  it('geht vorwärts', () => {
    expect(nachbarlasche(reihe, 'karten', 1)).toBe('suche')
  })

  it('geht rückwärts', () => {
    expect(nachbarlasche(reihe, 'suche', -1)).toBe('karten')
  })

  // **Zyklisch in beide Richtungen.** Wer am Ende steht, kommt an den Anfang —
  // sonst wäre der letzte Pfeil einer, der nichts tut, und das sieht aus wie
  // ein Fehler.
  it('läuft am Ende um', () => {
    expect(nachbarlasche(reihe, 'statistik', 1)).toBe('karten')
    expect(nachbarlasche(reihe, 'karten', -1)).toBe('statistik')
  })

  it('gibt nichts für eine Seite, die nicht in der Reihe steht', () => {
    expect(nachbarlasche(reihe, 'pakete', 1)).toBeUndefined()
  })

  it('gibt nichts für eine leere Reihe', () => {
    expect(nachbarlasche([], 'karten', 1)).toBeUndefined()
  })
})

// **Die Wache über den Bereich** verengt den Typ — deshalb hier geprüft und
// nicht nur benutzt.
describe('istSetseite', () => {
  // **Über alle Setseiten und nicht über die Leiste** — seit UX-08 stehen
  // sieben davon in einer Schublade (#1680), und sie sind Setseiten wie
  // die anderen.
  it('trennt die Setseiten von den Anwendungsseiten', () => {
    for (const schluessel of setschluessel) {
      expect(istSetseite(schluessel)).toBe(true)
    }

    for (const seite of anwendungsseiten) {
      expect(istSetseite(seite.schluessel)).toBe(false)
    }
  })
})

describe('Die Breite des Inhalts', () => {
  // **Eine Arbeitsfläche ist kein Fließtext**
  // (#697).
  //
  // Der Rahmen kappte jede Seite bei 6xl — 1.152 Punkte. Gemessen an der
  // laufenden Anwendung blieben davon für ein Eingabefeld der Karte 176; auf
  // einem 1.728 Punkte breiten Fenster lagen 336 ungenutzt daneben.
  //
  // Die Kappung ist für **Text** richtig: ohne sie liefe eine Zeile über einen
  // 34-Zoll-Bildschirm. Die Laschen des Sets sind kein Text — dort liegen
  // Liste, Formular und Karte nebeneinander, und jede von ihnen hat ihre
  // eigene Breite.
  it('gibt den Setlaschen mehr Platz als den Textseiten', () => {
    expect(inhaltsflaeche('karten')).not.toBe(inhaltsflaeche('hilfe'))
  })

  it('behandelt alle Setlaschen gleich', () => {
    for (const schluessel of setschluessel) {
      expect(inhaltsflaeche(schluessel)).toBe(inhaltsflaeche('karten'))
    }
  })

  it('behandelt alle Anwendungsseiten gleich', () => {
    for (const schluessel of anwendungsschluessel) {
      expect(inhaltsflaeche(schluessel)).toBe(inhaltsflaeche('hilfe'))
    }
  })
})

// **⌘9 und ⌘0 sind reserviert** — Werkstück U1 von Phase 5.1
// (#1412).
//
// Die Bereiche *Spielen* und *Handeln* entstehen in den Phasen 5.5 und 5.6;
// ihre Tasten stehen **vorher** fest, damit keine Seite dazwischen sie
// nimmt. Eine Reservierung, die nur in einem Dokument steht, hält nicht:
// wer `kuerzel: '9'` an eine Seite schriebe, bekäme einen grünen Lauf.
describe('reservierte Kürzel', () => {
  // **Beide sind vergeben**: ⌘9 seit G1 (#1458), ⌘0 seit T9 (#1474). Eine
  // Reservierung endet damit, dass ihr Bereich da ist — und die Liste bleibt
  // stehen, weil die nächste Reservierung eine Stelle braucht.
  it('hält keine Taste mehr frei', () => {
    expect(reservierteKuerzel).toEqual({})
  })

  it('gibt keine reservierte Taste an eine Seite', () => {
    for (const seite of seiten) {
      expect(
        seite.kuerzel in reservierteKuerzel,
        `⌘${seite.kuerzel} führt auf ${seite.schluessel} — die Taste ist reserviert`,
      ).toBe(false)
    }
  })

  // **Die zweite Reihe (⌥⌘) war von der Reservierung nie berührt** — sie
  // gehört den Laschen, und ⌥⌘9 und ⌥⌘0 führen auf *Ausgeben* und
  // *Druckbogen*. Seit die Liste leer ist, hält dieser Lauf das fest, indem
  // er die zweite Reihe **vollständig** nennt: eine leere Liste gegen eine
  // leere Liste zu halten sagte nichts (§10).
  it('lässt die zweite Reihe unberührt', () => {
    const zweite = seiten
      .filter((seite) => /^[09]$/.test(seite.zweitkuerzel))
      .map((seite) => seite.schluessel)

    expect(zweite).toEqual(['ausgeben', 'druckbogen'])
  })
})
