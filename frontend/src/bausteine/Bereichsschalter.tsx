// **Bereiche und Funktionen** — OPS-10
// (#1908,
// Artboard *Einstellungen*): je Schalter eine Wahl, ein Satz, was er
// abschaltet, und seine Reife als Marke, wenn sie nicht *stabil* ist.
//
// **Die Liste kommt aus den Einstellungen** — `schalter.*` in der Reihe, die
// Go liefert —, die Reife aus `Schalter`. Ein elfter Schalter erscheint hier,
// ohne dass jemand diese Datei anfasst.
//
// **Ein Schalter und keine Auswahlliste** — UXR-19
// (#1927),
// Regel R7 aus `docs/UIREVIEW.md`: was sofort wirkt und zwei Werte hat, ist
// ein Schalter. Hier standen zehn Auswahllisten mit „An" und „Aus", jede
// mit einem Knopf „Zurücksetzen auf An" darunter — umlegen ist dasselbe.
// Herkunft, Ebene und Wirkzeitpunkt stehen weiter an jedem (UX-04.AK-02).
//
// **Der Schalter selbst ist ein Baustein** (`Schalter`) — UXR-F2 (#2008):
// jede Einstellung mit zwei Werten, die an und aus sind, zeichnet ihn.
import { useSchalterreife } from '../haken/useSchalterreife'
import { useTexte } from '../sprache/kontext'
import { Einstellungsherkunft } from './Einstellungsherkunft'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Schalter } from './Schalter'
import type { EinstellungDTO } from '../api'

const vorsatz = 'schalter.'
const stabil = 'stabil'
const an = 'an'
const aus = 'aus'

const markenklassen =
  'rounded-full border border-marke px-2 py-0.5 text-xs text-marke'

export function Bereichsschalter({
  eintraege,
  angehalten,
  setze,
}: {
  eintraege: readonly EinstellungDTO[]
  angehalten: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.schalter
  const saetze: Record<string, string | undefined> = eigene.saetze
  const reifen: Record<string, string | undefined> = eigene.reife
  const namen: Record<string, string | undefined> =
    texte.seiten.einstellungen.namen
  const reife = useSchalterreife()

  return (
    <section
      aria-label={texte.seiten.einstellungen.gruppen.bereiche}
      className="flex flex-col gap-3"
    >
      <Erklaerungsabsatz text={eigene.erklaerung} />
      <Erklaerungsabsatz text={eigene.hinweis} />

      <div className="grid max-w-xl gap-5">
        {eintraege
          .filter((eintrag) => eintrag.schluessel.startsWith(vorsatz))
          .map((eintrag) => {
            const kennung = eintrag.schluessel.slice(vorsatz.length)
            const stufe = reife[kennung] ?? stabil

            return (
              <div key={eintrag.schluessel} className="flex flex-col gap-1">
                {/* **Die Marke steht neben dem Schalter, nicht in ihm** —
                    sonst hieße er „KI Vorschau", und ein Vorleseprogramm
                    sagte die Reife bei jedem Umlegen mit. */}
                <div className="flex items-center gap-2">
                  <label className="flex items-center gap-2 text-sm font-medium">
                    <Schalter
                      schluessel={eintrag.schluessel}
                      an={eintrag.wert === an}
                      angehalten={angehalten(eintrag.schluessel)}
                      aendern={(ein) => {
                        setze(eintrag.schluessel, ein ? an : aus)
                      }}
                    />
                    {namen[eintrag.schluessel] ?? eintrag.schluessel}
                  </label>
                  {stufe !== stabil && (
                    <span className={markenklassen}>
                      {reifen[stufe] ?? stufe}
                    </span>
                  )}
                </div>
                <p className="ml-6 text-sm text-leise">
                  {saetze[kennung] ?? ''}
                </p>
                {/* **Ohne Beschriftung des wirksamen Werts**: ein Schalter
                    kennt weder „erben“ noch eine Richtlinie, also gilt, was
                    er zeigt (`domain.Verschmolzen`) — der Satz darüber kommt
                    hier nicht vor (AP-11, #2368). */}
                <div className="ml-6">
                  <Einstellungsherkunft eintrag={eintrag} />
                </div>
              </div>
            )
          })}
      </div>
    </section>
  )
}
