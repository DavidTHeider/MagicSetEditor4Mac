// **Das Profil wählen, in dem gearbeitet wird** — Werkstück 3 von Phase 4.12
// (#1351).
//
// # Warum das gezeigte Profil keinen Knopf trägt
//
// **Ein „Wechseln", das nichts wechselt, ist keine Handlung.** Bis UXR-20
// (#1928)
// stand daneben eine zweite Liste, `Nutzerliste`, die nur zeigte, wen es gibt;
// seitdem steht diese hier neben den Laschen — wie die linke Spalte des
// Artboards `Nutzer` —, und die Menschen mit Bild stehen in `Nutzerwahl`.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Bindungsfehler } from '../bindungsfehler'
import { Knopf } from '../bausteine/Knopf'
import { Nachweisfeld } from '../bausteine/Nachweisfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useProfilwechsel } from '../profil'
import { useVorgang, zeigervorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Faltliste } from '../bausteine/Faltliste'

import type { NutzerDTO } from '../api'

interface Eigenschaften {
  nutzer: readonly NutzerDTO[]

  // gezeigt ist die Kennung des Profils, in dem gearbeitet wird.
  gezeigt: string
}

export function Profilwahl({ nutzer, gezeigt }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // verlangt ist das Profil, für das gerade ein zweiter Faktor gefragt wird —
  // leer, wenn keiner gefragt wird.
  //
  // **Erst versuchen, dann fragen** (§6): wer keinen Faktor hat, sieht kein
  // Feld. Wer einen hat, sieht es genau dann, wenn die Anwendung ihn verlangt
  // — und nicht vorsorglich neben jeder Zeile.
  const [verlangt, setzeVerlangt] = useState('')

  const gewechselt = useProfilwechsel()

  // **Ein Vorgang je Datensatz** (§7, Regel 4): zwei Wechsel gleichzeitig
  // gingen gegen einen Stand, den der Bestand noch nicht kennt.
  //
  // **Und der Datensatz ist der ZEIGER, nicht das Ziel.** Bis zum 2026-09-06
  // war der Schlüssel die Ziel-Kennung — zwei schnelle Klicks auf
  // **verschiedene** Profile hatten damit zwei Schlüssel und liefen
  // nebeneinander, obwohl beide dieselbe Zeile schreiben. Angestrichen vom
  // Nachlauf zum Zweitblick über Phase 4.12.
  const vorgang = useVorgang()

  // **Flach, und das gezeigte zuerst** — die übrigen in der Folge der
  // Menschen, wie bisher.
  const alle = nutzer.flatMap((einer) =>
    einer.profile.map((eines) => ({ einer, eines })),
  )
  const profile = [
    ...alle.filter(({ eines }) => eines.id === gezeigt),
    ...alle.filter(({ eines }) => eines.id !== gezeigt),
  ]

  const wechseln = async (profilID: string, nachweis = '') => {
    setzeFehler(undefined)

    await vorgang.starte(zeigervorgang, async () => {
      try {
        const antwort = await api().ProfilWechseln(profilID, nachweis)

        auspacken(antwort, () => antwort.profil)

        // **Und die Anwendung baut sich neu auf** (§7, Regel 1). Warum das
        // hier das Richtige ist, steht in `profil.tsx`: beim Profil ist das
        // Bezugsobjekt **alles**.
        setzeVerlangt('')
        gewechselt()
      } catch (grund) {
        // **Ein fehlender zweiter Faktor ist keine Abweisung, sondern ein
        // fehlender Schritt.** Die Zeile klappt ihr Feld auf, statt eine
        // Meldung zu zeigen, mit der niemand etwas anfangen kann (§6).
        if (
          grund instanceof Bindungsfehler &&
          grund.schluessel === 'zweiterFaktor'
        ) {
          setzeVerlangt(profilID)

          return
        }

        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  return (
    <section aria-label={eigene.wahl} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.wahl}</h2>

      <Erklaerungsabsatz
        text={eigene.wahlErklaerung}
        klasse="max-w-prose text-leise"
      />

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}

      {/* **Das gezeigte Profil vorn, dazu drei, der Rest auf Wunsch** —
          #2022: mit sieben Profilen standen sieben „Hierhin wechseln" in der
          Spalte, und die Lasche daneben riss R2. */}
      <Faltliste
        eintraege={profile}
        vorn={4}
        schritt={10}
        klasse="flex max-w-xl flex-col gap-2"
        zeile={({ einer, eines }) => (
          <li
            key={eines.id}
            className="flex flex-wrap items-center gap-x-3 gap-y-1"
          >
            {/* **Das Profil vorn, der Mensch darunter** — wie im Artboard
                  `Nutzer` (UXR-20, #1928): gewählt wird eine
                  Arbeitsumgebung, und wem sie gehört, sagt die zweite
                  Zeile. In der schmalen Spalte neben den Laschen bricht der
                  Knopf darunter um, statt die Zeile zu verbreitern. */}
            <span className="flex min-w-0 flex-1 flex-col">
              <span className="font-medium">{eines.name}</span>
              <span className="text-sm text-leise">{einer.name}</span>
            </span>

            {eines.id === gezeigt ? (
              // **Kein Knopf, wo nichts zu tun ist** (§6). Ein „Wechseln",
              // das nichts wechselt, ist eine Handlung, die niemand
              // rückgängig machen kann, weil sie nie stattfand.
              <span className="text-sm text-leise">{eigene.hier}</span>
            ) : (
              <Knopf
                gewicht="leise"
                bezeichnung={einsetzen(eigene.hierhinFuer, {
                  nutzer: einer.name,
                  profil: eines.name,
                })}
                angehalten={vorgang.laeuft(zeigervorgang)}
                aufKlick={() => {
                  void wechseln(eines.id)
                }}
              >
                {eigene.hierhin}
              </Knopf>
            )}

            {verlangt === eines.id && (
              <Nachweisfeld
                angehalten={vorgang.laeuft(zeigervorgang)}
                abgeschickt={(nachweis) => {
                  void wechseln(eines.id, nachweis)
                }}
              />
            )}
          </li>
        )}
      />
    </section>
  )
}
