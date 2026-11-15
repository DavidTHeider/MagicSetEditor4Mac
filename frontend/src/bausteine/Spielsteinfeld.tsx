// **Die vier Angaben eines Spielsteins** — CR 701.7a, Werkstück zu
// #1623.
//
// # Warum vier Felder und nicht ein Textfeld
//
// **Weil ein Spielstein seine Eigenschaften mitbringt** (CR 111.1): er
// vertritt keine Karte, und der Effekt, der ihn erzeugt, nennt sie. Wer sie
// als eine Zeile eingäbe, müsste sie danach zerteilen — und die Zerteilung
// wäre eine zweite Deutung derselben Felder (§2).
//
// **Die Kennung fehlt mit Absicht**: sie vergibt der Kern (#2480) — das
// Fenster sieht nur Sichtkennungen und kann nicht wissen, welche frei ist.
import { useTexte } from '../sprache/kontext'

// **Schmaler als ein gewöhnliches Feld**: vier davon stehen nebeneinander,
// und `Stärke` braucht drei Zeichen.
const eingabeklassen =
  'w-28 rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Spielsteinfeld({
  wert,
  setze,
}: {
  wert: string
  setze: (wert: string) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.handlungen

  const teile = wert.split('|')

  // **Die Reihenfolge ist die der Nutzlast** (`name`, `super type`, `power`,
  // `toughness`) — `spielsteinfelder` zerteilt sie an derselben Stelle.
  const angaben = [
    eigene.spielsteinName,
    eigene.spielsteinTyp,
    eigene.spielsteinStaerke,
    eigene.spielsteinWiderstand,
  ]

  return (
    <div className="flex flex-wrap gap-2">
      {angaben.map((beschriftung, stelle) => (
        <label key={beschriftung} className="flex flex-col gap-1 text-sm">
          <span className="font-medium">{beschriftung}</span>

          <input
            type="text"
            value={teile[stelle] ?? ''}
            onChange={(ereignis) => {
              const neu = [...angaben].map((_, lauf) =>
                lauf === stelle ? ereignis.target.value : (teile[lauf] ?? ''),
              )

              setze(neu.join('|'))
            }}
            className={eingabeklassen}
          />
        </label>
      ))}
    </div>
  )
}
