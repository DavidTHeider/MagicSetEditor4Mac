// Ein Optionsfeld einer Exportvorlage als Eingabe.
//
// **Eine eigene Datei**, weil `SetVorlagenausfuhr.tsx` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was** die Ausfuhr tut
// — laden, wählen, laufen —, hier, **wie** eine einzelne Wahl aussieht.
import type { ExportvorlageDTO } from '../vorlagen'

// Wahlfeld zeigt eine Wahlmöglichkeit der Vorlage als Eingabe.
//
// **Der Name kommt aus dem Paket und wird nicht übersetzt** (§9): er steht so
// in der Datei, und eine Übersetzung machte jede vorhandene Vorlage unlesbar.
export function Wahlfeld({
  option,
  wert,
  setzen,
}: {
  option: ExportvorlageDTO['optionen'][number]
  wert: string
  setzen: (wert: string) => void
}) {
  if (option.art === 'boolean') {
    return (
      <label className="mt-2 flex items-center gap-2">
        <input
          type="checkbox"
          checked={wert === 'yes'}
          onChange={(ereignis) => {
            setzen(ereignis.target.checked ? 'yes' : 'no')
          }}
        />
        <span>{option.beschreibung || option.name}</span>
      </label>
    )
  }

  if (option.auswahl.length > 0) {
    return (
      <label className="mt-2 block">
        <span className="text-sm text-leise">
          {option.beschreibung || option.name}
        </span>
        <select
          value={wert}
          onChange={(ereignis) => {
            setzen(ereignis.target.value)
          }}
          className="mt-1 block rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {option.auswahl.map((moeglich) => (
            <option key={moeglich} value={moeglich}>
              {moeglich}
            </option>
          ))}
        </select>
      </label>
    )
  }

  return (
    <label className="mt-2 block">
      <span className="text-sm text-leise">
        {option.beschreibung || option.name}
      </span>
      <input
        type="text"
        value={wert}
        onChange={(ereignis) => {
          setzen(ereignis.target.value)
        }}
        className="mt-1 block rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />
    </label>
  )
}
