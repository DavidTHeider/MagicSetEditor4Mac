// **Ein Feld für ein Kennwort** — Werkstück 6 von Phase 4.12
// (#1351).
//
// # Warum es nicht `Pflichtfeld` mit einem anderen `type` ist
//
// **Weil drei Dinge daran anders sind, und alle drei sind fachlich.**
//
//  1. **Es ist verdeckt** — und deshalb braucht es einen Weg, es zu zeigen:
//     wer sein Kennwort nicht sehen kann, tippt es falsch und erfährt es erst
//     bei der Anmeldung.
//  2. **Der Browser soll wissen, was es ist.** `autocomplete="new-password"`
//     heißt: hier nicht ausfüllen. Das gilt auch für das Anmelden, und zwar
//     mit Absicht — dasselbe Feld dient beiden Wegen, und ein automatisch
//     eingesetztes Kennwort in einem Formular, das auch **anlegen** kann,
//     legte ein Konto mit einem Kennwort an, das niemand für dieses gewählt
//     hat.
//  3. **Es hat eine Mindestlänge**, und die steht am Feld, bevor jemand auf
//     „anlegen" drückt (§6).
import { useId, useState } from 'react'

import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  beschriftung: string
  wert: string
  angehalten?: boolean

  // befund ist ein Fehler, den die Anwendung **zu diesem Feld** gemeldet hat.
  befund?: string | undefined

  // hinweis erklärt das Feld, **bevor** ein Fehler kommt (§6).
  hinweis?: string | undefined

  geaendert: (wert: string) => void
}

export function Kennwortfeld({
  beschriftung,
  wert,
  angehalten,
  befund,
  hinweis,
  geaendert,
}: Eigenschaften) {
  const texte = useTexte()
  const feldID = useId()
  const hinweisID = useId()

  const [sichtbar, setzeSichtbar] = useState(false)

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={feldID} className="text-sm">
        {beschriftung}
      </label>

      <div className="flex items-center gap-2">
        <input
          id={feldID}
          type={sichtbar ? 'text' : 'password'}
          value={wert}
          disabled={angehalten}
          autoComplete="new-password"
          aria-invalid={befund !== undefined}
          aria-describedby={
            befund !== undefined || hinweis !== undefined
              ? hinweisID
              : undefined
          }
          className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60 aria-[invalid=true]:border-fehler"
          onChange={(ereignis) => {
            geaendert(ereignis.target.value)
          }}
        />

        {/* **Ein Umschalter und kein Knopf mit Zustand im Namen**: er sagt,
            was er tun **wird**, und `aria-pressed` sagt, wie es gerade
            steht. */}
        <button
          type="button"
          aria-pressed={sichtbar}
          aria-label={
            sichtbar
              ? texte.bausteine.kennwort.verbergen
              : texte.bausteine.kennwort.zeigen
          }
          className="rounded-md border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          onClick={() => {
            setzeSichtbar((vorher) => !vorher)
          }}
        >
          {sichtbar
            ? texte.bausteine.kennwort.verbergenKurz
            : texte.bausteine.kennwort.zeigenKurz}
        </button>
      </div>

      {befund === undefined && hinweis !== undefined && (
        <p id={hinweisID} className="text-sm text-leise">
          {hinweis}
        </p>
      )}

      {befund !== undefined && (
        <div role="alert">
          <p id={hinweisID} className="text-sm text-fehler">
            {befund}
          </p>
        </div>
      )}
    </div>
  )
}
