// **Anmeldung über einen Anbieter** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Was diese Fläche verspricht — und was nicht
//
// **Sie sagt, was sie braucht.** Dieses Programm liefert keine Zugangsdaten
// mit; wer sich anmelden will, registriert seine eigene Anwendung beim
// Anbieter (Vermerk 0066). Steht keine Kennung, sagt der Knopf es beim Druck
// mit dem Satz, der dazugehört — statt einen Browser auf eine Adresse zu
// schicken, die der Anbieter abweist.
//
// **Und sie verspricht keinen zweiten Faktor.** Er kommt in Werkstück 7; bis
// dahin steht hier nichts darüber (§10: auch eine Ablehnung braucht Deckung).
import { useRef, useState } from 'react'

import { api, rufAbbrechen } from '../fassade'
import { auspacken } from '../umschlag'
import { Knopf } from '../bausteine/Knopf'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'

import type { AnbieterkontoDTO } from '../api'
import type { AnmeldeanbieterDTO } from '../apinutzer'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface Eigenschaften {
  // anbieter sind die Anbieter, über die man sich anmelden kann.
  anbieter: readonly string[]

  // konten sind die Konten des gezeigten Nutzers.
  konten: readonly AnbieterkontoDTO[]

  // nutzerID ist der Nutzer, dessen Konten hier stehen.
  nutzerID: string

  // geaendert meldet nach oben, dass neu zu lesen ist.
  geaendert: () => void

  // staende sagen je Anbieter, ob er eingerichtet ist und wogegen sein Weg
  // geprüft ist — ID-05.AK-01 (#1704). Fehlt einer, steht nichts dazu.
  staende?: readonly AnmeldeanbieterDTO[]
}

export function Anmeldung({
  anbieter,
  konten,
  nutzerID,
  geaendert,
  staende = [],
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **Ein Vorgang je Anbieter** (§7, Regel 4): zwei Anmeldungen gleichzeitig
  // öffneten zwei Browser und zwei Horcher, und nur einer bekäme den Rückruf.
  const vorgang = useVorgang()

  const verbunden = (name: string) =>
    konten.find((eines) => eines.anbieter === name)

  // **Was läuft, lässt sich abbrechen** (BK-07d, #2118): der Ruf wartet auf
  // den Browser, und wer dort nicht weitermacht, wartete sonst bis zur Frist.
  // Wer selbst abbricht, bekommt dafür keine Fehlermeldung.
  const laufend = useRef(new Map<string, Promise<unknown>>())
  const abgebrochen = useRef(new Set<string>())

  const anmelden = async (name: string) => {
    setzeFehler(undefined)
    abgebrochen.current.delete(name)

    await vorgang.starte(name, async () => {
      const lauf = api().Anmelden(name)
      laufend.current.set(name, lauf)

      try {
        const antwort = await lauf

        auspacken(antwort, () => antwort.nutzer)
        geaendert()
      } catch (grund) {
        if (!abgebrochen.current.has(name)) {
          setzeFehler(meldungAus(grund, texte))
        }
      } finally {
        laufend.current.delete(name)
      }
    })
  }

  const abbrechen = (name: string) => {
    abgebrochen.current.add(name)
    rufAbbrechen(laufend.current.get(name))
  }

  const loesen = async (name: string) => {
    setzeFehler(undefined)

    await vorgang.starte(name, async () => {
      try {
        const antwort = await api().KontoLoesen(nutzerID, name)

        auspacken(antwort, () => undefined)
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  return (
    <section aria-label={eigene.anmeldung} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß. */}
      <Erklaerungsabsatz
        text={eigene.anmeldungErklaerung}
        klasse="max-w-prose text-leise"
      />

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}

      <ul className="flex max-w-xl flex-col gap-2">
        {anbieter.map((name) => {
          const konto = verbunden(name)

          // **Der Name, wie der Anbieter sich schreibt** — und der rohe, wenn
          // der Katalog ihn nicht kennt. Der Fall ist keiner von heute: der
          // Vermerk sagt, OpenAI wäre „eine Zeile, sobald es einen Weg gibt",
          // und bis der Katalog nachzieht, steht der rohe Name da statt
          // nichts.
          const gezeigt = eigene.anbieternamen[name] ?? name
          const stand = staende.find((einer) => einer.anbieter === name)

          return (
            <li key={name} className="flex items-center gap-3">
              <span className="min-w-28">{gezeigt}</span>

              {/* **Eingerichtet, und wogegen geprüft** (ID-05.AK-01): kein
                  Anbieterweg lief bisher gegen den echten Anbieter, und das
                  steht hier, statt dass ein Knopf es verschweigt. */}
              {stand !== undefined && (
                <span className="min-w-0 flex-1 text-sm text-leise">
                  {einsetzen(eigene.anbieterstand, {
                    eingerichtet: stand.eingerichtet
                      ? eigene.eingerichtet
                      : eigene.nichtEingerichtet,
                    geprueft: eigene.geprueft[stand.geprueft],
                  })}
                </span>
              )}

              {konto === undefined ? (
                <Knopf
                  gewicht="leise"
                  bezeichnung={einsetzen(eigene.anmeldenBei, {
                    anbieter: gezeigt,
                  })}
                  angehalten={vorgang.laeuft(name)}
                  aufKlick={() => {
                    void anmelden(name)
                  }}
                >
                  {eigene.anmelden}
                </Knopf>
              ) : null}

              {konto === undefined && vorgang.laeuft(name) && (
                <Knopf
                  gewicht="leise"
                  bezeichnung={einsetzen(eigene.anmeldungAbbrechenBei, {
                    anbieter: gezeigt,
                  })}
                  aufKlick={() => {
                    abbrechen(name)
                  }}
                >
                  {eigene.anmeldungAbbrechen}
                </Knopf>
              )}

              {konto === undefined ? null : (
                <>
                  <span className="text-sm text-leise">
                    {konto.anzeigename === ''
                      ? eigene.verbunden
                      : einsetzen(eigene.verbundenAls, {
                          name: konto.anzeigename,
                        })}
                  </span>

                  <Knopf
                    gewicht="leise"
                    bezeichnung={einsetzen(eigene.loesenVon, {
                      anbieter: gezeigt,
                    })}
                    angehalten={vorgang.laeuft(name)}
                    aufKlick={() => {
                      void loesen(name)
                    }}
                  >
                    {eigene.loesen}
                  </Knopf>
                </>
              )}
            </li>
          )
        })}
      </ul>
    </section>
  )
}
