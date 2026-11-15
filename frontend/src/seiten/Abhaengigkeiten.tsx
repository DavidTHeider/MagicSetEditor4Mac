// Die Abhängigkeiten der installierten Pakete (#57).
//
// **Ein Befund mit Vorschlag, kein Abbruch.** Was fehlt, lässt sich beziehen
// oder aus einer MSE2-Installation übernehmen — beides steht auf derselben
// Seite. Was zu alt ist, rendert stillschweigend falsch; deshalb steht es hier
// und nicht erst an der Karte.
//
// **Der Satz entsteht hier, aus Angaben von dort** (§9): das Backend gibt einen
// Schlüssel (`fehlt`, `zualt`, `unlesbar`, `zyklus`) und die Werte dazu.
import { type AbhaengigkeitDTO } from '../api'
import { abhaengigkeitssatz } from './abhaengigkeitssatz'
import { useTexte } from '../sprache/kontext'

// **Gelesen hat die Seite** (U14, #2159) — einmal für diesen Abschnitt und
// die Zahlen an den Paketzeilen ([useAbhaengigkeiten]).
interface Eigenschaften {
  daten: AbhaengigkeitDTO[] | undefined
  fehler: string | undefined
}

export function Abhaengigkeiten({ daten, fehler }: Eigenschaften) {
  const texte = useTexte()

  // **Ohne eigene Überschrift** — seit UXR-17 (#1925) steht der Abschnitt
  // eingeklappt unter dem Arbeitsplatz, und der Schalter des Kastens trägt
  // den Namen.
  return (
    <section>
      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {daten?.length === 0 && (
        <p className="max-w-prose text-leise">
          {texte.seiten.pakete.abhaengigAllesGut}
        </p>
      )}

      {daten !== undefined && daten.length > 0 && (
        <ul className="max-w-prose space-y-2">
          {daten.map((befund) => (
            <li
              key={`${befund.paket}/${befund.braucht}/${befund.art}`}
              className="text-text"
            >
              {abhaengigkeitssatz(befund, texte)}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
