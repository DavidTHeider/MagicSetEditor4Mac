// **Sets aus Dateien: einlesen, sichern, verpacken, öffnen, löschen** — eine
// Gruppe des Vertragsrundgangs (B01, #2220).
//
// **Alles hier arbeitet an Sets, die diese Gruppe selbst hereinholt.** Das
// Saatset der ersten Gruppe (`k.set`) wird weder gesichert noch gelöscht: ein
// Sichern schriebe in `saat/` des Repositoriums zurück, und ein Löschen nähme
// den anderen Gruppen den Boden.
//
// Der Ablauf:
//
//  1. `SetAusDatei` liest ein selbst geschriebenes `.mse-set` (ein Verzeichnis
//     mit der Hauptdatei `set`, wie `fix-f02.mjs` es beschreibt) aus dem
//     Zwischenverzeichnis der Probe — **Quelle ist dort, also schreibt
//     `SetSpeichern` auch dorthin zurück**.
//  2. `UmschlagSchnueren` (Dialog; hier nur als Vorbedingung) packt es in einen
//     Umschlag, `UmschlagUebernehmen` holt es als **zweites** Set wieder herein.
//  3. `BeispielsetOeffnen` schlägt ein mitgeliefertes Set auf.
//  4. Löschen am Ende, jedes Set mit seinem Namen als Bestätigung.
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

// Die Hauptdatei: ein Set auf dem Spiel und dem Stil des Saatsets, mit einer
// Karte. Der Titel ist der Name, den `Sets` später nennt.
const setdatei = (k) => [
  'mse version: 2.0.0',
  `game: ${k.set.spiel}`,
  `stylesheet: ${k.set.stil}`,
  'set info:',
  '\ttitle: Vertragsset A',
  '\tcopyright: Probe',
  'card:',
  '\tname: Vertragskarte',
  '\tfarbe: rot',
  '\ttext: Aus dem Rundgang',
  '',
].join('\n')

const setort = (k) => join(k.ort.ort, 'vertrag', 'vertragsset-a.mse-set')
const umschlag = (k) => join(k.ort.ort, 'vertrag', 'vertragsset-a.msepackage')

// nameVon sucht den Namen eines Sets in der letzten Liste von `Sets`.
const nameVon = (k, id) => k.aSetliste?.find((s) => s.id === id)?.name ?? ''

// Eine Liste von `Sets` merken — die Namen sind die Bestätigung zum Löschen.
const liste = {
  name: 'Sets', nimm: 'a.sets.map((s) => ({ id: s.id, name: s.name }))',
  nach: (w, k) => { k.aSetliste = w },
}

export const aSets = {
  name: 'Sets aus Dateien (A)',
  rufe: [
    {
      name: 'SetAusDatei', argumente: (k) => [setort(k)],
      vorher: (_p, k) => {
        mkdirSync(setort(k), { recursive: true })
        writeFileSync(join(setort(k), 'set'), setdatei(k))
      },
      nimm: 'a.setID', nach: (w, k) => { k.aSet = w },
    },
    liste,
    // **Mit Herkunft**: das Set kam aus einer Datei, ⌘S schreibt dorthin
    // zurück und fragt nicht nach einem Ort (`Bindungen.SetSpeichern`).
    { name: 'SetSpeichern', argumente: (k) => [k.aSet] },
    // **Ein Umschlag trägt die Peeridentität des Profils** — ohne sie lehnt
    // `UmschlagSchnueren` ab (zweiter Rundgang).
    { name: 'PeeridentitaetSicherstellen' },
    {
      // Der Ausgang eines Dialogs ist eine Datei: die Probe antwortet mit
      // einem Pfad statt mit dem echten Dialog.
      name: 'UmschlagSchnueren', argumente: (k) => ['set', k.aSet],
      vorher: (p, k) => p.frage({ was: 'dialogantwort', datei: umschlag(k) }),
    },
    {
      // **Ein neues Set und nicht dasselbe**: „eine Quelle, ein Set" gilt für
      // den Ort der Datei, und ein Umschlag entfaltet sich in ein
      // Wegwerfverzeichnis (`internal/app/umschlag.go`, `UmschlagUebernehmen`).
      name: 'UmschlagUebernehmen', argumente: (k) => [umschlag(k)],
      nimm: 'a.setID', nach: (w, k) => { k.aUmschlagset = w },
    },
    {
      name: 'Beispielsets', nimm: "a.sets.map((s) => s.pfad)",
      nach: (w, k) => { k.aBeispielpfad = w.find((p) => p.endsWith('probe-werkstatt.mse-set')) ?? w[0] ?? '' },
    },
    {
      name: 'BeispielsetOeffnen', argumente: (k) => [k.aBeispielpfad],
      nimm: 'a.setID', nach: (w, k) => { k.aBeispielset = w },
    },
    liste,

    // --- Ans Ende, was wegnimmt: jedes Set, das diese Gruppe hereingeholt hat.
    { name: 'SetLoeschen', argumente: (k) => [k.aUmschlagset, nameVon(k, k.aUmschlagset)] },
    {
      // **Das aufgeschlagene Beispielset ist vielleicht gar keins, das diese
      // Gruppe angelegt hat**: stammt es aus derselben Quelle wie das Saatset,
      // meldet der Import `schonOffen` und gibt dessen Kennung zurück. Dann
      // wird nichts gelöscht — der Ruf geht mit der Kennung `0` hinaus und
      // wird fachlich abgewiesen. Deshalb gilt hier „erreicht" und nicht
      // „gelungen"; jeder andere Ausgang (abgewiesen, abgebrochen) bleibt rot.
      name: 'SetLoeschen', erwartet: 'erreicht',
      argumente: (k) => (k.aBeispielset && k.aBeispielset !== k.set.id && k.aBeispielset !== k.aSet
        ? [k.aBeispielset, nameVon(k, k.aBeispielset)]
        : ['0', '']),
    },
    { name: 'SetLoeschen', argumente: (k) => [k.aSet, nameVon(k, k.aSet)] },
  ],
}
