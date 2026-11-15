// **Was an einem Zonenfeld steht** — Befund (b) und (c) des Entwurfs der
// Zieloberfläche, jetzt PLAY-03.AK-06
// (#1711).
//
// # Die zwei Befunde
//
// **(b)** `pixinaht.ts` gab einem Zonenfeld einen 0,8-px-Strich und sonst
// nichts: keine Füllung, keinen Namen, kein Zeichen. Wer den Tisch ansah, sah
// dünne Striche und musste raten, welcher die Bibliothek ist.
//
// **(c)** `tischplan.ts` **berechnet** `Zonenzahl` — wie viele Karten in einer
// Zone liegen, von der nicht alles zu sehen ist — und legt sie an eine Stelle.
// **Gezeichnet hat sie niemand.** Eine berechnete Größe ohne Verbraucher ist
// dieselbe Gestalt, vor der §2 bei Paketen warnt.
//
// # Warum das hier steht und nicht im Zeichner
//
// **Weil es eine Entscheidung ist.** Wo ein Schild sitzt, woran es verankert
// ist und was darauf steht, lässt sich messen; `pixinaht.ts` und
// `pixibeiwerk.ts` stehen in `.covignore`, weil sie **zeichnen**. Dieselbe
// Trennung wie bei `markenpunkte` und `beziehungslinien`.
//
// # Und warum der Name von außen kommt
//
// **Weil eine Zone in zwei Sprachen zwei Namen hat.** Die Kataloge liegen in
// `sprache/`, und ein Zeichner, der sie läse, hätte die Oberflächensprache im
// Renderer. Der Aufrufer gibt das Wort; hier steht, wo es hinkommt.

import { Namenshoehe } from './masse'

import type { Tischplan } from './tischplan'

// Schildgroesse ist die Höhe der Schrift in Weltpunkten — in der Stufe
// „mittel" der Einstellung `tisch.beschriftung` (`schildmass.ts`).
//
// **Sechzehn in einer Leiste von vierundzwanzig** (`masse.ts`,
// [Namenshoehe]): zwei Drittel, und auch „groß" (ein Fünftel mehr) bleibt
// darin. Bei der Anfangskamera eines Fensters von 1.440 Punkten — rund 0,58
// Punkte je Einheit — sind das gut neun Punkte. **Zwölf** wären bei
// derselben Kamera sieben: gelesen hätte sie niemand gern.
export const Schildgroesse = 16

// Kurzschildgroesse ist die Schrift der Kurzform über einem Stapel.
//
// **Zwölf und ungesperrt**: „FRIEDH." braucht darin rund 46 Einheiten, und
// über einem Stapel stehen 57 zur Verfügung (63 abzüglich zweier
// Schildabstände) — auch in „groß" noch 55.
export const Kurzschildgroesse = 12

// Schildabstand ist, wie weit ein Schild von der Feldkante wegrückt.
export const Schildabstand = 3

// Zonenschild ist ein Wort an seinem Platz.
export interface Zonenschild {
  // kennung ist Zone und Sitz — der Zeichner hält seine Textobjekte daran.
  //
  // **Ohne sie baute er bei jedem Zug neue.** Text ist teuer: eine
  // Neuvermessung der Schrift je Bild wäre derselbe Fehler, den
  // `pixibeiwerk.ts` bei den Markenzahlen schon einmal vermieden hat.
  kennung: string

  x: number
  y: number
  text: string

  // groesse ist die Höhe der Schrift in Weltpunkten.
  //
  // **Am Schild und nicht am Zeichner**, seit es zwei Größen gibt: ein
  // Sitzschild nennt den Spieler und seinen Stand, ein Zonenschild sagt, was
  // in einem Feld liegt. Gleich groß wäre beides gleich wichtig.
  groesse: number

  // ankerY sagt, wo der Text senkrecht an seinem Punkt hängt — 0 oben, 0,5
  // in der Mitte, 1 unten.
  //
  // **Seit dem 2026-09-24 in der Mitte seines Bandes**: der Name und die
  // Anzahl in der Leiste über den Karten, der Kopf im Band an der
  // Außenkante des Sitzes. Vorher hingen sie von einer Kante in die Fläche —
  // der Name auf die erste Karte, der Kopf auf den Namen.
  ankerY: number

  // ankerX sagt dasselbe waagerecht — fehlt es, steht der Text rechts von
  // seinem Punkt. **Die Anzahl steht am rechten Ende der Leiste**, wie im
  // Artboard (`tzone-zahl`), und wächst nach links.
  ankerX?: number

  // art sagt, was das Schild ist — ein Sitzkopf, ein Name, eine Anzahl.
  //
  // **Daran hängt, ob es bei „aus" stehen bleibt** (`schildmass.ts`): eine
  // Anzahl ist eine Auskunft, die sonst nirgends steht; Kopf und Name stehen
  // auch in der Spielerleiste und in den Kartenknöpfen.
  art: Schildart
}

// Schildart unterscheidet die Arten von Schildern auf dem Tisch — der
// Kurzname ist der Name über einem Stapel.
export type Schildart = 'kopf' | 'name' | 'kurzname' | 'zahl'

// zonenschilder gibt Namen und Anzahlen zu einem Plan.
//
// **Der Name steht an jedem Feld, die Anzahl nur, wo es eine gibt.**
// `tischplan` legt eine `Zonenzahl` nur an, wo mehr drin ist als zu sehen
// oder wo ein Stapel liegt — eine Zahl an einer Zone, in der alles offen
// liegt, wäre eine Auskunft, die man abzählen kann.
export function zonenschilder(
  plan: Tischplan,
  wort: (zone: string, kurz: boolean) => string,
): Zonenschild[] {
  // **Über einem Stapel die Kurzform** — dort ist die Leiste so schmal wie
  // eine Karte ([Kurzschildgroesse]).
  const schilder: Zonenschild[] = plan.felder.map((feld) => {
    const stapel = feld.gestalt === 'stapel'

    return {
      kennung: `${feld.sitz}:${feld.zone}:name`,
      x: feld.x + Schildabstand,
      y: feld.y - Namenshoehe / 2,
      text: wort(feld.zone, stapel),
      groesse: stapel ? Kurzschildgroesse : Schildgroesse,
      ankerY: 0.5,
      art: stapel ? 'kurzname' : 'name',
    }
  })

  // **Die Anzahl steht in derselben Leiste, rechts** — `tischplan` legt ihren
  // Punkt auf die Unterkante des Feldes; die Leiste gehört zum Feld, und das
  // Feld findet sich über Sitz und Zone.
  for (const zahl of plan.zahlen) {
    const feld = plan.felder.find(
      (eines) => eines.sitz === zahl.sitz && eines.zone === zahl.zone,
    )
    if (feld === undefined) continue

    // **Auf einem Stapel sitzt die Anzahl unten rechts auf ihm** — neben
    // der Kurzform ist in seiner Leiste kein Platz. Der Zeichner legt eine
    // Marke darunter, damit sie auch auf einem Kartenrücken zu lesen ist.
    const stapel = feld.gestalt === 'stapel'

    schilder.push({
      kennung: `${zahl.sitz}:${zahl.zone}:zahl`,
      x: feld.x + feld.breite - Schildabstand,
      y: stapel
        ? feld.y + feld.hoehe - Schildabstand
        : feld.y - Namenshoehe / 2,
      text: String(zahl.zahl),
      groesse: Schildgroesse,
      ankerY: stapel ? 1 : 0.5,
      ankerX: 1,
      art: 'zahl',
    })
  }

  return schilder
}
