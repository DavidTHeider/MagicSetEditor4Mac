// **Die Lagen des Starts, am nativen Fenster** — F05
// (#2344),
// Befunde A-05 und A-21.
//
// Die Abnahme: belegt, voll, nur lesend, gesperrt, beschädigt, älter und
// neuer je mit richtiger Klasse und Rat; ein gesunder Bestand wird nie zum
// Ersetzen angeboten; „Erneut versuchen" nach gelöster Belegung öffnet
// denselben Bestand — Inhalt und Prüfsumme wie erwartet.
//
// # Wie die Lagen entstehen — echt, nicht eingespeist
//
//   - **belegt**: ein `sqlite3` hält eine Schreibtransaktion offen
//     (`BEGIN IMMEDIATE`); der Start wartet seine fünf Sekunden und gibt auf.
//   - **gesperrt** und **nur lesbar**: die Rechte der Datei (`000`, `400`).
//   - **beschädigt**: die erste Seite hinter dem Kopf mit Nullen.
//   - **neuer** und **älter**: die Buchführung — eine Wanderung aus der
//     Zukunft dazu, eine der Grundlinie weg.
//   - **voll**: ein Plattenabbild von 40 MB, randvoll geschrieben.
//
// # Was es nicht belegt
//
// **Ein Fehler des Datenträgers** (`IOERR`) lässt sich hier nicht echt
// herstellen; seine Klasse belegen `startlagen_intern_test.go` mit dem Code
// von SQLite und `wiederherstellungslagen.test.tsx` mit der Seite. Und
// VoiceOver — das hört nur ein Mensch.
import { execFileSync, spawn } from 'node:child_process'
import { createHash } from 'node:crypto'
import { chmodSync, copyFileSync, mkdirSync, openSync, readFileSync, rmSync, writeSync, closeSync } from 'node:fs'
import { join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const saat = resolve(wurzel, 'saat/probe-werkstatt.mse-set')

const seite = "document.querySelector('[data-wiederherstellung]')"
const text = `return ${seite}?.innerText ?? ''`

export default async function startlagen09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', saat)
  const sets = (o = ort) => p.mse(o, 'sets')
  const vorher = sets()
  const pruefsumme = (pfad) => createHash('sha256').update(readFileSync(pfad)).digest('hex')
  const heil = join(ort.ort, 'heil.db')
  copyFileSync(ort.bestand, heil)
  const zurueck = () => copyFileSync(heil, ort.bestand)

  // starteZu startet und erwartet die Wiederherstellung mit diesem Satz.
  const starteZu = async (lage, satz, o = ort) => {
    await p.starte({ ort: o, sprachen: 'de' })
    const gezeigt = await p.warte(`${lage}: die Wiederherstellung steht`, async () => (await p.seite(text)) || false, 30000)
    p.zeige(gezeigt.replace(/\s+/g, ' ').slice(0, 300))
    p.pruefe(`${lage}: der Satz sagt, woran es liegt`, gezeigt.includes(satz))

    return gezeigt
  }
  const ersetztNicht = (lage, gezeigt) => {
    p.pruefe(`${lage}: „Erneut versuchen" steht da`, gezeigt.includes('Erneut versuchen'))
    p.pruefe(`${lage}: kein Zurückspielen und kein Weg zurück`, !gezeigt.includes('Sicherung zurückspielen') && !gezeigt.includes('Der Weg zurück'))
  }
  const erneut = async () => {
    const gefunden = await p.seite(`const k = [...${seite}.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Erneut versuchen'); k?.focus(); return document.activeElement === k`)
    p.pruefe('»Erneut versuchen« nimmt den Fokus', gefunden)
    await p.taste('Enter')
  }
  const offen = (lage) => p.warte(`${lage}: die Anwendung steht`, async () => !(await p.seite(`return !!${seite}`)) && (await p.seite("return document.querySelectorAll('nav button').length > 3")), 30000)

  p.notiz('Belegt: ein sqlite3 hält eine Schreibtransaktion')
  const halter = spawn('sqlite3', [ort.bestand], { stdio: ['pipe', 'ignore', 'ignore'] })
  halter.stdin.write('BEGIN IMMEDIATE;\nINSERT INTO bestand (schluessel, wert) VALUES (\'probe_halter\', \'1\');\n')
  await new Promise((los) => setTimeout(los, 1000))
  const summeVorher = pruefsumme(ort.bestand)
  ersetztNicht('belegt', await starteZu('belegt', 'Der Bestand ist gerade belegt'))
  await p.bild('f05-belegt')
  p.pruefe('belegt: der gescheiterte Start hat die Datei nicht verändert', pruefsumme(ort.bestand) === summeVorher)
  halter.stdin.write('ROLLBACK;\n.quit\n')
  halter.stdin.end()
  await new Promise((los) => halter.once('exit', los))
  await erneut()
  await offen('belegt, erneut versucht')
  p.pruefe('belegt: derselbe Bestand — dieselben Sets', sets() === vorher)
  await p.bild('f05-belegt-erneut-offen')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Gesperrt: die Datei ohne jedes Recht')
  chmodSync(ort.bestand, 0o000)
  ersetztNicht('gesperrt', await starteZu('gesperrt', 'Auf diesen Ort darf dieses Programm nicht zugreifen'))
  chmodSync(ort.bestand, 0o600)
  await erneut()
  await offen('gesperrt, freigegeben und erneut versucht')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Nur lesbar: die Datei ohne Schreibrecht')
  chmodSync(ort.bestand, 0o400)
  ersetztNicht('nur lesbar', await starteZu('nur lesbar', 'lässt sich lesen, aber nicht schreiben'))
  await p.bild('f05-nur-lesbar')
  chmodSync(ort.bestand, 0o600)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Beschädigt: die erste Seite hinter dem Kopf mit Nullen')
  const griff = openSync(ort.bestand, 'r+')
  writeSync(griff, Buffer.alloc(3000), 0, 3000, 100)
  closeSync(griff)
  const beschaedigt = await starteZu('beschädigt', 'beschädigt oder kein Bestand')
  p.pruefe('beschädigt: das Zurückspielen steht da', beschaedigt.includes('Sicherung zurückspielen'))
  p.pruefe('beschädigt: kein „Erneut versuchen"', !beschaedigt.includes('Erneut versuchen'))
  await p.bild('f05-beschaedigt')
  p.pruefe('⌘Q beendet', await p.beende())
  zurueck()

  p.notiz('Neuer: eine Wanderung aus der Zukunft')
  execFileSync('sqlite3', [ort.bestand, "INSERT INTO migrationen (name, angewendet, pruefsumme) VALUES ('9999_aus_der_zukunft.sql', '2026-10-14T00:00:00.000Z', 'x')"])
  const neuer = await starteZu('neuer', 'aus einer neueren Fassung')
  p.pruefe('neuer: die Datei ist in Ordnung — ersetze sie nicht', neuer.includes('ersetze sie nicht'))
  p.pruefe('⌘Q beendet', await p.beende())
  zurueck()

  p.notiz('Älter: eine Wanderung der Grundlinie fehlt')
  execFileSync('sqlite3', [ort.bestand, 'DELETE FROM migrationen WHERE name = (SELECT min(name) FROM migrationen)'])
  const aelter = await starteZu('älter', 'Fassung vor 0.7.0')
  p.pruefe('älter: die Datei ist in Ordnung — ersetze sie nicht', aelter.includes('ersetze sie nicht'))
  p.pruefe('⌘Q beendet', await p.beende())
  zurueck()

  p.notiz('Voll: ein Plattenabbild, randvoll')
  const abbild = join(ort.ort, 'voll.dmg')
  const platte = join(ort.ort, 'voll')
  mkdirSync(platte)
  execFileSync('hdiutil', ['create', '-size', '40m', '-fs', 'HFS+', '-volname', 'msevoll', '-type', 'UDIF', abbild])
  execFileSync('hdiutil', ['attach', '-nobrowse', '-mountpoint', platte, abbild])
  try {
    const vollort = { ...ort, bestand: join(platte, 'bestand.db') }
    copyFileSync(heil, vollort.bestand)
    const fuell = join(platte, 'fuellung')
    // **Erst in großen Stücken, dann in kleinen** (#2407): mit `bs=1m` allein
    // blieb bis zu ein Mebibyte frei — am 2026-10-15 waren es 1016 KB, und
    // der Start gelang, wo er am 2026-10-14 mit 12 KB scheiterte.
    for (const [ziel, groesse] of [[fuell, '1m'], [`${fuell}-rest`, '4k']]) {
      try {
        execFileSync('dd', ['if=/dev/zero', `of=${ziel}`, `bs=${groesse}`], { stdio: 'ignore' })
      } catch {
        // **dd endet am vollen Datenträger mit einem Fehler** — das ist der Zweck.
      }
    }
    p.zeige(execFileSync('df', ['-k', platte]).toString().split('\n')[1] ?? '')
    // **Beides, seit #2407** — `SQLITE_IOERR_SHMSIZE` ist kein Beleg für
    // eine volle Platte, der knappe Platz steht dabei.
    const voll = await starteZu('voll', 'kaum Platz frei', vollort)
    ersetztNicht('voll', voll)
    await p.bild('f05-voll')
    rmSync(fuell)
    rmSync(`${fuell}-rest`)
    await erneut()
    await offen('voll, Platz geschaffen und erneut versucht')
    p.pruefe('voll: derselbe Bestand — dieselben Sets', sets(vollort) === vorher)
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    execFileSync('hdiutil', ['detach', platte, '-force'])
  }

  p.notiz('Zum Schluss: der heile Bestand öffnet')
  await p.starte({ ort, sprachen: 'de' })
  p.pruefe('der heile Bestand ist offen', !(await p.seite(`return !!${seite}`)))
  p.pruefe('dieselben Sets wie am Anfang', sets() === vorher)
  p.pruefe('⌘Q beendet', await p.beende())
}
