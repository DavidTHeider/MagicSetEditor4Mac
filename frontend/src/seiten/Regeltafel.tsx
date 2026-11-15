// **Nach welchen Regeln** — die zweite Tafel des Artboards *Eine Partie
// anfangen* (PLAY-02, S11.NAT-03 und S11.NAT-04).
//
// # Was hier steht und was darunter
//
// **Oben, was man vor dem Anfangen wissen muss**: wie viel das Programm
// übernimmt, was der Tisch tut (`Regelkurzschau`), was den Anfang anhält und
// was an einer Abwahl hängt (`Regelhinweise`), und dass „unterstützt" nicht
// „vollständig" heißt. **Darunter, eingeklappt, die langen Aufzählungen**
// (`Regelschau`): offen zwischen den Sitzen und dem Startknopf machten sie die
// Seite mit zwei gewählten Decks 4.890 Punkte hoch — gemessen am 2026-09-24
// (Befund F13).
//
// **Das Aufklappen ändert keine Eingabe** (S11.NAT-03): die Regelschau hat
// keine, und der Kasten sagt über `aria-expanded`, ob er offen ist.
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Flaeche } from '../bausteine/Flaeche'
import { Regelhinweise } from './Regelhinweise'
import { Regelkurzschau } from './Regelkurzschau'
import { Regelschau } from './Regelschau'
import { Anfaengermodus, Regelwahlfeld } from '../bausteine/Regelwahlfeld'
import { useTexte } from '../sprache/kontext'

import type { RegelfreigabeDTO } from '../apiregeln'

export function Regeltafel({
  freigabe,
  modus,
  setzeModus,
  abgewaehlt,
  setzeAbgewaehlt,
}: {
  freigabe: RegelfreigabeDTO
  modus: string
  setzeModus: (modus: string) => void
  abgewaehlt: string[]
  setzeAbgewaehlt: (abgewaehlt: string[]) => void
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <Flaeche
      ueberschrift={eigene.regelnTafel}
      marke={
        modus === Anfaengermodus ? eigene.modusAnfaenger : eigene.modusFrei
      }
    >
      <div className="flex flex-col gap-4">
        <Regelwahlfeld
          modus={modus}
          setzeModus={setzeModus}
          abgewaehlt={abgewaehlt}
          setzeAbgewaehlt={setzeAbgewaehlt}
          baum={freigabe.baum}
        />

        <Regelkurzschau freigabe={freigabe} />

        <Regelhinweise freigabe={freigabe} />

        {/* **Ein fester Hinweis und keine Meldung**: `Hinweiskasten` wäre
            eine Live-Region, und die sagte einem Vorleseprogramm bei jedem
            Zeichnen etwas, das sich nicht geändert hat. */}
        <p className="flex max-w-prose items-start gap-2 rounded-md border border-kante p-3 text-sm">
          <span aria-hidden="true">⚠</span>
          <span>
            <strong className="font-semibold">
              {eigene.unterstuetztTitel}
            </strong>{' '}
            {eigene.unterstuetztText}
          </span>
        </p>

        <Aufklappkasten titel={eigene.regelgrenzenTitel}>
          <Regelschau freigabe={freigabe} />
        </Aufklappkasten>
      </div>
    </Flaeche>
  )
}
