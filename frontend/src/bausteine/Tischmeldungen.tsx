// **Was der Tisch gerade sagt** — eine Absage und was noch wartet. Aus
// `Partietisch.tsx` herausgelöst, als P6
// (#2374)
// ihn über die Zeilengrenze aus §2 brachte.
//
// # Eine Absage steht neben dem Tisch
//
// **Eine abgewiesene Handlung wird hier erklärt und nicht am Kopf der
// Seite** — PLAY-04.AK-04
// (#1712).
// Der Tisch füllt den Bildschirm; wer eine Karte ablegt, die der Kern
// abweist, sieht die Begründung sonst dort, wo er gerade nicht hinsieht.
// `Spielen.tsx` zeigt sie deshalb nur, solange kein Tisch offensteht — zwei
// gleiche Meldungen ließen jemanden zwei Fehler suchen (§6).
//
// **Zurückgenommen wird dabei nichts**, und das ist eine Entscheidung: der
// Tisch zeichnet nicht optimistisch
// ([0126](../../../docs/entscheidungen/0126-der-tisch-zeichnet-nicht-optimistisch.md)).
//
// # Was wartet, ist zu sehen
//
// P5 (#2373):
// gezeichnet wird trotzdem nur, was der Kern bestätigt hat (0126).
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Tischmeldungen({
  fehler,
  wartend,
}: {
  fehler: string | undefined
  wartend: number
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <>
      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      <p aria-live="polite" className="text-sm text-leise">
        {wartend > 0 &&
          einsetzen(eigene.wuenscheWarten, { zahl: String(wartend) })}
      </p>
    </>
  )
}
