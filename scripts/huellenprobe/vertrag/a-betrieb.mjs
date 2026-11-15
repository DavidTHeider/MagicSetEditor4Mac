// **Betrieb: Diagnosepaket, Leistungsprofil, Befehlszentrale, Aufträge, Pflege,
// Rückkehren, Weiterarbeiten** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Nichts hier braucht ein Gegenüber, ein Konto oder das Netz. **Zwei Bindungen
// gelingen in einem einzelnen, sequentiellen Lauf nur eingeschränkt** — die
// Gründe stehen an ihnen (`AuftragAbbrechen`, `RueckkehrQuittieren`).
//
// Braucht `k.set` aus der ersten Gruppe (nur für `Anheften`).
import { mkdirSync } from 'node:fs'
import { join } from 'node:path'

const paket = (k) => join(k.ort.ort, 'vertrag', 'diagnose.msediagnose')

export const aBetrieb = {
  name: 'Betrieb und Pflege (A)',
  rufe: [
    // --- Diagnosepaket: die Vorschau, das Paket in eine Datei, die Datei wieder weg.
    // `DiagnosepaketLoeschen` nimmt **nur das zuletzt geschriebene Paket** — es
    // braucht also den Dialogweg davor (`DiagnosepaketSichern`, die Probe
    // beantwortet ihn mit einem Pfad). Die Endung hängt der Weg selbst an.
    { name: 'Diagnosepaket' },
    {
      name: 'DiagnosepaketSichern', argumente: () => [[]],
      vorher: (p, k) => { mkdirSync(join(k.ort.ort, 'vertrag'), { recursive: true }); return p.frage({ was: 'dialogantwort', datei: paket(k) }) },
    },
    { name: 'DiagnosepaketLoeschen' },

    // --- Leistungsprofil: so kurz wie erlaubt (eine Sekunde, `profildauerHoechstens` ist 60).
    { name: 'Leistungsprofil', argumente: () => [1] },

    // --- Die Befehlszentrale ruft ein Werkzeug des MCP-Servers unmittelbar.
    // Ein lesender Befehl ohne Pflichtargumente, `lesen.sets` wenn es ihn gibt:
    // nichts darin schreibt, und `bestaetigt` ist nur für Gefährliches nötig.
    {
      name: 'Befehlskatalog',
      nimm: 'a.befehle.filter((b) => !b.schreibend && !(b.schema?.required?.length)).map((b) => b.name)',
      nach: (w, k) => { k.aBefehl = w.includes('lesen.sets') ? 'lesen.sets' : w[0] },
    },
    { name: 'BefehlAusfuehren', argumente: (k) => [k.aBefehl, '', false] },

    // --- Aufträge: das Buch führt die Datei-Läufe dieses Rundgangs (Einlesen,
    // Verpacken). Anhalten lässt sich nur einer, der **läuft**
    // (`Auftragsbuch.Abbrechen`); im Rundgang ist jeder zu Ende, bevor der
    // nächste Ruf kommt.
    {
      name: 'Auftraege',
      nimm: 'a.auftraege.map((x) => x.id)', nach: (w, k) => { k.aAuftrag = w.at(-1) ?? '0' },
    },
    {
      name: 'AuftragAbbrechen', argumente: (k) => [k.aAuftrag ?? '0'], erwartet: 'fachfehler',
      grund: 'nur ein laufender Auftrag lässt sich anhalten; im Rundgang ist jeder beendet (auftragsbuch.go, Abbrechen)',
    },

    // --- Pflege: **ohne Wahl**. Die Vorschau nennt Maßnahmen, die den Bestand
    // des Rundgangs verändern würden (Aufräumen); gewählt wird keine. Der Ruf
    // prüft Zeitpunkt und Frist und gibt eine leere Bilanz — ohne Sicherung,
    // denn „nichts zu tun heißt: nicht sichern" (`Datenpflege.Pflegen`).
    { name: 'Pflegevorschau' },
    { name: 'Pflegen', argumente: () => [new Date().toISOString(), []] },

    // --- Rückkehren: quittiert wird, was ein Zurückspielen einer Sicherung
    // angelegt hat. **Dieser Lauf spielt nichts zurück** (das tut
    // `SicherungZurueckspielen` und beendet die Sitzung); gibt es keine offene
    // Rückkehr, ist das Quittieren ein Nein. „Erreicht" gilt also für beide.
    {
      name: 'Rueckkehren',
      nimm: "a.rueckkehren.filter((r) => !r.geklaertAm).map((r) => r.id)", nach: (w, k) => { k.aRueckkehr = w[0] ?? 0 },
    },
    { name: 'RueckkehrQuittieren', argumente: (k) => [k.aRueckkehr ?? 0], erwartet: 'erreicht' },

    // --- Weiterarbeiten: ein Objekt oben halten und wieder lösen.
    { name: 'Anheften', argumente: (k) => ['set', k.set.id] },
    { name: 'Weiterarbeiten' },
    { name: 'Loesen', argumente: (k) => ['set', k.set.id] },
  ],
}
