// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum die Liste erst auf Nachfrage kommt
//
// **Weil sie ein Ruf ins Netz ist.** Ein Kasten, der von selbst fragte, fragte
// jedes Mal, wenn jemand die Technikseite aufschlägt — und sagte dabei jedem im
// Netz, dass hier nach Gegenüber gesucht wird.
//
// # Und warum ein Fund kein Kontakt ist
//
// **Weil er nichts belegt.** Wer sich meldet, sagt eine Kennung und einen
// Namen; beides kann jeder sagen. Der Knopf trägt ihn deshalb als
// **unbestätigten** Kontakt ein — und der darf nichts, bis ein Mensch den
// Fingerabdruck verglichen hat.
import { useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { Diagnoseverweis } from './Diagnoseverweis'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Nachbar, Nachbarnantwort } from '../apipeerkontakt'

interface Nachbarneigenschaften {
  // aufGeaendert meldet, an welchem Gegenüber eben etwas geschrieben wurde —
  // aufgenommen oder umgezogen. **Im Kasten „Gegenüber aufnehmen oder
  // bestätigen"** (UXR-16, #1924) lädt die Liste darunter nach und schlägt
  // diese Zeile auf: wer eben jemanden aufgenommen hat, vergleicht als
  // Nächstes mit ihm.
  aufGeaendert?: ((kennung: string) => void) | undefined
}

// markenklasse ist die Gestalt der Marke neben dem Namen — dieselbe wie in
// der Liste der Gegenüber.
const markenklasse = 'rounded-full border border-kante px-2 text-xs text-leise'

export function Nachbarn({ aufGeaendert }: Nachbarneigenschaften = {}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()

  const [gefunden, setzeGefunden] = useState<Nachbarnantwort>()
  const [fehler, setzeFehler] = useState<string>()

  // ursache bleibt für „Diagnose öffnen" stehen — dort reist ihre
  // Vorgangskennung mit (NET-02, S13.NAT-04, #1706).
  const [ursache, setzeUrsache] = useState<unknown>()

  const laeuft = vorgang.laeuft('nachbarn')

  // handeln fährt einen Weg und trägt das Ergebnis ein.
  //
  // **Ein Vorgang für beide** (§7, Regel 4): suchen und aufnehmen gehen an
  // dieselbe Liste, und zwei Riegel ließen zwei zugleich laufen.
  //
  // **kennung** nennt das Gegenüber, an dem geschrieben wird — fehlt sie,
  // wird nur gesucht.
  const handeln = (
    weg: () => Promise<Nachbarnantwort>,
    kennung?: string,
  ): Promise<void> =>
    vorgang.starte('nachbarn', async () => {
      setzeFehler(undefined)
      setzeUrsache(undefined)

      try {
        const antwort = await weg()

        setzeGefunden(auspacken(antwort, () => antwort))

        if (kennung !== undefined) {
          aufGeaendert?.(kennung)
        }
      } catch (gescheitert) {
        setzeFehler(meldungAus(gescheitert, texte))
        setzeUrsache(gescheitert)
      }
    })

  return (
    <div className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.nachbarnErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <Knopf
        gewicht="leise"
        angehalten={laeuft}
        aufKlick={() => {
          void handeln(() => api().Nachbarn())
        }}
      >
        {laeuft ? eigene.nachbarnSucht : eigene.nachbarnSuchen}
      </Knopf>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      <Diagnoseverweis ursache={ursache} />

      {gefunden?.nachbarn.length === 0 && (
        <p className="text-sm text-leise">{eigene.nachbarnLeer}</p>
      )}

      {gefunden !== undefined && gefunden.nachbarn.length > 0 && (
        <ul className="flex flex-col gap-2">
          {gefunden.nachbarn.map((nachbar) => (
            <Nachbarzeile
              key={nachbar.kennung}
              nachbar={nachbar}
              laeuft={laeuft}
              aufnehmen={() => {
                void handeln(async () => {
                  await api().PeerkontaktAufnehmen(
                    nachbar.kennung,
                    nachbar.name,
                  )

                  // **Die Adresse kommt mit** — sie ist der Grund, warum
                  // dieser Weg mehr ist als ein Abtippen der Kennung.
                  //
                  // **Und zwar die zum Merken** (#1633): gewählt wird über
                  // die IP, aufgeschrieben wird der `.local`-Name, wo es
                  // einen gibt. Eine IP von heute ist nach dem nächsten
                  // DHCP-Wechsel falsch — und zeigte dann auf einen
                  // Rechner, der inzwischen jemand anderem gehört.
                  await api().PeerkontaktAdresse(
                    nachbar.kennung,
                    nachbar.merkadresse,
                  )

                  return api().Nachbarn()
                }, nachbar.kennung)
              }}
              uebernehmen={() => {
                void handeln(async () => {
                  // **Nur die Adresse** (NET-02.NAT-02, #1706) — der
                  // Stand bleibt, und die Leitung prüft weiter den
                  // Schlüssel hinter der Kennung.
                  await api().PeerkontaktAdresse(
                    nachbar.kennung,
                    nachbar.merkadresse,
                  )

                  return api().Nachbarn()
                }, nachbar.kennung)
              }}
            />
          ))}
        </ul>
      )}

      {/* **Was nicht zu deuten war, steht dabei** — sonst sähe „niemand
          gefunden" wie „nichts gehört" aus. */}
      {gefunden !== undefined && gefunden.verworfen > 0 && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.nachbarnVerworfen, {
            zahl: String(gefunden.verworfen),
          })}
        </p>
      )}
    </div>
  )
}

interface Nachbarzeileneigenschaften {
  nachbar: Nachbar
  laeuft: boolean
  aufnehmen: () => void

  // uebernehmen setzt die gefundene Adresse an einen bekannten Kontakt.
  uebernehmen: () => void
}

// Nachbarzeile zeigt einen Fund samt dem, was dieses Profil über ihn weiß.
//
// **Name, Stand und die eine Handlung vorn, das Technische dahinter** —
// UXR-16 (#1924),
// nach dem Artboard *Gegenüber*: dort steht ein Fund als Name mit seiner
// Marke, und Fingerabdruck und Adresse stehen unter „Technische Angaben".
// Bis dahin standen sie offen an jedem Fund — bei sechs Funden achtzehn
// Zeilen Kennziffern, und die Frage *wer ist da?* stand dazwischen.
function Nachbarzeile({
  nachbar,
  laeuft,
  aufnehmen,
  uebernehmen,
}: Nachbarzeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik

  return (
    <li className="flex flex-col gap-2 rounded-sm border border-kante p-3">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {/* **Ohne Namen der Fingerabdruck** — wie in der Liste der
            Gegenüber: eine Zeile ohne Namen ließe sich von der nächsten
            nicht unterscheiden. */}
        <span className="font-medium break-all">
          {nachbar.name || nachbar.fingerabdruck}
        </span>

        <span className={markenklasse}>
          {nachbar.bekannt
            ? standsatz(eigene, nachbar.stand)
            : eigene.nachbarnUnbekannt}
        </span>

        {!nachbar.bekannt && (
          <span className="ml-auto">
            <Knopf gewicht="leise" angehalten={laeuft} aufKlick={aufnehmen}>
              {eigene.nachbarnAufnehmen}
            </Knopf>
          </span>
        )}
      </div>

      {nachbar.name !== '' && (
        <Erklaerungsabsatz text={eigene.kontakteNenntHinweis} />
      )}

      {/* **Ein bekannter an einer neuen Adresse** (NET-02.NAT-02, #1706):
          was die Liste führte, steht neben dem, was gefunden wurde — und
          ein Knopf übernimmt es, statt dass jemand abtippt. **Beide Adressen
          stehen hier offen**, denn an ihnen hängt die Entscheidung. */}
      {nachbar.adresseGeaendert && (
        <div className="flex flex-col gap-1 text-sm">
          <p className="font-medium">{eigene.nachbarnAdresseNeu}</p>
          <dl className="flex flex-col gap-1">
            <dt className="font-medium">{eigene.nachbarnAdresseBisher}</dt>
            <dd className="font-mono text-xs">
              {nachbar.gemerkteAdresse === ''
                ? eigene.nachbarnAdresseKeine
                : nachbar.gemerkteAdresse}
            </dd>
            <dt className="font-medium">{eigene.nachbarnGemerkt}</dt>
            <dd className="font-mono text-xs">{nachbar.merkadresse}</dd>
          </dl>
          <Erklaerungsabsatz
            text={eigene.nachbarnAdresseHinweis}
            klasse="max-w-prose text-leise"
          />
          <div>
            <Knopf gewicht="leise" angehalten={laeuft} aufKlick={uebernehmen}>
              {eigene.nachbarnAdresseUebernehmen}
            </Knopf>
          </div>
        </div>
      )}

      {nachbar.bekannt && !nachbar.adresseGeaendert && (
        <p className="text-sm text-leise">{eigene.nachbarnSchonBekannt}</p>
      )}

      <Aufklappkasten titel={eigene.einladung.technik}>
        <dl className="flex flex-col gap-1 text-sm">
          <dt className="font-medium">{eigene.kontakteAbdruck}</dt>
          <dd className="font-mono text-xs break-all">
            {nachbar.fingerabdruck}
          </dd>
          <dt className="font-medium">{eigene.kontakteAdresse}</dt>
          <dd className="font-mono text-xs">{nachbar.adresse}</dd>

          {/* **Der Name steht dabei, wo es einen gibt** (#1633) — er ist es,
              der beim Aufnehmen in die Kontaktadresse geht. Wo er dieselbe
              Angabe wäre, steht er nicht: eine Zeile, die sich wiederholt,
              erklärt nichts. */}
          {nachbar.merkadresse !== nachbar.adresse && (
            <>
              <dt className="font-medium">{eigene.nachbarnGemerkt}</dt>
              <dd className="font-mono text-xs">{nachbar.merkadresse}</dd>
            </>
          )}
        </dl>
      </Aufklappkasten>
    </li>
  )
}

// standsatz gibt den Satz zu einem der drei Stände.
//
// **Der Rückfall ist der Stand selbst** — dieselbe Überlegung wie bei den
// Kontakten nebenan: ein vierter, den dieses Fenster noch nicht kennt, ist
// immer noch eine Auskunft.
function standsatz(
  eigene: {
    kontakteStandUnbestaetigt: string
    kontakteStandBestaetigt: string
    kontakteStandGesperrt: string
  },
  stand: string,
): string {
  const saetze: Record<string, string> = {
    unbestaetigt: eigene.kontakteStandUnbestaetigt,
    bestaetigt: eigene.kontakteStandBestaetigt,
    gesperrt: eigene.kontakteStandGesperrt,
  }

  return saetze[stand] ?? stand
}
