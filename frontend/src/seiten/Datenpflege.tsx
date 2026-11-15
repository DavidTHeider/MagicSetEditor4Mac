// **Nachsehen, wählen, aufräumen** — Werkstück O8 von Phase 5.2
// (#1432).
//
// # Warum die Klasse die Vorauswahl entscheidet
//
// **Weil sie die Entscheidung ist.** Was `sicher` heißt, zeigt auf nichts und
// hält nichts fest — es steht angehakt da. Was `entscheidung` heißt, steht
// **nicht** angehakt da: die Klasse heißt „ein Mensch entscheidet", und eine
// Vorauswahl nähme ihm die Entscheidung ab, während sie so aussieht, als hätte
// er sie getroffen. Was `nurInformieren` heißt, hat gar kein Kästchen.
import { useId, useState } from 'react'

import type { Pflegelauf } from '../haken/useDatenpflege'
import type { PflegebefundDTO, PflegevorschauAntwort } from '../apipflege'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

const knopf =
  'rounded-md border border-kante px-3 py-1.5 text-sm font-medium ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

// vorausgewaehlt sagt, was ohne Zutun angehakt ist.
function vorausgewaehlt(befunde: PflegebefundDTO[]): Set<string> {
  return new Set(
    befunde.filter((befund) => befund.klasse === 'sicher').map((b) => b.art),
  )
}

export function Datenpflege({ stand }: { stand: Pflegelauf }) {
  const texte = useTexte()
  const worte = texte.seiten.datenpflege
  const titelID = useId()

  // **Die Wahl gehört der Vorschau.** Ein `key` auf dem Kasten räumt sie beim
  // Wechsel mit ab (§7, Regel 1 und 2 in einer Gestalt — `docs/SCHREIBWEGE.md`).
  return (
    <section aria-labelledby={titelID} className="max-w-xl">
      <h2 id={titelID} className="text-lg font-semibold tracking-tight">
        {worte.titel}
      </h2>

      <p className="mt-1 max-w-prose text-sm text-leise">{worte.hinweis}</p>

      <div className="mt-4">
        <button
          type="button"
          className={knopf}
          disabled={stand.laeuft}
          onClick={stand.nachsehen}
        >
          {stand.laeuft ? worte.nachsehenLaeuft : worte.nachsehen}
        </button>
      </div>

      {stand.vorschau !== undefined && (
        <Vorschau
          key={stand.vorschau.erstellt}
          stand={stand}
          vorschau={stand.vorschau}
        />
      )}

      {stand.bilanz !== undefined && (
        <div className="mt-4 rounded-md border border-kante p-3">
          <h3 className="text-sm font-medium">{worte.bilanzTitel}</h3>

          {stand.bilanz.gesichert !== '' && (
            <p className="mt-1 text-sm break-all text-leise">
              {einsetzen(worte.gesichert, { ort: stand.bilanz.gesichert })}
            </p>
          )}

          <ul className="mt-2 text-sm">
            {stand.bilanz.zeilen.map((zeile) => (
              <li key={zeile.art}>
                {einsetzen(worte.bilanzZeile, {
                  art: worte.arten[zeile.art] ?? zeile.art,
                  ausgang: worte.ausgaenge[zeile.ausgang] ?? zeile.ausgang,
                  anzahl: String(zeile.anzahl),
                })}
              </li>
            ))}
          </ul>

          <p className="mt-2 text-sm text-leise">
            {einsetzen(worte.bilanzStaende, {
              vorher: String(stand.bilanz.vorher.length),
              nachher: String(stand.bilanz.nachher.length),
            })}
          </p>
        </div>
      )}

      {stand.fehler !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {stand.fehler}
        </p>
      )}
    </section>
  )
}

// Vorschau zeigt die Befunde und nimmt die Wahl entgegen.
//
// **Sie bekommt die Vorschau als eigene Eigenschaft.** Der erste Anlauf holte
// sie aus `stand` und brauchte dafür eine Wache gegen `undefined` — einen
// Zweig, den kein Aufruf betritt, weil der Aufrufer sie schon geprüft hat
// (§3). Als Eigenschaft ist die Zusage im Typ statt in einem `if`.
function Vorschau({
  stand,
  vorschau,
}: {
  stand: Pflegelauf
  vorschau: PflegevorschauAntwort
}) {
  const worte = useTexte().seiten.datenpflege
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  const [gewaehlt, setzeGewaehlt] = useState(() =>
    vorausgewaehlt(vorschau.befunde),
  )

  if (vorschau.befunde.length === 0) {
    return <p className="mt-3 text-sm text-leise">{worte.sauber}</p>
  }

  const umschalten = (art: string) => {
    setzeGewaehlt((bisher) => {
      const neu = new Set(bisher)

      if (neu.has(art)) {
        neu.delete(art)
      } else {
        neu.add(art)
      }

      return neu
    })
  }

  const waehlbar = vorschau.befunde.filter(
    (befund) => befund.klasse !== 'nurInformieren',
  )

  return (
    <div className="mt-4">
      <p className="text-sm text-leise">
        {einsetzen(worte.gueltigbis, {
          zeit: alsZeitpunkt(vorschau.gueltigbis, darstellung, zeitzone),
        })}
      </p>

      <ul className="mt-3 flex flex-col gap-3">
        {vorschau.befunde.map((befund) => (
          <li key={befund.art} className="rounded-md border border-kante p-3">
            {befund.klasse === 'nurInformieren' ? (
              <p className="text-sm font-medium">
                {worte.arten[befund.art] ?? befund.art}
              </p>
            ) : (
              <label className="flex items-start gap-2 text-sm font-medium">
                <input
                  type="checkbox"
                  className="mt-0.5 size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                  checked={gewaehlt.has(befund.art)}
                  onChange={() => {
                    umschalten(befund.art)
                  }}
                />
                {worte.arten[befund.art] ?? befund.art}
              </label>
            )}

            <p className="mt-1 text-sm text-leise">
              {einsetzen(worte.anzahl, { anzahl: String(befund.anzahl) })} ·{' '}
              {worte.klassen[befund.klasse] ?? befund.klasse}
            </p>

            <p className="mt-1 max-w-prose text-sm text-leise">
              {worte.klassenhinweise[befund.klasse] ?? ''}
            </p>

            {befund.beispiele.length > 0 && (
              <p className="mt-1 font-mono text-sm break-all text-leise">
                {worte.beispieleTitel}: {befund.beispiele.join(', ')}
              </p>
            )}
          </li>
        ))}
      </ul>

      {/* **Kein Knopf, wenn nichts wählbar ist** — er könnte nur nichts tun. */}
      {waehlbar.length > 0 && (
        <div className="mt-4">
          <button
            type="button"
            className={knopf}
            disabled={stand.laeuft || gewaehlt.size === 0}
            onClick={() => {
              stand.pflegen(
                vorschau.erstellt,
                vorschau.befunde
                  .filter((befund) => gewaehlt.has(befund.art))
                  .map((befund) => ({
                    art: befund.art,
                    revision: befund.revision,
                  })),
              )
            }}
          >
            {stand.laeuft ? worte.aufraeumenLaeuft : worte.aufraeumen}
          </button>

          <Erklaerungsabsatz
            text={worte.aufraeumenHinweis}
            klasse="mt-1 max-w-prose text-sm text-leise"
          />
        </div>
      )}
    </div>
  )
}
