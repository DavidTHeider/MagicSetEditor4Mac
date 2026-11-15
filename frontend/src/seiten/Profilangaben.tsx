// **Die Angaben eines Profils** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 10 bis 13.
//
// # Warum es das **gezeigte** Profil bearbeitet
//
// **Weil das die Rolle ist, in der jemand gerade arbeitet.** Ein anderes zu
// beschreiben, ist eine Verwaltungsaufgabe und steht unter *Nutzer und
// Profile bearbeiten*; hier steht, wofür dieses Profil da ist.
//
// # Und warum das Abzeichen daneben liegt und nicht im Formular
//
// **Weil es eine Datei ist und kein Text.** Es geht über den Dateidialog,
// nicht über ein Eingabefeld — dieselbe Trennung wie beim Profilbild
// (Werkstück 8).
import { Satz } from '../bausteine/Satz'
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Beschriftetesfeld } from '../bausteine/Beschriftetesfeld'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Nutzerbild } from '../bausteine/Nutzerbild'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { meldungAus } from '../meldung'
import { profilangabenAus } from './Nutzerpflege'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { ProfilangabenDTO, ProfilDTO } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// felder sind die drei freiwilligen Texte, in Anzeigereihenfolge.
const felder = ['subslogan', 'kurzebeschreibung', 'langebeschreibung'] as const

interface Eigenschaften {
  profil: ProfilDTO

  // badgeURI ist das Abzeichen als Bild — leer heißt: keines.
  badgeURI: string

  // geaendert meldet nach oben, dass neu zu lesen ist.
  geaendert: () => void
}

export function Profilangaben({ profil, badgeURI, geaendert }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Der Entwurf entsteht einmal, und ein `key` räumt ihn ab** — derselbe
  // Weg wie bei [Nutzerangaben]: er trägt Kennung **und** Fassung, und damit
  // beginnt er nach einem geglückten Schreiben bei dem, was jetzt im Bestand
  // steht (§7, Regel 1 und 2 in einer Gestalt).
  const [entwurf, setzeEntwurf] = useState<ProfilangabenDTO>(
    profilangabenAus(profil),
  )

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **Ein Vorgang je Profil** (§7, Regel 4) — **einer** für das Formular und
  // das Abzeichen: beide schreiben denselben Datensatz, und getrennte ließen
  // zwei Schreibvorgänge gegen dieselbe Fassung laufen.
  const vorgang = useVorgang()

  const fuehreAus = (arbeit: () => Promise<void>) =>
    vorgang.starte(profil.id, async () => {
      setzeFehler(undefined)

      try {
        await arbeit()
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  const sichern = () =>
    fuehreAus(async () => {
      const antwort = await api().ProfilAendern(entwurf)

      auspacken(antwort, () => antwort.profil)
    })

  const badgeWaehlen = () =>
    fuehreAus(async () => {
      const antwort = await api().ProfilbadgeWaehlen(profil.id, profil.fassung)

      auspacken(antwort, () => antwort.profil)
    })

  const badgeEntfernen = () =>
    fuehreAus(async () => {
      const antwort = await api().ProfilbadgeEntfernen(
        profil.id,
        profil.fassung,
      )

      auspacken(antwort, () => antwort.profil)
    })

  return (
    <section aria-label={eigene.profilangaben} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.profilangaben}</h2>

      <Erklaerungsabsatz
        text={eigene.profilangabenErklaerung}
        klasse="max-w-prose text-leise"
      />

      <div className="grid max-w-3xl gap-3 sm:grid-cols-2">
        <Pflichtfeld
          beschriftung={eigene.profilfelder.name}
          wert={entwurf.name}
          angehalten={vorgang.laeuft(profil.id)}
          geaendert={(wert) => {
            setzeEntwurf((bisher) => ({ ...bisher, name: wert }))
          }}
        />

        {felder.map((feld) => (
          <Beschriftetesfeld
            key={feld}
            bezeichnung={eigene.profilfelder[feld]}
            wert={entwurf[feld]}
            breite={24}
            setzeWert={(wert) => {
              setzeEntwurf((bisher) => ({ ...bisher, [feld]: wert }))
            }}
          />
        ))}

        {/* **Technisch und Ablauf** — Werkstück 10 (#1351), Punkt 12. Sie
            stehen im selben Formular wie die Beschreibungen, weil sie
            **dasselbe** Schreiben teilen: eine zweite Schaltfläche liefe gegen
            dieselbe Fassung (§7, Regel 4). */}
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={entwurf.technisch}
            disabled={vorgang.laeuft(profil.id)}
            onChange={(ereignis) => {
              setzeEntwurf((bisher) => ({
                ...bisher,
                technisch: ereignis.target.checked,
              }))
            }}
            className="size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          <span className="text-sm font-medium">{eigene.technisch}</span>
        </label>

        <Beschriftetesfeld
          bezeichnung={eigene.ablauf}
          wert={entwurf.ablauf}
          breite={24}
          setzeWert={(wert) => {
            setzeEntwurf((bisher) => ({ ...bisher, ablauf: wert }))
          }}
        />
      </div>

      <Erklaerungsabsatz
        text={eigene.technischErklaerung}
        klasse="max-w-prose text-leise"
      />
      <p className="max-w-prose text-leise">
        <Satz text={eigene.ablaufErklaerung} />
      </p>

      <div>
        <Aktionsknopf
          aktion="sichern"
          gewicht="haupt"
          angehalten={vorgang.laeuft(profil.id) || entwurf.name.trim() === ''}
          aufKlick={() => {
            void sichern()
          }}
        >
          {eigene.profilangabenSichern}
        </Aktionsknopf>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        {/* **Das Abzeichen über einem leeren Gesicht** — hier zeigt es sich
            allein, und genau so liegt es später über dem Profilbild. */}
        <Nutzerbild name={profil.name} datenURI="" badgeURI={badgeURI} />

        <Knopf
          gewicht="leise"
          angehalten={vorgang.laeuft(profil.id)}
          aufKlick={() => {
            void badgeWaehlen()
          }}
        >
          {eigene.badgeWaehlen}
        </Knopf>

        {profil.badge !== '' && (
          <Aktionsknopf
            aktion="entfernen"
            gewicht="leise"
            angehalten={vorgang.laeuft(profil.id)}
            aufKlick={() => {
              void badgeEntfernen()
            }}
          >
            {eigene.badgeEntfernen}
          </Aktionsknopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
