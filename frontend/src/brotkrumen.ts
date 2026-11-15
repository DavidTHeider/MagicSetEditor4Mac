// **Wo du bist** — UX-03
// (#1675),
// gezeichnet als `brot` in jedem Artboard des Entwurfs außer `Main`.
//
// # Was eine Krume sagt und was nicht
//
// **UX-03.AK-02**: *„Breadcrumbs zeigen fachliche Hierarchien, nicht die
// interne Komponentenstruktur."* Die Hierarchie dieser Anwendung ist
// **Bereich › Objekt › Fläche** — nicht `Rahmen › Setarbeitsplatz › Laschen`.
//
// # Warum eine eigene Datei
//
// **Weil die Rechnung ohne einen Baum prüfbar sein soll.** Was dasteht, ist
// eine Liste; *wie* sie aussieht, steht in `bausteine/Brotkrume.tsx`.
import { bereichFuer, ersteSeiteVon } from './bereiche'
import { beschriftung, istSetseite } from './navigation'
import { schubladeVon } from './schubladen'

import type { Katalog } from './sprache/kataloge'
import type { Seitenschluessel } from './navigation'

// Krume ist eine Station des Weges.
export interface Krume {
  beschriftung: string

  // ziel ist die Seite, auf die sie führt — **fehlt bei der letzten**, denn
  // die ist, wo man steht. Ein Verweis auf die eigene Seite wäre ein Knopf,
  // der nichts tut.
  ziel?: Seitenschluessel
}

// Bezugsobjekt ist das, woran gerade gearbeitet wird.
//
// **Optional, und zwar auf jeder Seite.** Die Setliste hat keines, die Karten
// eines offenen Sets haben eines. Wer es für Pflicht hielte, erfände auf der
// Setliste ein Objekt, das es nicht gibt.
export interface Bezugsobjekt {
  name: string
  ziel: Seitenschluessel
}

// brotkrumen gibt den Weg zur aufgeschlagenen Fläche.
//
// # Warum die Startseite keinen bekommt
//
// **Weil sie der Anfang ist.** Der Entwurf zeichnet sie ohne (`Main.json`
// trägt kein `brot`), und ein Weg mit einer einzigen Station, die auf sich
// selbst zeigt, sagt nichts.
//
// # Und was eine Seite selbst über ihren Weg sagt
//
// **`weg` steht an der Stelle ihres Namens** — UX-03.AK-02: über dem Tisch
// *Spielen › Partie gegen Bo › Am Tisch*, wie im Artboard *Tisch*, und über
// den Karten die aufgeschlagene statt „Karten". Bereich und Set gibt weiter
// der Rahmen; was dahinter kommt, weiß nur die Seite (`seitenweg.tsx`). Ein
// leerer Weg ist keiner.
export function brotkrumen(
  texte: Katalog,
  aktuell: Seitenschluessel,
  objekt?: Bezugsobjekt,
  weg?: readonly Krume[],
): readonly Krume[] {
  if (aktuell === 'start') {
    return []
  }

  const bereich = bereichFuer(aktuell)

  /* v8 ignore next 3 -- jede Seite außer den Laschen steht in einem Bereich,
     und die Laschen fallen auf *Gestalten* zurück; `bereichsordnung.test.ts`
     hält das. TypeScript sieht es nicht. */
  if (bereich === undefined) {
    return []
  }

  const stationen: Krume[] = [
    {
      beschriftung: texte.allgemein.bereiche[bereich.schluessel],
      ziel: ersteSeiteVon(bereich),
    },
  ]

  // **Das Objekt steht nur, wo die Fläche über es spricht** — also auf einer
  // **Lasche des offenen Sets**.
  //
  // # Nicht auf jeder Seite
  //
  // Ein Set bleibt offen, während jemand die Sammlung ansieht oder handelt;
  // mit `if (objekt !== undefined)` allein stünde dort „Sammeln › Werkstatt
  // der Dämmerung › Sammlung" — ein Zusammenhang, den es nicht gibt.
  //
  // **Der Entwurf zeichnet es genauso**: das Set steht im `brot` der
  // Artboards *Karten*, *Eigenschaften*, *Werkzeuge* und *Ausgeben* — und in
  // keinem anderen. Auch nicht auf der **Setliste**: sie führt alle Sets, und
  // eines davon hervorzuheben wäre eine zweite Auskunft über eine Liste.
  if (objekt !== undefined && istSetseite(aktuell)) {
    // **Kein Verweis, wenn das Ziel die eigene Seite ist.** Auf der
    // Kartenfläche *ist* man beim Set; ein Knopf dorthin führte auf die
    // Stelle, an der man steht.
    //
    // **Zwei Verbünde und kein `ziel: undefined`**: `exactOptionalPropertyTypes`
    // unterscheidet „die Eigenschaft fehlt" von „sie steht auf `undefined`",
    // und gemeint ist das erste.
    stationen.push(
      objekt.ziel === aktuell
        ? { beschriftung: objekt.name }
        : { beschriftung: objekt.name, ziel: objekt.ziel },
    )
  }

  // **Ein Weg der Seite steht an der Stelle ihres Namens** — und nur dort:
  // das Set davor bleibt stehen (*Gestalten › Werkstatt der Dämmerung ›
  // Aschenwächter*, Artboard *Karten*).
  if (weg !== undefined && weg.length > 0) {
    return [...stationen, ...weg]
  }

  // **Eine Seite in einer Schublade endet bei der Schublade** (UX-08,
  // #1680): *Gestalten › Set › Eigenschaften*, wie im Entwurf. Welche Lasche
  // darin offen ist, sagt die Schublade selbst.
  const schublade = schubladeVon(aktuell)

  stationen.push({
    beschriftung:
      schublade === undefined
        ? beschriftung(texte, aktuell)
        : texte.allgemein.arbeitsplatz.schubladen[schublade.schluessel].knopf,
  })

  return stationen
}
