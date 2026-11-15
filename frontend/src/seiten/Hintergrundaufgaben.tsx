// **Hintergrundaufgaben** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Was hier steht
//
// Was die Anwendung gerade im Hintergrund holt — Preise für die Sammlung, die
// Sammeldaten von Scryfall, eine neue Fassung — mit Stand, Fortschritt und
// Dauer; dazu die letzten fünf beendeten. Geführt wird das Buch in Go
// (`service.Auftragsbuch`), und diese Seite fragt es.
//
// # Gefragt wird nur, solange etwas läuft
//
// Ein Intervall, das immer liefe, wäre Arbeit ohne Betrachter (wie
// `useBestandsstand` sagt). Deshalb tickt die Uhr nur, während das Buch einen
// laufenden Auftrag nennt — und hört auf, sobald der letzte beendet ist. Der
// Abstand ist derselbe wie beim Bestandsstand; er ist ein Wert der Anzeige.
import { useEffect, useState } from 'react'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { laufsatz } from '../bausteine/laufsatz'
import { Flaeche } from '../bausteine/Flaeche'
import { abstand as bestandsabstand } from '../haken/useBestandsstand'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import type { AuftraegeAntwort, AuftragDTO } from '../apiaufmerksamkeit'

interface HintergrundaufgabenEigenschaften {
  stand: number

  // abstand ist die Zeit zwischen zwei Abfragen, solange etwas läuft — für
  // Tests, die nicht zwei Sekunden warten wollen.
  abstand?: number
}

async function holen(): Promise<AuftraegeAntwort> {
  const antwort = await api().Auftraege()

  return auspacken(antwort, () => antwort)
}

const standfarben: Record<AuftragDTO['stand'], string> = {
  laeuft: 'text-leise',
  fertig: 'text-erfolg',
  gescheitert: 'text-fehler',
  abgebrochen: 'text-leise',
}

const knopf =
  'rounded-sm border border-kante px-2 py-0.5 text-xs focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

// istArt sagt, ob der Katalog ein Wort für eine Art hat — sie kommt als
// Schlüssel aus Go, und eine neue fällt auf ihren Schlüssel zurück.
function istArt<T extends object>(
  arten: T,
  art: string,
): art is Extract<keyof T, string> {
  return art in arten
}

export function Hintergrundaufgaben({
  stand,
  abstand = bestandsabstand,
}: HintergrundaufgabenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.start.auftraege
  const [tick, setzeTick] = useState(0)
  // **Behalten, nicht räumen**: das Bezugsobjekt wechselt nie — es ist immer
  // dasselbe Buch —, und eine Liste, die alle zwei Sekunden verschwindet und
  // wiederkommt, ist keine Anzeige eines Fortschritts.
  const buch = useLaden(holen, [stand, tick], true)
  const laufend = buch.daten?.laufend ?? 0
  const [anhaltefehler, setzeAnhaltefehler] = useState<string>()

  // **Nach dem Anhalten wird gefragt, nicht angenommen** — beendet ist der
  // Lauf erst zwischen zwei Stücken, und das sagt das Buch.
  const anhalten = async (id: string) => {
    setzeAnhaltefehler(undefined)

    try {
      auspacken(await api().AuftragAbbrechen(id), () => undefined)
      setzeTick((bisher) => bisher + 1)
    } catch (ursache: unknown) {
      setzeAnhaltefehler(meldungAus(ursache, texte))
    }
  }

  useEffect(() => {
    if (laufend === 0) {
      return undefined
    }

    const uhr = setInterval(() => {
      setzeTick((bisher) => bisher + 1)
    }, abstand)

    return () => {
      clearInterval(uhr)
    }
  }, [laufend, abstand])

  if (buch.fehler !== undefined) {
    return (
      <p role="alert" className="mt-8 text-fehler">
        {texte.bausteine.zustaende.fehler} {buch.fehler}
      </p>
    )
  }

  // **Ein leeres Buch steht nicht da** — UXR-03
  // (#1911).
  // Hier stand ein Bereich mit Überschrift und einem Satz, der erklärte, was
  // erschiene, wenn etwas liefe — auf jeder Startseite, auf der nichts lief.
  // Das Artboard *Main* kennt den Bereich nicht. **Gefragt wird trotzdem**:
  // was läuft oder gerade beendet ist, steht da, sobald das Buch es führt.
  const daten = buch.daten
  if (daten === undefined || daten.auftraege.length === 0) {
    return null
  }

  return (
    <section aria-label={eigene.titel} className="mt-8 flex flex-col gap-4">
      <h2 className="text-lg font-semibold">{eigene.titel}</h2>

      <Flaeche>
        {anhaltefehler !== undefined && (
          <p role="alert" className="mb-2 text-sm text-fehler">
            {texte.bausteine.zustaende.fehler} {anhaltefehler}
          </p>
        )}
        {daten.laufend > 0 && (
          <p role="status" className="mb-2 text-sm text-leise">
            {zaehlwort(daten.laufend, eigene.laufend)}
          </p>
        )}
        <ul className="flex flex-col gap-1">
          {daten.auftraege.map((auftrag) => {
            const art = istArt(eigene.arten, auftrag.art)
              ? eigene.arten[auftrag.art]
              : auftrag.art
            const fortschritt =
              auftrag.einheit === ''
                ? auftrag.fortschritt
                : laufsatz(texte.bausteine.laufstand, auftrag)

            return (
              <li
                key={auftrag.id}
                className="flex flex-wrap items-baseline gap-x-2 px-2 py-1 text-sm"
              >
                <span className="font-medium">
                  {auftrag.was === '' ? art : `${art} · ${auftrag.was}`}
                </span>
                <span className={standfarben[auftrag.stand]}>
                  {eigene.staende[auftrag.stand]}
                </span>
                {fortschritt !== '' && (
                  <span className="text-leise">{fortschritt}</span>
                )}
                <span className="text-leise">
                  {einsetzen(eigene.dauer, {
                    sekunden: String(auftrag.sekunden),
                  })}
                </span>
                {auftrag.stand === 'laeuft' && auftrag.abbrechbar && (
                  <button
                    type="button"
                    className={knopf}
                    onClick={() => {
                      void anhalten(auftrag.id)
                    }}
                  >
                    {einsetzen(eigene.abbrechen, { art })}
                  </button>
                )}
                {auftrag.grund !== '' && (
                  <span className="block w-full text-xs text-fehler">
                    {auftrag.grund}
                  </span>
                )}
              </li>
            )
          })}
        </ul>
      </Flaeche>
    </section>
  )
}
