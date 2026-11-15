// **Was dieser Prozess gewirkt hat** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Warum die Grenze dabeisteht
//
// Weil sie die Zusage ist. Eine Tabelle, die mit jeder Karte wächst, ist nach
// einem Tag Arbeit unlesbar; diese hier hat **vierzig** mögliche Zeilen, und
// das steht daneben, damit es niemand nachschlagen muss.
//
// # Und warum sie leer sein darf
//
// **Gezählt werden geprüfte Wirkungen**, nicht Rufe: wer nur liest, löst
// keine aus. Eine leere Tabelle heißt „nichts nach außen gewirkt" und ist die
// richtige Auskunft — kein Fehler und kein fehlender Wert.
import { useLaden } from '../haken/useLaden'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useTexte } from '../sprache/kontext'
import { Betriebszahlen } from './Betriebszahlen'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Wirkungszaehler() {
  const eigene = useTexte().seiten.technik

  const stand = useLaden(async () => {
    const antwort = await api().Metriken()

    return auspacken(antwort, () => antwort)
  }, [])

  const zahlen = stand.daten

  return (
    // **Ohne eigene Überschrift und Linie seit #2465**: die Zahlen stehen in
    // einem Abschnitt der Einstellungen, und dessen Knopf trägt den Namen.
    <section className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.metrikenErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {zahlen !== undefined && (
        <>
          <p className="text-sm text-leise">
            {eigene.metrikenGrenze.replace('{grenze}', String(zahlen.grenze))}
          </p>

          {zahlen.reihen.length === 0 ? (
            <p className="text-sm">{eigene.metrikenLeer}</p>
          ) : (
            <table className="w-full max-w-xl text-sm">
              <caption className="sr-only">{eigene.metriken}</caption>

              <thead>
                <tr className="text-left">
                  <th scope="col">{eigene.spalteFlaeche}</th>
                  <th scope="col">{eigene.spalteWirkung}</th>
                  <th scope="col">{eigene.spalteAusgang}</th>
                  <th scope="col" className="text-right">
                    {eigene.spalteAnzahl}
                  </th>
                </tr>
              </thead>

              <tbody>
                {zahlen.reihen.map((reihe) => (
                  <tr
                    key={`${reihe.flaeche}/${reihe.wirkung}/${reihe.ausgang}`}
                  >
                    <td>{wortZu(eigene.flaechen, reihe.flaeche)}</td>
                    <td>{wortZu(eigene.wirkungen, reihe.wirkung)}</td>
                    <td>{wortZu(eigene.ausgaenge, reihe.ausgang)}</td>
                    <td className="text-right tabular-nums">{reihe.anzahl}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {/* **Und wie der Prozess dabei läuft** (OPS-05, #1722) —
              Laufzeiten, Verbrauch, Zeitraum und was fehlt. */}
          <Betriebszahlen zahlen={zahlen} />
        </>
      )}
    </section>
  )
}

// wortZu sucht das deutsche oder englische Wort zu einem Bezeichner aus Go.
//
// **Übersetzt und nicht roh**, obwohl die Kommandozeile, MCP und der Dienst
// die Bezeichner selbst geben: die Oberfläche ist zweisprachig (§9), und der
// Betriebszustand eine Überschrift weiter oben tut dasselbe mit seinem Grund.
// Zwei Nachbarn, die dieselbe Sorte Wert verschieden zeigen, wären der
// Widerspruch, den die Sichtprüfung angestrichen hat.
//
// **Der Rückfall ist der Bezeichner selbst.** Eine Wirkungsklasse, die dieses
// Fenster noch nicht kennt, ist immer noch eine Auskunft; eine leere Zelle
// wäre keine.
function wortZu(woerter: Record<string, string>, bezeichner: string): string {
  return woerter[bezeichner] ?? bezeichner
}
