// **Mitspieler an diesen Tisch lassen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum der Satz über dem Knopf steht und nicht daneben
//
// **Weil es zwei Entscheidungen sind und aussieht wie eine.** Wer erreichbar
// ist, hat noch niemandem erlaubt, an seiner Partie zu ziehen; wer hier
// öffnet, tut genau das. §6 verlangt die Erklärung dort, wo geklickt wird —
// und eine Zusage, die man erst hinterher versteht, ist keine.
//
// # Und warum die Sitze hier stehen
//
// **Weil sie der Wirt vergibt und nicht der Gast.** Eine Sitzung belegt einen
// **Rechner**, keinen Sitz; wer beides gleichsetzte, ließe einen Gast im Namen
// eines anderen ziehen. Am Fenster heißt das: die freien Sitze sind die, an
// denen niemand von hier sitzt — sie stehen da, damit man sieht, was man
// weggibt.
import { useEffect, useRef } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { Tischgespraech } from '../bausteine/Tischgespraech'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { usePartiefreigabe } from '../haken/usePartiefreigabe'
import { usePartienebenbei } from '../haken/usePartienebenbei'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { Partie } from '../apipartie'
import { Abschnittstitel } from '../bausteine/Abschnittstitel'

interface Freigabeeigenschaften {
  partie: Partie

  // meldeGastsitze sagt dem Tisch, welche Sitze Gäste spielen — `undefined`,
  // solange der Stand fehlt (P6, #2374: die Sicht folgt dem Beginn nie zu
  // einem Gastsitz).
  meldeGastsitze?:
    ((gastsitze: readonly string[] | undefined) => void) | undefined
}

export function Partiefreigabe({
  partie,
  meldeGastsitze,
}: Freigabeeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = usePartiefreigabe(partie.id, partie.fuer)

  // **Der Haken hängt oben und nicht am Knopf.** Ein Haken darf nicht hinter
  // einer Bedingung stehen — und er soll es hier auch gar nicht: sein Horcher
  // muss hängen, **bevor** der erste Satz kommt. Gezeigt wird die Leiste
  // trotzdem erst, wenn der Tisch offensteht.
  const gespraech = usePartienebenbei(partie.fuer, 'nebenbei-wirt', partie.id)

  // **Frei ist jeder Sitz außer dem eigenen.** Der Wirt gibt seinen nicht
  // weg — er sitzt darauf.
  const gastsitze = partie.sitze
    .map((sitz) => sitz.kennung)
    .filter((kennung) => kennung !== partie.fuer)

  const offen = lauf.stand?.laeuft === true

  const erwartet = partie.sitze
    .filter((sitz) => sitz.wartet)
    .map((sitz) => sitz.name)

  // **Der Stand beim Aufbau, einmal** — bis dahin kam er erst auf
  // Knopfdruck, und wer Gast ist, wusste der Tisch nicht (P6, #2374).
  // **Über einen Verweis**: `ansehen` bildet sich mit jedem Vorgang neu, und
  // als Abhängigkeit fragte der Effekt nach jedem Lesen gleich wieder.
  const ersteFrage = useRef(lauf.ansehen)
  useEffect(() => {
    ersteFrage.current()
  }, [])

  const gaeste =
    lauf.stand === undefined ? undefined : offen ? gastsitze.join(' ') : ''
  useEffect(() => {
    meldeGastsitze?.(
      gaeste === undefined ? undefined : gaeste.split(' ').filter(Boolean),
    )
  }, [gaeste, meldeGastsitze])

  return (
    <section className="flex flex-col gap-2">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.freigabe}
      </Abschnittstitel>

      <Erklaerungsabsatz
        text={eigene.freigabeErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <p className="text-sm">
        {offen ? eigene.freigabeOffen : eigene.freigabeZu}
      </p>

      <p className="text-sm text-leise">
        {eigene.freigabeSitze}: {gastsitze.join(', ')}
      </p>

      {/* **Wer erwartet wird** — P9c (#2377): der Sitz wartet auf das Deck
          seines Gastes. */}
      {erwartet.length > 0 && (
        <p className="text-sm">
          {einsetzen(eigene.freigabeWartet, { namen: erwartet.join(', ') })}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft || gastsitze.length === 0}
          aufKlick={() => {
            if (offen) {
              lauf.beenden()

              return
            }

            lauf.freigeben(gastsitze)
          }}
        >
          {offen ? eigene.freigabeBeenden : eigene.freigabeStarten}
        </Knopf>

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.freigabeNachsehen}
        </Knopf>
      </div>

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      <Besetzung lauf={lauf} />

      {/* **Erst wenn der Tisch offensteht.** Eine Chatleiste an einem Tisch,
          an dem niemand sitzen darf, wäre ein Feld ohne Empfänger — und ein
          Satz darin ginge mit einem Fehler zurück, den niemand verursacht hat
          (§6: leere Zustände erklären den nächsten Schritt). */}
      {offen && <Tischgespraech lauf={gespraech} />}

      <p className="max-w-prose text-xs text-leise">
        {eigene.freigabeOhneHorchposten}
      </p>
    </section>
  )
}

// Besetzung zeigt, wer schon sitzt.
//
// **Mit Fingerabdruck**, denn ein Sitzname sagt nicht, **wer** ihn hat. Der
// Abdruck ist die Gestalt, die zwei Menschen miteinander vergleichen können —
// die Peerkennung selbst ist ein Schlüssel und nichts zum Ablesen.
function Besetzung({ lauf }: { lauf: ReturnType<typeof usePartiefreigabe> }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const vergeben = lauf.stand?.vergeben ?? []

  if (vergeben.length === 0) {
    return <p className="text-sm text-leise">{eigene.freigabeNiemand}</p>
  }

  return (
    <div className="flex flex-col gap-1">
      <p className="text-sm">{eigene.freigabeBesetzt}</p>

      <ul className="flex flex-col gap-1 text-sm">
        {vergeben.map((eine) => (
          <li key={eine.sitz} className="flex flex-col">
            <span>{eine.sitz}</span>
            <span className="font-mono text-xs text-leise">
              {eine.fingerabdruck}
            </span>

            {/* **Und wer das ist** — Werkstück T7 (#1472): wer miteinander
                gespielt hat, kennt sich schon; eine zweite Paarung wäre eine
                Frage, die beantwortet ist. Sehen darf deshalb trotzdem
                niemand etwas — das entscheidet die Freigabe (T3). */}
            <Sitzpartner sitz={eine.sitz} />
          </li>
        ))}
      </ul>
    </div>
  )
}

// Sitzpartner sagt, welcher bekannte Kontakt auf einem Sitz sitzt.
//
// **Es öffnet nichts.** Es beantwortet die Frage, die nach einer Partie offen
// ist: mit wem habe ich gerade gespielt? Der Weg zum Handel steht daneben, im
// Bereich der Gegenüber — und mit `T9` (#1474) an einer Stelle.
function Sitzpartner({ sitz }: { sitz: string }) {
  const eigene = useTexte().seiten.spielen

  const partner = useLaden(async () => {
    const antwort = await api().Handelspartner(sitz)

    return auspacken(antwort, () => antwort)
  }, [sitz])

  // **Ein Fehler beim Nachschlagen nimmt dem Sitz nichts** (§7, Regel 3) —
  // er sagt dasselbe wie ein unbekanntes Gegenüber, denn genau das ist die
  // Auskunft: dieser Sitz führt zu keinem bekannten Kontakt.
  if (partner.daten === undefined && partner.fehler === undefined) {
    return null
  }

  return (
    <span className="text-xs text-leise">
      {partner.daten?.bekannt === true
        ? `${eigene.freigabePartner}: ${partner.daten.kontakt.eigenerName || partner.daten.kontakt.selbstgewaehlterName}`
        : eigene.freigabePartnerFremd}
    </span>
  )
}
