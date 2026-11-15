// **Die Zugriffstoken eines technischen Profils** — Werkstück 10 von Phase
// 4.12 (#1351),
// erweitert in Werkstück O2 von Phase 5.2
// (#1426)
// um Principal, Ablaufwahlen, Rotation, Sammelwiderruf und die Einrichtung
// eines MCP-Clients.
//
// # Warum der Klartext oben steht und nicht in der Zeile
//
// **Weil er genau einmal kommt** — beim Anlegen wie bei der Rotation — und
// danach nirgends mehr steht. Ein Feld in der Zeile sähe aus, als ließe es
// sich später nachschlagen; ein Kasten über der Liste sagt: jetzt kopieren.
//
// # Was geteilt ist
//
// Formular (`Tokenformular`), Zeile (`Tokenzeile`) und Anleitung
// (`Tokenonboarding`) sind eigene Bausteine; hier bleibt, was den Bestand
// ruft: laden, anlegen, ändern, rotieren, zurückziehen — je Datensatz **ein**
// Vorgang (§7, Regel 4), und der Sammelwiderruf fragt an Ort und Stelle nach.
import { Satz } from './Satz'
import { useState } from 'react'

import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { Tokenformular } from './Tokenformular'
import { Tokenonboarding } from './Tokenonboarding'
import { Tokenzeile } from './Tokenzeile'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { TokenangabenDTO, TokenDTO } from '../api'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface Eigenschaften {
  profilID: string
}

export function Zugriffstoken({ profilID }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const geladen = useLaden(async () => {
    const antwort = await api().Token(profilID)

    return auspacken(antwort, () => antwort.token)
  }, [profilID])

  // **Die Ablaufwahlen kommen aus Go** — einmal je Aufbau; sie hängen an der
  // Uhr, nicht am Profil.
  const wahlen = useLaden(async () => {
    const antwort = await api().Tokenablaufwahlen()

    return auspacken(antwort, () => antwort.wahlen)
  }, [])

  const vorgang = useVorgang()

  const [runde, setzeRunde] = useState(0)
  const [klartext, setzeKlartext] = useState('')
  const [gezogen, setzeGezogen] = useState<number | undefined>(undefined)
  const [widerrufGefragt, setzeWiderrufGefragt] = useState(false)
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const fuehreAus = (schluessel: string, arbeit: () => Promise<void>) =>
    vorgang.starte(schluessel, async () => {
      setzeFehler(undefined)

      try {
        await arbeit()

        geladen.nachladen()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  const anlegen = (angaben: TokenangabenDTO) =>
    fuehreAus('neu', async () => {
      const antwort = await api().TokenAnlegen(angaben)

      auspacken(antwort, () => antwort.token)
      setzeKlartext(antwort.klartext)
      setzeRunde((bisher) => bisher + 1)
    })

  const zurueckziehen = (eines: TokenDTO) =>
    fuehreAus(eines.id, async () => {
      const antwort = await api().TokenLoeschen(eines.id)

      auspacken(antwort, () => undefined)
    })

  const rotieren = (eines: TokenDTO) =>
    fuehreAus(eines.id, async () => {
      const antwort = await api().TokenRotieren(eines.id)

      auspacken(antwort, () => antwort.token)
      setzeKlartext(antwort.klartext)
    })

  const erlaubnisSetzen = (eines: TokenDTO, scope: string) =>
    fuehreAus(eines.id, async () => {
      const antwort = await api().TokenAendern({
        id: eines.id,
        profilID: eines.profilID,
        bezeichnung: eines.bezeichnung,
        principal: eines.principal,
        scope,
        ablauf: eines.ablauf,
        partie: eines.partie,
        sitz: eines.sitz,
        fassung: eines.fassung,
      })

      auspacken(antwort, () => antwort.token)
    })

  const alleWiderrufen = () =>
    fuehreAus('alle', async () => {
      const antwort = await api().TokenAlleWiderrufen(profilID)

      setzeGezogen(auspacken(antwort, () => antwort.gezogen))
      setzeWiderrufGefragt(false)
    })

  return (
    <section aria-label={eigene.token} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß; zwei
          gleiche Titel untereinander wären einer zu viel. */}
      <Erklaerungsabsatz
        text={eigene.tokenErklaerung}
        klasse="max-w-prose text-leise"
      />
      <p className="max-w-prose text-leise">
        <Satz text={eigene.tokenEinmalig} />
      </p>

      {klartext !== '' && (
        <p role="status" className="flex flex-col gap-1">
          <span>{eigene.tokenKlartext}</span>
          <code className="select-all break-all rounded-sm bg-flaeche px-2 py-1 font-mono">
            {klartext}
          </code>
        </p>
      )}

      {/* **Der `key` räumt das Formular nach einem Anlegen** — der Entwurf
          lebt im Baustein, und ein neuer Baustein beginnt leer. */}
      <Tokenformular
        key={runde}
        profilID={profilID}
        wahlen={wahlen.daten ?? []}
        laeuft={vorgang.laeuft('neu')}
        aufAnlegen={(angaben) => {
          void anlegen(angaben)
        }}
      />

      {geladen.daten?.length === 0 && (
        <p className="text-leise">{eigene.tokenLeer}</p>
      )}

      <ul className="flex flex-col gap-2">
        {(geladen.daten ?? []).map((eines) => (
          <Tokenzeile
            key={eines.id}
            eines={eines}
            laeuft={vorgang.laeuft(eines.id)}
            aufErlaubnis={(scope) => {
              void erlaubnisSetzen(eines, scope)
            }}
            aufRotieren={() => {
              void rotieren(eines)
            }}
            aufZurueckziehen={() => {
              void zurueckziehen(eines)
            }}
          />
        ))}
      </ul>

      {(geladen.daten?.length ?? 0) > 0 && !widerrufGefragt && (
        <div>
          <Knopf
            gewicht="leise"
            angehalten={vorgang.laeuft('alle')}
            aufKlick={() => {
              setzeWiderrufGefragt(true)
            }}
          >
            {eigene.tokenAlleWiderrufen}
          </Knopf>
        </div>
      )}

      {widerrufGefragt && (
        <Entscheidung
          frage={eigene.tokenAlleWiderrufenFrage}
          erklaerung={eigene.tokenAlleWiderrufenErklaerung}
          bestaetigen={eigene.tokenAlleWiderrufen}
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={vorgang.laeuft('alle')}
          aufBestaetigen={() => {
            void alleWiderrufen()
          }}
          aufAbbrechen={() => {
            setzeWiderrufGefragt(false)
          }}
        />
      )}

      {gezogen !== undefined && (
        <p role="status" className="text-leise">
          {einsetzen(eigene.tokenGezogen, { zahl: String(gezogen) })}
        </p>
      )}

      <Tokenonboarding />

      {(fehler ?? geladen.fehler ?? wahlen.fehler) !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler ?? geladen.fehler ?? wahlen.fehler}
        </p>
      )}
    </section>
  )
}
