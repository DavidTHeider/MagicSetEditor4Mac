// **Wie viel soll das Programm übernehmen?** — PLAY-02.AK-01 und AK-02
// (#1710).
//
// # Warum zwei Modi und nicht dreizehn Kästchen
//
// **Weil dreizehn Kästchen vor dem ersten Spiel eine Prüfung sind und kein
// Anfang.** Wer neu ist, wählt „so viel wie möglich" und fängt an; wer
// genauer weiß, was er will, stellt um und sieht dann die Bereiche.
//
// # Und warum hier nichts **zu**gewählt werden kann
//
// **Weil das eine Zusage wäre, die niemand hält.** Was dieses Programm
// durchsetzt, sagt die Matrix darunter; ein Kästchen, das einen Bereich auf
// „automatisch" stellte, ohne dass Code dahinter stünde, sähe aus wie ein
// Beleg (§10). Abwählen dagegen ist immer einlösbar.
//
// # Und warum die Gruppen eingeklappt stehen
//
// **Weil dreizehn Kästchen auch nach dem Umstellen eine Wand sind** — UXR-13
// (#1921).
// Neben den Feldern der Sitze waren es bei zwei Sitzen und zwei Profilen 21
// Felder auf der Seite (R3 in `docs/UIREVIEW.md` lässt zwölf zu). Jetzt steht
// je Gruppe ein Knopf mit ihrem Namen, und wer in ihr wählt, klappt sie auf.
// Was abgewählt ist, sagt die Kurzschau darunter, auch bei zugeklappter
// Gruppe.
import { Aufklappkasten } from './Aufklappkasten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { regelbereichwort, regelgruppenwort } from '../seiten/regelwoerter'

import type { Regelgruppe } from '../apiregeln'

interface Regelwahleigenschaften {
  // modus ist `anfaenger` oder `frei`.
  modus: string
  setzeModus: (modus: string) => void

  abgewaehlt: readonly string[]
  setzeAbgewaehlt: (abgewaehlt: string[]) => void

  // baum kommt aus dem Kern — **nicht** aus dieser Datei (§16).
  baum: readonly Regelgruppe[]
}

// Anfaengermodus und Freiermodus sind die zwei Kennungen des Kerns.
export const Anfaengermodus = 'anfaenger'
export const Freiermodus = 'frei'

export function Regelwahlfeld({
  modus,
  setzeModus,
  abgewaehlt,
  setzeAbgewaehlt,
  baum,
}: Regelwahleigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  return (
    <section className="flex max-w-prose flex-col gap-2">
      <h3 className="text-base font-semibold">{eigene.modusTitel}</h3>

      <div className="flex flex-wrap gap-4">
        {[
          { kennung: Anfaengermodus, wort: eigene.modusAnfaenger },
          { kennung: Freiermodus, wort: eigene.modusFrei },
        ].map((wahl) => (
          <label key={wahl.kennung} className="flex items-center gap-2 text-sm">
            <input
              type="radio"
              name="regelmodus"
              className={kaestchenstil}
              checked={modus === wahl.kennung}
              onChange={() => {
                // **Der Wechsel in den Anfängermodus räumt die Abwahl mit**
                // (§7, Regel 1): sie gehört zu einem Stand, den es nicht mehr
                // gibt — und der Kern wiese sie ohnehin ab.
                setzeModus(wahl.kennung)

                if (wahl.kennung === Anfaengermodus) {
                  setzeAbgewaehlt([])
                }
              }}
            />
            {wahl.wort}
          </label>
        ))}
      </div>

      <Erklaerungsabsatz
        text={eigene.modusErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {modus === Freiermodus &&
        baum.map((gruppe) => (
          <Aufklappkasten
            key={gruppe.kennung}
            titel={regelgruppenwort(eigene, gruppe.kennung)}
          >
            <fieldset className="flex flex-col gap-1">
              {/* **Die Legende bleibt für das Vorleseprogramm** — sichtbar
                  steht derselbe Name schon auf dem Knopf darüber. */}
              <legend className="sr-only">
                {regelgruppenwort(eigene, gruppe.kennung)}
              </legend>

              {gruppe.bereiche.map((bereich) => (
                <label
                  key={bereich}
                  className="flex items-center gap-2 text-sm"
                >
                  <input
                    type="checkbox"
                    className={kaestchenstil}
                    checked={abgewaehlt.includes(bereich)}
                    onChange={(ereignis) => {
                      setzeAbgewaehlt(
                        ereignis.target.checked
                          ? [...abgewaehlt, bereich]
                          : abgewaehlt.filter((name) => name !== bereich),
                      )
                    }}
                  />
                  {einsetzen(eigene.regelAbwaehlen, {
                    bereich: regelbereichwort(eigene, bereich),
                  })}
                </label>
              ))}
            </fieldset>
          </Aufklappkasten>
        ))}
    </section>
  )
}

// kaestchenstil ist derselbe wie an jedem anderen Kästchen dieses Fensters.
//
// **Mit sichtbarem Fokusring** (§6): ein Bedienelement, das man mit der
// Tastatur erreicht und dabei nicht sieht, ist nicht bedienbar — und ein
// Wächter zählt es nach (`internal/werkzeuge/fokusring_test.go`).
const kaestchenstil =
  'size-4 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'
