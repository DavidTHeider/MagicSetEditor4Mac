// **Die Felder eines Befehls, aus seinem Schema** — Werkstück U12 von Phase
// 5.1 (#1423). Wie es aussieht; was es tut, steht in `useKommandozentrale`.
import { Feld } from '../bausteine/Feld'
import { useTexte } from '../sprache/kontext'
import type { Argumentbefund, Parameter, Werte } from '../befehle'

interface FormularEigenschaften {
  parameter: Parameter[]
  werte: Werte
  befunde: Argumentbefund[]
  angehalten: boolean
  setzeWert: (name: string, wert: string | boolean) => void
}

const feldklassen =
  'rounded-md border border-kante bg-flaeche px-2 py-1 text-text ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'
const jsonfeldklassen =
  'rounded-md border border-kante bg-flaeche px-2 py-1 font-mono text-sm text-text ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

export function Befehlsformular({
  parameter,
  werte,
  befunde,
  angehalten,
  setzeWert,
}: FormularEigenschaften) {
  const texte = useTexte().bausteine.kommandozentrale

  if (parameter.length === 0) {
    return <p className="text-sm text-leise">{texte.keineParameter}</p>
  }

  const befundtexte = {
    pflicht: texte.befundPflicht,
    zahl: texte.befundZahl,
    json: texte.befundJson,
  }

  return (
    <div className="flex flex-col gap-3">
      {parameter.map((einer) => {
        const befund = befunde.find((b) => b.name === einer.name)
        const beschriftung = einer.pflicht
          ? `${einer.name} (${texte.pflicht})`
          : einer.name
        const wert = werte[einer.name]
        const text = typeof wert === 'string' ? wert : ''

        if (einer.art === 'wahrheit') {
          return (
            <label key={einer.name} className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={wert === true}
                disabled={angehalten}
                onChange={(ereignis) => {
                  setzeWert(einer.name, ereignis.target.checked)
                }}
              />
              <span>{beschriftung}</span>
              {einer.erklaerung !== '' && (
                <span className="text-leise">{einer.erklaerung}</span>
              )}
            </label>
          )
        }

        return (
          <Feld
            key={einer.name}
            beschriftung={beschriftung}
            {...(einer.erklaerung !== '' || einer.art === 'json'
              ? {
                  hinweis:
                    einer.art === 'json'
                      ? `${einer.erklaerung} ${texte.jsonHinweis}`.trim()
                      : einer.erklaerung,
                }
              : {})}
            {...(befund === undefined
              ? {}
              : { befund: befundtexte[befund.grund] })}
          >
            {(verbindung) =>
              einer.art === 'wahl' ? (
                <select
                  {...verbindung}
                  value={text}
                  disabled={angehalten}
                  className={feldklassen}
                  onChange={(ereignis) => {
                    setzeWert(einer.name, ereignis.target.value)
                  }}
                >
                  <option value="" />
                  {einer.werte.map((moeglich) => (
                    <option key={moeglich} value={moeglich}>
                      {moeglich}
                    </option>
                  ))}
                </select>
              ) : einer.art === 'json' ? (
                <textarea
                  {...verbindung}
                  value={text}
                  rows={3}
                  disabled={angehalten}
                  className={jsonfeldklassen}
                  onChange={(ereignis) => {
                    setzeWert(einer.name, ereignis.target.value)
                  }}
                />
              ) : (
                <input
                  {...verbindung}
                  type="text"
                  inputMode={einer.art === 'zahl' ? 'decimal' : 'text'}
                  value={text}
                  disabled={angehalten}
                  className={feldklassen}
                  onChange={(ereignis) => {
                    setzeWert(einer.name, ereignis.target.value)
                  }}
                />
              )
            }
          </Feld>
        )
      })}
    </div>
  )
}
