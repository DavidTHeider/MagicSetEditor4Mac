// Die maschinelle Prüfung auf Barrierefreiheit (AGENTS.md §3) — **jede Seite
// im Ausgangszustand**. Was hinter einer Falte steht, prüft
// `barrierefreiheitoffen.test.tsx` (Z4 des Zweitblicks über Phase 6.13).
//
// **Die Seitenliste kommt aus `navigation.ts`, nicht aus einer Kopie daneben.**
// Eine zu kurze Aufzählung ist grün, und niemand sieht es — deshalb steht hier
// kein Array von Seiten, sondern die Liste selbst.
//
// # Die eine Seite, die nicht darin steht
//
// **Die Wiederherstellung** (`seiten/Wiederherstellung.tsx`). Sie ist keine
// Station der Navigation: man landet dort, wenn der Bestand nicht aufgeht, und
// dann gibt es keine Navigation. Sie war damit bis zum 2026-09-07 **nie**
// maschinell geprüft, obwohl der Kopf hier „jede Seite" sagte.
//
// Geprüft wird sie jetzt in ihrem eigenen Lauf (`wiederherstellung.test.tsx`),
// über dieselbe Abfrage aus `test/axe.ts` — eine zweite Fassung der
// Markenliste liefe auseinander.
//
// # Ohne Matcher-Bibliothek
//
// `axe-core` unmittelbar statt `jest-axe` oder `vitest-axe`: die Zusicherung
// darunter ist fünf Zeilen lang, und eine Abhängigkeit weniger ist eine Zeile
// weniger in `DRITTANBIETER.md`, eine Pinnung weniger und ein
// Aktualisierungspfad weniger.
//
// # Was axe unter jsdom nicht sieht
//
// `color-contrast` gilt dort als **unentschieden**: ohne Layout gibt es keine
// gerechnete Farbe. Die Kontraste werden im Browser nachgemessen, nicht hier —
// §3 führt das als eigene Zusage, und dieser Test löst sie **nicht** ein.
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from './test/bestandspflegeprobe'
import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { verstoesse } from './test/axe'

import { seiten, type Seitenschluessel } from './navigation'
import { belegeProbeset, einstellungsprobe } from './test/seitenvorrichtung'
import { geladenesDokumentDer } from './test/geladeneseite'
import { Rahmen } from './rahmen/Rahmen'
import { Seiteninhalt } from './seiten/Seiteninhalt'
import { sprachen } from './sprache/kataloge'

// einstellungsprobe ist ein Einstellungsstand, der geladen und leer ist —
// genug, damit die Einstellungsseite ihre Kürzeltafel zeichnet.
import { Sprachraum } from './sprache/kontext'
import { stillerSpeicherlauf } from './test/speicherlauf'

// **Ein Merkposten je Seite — als vollständiger Record.**
//
// Der Zweck ist nicht die Prosa, sondern der Typ: wer eine Seite in
// `navigation.ts` einträgt, bekommt hier einen **Typfehler**, bevor irgendein
// Test läuft. Eine Aufzählung, die eine Seite vergisst, ist sonst grün, und
// niemand sieht es.
//
// Was jeweils dasteht, ist die Frage, die diese Seite eigen hat — der Rest gilt
// für alle und steht in der gemeinsamen Prüfung darunter.
const merkposten: Record<Seitenschluessel, string> = {
  start:
    'Jede Kennzahl ist ein Knopf, dessen Name Zahl und Wort trägt; Weiterarbeiten und Aufmerksamkeit sind benannte Bereiche, und jedes Anheften und Lösen nennt sein Objekt.',
  sets: 'Die Tabelle hat Kopfzellen, und jede Zeile führt zu ihren Karten.',
  karten:
    'Jede Karte ist per Tastatur zu wählen und anzuhaken; die Anhakung trägt den Kartennamen im Namen.',
  setangaben:
    'Jede Angabe trägt ihre Aufschrift als Label und ihren Feldnamen als `aria-label`; ein Fehler hängt über `aria-describedby` am verursachenden Feld.',
  schluesselwoerter:
    'Jedes Feld trägt seine Aufschrift als Label; die des Spiels stehen als benannte Liste aus Text, weil niemand sie ändern kann, und ein Schlüsselwort ohne greifendes Muster sagt es im Text.',
  pack: 'Packart und Same sind beschriftete Felder; die gezogenen Karten stehen als geordnete Liste, damit die Reihenfolge auch angesagt wird.',
  konsole:
    'Das Eingabefeld trägt seine Aufschrift als Label; das Ergebnis steht in einem `output`, und ein Fehler erscheint als `role="alert"`.',
  vorlagen:
    'Paketwahl, Dateiwahl und das Textfeld tragen ihre Aufschrift als Label; die Warnung, dass eine Änderung für alle Sets gilt, steht vor dem Feld und nicht dahinter.',
  sammlung:
    'Jedes Feld des Formulars trägt seine Aufschrift als Label, die Zustandswahl ist ein `select` mit Beschriftung, und jeder Knopf einer Zeile nennt im `aria-label` den Druck, den er meint — »Löschen« steht sonst zehnmal gleich da.',
  decks:
    'Jedes Feld der Anlage trägt seine Aufschrift als Label; jeder Knopf einer Deckzeile nennt im `aria-label` das Deck, das er meint, und das Urteil der Prüfung steht als Liste, damit die Reihenfolge angesagt wird.',
  turniere:
    'Jedes Formularfeld trägt seine Aufschrift als Label; die Rangliste ist eine Tabelle mit Kopfzellen, und die Paarungen einer Runde stehen als geordnete Liste — die Reihenfolge der Tische wird damit angesagt.',
  symbol:
    'Jeder Kontrollpunkt ist ein eigenes fokussierbares Element mit `aria-label` — Tabulator wechselt zwischen ihnen, die Pfeiltasten bewegen sie. Ein Zeichenwerkzeug auf einem Canvas hätte nichts, worauf ein Tabulator springen könnte (§6).',
  stil: 'Die Stilwahl ist ein beschriftetes Auswahlfeld; jede Stiloption trägt ihren Namen und ihre Erklärung über `aria-describedby`.',
  suche:
    'Die drei Filter sind beschriftet, und die Kürzung erscheint als `role="status"`.',
  ersetzen:
    'Suchtext und Ersatz tragen ihre Aufschrift als Label, die beiden Schalter ebenso; die betroffenen Stellen stehen als Liste in einem benannten Bereich, und ein Fehler erscheint als `role="alert"`.',
  statistik:
    'Jede Zahl steht als Text; die Balkenfläche ist ausgeblendet, statt zweimal vorgelesen zu werden.',
  bilder:
    'Jede Eingabe trägt ein Label, die Hilfetexte hängen über `aria-describedby` daran, und der Fortschritt steht als Satz in einem `role="status"` — der Balken daneben ist ausgeblendet, statt zweimal vorgelesen zu werden.',
  druckbogen:
    'Jede Einstellung trägt ein Label, die Ankreuzung der Schnittmarken ist per Tastatur erreichbar, und das Ergebnis steht als `role="status"`.',
  handeln:
    'Die Partnerliste ist eine Liste mit einem Knopf je Gegenüber, und der Knopf nennt im `aria-label` das Gegenüber, das er aufschlägt — »Aufschlagen« stünde sonst zehnmal gleich da. Jedes Feld des Handelsraums trägt seine Aufschrift als Label, und was sich bewegt hat, steht als Liste.',
  pakete: 'Ein leerer Zustand, der den nächsten Schritt erklärt.',
  ausgeben:
    'Die Frage „Wohin?“ ist eine Gruppe von Optionsfeldern mit Legende, jedes Ziel trägt seinen Satz über `aria-describedby`, und die Fläche des gewählten Ziels ist ein benannter Bereich mit eigener Überschrift.',
  spielen:
    'Jeder Sitz und jedes Deck trägt seine Aufschrift als Label; jede Karte am Tisch ist ein Knopf mit `aria-pressed`, und das Beenden fragt vor der endgültigen Handlung nach.',
  hilfe:
    'Die Begriffe stehen als Beschreibungsliste: der Name als `dt`, der Satz als `dd` — ein Vorleseprogramm sagt dann, was zu welchem gehört.',
  einstellungen: 'Ein leerer Zustand, der den nächsten Schritt erklärt.',
  nutzer:
    'Vier Laschen neben der Spalte „Wer hier arbeitet“, deren Wechselknöpfe Mensch und Profil im `aria-label` nennen; die eingeklappten Abschnitte sagen ihren Stand über `aria-expanded`, und die Hinweise unter den Feldern hängen über `aria-describedby` daran und stehen nicht im `label` — sonst würden sie Teil des Feldnamens.',
}

describe('Barrierefreiheit nach WCAG 2.2 AA', () => {
  it('führt für jede Seite einen Merkposten', () => {
    // Die Gegenprobe zum Typ: er verlangt Vollständigkeit, aber ein leerer
    // Satz erfüllte ihn auch.
    for (const seite of seiten) {
      expect(merkposten[seite.schluessel].length).toBeGreaterThan(20)
    }
  })

  for (const sprache of sprachen) {
    it.each(seiten)(`hält $schluessel in ${sprache} ein`, async (seite) => {
      belegeProbeset()

      // **Auf die Antworten der Seite gewartet** — sonst prüft axe „Wird
      // geladen …".
      //
      // **Am 2026-09-06 nachgemessen**: die Nutzerseite holt sechs Antworten
      // und zeigte bis dahin drei Wörter; `axe.run` ist zwar asynchron und
      // gab ihr damit einen Tick, aber verlassen konnte man sich darauf
      // nicht. Aufgefallen ist es beim Nachziehen desselben Fehlers im
      // Wächter über die Auszeichnungen. **Und eine Runde reichte auch nicht**
      // (G02): eine Setseite liest erst das Set, dann ihre Daten — gewartet
      // wird, bis die Seite steht (`geladeneseite.ts`).
      const dokument = await geladenesDokumentDer(seite.schluessel, sprache)

      expect(await verstoesse(dokument)).toEqual([])
    })

    // **Genau eine Überschrift erster Ebene** — von Hand gezählt, weil axe es
    // hier nicht kann
    // (#812).
    //
    // # Warum nicht über axe
    //
    // `page-has-heading-one` steht bei axe-core unter `best-practice` und läuft
    // mit der WCAG-Auswahl oben nicht mit. **Eingeschaltet meldet die Regel
    // `inapplicable`** — gemessen am 2026-08-22 über `bericht.inapplicable`:
    // sie prüft `html`, und dieser Lauf geht über einen Ausschnitt. Sie
    // hinzuzunehmen wäre ein Wächter, der nichts tut (§10).
    //
    // # Was gezählt wird
    //
    // **Zwei `<h1>` heißen für ein Vorleseprogramm: hier fangen zwei Dokumente
    // an.** Eine Setlasche ist ein Abschnitt des Arbeitsplatzes und keine
    // zweite Seite — der `Setarbeitsplatz` trägt die erste Ebene mit dem Namen
    // des Sets, die Lasche darunter eine zweite. Fünf taten es bis zum
    // 2026-08-22 anders: `ausgeben`, `bilder`, `druckbogen`, `statistik`,
    // `suche`.
    it.each(seiten)(
      `trägt auf $schluessel in ${sprache} genau eine h1`,
      async (seite) => {
        belegeProbeset()

        const dokument = await geladenesDokumentDer(seite.schluessel, sprache)

        const ueberschriften = [...dokument.querySelectorAll('h1')].map(
          (knoten) => knoten.textContent,
        )

        expect(ueberschriften).toHaveLength(1)

        // **Und keine Ebene wird übersprungen** (#834).
        //
        // „Genau eine h1" allein ließ `<h1>Mein Set</h1><h3>Suche</h3>`
        // durch — angestrichen vom Zweitblick über Phase O. Für ein
        // Vorleseprogramm ist eine übersprungene Ebene ein Abschnitt, der
        // fehlt: es sagt „Ebene drei" an, und der Hörer sucht die zwei.
        //
        // **Verlangt wird keine h2 je Lasche.** Drei tragen bewusst keine —
        // ihr Name steht in der Laschenleiste, und der Bereich verweist über
        // `aria-labelledby` darauf. Was gilt, ist die Folge: wo eine dritte
        // Ebene steht, muss eine zweite darüber sein.
        const ebenen = [
          ...dokument.querySelectorAll('h1, h2, h3, h4, h5, h6'),
        ].map((knoten) => Number(knoten.tagName[1]))

        for (const [stelle, ebene] of ebenen.entries()) {
          if (stelle === 0) {
            continue
          }

          expect(ebene).toBeLessThanOrEqual(
            Math.max(...ebenen.slice(0, stelle)) + 1,
          )
        }
      },
    )
  }

  // **Durchgängige Tastaturbedienung** (§6): jede Aktion, die per Maus
  // erreichbar ist, ist auch per Tastatur erreichbar.
  //
  // Geprüft wird das Erreichbare, nicht der Ring: ob ein Fokusring **sichtbar**
  // ist, weiß jsdom nicht — dort gibt es kein Layout und keine gerechnete
  // Farbe. Die Ringfarben stehen gemessen in `docs/KONTRASTE.md`, und
  // `internal/werkzeuge/kontraste_test.go` hält sie gegen den Quelltext.
  it.each(seiten)(
    'macht auf $schluessel alles per Tastatur erreichbar',
    async (seite) => {
      belegeProbeset()

      const dokument = await geladenesDokumentDer(seite.schluessel, 'de')

      const bedienbar = dokument.querySelectorAll<HTMLElement>(
        'a[href], button, input, select, textarea',
      )

      expect(bedienbar.length).toBeGreaterThan(0)

      // **Genau ein Halt in der Seitenleiste** (#1488) — nicht keiner, nicht
      // elf.
      expect(
        dokument.querySelectorAll('nav[aria-label] button[tabindex="0"]'),
      ).toHaveLength(1)

      // **Jedes Gitter hat genau einen Halt** (#2249) — die Zeile mit der
      // wandernden Tabstelle. Keiner hieße, es ist mit dem Tabulator nicht zu
      // betreten; mehrere, dass jede Zeile einer ist.
      for (const gitter of dokument.querySelectorAll('[role="grid"]')) {
        const halte = [...gitter.querySelectorAll<HTMLElement>('*')].filter(
          (knoten) =>
            knoten.tabIndex >= 0 &&
            !('disabled' in knoten && knoten.disabled === true),
        )

        expect(halte).toHaveLength(1)
      }

      for (const element of bedienbar) {
        // **Eine ungewählte Lasche ist die eine erlaubte Ausnahme.**
        //
        // Das ARIA-Muster für Laschen führt einen **wandernden** Tabindex: nur
        // die gewählte steht in der Reihenfolge, die übrigen erreicht man von
        // dort mit den Pfeiltasten. Ohne diese Ausnahme wäre entweder die
        // Prüfung rot oder die Leiste falsch — und fünf Stationen im Tabweg,
        // die ein Vorleseprogramm ohnehin über die Leiste ansagt
        // (#288).
        //
        // **Die Pfeiltasten sind geprüft** — `setarbeitsplatz.test.tsx`.
        if (element.getAttribute('role') === 'tab') {
          continue
        }

        // **Die Seitenleiste ist die zweite** — #1488: dieselbe Gestalt wie
        // die Laschen, ein wandernder Tabindex über elf Einträge, damit vor
        // jedem Arbeitsplatz nicht elf Halte liegen. Dass genau **ein**
        // Eintrag in der Reihenfolge steht, prüft die Schleife unten; die
        // Pfeiltasten prüft `rahmen/leistentabstelle.test.tsx`.
        if (element.closest('nav[aria-label]') !== null) {
          continue
        }

        // **Ein Gitter ist die dritte** (#2249): die Zeile trägt die Tabstelle,
        // die Knöpfe und Häkchen darin erreicht man mit → und ← (`inDieZeile`,
        // `kartenlistetasten.ts`, geprüft in `schluesselwortliste.test.tsx`
        // und `kartenliste*.test.tsx`). Den einen Halt prüft die Schleife oben.
        if (element.closest('[role="grid"]') !== null) {
          continue
        }

        // **Eine Symbolleiste ist die vierte** (AP-9, #2366): die
        // Kennzahlleiste hat einen wandernden Tabindex, damit drei Zahlen
        // nicht drei Halte vor dem Inhalt sind; die Pfeiltasten und die eine
        // Tabstelle prüft `bausteine/kennzahlleiste.test.tsx`.
        if (element.closest('[role="toolbar"]') !== null) {
          continue
        }

        // **Kein anderes Element wird aus der Reihenfolge genommen.**
        // `tabindex="-1"` macht etwas anklickbar und unerreichbar — der
        // häufigste Weg, eine Tastaturbedienung zu verlieren, ohne dass jemand
        // es merkt.
        expect(element.getAttribute('tabindex')).not.toBe('-1')

        // **Ein angehaltenes Bedienelement nimmt keinen Fokus** — und soll
        // keinen nehmen: `disabled` heißt „geht gerade nicht", und der Fokus
        // gehörte dann auf etwas, das nichts tut. Geprüft wird, dass es
        // **angehalten** ist und nicht bloß unerreichbar.
        if ('disabled' in element && element.disabled === true) {
          continue
        }

        element.focus()
        expect(document.activeElement).toBe(element)
      }
    },
  )

  // **Und jedes trägt einen sichtbaren Fokusring** (§6).
  //
  // # Was hier geprüft wird und was nicht
  //
  // **Dass eine Ringangabe dasteht** — nicht, wie sie aussieht. Unter jsdom
  // gibt es kein Layout und keine gerechnete Farbe; ob ein Ring **sichtbar**
  // ist, entscheidet sein Kontrast, und der steht gemessen in
  // `docs/KONTRASTE.md`, gehalten von `internal/werkzeuge/kontraste_test.go`.
  //
  // Der Test daneben prüft, dass ein Element den Fokus **annimmt**. Beides
  // zusammen ist die Zusage aus §6: erreichbar **und** zu sehen, wo man ist.
  // Erreichbar allein hilft niemandem, der hinsehen muss.
  //
  // # Warum auch ein Vorfahr zählt
  //
  // Manche Bauteile setzen den Ring an der Gruppe — eine Zeile, die als Ganzes
  // hervorgehoben wird, wenn eines ihrer Elemente den Fokus hat. Wer nur das
  // Element selbst prüfte, meldete dort einen Fehler, wo der Ring in
  // Wirklichkeit größer ist als das Element.
  //
  // **`has-focus-visible:` zählt wie `focus-within:`** — UXR-F2
  // (#2008):
  // das Optionsfeld eines Segments sieht man nicht, der Ring liegt um sein
  // Beschriftungsfeld und geht an, wenn der Fokus darin sichtbar ist.
  // `internal/werkzeuge/fokusring_test.go` führt beide Formen als Ring am
  // Rahmen.
  it.each(seiten)('zeigt auf $schluessel jeden Fokus', async (seite) => {
    belegeProbeset()

    const dokument = await geladenesDokumentDer(seite.schluessel, 'de')

    const ohneRing: string[] = []

    for (const element of dokument.querySelectorAll<HTMLElement>(
      'a[href], button, input, select, textarea',
    )) {
      if (/focus-visible:|focus:/.test(element.className)) {
        continue
      }

      if (
        element.closest(
          '[class*="focus-within:"], [class*="has-focus-visible:"]',
        )
      ) {
        continue
      }

      ohneRing.push(
        `${element.tagName.toLowerCase()} „${element.textContent.trim().slice(0, 30)}"`,
      )
    }

    expect(ohneRing).toEqual([])
  })

  // Und die Palette, die kein Seitenaufruf ist.
  it('hält die offene Befehlspalette ein', async () => {
    const { container } = render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()}>
          <Seiteninhalt
            speichern={stillerSpeicherlauf()}
            wechseln={vi.fn()}
            bestandGeaendert={vi.fn()}
            seite="start"
            fassung="0.0.0-test"
            zuUeber={false}
            einfuehrungZeigen={vi.fn()}
            ueberAngekommen={vi.fn()}
            setID={undefined}
            karteID={undefined}
            einstellungen={einstellungsprobe}
            bestandspflege={stillePflege()}
            ruecksprung={stillerRuecksprung()}
            pflege={stilleDatenpflege()}
            meldefehler={undefined}
            oeffnungsanstoss={false}
            oeffnungAngestossen={vi.fn()}
            neuanstoss={false}
            neuAngestossen={vi.fn()}
            bestandsstand={0}
            verlaufsstand={0}
            verlaufsfehler={undefined}
            waehleSet={vi.fn()}
            setEntfernt={vi.fn()}
            oeffneKarte={vi.fn()}
            oeffneSymbol={vi.fn()}
          />
        </Rahmen>
      </Sprachraum>,
    )

    container
      .querySelector<HTMLButtonElement>('nav button:last-of-type')
      ?.click()

    expect(await verstoesse(container)).toEqual([])
  })
})
