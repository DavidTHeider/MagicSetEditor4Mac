// **Marken an einer Karte** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Was eine Marke ist
//
// Ein Wort für die eigene Arbeit — „fertig", „prüfen", „noch kein bild". Sie
// steht **neben** der Karte und nicht in ihr: was in einem Kartenfeld stünde,
// ginge in jede weitergegebene `.mse-set` und käme dort nie wieder heraus
// ([0053](../../docs/entscheidungen/0053-marken-liegen-neben-dem-set.md)).
//
// # Warum neben den Notizen
//
// Weil beides dieselbe Frage beantwortet — *wie weit bin ich mit dieser
// Karte?* Die Notiz ist der Satz, die Marke der Schlüssel: nach ihr lässt sich
// filtern und suchen, nach einem Satz nicht.
import { useState } from 'react'

import { api } from '../api'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { schreibenBeginnt, schreibenEndet } from '../speicherstand'

interface KartenmarkenEigenschaften {
  setID: string
  karteID: string

  // marken sind die Marken dieser Karte — sie kommen aus der Liste und werden
  // hier **nicht** noch einmal geholt (§7, Regel 2: eine Sache, ein Stand).
  marken: readonly string[]

  // geaendert lädt nach — der Aufrufer weiß, was alles daran hängt.
  geaendert: () => void
}

export function Kartenmarken({
  setID,
  karteID,
  marken,
  geaendert,
}: KartenmarkenEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.karten.marken
  const vorgang = useVorgang()

  const [entwurf, setzeEntwurf] = useState('')
  const [fehler, setzeFehler] = useState<string>()

  // **Ein anderer kann dieselbe Marke inzwischen gesetzt haben** (U05,
  // #2121). Der Entwurf bleibt dann stehen, der Satz sagt es, und Setzen
  // geht nicht — es gäbe nichts mehr zu setzen.
  const schonGesetzt = entwurf.trim() !== '' && marken.includes(entwurf.trim())

  // **Ein Vorgang je Karte** (§7, Regel 4): zwei Schreibläufe auf dieselbe
  // gingen gegen einen Stand, den das Backend noch nicht kennt.
  const schreiben = async (lauf: () => Promise<void>) => {
    await vorgang.starte(karteID, async () => {
      setzeFehler(undefined)
      // **Der Speicherstand hört mit** (UX-05.AK-03) — die Marken schrieben
      // bis zum Zweitblick über Phase 6.3 an ihm vorbei (M8).
      schreibenBeginnt(setID)

      try {
        await lauf()
        schreibenEndet(setID, karteID, worte.ueberschrift, true)

        // **Erst nach dem Schreiben nachladen** (§7, Regel 3): ein Lesefehler
        // danach macht die Änderung nicht rückgängig, und nach einem
        // **gescheiterten** Schreiben wird gar nicht nachgeladen.
        geaendert()
      } catch (ursache: unknown) {
        schreibenEndet(setID, karteID, worte.ueberschrift, false)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const setzen = async () => {
    await schreiben(async () => {
      auspacken(
        await api().MarkeSetzen(setID, karteID, entwurf),
        () => undefined,
      )

      setzeEntwurf('')
    })
  }

  return (
    <section aria-label={worte.ueberschrift} className="mt-4">
      <h3 className="text-sm font-medium">{worte.ueberschrift}</h3>

      <Erklaerungsabsatz
        text={worte.erklaerung}
        klasse="mt-1 max-w-prose text-sm text-leise"
      />

      {marken.length === 0 ? (
        <p className="mt-2 text-sm text-leise">{worte.keine}</p>
      ) : (
        <ul className="mt-2 flex flex-wrap gap-2">
          {marken.map((marke) => (
            <li key={marke}>
              <span className="inline-flex items-center gap-1 rounded-full border border-kante px-2 py-0.5 text-sm">
                {marke}

                <button
                  type="button"
                  // **Der Name nennt die Marke** — „Abnehmen" allein sagte an
                  // fünf Knöpfen nebeneinander nicht, welcher welcher ist.
                  aria-label={einsetzen(worte.entfernen, { marke })}
                  disabled={vorgang.laeuft(karteID)}
                  onClick={() => {
                    void schreiben(async () => {
                      auspacken(
                        await api().MarkeEntfernen(setID, karteID, marke),
                        () => undefined,
                      )
                    })
                  }}
                  className="rounded-full px-1 leading-none focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
                >
                  ×
                </button>
              </span>
            </li>
          ))}
        </ul>
      )}

      <form
        className="mt-2 flex gap-2"
        onSubmit={(ereignis) => {
          ereignis.preventDefault()
          void setzen()
        }}
      >
        <label className="sr-only" htmlFor={`marke-${karteID}`}>
          {worte.setzen}
        </label>

        <input
          id={`marke-${karteID}`}
          value={entwurf}
          placeholder={worte.platzhalter}
          disabled={vorgang.laeuft(karteID)}
          onChange={(ereignis) => {
            setzeEntwurf(ereignis.target.value)
          }}
          className="min-w-0 flex-1 rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        />

        <button
          type="submit"
          disabled={
            vorgang.laeuft(karteID) || entwurf.trim() === '' || schonGesetzt
          }
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        >
          {worte.setzen}
        </button>
      </form>

      {schonGesetzt && (
        <p role="status" className="mt-1 text-sm">
          {einsetzen(worte.schonGesetzt, { marke: entwurf.trim() })}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="mt-1 text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
