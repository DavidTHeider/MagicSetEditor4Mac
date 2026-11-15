// **Nach einer Marke filtern** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Warum die Liste hier gefiltert wird und nicht im Dienst
//
// **Weil die Zeilen schon da sind.** Der Aufbau der Kartenliste bringt jede
// Zeile samt ihren Marken (`Listengestalt.marken`); ein Filter im Dienst wäre
// ein zweiter Ruf für eine Auswahl, die aus dem vorhandenen Stand hervorgeht —
// und zwei Stände derselben Liste liefen auseinander (§7, Regel 2).
//
// **Und weil er dann sofort greift.** Wer eine Marke anklickt, sieht die Liste
// im selben Zeichenschritt.
//
// # Warum eine Marke und nicht mehrere
//
// Die Frage, die jemand hier stellt, ist „zeig mir die, die noch zu prüfen
// sind". Zwei Marken zugleich brauchen eine zweite Frage — **und** oder
// **oder** —, und die ließe sich nicht anklicken, sondern nur erklären.
import { useTexte } from '../sprache/kontext'

interface MarkenfilterEigenschaften {
  // marken sind die Marken, die im Set vorkommen — jede einmal.
  marken: readonly string[]

  // gewaehlt ist die Marke, nach der gefiltert wird — leer heißt: alle.
  gewaehlt: string
  waehlen: (marke: string) => void
}

export function Markenfilter({
  marken,
  gewaehlt,
  waehlen,
}: MarkenfilterEigenschaften) {
  const worte = useTexte().seiten.karten.marken

  // **Ohne Marken keine Reihe.** Ein Filter über nichts ist eine Zeile, die
  // erklärt werden müsste; ein Set ohne Marken zeigt seine Liste wie zuvor.
  if (marken.length === 0) {
    return null
  }

  return (
    <fieldset className="mb-2 flex flex-wrap items-center gap-1">
      <legend className="sr-only">{worte.filter}</legend>

      {['', ...marken].map((marke) => (
        <label
          key={marke}
          className="cursor-pointer rounded-full border border-kante px-2 py-0.5 text-xs has-checked:border-marke has-checked:bg-betont focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring"
        >
          <input
            type="radio"
            name="markenfilter"
            value={marke}
            checked={gewaehlt === marke}
            onChange={() => {
              waehlen(marke)
            }}
            // **`sr-only` und nicht `hidden`**: ein so entferntes Optionsfeld
            // ist auch für die Tastatur weg.
            className="sr-only"
          />

          {marke === '' ? worte.alle : marke}
        </label>
      ))}
    </fieldset>
  )
}
