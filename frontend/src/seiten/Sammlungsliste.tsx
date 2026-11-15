import type { KeyboardEvent, MouseEvent } from 'react'

// **Die Liste der Sammlung in drei Ansichten** — Werkstück U8 von Phase 5.1
// (#1419): Tabelle, Gruppen mit Köpfen und Summen, Galerie mit Kacheln — alle
// über dieselbe Fensterliste (Vermerk 0085). Die Zeile ist die **Wahl** für
// den Inspector und trägt den **Haken** für die Massenvorgänge; bearbeitet
// wird nicht hier.
import type { ExemplarDTO } from '../api'
import { Fensterliste } from '../bausteine/Fensterliste'
import { Vorlaeufig } from '../bausteine/Vorlaeufig'
import { useWurzelschrift } from '../haken/useWurzelschrift'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { gruppiert, type Sicht } from './sammlungssicht'

interface ListenEigenschaften {
  exemplare: readonly ExemplarDTO[]
  sicht: Sicht
  gewaehlt: ReadonlySet<number>
  anhaken: (id: number, an: boolean) => void
  gehalten: number | undefined
  waehlen: (id: number) => void

  // aufFenster meldet, welche Zeilen gebraucht werden (#1500).
  aufFenster?: ((anfang: number, ende: number) => void) | undefined

  // anstossen nimmt Rechtsklick und ⇧F10 an einem Exemplar entgegen — AP-6,
  // Teil 2 (#2363): `useAnstoss` in `Sammlung` schlägt es zuerst auf.
  anstossen?:
    | ((
        ereignis: MouseEvent<HTMLElement> | KeyboardEvent<HTMLElement>,
        id: number,
      ) => void)
    | undefined
}

// zeilenhoehe ist die Höhe einer Tabellenzeile, kachelhoehe die einer
// Galeriereihe — fest, damit die Fensterliste rechnen kann.
// **Bei 16 Punkten Wurzelschrift** — gerechnet wird mit der wirksamen
// (Zweitblick über Phase 6 von 0.9.0, zweiter Teil, Befund 5): schmal trägt
// eine Zeile zwei Textzeilen, und bei „sehr groß“ schnitten 40 Punkte sie ab.
const zeilenhoeheBei16 = 40
const kachelhoeheBei16 = 96

// breiten gelten für Kopf und Zelle zugleich. **Schmal**, unter 42rem Breite
// der Liste, fallen Sprache, Set und Lagerort aus ihren Spalten und stehen
// unter der Bezeichnung (AP-12, #2369).
const breiten = {
  wahl: 'w-8 shrink-0',
  bezeichnung: 'min-w-0 flex-1',
  menge: 'w-28 shrink-0',
  sprache: 'hidden w-20 shrink-0 @2xl:block',
  set: 'hidden w-12 shrink-0 @2xl:block',
  lagerort: 'hidden w-24 shrink-0 @2xl:block',
}

// zellen sind die Klassen der Zellen — ihre Breite und was sie sonst tragen.
const leise = 'truncate text-leise '
const zellen = {
  bezeichnung:
    'flex flex-col justify-center font-medium ' + breiten.bezeichnung,
  sprache: leise + breiten.sprache,
  set: leise + breiten.set,
  lagerort: leise + breiten.lagerort,
}
const kachelnJeReihe = 4

type Listenzeile =
  | { art: 'kopf'; schluessel: string; anzahl: number; menge: number }
  | { art: 'zeile'; exemplar: ExemplarDTO }

interface Kachelreihe {
  schluessel: string
  exemplare: readonly ExemplarDTO[]
}

export function Sammlungsliste({
  exemplare,
  sicht,
  gewaehlt,
  anhaken,
  gehalten,
  waehlen,
  aufFenster,
  anstossen,
}: ListenEigenschaften) {
  const eigene = useTexte().seiten.sammlung
  const schrift = useWurzelschrift()
  const zeilenhoehe = Math.ceil((zeilenhoeheBei16 * schrift) / 16)
  const kachelhoehe = Math.ceil((kachelhoeheBei16 * schrift) / 16)
  const spalten = [
    { titel: eigene.spalteWahl, klasse: breiten.wahl },
    { titel: eigene.spalteBezeichnung, klasse: breiten.bezeichnung },
    { titel: eigene.spalteMenge, klasse: breiten.menge },
    { titel: eigene.spalteSprache, klasse: breiten.sprache },
    { titel: eigene.spalteSet, klasse: breiten.set },
    { titel: eigene.spalteLagerort, klasse: breiten.lagerort },
  ]

  // darunter ist, was schmal unter der Bezeichnung steht — ohne leere Teile.
  const darunter = (exemplar: ExemplarDTO) =>
    [
      exemplar.sprache,
      eigene[exemplar.finish],
      exemplar.setcode,
      exemplar.lagerort,
    ]
      .filter((teil) => teil !== '')
      .join(' · ')

  const zelle = (exemplar: ExemplarDTO) => (
    <>
      <div role="gridcell" className={breiten.wahl}>
        <input
          type="checkbox"
          // **Innerhalb des Gitters erreichbar, nicht über den Tabulator**
          // (#2249) — mit → und ←; sonst war jede Zeile ein eigener Halt.
          tabIndex={-1}
          checked={gewaehlt.has(exemplar.id)}
          aria-label={exemplar.bezeichnung}
          onChange={(ereignis) => {
            anhaken(exemplar.id, ereignis.target.checked)
          }}
          onClick={(ereignis) => {
            // **Der Haken ist kein Klick auf die Zeile**: er wählt nichts für
            // den Inspector — sonst stünden zwei »Löschen« nebeneinander.
            ereignis.stopPropagation()
          }}
          className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </div>
      {/* **Nur die Bezeichnung, nicht die Kennung**: die Sichtprüfung zeigte
          »Bi…«, weil die Kennung daneben den Platz nahm — sie steht im
          Inspector. Der volle Name hängt als `title` an der Zelle.

          **Schmal stehen Sprache, Set und Lagerort darunter** (AP-12, #2369,
          B-28): zwischen Navigator und Inspector blieben der Bezeichnung
          neben fünf festen Spalten keine 60 Punkte — „Counters…“, und
          umgebrochen liefen die Zeilen in ihre Nachbarn, denn eine Zeile ist
          40 Punkte hoch. Zwei Zeilen zu 20 und 16 passen hinein. */}
      <div
        role="gridcell"
        className={zellen.bezeichnung}
        title={exemplar.bezeichnung}
      >
        <span className="flex min-w-0 items-center gap-2">
          <span className="truncate">{exemplar.bezeichnung}</span>
          <Vorlaeufig vorlaeufig={exemplar.vorlaeufig} />
        </span>
        <span
          className="truncate text-xs/4 font-normal text-leise @2xl:hidden"
          title={darunter(exemplar)}
        >
          {darunter(exemplar)}
        </span>
      </div>
      <div role="gridcell" className={breiten.menge}>
        {exemplar.menge} × {eigene.zustaende[zustandsname(exemplar.zustand)]}
      </div>
      <div role="gridcell" className={zellen.sprache}>
        {exemplar.sprache} · {eigene[exemplar.finish]}
      </div>
      <div role="gridcell" className={zellen.set}>
        {exemplar.setcode}
      </div>
      <div role="gridcell" className={zellen.lagerort}>
        {exemplar.lagerort}
      </div>
    </>
  )

  if (sicht.ansicht === 'galerie') {
    const reihen: Kachelreihe[] = []
    for (let anfang = 0; anfang < exemplare.length; anfang += kachelnJeReihe) {
      const stueck = exemplare.slice(anfang, anfang + kachelnJeReihe)
      // **Die Stelle als Schlüssel**: eine Reihe hat keine Identität, ihre
      // Kacheln haben eine.
      reihen.push({ schluessel: String(anfang), exemplare: stueck })
    }

    return (
      <Fensterliste
        bezeichnung={eigene.listeBezeichnung}
        eintraege={reihen}
        zeilenhoehe={kachelhoehe}
        aufFenster={aufFenster}
        schluessel={(reihe) => reihe.schluessel}
        spalten={[eigene.listeBezeichnung]}
        zeichne={(reihe) => (
          <div role="gridcell" className="flex w-full gap-2">
            {reihe.exemplare.map((exemplar) => (
              <button
                key={exemplar.id}
                type="button"
                aria-label={einsetzen(eigene.kachelWaehlen, {
                  bezeichnung: exemplar.bezeichnung,
                })}
                aria-pressed={exemplar.id === gehalten}
                onClick={() => {
                  waehlen(exemplar.id)
                }}
                onContextMenu={(ereignis) => {
                  anstossen?.(ereignis, exemplar.id)
                }}
                onKeyDown={(ereignis) => {
                  anstossen?.(ereignis, exemplar.id)
                }}
                className="flex h-20 flex-1 flex-col items-start justify-between rounded-md border border-kante p-2 text-left hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-pressed:bg-betont"
              >
                <span className="flex items-center gap-2 truncate font-medium">
                  {exemplar.bezeichnung}
                  <Vorlaeufig vorlaeufig={exemplar.vorlaeufig} />
                </span>
                <span className="text-xs text-leise">
                  {exemplar.menge} ×{' '}
                  {eigene.zustaende[zustandsname(exemplar.zustand)]}
                </span>
              </button>
            ))}
          </div>
        )}
      />
    )
  }

  if (sicht.ansicht === 'gruppen') {
    const zeilen: Listenzeile[] = gruppiert(
      exemplare,
      sicht.gruppierung,
    ).flatMap((abschnitt) => [
      {
        art: 'kopf' as const,
        schluessel: abschnitt.schluessel,
        anzahl: abschnitt.exemplare.length,
        menge: abschnitt.menge,
      },
      ...abschnitt.exemplare.map((exemplar) => ({
        art: 'zeile' as const,
        exemplar,
      })),
    ])

    return (
      <Fensterliste
        bezeichnung={eigene.listeBezeichnung}
        eintraege={zeilen}
        zeilenhoehe={zeilenhoehe}
        aufFenster={aufFenster}
        schluessel={(zeile) =>
          zeile.art === 'kopf' ? `kopf:${zeile.schluessel}` : zeile.exemplar.id
        }
        spalten={spalten}
        zeichne={(zeile) =>
          zeile.art === 'kopf' ? (
            <div
              role="rowheader"
              className="flex flex-1 items-baseline gap-3 font-medium"
            >
              <span>
                {zeile.schluessel === '' ? eigene.gruppeOhne : zeile.schluessel}
              </span>
              <span className="text-sm font-normal text-leise">
                {einsetzen(eigene.gruppeSumme, {
                  anzahl: String(zeile.anzahl),
                  menge: String(zeile.menge),
                })}
              </span>
            </div>
          ) : (
            zelle(zeile.exemplar)
          )
        }
        aufKlick={(zeile) => {
          if (zeile.art === 'zeile') {
            waehlen(zeile.exemplar.id)
          }
        }}
        // **Auch die Tastatur wählt** — wie in der Tabelle; hier fehlte es,
        // und ein Exemplar ließ sich nur mit der Maus aufschlagen (Zweitblick
        // über Phase 6.3, G9). Ein Kopf wird gehalten, aber nicht gewählt.
        aufHalt={(zeile) => {
          if (zeile.art === 'zeile') {
            waehlen(zeile.exemplar.id)
          }
        }}
        ausgewaehlt={(zeile) =>
          zeile.art === 'zeile' && zeile.exemplar.id === gehalten
        }
        // **Ein Kopf hat kein Menü** — er ist kein Exemplar.
        aufAnstoss={(ereignis, zeile) => {
          if (zeile.art === 'zeile') {
            anstossen?.(ereignis, zeile.exemplar.id)
          }
        }}
      />
    )
  }

  return (
    <Fensterliste
      bezeichnung={eigene.listeBezeichnung}
      eintraege={exemplare}
      zeilenhoehe={zeilenhoehe}
      aufFenster={aufFenster}
      // **Die Stelle und nicht die Kennung** (#1500): eine nachgeholte Zeile
      // ersetzt einen Platzhalter — mit der Kennung als Schlüssel tauschte
      // React den Knoten aus, und der Fokus ginge mit ihm verloren.
      schluessel={(_exemplar, stelle) => stelle}
      spalten={spalten}
      zeichne={zelle}
      aufKlick={(exemplar) => {
        waehlen(exemplar.id)
      }}
      aufHalt={(exemplar) => {
        waehlen(exemplar.id)
      }}
      ausgewaehlt={(exemplar) => exemplar.id === gehalten}
      aufAnstoss={(ereignis, exemplar) => {
        anstossen?.(ereignis, exemplar.id)
      }}
    />
  )
}

function zustandsname(zustand: string): keyof typeof namen {
  return zustand in namen ? (zustand as keyof typeof namen) : 'nearMint'
}

const namen = {
  mint: true,
  nearMint: true,
  excellent: true,
  good: true,
  lightPlayed: true,
  played: true,
  poor: true,
} as const
