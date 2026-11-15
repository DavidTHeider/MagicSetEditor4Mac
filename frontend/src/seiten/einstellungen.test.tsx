// Die Einstellungsseite —
// #245.
//
// Vorher führte sie die **Sprache** und sonst nichts; darunter stand ein
// Absatz, der erklärte, dass das Erscheinungsbild der Systemeinstellung folgt.
// Das war keine Einstellung, sondern eine Entschuldigung.
//
// **Was hier steht, ist die Wahl selbst** — der **Aufbau** der Seite steht
// nebenan in `einstellungsaufbau.test.tsx` (abgetrennt am 2026-09-05, §2).
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'
import { schalternamen } from '../bausteine/dienstnamen'
import {
  aktualisierung,
  alleDienste,
  eintraege,
  zeigeEinstellungen as zeige,
  zumGebiet,
} from '../test/einstellungsseite'

const texte = kataloge.de

describe('Die Einstellungsseite', () => {
  // **Keine Auswahl ohne sichtbaren Wert** — gefunden bei der Sichtprüfung am
  // laufenden Fenster in Phase 4.8
  // (#1321).
  //
  // Die vier Dienstschalter standen dort als **leere** Auswahllisten: ihr Wert
  // war richtig gesetzt, aber `Wahl` schlägt die Beschriftung im Katalog unter
  // `<schluessel>.<wert>` nach, und `dienst.scryfall.aus` gibt es nicht. React
  // zeichnet dann eine Option ohne Text, ohne zu klagen. Man sah vier Kästen
  // und konnte nicht erkennen, ob ein Dienst an oder aus ist.
  //
  // **Warum kein Test das gesehen hat**: jeder von ihnen sucht das Feld über
  // seine Beschriftung und prüft den **Wert**. Der Text der Optionen kam in
  // keiner Zusicherung vor.
  //
  // **Deshalb prüft dieser Test die Klasse, nicht die vier Fälle.** Er sieht
  // jede Auswahlliste der Seite an — auch die, die es noch nicht gibt.
  it('zeichnet keine Auswahlmöglichkeit ohne Text', () => {
    // **Die Sprache mit allen neun Werten**, wie Go sie schickt: nur sie ist
    // hier eine Auswahlliste (ab fünf Werten), und der Lauf soll alle drei
    // Arten sehen.
    const sprache = eintraege.find(
      (eintrag) => eintrag.schluessel === 'sprache',
    )

    zeige({
      eintraege: [
        ...eintraege.filter((eintrag) => eintrag !== sprache),
        {
          schluessel: 'sprache',
          wert: 'de',
          wirksam: 'de',
          werte: ['erben', 'de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh'],
          bereich: 'anwendung',
        },
        aktualisierung,
        {
          schluessel: 'ersteschritte',
          wert: 'an',
          wirksam: 'an',
          werte: ['an', 'aus'],
          bereich: 'anwendung',
        },
      ],
      dienste: alleDienste,
      fragen: [
        {
          schluessel: 'frage.suche',
          wert: 'scryfall',
          wirksam: 'scryfall',
          werte: ['scryfall', 'mtgapi', 'mtgjson', 'scrydex'],
          bereich: 'fragen',
        },
      ],
    })

    // **Seit UXR-F2 (#2008) sind es drei Arten**: die Auswahlliste mit ihren
    // Möglichkeiten, die Segmente mit ihren Beschriftungen und der Schalter
    // mit seinem Namen. Der Fehler konnte in jeder von ihnen stehen — ein
    // Segment ohne Text ist ein leerer Kasten, ein Schalter ohne Namen ein
    // Kästchen, das niemand benennt.
    const listen = screen.getAllByRole('combobox')
    const reihen = screen.getAllByRole('radiogroup')
    const schalter = screen.getAllByRole('switch')
    expect(listen.length).toBeGreaterThan(0)
    expect(reihen.length).toBeGreaterThan(0)
    expect(schalter.length).toBeGreaterThan(0)

    for (const liste of listen) {
      for (const option of within(liste).getAllByRole('option')) {
        expect(
          option.textContent,
          `die Auswahl „${liste.getAttribute('aria-label') ?? liste.id}" hat ` +
            `eine Möglichkeit ohne Text (Wert „${option.getAttribute('value') ?? ''}")`,
        ).not.toBe('')
      }
    }

    for (const reihe of reihen) {
      for (const feld of within(reihe).getAllByRole('radio')) {
        expect(
          feld.closest('label')?.textContent,
          `die Reihe „${reihe.getAttribute('data-einstellung') ?? ''}" hat ` +
            `ein Segment ohne Text (Wert „${feld.getAttribute('value') ?? ''}")`,
        ).not.toBe('')
      }
    }

    for (const einer of schalter) {
      expect(
        einer.closest('label')?.textContent,
        `der Schalter „${einer.getAttribute('data-einstellung') ?? ''}" ` +
          'hat keinen Namen',
      ).not.toBe('')
    }
  })

  // **Die letzte Reihe** — was geschieht, wenn doch einmal ein Name fehlt.
  //
  // Vorher war die Antwort: **nichts**. `Wahl` griff über `as keyof` auf den
  // Katalog zu, bekam `undefined`, und React zeichnete eine Auswahl ohne Text.
  // Genau so standen die vier Dienstschalter im Fenster.
  //
  // Jetzt steht der rohe Wert da. Er ist nicht schön — aber er ist lesbar,
  // und er sagt dem Lesenden, was gewählt ist. **Dass es nie so weit kommt**,
  // hält `internal/werkzeuge/einstellungstexte_test.go` gegen die Go-Seite;
  // dieser Test hält, was passiert, wenn der Wächter einmal umgangen wird.
  it('zeigt den rohen Wert, wenn kein Name im Katalog steht', () => {
    zeige({
      eintraege: [
        {
          schluessel: 'unbekannt',
          wert: 'zwei',
          wirksam: 'zwei',
          werte: ['eins', 'zwei'],
          bereich: 'anwendung',
        },
      ],
    })

    // **Zwei Werte, die kein Paar bilden, sind zwei Segmente** (UXR-F2,
    // #2008) — auch ihre Namen fallen auf den rohen Wert zurück.
    const reihe = screen.getByRole('radiogroup', { name: 'unbekannt' })

    expect(
      within(reihe)
        .getAllByRole('radio')
        .map((feld) => feld.closest('label')?.textContent),
    ).toEqual(['eins', 'zwei'])
    expect(within(reihe).getByRole('radio', { name: 'zwei' })).toBeChecked()
  })

  // **Über die Lasche** — Werkstück I von Phase 4.10: die Dienste liegen im
  // Gebiet »Fremde Kartendienste«.
  //
  // **Ein Schalter je Dienst** — UXR-F2
  // (#2008):
  // zwei Werte, aus und an, sind ein Schalter, und der Stand steht an ihm
  // statt in zwei Möglichkeiten einer Auswahl.
  it('nennt jeden Dienst bei seinem Eigennamen und zeigt seinen Stand', async () => {
    const setze = zeige({ dienste: alleDienste })

    await zumGebiet(kataloge.de.seiten.einstellungen.gruppen.dienste)

    for (const dienst of alleDienste) {
      // **Und der Eigenname selbst ist Teil der Zusage.** Fehlt er, stünde am
      // Schalter der rohe Schlüssel; das hält
      // `internal/werkzeuge/einstellungstexte_test.go` gegen die Go-Seite.
      const name = schalternamen[dienst.schluessel]
      if (name === undefined) {
        throw new Error(`in dienstnamen.ts fehlt ${dienst.schluessel}`)
      }

      // Alle vier stehen auf „aus".
      const schalter = screen.getByRole('switch', { name })
      expect(schalter).not.toBeChecked()

      await userEvent.click(schalter)
      expect(setze).toHaveBeenLastCalledWith(dienst.schluessel, 'an')
    }
  })

  // **Seit UX-04 stehen sie in der Gruppe *Erscheinungsbild*** — der Entwurf
  // trennt „Sprache, Start, Kürzel" von „Thema, Schrift, Fenster", auch wo
  // beide im Bereich `anwendung` liegen
  // (#1676).
  it('führt jede Einstellung mit ihren Werten', async () => {
    zeige()

    expect(
      screen.getByLabelText(texte.seiten.einstellungen.namen.sprache),
    ).toBeInTheDocument()

    await zumGebiet(texte.seiten.einstellungen.gruppen.erscheinungsbild)

    for (const name of [
      texte.seiten.einstellungen.namen.erscheinungsbild,
      texte.seiten.einstellungen.namen.schriftgroesse,
    ]) {
      expect(screen.getByLabelText(name)).toBeInTheDocument()
    }

    // **Die Werte kommen aus der Antwort, nicht aus einer Liste in der
    // Komponente.** Sonst stünden die erlaubten Werte an zwei Stellen, und die
    // hier wäre die ungeprüfte. Drei Werte sind seit UXR-F2 (#2008) drei
    // Segmente, und das gewählte trägt den Stand.
    const reihe = screen.getByRole('radiogroup', {
      name: texte.seiten.einstellungen.namen.erscheinungsbild,
    })

    expect(
      within(reihe)
        .getAllByRole('radio')
        .map((feld) => feld.closest('label')?.textContent),
    ).toEqual([
      texte.seiten.einstellungen.werte['erscheinungsbild.system'],
      texte.seiten.einstellungen.werte['erscheinungsbild.hell'],
      texte.seiten.einstellungen.werte['erscheinungsbild.dunkel'],
    ])
    expect(
      within(reihe).getByRole('radio', {
        name: texte.seiten.einstellungen.werte['erscheinungsbild.system'],
      }),
    ).toBeChecked()
  })

  it('meldet eine Änderung mit Schlüssel und Wert', async () => {
    const setze = zeige()

    await zumGebiet(texte.seiten.einstellungen.gruppen.erscheinungsbild)

    await userEvent.click(
      screen.getByRole('radio', {
        name: texte.seiten.einstellungen.werte['erscheinungsbild.dunkel'],
      }),
    )

    expect(setze).toHaveBeenCalledWith('erscheinungsbild', 'dunkel')
  })

  // **Ein Vorgang je Einstellung** (§7, Regel 4): während einer läuft, ist
  // sie gesperrt — ein zweiter Klick ginge gegen einen Stand, den das Backend
  // noch nicht kennt. Gleich, ob sie eine Auswahlliste, Segmente oder ein
  // Schalter ist (UXR-F2, #2008).
  it('sperrt eine Einstellung, während ihr Vorgang läuft', async () => {
    zeige({ laeuft: (schluessel) => schluessel === 'schriftgroesse' })

    // **Die Gegenprobe steht in ihrer eigenen Gruppe** — seit UX-04 liegen
    // Sprache und Schriftgröße nicht mehr nebeneinander.
    for (const feld of within(
      screen.getByRole('radiogroup', {
        name: texte.seiten.einstellungen.namen.sprache,
      }),
    ).getAllByRole('radio')) {
      expect(feld).toBeEnabled()
    }

    await zumGebiet(texte.seiten.einstellungen.gruppen.erscheinungsbild)

    for (const feld of within(
      screen.getByRole('radiogroup', {
        name: texte.seiten.einstellungen.namen.schriftgroesse,
      }),
    ).getAllByRole('radio')) {
      // **Gesperrt mit `aria-disabled`**, damit der Fokus bleibt (`uxrf2`).
      expect(feld).toHaveAttribute('aria-disabled', 'true')
    }
  })

  it('meldet einen Ladefehler am Ort', () => {
    zeige({ ursache: new Error('Der Bestand schweigt.') })

    expect(screen.getByRole('alert')).toHaveTextContent('Der Bestand schweigt.')
  })

  it('meldet einen Fehler der Menüleiste am Ort', () => {
    zeige({ meldefehler: 'Die Leiste blieb englisch.' })

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Die Leiste blieb englisch.',
    )
  })
})

// **Ein Hinweis mit Auszeichnung wird ausgezeichnet, nicht abgetippt.**
//
// # Der Befund
//
// Unter „Nach neuen Fassungen sehen" stand wörtlich `**aus**`, `**melden**`
// und `**herunterladen**` — mit Sternchen. Die Kataloge benutzen `**…**` und
// `` `…` `` seit der Hilfeseite, und `bausteine/Satz.tsx` löst beides auf; der
// Hinweis unter einer Einstellung reichte den Text roh durch.
//
// **Gefunden bei der Sichtprüfung am laufenden Fenster** für die Freigabe
// 0.2.0, nicht von einem Test — genau die Sorte Befund, für die es sie gibt
// (§3).
describe('Der Hinweis unter einer Einstellung', () => {
  it('zeichnet aus, statt Sternchen zu zeigen', () => {
    zeige({ eintraege: [...eintraege, aktualisierung] })

    const hinweis = texte.seiten.einstellungen.hinweise.aktualisierung

    // **Kein Sternchen bleibt stehen.**
    expect(screen.queryByText(/\*\*/)).toBeNull()

    // **Und die Hervorhebung ist wirklich eine.** Ohne diese Hälfte wäre der
    // Test auch grün, wenn jemand die Sternchen nur wegließe — und dann stünde
    // der Hinweis ohne die Betonung da, die ihn lesbar macht.
    const betont = screen.getAllByText('aus', { selector: 'strong' })
    expect(betont.length).toBeGreaterThan(0)

    // **Sichtbar steht der erste Satz** — UX-15.AK-02
    // (#1771).
    // Hier stand bis zum 2026-09-22 der **ganze** Hinweis; er war 371
    // Zeichen lang und schob das nächste Feld nach unten.
    expect(
      screen.getByText(
        (_, knoten) =>
          knoten?.tagName === 'SPAN' &&
          knoten.textContent === kurzform(hinweis).replaceAll('**', ''),
      ),
    ).toBeTruthy()

    // **Und der Rest steht hinter dem Fragezeichen** — nichts geht verloren.
    // **Der zu dieser Einstellung**, nicht irgendeiner: die Fläche zeigt
    // mehrere Felder, und jedes bringt sein eigenes Fragezeichen mit.
    expect(
      screen
        .getAllByRole('tooltip', { hidden: true })
        .map((eines) => eines.textContent),
    ).toContain(langform(ohneAuszeichnung(hinweis)))
  })

  // **Die Suche führt ins Gebiet** — Werkstück U11 von Phase 5.1 (#1422): der
  // Knopf am Treffer öffnet die Lasche, in der die Einstellung wohnt.
  // **Die Suche führt in das Gebiet des Treffers** — UX-04.AK-03
  // (#1676).
  //
  // **Über die Gruppenliste und nicht über einen Knopf.** Bis zum
  // 2026-09-21 stand unter jedem Treffer „Im Gebiet „X" zeigen" — in einer
  // Trefferliste, die den Treffer ein zweites Mal zeichnete. Jetzt schaltet
  // die Liste selbst um, sobald die offene Gruppe keinen Treffer hat.
  it('schlägt das Gebiet des Treffers auf', async () => {
    zeige()

    await userEvent.type(
      screen.getByLabelText(texte.seiten.einstellungen.suche),
      'schriftgröße',
    )

    expect(
      await screen.findByRole('tab', {
        name: texte.seiten.einstellungen.gruppen.erscheinungsbild,
        selected: true,
      }),
    ).toBeInTheDocument()
  })
})
