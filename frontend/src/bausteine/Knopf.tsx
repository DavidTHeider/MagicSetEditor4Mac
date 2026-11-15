// Ein Knopf in drei Gewichtungen.
//
// # Warum drei und nicht beliebig viele
//
// Eine Seite hat **eine** Haupthandlung. Alles andere ist Nebenhandlung, und
// was nur ein Verweis ist, soll nicht wie ein Knopf aussehen. Mehr
// Gewichtungen hießen: auf jeder Seite neu entscheiden, und damit wieder
// verschieden.
//
// # Der Fokusring steht an jedem
//
// AGENTS.md §6 verlangt ihn sichtbar. Er sitzt hier **einmal** statt an jeder
// Fundstelle; eine vergessene Stelle wäre sonst nicht zu sehen, bis jemand mit
// der Tastatur dort ankommt.
import type { Tooltipgriffe } from './Tooltip'
import { useRef, type ReactNode, type RefObject } from 'react'

import { useFokusUeberAnhalten } from '../haken/useFokusUeberAnhalten'
import { useSchreibsperre } from '../haken/schreibrecht'

import { hoverHaupt, hoverNeben, rueckmeldung } from './knopfklassen'

// Gewicht sagt, wie laut ein Knopf spricht.
export type Gewicht = 'haupt' | 'neben' | 'leise'

interface KnopfEigenschaften {
  // gewicht ist **`neben`, wenn niemand etwas sagt**: der häufigste Fall, und
  // der, bei dem ein Versehen am wenigsten kostet. Ein versehentlicher
  // Hauptknopf zöge den Blick auf die falsche Handlung.
  gewicht?: Gewicht

  // angehalten heißt: der Vorgang läuft noch (§7 Regel 4). Der Knopf sagt es
  // über `disabled` und nicht über die Farbe allein.
  angehalten?: boolean

  // bezeichnung ist der Name für ein Vorleseprogramm, wenn die Beschriftung
  // allein nicht reicht.
  //
  // **In einer Liste steht „Sichern" fünfmal.** Wer nur die Beschriftung hört,
  // weiß nicht, welches Set gemeint ist; `aria-label` nennt es dazu. Ohne
  // Angabe steht kein Attribut da — und die Beschriftung gilt.
  bezeichnung?: string

  // beschreibung sind die Griffe eines [Tooltip] — er **beschreibt** den
  // Knopf, er benennt ihn nicht.
  //
  // **Warum der Knopf sie kennt und nicht der Aufrufer sie streut**: `Knopf`
  // gibt keine unbekannten Eigenschaften weiter, und ein zweites `<span>`
  // zwischen Tooltip und Knopf zerbräche das Gitter der Platzliste.
  beschreibung?: Tooltipgriffe

  // beschriebenVon nennt ein Element, das den Knopf beschreibt, **ohne
  // Tooltip** — `aria-describedby`; `beschreibung` geht vor. Und titel ist
  // der Hinweis unter dem Zeiger, `title`.
  //
  // **Für den Knopf über einer Spalte des Arbeitsplatzes** (UXR-02, #1910):
  // er nennt sein Kürzel im Titel wie der Knopf der Seitenleiste, und bei
  // Enge beschreibt ihn ein Satz, der nicht mehr in der Zeile steht. Ein
  // Tooltip verbrauchte dort Escape, solange er unter dem ruhenden Zeiger
  // stand — und das erste Escape schloss den Hinweis statt der Spalte.
  beschriebenVon?: string | undefined
  titel?: string | undefined

  // aufKlick darf fehlen — aber **nur** bei einem Knopf, der dauerhaft
  // angehalten ist.
  //
  // Ein Knopf am Rand einer Liste („nach oben" auf der ersten Karte) hat
  // nichts zu tun. Ihm eine leere Behandlung zu geben wäre eine Funktion, die
  // kein Test je aufruft — und eine Zeile, die aussieht, als könnte etwas
  // geschehen.
  //
  // **Geprüft wird das nicht.** Wer sie bei einem bedienbaren Knopf vergisst,
  // bekommt einen, der nichts tut; das fällt beim ersten Klick auf und in
  // keinem Wächter (§10).
  aufKlick?: () => void

  // aufDruck greift schon beim **Drücken** statt beim Klick.
  //
  // # Wofür
  //
  // Für Knöpfe, die nur da sind, solange ein Feld den Fokus hat — die
  // Vervollständigung im Quelltextfeld
  // (#1164).
  // Ein Klick nimmt dem Feld zuerst den Fokus; räumt dessen `onBlur` die
  // Liste, ist der Knopf weg, bevor er gedrückt ist.
  //
  // **Für die Tastatur bleibt `aufKlick` der Weg**: ein `mousedown` gibt es
  // dort nicht, und ein Knopf, der nur mit der Maus geht, ist keiner (§6).
  // Wer `aufDruck` setzt, setzt es zusätzlich.
  aufDruck?: () => void

  // ausgeklappt sagt, ob das, was dieser Knopf auf- und zuklappt, gerade offen
  // steht — `aria-expanded`.
  //
  // **Ohne Angabe steht kein Attribut da.** Ein `aria-expanded="false"` an
  // einem Knopf, der nichts aufklappt, sagt einem Vorleseprogramm, es gebe
  // etwas zu öffnen (#697).
  ausgeklappt?: boolean

  // steuert nennt die Kennung dessen, was der Knopf auf- und zuklappt —
  // `aria-controls`. Steht das Ziel nicht da, entfällt es: eine Kennung, die
  // ins Leere zeigt, ist schlechter als keine.
  steuert?: string

  // gedrueckt macht den Knopf zu einem Umschalter in einer Gruppe —
  // `aria-pressed`, für die Zoomleiste (U5, #1416). Ohne Angabe ist er kein
  // Umschalter, und das Attribut fehlt.
  //
  // **Gedrückt sieht er aus wie ein Hauptknopf** — bis zum 2026-09-30
  // wechselte nur das Attribut, und welcher Umschalter galt, sah niemand.
  gedrueckt?: boolean

  // griff gibt dem Aufrufer den Knopf in die Hand — für den Fokus, den eine
  // `Entscheidung` zurückgibt oder ein `Menue` beim Schließen holt. Dieselbe
  // Form wie `Pflichtfeld.griff`.
  griff?: RefObject<HTMLButtonElement | null>

  // hatMenue sagt einem Vorleseprogramm, dass hinter dem Knopf ein Menü
  // liegt — `aria-haspopup="menu"` (Werkstück U2 von Phase 5.1, #1413).
  hatMenue?: boolean

  // aufPfeil greift bei ↓ und ↑ auf dem Knopf — das Muster eines Menüknopfs:
  // ↓ öffnet auf dem ersten Eintrag, ↑ auf dem letzten. Der Wert ist die
  // Stelle, von der aus gezählt wird: `0` von vorn, `-1` von hinten.
  aufPfeil?: (von: 0 | -1) => void

  // klein macht Schrift und Rand kleiner, nicht die Gewichtung — für eine
  // Reihe gleichrangiger Handlungen, wie die Spalte neben dem Tisch sie führt
  // (UX-08, #1680). Das Artboard *Tisch* zeigt sie so; in voller Größe
  // brauchten zwanzig Handlungen einer Karte zehn Zeilen.
  klein?: boolean

  // schreibt heißt: diese Handlung ändert etwas — G10 (#2276). Wer in
  // diesem Profil nur lesen darf, bekommt den Knopf **gesperrt und erklärt**
  // (`aria-disabled` und ein Titel), statt dass der Klick erst am Tor
  // scheitert. **Fokussierbar bleibt er**: ein Vorleseprogramm soll lesen
  // können, warum er nicht geht — `disabled` nähme ihn aus der Reihe.
  schreibt?: boolean

  // ohneTabstelle nimmt den Knopf aus der Tabulatorreihe — für einen Knopf in
  // einer Gitterzeile, die selbst die Tabstelle ist: erreicht wird er dann
  // mit → aus der Zeile (`gitterzeilentaste`), wie die Zelle daneben. Ohne
  // das hätte eine Liste mit 331 Zeilen 331 Tabstellen mehr.
  ohneTabstelle?: boolean

  // kuerzel ist das Tastenkürzel in der Schreibweise von
  // `aria-keyshortcuts` — gesetzt vom [Aktionsknopf], und nur dort, wo der
  // Aufrufer es verdrahtet hat (UX-10.AK-01, #1682).
  kuerzel?: string | undefined

  // aktion ist die Art aus dem Aktionskatalog — `data-aktion`, damit eine
  // Probe und ein Tastenhorcher der Zeile den Knopf finden, ohne seine
  // Aufschrift zu kennen.
  aktion?: string

  children: ReactNode
}

const gemeinsam =
  'inline-flex items-center gap-2 rounded-md font-medium ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-50 ' +
  `aria-disabled:cursor-not-allowed aria-disabled:opacity-50 ${rueckmeldung} `

// masse sind Rand und Schrift je Größe.
const masse = {
  gross: 'px-3 py-2 ',
  klein: 'px-2 py-1 text-sm ',
}

// gewichte ist die Farbe je Gewichtung.
//
// **`text-marketext` auf `bg-marke`** — gemessen und in `docs/KONTRASTE.md`
// eingetragen. Die Marke ist in beiden Erscheinungsbildern hell genug für
// dunklen Text und dunkel genug für hellen; `marketext` ist die Farbe, die
// jeweils dazugehört, und deshalb braucht der Hauptknopf keine `dark:`-Zeile.
const gewichte: Record<Gewicht, string> = {
  haupt: `bg-marke text-marketext ${hoverHaupt}`,
  neben: `border border-kante bg-flaeche text-text ${hoverNeben}`,
  leise: `text-marke ${hoverNeben}`,
}

// **Ein gedrückter Umschalter ist ein Segment mit Rand, nie der Hauptknopf** —
// AP-3 (#2360),
// Befund B-08, Entwurf `Zustaende09`. Bis dahin bekam er die Füllung des
// Hauptknopfs, und bis zu vier gefüllte Flächen standen in einer Sicht:
// „Kacheln“ gedrückt sah aus wie „+ Neues Set“. Gewählt ist jetzt, was in
// `Segmente` gewählt ist — Rand in der Marke, Grund `betont`, `text-text`
// darauf (gemessen in `docs/KONTRASTE.md`).
const gewaehlt = 'border border-marke bg-betont text-text'

// **Ein ruhender Umschalter trägt nie das Hauptgewicht**, auch wenn der
// Aufrufer es verlangt — der gedrückte sähe sonst schwächer aus als der
// ungedrückte. Der leise behält seinen Platz für den Rand, damit das Drücken
// die Nachbarn nicht verschiebt.
const ungewaehlt: Record<Gewicht, string> = {
  haupt: gewichte.neben,
  neben: gewichte.neben,
  leise: `border border-transparent text-marke ${hoverNeben}`,
}

// stilFuer ist die Farbe eines Knopfs — eines Umschalters, wenn `gedrueckt`
// gesetzt ist.
function stilFuer(gewicht: Gewicht, gedrueckt: boolean | undefined): string {
  if (gedrueckt === undefined) {
    return gewichte[gewicht]
  }

  return gedrueckt ? gewaehlt : ungewaehlt[gewicht]
}

export function Knopf({
  gewicht = 'neben',
  angehalten = false,
  beschreibung,
  beschriebenVon,
  titel,
  bezeichnung,
  ausgeklappt,
  steuert,
  gedrueckt,
  aufKlick,
  aufDruck,
  griff,
  hatMenue,
  aufPfeil,
  ohneTabstelle,
  schreibt = false,
  klein = false,
  kuerzel,
  aktion,
  children,
}: KnopfEigenschaften) {
  // **Der eigene Griff, wo der Aufrufer keinen reicht** —
  // `useFokusUeberAnhalten` braucht den Knopf (#1990).
  const eigenerRef = useRef<HTMLButtonElement | null>(null)
  const knopfRef = griff ?? eigenerRef
  const merkeFokus = useFokusUeberAnhalten(knopfRef, angehalten)
  const sperre = useSchreibsperre(schreibt)

  // **Der Grund ergänzt die Beschreibung, statt sie zu ersetzen**: ein Knopf
  // aus dem Aktionskatalog behält seine Bedeutung, und der Titel allein käme
  // neben einer Beschreibung bei keinem Vorleseprogramm an.
  const beschrieben =
    [
      beschreibung?.['aria-describedby'] ?? beschriebenVon,
      sperre.knopf['aria-describedby'],
    ]
      .filter((teil) => teil !== undefined)
      .join(' ') || undefined

  return (
    <button
      ref={knopfRef}
      type="button"
      tabIndex={ohneTabstelle === true ? -1 : undefined}
      data-aktion={aktion}
      aria-keyshortcuts={kuerzel}
      aria-label={bezeichnung}
      title={beschreibung === undefined ? (sperre.titel ?? titel) : titel}
      aria-disabled={sperre.knopf['aria-disabled']}
      aria-describedby={beschrieben}
      onFocus={beschreibung?.onFocus}
      onBlur={beschreibung?.onBlur}
      aria-haspopup={hatMenue === true ? 'menu' : undefined}
      aria-expanded={ausgeklappt}
      aria-pressed={gedrueckt}
      aria-controls={ausgeklappt === true ? steuert : undefined}
      disabled={angehalten}
      onClick={
        aufKlick === undefined
          ? undefined
          : () => {
              if (sperre.gesperrt) return

              merkeFokus()
              aufKlick()
            }
      }
      onKeyDown={
        aufPfeil === undefined
          ? undefined
          : (ereignis) => {
              if (ereignis.key === 'ArrowDown' || ereignis.key === 'ArrowUp') {
                ereignis.preventDefault()
                aufPfeil(ereignis.key === 'ArrowDown' ? 0 : -1)
              }
            }
      }
      onMouseDown={
        aufDruck === undefined
          ? undefined
          : (ereignis) => {
              // **Der Fokus bleibt, wo er ist.** Ohne das nähme der Knopf ihn
              // dem Feld ab, und der Zeiger stünde nicht mehr dort, wo der
              // Vorschlag hin soll.
              ereignis.preventDefault()
              aufDruck()
            }
      }
      className={`${gemeinsam}${klein ? masse.klein : masse.gross}${stilFuer(gewicht, gedrueckt)}`}
    >
      {children}
      {sperre.satz}
    </button>
  )
}
