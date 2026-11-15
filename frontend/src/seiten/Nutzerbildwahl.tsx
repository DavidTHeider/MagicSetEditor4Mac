// **Ein Profilbild wählen und wieder wegnehmen** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Warum der Dateidialog in Go steht
//
// **Ein Fenster bekäme einen Namen und keinen Pfad** — dieselbe Begründung
// wie bei `BildSetzen` für ein Kartenbild. `NutzerbildWaehlen` öffnet den
// nativen Dialog und legt das Gewählte gleich ab; wer abbricht, bekommt eine
// leere Antwort und keinen Fehler.
//
// # Und warum das Entfernen daneben steht und nicht darunter
//
// **Weil es dieselbe Frage beantwortet.** Ein Bild hat zwei Zustände, und
// beide Handgriffe gehören an dieselbe Stelle — sonst sucht man den zweiten.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { meldungAus } from '../meldung'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'

import type { NutzerDTO } from '../api'

interface Eigenschaften {
  nutzer: NutzerDTO

  // geaendert meldet nach oben, dass neu zu lesen ist.
  geaendert: () => void
}

export function Nutzerbildwahl({ nutzer, geaendert }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **Ein Vorgang je Datensatz** (§7, Regel 4): ein zweiter Ruf, während der
  // erste läuft, ginge gegen eine Fassung, die es nicht mehr gibt.
  const vorgang = useVorgang()

  // fuehreAus lässt laufen, zeigt den Fehler, und lädt **nur nach einem
  // geglückten** Ruf neu (§7, Regel 3).
  const fuehreAus = (arbeit: () => Promise<void>) =>
    vorgang.starte(nutzer.id, async () => {
      setzeFehler(undefined)

      try {
        await arbeit()
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  const waehlen = () =>
    fuehreAus(async () => {
      const antwort = await api().NutzerbildWaehlen(nutzer.id, nutzer.fassung)

      auspacken(antwort, () => antwort.nutzer)
    })

  const entfernen = () =>
    fuehreAus(async () => {
      const antwort = await api().NutzerbildEntfernen(nutzer.id, nutzer.fassung)

      auspacken(antwort, () => antwort.nutzer)
    })

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="leise"
          angehalten={vorgang.laeuft(nutzer.id)}
          aufKlick={() => {
            void waehlen()
          }}
        >
          {eigene.bildWaehlen}
        </Knopf>

        {/* **Der Knopf steht nur da, wenn es etwas wegzunehmen gibt** — ein
            angehaltener Knopf ohne Anlass ist eine Frage, die niemand
            gestellt hat. */}
        {nutzer.bild !== '' && (
          <Aktionsknopf
            aktion="entfernen"
            gewicht="leise"
            angehalten={vorgang.laeuft(nutzer.id)}
            aufKlick={() => {
              void entfernen()
            }}
          >
            {eigene.bildEntfernen}
          </Aktionsknopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
