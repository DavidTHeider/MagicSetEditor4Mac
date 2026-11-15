// **Einstellungen, Austausch, Fenster- und Tischbild, Anbietergeheimnis** —
// eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// **Alles hier ändert Zustand des einen Nutzers und seines Profils, und alles
// wird am Ende zurückgestellt.** Die Probe läuft in einem eigenen Bestand mit
// eigenem Schlüsselbund (`Probe.ort`, `MSE_SCHLUESSELBUND`), also bleibt der
// Anmeldebund des Rechners unberührt (§3, „der Anmeldebund dieses Rechners ist
// für Testläufe gesperrt").
//
// Es gibt **einen Nutzer** mit der Kennung `"1"` (Auftrag des Rundgangs).
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { png } from '../../pngbau.mjs'

const nutzer = '1'
const austausch = (k) => join(k.ort.ort, 'vertrag', 'einstellungen.json')
const bild = (k, wofuer) => join(k.ort.ort, 'vertrag', `${wofuer}.png`)

// bildvorher legt ein kleines Bild an und beantwortet den **nächsten**
// Dateidialog damit — `ruf` hat ihn davor leer beantwortet.
const bildvorher = (wofuer) => async (p, k) => {
  mkdirSync(join(k.ort.ort, 'vertrag'), { recursive: true })
  writeFileSync(bild(k, wofuer), png({ breite: 16, hoehe: 16 }))
  await p.frage({ was: 'dialogantwort', datei: bild(k, wofuer) })
}

export const aEinstellungen = {
  name: 'Einstellungen, Bilder, Geheimnis (A)',
  rufe: [
    // --- Einstellungen. `feldtiefe` ist eine Einstellung des Profils (sie reist
    // im Austausch mit) und ändert nur, welche Felder die Oberfläche zeigt.
    { name: 'EinstellungSetzen', argumente: () => ['feldtiefe', 'erweitert'] },
    {
      // **Der Austausch beginnt mit einer Ausfuhr**: nur sie liefert den Text,
      // den `EinstellungenUebernehmen` danach verlangt (`inhalt`). Ein Dialog,
      // den die Probe mit einem Pfad beantwortet. Leere Abschnitte heißen:
      // das Bündel aus Nutzer und Profil.
      name: 'EinstellungenSichern', argumente: () => [[]],
      vorher: (p, k) => { mkdirSync(join(k.ort.ort, 'vertrag'), { recursive: true }); return p.frage({ was: 'dialogantwort', datei: austausch(k) }) },
    },
    // Ein Unterschied, den die Vorschau zeigt und die Übernahme zurückholt.
    { name: 'EinstellungSetzen', argumente: () => ['feldtiefe', 'normal'] },
    {
      name: 'EinstellungenEinlesen', argumente: () => [],
      vorher: (p, k) => p.frage({ was: 'dialogantwort', datei: austausch(k) }),
      // `revision` und `inhalt` gehören zusammen: die Übernahme rechnet die
      // Vorschau noch einmal und vergleicht (`Einstellungsaustausch.Uebernehmen`).
      nimm: '({ inhalt: a.sicht.inhalt, revision: a.sicht.revision })',
      nach: (w, k) => { k.aAustausch = w },
    },
    {
      // `nur` nennt Zeilen als `<ebene>:<schluessel>`.
      name: 'EinstellungenUebernehmen',
      argumente: (k) => [k.aAustausch.inhalt, k.aAustausch.revision, ['profil:feldtiefe']],
    },
    // Auf die Vorgabe zurück.
    { name: 'EinstellungSetzen', argumente: () => ['feldtiefe', 'normal'] },

    // --- Fenster- und Tischbild des einen Nutzers. Die Dateidialoge beantwortet
    // die Probe mit einem kleinen PNG (`bildvorher`).
    { name: 'FensterbildWaehlen', argumente: () => [nutzer], vorher: bildvorher('fensterbild') },
    { name: 'TischbildWaehlen', argumente: () => [nutzer], vorher: bildvorher('tischbild') },
    { name: 'FensterbildEntfernen', argumente: () => [nutzer] },
    { name: 'TischbildEntfernen', argumente: () => [nutzer] },

    // --- Das Geheimnis eines Anmeldeanbieters. GitHub verlangt eines
    // (`GeheimnisPflicht`), Microsoft verbietet es; der Wert ist ein erfundener
    // Platzhalter im eigenen Schlüsselbund der Probe und wird am Ende gelöscht.
    { name: 'AnbietergeheimnisSetzen', argumente: () => ['github', 'vertrag-platzhalter-ohne-wert'] },
    { name: 'AnbietergeheimnisLoeschen', argumente: () => ['github'] },
  ],
}
