// Einen Umschlag ansehen und dann übernehmen — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Warum zwei Schritte und nicht ein Knopf
//
// **Weil #1457 die fachliche Übernahmefreigabe verlangt**: „Import zeigt
// Objekte/Abhängigkeiten/Konflikte und benötigt fachliche
// Übernahmefreigabe." Was hier hereinkommt, hat jemand anderes geschrieben —
// und §16 Regel 6 sagt es allgemein: *nichts wird still übernommen.*
//
// # Und warum der Abdruck einen Satz danebenbekommt
//
// **Weil eine Kennung allein aussieht wie ein Nachweis.** Sie sagt, **wer**
// unterschrieben hat, nicht, ob man ihm trauen sollte. Ohne diesen Satz läse
// sich eine Zeile „Unterschrieben von k51q…" wie eine Unbedenklichkeit.
import { type Umschlaglauf } from '../haken/useUmschlag'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface UmschlagannahmeEigenschaften {
  // umschlag ist der Haken der **Seite** — derselbe, der auch die Knöpfe an
  // den Setkarten bedient.
  //
  // **Einer und nicht zwei**: §7 Regel 4 gilt je Datensatz, und ein zweiter
  // Haken hier hieße eine zweite Vorgangsliste — wer gleichzeitig schnürt und
  // hereinnimmt, bekäme zwei Läufe, die nichts voneinander wissen.
  umschlag: Umschlaglauf
}

export function Umschlagannahme({ umschlag }: UmschlagannahmeEigenschaften) {
  const texte = useTexte()

  const angesehen = umschlag.angesehen
  const laeuftWahl = umschlag.laeuft('umschlag-waehlen')
  const laeuftAnnahme = umschlag.laeuft('umschlag-uebernehmen')

  const knopf =
    'rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'

  // **Ein eigener Name und keine Zeichenkette mit Platzhalter.** Ein
  // `` `mt-3 ${knopf}` `` im JSX sieht für `keintextimjsx` aus wie ein Text
  // mit eingesetztem Wert — und dieser Wächter hat recht, wenn er nicht
  // unterscheiden kann, ob dort eine Klasse oder ein Satz steht.
  const ersterKnopf = 'mt-3 ' + knopf

  return (
    <section className="mt-6 border-t border-kante pt-4">
      <h2 className="text-lg font-semibold">
        {texte.seiten.sets.umschlagTitel}
      </h2>

      <Erklaerungsabsatz
        text={texte.seiten.sets.umschlagErklaerung}
        klasse="mt-1 max-w-prose text-leise"
      />

      {/* **Auch während einer Übernahme wird nicht gewählt** — Befund 4 des
          Zweitblicks über Phase 5.4. Der Knopf sperrte nur gegen sich
          selbst; wer während der Übernahme einen zweiten Umschlag wählte,
          bekam ihn kurz zu sehen, und die verspätete Antwort der Übernahme
          nahm ihn wieder weg. Der Zähler in `useUmschlag` fängt das ab —
          hier steht, dass es gar nicht erst dazu kommt (§7, Regel 4). */}
      <button
        type="button"
        disabled={laeuftWahl || laeuftAnnahme}
        onClick={umschlag.waehlen}
        className={ersterKnopf}
      >
        {texte.seiten.sets.umschlagWaehlen}
      </button>

      {/* **Hier stehen die Meldungen der Annahme.** Was das Ablegen meldet,
          steht seit UXR-04 (#1912) an der Setkarte, aus der es kam — vorher
          stand es nur hier, und dieser Kasten war zugeklappt, also gar nicht
          gezeichnet. `fuer` sagt, wem ein Fehler gehört. */}
      {umschlag.fehler !== undefined && umschlag.fuer === undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {umschlag.fehler}
        </p>
      )}

      {umschlag.uebernommen !== undefined && (
        <p role="status" className="mt-3 max-w-prose text-leise">
          {texte.seiten.sets.umschlagUebernommen}
        </p>
      )}

      {angesehen !== undefined && (
        <div className="mt-4 rounded-lg border border-kante bg-flaeche p-4">
          {angesehen.urheber !== '' && (
            <>
              <p className="text-sm">
                <span className="text-leise">
                  {`${texte.seiten.sets.umschlagUrheber}: `}
                </span>

                {/* **Der Abdruck bricht um und wird nicht abgeschnitten.**
                    Er ist das, was zwei Menschen vergleichen; eine Ellipse
                    darin machte den Vergleich wertlos. */}
                <span className="break-all font-mono">
                  {angesehen.fingerabdruck}
                </span>
              </p>

              <Erklaerungsabsatz
                text={texte.seiten.sets.umschlagAbdruckHinweis}
                klasse="mt-1 max-w-prose text-sm text-leise"
              />
            </>
          )}

          <h3 className="mt-3 text-sm font-semibold">
            {texte.seiten.sets.umschlagInhalt}
          </h3>

          <ul className="mt-1 space-y-1 text-sm">
            {angesehen.objekte.map((objekt) => (
              <li key={objekt.bezeichnung} className="flex gap-2">
                <span className="text-leise">{objekt.art}</span>
                <span className="truncate">{objekt.bezeichnung}</span>
              </li>
            ))}
          </ul>

          <div className="mt-4 flex flex-wrap gap-2">
            <button
              type="button"
              disabled={laeuftAnnahme}
              onClick={() => {
                umschlag.uebernehmen(angesehen.pfad)
              }}
              className={knopf}
            >
              {texte.seiten.sets.umschlagUebernehmen}
            </button>

            <button
              type="button"
              disabled={laeuftAnnahme}
              onClick={umschlag.raeumen}
              className={knopf}
            >
              {texte.seiten.sets.umschlagVerwerfen}
            </button>
          </div>
        </div>
      )}
    </section>
  )
}
