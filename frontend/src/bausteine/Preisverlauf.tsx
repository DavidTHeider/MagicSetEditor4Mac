// **Der Preisverlauf einer Reihe** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Eine Reihe je Währung und Ausführung
//
// Der Euro-Preis eines Drucks sagt nichts über seinen Foilpreis in Dollar. Wer
// beides in eine Reihe legte, hätte einen Verlauf, der zwischen zwei Größen
// springt — und niemand sähe, wo.
//
// # Warum eine Liste und kein Diagramm
//
// Weil ein Verlauf hier aus **wenigen** Punkten besteht: er entsteht, wenn
// jemand Preise holt, und das geschieht auf Anweisung. Ein Diagramm über drei
// Werte ist ein Diagramm über drei Werte.
import type { Preisreihe } from '../antworten'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { geldAusText } from '../geld'

interface PreisverlaufEigenschaften {
  reihe: Preisreihe
}

export function Preisverlauf({ reihe }: PreisverlaufEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const eigene = texte.seiten.sammlung

  // **Die Reihe geht als Ganzes hinein** (A3, #1442): sie hat sieben Teile,
  // und wer sie hier zerlegte, setzte sie drüben wieder zusammen.
  //
  // **Der Schlüssel ist die Reihe selbst** und nicht ihre Bestandteile einzeln
  // — sonst liefe ein Laden an, sobald irgendein Teil sich ändert, und zwei
  // Änderungen in einem Zug ergäben zwei Läufe.
  const schluessel = JSON.stringify(reihe)

  const stand = useLaden(async () => {
    const antwort = await api().Preisverlauf(reihe)

    return auspacken(antwort, () => antwort.staende)
  }, [schluessel])

  const staende = stand.daten ?? []

  return (
    <div className="w-full">
      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {staende.length === 0 && stand.fehler === undefined && (
        <p className="text-sm text-leise">{eigene.keinVerlauf}</p>
      )}

      {staende.length > 0 && (
        <ol className="flex flex-col gap-1">
          {staende.map((preis) => (
            <li key={preis.erhoben} className="text-sm text-leise">
              <span className="font-mono">{preis.erhoben}</span>{' '}
              <span className="font-medium">
                {geldAusText(preis.wert, preis.waehrung, darstellung)}
              </span>
            </li>
          ))}
        </ol>
      )}
    </div>
  )
}
