// **Ein beschriftetes Eingabefeld** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum ein eigener Baustein und nicht `seiten/Feldzeile.tsx`
//
// Jene ist die Zeile eines **Kartenfeldes**: sie kennt die Beschreibung aus
// dem Spielpaket, Paletten, Wortlisten, hergeleitete Werte und den
// Bilddialog — dreizehn Eigenschaften für eine Sache, die die Sammlung nicht
// ist. Hier steht eine Beschriftung über einem Feld, und mehr soll es nicht
// sein (§2: der Schnitt folgt der Frage).
interface BeschriftetesfeldEigenschaften {
  bezeichnung: string
  wert: string
  setzeWert: (wert: string) => void

  // art ist der `type` des Feldes — `text`, wenn niemand etwas sagt.
  art?: string

  // breite ist die Spaltenzahl. **Sie steht hier**, weil eine Menge zwei
  // Stellen braucht und eine Scryfall-Kennung sechsunddreißig.
  breite?: number

  // wachsen lässt das Feld den freien Platz seiner Reihe nehmen — für einen
  // Namen ohne Längengrenze, den eine feste Spaltenzahl abschnitte (P01,
  // #2165).
  wachsen?: boolean
}

// eingabeklasse ist die Gestalt eines Eingabefelds.
//
// **Ausgeführt, weil ein Feld mit Befund sie auch braucht** — der
// Erwerbspreis steht über `Feld`, damit ein Befund am Feld stehen kann
// (§6), und soll trotzdem aussehen wie seine Nachbarn.
export const eingabeklasse =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

const breiteeingabe = `${eingabeklasse} w-full`

export function Beschriftetesfeld({
  bezeichnung,
  wert,
  setzeWert,
  art = 'text',
  breite = 12,
  wachsen = false,
}: BeschriftetesfeldEigenschaften) {
  return (
    <label
      className={
        wachsen ? 'flex min-w-48 flex-1 flex-col gap-1' : 'flex flex-col gap-1'
      }
    >
      <span className="text-sm font-medium">{bezeichnung}</span>
      <input
        type={art}
        value={wert}
        size={breite}
        onChange={(ereignis) => {
          setzeWert(ereignis.target.value)
        }}
        className={wachsen ? breiteeingabe : eingabeklasse}
      />
    </label>
  )
}
