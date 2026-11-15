// **Ein Gegenüber aufnehmen und bestätigen, wo es gebraucht wird** —
// UX-03.AK-03 (#1675):
// *„Anlegen oder Bestätigen eines Kontakts ist aus dem Handels- und
// Spielworkflow möglich, ohne vorher eine technische Verwaltungsseite
// aufsuchen zu müssen."*
//
// # Dasselbe Bauteil wie in den Einstellungen
//
// **`Peerkontakte`, eingebettet** — mit Aufnehmen, Vergleich des Abdrucks,
// Sperren und Freigaben. Ein zweites, schmaleres Bauteil hätte zwei Wege zu
// derselben Bestätigung gegeben, und einer von beiden wäre der, den niemand
// nachführt.
//
// # Zu, bis jemand ihn öffnet
//
// **Weil er eine Frage an den Bestand ist** — dieselbe Überlegung wie bei der
// Liste in den Einstellungen, die erst auf Nachfrage kommt.
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { useTexte } from '../sprache/kontext'
import { Peerkontakte } from './Peerkontakte'

export function Gegenueberkasten({
  aufGeaendert,
}: {
  aufGeaendert?: (() => void) | undefined
}) {
  const texte = useTexte()

  return (
    <Aufklappkasten titel={texte.bausteine.gegenueberTitel}>
      <Peerkontakte eingebettet aufGeaendert={aufGeaendert} />
    </Aufklappkasten>
  )
}
