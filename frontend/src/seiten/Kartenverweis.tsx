// **Der Verweis einer Karte auf einen echten Druck** — Werkstück 2 von
// Phase 3.6
// (#1167).
//
// # Was er löst
//
// Ohne Verweis holt jeder Abgleich **irgendeinen** Druck: `/cards/named`
// wählt selbst aus. Wer in der Trefferliste die Alpha-Fassung angeklickt hat,
// bekam beim nächsten Abgleich den neuesten Druck — und nichts sagte, warum.
//
// # `new`, und wo gesucht wurde
//
// MSE2 kennt Scryfall nicht: gesucht am 2026-08-27 in
// `/Users/david/mse2-quelle` (Fassung f9d9356 vom 2021-07-06) nach `scryfall`,
// `gatherer`, `oracle`, `multiverse` über `src/`. Es gibt dort nichts, was
// eine Karte an einen fremden Datensatz bände.
//
// # Ein Vorgang je Karte, nicht je Tätigkeit
//
// **Ein Schlüssel fürs Setzen und einer fürs Lösen** genügten nur, wenn beide
// Knöpfe nie zugleich sichtbar wären. **Sie sind es**: eine Karte mit Verweis
// und ein angeklickter Suchtreffer zeigen beide. Zwei Schlüssel sperrten
// einander dann nicht, und der zweite Ruf ginge gegen einen Stand, den der
// Bestand noch nicht kennt (§7, Regel 4).
//
// **Der Datensatz ist die Karte**, nicht die Handlung — und Setzen und Lösen
// sind Gegenrichtungen auf demselben.
//
// # Der offene Kontext gehört der Karte
//
// **Der `key` beim Aufrufer räumt ihn** (§7, Regeln 1 und 2 in einer
// Gestalt — `docs/SCHREIBWEGE.md`): wer die Karte wechselt, bekommt ein neues
// Bauteil und nicht ein nachgezogenes altes.
import { Besitzstand } from '../bausteine/Besitzstand'
import { Vorlaeufig } from '../bausteine/Vorlaeufig'
import { api } from '../fassade'
import { Knopf } from '../bausteine/Knopf'
import { useLaden } from '../haken/useLaden'
import type { Vorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useZeitzone, useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { alsZeitpunkt } from '../zeitpunkt'
import type { SuchtrefferDTO } from '../apiscryfall'

interface KartenverweisEigenschaften {
  karteID: string

  // vorschlag ist der zuletzt in der Suche angeklickte Treffer — oder keiner.
  //
  // **Er kommt von oben und wird hier nicht gesucht.** Zwei Suchflächen
  // nebeneinander wären zwei Fragen an dieselbe Gegenstelle; die eine steht
  // schon da, und ihr Treffer trägt die Kennung des Drucks.
  vorschlag: SuchtrefferDTO | undefined

  // vorgang ist der **des Aufrufers**.
  //
  // # Warum nicht ein eigener
  //
  // Dieses Bauteil wird beim Kartenwechsel **abgebaut** (sein `key`); ein
  // Schloss aus einem eigenen `useVorgang()` lebte nur so lange wie die
  // Instanz. Wer von Karte A zu B und zurück zu A geht, bekäme eine **neue,
  // leere** Vorgangsliste — während der erste Ruf auf A noch läuft. §7, Regel
  // 4 meint einen Vorgang je **Datensatz**, nicht je Bauteil.
  //
  // **Der Aufrufer überlebt den Wechsel**, und er hält den Vorgang schon für
  // Kopieren, Ordnen und Löschen.
  vorgang: Vorgang

  // aufFehler meldet nach oben — wie bei den Notizen, und aus demselben
  // Grund: der Kasten wird beim Kartenwechsel abgebaut, und ein Vorgang, der
  // ihn überlebt, meldete an ein Bauteil, das es nicht mehr gibt.
  aufFehler: (meldung: string | undefined) => void
}

export function Kartenverweis({
  karteID,
  vorschlag,
  vorgang,
  aufFehler,
}: KartenverweisEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const eigene = texte.seiten.karten.verweis

  const stand = useLaden(async () => {
    const antwort = await api().KarteVerweis(karteID)

    return auspacken(antwort, () => antwort.verweis)
  }, [karteID])

  const setzen = (printing: string) =>
    vorgang.starte(`verweis-${karteID}`, async () => {
      try {
        const antwort = await api().KarteVerweisSetzen(karteID, printing)

        auspacken(antwort, () => antwort.verweis)
        aufFehler(undefined)

        // **Regel 3**: geschrieben ist geschrieben. Das Nachladen ist
        // Bestenfalls-Arbeit — scheitert es, steht der Verweis trotzdem.
        stand.nachladen()
      } catch (ursache: unknown) {
        aufFehler(meldungAus(ursache, texte))
      }
    })

  const loesen = () =>
    vorgang.starte(`verweis-${karteID}`, async () => {
      try {
        const antwort = await api().KarteVerweisLoesen(karteID)

        auspacken(antwort, () => antwort)
        aufFehler(undefined)
        stand.nachladen()
      } catch (ursache: unknown) {
        aufFehler(meldungAus(ursache, texte))
      }
    })

  const verweis = stand.daten ?? undefined

  return (
    <section aria-label={eigene.aufschrift} className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.aufschrift}</h3>

      {verweis === undefined ? (
        // **Ein leerer Zustand erklärt den nächsten Schritt** (§6) — und
        // nennt die Folge: ohne Verweis nimmt der Abgleich irgendeinen Druck.
        <p className="text-sm text-leise">{eigene.keiner}</p>
      ) : (
        <p className="text-sm">
          <a
            href={verweis.adresse}
            target="_blank"
            rel="noreferrer"
            className="underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {verweis.name}
          </a>{' '}
          <span className="text-leise">
            {einsetzen(eigene.druck, {
              setcode: verweis.setcode.toUpperCase(),
              nummer: verweis.nummer,
            })}
          </span>{' '}
          {/* **Das Alter steht dabei** (§16, Regel 5): Name, Setcode und
              Nummer stammen aus der Antwort von damals. */}
          <span className="text-leise">
            {einsetzen(eigene.gesetzt, {
              wann: alsZeitpunkt(verweis.gesetzt, darstellung, zeitzone),
            })}
          </span>{' '}
          <Vorlaeufig vorlaeufig={verweis.vorlaeufig} />
        </p>
      )}

      {/* **Besitze ich diese Karte?** (#1292, Vermerk 0057). Hier steht die
          Frage, weil hier der Druck steht — und ein Wunsch zählt nicht als
          Besitz. */}
      {verweis !== undefined && <Besitzstand printing={verweis.printing} />}

      <div className="flex flex-wrap items-center gap-2">
        {vorschlag !== undefined && (
          <Knopf
            angehalten={vorgang.laeuft(`verweis-${karteID}`)}
            aufKlick={() => {
              void setzen(vorschlag.kennung)
            }}
          >
            {einsetzen(eigene.setzen, {
              name: vorschlag.name,
              setcode: vorschlag.setcode.toUpperCase(),
              nummer: vorschlag.nummer,
            })}
          </Knopf>
        )}

        {verweis !== undefined && (
          <Knopf
            angehalten={vorgang.laeuft(`verweis-${karteID}`)}
            aufKlick={() => {
              void loesen()
            }}
          >
            {eigene.loesen}
          </Knopf>
        )}
      </div>

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}
    </section>
  )
}
