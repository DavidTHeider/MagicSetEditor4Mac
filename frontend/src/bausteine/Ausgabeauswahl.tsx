// **Welche Karten hinausgehen** — Werkstück 4 von Phase 3.6
// (#1167).
//
// # ported, und die Fundstelle ist genau
//
//	enum ExportCardSelectionType
//	{  EXPORT_SEL_ENTIRE_SET
//	,  EXPORT_SEL_SUBSET
//	,  EXPORT_SEL_CUSTOM
//	};
//
// `src/gui/card_select_window.hpp`, Zeilen 20–24, Fassung f9d9356 vom
// 2021-07-06. MSE2 stellt die Wahl vor **vier** seiner sieben Ausgabewege:
// Drucken, Druckvorschau, HTML-Ausgabe und **mehrere Bilder**
// (`src/gui/set/window.cpp`, Zeilen 656–712).
//
// **Nicht** vor dem Einzelbild — das nimmt die aufgeschlagene Karte —, nicht
// vor Apprentice, nicht vor MWS und **nicht** beim Speichern des Sets: ein
// gespeichertes Teilset wäre ein anderes Set.
//
// # Zwei Wahlen und nicht drei
//
// Die mittlere ist bei MSE2 eine Liste, die ein Reiter beisteuert; im
// Auslieferungsstand tut das genau einer — `RandomPackPanel` mit „export
// generated packs" (`random_pack_panel.cpp`, Zeilen 627–633). Hier gibt es
// diesen Beitrag nicht, und eine Wahl ohne Inhalt wäre eine leere Zeile in
// jedem Ausgabefenster.
//
// # Kein Dialog über einem Dialog
//
// MSE2 öffnet dafür ein `CardSelectWindow`. §6 sagt: „Keine verschachtelten
// Dialoge. Bearbeitung im Seitenbereich, nicht in einem Modal über einem
// Modal." Die Liste steht deshalb **auf der Seite**, eingeklappt, solange das
// ganze Set gemeint ist.
//
// # Ein Set, wie es wirklich ist — UXR-09 (#1917)
//
// **Dreihundert Karten sind dreihundert Kästchen.** Ein Grundset von Magic
// hat 250 bis 280 Karten; die Teilmenge zeichnete je Karte ein Kästchen, und
// R6 aus `docs/UIREVIEW.md` erlaubt auf einer Fläche zweihundert
// Bedienelemente. **Adapted**: MSE2s `CardSelectWindow` ist eine Liste in
// einem eigenen Fenster, die alle Karten trägt. Hier stehen sechzig auf
// einmal, weitere auf Knopfdruck, und ab fünfzehn Karten ein Feld, das
// eingrenzt — die Gestalt der Sets- und Deckliste (`Listenansicht`,
// UX-11.AK-03), mit deren Seitengröße und deren Sätzen.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Knopf } from './Knopf'
import { Seitenladung } from './Seitenladung'
import { seitengroesse } from './Listenansicht'

// suchgrenze ist die Zahl, ab der eingegrenzt werden kann — dieselbe wie in
// R7: über fünfzehn Werte gehört zu einer Wahl eine Suche.
export const suchgrenze = 15

interface AusgabeauswahlEigenschaften {
  setID: string

  // gewaehlt sind die Kennungen — **leer heißt: das ganze Set**, wie in
  // `domain.Kartenauswahl`.
  gewaehlt: readonly string[]

  setzeGewaehlt: (gewaehlt: readonly string[]) => void

  // angehalten sperrt die Wahl, solange die Ausgabe läuft (§7, Regel 4).
  angehalten: boolean
}

const fokus =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// **Der Ring steht ausgeschrieben** und nicht aus `fokus` eingesetzt: der
// Wächter `fokusring_test.go` folgt einer Größe bis zu ihrer Zuweisung und
// liest dort, nicht in einer zweiten.
const suchklasse =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Ausgabeauswahl({
  setID,
  gewaehlt,
  setzeGewaehlt,
  angehalten,
}: AusgabeauswahlEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.ausgabeauswahl
  const liste = texte.bausteine.listenansicht
  const gruppe = useId()
  const suchfeld = useId()

  const [suche, setzeSuche] = useState('')
  const [seiten, setzeSeiten] = useState(1)

  const stand = useLaden(async () => {
    const antwort = await api().Kartenzeilen(setID)

    return auspacken(antwort, () => antwort.zeilen)
  }, [setID])

  const zeilen = stand.daten ?? []
  const teilmenge = gewaehlt.length > 0

  // name ist, was an einem Kästchen steht — und wonach eingegrenzt wird.
  const name = (zeile: (typeof zeilen)[number]) =>
    zeile.name === '' ? eigene.ohneNamen : zeile.name

  // **Groß oder klein geschrieben, gleich** — wie die Suche der Stilwahl;
  // eine Kartenzahl im Namen findet man über ihre Ziffern.
  const gesucht = suche.trim().toLowerCase()
  const treffer = zeilen.filter((zeile) =>
    name(zeile).toLowerCase().includes(gesucht),
  )
  const gezeigt = treffer.slice(0, seiten * seitengroesse)

  return (
    <fieldset className="flex flex-col gap-2">
      <legend className="text-sm font-medium">{eigene.aufschrift}</legend>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="radio"
          name={gruppe}
          checked={!teilmenge}
          disabled={angehalten}
          onChange={() => {
            setzeGewaehlt([])
          }}
          className={fokus}
        />
        {/* **Solange die Zeilen ausstehen, keine Zahl** — G11 (#2277):
            „(0 Karten)" sähe aus wie ein Set ohne Karten. */}
        {stand.daten === undefined
          ? eigene.ganzesSetOhneZahl
          : einsetzen(eigene.ganzesSet, { zahl: String(zeilen.length) })}
      </label>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="radio"
          name={gruppe}
          checked={teilmenge}
          disabled={angehalten || zeilen.length === 0}
          onChange={() => {
            // **Die erste Karte wird vorgewählt.** Eine Teilmenge ohne Karte
            // wäre eine Wahl, die keine Ausgabe zulässt — und der Knopf
            // daneben meldete einen Fehler, den niemand verursacht hat.
            setzeGewaehlt(zeilen.slice(0, 1).map((zeile) => zeile.id))
          }}
          className={fokus}
        />
        {einsetzen(eigene.teilmenge, { zahl: String(gewaehlt.length) })}
      </label>

      {teilmenge && zeilen.length >= suchgrenze && (
        <div className="flex max-w-md flex-col gap-1">
          <label htmlFor={suchfeld} className="text-sm">
            {eigene.eingrenzen}
          </label>
          <input
            id={suchfeld}
            type="search"
            value={suche}
            disabled={angehalten}
            onChange={(ereignis) => {
              setzeSuche(ereignis.target.value)
              setzeSeiten(1)
            }}
            className={suchklasse}
          />
        </div>
      )}

      {/* **Gefiltert leer ist nicht leer** — dieselbe Auskunft wie in der
          Listenansicht. Eine leere Liste ohne Satz sähe aus wie ein Set ohne
          Karten. */}
      {teilmenge && stand.daten !== undefined && treffer.length === 0 && (
        <p className="text-sm text-leise">
          {einsetzen(liste.nichtsPasst, { suche: suche.trim() })}
        </p>
      )}

      {teilmenge && gezeigt.length > 0 && (
        <ul
          aria-label={eigene.aufschrift}
          className="max-h-64 overflow-y-auto rounded-sm border border-kante p-2"
        >
          {gezeigt.map((zeile) => (
            <li key={zeile.id}>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={gewaehlt.includes(zeile.id)}
                  disabled={angehalten}
                  onChange={(ereignis) => {
                    setzeGewaehlt(
                      ereignis.target.checked
                        ? [...gewaehlt, zeile.id]
                        : gewaehlt.filter((eine) => eine !== zeile.id),
                    )
                  }}
                  className={fokus}
                />
                {name(zeile)}
              </label>
            </li>
          ))}
        </ul>
      )}

      {teilmenge && treffer.length > gezeigt.length && (
        <div className="flex flex-wrap items-center gap-3">
          <p className="text-sm text-leise">
            {einsetzen(liste.zahl, {
              gezeigt: String(gezeigt.length),
              gesamt: String(treffer.length),
            })}
          </p>
          <Knopf
            klein
            aufKlick={() => {
              setzeSeiten(seiten + 1)
            }}
          >
            {einsetzen(liste.weitere, {
              anzahl: String(
                Math.min(seitengroesse, treffer.length - gezeigt.length),
              ),
            })}
          </Knopf>
        </div>
      )}

      {/* **Laden und Scheitern sagen sich** — G11 (#2277), mit einem
          Rückweg statt der Seite, die man verlassen müsste. */}
      <Seitenladung zustaende={[stand]} />
    </fieldset>
  )
}
