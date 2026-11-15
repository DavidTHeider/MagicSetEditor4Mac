// **Die Angaben zur Person** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 1 bis 5, 14 und 15.
//
// # Warum ein Formular und nicht elf Zeilen mit Stiften
//
// **Weil sie zusammen ausgefüllt werden.** Wer seinen Namen einträgt, trägt
// meist auch den Vornamen ein; elf einzeln bestätigte Zeilen wären elf
// Schreibvorgänge für einen Handgriff — und elf Gelegenheiten, an denen
// jemand anderes dazwischenschreibt.
//
// # Und warum es **einen** Nutzer bearbeitet
//
// **Den, dem das gezeigte Profil gehört.** Das ist der Mensch, der hier
// arbeitet; einen anderen zu ändern, ist eine Verwaltungsaufgabe und steht
// unter *Nutzer und Profile bearbeiten*.
//
// # Was hier nicht steht
//
// **Adresse, Rufnummer und Kennwort** gehören zur Anmeldung (Werkstück 6),
// die zweiten Faktoren zu ihrem eigenen Weg (Werkstück 7), das Bild zu seinem
// (Punkt 17). Ein Formular, das alles zugleich änderte, hätte für jedes Feld
// eine andere Zusage.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Beschriftetesfeld } from '../bausteine/Beschriftetesfeld'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { angabenAus } from './Nutzerpflege'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { NutzerangabenDTO, NutzerDTO } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// felder sind die acht freiwilligen, in Anzeigereihenfolge.
//
// **Als Liste und nicht als acht Blöcke**: sie unterscheiden sich in nichts
// als ihrem Namen, und acht Kopien desselben Bauteils wären acht Stellen, an
// denen eine Änderung ankommen muss.
const felder = [
  'anrede',
  'namenszusatz',
  'vorname',
  'weiterenamen',
  'nachname',
  'slogan',
  'kurzebio',
  'langebio',
] as const

interface Eigenschaften {
  nutzer: NutzerDTO

  // geaendert meldet nach oben, dass neu zu lesen ist.
  geaendert: () => void
}

export function Nutzerangaben({ nutzer, geaendert }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Der Entwurf entsteht einmal, und ein `key` räumt ihn ab.**
  //
  // # Warum kein Effekt, der ihn nachzieht
  //
  // **Weil ein `setState` im Effekt kaskadierende Zeichnungen auslöst** — der
  // React-Linter weist es ab, und er hat recht. Der Weg dieses Projekts ist
  // ein `key` am Bauteil (§7, „Regel 1 und 2 haben drei Gestalten"): er
  // steht in `Nutzer.tsx` und trägt Kennung **und Fassung**.
  //
  // **Beide, und die zweite ist der Punkt**: nach einem geglückten Schreiben
  // ist die Fassung eine andere, das Bauteil entsteht neu, und der Entwurf
  // beginnt bei dem, was jetzt im Bestand steht. Ohne sie schriebe der
  // nächste Ruf gegen einen Stand, den es nicht mehr gibt.
  const [entwurf, setzeEntwurf] = useState<NutzerangabenDTO>(angabenAus(nutzer))
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const vorgang = useVorgang()

  const sichern = () =>
    vorgang.starte(nutzer.id, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().NutzerAendern(entwurf)

        auspacken(antwort, () => antwort.nutzer)

        // **Nur nach einem geglückten Ruf** (§7, Regel 3): `auspacken` wirft
        // vorher, und ein Lesefehler überschriebe sonst die Meldung, die
        // sagt, was zu tun ist.
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  return (
    <section aria-label={eigene.angaben} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.angaben}</h2>

      <Erklaerungsabsatz
        text={eigene.angabenErklaerung}
        klasse="max-w-prose text-leise"
      />

      <div className="grid max-w-3xl gap-3 sm:grid-cols-2">
        <Pflichtfeld
          beschriftung={eigene.felder.name}
          wert={entwurf.name}
          angehalten={vorgang.laeuft(nutzer.id)}
          geaendert={(wert) => {
            setzeEntwurf((bisher) => ({ ...bisher, name: wert }))
          }}
        />

        {felder.map((feld) => (
          <Beschriftetesfeld
            key={feld}
            bezeichnung={eigene.felder[feld]}
            wert={entwurf[feld]}
            breite={24}
            setzeWert={(wert) => {
              setzeEntwurf((bisher) => ({ ...bisher, [feld]: wert }))
            }}
          />
        ))}
      </div>

      <div>
        {/* **Im Nebengewicht** — AP-3 (#2360): das Artboard *Nutzer* füllt
            nur „+ Profil anlegen“ im Kopf, und je Blick führt einer
            (`Zustaende09`). */}
        <Aktionsknopf
          aktion="sichern"
          angehalten={vorgang.laeuft(nutzer.id) || entwurf.name.trim() === ''}
          aufKlick={() => {
            void sichern()
          }}
        >
          {eigene.angabenSichern}
        </Aktionsknopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
