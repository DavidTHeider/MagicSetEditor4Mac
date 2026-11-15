// **Kurve und Farbverteilung eines Decks** — Werkstück I von Phase 4.10
// (#1325).
//
// Eigene Datei, weil `Deckwerkbank.tsx` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, was mit einem Deck **geschieht**,
// hier, was es über sich **sagt**.
//
// # Warum Balken aus `div` und kein Diagramm
//
// Ein Balken ist eine Länge, und eine Länge ist ein `style`. Eine Bibliothek
// dafür kostet ein Bündel, eine Abhängigkeit und einen Lizenzvermerk — für
// zwei Listen mit je höchstens acht Zeilen.
//
// **Und die Zahl steht daneben, nicht nur im Balken.** Ein Balken ist die
// Form; die Zahl ist die Antwort. Wer die Seite mit einer Vorlesehilfe hört,
// bekommt sie so auch.
import {
  farbverteilung,
  kurve,
  type Farbe,
  type Kartenposten,
} from './deckrechnung'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface Auswertungseigenschaften {
  posten: readonly Kartenposten[]
}

export function Deckauswertung({ posten }: Auswertungseigenschaften) {
  const texte = useTexte().seiten.decks

  // **Die Farbnamen flach im Katalog.** Ein verschachtelter Verbund darin
  // bräche die Zusage `Record<string, string>`, an der `Deckurteil` hängt;
  // diese Zuordnung ist der Preis dafür und steht an **einer** Stelle.
  const farbname: Record<Farbe, string> = {
    weiss: texte.farbeWeiss,
    blau: texte.farbeBlau,
    schwarz: texte.farbeSchwarz,
    rot: texte.farbeRot,
    gruen: texte.farbeGruen,
  }

  const zeilen = kurve(posten)
  const farben = farbverteilung(posten)

  // **Kein Diagramm über null Karten.** Ein leerer Balkenkasten sähe aus wie
  // eine Aussage und wäre keine.
  if (zeilen.length === 0) {
    return null
  }

  const hoechste = Math.max(...zeilen.map(([, zahl]) => zahl), 1)

  return (
    <div className="flex flex-wrap gap-8">
      <section className="min-w-56 flex-1">
        <h3 className="font-medium">{texte.kurve}</h3>

        <Erklaerungsabsatz
          text={texte.kurveHinweis}
          klasse="mb-2 max-w-prose text-sm text-leise"
        />

        <ul aria-label={texte.kurve} className="flex flex-col gap-1">
          {zeilen.map(([wert, zahl]) => (
            <li
              key={wert}
              aria-label={String(wert)}
              className="flex items-center gap-2"
            >
              <span className="w-6 text-right tabular-nums text-leise">
                {wert}
              </span>

              <span
                aria-hidden="true"
                style={{ width: `${String((zahl / hoechste) * 100)}%` }}
                className="h-4 rounded-sm bg-marke"
              />

              <span className="tabular-nums">{zahl}</span>
            </li>
          ))}
        </ul>
      </section>

      {farben.length > 0 && (
        <section className="min-w-56 flex-1">
          <h3 className="font-medium">{texte.farben}</h3>

          <Erklaerungsabsatz
            text={texte.farbenHinweis}
            klasse="mb-2 max-w-prose text-sm text-leise"
          />

          <ul aria-label={texte.farben} className="flex flex-col gap-1">
            {farben.map(([farbe, zahl]) => (
              <li
                key={farbe}
                aria-label={farbname[farbe]}
                className="flex items-center gap-2"
              >
                <span className="w-20 text-leise">{farbname[farbe]}</span>

                <span
                  aria-hidden="true"
                  style={{
                    width: `${String(
                      (zahl / Math.max(...farben.map(([, n]) => n), 1)) * 100,
                    )}%`,
                  }}
                  className="h-4 rounded-sm bg-marke"
                />

                <span className="tabular-nums">{zahl}</span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}
