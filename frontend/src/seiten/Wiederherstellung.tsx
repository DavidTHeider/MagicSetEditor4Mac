// **Das Fenster, wenn der Bestand nicht aufging** —
// #962.
//
// # Warum es diese Seite gibt
//
// Die Abnahme von Phase S lautete: „Ein absichtlich abbrechendes Upgrade
// lässt sich **ohne Handarbeit** auf den alten Stand zurückführen." Der Weg
// dorthin war gebaut — und im gescheiterten Fall **nicht erreichbar**:
// `main.lauf` kehrte vor `wails.Run` zurück, und es gab kein Fenster, keine
// Bindung und keinen Knopf.
//
// Eine Anwendung, die sich bei einem Datenbankfehler wortlos beendet, lässt
// genau die Menschen allein, für die es die Sicherung gibt.
//
// # Was sie **nicht** ist
//
// **Keine zweite Fassung der Bestandspflege.** Sie zeigt dieselbe Seite mit
// denselben Bindungen; nur der Sicherungsknopf fehlt, weil es nichts zu
// sichern gibt. Zwei Fassungen derselben Frage liefen auseinander, und die
// eine, die niemand ansieht, wäre die hier.
//
// # Und kein Ersetzen, wo die Datei heil ist
//
// **F05** (#2344):
// belegt, voll, gesperrt, nur lesbar oder ein Fehler des Datenträgers — dort
// liegt es an etwas um die Datei herum, und der Rat ist „erneut versuchen".
// Zurückspielen und der Weg zurück stehen dann **nicht** da: wer sie nähme,
// verlöre, was seit der Sicherung dazukam. Bei einer ungeklärten Lage steht
// beides.
import { useId } from 'react'

import { Bestandspflege } from './Bestandspflege'
import type { Rueckweg } from '../antwortenbestand'
import { useBestandspflege } from '../haken/useBestandspflege'
import { useStartWiederholen } from '../haken/useStartWiederholen'
import { Knopf } from '../bausteine/Knopf'
import { ausKatalog } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// heileLagen sind die Klassen, in denen die Datei vermutlich in Ordnung ist.
const heileLagen = new Set([
  'belegt',
  'voll',
  'gesperrt',
  'bestandNurLesbar',
  'bestandEA',
  'kaumPlatz',
])

export function Wiederherstellung({
  grund,
  schluessel,
  rueckweg,
  bestandErsetzt,
}: {
  // grund ist die Meldung des gescheiterten Starts.
  grund: string

  // schluessel ordnet ihn einer Klasse zu — oder ist leer.
  //
  // **Er trägt den Satz, den jemand liest** (Werkstück 3 von Phase 4.13,
  // #1352). Mit `grund` allein stünde im Ernstfall „journal_mode: sqlite3:
  // file is not a database" im Fenster — ein roher Treibertext, den §8
  // verbietet.
  schluessel: string

  // rueckweg nennt die Sicherung, die diese Fassung öffnen kann — sonst `null`.
  //
  // **Weil der Satz oben sonst ins Leere zeigt** (Werkstück 5 derselben
  // Phase): `bestandNeuereFassung` rät „spiele eine Sicherung von vorher
  // zurück", und neben dem Bestand liegen bis zu dreißig davon. Welche
  // aufgeht, hat Go nachgesehen — hier steht sie.
  rueckweg: Rueckweg | null

  bestandErsetzt: () => void
}) {
  const texte = useTexte()
  const titelID = useId()
  const worte = texte.seiten.bestandspflege

  const pflege = useBestandspflege(bestandErsetzt, bestandErsetzt)
  const versuch = useStartWiederholen(bestandErsetzt)

  const heil = heileLagen.has(schluessel)
  const wiederholbar = heil || schluessel === 'startUngeklaert'

  // **Über `ausKatalog` und nicht über einen Index auf `texte.fehler`**: der
  // Katalog trägt neben Sätzen auch den Verbund der Skriptmeldungen, und ein
  // Index darauf gäbe ein Objekt statt einer Zeichenkette.
  const satz = ausKatalog(texte, schluessel)

  return (
    <main data-wiederherstellung className="mx-auto max-w-3xl px-6 py-10">
      <section aria-labelledby={titelID}>
        <h1 id={titelID} className="text-xl font-semibold tracking-tight">
          {worte.wiederherstellungTitel}
        </h1>

        <Erklaerungsabsatz
          text={
            heil
              ? worte.wiederherstellungHinweisHeil
              : worte.wiederherstellungHinweis
          }
          klasse="mt-2 max-w-prose text-sm text-leise"
        />

        {/* **Der Satz zuerst, die Einzelheit darunter.** Der Schlüssel wählt
            ihn aus dem Katalog; kennt der Katalog ihn nicht, bleibt der Satz
            aus Go — eine unübersetzte Meldung ist besser als eine leere (§9).

            **Und der rohe Grund bleibt stehen.** Wer auf dieser Seite steht,
            sucht einen Fehler, und dann ist er die einzige Spur. */}
        {satz !== undefined && <p className="mt-4 max-w-prose">{satz}</p>}

        {/* **Der benannte Weg zurück steht vor dem rohen Grund.** Wer hier
            steht, will zuerst weiterarbeiten; die Einzelheit ist für den,
            der wissen will, warum. */}
        {wiederholbar && (
          <div className="mt-4 flex flex-col items-start gap-2">
            <Knopf
              gewicht="haupt"
              angehalten={versuch.laeuft}
              aufKlick={versuch.wiederholen}
            >
              {worte.erneutVersuchen}
            </Knopf>
            {versuch.weiterGescheitert && (
              <p role="status" className="text-sm">
                {worte.weiterGescheitert}
              </p>
            )}
            {versuch.fehler !== undefined && (
              <p role="alert" className="text-sm text-fehler">
                {versuch.fehler}
              </p>
            )}
          </div>
        )}

        {!heil && rueckweg !== null && (
          <div className="mt-4 rounded-md border border-kante p-3">
            <h2 className="text-sm font-medium">{worte.wegZurueckTitel}</h2>
            <p className="mt-1 max-w-prose text-sm">
              {einsetzen(worte.wegZurueck, {
                datei: rueckweg.datei,
                stand: rueckweg.schemastand,
                sets: String(rueckweg.sets),
                karten: String(rueckweg.karten),
              })}
            </p>
          </div>
        )}

        <div className="mt-4 rounded-md border border-kante p-3">
          <h2 className="text-sm font-medium">
            {worte.wiederherstellungGrund}
          </h2>
          <p className="mt-1 font-mono text-sm break-words text-leise">
            {grund}
          </p>
        </div>
      </section>

      {!heil && (
        <Bestandspflege
          nurZurueckspielen
          sichern={pflege.sicherung.sichern}
          sichertGerade={pflege.sicherung.laeuft}
          waehlen={pflege.rueckspiel.waehlen}
          gesperrt={pflege.rueckspiel.gesperrt}
          entsperren={pflege.rueckspiel.entsperren}
          zurueckspielen={pflege.rueckspiel.zurueckspielen}
          spieltGerade={pflege.rueckspiel.laeuft}
          auswahl={pflege.rueckspiel.auswahl}
          lage={pflege.rueckspiel.lage}
          beiseite={pflege.rueckspiel.beiseite}
          fehler={pflege.rueckspiel.fehler ?? pflege.sicherung.fehler}
          abbrechen={pflege.rueckspiel.raeumen}
        />
      )}
    </main>
  )
}
