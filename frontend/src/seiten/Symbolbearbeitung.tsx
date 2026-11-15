// **Die Leiste über dem Editor, wenn er ein Symbol der Bibliothek bearbeitet**
// — DOM-02.AK-02 (#1697).
//
// Der Weg zurück zum Setsymbol und, nach dem Sichern, die Sets, die die
// vorige Fassung tragen — samt dem Angebot, sie zu übernehmen. **Übernommen
// wird nur auf Wunsch**: die Sets tragen eine Kopie, und wer ein Symbol
// umzeichnet, meint nicht jedes Set, in dem es einmal eingesetzt wurde.
import type {
  Bearbeitetes,
  Symbolbearbeitungsstand,
} from '../haken/useSymbolbearbeitung'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Symbolbearbeitung({
  stand,
  bearbeitet,
}: {
  stand: Symbolbearbeitungsstand
  bearbeitet: Bearbeitetes
}) {
  const eigene = useTexte().seiten.karten.symbolbibliothek
  const { veraltet } = stand

  return (
    <div className="flex flex-col gap-2 rounded-sm border border-kante p-3">
      <div>
        <Knopf aufKlick={stand.schliessen}>{eigene.zumSetsymbol}</Knopf>
      </div>

      {veraltet !== undefined &&
        (veraltet.length === 0 ? (
          <p className="text-sm text-leise">{eigene.keineVeraltet}</p>
        ) : (
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm">
              {einsetzen(eigene.veraltet, { sets: veraltet.join(', ') })}
            </p>
            <Knopf
              gewicht="haupt"
              angehalten={stand.laeuft}
              aufKlick={() => void stand.nachziehen(bearbeitet)}
            >
              {eigene.nachziehen}
            </Knopf>
          </div>
        ))}

      {stand.hinweis !== undefined && (
        <p role="status" className="text-sm text-leise">
          {stand.hinweis}
        </p>
      )}
      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}
    </div>
  )
}
