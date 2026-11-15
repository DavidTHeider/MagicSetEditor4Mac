// Vorlagen aus einer MSE2-Installation übernehmen (#55).
//
// **Der Hauptweg**, weil er der einzige ist, bei dem die Rechtslage klar ist:
// die Dateien liegen bereits auf dem Rechner der Nutzenden (§11).
//
// # Warum Vorschläge und kein Suchlauf
//
// Die üblichen Orte werden **vorgeschlagen, nicht ungefragt gelesen**. Ein
// Programm, das die Platte durchsucht, liest Verzeichnisse, die niemand ihm
// genannt hat. Wer keinen der Vorschläge nimmt, wählt selbst.
import { useCallback, useState } from 'react'

import { api, auspacken, type UebernahmeAntwort } from '../api'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { istAbgebrochen, meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Laufstand } from '../bausteine/Laufstand'

interface Eigenschaften {
  // uebernommen meldet, dass der Katalog neu zu laden ist.
  uebernommen: () => void
}

export function Uebernahme({ uebernommen }: Eigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().MSE2Vorschlaege()

    return antwort.orte
  }, [])
  const { daten } = useLaden<string[]>(holen, [])

  const [ergebnis, setzeErgebnis] = useState<UebernahmeAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  // **Eine Anzeige für zwei Quellen** — ein Verzeichnis von MSE2 und ein
  // `.mse-installer` geben dieselbe Antwort (FMT-02.AK-04, #1691), und ein
  // Vorgang zur Zeit gilt für beide (§7, Regel 4).
  const ausfuehren = async (ruf: () => Promise<UebernahmeAntwort>) => {
    await vorgang.starte('uebernahme', async () => {
      setzeFehler(undefined)
      setzeErgebnis(undefined)

      // **Wurde etwas angefasst?** Ein abgebrochener Dialog hat nichts
      // geschrieben; nach ihm nachzuladen wäre Arbeit über ein Nichts. Alles
      // andere — auch ein Fehler — kann die Platte verändert haben.
      let angefasst = true

      try {
        const antwort = await ruf()

        auspacken(antwort, () => undefined)

        if (antwort.abgebrochen) {
          angefasst = false

          return
        }

        setzeErgebnis(antwort)
      } catch (ursache: unknown) {
        // **Ein Anhalten ist keine Fehlermeldung** (FMT-03): die Pakete davor
        // sind ganz und vermerkt, und der Laufstand sagt es.
        if (!istAbgebrochen(ursache)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      } finally {
        // **Geschrieben ist geschrieben** (§7, Regel 3) — und deshalb im
        // `finally`: die Übernahme kopiert, bevor sie abgleicht. Scheitert der
        // Katalog danach, liegen die Pakete trotzdem da.
        if (angefasst) {
          uebernommen()
        }
      }
    })
  }

  const uebernehmen = (vorschlag: string) =>
    ausfuehren(() => api().VorlagenUebernehmen(vorschlag))

  // **Ohne eigene Überschrift** — seit UXR-17 (#1925) steht die Übernahme
  // in einem Kasten der Schublade *Paket beziehen*, und dessen Schalter trägt
  // den Namen; zwei Namen für dieselbe Sache sind einer zu viel.
  return (
    <section>
      <Erklaerungsabsatz
        text={texte.seiten.pakete.mse2Erklaerung}
        klasse="max-w-prose text-leise"
      />

      {daten !== undefined && daten.length > 0 && (
        <ul className="mt-3 space-y-2">
          {daten.map((ort) => (
            <li key={ort} className="flex flex-wrap items-center gap-2">
              <span className="font-mono text-sm">{ort}</span>
              <button
                type="button"
                disabled={vorgang.laeuft('uebernahme')}
                onClick={() => {
                  void uebernehmen(ort)
                }}
                aria-label={einsetzen(texte.seiten.pakete.mse2VonHier, { ort })}
                className="rounded-sm border border-kante px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              >
                {texte.seiten.pakete.mse2Uebernehmen}
              </button>
            </li>
          ))}
        </ul>
      )}

      {daten?.length === 0 && (
        <p className="mt-3 max-w-prose text-leise">
          {texte.seiten.pakete.mse2KeinVorschlag}
        </p>
      )}

      <button
        type="button"
        disabled={vorgang.laeuft('uebernahme')}
        onClick={() => {
          void uebernehmen('')
        }}
        className="mt-3 rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
      >
        {texte.seiten.pakete.mse2Waehlen}
      </button>

      <Erklaerungsabsatz
        text={texte.seiten.pakete.installerErklaerung}
        klasse="mt-4 max-w-prose text-leise"
      />

      <button
        type="button"
        disabled={vorgang.laeuft('uebernahme')}
        onClick={() => {
          void ausfuehren(() => api().InstallerEinlesen())
        }}
        className="mt-2 rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
      >
        {texte.seiten.pakete.installerWaehlen}
      </button>

      <Laufstand art="installer" />

      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {ergebnis && (
        <div role="status" className="mt-4">
          <p className="text-leise">
            {einsetzen(texte.seiten.pakete.mse2Ergebnis, {
              anzahl: String(ergebnis.uebernommen.length),
            })}
          </p>

          {ergebnis.uebergangen.length > 0 && (
            <>
              <p className="mt-2 max-w-prose text-leise">
                {texte.seiten.pakete.mse2Uebergangen}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {ergebnis.uebergangen.map((name) => (
                  <li key={name}>{name}</li>
                ))}
              </ul>
            </>
          )}

          {ergebnis.befunde.length > 0 && (
            <>
              <p className="mt-2 max-w-prose text-leise">
                {texte.seiten.pakete.katalogBefunde}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {ergebnis.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </section>
  )
}
