// **Sitzung, Konten, zweite Faktoren, Kennwörter, Wechsel** — eine Gruppe des
// Vertragsrundgangs (B01, #2220).
//
// **Diese Gruppe gehört ans Ende** (hinter `b-peers`, `b-handel`,
// `b-partien`, `b-dienste` und alle, die Daten des arbeitenden Menschen
// brauchen): sie setzt ihm ein Kennwort, sperrt, meldet ab und an, wechselt
// den Menschen und das Profil. Nach ihr steht die Sitzung wieder da, wo sie
// anfing — angemeldet, derselbe Mensch, dasselbe Profil —, **mit einem
// Unterschied**: der arbeitende Mensch hat jetzt ein Kennwort
// (`vertrag-kennwort-2`). Es gibt keinen Weg, eines wieder zu löschen.
//
// # Wie die Kette zusammenhängt
//
// Jeder Übergang der Sitzung hat eine Vorbedingung (`internal/domain/sitzung.go`):
//
// - `Sperren` verlangt einen Weg zurück — ein Kennwort oder Touch ID
//   (`internal/service/sitzung.go`, „Ohne Kennwort und ohne Sensor ließe sich
//   die Sperre nicht lösen"). Also zuerst `KennwortWechseln`.
// - `Entsperren` geht nur aus `gesperrt`; `LokalAnmelden` nur aus
//   `abgemeldet`; `KennwortWiederherstellen` aus `abgemeldet` und `gesperrt`.
//   Also: sperren → entsperren; abmelden → wiederherstellen → anmelden.
// - **Nach `Abmelden` weist das Tor jeden Ruf ab**, bis auf die wenigen, die
//   es offen lässt (`Entsperren`, `LokalAnmelden`, `KennwortWiederherstellen`
//   …). Zwischen `Abmelden` und `LokalAnmelden` steht deshalb nur
//   `KennwortWiederherstellen`.
//
// # Der zweite Faktor, und warum an einem dritten Menschen
//
// Der arbeitende Mensch bekommt keinen Faktor — mit einem würden `Sperren`,
// `Entsperren` und der Wechsel von ihm abhängen. Ein Konto (`Sitzungskonto`)
// bekommt ihn, und das Konto ist das Ziel von `NutzerWechseln` (mit einem
// Wiederherstellungscode als Nachweis), von `KennwortWiederherstellen` (mit
// einem zweiten) und von `FaktorEntfernen`: **wer einen Faktor führt, den
// verwaltet von außen niemand** (`internal/service/ueberfremde.go`), also
// wechselt der arbeitende Mensch zu ihm und entfernt ihn dort — das ist der
// Weg, den die Anwendung selbst vorsieht. Gegen das Kommandozeilenprogramm
// geprüft (2026-10-09): Einrichten, Bestätigen mit einem hier gerechneten
// Code, Wechseln mit Wiederherstellungscode, Entfernen als er selbst.
//
// **Nicht gelingbar**: `BiometrieEinrichten`, `EntsperrenMitSensor` — Touch
// ID ist ein Systemdialog (`scripts/bindungsreichweite.ausnahmen.json`).
//
// # Wie weit das geprüft ist
//
// Die Abfolge ist gegen den App-Bau gefahren worden (`app.NeueBindungen` über
// einem frisch gestarteten `app.NeueApp`, Wegwerfprobe vom 2026-10-09, nicht
// eingecheckt), die Argumentverbünde strikt gegen die Go-Typen und die
// `nimm`-Ausdrücke gegen die serialisierten Antworten. **Nicht geprüft** ist,
// was nur das native Fenster hat: die Wails-Naht selbst, die echten
// Dateidialoge (dort stand ein eingesetzter Dateiwähler) und jeder
// Unterprozess.

import { b, kennwort, totp } from './hilfen-b.mjs'

const konto = (k) => b(k).konto
const faktor = (k) => b(k).faktor
const sp = (k) => b(k).sp

const merke = (name) => (w, k) => { b(k)[name] = w }

export const bSitzung = {
  name: 'Sitzung, Konten, Faktoren, Kennwörter, Wechsel',
  rufe: [
    // Wer arbeitet und in welchem Profil — gelesen, damit der Rückweg stimmt.
    { name: 'WerArbeitet', nimm: 'a.nutzerID', nach: (w, k) => { b(k).ichID = w; k.nutzer = w } },
    { name: 'GezeigtesProfil', nimm: 'a.profil.id', nach: (w, k) => { b(k).ichProfil = w; k.profil = w } },

    // --- ein Konto mit einem zweiten Faktor (alles Testwerte) ---
    {
      name: 'KontoEintragen', argumente: () => ['sitzung@example.test', kennwort.eins, 'Sitzungskonto'],
      nimm: '({ id: a.nutzer.id, name: a.nutzer.name })', nach: merke('konto'),
    },
    // Das Geheimnis und die zehn Codes stehen **einmal** in dieser Antwort.
    { name: 'FaktorEinrichten', argumente: (k) => [konto(k).id], nimm: '({ geheimnis: a.geheimnis, codes: a.codes })', nach: merke('faktor') },
    // **Ein Code aus dem Geheimnis**, zum Zeitpunkt des Rufs gerechnet.
    { name: 'FaktorBestaetigen', argumente: (k) => [konto(k).id, totp(faktor(k).geheimnis)] },

    // --- ein zweites Profil des arbeitenden Menschen, zum Wechseln ---
    {
      name: 'ProfilAnlegen', argumente: (k) => [b(k).ichID, 'Sitzungsprofil'],
      nimm: '({ id: a.profil.id, name: a.profil.name })', nach: merke('sp'),
    },

    // --- sperren und entsperren ---
    // Ohne Kennwort ließe sich die Sperre nicht lösen; der Mensch hat bis
    // hierher keines, `alt` wird dann nicht geprüft.
    { name: 'KennwortWechseln', argumente: () => ['', kennwort.eins] },
    { name: 'Sperren' },
    { name: 'Entsperren', argumente: () => [kennwort.eins] },

    // --- Profil wechseln und zurück ---
    { name: 'ProfilWechseln', argumente: (k) => [sp(k).id, ''] },
    { name: 'ProfilWechseln', argumente: (k) => [b(k).ichProfil, ''] },

    // --- den Menschen wechseln (mit Wiederherstellungscode) und zurück ---
    { name: 'NutzerWechseln', argumente: (k) => [konto(k).id, faktor(k).codes[0]] },
    { name: 'NutzerWechseln', argumente: (k) => [b(k).ichID, ''] },

    // --- abmelden, das Kennwort des Kontos über den Zettel, wieder anmelden ---
    { name: 'Abmelden' },
    { name: 'KennwortWiederherstellen', argumente: (k) => [konto(k).id, faktor(k).codes[1], kennwort.drei] },
    { name: 'LokalAnmelden', argumente: (k) => [b(k).ichID, kennwort.eins, ''] },

    // --- den Faktor des Kontos nimmt nur das Konto selbst heraus ---
    { name: 'NutzerWechseln', argumente: (k) => [konto(k).id, faktor(k).codes[2]] },
    { name: 'FaktorEntfernen', argumente: (k) => [konto(k).id, 'totp'] },
    { name: 'NutzerWechseln', argumente: (k) => [b(k).ichID, ''] },

    // --- weg, was diese Gruppe angelegt hat ---
    { name: 'NutzerLoeschen', argumente: (k) => [konto(k).id, konto(k).name] },
    { name: 'ProfilLoeschen', argumente: (k) => [sp(k).id, sp(k).name] },

    // --- zuletzt: das Kennwort des Menschen wechseln (mit dem alten) ---
    { name: 'KennwortWechseln', argumente: () => [kennwort.eins, kennwort.zwei] },
  ],
}
