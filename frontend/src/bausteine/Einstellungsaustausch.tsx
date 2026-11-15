// **Einstellungen mitnehmen und einlesen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum zwei Schritte und nicht einer
//
// **§16, Regel 6: „Nichts wird still übernommen."** Der erste Knopf liest eine
// Datei und **zeigt**, was sie ändern würde — Zeile für Zeile, mit dem alten
// Wert daneben und der Auskunft, wessen Entscheidung der alte war. Der zweite
// nimmt auf, was entschieden ist.
//
// Dieselbe Gestalt wie bei der Kartenliste in `Listeneinfuhr.tsx`, und aus
// demselben Grund: eine Datei kommt von irgendwoher.
//
// # Warum die Revision durch das Fenster geht
//
// **Weil zwischen Vorschau und Klick jemand geschrieben haben kann** (§7,
// Regel 2) — der MCP-Server schreibt auf dieselbe Datei. Die Marke aus der
// Vorschau geht zurück; stimmt sie nicht mehr, bleibt alles stehen und die
// Meldung sagt es.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { Hinweiskasten } from './Hinweiskasten'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { Satz } from './Satz'
import { Austauschzeilen, zeilenname } from './Austauschzeilen'

import type { Einstellungsbericht, Einstellungssicht } from '../api'

// abschnitteFuer gibt die Abschnitte, die mitreisen sollen.
//
// **Leer heißt »nutzer« und »profil«** — so steht es in `service`, und eine
// zweite Liste hier liefe zurück. Wer den Arbeitsplatz will, nennt alle drei.
function abschnitteFuer(mitArbeitsplatz: boolean): string[] {
  return mitArbeitsplatz ? ['nutzer', 'profil', 'arbeitsplatz'] : []
}

export function Einstellungsaustausch() {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.austausch

  const [mitArbeitsplatz, setzeMitArbeitsplatz] = useState(false)
  const [sicht, setzeSicht] = useState<Einstellungssicht | null>(null)
  const [bericht, setzeBericht] = useState<Einstellungsbericht | null>(null)
  const [abgewaehlt, setzeAbgewaehlt] = useState<Record<string, boolean>>({})
  const [fehler, setzeFehler] = useState<string>()

  // vorgang hält einen Lauf je Handlung an (§7, Regel 4).
  const vorgang = useVorgang()

  // lauf zählt die Wechsel mit (§7, Regel 2): wer eine zweite Datei einliest,
  // während die erste Antwort unterwegs ist, bekäme sonst die alte Vorschau
  // eingetragen.
  const lauf = useLauf()

  // raeumen schließt die offene Vorschau (§7, Regel 1).
  const raeumen = () => {
    setzeSicht(null)
    setzeAbgewaehlt({})
  }

  const sichern = () => {
    void vorgang.starte('sichern', async () => {
      const meiner = lauf.naechster()

      try {
        const antwort = await api().EinstellungenSichern(
          abschnitteFuer(mitArbeitsplatz),
        )

        auspacken(antwort, () => antwort)
      } catch (ursache: unknown) {
        if (lauf.istAktuell(meiner)) {
          setzeFehler(meldungAus(ursache, texte))
        }

        return
      }

      if (lauf.istAktuell(meiner)) {
        setzeFehler(undefined)
      }
    })
  }

  const einlesen = () => {
    void vorgang.starte('einlesen', async () => {
      const meiner = lauf.naechster()

      try {
        const antwort = await api().EinstellungenEinlesen()
        const gelesen = auspacken(antwort, () => antwort.sicht)

        if (!lauf.istAktuell(meiner)) {
          return
        }

        setzeFehler(undefined)
        setzeBericht(null)
        setzeAbgewaehlt({})

        // **Eine Vorschau ohne Zeilen und ohne Befunde ist ein abgebrochener
        // Dialog** — sie steht dann als leere Tabelle da und sagt, dass
        // nichts zu ändern wäre. Das ist die richtige Antwort: wer den Dialog
        // schließt, hat nichts falsch gemacht.
        setzeSicht(gelesen)
      } catch (ursache: unknown) {
        if (lauf.istAktuell(meiner)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const uebernehmen = (offen: Einstellungssicht) => {
    void vorgang.starte('uebernehmen', async () => {
      const meiner = lauf.naechster()

      try {
        const antwort = await api().EinstellungenUebernehmen(
          offen.inhalt,
          offen.revision,
          offen.zeilen
            .filter((zeile) => !abgewaehlt[zeilenname(zeile)])
            .map(zeilenname),
        )
        const ergebnis = auspacken(antwort, () => antwort.bericht)

        if (!lauf.istAktuell(meiner)) {
          return
        }

        setzeFehler(undefined)
        setzeBericht(ergebnis)
        raeumen()
      } catch (ursache: unknown) {
        // **Nach einem gescheiterten Schreiben wird nicht geräumt**
        // (§7, Regel 3): die Vorschau bleibt stehen, damit man es noch
        // einmal versuchen kann, ohne die Datei erneut zu suchen.
        if (lauf.istAktuell(meiner)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  return (
    <div className="flex max-w-3xl flex-col gap-4">
      <p className="max-w-prose text-leise">
        <Satz text={eigene.erklaerung} />
      </p>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={mitArbeitsplatz}
          disabled={vorgang.laeuft('sichern')}
          onChange={(ereignis) => {
            setzeMitArbeitsplatz(ereignis.target.checked)
          }}
          className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />

        {eigene.mitArbeitsplatz}
      </label>

      <div className="flex flex-wrap gap-2">
        <Aktionsknopf
          aktion="sichern"
          gewicht="neben"
          angehalten={vorgang.laeuft('sichern')}
          aufKlick={sichern}
        >
          {eigene.sichern}
        </Aktionsknopf>

        <Knopf
          gewicht="neben"
          angehalten={vorgang.laeuft('einlesen')}
          aufKlick={einlesen}
        >
          {eigene.einlesen}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <Hinweiskasten
          art="fehler"
          meldung={`${texte.bausteine.zustaende.fehler} ${fehler}`}
        />
      )}

      {bericht !== null && (
        <p role="status" className="text-leise">
          {einsetzen(eigene.uebernommen, {
            zahl: String(bericht.uebernommen),
          })}
        </p>
      )}

      {sicht !== null && (
        <Austauschzeilen
          sicht={sicht}
          abgewaehlt={abgewaehlt}
          umschalten={(name) => {
            setzeAbgewaehlt((vorher) => ({
              ...vorher,
              [name]: !vorher[name],
            }))
          }}
          angehalten={vorgang.laeuft('uebernehmen')}
          uebernehmen={() => {
            uebernehmen(sicht)
          }}
          verwerfen={raeumen}
        />
      )}
    </div>
  )
}
