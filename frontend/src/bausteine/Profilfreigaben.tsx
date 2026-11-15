// **Wer an diesem Profil was darf** — Werkstück 11 von Phase 4.12
// (#1351).
//
// # Warum die eigene Rolle die Seite bestimmt
//
// **Weil ein Knopf, den man nicht drücken darf, eine Zumutung ist.** Wer kein
// Verwalter ist, bekommt die Liste zu sehen und kein Formular — und den Satz
// dazu, warum.
//
// # Warum `leser` erklärt wird
//
// **Weil das Wort mehr verspricht, als es hält.** Ein Leser sieht, **dass** es
// das Profil gibt, nicht was darin steht; wer das nicht dazuschreibt, vergibt
// eine Rolle, die er sich anders vorstellt
// ([0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).
import { Satz } from './Satz'
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Knopf } from './Knopf'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { NutzerDTO } from '../api'

// rollen sind die drei, aufsteigend.
const rollen = ['leser', 'mitarbeiter', 'verwalter'] as const

interface Eigenschaften {
  profilID: string

  // nutzer sind alle des Bestands — die Auswahl baut daraus.
  nutzer: readonly NutzerDTO[]

  // eigner ist der Nutzer, dem dieses Profil gehört — er steht nicht zur Wahl.
  eigner: string
}

export function Profilfreigaben({ profilID, nutzer, eigner }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  // **Beim Profilwechsel geräumt** (§7, Regel 1).
  const geladen = useLaden(async () => {
    const antwort = await api().Freigaben(profilID)

    return auspacken(antwort, () => antwort)
  }, [profilID])

  // **Ein Vorgang je Nutzer** (§7, Regel 4).
  const vorgang = useVorgang()

  const [gewaehlt, setzeGewaehlt] = useState('')
  const [rolle, setzeRolle] = useState<string>(rollen[0])
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const fuehreAus = (schluessel: string, arbeit: () => Promise<void>) =>
    vorgang.starte(schluessel, async () => {
      setzeFehler(undefined)

      try {
        await arbeit()

        // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3).
        geladen.nachladen()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  const vergeben = () =>
    fuehreAus('neu', async () => {
      const antwort = await api().FreigabeSetzen({
        profilID,
        nutzerID: gewaehlt,
        rolle,
      })

      auspacken(antwort, () => antwort)
      setzeGewaehlt('')
    })

  const zuruecknehmen = (nutzerID: string) =>
    fuehreAus(nutzerID, async () => {
      const antwort = await api().FreigabeLoeschen(profilID, nutzerID)

      auspacken(antwort, () => antwort)
    })

  const stand = geladen.daten
  const darfVergeben = stand?.eigeneRolle === 'verwalter'

  // **Ein unbekannter Rollenname wird gezeigt, nicht verschluckt** (§16:
  // Drift wird erkannt, nicht erlitten). Bekäme das Fenster aus einem
  // neueren Bestand eine vierte Rolle, stünde sie roh da statt gar nicht.
  //
  // **An einer Stelle**, weil sie sonst zweimal danebenstünde — und der
  // zweite Rückfall wäre der, den niemand betritt.
  const rollenname = (eine: string) => eigene.rollennamen[eine] ?? eine

  // **Wer schon eine Freigabe hat, steht nicht noch einmal zur Wahl** — und
  // der Eigner auch nicht: er verwaltet es ohnehin.
  const offen = nutzer.filter(
    (einer) =>
      einer.id !== eigner &&
      !(stand?.freigaben ?? []).some((eine) => eine.nutzerID === einer.id),
  )

  return (
    <section aria-label={eigene.freigaben} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß; zwei
          gleiche Titel untereinander wären einer zu viel. */}
      <p className="max-w-prose text-leise">
        <Satz text={eigene.freigabenErklaerung} />
      </p>

      {stand !== undefined && !darfVergeben && (
        // **§6: ein Zustand erklärt sich, statt einen toten Knopf zu zeigen.**
        <p className="max-w-prose text-leise">{eigene.freigabenNurLesen}</p>
      )}

      {darfVergeben && offen.length > 0 && (
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">{eigene.freigabeNutzer}</span>
            <select
              value={gewaehlt}
              onChange={(ereignis) => {
                setzeGewaehlt(ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              <option value="">{eigene.freigabeNiemand}</option>
              {offen.map((einer) => (
                <option key={einer.id} value={einer.id}>
                  {einer.name}
                </option>
              ))}
            </select>
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">{eigene.freigabeRolle}</span>
            <select
              value={rolle}
              onChange={(ereignis) => {
                setzeRolle(ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {rollen.map((eine) => (
                <option key={eine} value={eine}>
                  {rollenname(eine)}
                </option>
              ))}
            </select>
          </label>

          <Knopf
            gewicht="haupt"
            angehalten={vorgang.laeuft('neu') || gewaehlt === ''}
            aufKlick={() => {
              void vergeben()
            }}
          >
            {eigene.freigabeVergeben}
          </Knopf>
        </div>
      )}

      {stand?.freigaben.length === 0 && (
        <p className="text-leise">{eigene.freigabenLeer}</p>
      )}

      <ul className="flex flex-col gap-2">
        {(stand?.freigaben ?? []).map((eine) => (
          <li key={eine.nutzerID} className="flex flex-wrap items-center gap-3">
            <span className="font-medium">{eine.name}</span>
            <span className="text-leise">{rollenname(eine.rolle)}</span>

            {darfVergeben && (
              <Knopf
                gewicht="leise"
                angehalten={vorgang.laeuft(eine.nutzerID)}
                aufKlick={() => {
                  void zuruecknehmen(eine.nutzerID)
                }}
              >
                {eigene.freigabeZurueckziehen}
              </Knopf>
            )}
          </li>
        ))}
      </ul>

      {(fehler ?? geladen.fehler) !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler ?? geladen.fehler}
        </p>
      )}
    </section>
  )
}
