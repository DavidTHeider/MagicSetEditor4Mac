// **Den zweiten Faktor einrichten** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Zwei Schritte, und der zweite ist der wichtige
//
// **Ein TOTP-Faktor gilt erst, wenn ein Code aus ihm stimmt.** Wer ihn beim
// Anzeigen des Musters einträgt und den Scan vergisst, verlangt danach einen
// Code, den niemand erzeugen kann — er hätte sich selbst ausgesperrt.
//
// # Was einmal dasteht und nie wieder
//
// **Geheimnis und Wiederherstellungscodes.** Die Seite sagt es, bevor sie sie
// zeigt; wer sie wegklickt, richtet neu ein.
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Knopf } from '../bausteine/Knopf'
import { Nachweisfeld } from '../bausteine/Nachweisfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { Einrichtungsantwort, FaktorDTO, FaktorenAntwort } from '../api'

// Die Schlüssel der Vorgänge (§7, Regel 4).
const TOTP = 'totp'
const BIOMETRIE = 'biometrie'

export function Zweiterfaktor({ nutzerID }: { nutzerID: string }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.faktor

  const [frisch, setzeFrisch] = useState<Einrichtungsantwort | undefined>(
    undefined,
  )
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **Eine Zahl und kein Rückruf** — dieselbe Form wie bei der Nutzerliste:
  // wer neu laden will, zählt hoch, und der Effekt sieht es.
  const [runde, setzeRunde] = useState(0)

  // **Entfernt wird erst nach einer Frage** — UX-10.AK-04 (#1682): danach
  // verlangt die Anmeldung einen Faktor weniger. Bis zum 2026-09-25 genügte
  // ein Klick.
  const [fragt, setzeFragt] = useState<string>()

  const vorgang = useVorgang()

  // **Über `useLaden`, wie jeder Ladeeffekt dieses Fensters** (§7, Regel 2):
  // er zählt den Stand mit, aus dem eine Antwort stammt, und wirft eine
  // überholte weg. **Und er räumt beim Wechsel des Nutzers** (Regel 1) — die
  // Faktoren gehören ihm.
  const holen = useCallback(async () => {
    const antwort = await api().Faktoren(nutzerID)

    auspacken(antwort, () => antwort.faktoren)

    return antwort
  }, [nutzerID])

  const geladen = useLaden<FaktorenAntwort>(holen, [nutzerID], false, runde)
  const stand = geladen.daten

  const fuehre = async (schluessel: string, arbeit: () => Promise<void>) => {
    setzeFehler(undefined)

    await vorgang.starte(schluessel, async () => {
      try {
        await arbeit()

        // **Geschrieben ist geschrieben** (§7, Regel 3): das Nachladen ist
        // Bestenfalls-Arbeit und macht die Änderung nicht ungeschehen.
        setzeRunde((zahl) => zahl + 1)
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  const einrichten = () =>
    fuehre(TOTP, async () => {
      const antwort = await api().FaktorEinrichten(nutzerID)

      auspacken(antwort, () => antwort.adresse)
      setzeFrisch(antwort)
    })

  const bestaetigen = (code: string) =>
    fuehre(TOTP, async () => {
      const antwort = await api().FaktorBestaetigen(nutzerID, code)

      auspacken(antwort, () => undefined)
      setzeFrisch(undefined)
    })

  const biometrie = () =>
    fuehre(BIOMETRIE, async () => {
      const antwort = await api().BiometrieEinrichten(nutzerID)

      auspacken(antwort, () => undefined)
    })

  const entfernen = (art: string) =>
    fuehre(art, async () => {
      const antwort = await api().FaktorEntfernen(nutzerID, art)

      auspacken(antwort, () => undefined)
    })

  const hat = (art: string) =>
    stand?.faktoren.some((einer: FaktorDTO) => einer.art === art) === true

  return (
    <section aria-label={eigene.titel} className="flex flex-col gap-3">
      <h2 className="text-base font-medium">{eigene.titel}</h2>

      <Erklaerungsabsatz
        text={eigene.erklaerung}
        klasse="max-w-prose text-leise"
      />

      {geladen.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {geladen.fehler}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      <ul className="flex max-w-xl flex-col gap-2">
        <li className="flex items-center gap-3">
          <span className="min-w-48">{eigene.totp}</span>

          {hat(TOTP) ? (
            <Aktionsknopf
              aktion="entfernen"
              gewicht="leise"
              bezeichnung={eigene.totpEntfernen}
              angehalten={vorgang.laeuft(TOTP)}
              aufKlick={() => {
                setzeFragt(TOTP)
              }}
            >
              {eigene.entfernen}
            </Aktionsknopf>
          ) : (
            <Knopf
              gewicht="leise"
              bezeichnung={eigene.totpEinrichten}
              angehalten={vorgang.laeuft(TOTP)}
              aufKlick={() => {
                void einrichten()
              }}
            >
              {eigene.einrichten}
            </Knopf>
          )}
        </li>

        <li className="flex items-center gap-3">
          <span className="min-w-48">
            {stand?.biometrieart === undefined || stand.biometrieart === ''
              ? eigene.biometrie
              : stand.biometrieart}
          </span>

          {hat(BIOMETRIE) ? (
            <Aktionsknopf
              aktion="entfernen"
              gewicht="leise"
              bezeichnung={eigene.biometrieEntfernen}
              angehalten={vorgang.laeuft(BIOMETRIE)}
              aufKlick={() => {
                setzeFragt(BIOMETRIE)
              }}
            >
              {eigene.entfernen}
            </Aktionsknopf>
          ) : (
            <Knopf
              gewicht="leise"
              bezeichnung={eigene.biometrieEinrichten}
              angehalten={
                vorgang.laeuft(BIOMETRIE) || stand?.biometrie !== true
              }
              aufKlick={() => {
                void biometrie()
              }}
            >
              {eigene.einrichten}
            </Knopf>
          )}

          {/* **Warum es nicht geht, steht daneben** (§6) — ein ausgegrauter
              Knopf ohne Grund ist eine Ablehnung ohne Auskunft. */}
          {stand !== undefined && !stand.biometrie && (
            <span className="text-sm text-leise">{stand.biometriegrund}</span>
          )}
        </li>
      </ul>

      {/* **Nur für einen Faktor, den es noch gibt** — nach dem Entfernen
          steht die Frage nicht mehr da, auch wenn niemand sie geschlossen
          hat. */}
      {fragt !== undefined && hat(fragt) && (
        <Entscheidung
          frage={
            fragt === TOTP
              ? eigene.totpEntfernenFrage
              : eigene.biometrieEntfernenFrage
          }
          erklaerung={eigene.entfernenErklaerung}
          bestaetigen={eigene.entfernenJa}
          aktion="entfernen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={vorgang.laeuft(fragt)}
          aufBestaetigen={() => {
            setzeFragt(undefined)
            void entfernen(fragt)
          }}
          aufAbbrechen={() => {
            setzeFragt(undefined)
          }}
        />
      )}

      {frisch !== undefined && (
        <div className="flex max-w-prose flex-col gap-2 rounded-md border border-kante p-3">
          <p className="font-medium">{eigene.einmal}</p>

          <Erklaerungsabsatz
            text={eigene.einmalErklaerung}
            klasse="text-leise"
          />

          <code className="break-all text-sm">{frisch.geheimnis}</code>

          <p className="text-leise">
            {einsetzen(eigene.codesUeberschrift, {
              zahl: String(frisch.codes.length),
            })}
          </p>

          <ul className="font-mono text-sm">
            {frisch.codes.map((code) => (
              <li key={code}>{code}</li>
            ))}
          </ul>

          <p className="text-leise">{eigene.bestaetigen}</p>

          <Nachweisfeld
            angehalten={vorgang.laeuft(TOTP)}
            abgeschickt={(code: string) => {
              void bestaetigen(code)
            }}
          />
        </div>
      )}
    </section>
  )
}
