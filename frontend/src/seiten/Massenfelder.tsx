// **Feld und Wert der Massenänderung** — aus `Massenaenderung.tsx`
// herausgelöst, als die Vorschau dazukam (U6, #1417) und die Seite über die
// Zeilengrenze ging (§2). Der Schnitt folgt der Frage: hier steht, **was**
// jemand wählt; dort, was daraus wird.
import type { FeldbeschreibungDTO } from '../api'
import { useTexte } from '../sprache/kontext'
import { aufschriftVon, gruppiereFelder } from './feldgruppen'

interface MassenfelderEigenschaften {
  angeboten: readonly FeldbeschreibungDTO[]
  feld: string
  gewaehltesFeld: FeldbeschreibungDTO | undefined
  wert: string
  feldWaehlen: (feld: string) => void
  wertSetzen: (wert: string) => void
}

const feldklasse =
  'rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Massenfelder({
  angeboten,
  feld,
  gewaehltesFeld,
  wert,
  feldWaehlen,
  wertSetzen,
}: MassenfelderEigenschaften) {
  const worte = useTexte().seiten.karten.massen

  return (
    <div className="mt-3 flex flex-wrap items-end gap-4">
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{worte.feld}</span>
        <select
          value={feld}
          onChange={(ereignis) => {
            feldWaehlen(ereignis.target.value)
          }}
          className={feldklasse}
        >
          {/* **Eine leere Wahl zuerst.** Ohne sie stünde beim Öffnen das
              erste Feld da, und ein Klick schriebe es — ohne dass jemand es
              gewählt hätte. */}
          <option value="">{worte.keinFeld}</option>
          {/* **Die Glieder einer Feldreihe stehen unter ihrem Grundfeld**
              (#1167): `Name`, `Name 2`, `Name 3` als gleichrangige Einträge
              ließen niemanden sehen, dass sie dasselbe Feld an drei Stellen
              sind. **Die Aufschrift und nicht der Bezeichner** (§9). */}
          {gruppiereFelder(angeboten).map((gruppe) =>
            gruppe.reihe === '' ? (
              gruppe.felder.map((eines) => (
                <option key={eines.name} value={eines.name}>
                  {aufschriftVon(eines)}
                </option>
              ))
            ) : (
              <optgroup key={gruppe.reihe} label={gruppe.aufschrift}>
                {gruppe.felder.map((eines) => (
                  <option key={eines.name} value={eines.name}>
                    {aufschriftVon(eines)}
                  </option>
                ))}
              </optgroup>
            ),
          )}
        </select>
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{worte.wert}</span>
        {/* **Ein Wahlfeld bekommt seine Wahlmöglichkeiten** (Z3-01) — nur
            für `choice`; ein Text bleibt ein Textfeld. */}
        {gewaehltesFeld?.art === 'choice' &&
        gewaehltesFeld.wahlen.length > 0 ? (
          <select
            value={wert}
            onChange={(ereignis) => {
              wertSetzen(ereignis.target.value)
            }}
            className={feldklasse}
          >
            <option value="">{worte.keinWert}</option>
            {gewaehltesFeld.wahlen.map((wahl) => (
              <option key={wahl} value={wahl}>
                {wahl}
              </option>
            ))}
          </select>
        ) : (
          <input
            type="text"
            value={wert}
            onChange={(ereignis) => {
              wertSetzen(ereignis.target.value)
            }}
            // **Nicht von selbst berichtigt** (G03, #2250) — wie `Feldeingabe`.
            autoCorrect="off"
            autoCapitalize="off"
            className={feldklasse}
          />
        )}
      </label>
    </div>
  )
}
