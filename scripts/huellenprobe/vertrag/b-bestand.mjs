// **Der Bestand selbst: verschlüsseln, sichern, zurückspielen, zurücksetzen** —
// eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// **Diese Gruppe gehört ganz ans Ende**, hinter alle anderen, auch hinter
// `b-sitzung.mjs` und `b-dienste.mjs`. Der Grund ist der Zustand, nicht die
// Ordnung der Wirkungen:
//
// - `VerschluesselungUmschalten` und `SicherungZurueckspielen` **tauschen die
//   Datenbank** unter dem laufenden Fenster aus (`App.imTausch`). Beide
//   verlangen, **allein** zu sein: ein laufender `mse-rest` oder `mse-mcp`
//   hält den Bestand mit offen (`internal/service/bestandsrueckspielen.go`).
//   `b-dienste.mjs` beendet beide, und diese Gruppe setzt voraus, dass es vor
//   ihr lief.
// - Das Zurückspielen holt den Stand von **eben** zurück. Alles, was danach
//   in einer anderen Gruppe angelegt würde, ginge verloren; vor ihr ist das
//   Zurückspielen eine Rundreise ohne Folgen.
//
// # Eine Rundreise, die nichts ändert
//
// Verschlüsseln und wieder entschlüsseln gibt denselben Inhalt in derselben
// Gestalt zurück — der Schlüssel liegt zwischendurch im Schlüsselbund **der
// Probe** (`MSE_SCHLUESSELBUND`), nicht im Anmeldebund dieses Rechners.
// Gesichert wird ohne Kennwort und ohne Klartextwunsch (der Bestand ist dann
// unverschlüsselt); zurückgespielt wird dieselbe Datei **sofort**.
//
// **`Zuruecksetzen` mit der Art `nutzereinstellungen`** nimmt die Wahlen des
// arbeitenden Menschen zurück und löscht nichts: kein Name, keine Revision,
// kein Nachweis (`internal/service/ruecksetzungslauf.go`). Die löschenden
// Arten verlangten die Revision einer Vorschau und scheitern an einem
// laufenden Auftrag — ein Zustand, den diese Gruppe nicht kennt.
//
// **Die ganze Anwendung zurückzusetzen** (`anwendung`) verlangt einen
// Nachweis und legt den Bestand beiseite; sie ist hier nicht Teil des Laufs.
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

import { join } from 'node:path'

import { b } from './hilfen-b.mjs'

const merke = (name) => (w, k) => { b(k)[name] = w }

export const bBestand = {
  name: 'Bestand: Verschlüsselung, Sicherung, Rücksetzung',
  rufe: [
    // Der Stand wird gelesen und nicht vorausgesetzt: die Richtung des
    // ersten Umschaltens folgt ihm, die des zweiten stellt ihn wieder her.
    { name: 'Verschluesselungsstand', nimm: 'a.verschluesselt', nach: merke('verschluesselt') },
    { name: 'VerschluesselungUmschalten', argumente: (k) => [!b(k).verschluesselt] },
    { name: 'VerschluesselungUmschalten', argumente: (k) => [b(k).verschluesselt] },

    // Eine Sicherung des ganzen Bestands; der Speichern-Dialog wird mit dem
    // Ort im Zwischenverzeichnis der Probe beantwortet. Die Endung gehört
    // dazu (`internal/sicherung/archiv.go`).
    {
      name: 'BestandSichern',
      vorher: async (p, k) => { await p.frage({ was: 'dialogantwort', datei: join(k.ort.ort, 'vertrag', 'b-bestand.msesicherung') }) },
      argumente: () => ['', false], nimm: 'a.pfad', nach: merke('sicherung'),
    },
    // **Ansehen, bevor sie zurückkommt** (C01): beide lesen die Sicherung,
    // die eben entstand — ohne sie blieben sie in „Übrige" bei einer
    // Datei, die nicht daliegt.
    { name: 'SicherungBeschauen', argumente: (k) => [b(k).sicherung, ''] },
    { name: 'Ruecksprunglage', argumente: (k) => [b(k).sicherung, ''] },
    { name: 'SicherungZurueckspielen', argumente: (k) => [b(k).sicherung, ''] },

    { name: 'Zuruecksetzen', argumente: () => ['nutzereinstellungen', '', '', '', '', ''] },
  ],
}
