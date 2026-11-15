// **Die Liste eines Menüs** — Werkstück U2 von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Warum ein eigener Baustein
//
// `Menue` hängt an einem Knopf, `Kontextmenue` an einem Rechtsklick — was
// **danach** aufgeht, ist dasselbe: `role="menu"`, Einträge als
// `menuitem`, ein rovierender Tabindex, ein ruhender Eintrag mit
// `aria-disabled`. Zweimal geschrieben liefe es auseinander, und ein
// Vorleseprogramm hörte zwei verschiedene Menüs, wo eines gemeint ist.
//
// **Die Tasten deutet der Aufrufer.** Wer das Menü öffnet, weiß, wohin der
// Fokus beim Schließen soll; die Liste weiß es nicht.
import { objektkuerzel, type Objektkuerzel } from '../objekthandlungen'
import {
  useEffect,
  useRef,
  type CSSProperties,
  type KeyboardEvent,
} from 'react'

import { aktionsarten, type Aktion } from '../aktionen'
import { useSchreibsperre } from '../haken/schreibrecht'

// Menueeintrag ist eine Zeile des Menüs.
//
// **Ein Verbund, kein Feld mit Vorbehalt**: ein Eintrag, der ruht, weil es
// nichts zu tun gibt („nach oben" auf der ersten Karte), hat keine Handlung —
// und einer, der ruht, weil sein Vorgang läuft (§7, Regel 4), hat eine, die
// gerade nicht gilt. Beide stehen im Menü; **er verschwindet nicht**, denn
// was fehlt, kann niemand vermissen.
export type Menueeintrag = {
  schluessel: string

  // name steht im Menü — aus dem Katalog des Aufrufers (§9).
  name: string

  // aktion ist die Katalogaktion dieser Zeile — sie trägt dann dasselbe
  // Symbol wie der Knopf derselben Aktion anderswo (UX-10.AK-01, #1682).
  aktion?: Aktion

  // mitKuerzel sagt, dass der Aufrufer das Kürzel der Aktion **verdrahtet**
  // hat — dieselbe Zusage wie am `Aktionsknopf`. Die Zeile nennt es dann
  // rechts, wie jedes Menü von macOS, und über `aria-keyshortcuts` (UXR-06,
  // #1914: Kopieren zog aus einem Knopf ins Menü „Diese Karte").
  mitKuerzel?: boolean

  // kuerzel ist das Kürzel eines Objektmenüs — AP-7 (#2364). Die Taste
  // verdrahtet die Leiste (`ObjekthandlungenMelden`); die Zeile nennt sie.
  kuerzel?: Objektkuerzel

  // schreibt heißt: diese Zeile ändert etwas im Profil — #2288, dieselbe
  // Zusage wie `schreibt` am `Knopf` (G10). Für einen Leser steht sie
  // gesperrt und erklärt da, **fokussierbar**, und ihre Handlung läuft nicht.
  schreibt?: boolean
} & ({ aufWahl: () => void; angehalten?: boolean } | { angehalten: true })

// naechsteStelle rechnet, wohin ein Pfeil führt — zyklisch, über ruhende
// Einträge hinweg.
//
// **Als eigene Funktion**, damit sie ohne DOM prüfbar ist: eine Tabelle aus
// Stelle, Taste und Ziel ist zu lesen, ein Test über `keyDown` nicht.
export function naechsteStelle(
  eintraege: readonly Menueeintrag[],
  stelle: number,
  taste: string,
): number {
  const anzahl = eintraege.length

  const bedienbar = (bei: number) => eintraege[bei]?.angehalten !== true

  const suche = (von: number, richtung: 1 | -1): number => {
    for (let schritt = 1; schritt <= anzahl; schritt += 1) {
      const kandidat = (((von + richtung * schritt) % anzahl) + anzahl) % anzahl
      if (bedienbar(kandidat)) {
        return kandidat
      }
    }

    return stelle
  }

  switch (taste) {
    case 'ArrowDown':
      return suche(stelle, 1)
    case 'ArrowUp':
      return suche(stelle, -1)
    case 'Home':
      return suche(-1, 1)
    case 'End':
      return suche(anzahl, -1)
    default:
      return stelle
  }
}

interface MenuelisteEigenschaften {
  id: string
  bezeichnung: string
  eintraege: readonly Menueeintrag[]

  // stelle ist der Eintrag, der den Fokus hat — der Aufrufer hält sie, weil
  // er sie beim Öffnen setzt.
  stelle: number

  aufTaste: (ereignis: KeyboardEvent<HTMLUListElement>) => void

  // aufWahl läuft, bevor der Eintrag seine Arbeit tut — der Aufrufer schließt
  // damit das Menü und holt den Fokus.
  aufWahl: () => void

  // lage setzt das Menü an eine Stelle im Kasten — das Kontextmenü an den
  // Zeiger. Ohne Angabe steht es unter seinem Knopf.
  lage?: CSSProperties
}

export function Menueliste({
  id,
  bezeichnung,
  eintraege,
  stelle,
  aufTaste,
  aufWahl,
  lage,
}: MenuelisteEigenschaften) {
  const liste = useRef<HTMLUListElement>(null)

  // **Ein Satz für alle gesperrten Zeilen** — die Sperre hängt am Raum, nicht
  // an der Zeile, und ein Grund je Zeile sagte dasselbe mehrmals.
  const sperre = useSchreibsperre(true)
  const gesperrt = (eintrag: Menueeintrag) =>
    sperre.gesperrt && eintrag.schreibt === true

  // **Der Fokus folgt der Stelle.** Ein Menü, in dem der Fokus auf dem Knopf
  // bleibt, liest ein Vorleseprogramm nicht vor.
  useEffect(() => {
    liste.current
      ?.querySelector<HTMLElement>(`[data-stelle="${String(stelle)}"]`)
      ?.focus()
  }, [stelle])

  return (
    <>
      {/* **Neben der Liste und nicht darin**: ein `ul` trägt nur Zeilen. */}
      {sperre.satz}
      <ul
        ref={liste}
        id={id}
        role="menu"
        aria-label={bezeichnung}
        onKeyDown={aufTaste}
        style={lage}
        // **Eine Ebene über der Seite, und sie taucht auf** — AP-4 (#2361),
        // Vermerk 0065 Festlegungen 10 und 11. **Ein langes Menü rollt in
        // sich**, statt über den Fensterrand zu laufen (Zweitblick über
        // Phase 6 von 0.9.0, Befund 5).
        className="auftauchen absolute z-40 max-h-[calc(100vh-1rem)] min-w-48 overflow-y-auto rounded-md border-2 border-kante bg-flaeche py-1 shadow-ueberlage contrast-more:shadow-none"
      >
        {eintraege.map((eintrag, bei) => (
          <li key={eintrag.schluessel} role="none">
            <button
              type="button"
              role="menuitem"
              data-stelle={bei}
              tabIndex={bei === stelle ? 0 : -1}
              aria-disabled={
                eintrag.angehalten === true || gesperrt(eintrag)
                  ? true
                  : undefined
              }
              aria-describedby={
                gesperrt(eintrag) ? sperre.knopf['aria-describedby'] : undefined
              }
              title={gesperrt(eintrag) ? sperre.titel : undefined}
              aria-keyshortcuts={kuerzelVon(eintrag)?.tasten}
              onClick={() => {
                if (eintrag.angehalten === true || gesperrt(eintrag)) {
                  return
                }

                aufWahl()
                eintrag.aufWahl()
              }}
              className="w-full px-3 py-1.5 text-left hover:bg-betont focus-visible:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring aria-disabled:opacity-50"
            >
              {eintrag.aktion !== undefined && (
                <span aria-hidden="true" className="mr-2 inline-block w-4">
                  {aktionsarten[eintrag.aktion].symbol}
                </span>
              )}
              {eintrag.name}
              <Kuerzel eintrag={eintrag} />
            </button>
          </li>
        ))}
      </ul>
    </>
  )
}

// Kuerzel zeigt das Zeichen der Tasten rechts in der Zeile.
//
// **Das Zeichen ist stumm** — die Tasten nennt `aria-keyshortcuts`, und ein
// Vorleseprogramm läse sonst „Kopieren, Befehl C" zweimal.
function Kuerzel({ eintrag }: { eintrag: Menueeintrag }) {
  const kuerzel = kuerzelVon(eintrag)

  if (kuerzel === undefined) {
    return null
  }

  return (
    <span aria-hidden="true" className="float-right ml-6 text-leise">
      {kuerzel.zeichen}
    </span>
  )
}

// kuerzelVon gibt das Kürzel einer Zeile — nur, wenn sie eine Aktion trägt
// **und** ihr Aufrufer es verdrahtet hat.
function kuerzelVon(
  eintrag: Menueeintrag,
): { tasten: string; zeichen: string } | undefined {
  if (eintrag.kuerzel !== undefined) {
    return objektkuerzel[eintrag.kuerzel]
  }

  return eintrag.mitKuerzel === true && eintrag.aktion !== undefined
    ? aktionsarten[eintrag.aktion]
    : undefined
}
