// **Über fremde Netze erreichbar sein** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum ein dritter Kasten und keine Einstellung am zweiten
//
// **Weil es eine andere Preisgabe ist.** Erreichbar sein gilt dem, der die
// Adresse kennt; sich ankündigen gilt jedem im selben Netz; eine Einladung
// gilt jedem, der sie bekommt — über ein Relais, das jemand anderes betreibt.
// Drei Entscheidungen, drei Schalter.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { inDieZwischenablage } from '../zwischenablage'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Fernwege } from './Fernwege'

import type { Fernantwort } from '../apipeerkontakt'

export function Fernhorchposten() {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [geaendert, setzeGeaendert] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  const [kopiert, setzeKopiert] = useState(false)

  const stand = useLaden<Fernantwort>(async () => {
    const antwort = await api().FernHorchStand()

    return auspacken(antwort, () => antwort)
  }, [geaendert])

  const laeuft = vorgang.laeuft('fernhorchposten')

  // **Ein Vorgang für beide Wege** (§7, Regel 4): sie gehen an denselben
  // Tunnel, und zwei Riegel ließen ein Beenden und ein Starten zugleich
  // laufen.
  const handeln = (weg: () => Promise<Fernantwort>): Promise<void> =>
    vorgang.starte('fernhorchposten', async () => {
      setzeFehler(undefined)
      setzeKopiert(false)

      try {
        auspacken(await weg(), () => undefined)

        // **Geschrieben ist geschrieben** (§7, Regel 3).
        setzeGeaendert((runde) => runde + 1)
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const daten = stand.daten
  const laufend = daten?.stand.laeuft === true ? daten.stand : undefined

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.fernTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.fernErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {laufend !== undefined ? (
        <div className="flex flex-col gap-2">
          <p className="max-w-prose text-sm text-leise">
            {eigene.fernWeitergeben}
          </p>

          <p className="font-mono text-sm break-all select-all">
            {laufend.einladung}
          </p>

          <div className="flex items-center gap-2">
            <Aktionsknopf
              aktion="kopieren"
              gewicht="leise"
              aufKlick={() => {
                // **Ein Wahrheitswert und keine Ablehnung.**
                // `inDieZwischenablage` fängt selbst und gibt `false`; ein
                // zweiter `then`-Zweig wäre toter Code — und **war** es:
                // die Abdeckung hat ihn angestrichen, nachdem ich ihn
                // geschrieben hatte.
                //
                // **Derselbe Satz wie bei den Werkzeugen** — es ist dieselbe
                // Sache, und zwei Schlüssel dafür hießen im Englischen
                // zweierlei (`sammelbegriffe.test.ts`).
                void inDieZwischenablage(laufend.einladung).then(
                  (geglueckt) => {
                    if (geglueckt) {
                      setzeKopiert(true)

                      return
                    }

                    setzeFehler(eigene.kopierfehler)
                  },
                )
              }}
            >
              {eigene.fernKopieren}
            </Aktionsknopf>

            {kopiert && (
              <span role="status" className="text-sm text-leise">
                {eigene.fernKopiert}
              </span>
            )}
          </div>

          <p className="max-w-prose text-sm text-leise">{eigene.fernStirbt}</p>

          <Fernwege
            wege={laufend.wege}
            angebunden={laufend.angebunden}
            neuLesen={() => {
              setzeGeaendert((runde) => runde + 1)
            }}
          />

          <Knopf
            gewicht="leise"
            angehalten={laeuft}
            aufKlick={() => {
              void handeln(() => api().FernHorchBeenden())
            }}
          >
            {eigene.fernBeenden}
          </Knopf>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          <p className="text-sm text-leise">{eigene.fernStill}</p>

          <Erklaerungsabsatz
            text={eigene.fernHinweis}
            klasse="max-w-prose text-sm text-leise"
          />

          <Knopf
            angehalten={laeuft}
            aufKlick={() => {
              void handeln(() => api().FernHorchStarten())
            }}
          >
            {eigene.fernStarten}
          </Knopf>
        </div>
      )}
    </section>
  )
}
