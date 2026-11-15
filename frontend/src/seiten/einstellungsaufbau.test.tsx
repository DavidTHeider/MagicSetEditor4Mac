// **Der Aufbau der Einstellungsseite** — Laschen und Kürzeltafel.
//
// **Abgetrennt am 2026-09-05** (§2), als `einstellungen.test.tsx` mit dem Feld
// `wirksam` aus Werkstück 9 von Phase 4.12
// (#1351)
// über die Zeilengrenze ging. **Der Schnitt folgt der Frage**: nebenan steht,
// was eine **Wahl** tut, hier, wie die Seite **gegliedert** ist.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import {
  zeigeEinstellungen as zeige,
  zumGebiet,
} from '../test/einstellungsseite'

const texte = kataloge.de

// **Die Kürzeltafel** — der dritte Abnahmekasten aus #245: „eine Übersicht der
// Tastaturkürzel ist in der Anwendung zu sehen".
//
// **Seit UXR-19 hinter einem Aufklapper**
// (#1927):
// die Tafel war höher als das Fenster und stand vor jeder anderen Einstellung
// der Gruppe. Zu sehen ist sie weiterhin — einen Klick weiter, und in der
// Hilfe ohne ihn.
async function kuerzelAufklappen() {
  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.seiten.einstellungen.kuerzelUeberschrift,
    }),
  )
}

describe('Die Kürzeltafel', () => {
  it('zeigt die Kürzel, die nicht auf eine Seite führen', async () => {
    zeige()

    await zumGebiet(kataloge.de.seiten.einstellungen.gruppen.allgemein)
    await kuerzelAufklappen()

    for (const taste of ['⌘K', '⌘F', '⌘Z', '⇧⌘Z', '⌘S', '⇧⌘S']) {
      expect(screen.getByText(taste)).toBeInTheDocument()
    }

    expect(
      screen.getByText(texte.seiten.einstellungen.kuerzel.palette),
    ).toBeInTheDocument()
  })

  // **Und die Seitenkürzel**, aus `navigation.ts` und nicht aus einer Kopie:
  // wer eine Seite hinzufügt, bekommt ihre Zeile geschenkt.
  it('zeigt jedes Seitenkürzel mit dem Namen seiner Seite', async () => {
    zeige()

    await zumGebiet(kataloge.de.seiten.einstellungen.gruppen.allgemein)
    await kuerzelAufklappen()

    expect(screen.getByText('⌘1')).toBeInTheDocument()
    expect(screen.getByText('⌘,')).toBeInTheDocument()

    // **⌘6 ist seit UX-02 die letzte Ziffer der ersten Reihe**
    // (#1674).
    //
    // Hier stand ⌘7, davor ⌘0. Die Reihe ist jetzt **kürzer und geordnet**:
    // sechs Ziffern für Start und die fünf Bereiche, und jede führt auf die
    // erste Seite ihres Bereichs. ⌘7 bis ⌘0 sind frei.
    expect(screen.getByText('⌘6')).toBeInTheDocument()

    // **Und die Tafel nennt den Namen der Seite, auf die die Taste führt.**
    // ⌘6 ist *Ressourcen*, und dessen erste Seite sind die Pakete.
    expect(screen.getByText(texte.seiten.pakete.titel)).toBeInTheDocument()
  })
})

// **Jede Sprungmarke findet ihren Abschnitt** — Werkstück C von Phase 4.10
// (#1325).
//
// Die Seite ist 5,1 Bildschirme hoch, und bis hierher gab es keinen anderen
// Weg als das Rollen. **Der Test hält die Zusage, die auseinanderläuft**: ein
// Verweis, der nirgends ankommt, ist schlimmer als keiner.
describe('Die Laschen der Einstellungen', () => {
  // **Hier stand ein Test über Sprungmarken** (Werkstück C von Phase 4.10).
  // Seit Werkstück I gliedern Laschen die Seite: auf einer Seite, auf der man
  // **eine** Einstellung sucht und nicht liest, ist eine Lasche die bessere
  // Antwort als ein Verweis, der zu einem Abschnitt springt, während alles
  // andere stehen bleibt
  // (#1325).
  it('führt seine Gebiete, und immer genau eines ist offen', async () => {
    zeige()

    const leiste = await screen.findByRole('tablist', {
      name: kataloge.de.seiten.hilfe.aufDieserSeite,
    })

    const laschen = within(leiste).getAllByRole('tab')

    // **Sechs ohne Dienste, sieben mit.** Die Zahl steht hier, damit eine
    // leere Leiste nicht grün ist — und die Vorrichtung belegt keine Dienste.
    //
    // **Fünf seit dem 2026-09-06**: *Mein Zugang* ist auf die Nutzerseite
    // gezogen, wo die übrigen Einstellungen eines Menschen stehen. Der Grund
    // steht in `Persoenliches.tsx`; angestrichen hat es der Zweitblick über
    // Phase 4.12 (#1351).
    //
    // **Sechs seit dem 2026-09-08**: der Austausch kam dazu — Einstellungen
    // mitnehmen und einlesen (Werkstück O3 von Phase 5.2,
    // #1427).
    //
    // **Sieben seit dem 2026-09-14**: *Netz und Erreichbarkeit* ist aus
    // *Werkzeuge ohne Fenster* herausgeschnitten worden. Die vier Kästen
    // standen unter einer Aufschrift, die sie nicht nennt, und dort sucht
    // sie niemand — angestrichen bei der Sichtprüfung zu N2 (#1455).
    //
    // **Neun seit dem 2026-09-21**, und diesmal ist die Zahl das Kriterium
    // selbst: UX-04.AK-01
    // (#1676)
    // nennt „Allgemein, Erscheinungsbild, Nutzer/Profil, Fachfunktionen, KI,
    // Verbindungen, Dienste, Daten/Sicherung und Diagnose" — und der Entwurf
    // zeichnet genau diese neun. Sie hängen **nicht** mehr davon ab, ob
    // Dienste belegt sind: eine Gruppe ist ein Ort, kein Inhalt. **Und seit
    // OPS-10 (#1908) zehn** — die Schalter stehen in einer eigenen.
    expect(laschen).toHaveLength(10)

    expect(
      laschen.filter((l) => l.getAttribute('aria-selected') === 'true'),
    ).toHaveLength(1)

    expect(screen.getAllByRole('tabpanel')).toHaveLength(1)
  })

  // **Jede Zeile zeigt etwas.** Ein Gebiet, hinter dem nichts steht, wäre ein
  // Bedienelement, das etwas verspricht, was nicht kommt.
  it('zeigt hinter jeder Zeile einen Inhalt', async () => {
    zeige()

    const leiste = await screen.findByRole('tablist', {
      name: kataloge.de.seiten.hilfe.aufDieserSeite,
    })

    for (const lasche of within(leiste).getAllByRole('tab')) {
      await userEvent.click(lasche)

      const feld = screen.getByRole('tabpanel')

      // **Die Fläche nennt ihre Zeile** — und zwar deren **Namen** und
      // nicht ihren ganzen Text: seit UX-04.AK-01 steht der Untertitel mit
      // in der Zeile, und die Namensrechnung folgt einem `aria-labelledby`
      // nur eine Ebene weit. Ein Verweis auf den Knopf ergäbe „KI Anbieter,
      // Modelle, Kosten" als Namen der Fläche.
      const benennt = document.getElementById(
        feld.getAttribute('aria-labelledby') ?? '',
      )

      expect(lasche.contains(benennt)).toBe(true)
      expect(benennt).toHaveTextContent(lasche.firstChild?.textContent ?? '')
      expect(feld.textContent.trim(), lasche.textContent).not.toBe('')
    }
  })
})

// **Die Sprachmodelle stehen unter ihrer eigenen Aufschrift** — UX-04.AK-04
// (#1676).
//
// **Gefunden in der Sichtprüfung am 2026-09-21**: die Gruppe *KI* benutzt
// denselben `Dienstschalter` wie die Kartendienste, und darüber stand
// deshalb „Fremde Kartendienste". Dass das Bauteil eine eigene Aufschrift
// annimmt, hält `bausteine/dienstschalter.test.tsx`; hier steht, dass die
// Seite sie auch **übergibt**.
describe('die Gruppe KI', () => {
  it('nennt die Sprachmodelle und nicht die Kartendienste', async () => {
    const eigene = kataloge.de.seiten.einstellungen

    zeige({
      dienste: [
        {
          schluessel: 'dienst.anthropic',
          wert: 'aus',
          wirksam: 'aus',
          werte: ['aus', 'an'],
          bereich: 'dienste',
        },
      ],
    })

    await userEvent.click(screen.getByRole('tab', { name: eigene.gruppen.ki }))

    const feld = screen.getByRole('tabpanel')

    expect(
      within(feld).getByRole('region', { name: eigene.kiAnbieter }),
    ).toBeInTheDocument()

    expect(within(feld).queryByText(eigene.dienste)).not.toBeInTheDocument()
  })
})
