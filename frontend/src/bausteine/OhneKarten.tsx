// **Ein Set ohne Karte sagt es, und der Knopf führt zu den Karten** — G02
// (#2249).
//
// Eine Lasche, die mit den Karten arbeitet — suchen, ersetzen, zählen, einen
// Stil zeigen, ein Pack ziehen, ausgeben —, hat in einem leeren Set nichts zu
// tun. Bis G02 bot sie trotzdem ihr Feld an, und erst das Ergebnis sagte,
// dass nichts da ist; der Weg zu den Karten stand nirgends.
//
// **`karten` kommt vom Aufrufer** und nicht aus dem Raum des offenen Sets:
// Ausgeben liest sein Set selbst, die übrigen Laschen über
// `useGelesenesSet`. **Solange es unbekannt ist, schweigt er** — ein Hinweis,
// der beim Laden aufblitzt, wäre falsch, sobald die Antwort kommt.
import { Knopf } from './Knopf'
import { Leerzustand } from './Leerzustand'
import { useTexte } from '../sprache/kontext'

export function OhneKarten({
  karten,
  erklaerung,
  wechseln,
}: {
  karten: number | undefined

  // erklaerung sagt, was hier ohne Karten fehlt — ohne sie der allgemeine
  // Satz.
  erklaerung?: string
  wechseln: (seite: 'karten') => void
}) {
  const worte = useTexte().allgemein.arbeitsplatz

  if (karten !== 0) {
    return null
  }

  return (
    <div className="mb-4">
      <Leerzustand
        titel={worte.ohneKartenTitel}
        erklaerung={erklaerung ?? worte.ohneKarten}
      >
        <Knopf
          aufKlick={() => {
            wechseln('karten')
          }}
        >
          {worte.zuDenKarten}
        </Knopf>
      </Leerzustand>
    </div>
  )
}
