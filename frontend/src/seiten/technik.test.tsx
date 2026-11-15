// **Der technische Bereich** — Werkstück 5 von Phase 3.3
// (#1164).
//
// # Die Zusage, um die es geht
//
// „Keine beliebige Shell. Was hier steht, ist eine Vorlage zum Kopieren, kein
// Ausführungsweg." Der letzte Lauf dieser Datei hält sie fest: es gibt keinen
// Weg, über den diese Seite etwas ausführt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Technik, Werkzeuge } from './Technik'
import { Leistungsprofil } from './Leistungsprofil'
import { Suchindex } from './Suchindex'
import { Zwischenspeicher } from './Zwischenspeicher'
import { belegeApi, raeumeApi } from '../test/api'
import { volleMetriken } from '../test/metriken'
import { kataloge } from '../sprache/kataloge'

// Fragezeichen ist, was auf einem Hilfeknopf steht.
const Fragezeichen = '?'

const texte = kataloge.de.seiten.technik

// angaben belegt die Bindung mit einem festen Ort und zwei Zeilen.
function angaben(vorhanden: string[]) {
  belegeApi({
    // **Der Betriebszustand steht seit O9 auf derselben Seite** (#1433) und
    // fragt beim Aufbau; ohne Beleg meldete er einen Fehler, und der stünde
    // als erste `alert`-Rolle im Baum.
    Gesundheit: () =>
      Promise.resolve({
        lebt: true,
        bereit: true,
        grund: '',
        fremde: [],
        fehler: null,
      }),
    // **Und der Wirkungszähler daneben** — aus demselben Grund.
    Metriken: () => Promise.resolve(volleMetriken()),
    // **Und die Zwischenspeicher** (OPS-08, #1725) — aus demselben Grund.
    Zwischenspeicher: () => Promise.resolve({ speicher: [], fehler: null }),
    // **Und die Schnittstellen** (OPS-09, #1907) — aus demselben Grund.
    Schnittstellen: () =>
      Promise.resolve({ schnittstellen: [], nurEintrag: '', fehler: null }),
    // **Und die Lauschwahl** (G08, #2272) — aus demselben Grund.
    Netzschnittstellen: () =>
      Promise.resolve({ schnittstellen: [], fehler: null }),
    Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
    Befehlszeile: () =>
      Promise.resolve({
        verzeichnis: '',
        imPfad: true,
        pfadzeile: '',
        pruefbefehl: 'mse meta',
        eintraege: [],
        fehler: null,
      }),

    // **Und der Horchstand seit N2** (#1455) — auch er fragt beim Aufbau.
    HorchStand: () =>
      Promise.resolve({
        stand: {
          laeuft: false,
          adresse: '',
          sichtbar: false,
          erreichbarNamen: [],
          erreichbarAdressen: [],
        },
        fehler: null,
      }),

    // **Und der Suchindex seit A15** (#1514) — er fragt beim Aufbau, und
    // ohne Beleg stünde seine Fehlermeldung als erste `alert`-Rolle im Baum.
    Indexstand: () =>
      Promise.resolve({
        aktiv: false,
        manifest: {
          modell: '',
          modellrevision: '',
          laufzeit: '',
          abstandsmass: '',
          dimension: 0,
          zerteilfassung: '',
          suchkonfiguration: '',
          korpusstaende: [],
        },
        aufbau: { laeuft: false, generation: 0, geschrieben: 0, fehler: '' },
        suchbar: false,
        grund: '',
        fehler: null,
      }),

    // **Das Paket fragt erst auf Knopfdruck** — der Beleg steht hier
    // trotzdem, damit ein Klick in einem anderen Lauf nicht ins Leere geht.
    Diagnosepaket: () =>
      Promise.resolve({
        teile: [],
        bytes: 0,
        grenze: 8388608,
        von: '',
        bis: '',
        ohne: [],
        fehler: null,
      }),
    Technikangaben: () =>
      Promise.resolve({
        ort: '/Applications/MagicSetEditor4Mac.app/Contents/MacOS',
        vorhanden,
        zeilen: [
          {
            schluessel: 'hilfe',
            befehl:
              '/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse hilfe',
          },
          { schluessel: 'unbekannt', befehl: 'mse etwas-neues' },
        ],
        agenten: [
          {
            agent: 'claude',
            datei: '~/.claude.json',
            form: 'json',
            inhalt: '{ "mcpServers": {} }',
            befehl: 'claude mcp add …',
            nachgelesen: 'code.claude.com, abgerufen am 2026-09-13',
          },
        ],
        fehler: null,
      }),
  })
}

describe('Der technische Bereich', () => {
  afterEach(raeumeApi)

  it('nennt den Ort und die vorhandenen Werkzeuge', async () => {
    angaben(['mse', 'mse-mcp'])

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    // **Der Ort steht als eigener Satz da** — die Befehlszeilen tragen ihn
    // ebenfalls, deshalb wird hier auf den Satz und nicht auf den Pfad
    // geprüft.
    expect(
      await screen.findByText(
        'Sie liegen unter /Applications/MagicSetEditor4Mac.app/Contents/MacOS.',
      ),
    ).toBeInTheDocument()

    expect(screen.getByText(/mse, mse-mcp/)).toBeInTheDocument()
  })

  // **Ohne Werkzeuge steht dort, warum** (§6: leere Zustände erklären den
  // nächsten Schritt) — und nicht „Vorhanden: ".
  it('sagt es, wenn keines der Werkzeuge da ist', async () => {
    angaben([])

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    expect(await screen.findByText(texte.keineWerkzeuge)).toBeInTheDocument()
  })

  it('kopiert eine Zeile in die Zwischenablage', async () => {
    angaben(['mse'])

    const geschrieben = vi.fn(() => Promise.resolve())

    Object.assign(navigator, { clipboard: { writeText: geschrieben } })

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    const knoepfe = await screen.findAllByRole('button', {
      name: texte.kopieren,
    })

    const erster = knoepfe[0]

    if (erster === undefined) {
      throw new Error('kein Kopierknopf')
    }

    await userEvent.click(erster)

    await waitFor(() => {
      expect(geschrieben).toHaveBeenCalledWith(
        '/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse hilfe',
      )
    })

    expect(await screen.findByText(texte.kopiert)).toBeInTheDocument()
  })

  // **Ein Schlüssel ohne Satz bleibt der Schlüssel.**
  //
  // Ein Befehl ohne Erklärung ist immer noch ein Befehl; ein leerer Name
  // daneben wäre eine Zeile, die niemand einordnen kann.
  it('zeigt einen unbekannten Schlüssel statt nichts', async () => {
    angaben(['mse'])

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    expect(await screen.findByText('unbekannt')).toBeInTheDocument()
  })

  // **Die Kennung des offenen Sets geht mit.**
  it('reicht das offene Set weiter', async () => {
    const gefragt = vi.fn(() =>
      Promise.resolve({
        ort: '/x',
        vorhanden: [],
        zeilen: [],
        agenten: [],
        fehler: null,
      }),
    )

    belegeApi({ Technikangaben: gefragt })

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID="7" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalledWith('7')
    })
  })

  // **Und es gibt keinen Weg, etwas auszuführen** — die Zusage aus dem Issue.
  //
  // # Warum das ein Lauf ist und kein Satz im Kommentar
  //
  // Weil ein späterer Handgriff sie sonst still aufhöbe: ein „Ausführen"-Knopf
  // neben der Zeile sähe hilfreich aus, und niemand käme auf die Idee, dass er
  // eine Zusage bricht. Hier ist er ein roter Lauf.
  it('bietet keinen Weg an, etwas auszuführen', async () => {
    angaben(['mse'])

    // **Alles, was bis UXR-19 ein Bauteil war** (#1927): seitdem stehen
    // Zustand und Diagnose offen und der Rest in Abschnitten — die Zusage
    // gilt für jedes Stück, also zeichnet der Lauf jedes.
    render(
      <Sprachraum sprache="de">
        <Technik />
        <Leistungsprofil />
        <Zwischenspeicher />
        <Suchindex />
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    await screen.findAllByRole('button', { name: texte.kopieren })

    // **Jeder Knopf ist einer aus dieser Liste.**
    //
    // Hier stand bis zum 2026-08-27 `toBe(texte.kopieren)` — jeder Knopf
    // müsse »Befehl kopieren« heißen. Das war dieselbe Zusage, solange es nur
    // einen gab, und wurde falsch, als die Diagnose dazukam (#1166): sie führt
    // ebenfalls nichts aus, heißt aber anders.
    //
    // **Eine Aufzählung und kein einzelner Name**: ein **neuer** Knopf fällt
    // weiter auf, und das ist die Zusage — »Keine beliebige Shell« hält nur,
    // solange jeder Knopf benannt ist.
    const erlaubt = [
      texte.kopieren,

      // **Und das Fragezeichen neben einer Erläuterung** (UX-15,
      // #1771):
      // es schlägt einen Text auf und sonst nichts. Sein **Name** ist der
      // Satz, den es fortsetzt; auf dem Knopf steht ein Fragezeichen.
      Fragezeichen,
      texte.diagnose.anzeigen,
      texte.diagnose.kopieren,

      // **Und die Kommandozeile einrichten** (OPS-03, #1720): die zwei Knöpfe
      // legen Verknüpfungen in `~/.local/bin` und nehmen eigene heraus — sie
      // führen keinen Befehl aus.
      texte.befehlszeileEinrichten,
      texte.befehlszeileHerausnehmen,
      texte.diagnose.sichern,

      // **Und die des Diagnosepakets** (O10, #1434): auch sie führen nichts
      // aus — sie stellen zusammen, schreiben eine Datei und nehmen sie
      // wieder weg.
      texte.paket.zusammenstellen,
      texte.paket.sichern,
      texte.paket.loeschen,

      // **Und der des Leistungsprofils** (OPS-05, #1722): er zeichnet ein
      // Profil dieses Prozesses auf — innerhalb der Anwendung, ohne Shell.
      texte.leistungsprofil.aufzeichnen,

      // **Und der der Zwischenspeicher** (OPS-08, #1725): er wirft weg, was
      // sich neu holen lässt — innerhalb der Anwendung.
      texte.zwischenspeicher.leeren,

      // **Und die des Suchindex** (A15, #1514): er sucht lokal und baut
      // einen Index — beides greift diese Anwendung an und nicht die Shell.
      texte.suchindexSuchen,
      texte.suchindexAufbauen,

      // **Und der der Peeridentität** (N1, #1454): er schlägt den Kasten auf
      // und fragt danach die Anwendung — er startet nichts.
      texte.peerZeigen,

      // **Und der der Kontakte** (N1, #1454): dasselbe — er schlägt auf und
      // fragt. Die Knöpfe **in** der Liste stehen erst da, wenn jemand
      // aufgeschlagen hat, und dieser Lauf tut es nicht.
      texte.kontakteZeigen,

      // **Und der des Horchpostens** (N2, #1455): er öffnet einen Port
      // dieser Anwendung — keine Shell, kein fremdes Programm. Dass er von
      // selbst nichts tut, hält `horchposten.test.tsx`.
      texte.horchenStarten,

      // **Und der der Nachbarschaft** (N2, #1455): er ruft einmal ins
      // Netz und wartet auf Antworten — kein Programm startet dabei.
      texte.nachbarnSuchen,

      // **Und der der Lauschwahl** (G08, #2272): er schreibt die Einstellung
      // `netz.lauschen` — kein Programm startet dabei.
      texte.schnittstellen.lauschwahl.speichern,
    ]

    // **Die Aufschrift ohne das stumme Symbol** — das Zeichen des
    // Aktionskatalogs (UX-10.AK-01, #1682) steht davor und ist für
    // Vorleseprogramme verborgen; benannt wird der Knopf durch das Wort.
    const aufschrift = (knopf: HTMLElement) => {
      const kopie = knopf.cloneNode(true) as HTMLElement
      if (kopie.dataset.aktion !== undefined) {
        kopie.querySelector(':scope > [aria-hidden="true"]')?.remove()
      }

      return kopie.textContent
    }

    for (const knopf of screen.getAllByRole('button')) {
      expect(erlaubt).toContain(aufschrift(knopf))
    }

    // **Und der Satz steht auf der Seite**, nicht nur in einem Kommentar:
    // wer einen Knopf sieht, nimmt an, dass er etwas tut.
    expect(screen.getByText(texte.keineShell)).toBeInTheDocument()
  })

  // **Und sagt es, wenn die Zwischenablage nicht darf** — Befund 5 des
  // Nachlaufs zum Zweitblick über Phase 5.1.
  it('meldet, wenn das Kopieren nicht geht', async () => {
    angaben(['mse'])

    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn(() => Promise.reject(new Error('keine Erlaubnis'))),
      },
    })

    render(
      <Sprachraum sprache="de">
        <Werkzeuge setID={undefined} />
      </Sprachraum>,
    )

    const knoepfe = await screen.findAllByRole('button', {
      name: texte.kopieren,
    })

    const erster = knoepfe[0]

    if (erster === undefined) {
      throw new Error('kein Kopierknopf')
    }

    await userEvent.click(erster)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.kopierfehler,
    )
  })
})
