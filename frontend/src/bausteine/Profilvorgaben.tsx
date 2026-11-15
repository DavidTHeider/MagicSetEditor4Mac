// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Warum eine Auswahl und kein Textfeld
//
// **Weil der Wert eine Kennung ist.** Wer `7` tippt, weiß nicht, welches Deck
// er meint — und wer sich vertippt, bekommt ein Nein für etwas, das er nie
// wollte. Die Liste kommt mit der Antwort; sie ist dieselbe, gegen die danach
// geprüft wird.
//
// # Und warum die mehrwertigen Kästchen bekommen
//
// **Weil eine Mehrfachauswahl in einer Liste niemand bedient.** Ein
// `<select multiple>` verlangt Strg-Klick und zeigt nicht, was gewählt ist,
// sobald die Liste scrollt. Kästchen sagen es ohne Erklärung.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useState } from 'react'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { VorgabenstandDTO } from '../api'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'

interface Eigenschaften {
  profilID: string
}

export function Profilvorgaben({ profilID }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Beim Profilwechsel geräumt** (§7, Regel 1): `useLaden` ohne `behalten`
  // wirft den Stand des vorigen weg, statt ihn unter dem neuen Namen stehen zu
  // lassen.
  const geladen = useLaden(async () => {
    const antwort = await api().Profilvorgaben(profilID)

    return auspacken(antwort, () => antwort.vorgaben)
  }, [profilID])

  // **Ein Vorgang je Art** (§7, Regel 4): zwei Klicks auf dieselbe Art gingen
  // gegen einen Stand, den der Bestand noch nicht kennt.
  const vorgang = useVorgang()

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const setze = (art: string, werte: string[]) =>
    vorgang.starte(art, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().ProfilvorgabeSetzen({
          profilID,
          art,
          werte,
        })

        auspacken(antwort, () => antwort.vorgaben)

        // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3): geglückt ist
        // geglückt, und ein Lesefehler danach macht es nicht rückgängig.
        geladen.nachladen()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  return (
    <section aria-label={eigene.vorgaben} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß; zwei
          gleiche Titel untereinander wären einer zu viel. */}
      <Erklaerungsabsatz
        text={eigene.vorgabenErklaerung}
        klasse="max-w-prose text-leise"
      />

      {(geladen.daten ?? []).map((stand) => (
        <Vorgabenzeile
          key={stand.art}
          stand={stand}
          angehalten={vorgang.laeuft(stand.art)}
          setze={(werte) => {
            void setze(stand.art, werte)
          }}
        />
      ))}

      {(fehler ?? geladen.fehler) !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler ?? geladen.fehler}
        </p>
      )}
    </section>
  )
}

interface Zeileneigenschaften {
  stand: VorgabenstandDTO
  angehalten: boolean
  setze: (werte: string[]) => void
}

// Vorgabenzeile ist eine Art samt ihrer Wahl.
function Vorgabenzeile({ stand, angehalten, setze }: Zeileneigenschaften) {
  const eigene = useTexte().seiten.nutzer
  const name = eigene.vorgabenamen[stand.art] ?? stand.art

  return (
    <div className="flex flex-col gap-1">
      <span className="text-sm font-medium">{name}</span>

      {stand.moeglich.length === 0 ? (
        // **Ein leerer Zustand erklärt den nächsten Schritt** (§6) — „keine
        // Auswahl" sagte nur, dass etwas fehlt, nicht was.
        <p className="text-leise">{eigene.vorgabenLeer}</p>
      ) : stand.mehrwertig ? (
        <Kaestchen
          stand={stand}
          angehalten={angehalten}
          setze={setze}
          name={name}
        />
      ) : (
        <select
          aria-label={name}
          value={stand.werte[0] ?? ''}
          disabled={angehalten}
          onChange={(ereignis) => {
            setze(ereignis.target.value === '' ? [] : [ereignis.target.value])
          }}
          className="w-fit rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          <option value="">{eigene.vorgabenOhne}</option>
          {stand.moeglich.map((wahl) => (
            <option key={wahl.wert} value={wahl.wert}>
              {wahl.beschriftung}
            </option>
          ))}
        </select>
      )}

      {stand.verwaist.length > 0 && (
        // **Sie steht als Hinweis da und nicht als Fehler**: sie ist nichts,
        // was jemand falsch gemacht hat — worauf sie zeigte, ist weg.
        <p role="status" className="text-leise">
          {eigene.vorgabenVerwaist} {stand.verwaist.join(', ')}
        </p>
      )}
    </div>
  )
}

interface Kaestcheneigenschaften extends Zeileneigenschaften {
  name: string
}

// sichtbareWahlen ist, wie viele Kästchen ungefaltet dastehen (#2022) —
// Präferenzdecks und Lagerorte wachsen mit dem Bestand (R6).
const sichtbareWahlen = 8

// Kaestchen ist die Wahl einer mehrwertigen Art.
//
// **Gefaltet wird nur, was nicht gewählt ist** — ein gewähltes Deck steht
// immer da, sonst ließe es sich nicht abwählen. Die Folge bleibt die des
// Bestands; bei den Präferenzdecks sagt die Wahl ihre Reihenfolge selbst.
function Kaestchen({ stand, angehalten, setze, name }: Kaestcheneigenschaften) {
  const worte = useTexte().bausteine.listenansicht
  const [alle, setzeAlle] = useState(false)
  const gefaltet = !alle && stand.moeglich.length > sichtbareWahlen
  const gezeigt = gefaltet
    ? stand.moeglich.filter(
        (wahl, stelle) =>
          stelle < sichtbareWahlen || stand.werte.includes(wahl.wert),
      )
    : stand.moeglich

  return (
    <fieldset aria-label={name} className="flex flex-col gap-1">
      {gezeigt.map((wahl) => (
        <label key={wahl.wert} className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={stand.werte.includes(wahl.wert)}
            disabled={angehalten}
            onChange={(ereignis) => {
              // **Angehängt und nicht einsortiert**: die Reihenfolge ist bei
              // den Präferenzdecks die Aussage, und wer eines dazunimmt, meint
              // es als das nächste — nicht als das erste.
              setze(
                ereignis.target.checked
                  ? [...stand.werte, wahl.wert]
                  : stand.werte.filter((wert) => wert !== wahl.wert),
              )
            }}
            className="size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          <span>{wahl.beschriftung}</span>
        </label>
      ))}

      {gefaltet && (
        <div>
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              setzeAlle(true)
            }}
          >
            {einsetzen(worte.weitere, {
              anzahl: String(stand.moeglich.length - gezeigt.length),
            })}
          </Knopf>
        </div>
      )}
    </fieldset>
  )
}
