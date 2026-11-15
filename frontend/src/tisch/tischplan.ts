// **Der Tisch als Plan** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum ein Plan und nicht gleich ein Bild
//
// **Weil dieses Projekt die Naht schon einmal so geschnitten hat.** §2 sagt
// über den Rasterer: „`internal/render` bekommt einen **Zeichenplan**, keine
// Skripte" — und die Probe darauf war `internal/svg`, ein zweiter Verbraucher
// desselben Plans. Hier ist es dieselbe Naht: was wo liegt, ist eine
// **Rechnung**; sie auf die Grafikkarte zu bringen, ist eine **Naht**. Der
// Plan wird gemessen, die Naht ist dünn, und die zweite Erscheinung — die
// semantische Schicht für Tastatur und VoiceOver — liest denselben Plan
// statt einer eigenen Wahrheit.
//
// # Wo ein Objekt liegt, entscheidet nicht immer derselbe
//
// **CR 403.1** (Fassung 2026-08-19): Permanente liegen vor dem, der sie
// *kontrolliert*. **CR 400.3**: was in eine Bibliothek, einen Friedhof oder
// eine Hand ginge, die nicht dem *Eigner* gehört, geht in die des Eigners.
// Eine gestohlene Kreatur steht also beim Dieb und fällt in den Friedhof
// ihres Eigners — und genau das rechnet `sitzFuer`.
import { sitzflaechen } from './sitzflaeche'
import { zonenDerPartie } from '../spiel/partiezonenliste'
import { zonenfelder } from './zonenfeld'
import { schlachtfeldlagen } from './schlachtfeld'
import { eigeneLagen, type Stelle } from './eigenelagen'
import { normalisiert, stapelfeld, zurMitteVon } from './tischgeometrie'
import { zonenlagen } from './zonenlage'

import type { Partie, Spielobjekt } from '../apipartie'
import type { Sitzflaeche } from './sitzflaeche'
import type { Zonenfeld } from './zonenfeld'
import type { Gelegtes, Lage } from './zonenlage'

// Tischsicht sind die Entscheidungen, die nur diesem Fenster gehören.
//
// **Sie gehen nicht an den Wirt.** Eine gewendete Karte ist eine Frage der
// Darstellung; die Partie ändert sich dadurch nicht.
export interface Tischsicht {
  // gewendet sind die Objekte, von denen die zweite Seite gezeigt wird.
  gewendet: string[]

  // eigene sind die Lagen, die jemand selbst gewählt hat — PLAY-04.AK-03
  // (#1712).
  //
  // **Leer heißt: der Tisch ordnet.** Das ist die Vorgabe
  // (`domain.AutoanordnungAn`), und sie ist es, weil ein Tisch, der nichts
  // ordnet, bis jemand jede Karte angefasst hat, beim ersten Öffnen ein
  // Haufen wäre.
  eigene?: Readonly<Record<string, Stelle>> | undefined
}

// Tischlage ist eine Lage samt allem, was der Zeichner über sie wissen muss.
export interface Tischlage extends Lage {
  zone: string
  sitz: string

  // seite ist `vorne` oder `zweite`.
  seite: string

  karteID: string
  name: string

  // geheim heißt: **dieser** Empfänger darf nicht hineinsehen. Es führt zum
  // selben Bild wie `verdeckt` und ist doch etwas anderes — das eine ist eine
  // Tatsache über die Karte, das andere über den Betrachter.
  geheim: boolean

  zaehler: Record<string, number>
  haengtAn: string

  // art ist, **was** dieses Objekt ist — `karte`, `spielstein`, `emblem`
  // oder `faehigkeit` (`domain.Objektart`). PLAY-03.AK-04
  // (#1711)
  // verlangt, dass ein Spielstein als solcher zu erkennen ist — und zwar
  // auch für den, der die Karte nicht sieht, sondern hört.
  art: string

  // haufenzahl ist, wie viele Objekte auf diesem Haufen liegen — **1, wo
  // keiner ist**.
  //
  // **Sie steht hier und wird nicht am Knopf gezählt**: der Plan gruppiert
  // die Haufen ohnehin (`zonenlage.haufenweise`), und zweimal zu zählen
  // hieße, zwei Antworten auf dieselbe Frage zu haben (§10).
  haufenzahl: number

  // haufen ist der **sichtbare** Haufen, in dem ein Objekt liegt — nie der
  // Stapel der Regeln. Die Abnahme von #1460 nennt das ausdrücklich, und
  // CR 400.5 trägt es: die Reihenfolge des Stapels darf sich nicht ändern,
  // die eines Haufens auf dem Schlachtfeld sehr wohl.
  haufen: string

  // untergelegt heißt: diese Karte liegt in einem **Stapel** unter einer
  // anderen (#2472). Gezeichnet wird sie — sie zeigt die Dicke —, ein
  // eigener Knopf am gezeichneten Tisch ist sie nicht: 0,6 Punkte darunter
  // wäre er ein verdecktes Ziel (WCAG 2.2, 2.5.8), und treffen könnte ihn
  // ohnehin niemand. Die Liste und „Hineinsehen …“ erreichen jede einzeln.
  untergelegt: boolean
}

// Zonenzahl ist eine Zone, von der mehr drin ist, als zu sehen.
export interface Zonenzahl {
  zone: string
  sitz: string
  x: number
  y: number
  zahl: number
}

export interface Tischplan {
  breite: number
  hoehe: number

  sitze: Sitzflaeche[]
  felder: Zonenfeld[]
  lagen: Tischlage[]
  zahlen: Zonenzahl[]

  // heimatlos sind die Objekte, für die kein **Feld** gefunden wurde.
  //
  // **Sie verschwinden nicht still.** Eine unbekannte Zone — CR 400.1 nennt
  // die Ante-Zone älterer Karten — oder ein Sitz, den diese Sicht nicht
  // führt, hätte ein Objekt sonst lautlos vom Tisch genommen. §8 sagt, wie
  // das ausgeht: was nicht gelesen werden kann, wird **benannt**, und der
  // Rest wird gezeigt.
  //
  // # „Kein Feld" und nicht „keine Lage"
  //
  // **Der Unterschied ist der ganze Nutzen dieser Liste.** Der erste Anlauf
  // zählte jedes Objekt ohne Lage — und eine Bibliothek mit 53 Karten meldete
  // **45 Befunde**, weil ein Stapel nur seine obersten acht zeichnet
  // (`Stapelsicht`). Eine Diagnose, die bei jeder gewöhnlichen Partie
  // dreistellig ausfällt, liest niemand mehr. Gefunden am 2026-09-16 von den
  // zehn Szenen, beim allerersten Lauf.
  heimatlos: string[]
}

// sitzFuer sagt, an wessen Platz ein Objekt liegt.
//
// **CR 400.3 für die drei eigenen Zonen, CR 403.1 für das Schlachtfeld.**
export function sitzFuer(objekt: Spielobjekt): string {
  return objekt.zone === 'bibliothek' ||
    objekt.zone === 'friedhof' ||
    objekt.zone === 'hand'
    ? objekt.eigner
    : objekt.beherrscher
}

export function tischplan(
  partie: Partie,
  sicht: Tischsicht = { gewendet: [] },
): Tischplan {
  const flaechen = sitzflaechen(
    partie.sitze.map((einer) => einer.kennung),
    partie.fuer,
  )
  // **Nur die Zonen, die diese Partie braucht** — PLAY-03.AK-01
  // (#1711).
  const zonen = zonenDerPartie(partie)
  const felder = flaechen.flatMap((flaeche) => zonenfelder(flaeche, zonen))
  const stapel = stapelfeld(flaechen)
  if (stapel) felder.push(stapel)

  const lagen: Tischlage[] = []
  const zahlen: Zonenzahl[] = []
  const untergebracht = new Set<string>()

  for (const feld of felder) {
    const gelegte = objekteIn(partie.objekte, feld)
    for (const eines of gelegte) untergebracht.add(eines.kennung)

    // **Das Schlachtfeld liegt in Gruppen** — PLAY-03.AK-02: Kreaturen zur
    // Mitte, Länder und sonstiges Bleibendes dahinter. Jede andere Zone
    // rechnet wie bisher; eine Hand ist keine Permanentenreihe.
    const gerechnet =
      feld.zone === 'tisch'
        ? schlachtfeldlagen(feld, gelegte, zurMitteVon(flaechen, feld.sitz))
        : zonenlagen(feld, gelegte)

    // **Wie viele auf einem Haufen liegen** — PLAY-03.AK-04
    // (#1711).
    // Gezählt wird je Zone, denn ein Haufenname gilt in seiner Zone; und
    // hier, weil die Gelegten ohnehin schon dastehen.
    const jeHaufen = new Map<string, number>()

    for (const eines of gelegte)
      if (eines.haufen !== '')
        jeHaufen.set(eines.haufen, (jeHaufen.get(eines.haufen) ?? 0) + 1)

    // **Und dann kommt, was jemand selbst gelegt hat** — je Zone, denn
    // eine eigene Lage gilt in dem Feld, in dem sie entstanden ist.
    const nachWunsch = eigeneLagen(
      gerechnet.lagen.map((lage) =>
        tischlage(
          lage,
          feld,
          lage.gelegtes.objekt,
          sicht,
          jeHaufen.get(lage.gelegtes.haufen) ?? 1,
        ),
      ),
      sicht.eigene ?? {},
    )

    // **Im Stapel ist nur die oberste greifbar** (#2472).
    if (feld.gestalt === 'stapel') unterlegen(nachWunsch)

    for (const lage of nachWunsch) lagen.push(lage)

    if (gerechnet.zahl > gerechnet.lagen.length || feld.gestalt === 'stapel')
      zahlen.push({
        zone: feld.zone,
        sitz: feld.sitz,
        x: feld.x,
        y: feld.y + feld.hoehe,
        zahl: gerechnet.zahl,
      })
  }

  return normalisiert({
    breite: 0,
    hoehe: 0,
    sitze: flaechen,
    felder,
    lagen,
    zahlen,
    heimatlos: partie.objekte
      .filter((eines) => !untergebracht.has(eines.kennung))
      .map((eines) => eines.kennung),
  })
}

interface Zugeordnetes extends Gelegtes {
  objekt: Spielobjekt
}

// objekteIn sammelt die Objekte einer Zone in ihrer Reihenfolge.
//
// **Nach `ordnung`, und die ist nur in offenen Zonen gefüllt** — in einer
// verborgenen bleibt die Reihenfolge der Liste, und das ist richtig so: sie
// ist dort ohnehin niemandes Wissen (CR 401.2).
function objekteIn(objekte: Spielobjekt[], feld: Zonenfeld): Zugeordnetes[] {
  return objekte
    .filter(
      (eines) =>
        eines.zone === feld.zone &&
        (feld.sitz === '' || sitzFuer(eines) === feld.sitz),
    )
    .sort((eines, anderes) => eines.ordnung - anderes.ordnung)
    .map((objekt) => ({
      kennung: objekt.kennung,
      getappt: objekt.getappt,
      verdeckt: objekt.verdeckt || objekt.geheim,

      // **Der Haufen geht in die Rechnung** — er entscheidet, wie viele
      // Plätze eine Zone braucht
      // (#1756).
      haufen: objekt.haufen,
      objekt,
    }))
}

function tischlage(
  lage: Lage,
  feld: Zonenfeld,
  objekt: Spielobjekt,
  sicht: Tischsicht,
  haufenzahl: number,
): Tischlage {
  return {
    ...lage,
    zone: feld.zone,
    sitz: feld.sitz,
    seite: sicht.gewendet.includes(objekt.kennung) ? 'zweite' : 'vorne',
    karteID: objekt.karteID,
    // **Der Name der Karte, sonst der des Spielsteins** (M2, #1604).
    //
    // `Sichtobjekt.Name` trägt den Namen dessen, wo **keine** Karte
    // dahintersteht; bei einer Karte ist er leer, und am Tisch stand
    // deshalb die Objektkennung statt „Goblin Guide" — gesehen bei der
    // ersten Sichtprüfung mit einem echten Deck am 2026-09-17.
    // **`?:` und nicht `??`**: ein **leerer** Name soll durchfallen, und
    // `??` fiele nur bei `null` durch. Eine Karte ohne Eigenschaften hat
    // keinen Namen, und ein Spielstein hat seinen an `objekt.name`.
    name:
      eigenschaftenname(objekt) === ''
        ? objekt.name
        : eigenschaftenname(objekt),
    geheim: objekt.geheim,
    zaehler: objekt.zaehler,
    haengtAn: objekt.haengtAn,
    art: objekt.art,
    haufenzahl: haufenzahl,
    haufen: objekt.haufen,
    untergelegt: false,
  }
}

// unterlegen markiert in einem Stapel jede Lage unter der obersten.
function unterlegen(lagen: Tischlage[]): void {
  const oben = Math.max(...lagen.map((eine) => eine.hoehenrang))

  for (const lage of lagen) lage.untergelegt = lage.hoehenrang !== oben
}

// Zug ist eine Verschiebung, die **nur** die Darstellung betrifft.
//
// **#1460 verlangt die Trennung wörtlich**: „Pointer-Drag bewegt nur
// Darstellung; Drop sendet validierten Command, Antwort bestätigt oder
// korrigiert." Solange gezogen wird, weiß der Wirt von nichts — und das ist
// der Grund, warum ein abgelehnter Drop nichts kaputt machen kann: er hat nie
// etwas geändert.
export interface Zug {
  kennung: string

  // dx und dy sind die Verschiebung in Millimetern.
  dx: number
  dy: number
}

// mitZug gibt den Plan, wie er **während** eines Zuges aussieht.
//
// **Nur die eine Lage wandert**, und sie wandert nach ganz oben: was man in
// der Hand hält, liegt über allem anderen. Alles übrige bleibt, wo es ist —
// das ist dieselbe Zusage wie in §7, Regel 4 („eine Rücknahme stellt nur den
// betroffenen Datensatz wieder her"), nur vorwärts gelesen.
export function mitZug(plan: Tischplan, zug: Zug | undefined): Tischplan {
  if (!zug) return plan

  const oben = plan.lagen.reduce(
    (hoechster, eine) => Math.max(hoechster, eine.hoehenrang),
    0,
  )

  return {
    ...plan,
    lagen: plan.lagen.map((eine) =>
      eine.kennung === zug.kennung
        ? {
            ...eine,
            x: eine.x + zug.dx,
            y: eine.y + zug.dy,
            hoehenrang: oben + 1,
          }
        : eine,
    ),
  }
}

// eigenschaftenname gibt den Kartennamen — leer, wo es keinen gibt.
function eigenschaftenname(objekt: Spielobjekt): string {
  return objekt.eigenschaften?.name ?? ''
}
