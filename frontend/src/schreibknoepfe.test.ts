// **Welcher Einstieg für einen Leser gesperrt ist** — G10
// (#2276).
//
// B03 hat jede Seite mit einem Leser geöffnet und auf sechzehn Schreibknöpfe
// bedienbar gefunden (`docs/belege/huellenprobe/zustaende.md`). Gesperrt wird
// am Knopf mit `schreibt`, und **welcher schreibt, entscheidet der Träger der
// Sache** in `docs/AUTORISIERUNG.md`: was einem Profil gehört, verlangt dort
// die Rolle »Mitarbeiter«; was dem Rechner oder dem Menschen gehört, keine.
//
// **Was dieser Wächter hält**: jeder Einstieg aus der Tafel trägt `schreibt`,
// und jeder ausdrücklich offene trägt es nicht. **Was er nicht hält**: ob die
// Tafel vollständig ist — ein neuer Schreibknopf erscheint hier nicht von
// selbst. Das sieht die Zelle „nur lesen" des Drehbuchs `zustaende` am
// nativen Fenster, über alle 25 Seiten.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const quelle = (datei: string) => readFileSync(join(__dirname, datei), 'utf8')

// Ein Einstieg: Datei, der Ausdruck seiner Aufschrift, und warum.
type Eintrag = readonly [datei: string, aufschrift: string, grund: string]

// **Gesperrt**: die Fähigkeit dahinter verlangt am Profil »Mitarbeiter«.
const gesperrt: readonly Eintrag[] = [
  ['seiten/Sets.tsx', 'texte.seiten.sets.neuKopf', 'Ein leeres Set anlegen'],
  [
    'seiten/Seteinfuhr.tsx',
    'texte.seiten.sets.einlesenTitel',
    'Ein Set öffnen — es landet im Bestand des Profils',
  ],
  [
    'seiten/Paketliste.tsx',
    'eigene.anlegen',
    'Ein Paket aus einem Gerüst anlegen',
  ],
  ['seiten/Sammlung.tsx', 'eigene.aufnehmenOeffnen', 'Ein Exemplar aufnehmen'],
  // **Bis AP-6, Teil 2 (#2363) fehlten diese zwei** — ein Leser fand sie
  // bedienbar und scheiterte erst am Bestand (`query_only`).
  [
    'bausteine/Sammlungsinspector.tsx',
    'eigene.bearbeiten',
    'Ein Exemplar ändern',
  ],
  [
    'bausteine/Sammlungsinspector.tsx',
    'eigene.loeschen',
    'Ein Exemplar ändern',
  ],
  ['bausteine/Deckanlage.tsx', 'eigene.anlegen', 'Ein Deck anlegen'],
  [
    'bausteine/Objektmarken.tsx',
    'eigene.setzen',
    'Marken an jedem Fachobjekt führen',
  ],
  [
    'bausteine/Markenzeile.tsx',
    'eigene.setzen',
    'Marken an jedem Fachobjekt führen — der Schalter öffnet nur das Setzen',
  ],
  ['seiten/Turniere.tsx', 'eigene.neuOeffnen', 'Ein Turnier anlegen'],
  [
    'bausteine/Turnierdatei.tsx',
    'eigene.einlesen',
    'Ein Turnier aus einer Datei einlesen',
  ],
  // **Ein Sitz schreibt noch nichts** — aber die Liste dient allein dem
  // Anfang einer Partie, und ein Formular, das nicht abzuschicken ist,
  // füllt niemand gern.
  ['seiten/Sitzliste.tsx', 'eigene.sitzHinzu', 'Eine Partie anfangen'],
  ['seiten/Partieanfang.tsx', 'eigene.anfangen', 'Eine Partie anfangen'],
  [
    'bausteine/Partienliste.tsx',
    'eigene.partienEinlesen',
    'Eine Partie ausführen und einlesen',
  ],
  [
    'bausteine/Identitaetanlegen.tsx',
    'eigene.ohneIdentitaetAnlegen',
    'Die erste Identität anlegen',
  ],
  ['seiten/Kartenwerkzeuge.tsx', 'eigene.anlegen', 'Eine Karte anlegen'],
  // **Befund 4 des Zweitblicks über Phase G**: derselbe Weg wie im Kopf,
  // und eine Freigabe, die der Dienst nur einem Mitarbeiter erteilt.
  [
    'seiten/Sammlungszustaende.tsx',
    'eigene.neuesTitel',
    'Ein Exemplar aufnehmen — derselbe Weg wie im Kopf',
  ],
  [
    'seiten/Ruecksetzung.tsx',
    'worte.freigeben',
    'Den Neuanfang freigeben — verlangt »Mitarbeiter«',
  ],
  [
    'seiten/Kartenleer.tsx',
    'texte.seiten.karten.ablage.einfuegen',
    'Karten in ein leeres Set einfügen',
  ],
]

// **Offen**: die Sache gehört nicht dem Profil, oder der Einstieg führt
// auch zu einer Handlung, die ein Leser darf.
const offen: readonly Eintrag[] = [
  [
    'seiten/Nutzer.tsx',
    'eigene.neuerNutzer',
    'Nutzer und Profile führen — Träger ist der Nutzer, nicht das Profil',
  ],
  [
    'seiten/Nutzerseite.tsx',
    'eigene.profilAnlegenOeffnen',
    'Nutzer und Profile führen — Träger ist der Nutzer, nicht das Profil',
  ],
  [
    'seiten/Sammlung.tsx',
    'eigene.einfuhrOeffnen',
    'Die Sammlung ausgeben darf ein Leser; gesperrt ist darin das Einlesen',
  ],
  [
    'seiten/Sammlungszustaende.tsx',
    'eigene.listeTitel',
    'Dieselbe Schublade wie im Kopf — ausgeben darf ein Leser',
  ],
]

// oeffnenderTag findet den Knopf, dessen Aufschrift der Ausdruck ist, und
// gibt seinen Text vom Öffnen bis zur Aufschrift — und, steht sie in einem
// Attribut, bis zum Ende des Tags.
function oeffnenderTag(text: string, aufschrift: string): string[] {
  const stellen: string[] = []
  // **Die Aufschrift endet den Ausdruck** — `{eigene.setzen}` wie
  // `{offen ? … : eigene.setzen}` und `titel={eigene.anlegen}`.
  const muster = new RegExp(
    `(?<![\\w.])${aufschrift.replaceAll('.', '\\.')}\\}`,
    'g',
  )

  for (const treffer of text.matchAll(muster)) {
    const davor = text.slice(0, treffer.index)
    const beginn = Math.max(
      ...['<Knopf', '<Aktionsknopf', '<Aufklappkasten', '<button'].map((tag) =>
        davor.lastIndexOf(tag),
      ),
    )

    // **Bis zum Ende des öffnenden Tags**, wo die Aufschrift ein Attribut
    // ist — `<Aufklappkasten titel={…} schreibt>`; ein `=>` schließt ihn
    // nicht.
    const rest = text.slice(treffer.index)
    const ende = /(?<!=)>/.exec(rest)?.index ?? 0

    stellen.push(text.slice(beginn, treffer.index + ende + 1))
  }

  return stellen
}

describe('Die Schreibknöpfe für einen Leser', () => {
  it.each(gesperrt)('%s: %s ist gesperrt (%s)', (datei, aufschrift) => {
    const stellen = oeffnenderTag(quelle(datei), aufschrift)

    expect(stellen.length, 'die Aufschrift steht nicht in der Datei').toBe(1)
    expect(stellen[0]).toMatch(/\sschreibt(\s|=|\/?>)|\{\.\.\.schreibt\./)
  })

  it.each(offen)('%s: %s bleibt offen (%s)', (datei, aufschrift) => {
    const stellen = oeffnenderTag(quelle(datei), aufschrift)

    expect(stellen.length, 'die Aufschrift steht nicht in der Datei').toBe(1)
    expect(stellen[0]).not.toMatch(/\sschreibt(\s|=|\/?>)/)
  })
})
