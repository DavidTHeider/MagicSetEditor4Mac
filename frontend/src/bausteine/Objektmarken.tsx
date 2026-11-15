// **Die Marken eines Objekts** — UX-12
// (#1684),
// Artboard *Sammlung*: „+ Marke" im Inspector.
//
// Dieselbe Gestalt wie die Marken einer Karte (`seiten/Kartenmarken.tsx`) —
// Kacheln mit einem × und ein Feld darunter —, für jede Art. **Die
// Normalform macht Go**: was hier getippt wird, kommt so an, wie es gemeint
// ist, und steht danach so da, wie es gespeichert wurde.
import { useId, useState } from 'react'

import { api } from '../api'
import type { Markenart } from '../apiobjektmarken'
import { useSchreibsperre } from '../haken/schreibrecht'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'

interface ObjektmarkenEigenschaften {
  art: Markenart
  kennung: string
  marken: readonly string[]

  // geaendert sagt dem Aufrufer, dass neu zu lesen ist — **nach** einem
  // geglückten Schreiben (§7, Regel 3).
  geaendert: () => void
}

export function Objektmarken({
  art,
  kennung,
  marken,
  geaendert,
}: ObjektmarkenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.objektmarken
  const vorgang = useVorgang()
  const feld = useId()

  const [entwurf, setzeEntwurf] = useState('')

  // **Marken schreiben ins Profil** — G10 (#2276): ein Leser sieht sie,
  // setzt und entfernt aber keine; Feld und Knöpfe sagen es, statt erst am
  // Tor zu scheitern.
  const schreibt = useSchreibsperre(true)
  const [fehler, setzeFehler] = useState<string>()

  // **Ein Vorgang je Objekt** (§7, Regel 4).
  const schreiben = (lauf: () => Promise<void>) =>
    vorgang.starte(kennung, async () => {
      try {
        await lauf()
        setzeFehler(undefined)
        geaendert()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section aria-label={eigene.ueberschrift} className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.ueberschrift}</h3>

      {marken.length === 0 ? (
        <p className="max-w-prose text-sm text-leise">{eigene.keine}</p>
      ) : (
        <ul className="flex flex-wrap gap-2">
          {marken.map((marke) => (
            <li key={marke}>
              <span className="inline-flex items-center gap-1 rounded-full border border-kante px-2 py-0.5 text-sm">
                {marke}
                <button
                  type="button"
                  aria-label={einsetzen(eigene.entfernen, { marke })}
                  {...schreibt.knopf}
                  title={schreibt.titel}
                  disabled={vorgang.laeuft(kennung)}
                  onClick={() => {
                    if (schreibt.gesperrt) return

                    void schreiben(async () => {
                      auspacken(
                        await api().ObjektmarkeEntfernen(art, kennung, marke),
                        () => undefined,
                      )
                    })
                  }}
                  className="rounded-full px-1 leading-none focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-disabled:cursor-not-allowed aria-disabled:opacity-50"
                >
                  ×
                </button>
              </span>
            </li>
          ))}
        </ul>
      )}

      <form
        className="flex gap-2"
        onSubmit={(ereignis) => {
          ereignis.preventDefault()
          void schreiben(async () => {
            auspacken(
              await api().ObjektmarkeSetzen(art, kennung, entwurf),
              () => undefined,
            )
            setzeEntwurf('')
          })
        }}
      >
        <label className="sr-only" htmlFor={feld}>
          {eigene.setzenFuer}
        </label>
        <input
          id={feld}
          value={entwurf}
          placeholder={eigene.platzhalter}
          // **Ein Leser tippt nicht, was er nicht setzen darf** — und ohne
          // Entwurf bleibt das Absenden zu, auch über die Eingabetaste.
          readOnly={schreibt.gesperrt}
          // **Das Feld sagt den Grund wie seine Knöpfe** (H01, #2255).
          aria-describedby={schreibt.knopf['aria-describedby']}
          disabled={vorgang.laeuft(kennung)}
          onChange={(ereignis) => {
            setzeEntwurf(ereignis.target.value)
          }}
          className="min-w-0 flex-1 rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
        <button
          type="submit"
          {...schreibt.knopf}
          title={schreibt.titel}
          disabled={vorgang.laeuft(kennung) || entwurf.trim() === ''}
          className="rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-disabled:cursor-not-allowed aria-disabled:opacity-50"
        >
          {eigene.setzen}
        </button>
      </form>
      {schreibt.satz}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
