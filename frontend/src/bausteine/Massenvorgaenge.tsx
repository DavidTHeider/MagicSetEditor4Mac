// **Massenvorgänge über eine Auswahl** — Werkstück 6 von Phase 4.5
// (#1292).
//
// # Warum die Auswahl draußen liegt
//
// Weil sie zur **Liste** gehört, nicht zu diesem Kasten: die Ankreuzfelder
// stehen an den Zeilen, und wer eine Zeile löscht, nimmt sie aus der Auswahl.
// Ein Kasten, der die Auswahl hielte, hätte einen zweiten Halter derselben
// Sache (§7, Regel 2).
//
// # Warum jede Zeile ihre Fassung mitbringt
//
// **§7, Regel 2.** Ein Massenvorgang über bloße Kennungen überschriebe jede
// Änderung seit dem Laden der Liste — fünfzigfach auf einmal. Die Fassungen
// stehen in den Zeilen, die der Aufrufer ohnehin zeigt.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { ruecknahmeSchreiben } from './massenruecknahme'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type {
  MassenberichtDTO,
  MassenstandDTO,
  MassenzeileDTO,
} from '../antwortensammlung'

export interface MassenvorgaengeEigenschaften {
  // gewaehlt sind die Zeilen samt ihrer gelesenen Fassung.
  gewaehlt: MassenzeileDTO[]

  // gesamt ist die Zahl der Zeilen, aus denen gewählt wurde.
  gesamt: number

  // aufGeaendert läuft nach jedem geglückten Vorgang.
  aufGeaendert: () => void
}

export function Massenvorgaenge({
  gewaehlt,
  gesamt,
  aufGeaendert,
}: MassenvorgaengeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const lagerortID = useId()

  const [lagerort, setzeLagerort] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [bericht, setzeBericht] = useState<string>()
  const [schub, setzeSchub] = useState<string>()

  // **Die Rückfrage vor dem Massenlöschen** — UX-10.AK-04
  // (#1682).
  //
  // **Weil es hier keine Rücknahme gibt.** Ein Exemplar trägt, was jemand von
  // Hand eingetragen hat — Menge, Zustand, Lagerort, Erwerbspreis —, und der
  // Sammlungsbestand hat keinen Verlauf wie ein Set. Und es trifft **mehrere
  // auf einen Klick**: ein Fehlgriff kostet nicht einen Eintrag, sondern die
  // ganze Auswahl.
  //
  // **Sie hält die Auswahl, für die gefragt wird** (§7, Regel 1): ändert sie
  // sich, gilt die Frage nicht mehr — sonst bestätigte, wer „2 löschen?" las,
  // das Löschen einer anderen Auswahl (Nachlauf über UX-12, Befund 3).
  const [fragtLoeschen, setzeFragtLoeschen] = useState<MassenzeileDTO[]>()
  const rueckfrage =
    fragtLoeschen !== undefined && gleicheAuswahl(fragtLoeschen, gewaehlt)
      ? fragtLoeschen
      : undefined

  // befunde sind die Zeilen, die nicht durchkamen (Befund 4).
  const [befunde, setzeBefunde] = useState<string[]>([])

  // ruecknahme ist der Stand vor der letzten Lagerortänderung — DOM-03.AK-03
  // (#1698). **Die Sammlung hat keinen Verlauf**, und ein Lagerort für viele
  // Zeilen überschreibt viele; die alten Werte samt neuer Fassung nennt nur
  // der Bericht dieser Änderung. Jeder neue Vorgang räumt das Angebot.
  const [ruecknahme, setzeRuecknahme] = useState<MassenstandDTO[]>()

  const vorgang = useVorgang()

  // fuehreAus führt einen Vorgang und meldet sein Ergebnis.
  const fuehreAus = (
    schluessel: string,
    weg: () => Promise<MassenberichtDTO | null>,
  ) => {
    void vorgang.starte(schluessel, async () => {
      setzeRuecknahme(undefined)

      try {
        const ergebnis = await weg()

        setzeFehler(undefined)
        setzeBericht(
          ergebnis === null
            ? undefined
            : einsetzen(eigene.masseBericht, {
                geaendert: String(ergebnis.geaendert),
              }),
        )

        // **Was nicht durchkam, steht da** — Befund 4 des Zweitblicks über
        // Phase 4.5. Bei einem Fassungskonflikt stand nur »N geändert«, und
        // die übergangene Zeile war unsichtbar.
        setzeBefunde(ergebnis?.befunde ?? [])

        // **Erst nach dem geglückten Schreiben** (§7, Regel 3).
        aufGeaendert()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const aendern = () => {
    fuehreAus('masse', async () => {
      const antwort = await api().MasseAendern({
        zeilen: gewaehlt,
        menge: null,
        zustand: null,

        // **Nur der Lagerort, und der auch leer.** Ein leeres Feld heißt hier
        // „nimm die Ortsangabe weg" — das ist eine Angabe, keine Auslassung.
        lagerort,
        wunsch: null,
      })

      const bericht = auspacken(antwort, () => antwort.bericht)

      if (bericht !== null && bericht.vorher.length > 0) {
        setzeRuecknahme(bericht.vorher)
      }

      return bericht
    })
  }

  // zuruecknehmen schreibt jeden alten Lagerort an seine Zeilen zurück
  // (`massenruecknahme.ts`). Was seither jemand anderes schrieb, bleibt und
  // wird zum Befund (§7, Regel 2); **scheitert ein Schreiben, bleibt der Rest
  // rücknehmbar**, und was geschrieben ist, wird nachgeladen (Regel 3).
  const zuruecknehmen = (stand: MassenstandDTO[]) => {
    void vorgang.starte('masse', async () => {
      setzeRuecknahme(undefined)

      const ausgang = await ruecknahmeSchreiben(stand, async (zeilen, ort) => {
        const antwort = await api().MasseAendern({
          zeilen,
          menge: null,
          zustand: null,
          lagerort: ort,
          wunsch: null,
        })

        return auspacken(antwort, () => antwort.bericht)
      })

      setzeBericht(
        ausgang.geschrieben
          ? einsetzen(eigene.masseZurueckgenommen, {
              anzahl: String(ausgang.zurueck),
            })
          : undefined,
      )
      setzeBefunde(ausgang.befunde)
      setzeFehler(
        ausgang.fehler === undefined
          ? undefined
          : meldungAus(ausgang.fehler, texte),
      )

      if (ausgang.offen.length > 0) {
        setzeRuecknahme(ausgang.offen)
      }

      if (ausgang.geschrieben) {
        aufGeaendert()
      }
    })
  }

  const loeschen = (zeilen: MassenzeileDTO[]) => {
    setzeFragtLoeschen(undefined)

    fuehreAus('masse', async () => {
      const antwort = await api().MasseLoeschen(zeilen)

      return auspacken(antwort, () => antwort.bericht)
    })
  }

  // schubtext sagt in einem Satz, was der Schub bewegt hat.
  //
  // **Drei Fälle, nicht zwei**: mit »fertig« und »offen« allein sähe ein
  // Schub, der nur Fehlschläge hatte, aus wie ein Fortschritt (Nachlauf über
  // Phase 4.5, Befund 7).
  const schubtext = (schub: {
    geholt: number
    offen: number
    fehlgeschlagen: number
    fertig: boolean
    steht: boolean
  }) => {
    if (schub.fertig) {
      return eigene.massePreiseFertig
    }

    const teile = [
      einsetzen(eigene.massePreisschub, {
        geholt: String(schub.geholt),
        offen: String(schub.offen),
      }),
    ]

    if (schub.fehlgeschlagen > 0) {
      teile.push(
        einsetzen(eigene.massePreiseFehlgeschlagen, {
          anzahl: String(schub.fehlgeschlagen),
        }),
      )
    }

    if (schub.steht) {
      teile.push(eigene.massePreiseSteht)
    }

    return teile.join(' — ')
  }

  // preise holt einen Schub und sagt, was bleibt.
  //
  // **Ein Schub je Klick** — der Grund ist die Drossel (§16, Regel 4):
  // zwanzig Versuche sind bei 600 ms Abstand rund zwölf Sekunden.
  const preise = () => {
    void vorgang.starte('preise', async () => {
      try {
        const antwort = await api().MassePreise(24)
        const geholt = auspacken(antwort, () => antwort.schub)

        setzeFehler(undefined)
        setzeBefunde(geholt?.befunde ?? [])
        setzeSchub(geholt === null ? undefined : schubtext(geholt))

        aufGeaendert()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <section
      aria-label={eigene.masseTitel}
      className="flex flex-col gap-3 rounded-sm border border-kante p-3"
    >
      <h3 className="text-sm font-medium">{eigene.masseTitel}</h3>

      <p className="text-sm text-leise">
        {einsetzen(eigene.masseGewaehlt, {
          gewaehlt: String(gewaehlt.length),
          gesamt: String(gesamt),
        })}
      </p>

      {/* **Ändern und Löschen erst mit einer Auswahl** — UXR-10 (#1918).
          Ohne eine standen Feld und Knöpfe da und sagten auf Klick nur, dass
          nichts gewählt ist. Bericht und Rücknahme bleiben, wenn die Seite
          die Auswahl nach einem Vorgang räumt. */}
      {gewaehlt.length > 0 && (
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-sm" id={lagerortID}>
              {eigene.masseLagerortSetzen}
            </span>
            <input
              type="text"
              aria-labelledby={lagerortID}
              value={lagerort}
              onChange={(ereignis) => {
                setzeLagerort(ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              size={16}
            />
          </label>

          <Aktionsknopf
            aktion="bearbeiten"
            klein
            angehalten={vorgang.laeuft('masse')}
            aufKlick={aendern}
          >
            {eigene.masseAendern}
          </Aktionsknopf>

          {rueckfrage !== undefined && (
            <Entscheidung
              frage={einsetzen(eigene.masseLoeschenFrage, {
                anzahl: String(rueckfrage.length),
              })}
              erklaerung={eigene.masseLoeschenErklaerung}
              bestaetigen={eigene.masseLoeschenJa}
              aktion="loeschen"
              abbrechen={texte.bausteine.aktionen.abbrechen}
              laeuft={vorgang.laeuft('masse')}
              aufBestaetigen={() => {
                loeschen(rueckfrage)
              }}
              aufAbbrechen={() => {
                setzeFragtLoeschen(undefined)
              }}
            />
          )}

          <Aktionsknopf
            aktion="loeschen"
            klein
            angehalten={vorgang.laeuft('masse')}
            aufKlick={() => {
              setzeFragtLoeschen(gewaehlt)
            }}
          >
            {eigene.masseLoeschen}
          </Aktionsknopf>
        </div>
      )}

      {/* **Die Preise brauchen keine Auswahl** — sie gelten allen. */}
      <div>
        <Knopf klein angehalten={vorgang.laeuft('preise')} aufKlick={preise}>
          {eigene.massePreise}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {bericht !== undefined && <p className="text-sm text-text">{bericht}</p>}

      {ruecknahme !== undefined && (
        <Aktionsknopf
          aktion="zuruecknehmen"
          klein
          bezeichnung={einsetzen(eigene.masseZuruecknehmenVon, {
            anzahl: String(ruecknahme.length),
          })}
          angehalten={vorgang.laeuft('masse')}
          aufKlick={() => {
            zuruecknehmen(ruecknahme)
          }}
        >
          {eigene.masseZuruecknehmen}
        </Aktionsknopf>
      )}

      {schub !== undefined && <p className="text-sm text-text">{schub}</p>}

      {/* **Was nicht durchkam, steht da** (Befund 4). */}
      {befunde.map((befund) => (
        <p key={befund} className="text-sm text-hinweis">
          {befund}
        </p>
      ))}
    </section>
  )
}

// gleicheAuswahl sagt, ob zwei Auswahlen dieselben Zeilen in derselben Fassung
// nennen.
function gleicheAuswahl(a: MassenzeileDTO[], b: MassenzeileDTO[]): boolean {
  return (
    a.length === b.length &&
    a.every((zeile, stelle) => {
      const gegen = b[stelle]

      return zeile.id === gegen?.id && zeile.fassung === gegen.fassung
    })
  )
}
