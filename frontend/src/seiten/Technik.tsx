// **Der technische Bereich** — Werkstück 5 von Phase 3.3
// (#1164).
//
// # Was hier ausdrücklich nicht geschieht
//
// **Es wird nichts ausgeführt.** Das Issue sagt es: „Keine beliebige Shell.
// Was hier steht, ist eine Vorlage zum Kopieren." Eine Anwendung, die auf
// Knopfdruck Befehle führte, wäre ein Werkzeug für jeden, der ihr etwas
// unterschieben kann — und der Gewinn wäre ein gesparter Wechsel ins Terminal.
//
// Der Satz steht auch **auf der Seite**, nicht nur hier: wer einen Knopf sieht,
// nimmt an, dass er etwas tut.
//
// # Warum die Zeilen aus Go kommen
//
// Weil sie Pfade tragen, die nur das Programm kennt — den Ort seines eigenen
// Bündels. Eine Zeile, die das Fenster zusammensetzte, wäre eine zweite
// Fassung derselben Frage (§10).
import { useState } from 'react'

import { Betriebszustand } from '../bausteine/Betriebszustand'
import { Agentenkonfiguration } from '../bausteine/Agentenkonfiguration'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { inDieZwischenablage } from '../zwischenablage'
import { Diagnose } from './Diagnose'
import { Peeridentitaet } from './Peeridentitaet'
import { Horchposten } from '../bausteine/Horchposten'
import { Nachbarschaft } from '../bausteine/Nachbarschaft'
import { Fernhorchposten } from '../bausteine/Fernhorchposten'
import { Netzweg } from '../bausteine/Netzweg'
import { Dienstposten } from '../bausteine/Dienstposten'
import { Peerkontakte } from './Peerkontakte'
import { Diagnosepaket } from './Diagnosepaket'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Lauschwahl } from '../bausteine/Lauschwahl'
import { Werkzeugeinrichtung } from '../bausteine/Werkzeugeinrichtung'

interface TechnikEigenschaften {
  // setID ist das offene Set — leer, wenn keines offen ist.
  setID: string | undefined
}

// Technik ist, was in der Gruppe *Diagnose* offen steht: der Zustand und der
// Weg zu einem Fehlerbericht. **Was er seit dem Start gewirkt hat**, steht
// seit #2465 in einem Abschnitt darunter.
//
// **Seit UXR-19 nur noch das**
// (#1927).
// Die Gruppe zeichnete 23 Felder und alles untereinander — Protokoll,
// Messung, Zwischenspeicher, Suchindex und die Werkzeuge ohne Fenster. Wer
// hierherkommt, fragt *warum geht gerade etwas nicht?*, und die Antwort
// steht oben; das Übrige steht in Abschnitten darunter
// (`einstellungsabschnitte.ts`).
export function Technik() {
  return (
    <>
      {/* **Der Betriebszustand zuerst** (O9, #1433): er beantwortet dieselbe
          Frage wie die Diagnose darunter — *warum geht gerade etwas
          nicht?* —, steht aber immer da und kostet nichts. */}
      <Betriebszustand />

      <section className="flex flex-col gap-3 border-t border-kante pt-4">
        {/* **Die Diagnose steht hier und nicht auf einer eigenen Seite**
            (#1166): sie gehört zu demselben Bereich — was man braucht, wenn
            etwas klemmt. Eine eigene Seite hieße, sie erst zu suchen. */}
        <Diagnose />

        {/* **Und das Paket darunter** (O10, #1434): dieselbe Auskunft in
            Teilen, mit Manifest und Prüfsummen — für einen Bericht, der mehr
            braucht als einen Text. */}
        <Diagnosepaket />
      </section>
    </>
  )
}

// Werkzeuge sind die Werkzeuge ohne Fenster — Werkstück 5 von Phase 3.3
// (#1164), und seit OPS-03 und OPS-09 ihre Einrichtung per Knopf.
//
// **Seit UXR-19 ein Abschnitt** (#1927): der Knopf, der ihn aufklappt,
// trägt die Überschrift, die hier über allem stand.
export function Werkzeuge({ setID }: TechnikEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik

  const [kopiert, setzeKopiert] = useState('')
  const [kopierfehler, setzeKopierfehler] = useState(false)

  const stand = useLaden(async () => {
    const antwort = await api().Technikangaben(setID ?? '')

    return auspacken(antwort, () => antwort)
  }, [setID])

  const angaben = stand.daten

  return (
    <section className="flex flex-col gap-3">
      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      {/* **Der Satz steht vor den Knöpfen.** Danach hätte ihn niemand
          gelesen, der schon geklickt hat. */}
      <p className="text-sm text-hinweis">{eigene.keineShell}</p>

      {/* **Und das Nein der Zwischenablage** — Befund 5 des Nachlaufs zum
          Zweitblick über Phase 5.1. */}
      {kopierfehler && (
        <p role="alert" className="text-sm text-fehler">
          {eigene.kopierfehler}
        </p>
      )}

      {angaben !== undefined && (
        <>
          <p className="font-mono text-sm break-all text-leise">
            {einsetzen(eigene.ort, { ort: angaben.ort })}
          </p>

          <p className="text-sm text-leise">
            {angaben.vorhanden.length === 0
              ? eigene.keineWerkzeuge
              : einsetzen(eigene.vorhanden, {
                  namen: angaben.vorhanden.join(', '),
                })}
          </p>

          <ul className="flex flex-col gap-3">
            {angaben.zeilen.map((zeile) => (
              <li key={zeile.schluessel} className="flex flex-col gap-1">
                <span className="text-sm font-medium">
                  {satzZu(eigene.befehle, zeile.schluessel)}
                </span>

                <div className="flex flex-wrap items-center gap-2">
                  <code className="overflow-x-auto rounded-sm bg-betont/60 px-2 py-1 font-mono text-sm">
                    {zeile.befehl}
                  </code>

                  <Aktionsknopf
                    aktion="kopieren"
                    bezeichnung={eigene.kopieren}
                    aufKlick={() => {
                      // **Auch das Nein wird gezeigt** — Befund 5 des
                      // Nachlaufs zum Zweitblick über Phase 5.1.
                      void inDieZwischenablage(zeile.befehl).then(
                        (geglueckt) => {
                          setzeKopiert(geglueckt ? zeile.schluessel : '')
                          setzeKopierfehler(!geglueckt)
                        },
                      )
                    }}
                  >
                    {kopiert === zeile.schluessel
                      ? eigene.kopiert
                      : eigene.kopieren}
                  </Aktionsknopf>
                </div>
              </li>
            ))}
          </ul>

          {/* **Fertig für einen Agenten** (A14, #1453) — dieselbe Zusage
              wie oben: kopiert wird, nicht ausgeführt. */}
          <Agentenkonfiguration agenten={angaben.agenten} />
        </>
      )}

      {/* **Und die Kommandozeile einrichten** (OPS-03, #1720): statt einer
          Zeile `ln -s …` zum Abschreiben — in ein Verzeichnis, das dem
          Menschen gehört, und nie über etwas Fremdes. **Und der PATH und der
          Werkzeugserver per Knopf** (OPS-09, #1907): ein markierter Block,
          der byte-gleich wieder geht. Beide sehen einander (OPS03, #2349). */}
      <Werkzeugeinrichtung />

      {/* **Und wo die Dienste lauschen** (G08, #2272): nur dieser Mac,
          ausgewählte Adressen oder alle — die Wahl gilt für den Dienst und
          den Werkzeugserver über HTTP. */}
      <Lauschwahl />
    </section>
  )
}

// Netzbereich ist alles, was dieses Profil erreichbar macht.
//
// **Ein eigenes Gebiet seit dem 2026-09-14**, und zwar auf einen Befund hin:
// die vier Kästen standen unter der Lasche „Werkzeuge ohne Fenster", und
// dort sucht sie niemand. Ein Kopf, der weniger nennt, als sein Rumpf hält,
// ist derselbe Fehler wie einer, der mehr verspricht (§10) — nur sieht man
// ihn später.
export function Netzbereich() {
  // **Eine Aufnahme unter „Wer in der Nähe ist" lädt die Liste darüber
  // nach** (§7, Regel 3) — Befund aus UXR-16
  // (#1924).
  // Liste und Suche sind hier zwei Geschwister; im Kasten „Gegenüber
  // aufnehmen oder bestätigen" sind sie ein Bauteil und brauchen das nicht.
  const [aufgenommen, setzeAufgenommen] = useState(0)

  return (
    <section className="flex flex-col gap-6">
      {/* **Womit sich dieses Profil ausweist** (N1, #1454) — kein
          Geschmack, sondern ein Schlüssel. */}
      <Peeridentitaet />

      <Peerkontakte nachladen={aufgenommen} />
      <Horchposten />
      <Nachbarschaft
        aufGeaendert={() => {
          setzeAufgenommen((zahl) => zahl + 1)
        }}
      />
    </section>
  )
}

// Netzfachbereich ist, was im Netzbereich nur Fachleute brauchen — es steht
// im eingeklappten Kasten „Was Fachleute hier sonst noch einstellen"
// (UX-04.AK-04, #1676, nach dem Artboard *Einstellungen*). **Erreichbar im
// eigenen Netz bleibt draußen**: Identität, Gegenüber, Horchposten und
// Ankündigung sind der Weg zum Handeln, nicht seine Stellschrauben.
export function Netzfachbereich() {
  return (
    <section className="flex flex-col gap-6">
      <Fernhorchposten />

      {/* **Und wie dieses Programm hinausgeht** (N5, #1571) — nach dem
          Weg über fremde Netze, der sagt, wer **hereinkommt**. */}
      <Netzweg />

      {/* **Und die zwei Dienste, die man schalten kann** (N6, #1572) —
          zuletzt, weil sie den Rechner am weitesten öffnen. */}
      <Dienstposten art="rest" />
      <Dienstposten art="mcp" />
    </section>
  )
}

// satzZu sucht den erklärenden Satz zu einem Schlüssel aus Go.
//
// **Über eine Nachschlagetabelle und nicht über einen Typ.** Der Schlüssel
// kommt aus Go und ist zur Bauzeit eine gewöhnliche Zeichenkette; dieselbe
// Überlegung wie bei den Fehlerschlüsseln in `meldung.ts`.
//
// **Der Rückfall ist der Schlüssel selbst.** Ein Befehl ohne Satz ist immer
// noch ein Befehl; ein leerer Name daneben wäre eine Zeile, die niemand
// einordnen kann.
function satzZu(saetze: Record<string, string>, schluessel: string): string {
  return saetze[schluessel] ?? schluessel
}
