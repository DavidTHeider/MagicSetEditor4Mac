// **Das Band für eine neuere Fassung** —
// #1100.
//
// # Warum neben dem Meldungsband und nicht darin
//
// Das Meldungsband zeigt **eine** Meldung und wird weggeklickt. Hier steht
// eine Meldung mit **Handlung** — laden —, die einen Fortschritt zeigt und
// danach einen Ort nennt. Beides in einen Verbund zu zwingen hieße, dem
// Meldungsband eine Schaltfläche und einen Balken beizubringen, die alle
// anderen Meldungen nicht haben.
//
// # `role="status"` und nicht `role="alert"`
//
// Aus demselben Grund wie beim Meldungsband: eine neue Fassung ist keine
// Störung, die einen Vorleser mitten im Satz unterbricht.
import { Vertraeglichkeit } from './Vertraeglichkeit'
import { useState } from 'react'

import { api } from '../fassade'
import { useAktualisierung } from '../haken/useAktualisierung'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { auspacken } from '../umschlag'

import type { VertraeglichkeitAntwort } from '../apivertraeglichkeit'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

const bandrahmen =
  'mt-2 flex flex-wrap items-center gap-3 rounded-sm border border-marke px-3 py-2 text-sm text-marke'

const knopf =
  'rounded-sm border border-marke px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// mengeAus gibt eine Bytezahl in Mebibyte, auf eine Stelle gerundet.
//
// **Ohne Einheitenname im Quelltext**: `MiB` ist keine Übersetzung, sondern
// ein Zeichen — es steht in beiden Katalogen gleich und trotzdem dort, weil
// `keintextimjsx.test.ts` jede Zeichenkette im JSX abweist.
function mengeAus(bytes: number): string {
  return (bytes / (1 << 20)).toFixed(1) + ' MiB'
}

export function Fassungsband() {
  const texte = useTexte()
  const stand = useAktualisierung()
  const vorgang = useVorgang()

  const band = texte.allgemein.fassungsband

  // **Das Urteil holt das Band** (O6, #1430), nicht das Bauteil daneben: es
  // braucht es selbst, um den Ladeknopf wegzulassen, wenn die Fassung diesen
  // Bestand gar nicht öffnen kann. `useLaden` verwirft dabei die Antwort
  // eines überholten Laufs (§7, Regel 2).
  const geholt = useLaden(async () => {
    if (stand.fassung === '') {
      return undefined
    }

    const antwort = await api().Vertraeglichkeit()

    return auspacken(antwort, () => antwort)
  }, [stand.fassung])

  const [vorbereitet, setzeVorbereitet] = useState<VertraeglichkeitAntwort>()
  const [urteilsfehler, setzeUrteilsfehler] = useState<string>()

  const urteil = vorbereitet ?? geholt.daten

  const vorbereiten = () =>
    void vorgang.starte('vertraeglichkeit', async () => {
      try {
        const antwort = await api().WechselVorbereiten()

        setzeVorbereitet(auspacken(antwort, () => antwort))
        setzeUrteilsfehler(undefined)
      } catch (ursache: unknown) {
        setzeUrteilsfehler(meldungAus(ursache, texte))
      }
    })

  // **Der Bereich steht immer da, auch leer** — dieselbe Überlegung wie beim
  // Meldungsband: ein Vorleser meldet nur, was sich in einem Bereich ändert,
  // den er schon kennt.
  return (
    <div
      role="status"
      aria-live="polite"
      aria-label={band.bereich}
      className="px-6"
    >
      {stand.fassung === '' ? null : (
        <div className={bandrahmen}>
          <span>{einsetzen(band.neu, { fassung: stand.fassung })}</span>

          {stand.seite === '' ? null : (
            <a
              href={stand.seite}
              className="underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {band.seite}
            </a>
          )}

          {/* **Ob sie diesen Bestand öffnen kann** (O6, #1430) — vor dem
              Laden, nicht danach. */}
          <Vertraeglichkeit
            urteil={urteil}
            aufVorbereiten={vorbereiten}
            fehler={urteilsfehler ?? geholt.fehler}
          />

          {/* **Kein Ladeknopf für eine Fassung, die diesen Bestand nicht
              öffnen kann** — das fand die Sichtprüfung zu O6 (#1430): das
              Band sagte es und bot das Laden trotzdem an. */}
          {/* **Und keiner unter „melden“** — ob geladen werden darf, sagt
              die Antwort aus Go (AP-12, #2369); bis dahin stand er da, und
              ein Druck kam als Fehler zurück. */}
          {stand.pfad === '' &&
          stand.holen &&
          urteil?.urteil !== 'unvertraeglich' ? (
            <button
              type="button"
              onClick={() => void stand.laden()}
              disabled={stand.laeuft}
              className={knopf}
            >
              {stand.laeuft ? band.laeuft : band.laden}
            </button>
          ) : null}

          {/* **Erst ab dem ersten Block.** Vor ihm gibt es keine Zahl, und
              „0,0 MiB geladen" wäre eine Messung, die nicht stattgefunden hat
              (§10). Dass etwas läuft, sagt die Schaltfläche daneben. */}
          {stand.laeuft && stand.gelesen > 0 ? (
            <span>
              {stand.anteil === undefined
                ? einsetzen(band.fortschrittOhneLaenge, {
                    gelesen: mengeAus(stand.gelesen),
                  })
                : einsetzen(band.fortschritt, {
                    gelesen: mengeAus(stand.gelesen),
                    ganz: mengeAus(stand.gelesen / stand.anteil),
                  })}
            </span>
          ) : null}

          {stand.pfad === '' ? null : (
            <span>
              {einsetzen(band.geladen, {
                fassung: stand.fassung,
                pfad: stand.pfad,
              })}
            </span>
          )}

          {stand.fehler === '' ? null : (
            <span className="text-fehler">{stand.fehler}</span>
          )}

          <button
            type="button"
            onClick={stand.weglegen}
            aria-label={band.weglegen}
            className={knopf}
          >
            ×
          </button>
        </div>
      )}
    </div>
  )
}
