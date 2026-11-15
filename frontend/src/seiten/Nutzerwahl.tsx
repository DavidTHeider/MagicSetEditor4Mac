// **Den Menschen wählen, der arbeitet** — Werkstück 17 von Phase 4.12
// (#1400),
// nach Vermerk 0074.
//
// # Und seit UXR-20 die eine Liste der Menschen
//
// **Bis UXR-20** (#1928)
// stand darüber eine zweite, `Nutzerliste`: dieselben Menschen mit Bild und
// Profilen, aber ohne Handlung — und darunter in *Nutzer und Profile
// bearbeiten* noch einmal. Jetzt trägt jede Zeile hier das Bild samt dem
// Abzeichen der Rolle, in der gerade gearbeitet wird; die Profile stehen in
// der Profilwahl darüber, und wer keines hat, dem steht es hier.
//
// # Warum das neben der Profilwahl steht und nicht darin
//
// **Weil eine Geste nicht zwei Bedeutungen haben soll.** Die Wahl daneben
// heißt „ich sehe da hinein" und prüft **meine** Freigaben; diese heißt „ich
// bin jetzt jemand anderes" und verlangt **dessen** Ausweis. Wer beides an
// eine Liste hängte, machte aus einer abgelehnten Freigabe stillschweigend
// einen Identitätswechsel — und niemand sähe, wann er geschah.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Bindungsfehler } from '../bindungsfehler'
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import { Nachweisfeld } from '../bausteine/Nachweisfeld'
import { Nutzerbild } from '../bausteine/Nutzerbild'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useProfilwechsel } from '../profil'
import { useVorgang, zeigervorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'

import type { NutzerDTO } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface Eigenschaften {
  nutzer: readonly NutzerDTO[]

  // arbeitet ist die Kennung des Menschen, der gerade arbeitet.
  //
  // **Nicht der Eigner des gezeigten Profils** — wer hineingewechselt ist,
  // bleibt, wer er ist. Genau das ist der Inhalt dieses Werkstücks.
  arbeitet: string

  // bilder sind die Profilbilder als Daten-URI, je Nutzerkennung — wer
  // keines hat, steht nicht darin.
  bilder: Record<string, string>

  // badges sind die Abzeichen, je **Profil**kennung (Werkstück 9, #1351).
  badges: Record<string, string>

  // gezeigt ist das Profil, in dem gearbeitet wird.
  //
  // **Es entscheidet, welches Abzeichen über dem Gesicht liegt**: ein Mensch
  // hat mehrere Rollen, und zu sehen ist die, in der er gerade arbeitet.
  gezeigt: string
}

export function Nutzerwahl({
  nutzer,
  arbeitet,
  bilder,
  badges,
  gezeigt,
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // verlangt ist der Nutzer, für den gerade ein Ausweis gefragt wird.
  //
  // **Erst versuchen, dann fragen** (§6) — dieselbe Form wie bei der
  // Profilwahl: wer keinen Faktor führt, sieht kein Feld.
  const [verlangt, setzeVerlangt] = useState('')

  const gewechselt = useProfilwechsel()

  // **Ein Vorgang je Datensatz** (§7, Regel 4) — und der Datensatz ist der
  // **Zeiger**, nicht das Ziel. Der Grund steht in `Profilwahl.tsx`: ein
  // Nutzerwechsel schreibt Nutzer- **und** Profilzeiger, und zwei Wechsel auf
  // verschiedene Menschen schreiben dieselben zwei Zeilen.
  const vorgang = useVorgang()

  const wechseln = async (nutzerID: string, nachweis = '') => {
    setzeFehler(undefined)

    await vorgang.starte(zeigervorgang, async () => {
      try {
        const antwort = await api().NutzerWechseln(nutzerID, nachweis)

        auspacken(antwort, () => antwort.profil)

        // **Und die Anwendung baut sich neu auf** (§7, Regel 1): der
        // Profilzeiger zieht mit, und damit ist das Bezugsobjekt **alles**.
        setzeVerlangt('')
        gewechselt()
      } catch (grund) {
        if (
          grund instanceof Bindungsfehler &&
          grund.schluessel === 'zweiterFaktor'
        ) {
          setzeVerlangt(nutzerID)

          return
        }

        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  return (
    <section aria-label={eigene.nutzerwahl} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.nutzerwahl}</h2>

      <Erklaerungsabsatz
        text={eigene.nutzerwahlErklaerung}
        klasse="max-w-prose text-leise"
      />

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}

      {/* **Eine leere Liste erklärt, warum sie leer ist** (§6). Es gibt immer
          einen Menschen (Wanderung 0034); stünde trotzdem keiner da, sagt
          die Seite es, statt eine leere Fläche zu zeigen. */}
      {nutzer.length === 0 ? (
        <Leerzustand
          titel={eigene.niemand}
          erklaerung={eigene.profilErklaerung}
        />
      ) : (
        <ul className="flex max-w-xl flex-col gap-3">
          {nutzer.map((einer) => (
            <li key={einer.id} className="flex items-start gap-3">
              <Nutzerbild
                name={einer.name}
                datenURI={bilder[einer.id] ?? ''}
                badgeURI={
                  einer.profile.some((eines) => eines.id === gezeigt)
                    ? (badges[gezeigt] ?? '')
                    : ''
                }
              />

              {/* **Name, Stand und Handlung untereinander** — in der
                  schmalen Spalte neben den Laschen (UXR-20, #1928) passte
                  „Dieser Mensch werden" nicht neben den Namen, und ein
                  kurzer Name lief in den Knopf hinein. */}
              <div className="flex min-w-0 flex-1 flex-col items-start gap-1">
                <span className="font-medium">{einer.name}</span>

                {einer.profile.length === 0 && (
                  <span className="text-sm text-leise">
                    {eigene.keineProfile}
                  </span>
                )}

                {einer.id === arbeitet ? (
                  <span className="text-sm text-leise">{eigene.dasBinIch}</span>
                ) : (
                  <Knopf
                    gewicht="leise"
                    bezeichnung={einsetzen(eigene.werdenFuer, {
                      nutzer: einer.name,
                    })}
                    angehalten={vorgang.laeuft(zeigervorgang)}
                    aufKlick={() => {
                      void wechseln(einer.id)
                    }}
                  >
                    {eigene.werden}
                  </Knopf>
                )}

                {verlangt === einer.id && (
                  <Nachweisfeld
                    angehalten={vorgang.laeuft(zeigervorgang)}
                    abgeschickt={(nachweis) => {
                      void wechseln(einer.id, nachweis)
                    }}
                  />
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
