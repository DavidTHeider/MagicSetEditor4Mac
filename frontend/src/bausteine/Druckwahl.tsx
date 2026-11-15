// **Drucke aus der eigenen Sammlung wählen** — UXR-F18
// (#2033).
//
// # Wofür
//
// Eine Sichtfreigabe mit dem Umfang „Ausgewählte Drucke" benennt die Drucke,
// die ein Gegenüber sieht (`Sammlungsfreigabe.Auswahl`, `Zeigt` vergleicht
// `Printing`). Bis dahin bot das Fenster den Umfang an und ließ keinen Druck
// wählen — die Domäne weist eine Auswahl ohne Drucke ab, jede Freigabe auf
// diesem Umfang scheiterte. An der Kommandozeile stehen die Drucke hinter
// dem Umfang (`mse kontakt-freigeben <kennung> <zweck> auswahl druck…`).
//
// # Das Bedienelement (R7 in `docs/UIREVIEW.md`)
//
// **Bis vierzehn Drucke eine Liste mit Kästchen, ab fünfzehn ein Suchfeld
// davor.** Mit Suchfeld zeichnet die Liste nie mehr als die Wahl und höchstens
// zwölf Treffer (R6): eine Sammlung mit tausend Drucken stellt keine tausend
// Kästchen. Ohne Suchwort steht nur die Wahl da — wer noch nichts gewählt hat,
// sieht einen Satz, der sagt, wie es weitergeht.
//
// **Die Wahl bleibt stehen, auch wenn die Suche sie nicht trifft.** Sonst fiele
// sie aus dem Bild, und niemand könnte sie zurücknehmen (wie die gewählte
// Karte in `Deckkartenwahl`).
//
// # Warum eine Zeile je Druck und nicht je Exemplar
//
// **Weil die Freigabe den Druck benennt.** Zwei Exemplare desselben Drucks
// — in zwei Ausführungen, in zwei Zuständen — sind für das Gegenüber ein
// Eintrag, und die Menge davor ist ihre Summe. Ein Wunsch ist kein Besitz
// (`Exemplar.Besitzt`, 0057) und steht hier nicht.
//
// # Warum der Befund hier steht und nicht in `Feld`
//
// **Weil `Feld` ein einzelnes Bedienelement beschriftet.** Die Wahl ist eine
// Gruppe aus Kästchen, und ein Suchfeld gibt es erst ab fünfzehn Drucken. Der
// Befund steht deshalb an der Gruppe: als `role="alert"` im `fieldset`, das
// ihn in seiner Beschreibung führt — dieselbe Zusage wie in `Feld` (§6).
import { useId, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useSprache, useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'

import type { ExemplarDTO } from '../api'
import { falte } from '../sprache/vergleich'

// Druck ist eine Zeile der Wahl: ein Druck, den jemand besitzt.
interface Druck {
  printing: string
  bezeichnung: string

  // menge ist die Summe über alle Exemplare dieses Drucks.
  menge: number
}

interface Druckwahleigenschaften {
  // gewaehlt sind die Kennungen der gewählten Drucke — in der Reihenfolge der
  // Wahl.
  gewaehlt: readonly string[]
  waehlen: (neu: string[]) => void

  // befund ist der Satz, der an der Wahl steht: warum sie so nicht gilt.
  befund?: string | undefined

  // angehalten hält jedes Bedienelement an, solange ein Vorgang läuft (§7,
  // Regel 4).
  angehalten?: boolean
}

// sucheAb ist R7: ab so vielen Werten steht ein Suchfeld vor der Auswahl.
const sucheAb = 15

// trefferMax ist, wie viele Treffer die Liste zeichnet (R6).
const trefferMax = 12

const suchklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

const kaestchenklassen =
  'size-4 shrink-0 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Druckwahl({
  gewaehlt,
  waehlen,
  befund,
  angehalten = false,
}: Druckwahleigenschaften) {
  const texte = useTexte()
  const sprache = useSprache()
  const worte = texte.bausteine.druckwahl
  const suchfeld = useId()
  const hinweisID = useId()
  const befundID = useId()

  const [suche, setzeSuche] = useState('')

  // **Die ganze Sammlung, einmal** — dieselbe Bindung wie die Startseite und
  // die Deckzuordnung. Gefiltert wird hier, über Namen und Kennung; geordnet
  // nach dem Namen, in der Sprache des Fensters.
  const sammlung = useLaden(async () => {
    const antwort = await api().Sammlung()

    return auspacken(antwort, () => druckeAus(antwort.exemplare, sprache))
  }, [sprache])

  const alle = sammlung.daten ?? []
  const mitSuche = alle.length >= sucheAb
  const frage = falte(suche.trim(), sprache)

  // **Ohne Suchwort kein Treffer**: ab fünfzehn Drucken beginnt die Liste mit
  // der Wahl und nicht mit der ganzen Sammlung.
  const treffer =
    mitSuche && frage !== ''
      ? alle.filter((druck) => passt(druck, frage, sprache))
      : []
  const sichtbar = new Set(treffer.slice(0, trefferMax))
  const gewaehltMenge = new Set(gewaehlt)

  const gezeigt = mitSuche
    ? alle.filter(
        (druck) => gewaehltMenge.has(druck.printing) || sichtbar.has(druck),
      )
    : alle

  const beschriebenVon = [
    befund === undefined ? undefined : befundID,
    hinweisID,
  ]
    .filter((kennung) => kennung !== undefined)
    .join(' ')

  const umschalten = (printing: string, an: boolean) => {
    waehlen(
      an
        ? [...gewaehlt, printing]
        : gewaehlt.filter((eine) => eine !== printing),
    )
  }

  return (
    <fieldset
      aria-describedby={beschriebenVon}
      className="flex max-w-md flex-col gap-2"
    >
      <legend className="text-sm font-medium">{worte.drucke}</legend>

      <p id={hinweisID} className="text-sm text-leise">
        {worte.hinweis}
      </p>

      {sammlung.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {sammlung.fehler}
        </p>
      )}

      {/* **Eine leere Sammlung nennt den nächsten Schritt** (§6). */}
      {sammlung.daten !== undefined && alle.length === 0 && (
        <p className="text-sm text-leise">{worte.leer}</p>
      )}

      {mitSuche && (
        <label className="flex flex-col gap-1 text-sm" htmlFor={suchfeld}>
          {worte.suchen}
          <input
            id={suchfeld}
            type="search"
            value={suche}
            disabled={angehalten}
            onChange={(ereignis) => {
              setzeSuche(ereignis.target.value)
            }}
            className={suchklassen}
          />
        </label>
      )}

      {gezeigt.length > 0 && (
        <ul className="flex max-h-64 flex-col gap-1 overflow-y-auto">
          {gezeigt.map((druck) => (
            <li key={druck.printing} className="text-sm">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={gewaehltMenge.has(druck.printing)}
                  disabled={angehalten}
                  onChange={(ereignis) => {
                    umschalten(druck.printing, ereignis.target.checked)
                  }}
                  className={kaestchenklassen}
                />
                <span className="min-w-0 truncate">
                  {druck.menge} × {druck.bezeichnung}
                </span>
              </label>
            </li>
          ))}
        </ul>
      )}

      {mitSuche && (
        <p role="status" className="text-sm text-leise">
          {einsetzen(worte.gewaehlt, {
            anzahl: zaehlwort(
              gewaehlt.length,
              texte.allgemein.zaehlwoerter.drucke,
            ),
          })}
        </p>
      )}

      {mitSuche && frage === '' && (
        <p className="text-sm text-leise">
          {einsetzen(worte.nurSuche, {
            anzahl: zaehlwort(alle.length, texte.allgemein.zaehlwoerter.drucke),
          })}
        </p>
      )}

      {/* **Gefiltert leer ist nicht leer** (§6): die Sammlung hat Drucke, nur
          passt keiner zum Suchwort — der nächste Schritt ist ein anderes
          Wort. */}
      {mitSuche && frage !== '' && treffer.length === 0 && (
        <p className="text-sm text-leise">
          {einsetzen(worte.keiner, { suche: suche.trim() })}
        </p>
      )}

      {treffer.length > trefferMax && (
        <p className="text-sm text-leise">{worte.genauer}</p>
      )}

      {befund !== undefined && (
        <div role="alert">
          <p id={befundID} className="text-sm text-fehler">
            {befund}
          </p>
        </div>
      )}
    </fieldset>
  )
}

// druckeAus macht aus den Exemplaren die Drucke, die jemand besitzt — einen
// je Kennung, mit der Summe der Mengen, nach ihrem Namen geordnet.
//
// **Geordnet wird hier**, denn die Bindung gibt die Sammlung nach der Kennung
// des Drucks (`Sammlungsspeicher.Liste`), und eine Liste von Namen in der
// Ordnung einer Zahlenfolge liest sich wie zufällig. **In der Sprache des
// Fensters** (`docs/LOKALISIERUNG.md`); bei gleichem Namen — zwei Drucke
// derselben Karte — bleibt die Reihenfolge der Bindung.
function druckeAus(
  exemplare: readonly ExemplarDTO[],
  sprache: string,
): Druck[] {
  const drucke = new Map<string, Druck>()

  for (const exemplar of exemplare) {
    if (exemplar.wunsch) {
      continue
    }

    const bekannt = drucke.get(exemplar.printing)

    if (bekannt === undefined) {
      drucke.set(exemplar.printing, {
        printing: exemplar.printing,
        bezeichnung: exemplar.bezeichnung,
        menge: exemplar.menge,
      })
    } else {
      bekannt.menge += exemplar.menge
    }
  }

  return [...drucke.values()].sort((eines, anderes) =>
    eines.bezeichnung.localeCompare(anderes.bezeichnung, sprache),
  )
}

// passt sagt, ob ein Druck zu dem passt, was jemand tippt — über Bezeichnung
// und Kennung, ohne Rücksicht auf Groß und klein.
function passt(druck: Druck, frage: string, sprache: string): boolean {
  return (
    falte(druck.bezeichnung, sprache).includes(frage) ||
    falte(druck.printing, sprache).includes(frage)
  )
}
