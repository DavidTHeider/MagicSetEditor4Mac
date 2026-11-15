// **Das Gedächtnis des Arbeitsplatzes** — Werkstück U4 von Phase 5.1
// (#1415).
//
// # Wozu
//
// Ob Navigator und Inspector stehen und wie breit sie sind, ist eine Frage an
// das Profil und den Menschen — nicht an einen Baum, der beim nächsten Öffnen
// neu entsteht. Werkstück O3 (#1427)
// hat dafür vier Schlüssel der Ebene `arbeitsplatz` angelegt, aufgelöst wie
// jede Einstellung: Profil → Mensch → Vorgabe. Hier werden sie gelesen und
// geschrieben — über die Einstellungen, die das Fenster ohnehin lädt, und
// über `EinstellungSetzen`, das es ohnehin hat. Eine eigene Bindung gibt es
// nicht, mit Absicht ([0079](../../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md)).
//
// # Warum der Baustein sie nicht selbst liest
//
// `Arbeitsplatz` weiß nicht, was ein Profil ist
// ([0076](../../../docs/entscheidungen/0076-die-ordnung-der-arbeitsplaetze.md)):
// er bekommt das Gedächtnis als Eigenschaft und ruft `merke`, wenn jemand
// klappt. So bleibt er ohne Bestand prüfbar, und eine Seite, die kein
// Gedächtnis will, lässt die Eigenschaft weg.
//
// # Was er nicht tut
//
//   - **Keine Breite von selbst.** Geschrieben wird sie, wenn jemand den
//     Trenner zwischen den Spalten zieht oder mit den Pfeiltasten bewegt
//     (`breiteWuenschen`, AP-12, #2369) — in dieselben Schlüssel, die die
//     Einstellungen, MCP und REST setzen.
//   - **Nicht warten.** Ein Wechsel ist sofort zu sehen; scheitert das
//     Schreiben, kommt nur er zurück, nicht der ganze Stand (§7).
//   - **Nicht schreiben, was es nicht gibt.** Trägt die Antwort die vier
//     Schlüssel nicht — ein älterer Go-Teil —, gelten die Zahlen aus 0076,
//     und ein Wechsel bleibt im Baum.

import { useCallback, useRef, useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useEinstellungen } from './useEinstellungen'
import { useVorgang } from './useVorgang'

export type Arbeitsplatzregion = 'navigator' | 'inspector'

// Die Zahlen aus Vermerk 0076 — dieselben, die `domain.BreiteNavigatorVorgabe`
// und `BreiteInspektorVorgabe` tragen.
export const breiteNavigatorVorgabe = 240
export const breiteInspectorVorgabe = 320

// Die Schlüssel aus `internal/domain/arbeitsplatz.go`.
const sichtschluessel: Record<Arbeitsplatzregion, string> = {
  navigator: 'arbeitsplatz.navigator',
  inspector: 'arbeitsplatz.inspektor',
}

const breitenschluessel: Record<Arbeitsplatzregion, string> = {
  navigator: 'arbeitsplatz.navigatorbreite',
  inspector: 'arbeitsplatz.inspektorbreite',
}

const sichtbar = 'sichtbar'
const verborgen = 'verborgen'

export interface Arbeitsplatzgedaechtnis {
  // geladen sagt, ob die Einstellungen schon einmal da waren.
  geladen: boolean

  // bereit sagt, ob die vier Schlüssel darin standen — erst dann wird
  // geschrieben.
  bereit: boolean

  navigator: boolean
  inspector: boolean
  navigatorbreite: number
  inspectorbreite: number

  // fehler ist der Satz zum letzten gescheiterten Schreiben — `undefined`,
  // solange keines scheiterte.
  fehler: string | undefined

  laeuft: (region: Arbeitsplatzregion) => boolean
  merke: (region: Arbeitsplatzregion, offen: boolean) => void

  // breiteWuenschen setzt die Breite einer Spalte — sofort zu sehen, und
  // geschrieben wird der letzte Wunsch, wenn das vorige Schreiben fertig ist.
  breiteWuenschen: (region: Arbeitsplatzregion, punkte: number) => void
}

// breiteAus liest eine Breite — und nimmt die Vorgabe, wo keine Zahl steht.
function breiteAus(wert: string, vorgabe: number): number {
  const zahl = Number(wert)

  return wert.trim() !== '' && Number.isInteger(zahl) ? zahl : vorgabe
}

export function useArbeitsplatzgedaechtnis(): Arbeitsplatzgedaechtnis {
  const texte = useTexte()
  const stand = useEinstellungen()
  const vorgang = useVorgang()

  // gewuenscht hält, was seit dem Laden geklappt wurde — je Region, und nur
  // dort, wo jemand geklappt hat. Wo nichts steht, gilt der Bestand.
  const [gewuenscht, setzeGewuenscht] = useState<
    Partial<Record<Arbeitsplatzregion, boolean>>
  >({})
  const [fehler, setzeFehler] = useState<string>()

  // **Die Breiten wie die Sichtbarkeit**: was jemand gezogen hat, gilt, bis
  // das Schreiben scheitert. `ausstehend` hält den Wunsch, der noch nicht
  // geschrieben ist — **nur den letzten**: wer drei Schritte geht, während
  // der erste geschrieben wird, will den dritten (§7, Regel 4).
  const [breitenwunsch, setzeBreitenwunsch] = useState<
    Partial<Record<Arbeitsplatzregion, number>>
  >({})
  const ausstehend = useRef<Partial<Record<Arbeitsplatzregion, number>>>({})
  // bestaetigt ist die zuletzt geschriebene Breite je Region — Nachlauf zum
  // Zweitblick über Phase 6 von 0.9.0, zweiter Teil, R4: ohne sie fiel ein
  // gescheiterter Wunsch hinter eine schon geschriebene Breite auf die
  // geladene zurück.
  const bestaetigt = useRef<Partial<Record<Arbeitsplatzregion, number>>>({})

  const bereit =
    stand.geladen &&
    Object.values(sichtschluessel).every((name) =>
      stand.eintraege.some((eintrag) => eintrag.schluessel === name),
    )

  const offen = (region: Arbeitsplatzregion): boolean =>
    gewuenscht[region] ?? stand.wert(sichtschluessel[region]) !== verborgen

  const merke = useCallback(
    (region: Arbeitsplatzregion, soll: boolean) => {
      // **Ein Vorgang je Datensatz** (§7, Regel 4): läuft einer, unterbleibt
      // der zweite — der Knopf ist solange angehalten.
      if (vorgang.laeuft(region)) {
        return
      }

      setzeGewuenscht((war) => ({ ...war, [region]: soll }))

      if (!bereit) {
        return
      }

      void vorgang.starte(region, async () => {
        try {
          const antwort = await api().EinstellungSetzen(
            sichtschluessel[region],
            soll ? sichtbar : verborgen,
          )
          auspacken(antwort, () => undefined)
          setzeFehler(undefined)
        } catch (ursache: unknown) {
          // **Zurück auf den Bestand** — nur diese Region: der Wunsch fällt
          // weg, und darunter steht, was geschrieben ist.
          setzeGewuenscht((war) =>
            Object.fromEntries(
              Object.entries(war).filter(([name]) => name !== region),
            ),
          )
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [bereit, texte, vorgang],
  )

  const breiteSchreibbar = (region: Arbeitsplatzregion) =>
    bereit &&
    stand.eintraege.some(
      (eintrag) => eintrag.schluessel === breitenschluessel[region],
    )

  // schreibeBreite schreibt, was aussteht — und danach, was inzwischen dazu
  // kam. Ein Fehlschlag nimmt **nur diese Breite** zurück (§7).
  const schreibeBreite = (region: Arbeitsplatzregion) => {
    const punkte = ausstehend.current[region]
    if (punkte === undefined || vorgang.laeuft(`breite:${region}`)) {
      return
    }
    ausstehend.current = { ...ausstehend.current, [region]: undefined }

    void vorgang
      .starte(`breite:${region}`, async () => {
        try {
          const antwort = await api().EinstellungSetzen(
            breitenschluessel[region],
            String(punkte),
          )
          auspacken(antwort, () => undefined)
          bestaetigt.current = { ...bestaetigt.current, [region]: punkte }
          setzeFehler(undefined)
        } catch (ursache: unknown) {
          // **Nur zurück, wenn nichts Jüngeres wartet** — Zweitblick über
          // Phase 6 von 0.9.0, zweiter Teil, Befund 4: sonst nähme ein
          // älterer Fehlschlag den neuesten Wunsch mit, und der nächste
          // Durchgang fände nichts mehr zu schreiben. **Und zurück auf die
          // zuletzt geschriebene Breite** (R4), erst ohne sie auf die geladene.
          if (ausstehend.current[region] === undefined) {
            setzeBreitenwunsch((war) => ({
              ...war,
              [region]: bestaetigt.current[region],
            }))
          }
          setzeFehler(meldungAus(ursache, texte))
        }
      })
      .then(() => {
        schreibeBreite(region)
      })
  }

  const breiteWuenschen = (region: Arbeitsplatzregion, punkte: number) => {
    setzeBreitenwunsch((war) => ({ ...war, [region]: punkte }))

    if (!breiteSchreibbar(region)) {
      return
    }

    ausstehend.current = { ...ausstehend.current, [region]: punkte }
    schreibeBreite(region)
  }

  return {
    geladen: stand.geladen,
    bereit,
    navigator: offen('navigator'),
    inspector: offen('inspector'),
    navigatorbreite:
      breitenwunsch.navigator ??
      breiteAus(
        stand.wert(breitenschluessel.navigator),
        breiteNavigatorVorgabe,
      ),
    inspectorbreite:
      breitenwunsch.inspector ??
      breiteAus(
        stand.wert(breitenschluessel.inspector),
        breiteInspectorVorgabe,
      ),
    fehler,
    laeuft: vorgang.laeuft,
    merke,
    breiteWuenschen,
  }
}
