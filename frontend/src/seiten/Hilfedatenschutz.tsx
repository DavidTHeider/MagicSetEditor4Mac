// **Was diese Anwendung über dich weiß** — Werkstück 13 von Phase 4.12
// (#1351).
//
// # Warum es diese Seite gibt
//
// **Weil in der Hilfe „Sonst nichts. Keine Telemetrie, kein Konto, keine
// Anmeldung." stand.** Der Satz war wahr, als er geschrieben wurde; seit
// Phase 4.12 gibt es Konten, eine Anmeldung über vier fremde Anbieter, einen
// zweiten Faktor und ein Formular mit Anrede, Namensteilen und Biografien.
//
// # Eigene Datei
//
// Wie bei `Hilfeschnittstellen`: `Hilfe.tsx` ginge sonst über 400 Zeilen (§2).
//
// # Und ein Wächter hält den Text gegen den Quelltext
//
// `internal/werkzeuge/datenschutz_test.go` verlangt, dass **jedes**
// personenbezogene Feld aus `domain.Nutzer` und **jede** angesprochene Adresse
// hier vorkommt. Eine Datenschutzangabe, die niemand nachzählt, veraltet
// still — genau das ist hier passiert.
import { Flaeche } from '../bausteine/Flaeche'
import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'

// abschnitte sind die Paare aus Überschrift und Satz, in ihrer Reihenfolge.
//
// **Von innen nach außen**: was erfasst wird, wo es liegt, in welcher Form —
// und dann erst, was den Rechner verlässt. Wer wissen will, was hinausgeht,
// muss vorher wissen, worum es geht.
const abschnitte = [
  'was',
  'wo',
  'form',
  'aus',
  'nicht',
  'dienst',
  'ki',
  'auskunft',
] as const

export function Hilfedatenschutz() {
  const { hilfe } = useTexte()
  const eigene = hilfe.datenschutz

  return (
    <Flaeche ueberschrift={eigene.ueberschrift} sprungmarke="hilfe-datenschutz">
      <p className="mb-4 max-w-prose text-leise">
        <Satz text={eigene.einleitung} />
      </p>

      <dl className="grid gap-3 sm:grid-cols-[12rem_minmax(0,1fr)]">
        {abschnitte.map((schluessel) => (
          <div key={schluessel} className="contents">
            <dt className="font-semibold">{eigene[`${schluessel}Name`]}</dt>
            <dd className="max-w-prose text-leise">
              <Satz text={eigene[`${schluessel}Satz`]} />
            </dd>
          </div>
        ))}
      </dl>
    </Flaeche>
  )
}
