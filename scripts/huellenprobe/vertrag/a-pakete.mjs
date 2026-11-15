// **Pakete und Paketdateien** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Legt ein eigenes Include an (`mse-include`, kein Spiel nötig), kopiert es,
// ersetzt es durch eine zweite Fassung, geht zurück, arbeitet an einer Datei
// darin und räumt beides am Ende wieder weg. **Gelöscht wird nur, was diese
// Gruppe selbst angelegt hat** — die Saat des Rundgangs bleibt unberührt.
//
// Warum ein Include: `PaketAnlegen` verlangt bei einem Stil ein Spiel, das es
// im verwalteten Verzeichnis geben muss (`internal/service/paketanlage.go`,
// `pruefe`); ein Include braucht nichts und ist trotzdem ein vollwertiges
// Paket mit Hauptdatei `include`.
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const art = 'mse-include'
const name = 'vertrag-inc'
const kopie = 'vertrag-inc-kopie'

// ersatzort ist die Quelle der zweiten Fassung — ein Verzeichnis mit dem
// Namen `<name>.<art>`, denn aus dem Verzeichnisnamen liest `paketLesen`, welches
// Paket gemeint ist (`internal/service/paketlesen.go`).
const ersatzort = (k) => join(k.ort.ort, 'vertrag', 'ersatz', `${name}.${art}`)

const ersatzinhalt = [
  'mse version: 2.0.0',
  `short name: ${name}`,
  'full name: Vertrags-Include, zweite Fassung',
  // **Später als heute**: das Anlegen schreibt das Tagesdatum, und eine
  // Fassung, die älter wäre, sähe in der Paketübersicht nach einem Rückschritt aus.
  'version: 2026-12-31',
  '',
].join('\n')

export const aPakete = {
  name: 'Pakete und Paketdateien (A)',
  rufe: [
    {
      name: 'PaketAnlegen', argumente: () => [art, name, 'Vertrags-Include', ''],
      nimm: 'a.paket.pfad', nach: (w, k) => { k.aPaketpfad = w },
    },
    { name: 'PaketKopieren', argumente: () => [art, name, kopie] },
    // **Mit Bestätigung**, obwohl das frische Paket unverändert ist: so geht
    // der Ruf in jedem Fall durch, auch wenn der Abgleich es als geändert führte.
    {
      name: 'PaketErsetzen', argumente: (k) => [ersatzort(k), true],
      vorher: (_p, k) => {
        mkdirSync(ersatzort(k), { recursive: true })
        writeFileSync(join(ersatzort(k), 'include'), ersatzinhalt)
      },
    },
    // **Zurück tauscht die beiden Fassungen** — danach liegt wieder die
    // erste da, und die zweite ist die vorige.
    { name: 'PaketZurueck', argumente: () => [art, name] },
    // Die Datei eines Pakets heißt im Dienst nach seinem **Pfad**
    // (`Paketwerkstatt.ort`), nicht nach seinem Namen.
    {
      name: 'PaketdateiAnlegen', argumente: (k) => [k.aPaketpfad, 'notiz.txt', 'Vertrag, erste Fassung\n'],
      nimm: 'a.stand', nach: (w, k) => { k.aStand = w },
    },
    {
      name: 'PaketdateiSchreiben', argumente: (k) => [k.aPaketpfad, 'notiz.txt', 'Vertrag, zweite Fassung\n', k.aStand],
      nimm: 'a.stand', nach: (w, k) => { k.aStand = w },
    },
    // **Mit dem Stand, den das Schreiben gab** — eine Vorlage lässt sich nur
    // gegen ihren Stand umbenennen oder entfernen (`mitStand`).
    { name: 'PaketdateiUmbenennen', argumente: (k) => [k.aPaketpfad, 'notiz.txt', 'notiz-neu.txt', k.aStand] },
    { name: 'PaketdateiEntfernen', argumente: (k) => [k.aPaketpfad, 'notiz-neu.txt', k.aStand] },
    // Ans Ende, was wegnimmt.
    { name: 'PaketEntfernen', argumente: () => [art, kopie] },
    { name: 'PaketEntfernen', argumente: () => [art, name] },
  ],
}
