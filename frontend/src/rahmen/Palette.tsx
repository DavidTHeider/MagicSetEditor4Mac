// Die Befehlspalette (⌘K).
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** — der Schnitt
// aus AGENTS.md §2. Die Auswahl selbst ist in `passende` ausgelagert und ohne
// DOM prüfbar.
import { useId, useRef, useState, type KeyboardEvent } from 'react'

import { api } from '../api'
import { useAnschluss } from '../anschlussraum'
import { Dialog } from '../bausteine/Dialog'
import { useLaden } from '../haken/useLaden'
import { useNurLesend } from '../haken/schreibrecht'
import { usePalettendaten } from '../haken/usePalettendaten'
import { auspacken } from '../umschlag'
import {
  findziele,
  Ohneeingabe,
  passendeZiele,
  tunziele,
  type Palettenwege,
  type Palettenziel,
} from './palettenziele'

import { beschriftung, seiten, type Seitenschluessel } from '../navigation'
import type { Katalog } from '../sprache/kataloge'
import { useTexte } from '../sprache/kontext'
import { useVerborgeneSeiten } from '../abschaltung'
import type { BefehlDTO, BefehlskatalogAntwort } from '../apibefehle'
import { schluesselvergleich } from '../ordnung'

// passende filtert die Seiten nach dem, was jemand tippt.
//
// **Ohne Beachtung der Groß- und Kleinschreibung und ohne Anfangsbindung**: wer
// „stell" tippt, meint die Einstellungen. Eine Suche, die nur am Wortanfang
// greift, zwingt zum Raten, wie der Eintrag beginnt.
//
// **Und ohne, was abgeschaltet ist** (OPS-10, #1908).
export function passende(
  texte: Katalog,
  eingabe: string,
  verborgen: ReadonlySet<Seitenschluessel> = new Set(),
): Seitenschluessel[] {
  const gesucht = eingabe.trim().toLowerCase()

  return seiten
    .filter(
      (seite) =>
        !verborgen.has(seite.schluessel) &&
        (gesucht === '' ||
          beschriftung(texte, seite.schluessel)
            .toLowerCase()
            .includes(gesucht)),
    )
    .map((seite) => seite.schluessel)
}

// passendeBefehle filtert die Befehle der Kommandozentrale — über Name,
// Fähigkeit und Erklärung, höchstens dreißig (#1423).
export function passendeBefehle(
  befehle: BefehlDTO[],
  eingabe: string,
): BefehlDTO[] {
  const gesucht = eingabe.trim().toLowerCase()
  if (gesucht === '') {
    return []
  }

  // **Der Name zuerst**: wer `lesen.sets` tippt, meint `lesen.sets` — nicht
  // die dreißig Befehle, deren Erklärung „Sets" sagt.
  const rang = (befehl: BefehlDTO): number => {
    const name = befehl.name.toLowerCase()
    if (name === gesucht) {
      return 0
    }
    if (name.startsWith(gesucht)) {
      return 1
    }
    if (name.includes(gesucht)) {
      return 2
    }
    if (
      `${befehl.faehigkeit} ${befehl.erklaerung}`
        .toLowerCase()
        .includes(gesucht)
    ) {
      return 3
    }

    return -1
  }

  return (
    befehle
      .map((befehl) => ({ befehl, rang: rang(befehl) }))
      .filter((eintrag) => eintrag.rang >= 0)
      // **Nach Codepunkten**: `lesen.sets` ist eine Kennung, kein Menschentext
      // — ihre Ordnung ist auf jedem Rechner dieselbe (F01, `ordnung.ts`).
      .sort(
        (a, b) =>
          a.rang - b.rang || schluesselvergleich(a.befehl.name, b.befehl.name),
      )
      .map((eintrag) => eintrag.befehl)
      .slice(0, 30)
  )
}

// zielklasse und zeilenklasse sind die Zeilen aus Tun und Finden und die der
// Seiten und Befehle — die vorgewählte bekommt dazu `bg-betont` (G07).
const zielklasse =
  'flex w-full items-baseline gap-2 px-4 py-2 text-left hover:bg-betont focus-visible:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring'
const zeilenklasse =
  'w-full px-4 py-2 text-left hover:bg-betont focus-visible:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring'

// abschnittskopf ist die Zeile über einem Abschnitt — *Tun*, *Finden*,
// *Seiten*, *Befehle* —, wie `palette-gruppe` im Artboard *Befehle*.
const abschnittskopf = 'px-4 pt-2 text-xs font-semibold uppercase text-leise'

async function katalogHolen(): Promise<BefehlskatalogAntwort> {
  const antwort = await api().Befehlskatalog()

  return auspacken(antwort, () => antwort)
}

interface PalettenEigenschaften {
  schliessen: () => void
  waehlen: (seite: Seitenschluessel) => void

  // waehleBefehl öffnet die Kommandozentrale über einem Befehl (#1423).
  waehleBefehl: (befehl: BefehlDTO) => void

  // wege machen aus der Palette mehr als ein Seitenverzeichnis: Handlungen
  // und Objekte (UX-02.AK-02, #1674). **Ohne sie** — wie in den Läufen, die
  // nur Seiten und Befehle prüfen — fragt sie den Bestand gar nicht.
  wege?: Pick<Palettenwege, 'setID' | 'waehleSet' | 'oeffneKarte'> | undefined
}

// Palette wird **nur gezeigt, wenn sie offen ist** — sie kennt kein `offen`.
//
// Der erste Anlauf hatte eines und gab `null` zurück, solange es falsch war. Das
// blendet aus, baut aber nicht ab: der Zustand bleibt stehen, und die Palette
// zeigte beim nächsten Öffnen die alte Sucheingabe. Wer sie abbaut, hat das
// Problem nicht — ein abgebauter Verbund hat nichts, was er sich merken könnte.
//
// Der erste Anlauf leerte statt dessen im Effekt. Das ist eine Zustandsänderung
// als Folge einer Zustandsänderung, und React rät zu Recht davon ab.
export function Palette({
  schliessen,
  waehlen,
  waehleBefehl,
  wege,
}: PalettenEigenschaften) {
  const texte = useTexte()
  const [eingabe, setzeEingabe] = useState('')
  const listenId = useId()
  const rahmen = useRef<HTMLDivElement>(null)
  const anschluss = useAnschluss()

  // **Ein Leser bekommt nur, was er tun darf** (G07, #2254): seine
  // Verbindung liest nur, und eine Handlung, die schreibt, scheiterte erst
  // an der Bindung — dieselbe Klasse wie die Menüs (G10, #2288).
  const lesend = useNurLesend()

  // **Tun und Finden** — was geht, und was es gibt (UX-02.AK-02).
  const daten = usePalettendaten(eingabe, wege !== undefined)
  const volleWege: Palettenwege | undefined =
    wege === undefined
      ? undefined
      : { ...wege, uebergib: anschluss.uebergib, wechseln: waehlen }
  const alleTun =
    volleWege === undefined ? [] : tunziele(daten, volleWege, texte, lesend)
  const tun =
    eingabe.trim() === ''
      ? alleTun.slice(0, Ohneeingabe)
      : passendeZiele(alleTun, eingabe)
  const finden =
    volleWege === undefined || eingabe.trim() === ''
      ? []
      : passendeZiele(findziele(daten, volleWege, texte), eingabe)

  // **Die Pfeile wandern**, wie der Fuß es sagt: vom Feld in die erste
  // Zeile und zurück. Tab geht weiter wie bisher (ux07-palette).
  const wandern = (ereignis: KeyboardEvent<HTMLElement>) => {
    const richtung =
      ereignis.key === 'ArrowDown' ? 1 : ereignis.key === 'ArrowUp' ? -1 : 0
    if (richtung === 0 || rahmen.current === null) return

    const halte = [
      ...rahmen.current.querySelectorAll<HTMLElement>(
        'input, [data-palettenzeile]',
      ),
    ]
    const naechste = halte[halte.indexOf(ereignis.currentTarget) + richtung]
    if (naechste === undefined) return

    ereignis.preventDefault()
    naechste.focus()
  }

  // zeile zeichnet einen Eintrag aus Tun oder Finden.
  const zeile = (ziel: Palettenziel) => (
    <li key={ziel.schluessel}>
      <button
        type="button"
        data-palettenzeile=""
        {...markiert(ziel.schluessel)}
        onKeyDown={wandern}
        onClick={() => {
          schliessen()
          ziel.tun()
        }}
        className={hervor(zielklasse, ziel.schluessel)}
      >
        <span>{ziel.text}</span>
        <span className="ml-auto shrink-0 text-xs text-leise">{ziel.wo}</span>
      </button>
    </li>
  )

  const abschnitt = (titel: string, ziele: Palettenziel[]) =>
    ziele.length > 0 && (
      <>
        <p className={abschnittskopf}>{titel}</p>
        <ul aria-label={titel}>{ziele.map(zeile)}</ul>
      </>
    )

  // **Die Befehle kommen, sobald die Palette steht** — ein Fehler dabei
  // nimmt den Seiten nichts: dann gibt es nur sie, wie vor U12.
  const katalog = useLaden(katalogHolen, [])
  const befehle = passendeBefehle(
    (katalog.daten?.befehle ?? []).filter(
      (befehl) => !lesend || !befehl.schreibend,
    ),
    eingabe,
  )

  const treffer = passende(texte, eingabe, useVerborgeneSeiten())

  // **Die oberste Zeile ist vorgewählt, sobald etwas getippt ist** (G07,
  // #2254) — ⏎ im Feld führt sie aus, wie der Fuß es sagt. Ohne Eingabe
  // gibt es keine Absicht und keine Vorwahl.
  const vorwahl =
    eingabe.trim() === ''
      ? undefined
      : [
          ...tun.map((ziel) => ziel.schluessel),
          ...finden.map((ziel) => ziel.schluessel),
          ...treffer.map((schluessel) => `seite:${schluessel}`),
          ...befehle.map((befehl) => `befehl:${befehl.name}`),
        ][0]
  const markiert = (schluessel: string) =>
    schluessel === vorwahl ? { 'data-vorwahl': '' } : {}
  const hervor = (klasse: string, schluessel: string) =>
    schluessel === vorwahl ? `${klasse} bg-betont` : klasse

  // ausfuehren ist ⏎ im Feld. **Nicht während einer Eingabemethode** — bei
  // Japanisch oder Chinesisch bestätigt ⏎ dort erst die Zeichenwahl.
  const ausfuehren = (ereignis: KeyboardEvent<HTMLInputElement>) => {
    if (ereignis.key !== 'Enter' || ereignis.nativeEvent.isComposing) {
      wandern(ereignis)

      return
    }

    ereignis.preventDefault()
    rahmen.current?.querySelector<HTMLElement>('[data-vorwahl]')?.click()
  }

  // **Fokus und Escape gehören dem Dialog** — seit Werkstück U2 von Phase
  // 5.1 (#1413): er setzt den Fokus auf das erste Bedienelement, horcht
  // Escape am Fenster und gibt den Fokus beim Schließen zurück.

  return (
    <Dialog
      bezeichnung={texte.allgemein.palette.bezeichnung}
      schliessen={schliessen}
    >
      <div
        ref={rahmen}
        // **Eine Ebene über der Seite** — AP-4 (#2361), Festlegung 11.
        className="w-full max-w-md overflow-hidden rounded-lg border-2 border-kante bg-flaeche shadow-ueberlage contrast-more:shadow-none"
      >
        <input
          type="text"
          value={eingabe}
          aria-label={texte.allgemein.palette.eingabe}
          aria-controls={listenId}
          placeholder={texte.allgemein.palette.befehlEingabe}
          onChange={(ereignis) => {
            setzeEingabe(ereignis.target.value)
          }}
          onKeyDown={ausfuehren}
          // **Der Ring steht auch hier** (#828). Das Feld hat den Fokus,
          // sobald die Palette aufgeht — aber „hat ihn ohnehin" ist keine
          // Begründung dafür, ihn nicht zu zeigen: wer mit der Tastatur
          // hineinkommt, sieht sonst nicht, wo er ist.
          className="w-full border-b border-kante bg-transparent px-4 py-3 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring"
        />
        {treffer.length === 0 &&
        befehle.length === 0 &&
        tun.length === 0 &&
        finden.length === 0 ? (
          <p className="px-4 py-3 text-leise">
            {texte.allgemein.palette.keineTreffer}
          </p>
        ) : (
          <div id={listenId} className="max-h-96 overflow-y-auto py-1">
            {abschnitt(texte.allgemein.palette.tun, tun)}
            {abschnitt(texte.allgemein.palette.finden, finden)}
            {/* **Auch die Seiten tragen ihre Überschrift** — UXR-02
                (#1910): ohne sie lasen sie sich als Fortsetzung von *Tun*. */}
            {treffer.length > 0 && (
              <>
                <p className={abschnittskopf}>
                  {texte.allgemein.palette.seiten}
                </p>
                <ul aria-label={texte.allgemein.palette.seiten}>
                  {treffer.map((schluessel) => (
                    <li key={schluessel}>
                      <button
                        type="button"
                        data-palettenzeile=""
                        data-seite={schluessel}
                        {...markiert(`seite:${schluessel}`)}
                        onKeyDown={wandern}
                        onClick={() => {
                          waehlen(schluessel)
                        }}
                        className={hervor(zeilenklasse, `seite:${schluessel}`)}
                      >
                        {beschriftung(texte, schluessel)}
                      </button>
                    </li>
                  ))}
                </ul>
              </>
            )}
            {befehle.length > 0 && (
              <>
                <p className={abschnittskopf}>
                  {texte.allgemein.palette.befehle}
                </p>
                <ul aria-label={texte.allgemein.palette.befehle}>
                  {befehle.map((befehl) => (
                    <li key={befehl.name}>
                      <button
                        type="button"
                        data-palettenzeile=""
                        {...markiert(`befehl:${befehl.name}`)}
                        onKeyDown={wandern}
                        onClick={() => {
                          waehleBefehl(befehl)
                        }}
                        className={hervor(
                          zeilenklasse,
                          `befehl:${befehl.name}`,
                        )}
                      >
                        {/* **Der lesbare Name zuerst** (AP-11, #2368, B-27),
                            der Befehlsname klein dahinter; ohne Fähigkeit
                            bleibt er allein. */}
                        {befehl.faehigkeit === '' ? (
                          <code className="text-sm">{befehl.name}</code>
                        ) : (
                          <>
                            <span className="text-sm">{befehl.faehigkeit}</span>
                            <code className="ml-2 text-xs text-leise">
                              {befehl.name}
                            </code>
                          </>
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              </>
            )}
          </div>
        )}

        {/* **Wie man sich bewegt, steht darunter** — wie im Artboard
         *Befehle*. */}
        <p className="border-t border-kante px-4 py-2 text-xs text-leise">
          {texte.allgemein.palette.fuss}
        </p>
      </div>
    </Dialog>
  )
}
