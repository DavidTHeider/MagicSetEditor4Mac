// **Die Trefferliste einer Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Die Herkunft steht am Treffer
//
// **Drei Angaben und nicht eine**: Revision, Gültigkeitsdatum und Abrufzeit
// sind drei verschiedene Dinge. Wann etwas geholt wurde, sagt nicht, ab wann
// es gilt.
//
// # Und ein Rang ist keine Wahrscheinlichkeit
//
// **Deshalb steht keine Prozentzahl daneben.** Das Issue sagt es als eigenen
// Satz: „Rang ist keine kalibrierte Wahrscheinlichkeit."
import type { RechercheAntwort, RecherchetrefferDTO } from '../apirecherche'
import { KIEinrichtungsverweis } from '../bausteine/KIEinrichtungsverweis'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'

interface TrefferEigenschaften {
  ergebnis: RechercheAntwort

  // oeffnen bringt einen Treffer aus diesem Bestand in den Karteneditor.
  //
  // **`undefined` ist ein Wert und kein fehlendes Feld**: neben einer Karte,
  // die schon offen ist, gibt es keinen Weg dorthin — und `exactOptional`
  // verlangt, dass das dasteht statt weggelassen zu werden.
  oeffnen: ((setID: string, karteID: string) => void) | undefined
}

export function Recherchetreffer({ ergebnis, oeffnen }: TrefferEigenschaften) {
  const eigene = useTexte().seiten.recherche

  const modus =
    {
      hybrid: eigene.rechercheModusHybrid,
      lexikalisch: eigene.rechercheModusLexikalisch,
      'ohne-index': eigene.rechercheModusOhneIndex,
    }[ergebnis.modus] ?? ergebnis.modus

  return (
    <>
      <p className="text-sm text-leise">
        {eigene.rechercheModus}: {modus}
      </p>

      {/* **Die Stände stehen je Quelle** — einer für alle wäre die Mischung,
          die A15 ausschließt. */}
      {ergebnis.korpusstaende.length > 0 && (
        <p className="text-xs text-leise">
          {eigene.rechercheStaende}:{' '}
          {ergebnis.korpusstaende
            .map((stand) => `${stand.quelle} ${stand.revision}`)
            .join(' · ')}
        </p>
      )}

      {ergebnis.einschraenkungen.map((satz) => (
        <p key={satz} role="status" className="text-xs text-leise">
          {eigene.rechercheEinschraenkung}: {satz}
        </p>
      ))}

      {ergebnis.treffer.length === 0 ? (
        <div className="flex flex-col items-start gap-1">
          <p className="max-w-prose text-sm text-leise">
            {leeresatz(ergebnis, eigene)}
          </p>
          {/* **Ohne Einbettungsmodell der Weg dorthin** (AI-01.AK-04,
              #1742) — der Fall, den das Issue nennt. Ein fehlender Index
              wird unter *Diagnose* gebaut, und der Satz sagt es. */}
          {ergebnis.leere === 'ohne-modell' && <KIEinrichtungsverweis />}
        </div>
      ) : (
        <ol className="divide-y divide-kante">
          {ergebnis.treffer.map((treffer) => (
            <Einzeltreffer
              key={treffer.kennung}
              treffer={treffer}
              oeffnen={oeffnen}
            />
          ))}
        </ol>
      )}
    </>
  )
}

// leeresatz sagt, **warum** nichts kam — und was dagegen zu tun ist.
//
// **Sechs Gründe, und nur einer heißt „es gibt die Antwort nicht".** Ein
// gemeinsames „Nichts gefunden" verschwiege fünfmal etwas, das sich beheben
// lässt.
//
// # Zwei Sätze, und sie sagen Verschiedenes
//
// **Der Dienst nennt den Fall, der Katalog den nächsten Schritt.** Deshalb
// steht der Grund **vorn**: „Es ist kein Suchindex aktiv." ist die Lage,
// „Unter »Technik« lässt sich einer bauen." der Rat. Beim ersten Anlauf
// standen beide Hälften in beiden Sätzen, und am Fenster stand zweimal
// dasselbe.
function leeresatz(
  ergebnis: RechercheAntwort,
  eigene: ReturnType<typeof useTexte>['seiten']['recherche'],
): string {
  const satz = {
    'nichts-gefunden': eigene.rechercheLeereNichtsGefunden,
    'ohne-index': eigene.rechercheLeereOhneIndex,
    'ohne-modell': eigene.rechercheLeereOhneModell,
    'ohne-quelle': eigene.rechercheLeereOhneQuelle,
    gesperrt: eigene.rechercheLeereGesperrt,
    budget: eigene.rechercheLeereBudget,
  }[ergebnis.leere]

  return satz === undefined ? ergebnis.grund : `${ergebnis.grund} ${satz}`
}

// Einzeltreffer zeigt einen Fund samt Herkunft.
function Einzeltreffer({
  treffer,
  oeffnen,
}: {
  treffer: RecherchetrefferDTO
  oeffnen: ((setID: string, karteID: string) => void) | undefined
}) {
  const eigene = useTexte().seiten.recherche

  return (
    <li className="flex flex-col gap-1 py-2">
      <span className="font-semibold">{treffer.titel}</span>

      {/* **Als Text und nicht als Auszeichnung** (Issue, Werkstück 3):
          Quelltexte sind unvertraute Daten. Ein Regeltext, der Markdown
          mitbrächte, bekäme hier keine Verlinkung. */}
      <span className="text-sm text-leise">
        {treffer.ausschnitt || treffer.text}
      </span>

      <span className="text-xs text-leise">
        {treffer.quelle} ·{' '}
        {einsetzen(eigene.rechercheHerkunft, {
          revision: treffer.quellrevision,
          datum: treffer.quelldatum,
          abgerufen: treffer.abgerufen,
        })}
      </span>

      {treffer.weitere.length > 0 && (
        <span className="text-xs text-leise">
          {einsetzen(eigene.rechercheWeitere, {
            anzahl: String(treffer.weitere.length),
          })}
        </span>
      )}

      {/* **Nur eine eigene Karte lässt sich öffnen.** Eine fremde Adresse
          wird **nicht** angeklickt: eine Trefferaktion lädt keine fremden
          URLs ohne geltende Netzfreigabe (Issue, Werkstück 4). */}
      {treffer.eigeneKarteID > 0 && oeffnen !== undefined && (
        <button
          type="button"
          onClick={() => {
            oeffnen('', String(treffer.eigeneKarteID))
          }}
          className="self-start text-sm text-marke underline underline-offset-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {eigene.rechercheOeffnen}
        </button>
      )}
    </li>
  )
}
