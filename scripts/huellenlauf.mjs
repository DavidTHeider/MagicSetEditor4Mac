// **Der Lauf über die Drehbücher** — der zweite Teil der Hüllenprobe
// (`scripts/huellenprobe.mjs`): prüft das Bündel, wählt die Drehbücher, fährt
// jedes mit einer frischen Probe und schreibt seinen Beleg.
//
// **Eine eigene Datei**, seit die Probe über die Zeilengrenze aus §2 wuchs
// (2026-10-05). Die Klasse `Probe` reicht der Aufrufer hinein — ein Import von
// hier zurück wäre ein Ring zwischen zwei Modulen, von denen eines mit einem
// `await` auf oberster Ebene endet.
import { execFileSync } from 'node:child_process'
import { copyFileSync, existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { userInfo } from 'node:os'
import { join } from 'node:path'

import { bauabdruck, dateiabdruck, laeufeRaeumen, laufkennung, verzeichnisabdruck } from './baustand.mjs'
import { bildschirmAus, vermerkeLauf } from './bildschirmzeit.mjs'
import { abweichungen, anwendungLaeuft, kontrollprobe, rechnerabdruck, unlesbare } from './huellenorte.mjs'
import { offeneZeilen } from './huellenbeleg.mjs'
import { cspVerstoesse } from './huellenschutz.mjs'
import { rufbuchHolen } from './huellenrufbuch.mjs'

// drehbuchAus gibt das Drehbuch eines Moduls — seine Vorgabefunktion — oder
// nichts. **Ein Modul ohne sie ist ein Teil**, den mehrere Drehbücher teilen
// (`gastlauf`, `aufstiegsteile`, `budgetsorte09`), und kein roter Lauf.
export function drehbuchAus(modul) {
  return typeof modul.default === 'function' ? modul.default : undefined
}

// Laufbehalt ist die Zahl der Läufe, deren Verzeichnis stehen bleibt.
const Laufbehalt = 10

export async function fahreDrehbuecher(Probe, { wurzel, buendel, abnahmekennung, namen: genannt }) {
  if (!existsSync(buendel)) {
    console.error('Aufruf: node scripts/huellenprobe.mjs [<drehbuch> …] (vorher ./scripts/bauen.sh abnahme)')
    process.exit(2)
  }
  // **Das Bündel der Probe trägt eine eigene Kennung** — mit der der
  // Anwendung teilte es deren Vorgaben (die Erinnerung der Dateidialoge) und
  // den Speicher ihrer Webansicht (`localStorage`). Gefunden am Beleg von
  // `eintraege`: der Dialog öffnete, wo die Anwendung dieses Rechners zuletzt
  // war.
  const kennung = execFileSync('plutil', ['-extract', 'CFBundleIdentifier', 'raw', '-o', '-', join(buendel, 'Contents/Info.plist')]).toString().trim()
  if (kennung !== abnahmekennung) {
    console.error(`Das Bündel trägt die Kennung ${kennung} und nicht ${abnahmekennung}: es teilte Vorgaben und Webspeicher mit der Anwendung dieses Rechners (./scripts/bauen.sh abnahme)`)
    process.exit(2)
  }
  // **Jeder Lauf beginnt ohne die Vorgaben des vorigen** — die eigenen der
  // Probe, keine fremden.
  try {
    execFileSync('defaults', ['delete', abnahmekennung], { stdio: 'ignore' })
  } catch {
    // noch keine
  }
  const lauf = laufVorbereiten(wurzel, buendel)
  Probe.laufort = lauf.ort

  // **Ohne Namen fährt sie jedes Drehbuch** — die Liste ist das Verzeichnis,
  // nicht eine Aufzählung daneben.
  const namen = genannt.length > 0
    ? genannt
    : readdirSync(join(wurzel, 'scripts/huellenprobe')).filter((n) => n.endsWith('.mjs')).map((n) => n.slice(0, -4)).sort()
  // **Die echten Orte dieses Rechners, je Drehbuch vor und nach dem Lauf** —
  // T01 (#2095). Das Zuhause aus dem Benutzereintrag und nicht aus `HOME`:
  // wer die Probe selbst unter einem anderen startet, prüft sonst das falsche.
  const echt = userInfo().homedir
  // **Und einmal je Lauf die Positivkontrolle** — A01 (#2215): ein leerer
  // Vergleich sagt nur dann etwas, wenn derselbe Abdruck eine absichtlich
  // veränderte Datei sieht.
  if (!kontrollprobe().erkannt) {
    console.error('Die Positivkontrolle des Ortsabdrucks sah ihre veränderte Datei nicht — ein Vergleich der echten Orte sagte nichts.')
    process.exit(2)
  }
  let rot = 0
  for (const name of namen) {
    const modul = await import(join(wurzel, 'scripts/huellenprobe', `${name}.mjs`))
    // **Ein Diagnosedrehbuch fährt nur, wer es nennt** (`nurEinzeln`) —
    // `dauerteile` (B04) kostete den vollen Lauf sonst 36 Minuten.
    if (genannt.length === 0 && modul.nurEinzeln) {
      console.log(`– ${name}: nur einzeln, im vollen Lauf ausgelassen`)
      continue
    }
    const drehbuch = drehbuchAus(modul)
    if (!drehbuch) {
      console.log(`– ${name}: kein Drehbuch, ein Teil für andere`)
      continue
    }
    const probe = new Probe(name)
    probe.lauf = lauf
    const beginn = new Date().toISOString()
    const vorher = rechnerabdruck(echt)
    Probe.orte = []
    try {
      await drehbuch(probe)
    } catch (grund) {
      probe.pruefe('das Drehbuch lief durch', false, grund.message)

      // **Ein Abbruch hinterlässt sein Fenster.** Bis zum 2026-10-05 blieb
      // das Bildverzeichnis eines abgebrochenen Drehbuchs leer, und jeder
      // rote Lauf brauchte einen zweiten, um zu sehen, wo er stand.
      try {
        await probe.bild('abbruch')
      } catch {
        // kein Fenster mehr — dann gibt es auch nichts zu sehen
      }
    }
    // **Das Rufbuch auch eines abgebrochenen Drehbuchs** — B01 (#2220): die
    // Anwendung steht dann noch, und ihre Rufe zählen mit.
    if (probe.leitung) await rufbuchHolen(probe)
    mkdirSync(join(lauf.ort, 'rufbuch'), { recursive: true })
    writeFileSync(join(lauf.ort, 'rufbuch', `${name}.json`), `${JSON.stringify(probe.rufbuch ?? {}, null, 2)}\n`)
    probe.aufraeumen()
    const nachher = rechnerabdruck(echt)
    const verändert = abweichungen(vorher, nachher)
    // **Läuft die Anwendung dieses Rechners nebenher, verändert sie dieselben
    // Orte** — rot bleibt es trotzdem, aber mit dem Grund daneben.
    const nebenher = verändert.length > 0 && anwendungLaeuft() ? ' (die Anwendung dieses Rechners lief nebenher)' : ''
    probe.pruefe('die echten Orte dieses Rechners blieben unberührt', verändert.length === 0, verändert.join('; ') + nebenher)
    // **Und sie ließen sich lesen** — Nachlauf über Phase A: ein Ort, der sich
    // vorher und nachher nicht lesen ließ, gibt zwei gleiche Abdrücke und
    // damit ein „unberührt", das nichts prüft.
    const ungelesen = [...new Set([...unlesbare(vorher), ...unlesbare(nachher)])]
    probe.pruefe('die echten Orte ließen sich vollständig lesen', ungelesen.length === 0, ungelesen.slice(0, 3).join(', '))
    // **Die Schutzrichtlinie nahm dem Drehbuch nichts weg** — SEC-17 (#2164):
    // jeder Verstoß, den es nicht erwartet, ist eine Ressource, die fehlte.
    const verstoesse = cspVerstoesse(Probe.orte, probe.erwarteteVerstoesse)
    probe.pruefe('die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht', verstoesse.length === 0,
      verstoesse.slice(0, 3).map((z) => z.slice(z.indexOf('anweisung='))).join(' | '))
    const ziel = probe.beleg()
    const kopie = join(lauf.ort, 'belege', `${name}.md`)
    copyFileSync(ziel, kopie)
    lauf.faelle.push({
      name,
      befunde: probe.fehler,
      ohneBild: probe.ohneBild,
      bindungen: Object.keys(probe.rufbuch ?? {}).length,
      beginn,
      ende: new Date().toISOString(),
      belege: [ziel, kopie],
    })
    console.log(`${probe.fehler === 0 ? '✓' : '✗'} ${name} → ${ziel}`)
    rot += probe.fehler
  }
  lauf.ende = new Date().toISOString()
  // **War der Bildschirm aus, während ein Drehbuch lief?** — C03 (#2228).
  // Dann ist sein Rot zuerst die Umgebung, und Beleg wie Manifest sagen es;
  // **das Urteil bleibt**, denn ob die Anwendung bei hellem Bildschirm
  // bestanden hätte, weiß niemand. Ein unlesbares Protokoll heißt `null` —
  // nicht „war an".
  lauf.bildschirmAus = bildschirmAus(lauf.beginn, lauf.ende) ?? null
  if (lauf.bildschirmAus === null) console.log('  ? ob der Bildschirm an war, ist unbekannt — das Energieprotokoll (pmset -g log) ließ sich nicht lesen oder deckt den Beginn nicht ab')
  const imDunkeln = vermerkeLauf(lauf.faelle, lauf.bildschirmAus ?? [], {
    lesen: (pfad) => readFileSync(pfad, 'utf8'),
    schreiben: writeFileSync,
  })
  if (imDunkeln.length > 0) console.log(`  ⚠ der Bildschirm war aus, während ${String(imDunkeln.length)} Drehbücher liefen: ${imDunkeln.join(' ')}`)
  writeFileSync(join(lauf.ort, 'manifest.json'), `${JSON.stringify(lauf, null, 2)}\n`)
  console.log(`  Lauf ${lauf.kennung} → ${lauf.ort}`)
  process.exit(rot === 0 ? 0 : 1)
}

// laufVorbereiten prüft das Bündel gegen den Baum, baut `mse` und legt das
// Verzeichnis des Laufs an — T06 (#2099).
//
// **Ein Bündel aus einem anderen Stand bricht ab**: es misst nicht, was im
// Baum steht. Wer das bewusst will, setzt `MSE_HUELLE_ANDERER_STAND=1`; dann
// sagen Manifest und jeder Beleg es. **`mse` wird hier gebaut**, weil ein
// Aufruf ohne `huellenprobe.sh` sonst das von gestern nahm.
//
// **Jeder Lauf hat sein Verzeichnis** unter `.huellenprobe/laeufe/` — Bilder,
// Belege, Manifest —, und die jüngsten zehn bleiben.
function laufVorbereiten(wurzel, buendel) {
  const stempel = join(buendel, 'Contents/Resources/baustand.json')
  const stand = existsSync(stempel) ? JSON.parse(readFileSync(stempel, 'utf8')) : undefined
  const abdruck = bauabdruck(wurzel)
  const veraltet = stand?.abdruck !== abdruck
  if (veraltet && process.env.MSE_HUELLE_ANDERER_STAND !== '1') {
    console.error(`Das Bündel ist nicht aus diesem Stand gebaut (Bündel ${stand?.abdruck?.slice(0, 12) ?? 'ohne Baustand'}, Baum ${abdruck.slice(0, 12)}) — ./scripts/bauen.sh abnahme`)
    process.exit(2)
  }
  execFileSync('bash', ['-c', 'source scripts/activate.sh >/dev/null 2>&1 && go build -o .go/bin/mse ./cmd/mse'], { cwd: wurzel, stdio: 'inherit' })

  const laeufe = join(wurzel, '.huellenprobe/laeufe')
  const kennung = laufkennung()
  const ort = join(laeufe, kennung)
  mkdirSync(join(ort, 'belege'), { recursive: true })
  laeufeRaeumen(laeufe, Laufbehalt)
  const offen = offeneZeilen(wurzel)

  return {
    kennung,
    ort,
    beginn: new Date().toISOString(),
    head: execFileSync('git', ['-C', wurzel, 'rev-parse', 'HEAD']).toString().trim(),
    offen,
    bauabdruck: abdruck,
    veraltet,
    buendel: { stand: stand ?? null, programm: dateiabdruck(join(buendel, 'Contents/MacOS/MagicSetEditor4Mac')) },
    oberflaeche: verzeichnisabdruck(join(wurzel, 'frontend/dist')),
    programme: Object.fromEntries(['mse', 'mse-rest', 'mse-mcp'].map((n) => [n, dateiabdruck(join(wurzel, '.go/bin', n))])),
    umwelt: {
      macos: execFileSync('sw_vers', ['-productVersion']).toString().trim(),
      modell: execFileSync('sysctl', ['-n', 'hw.model']).toString().trim(),
      node: process.version,
      schalter: Object.fromEntries(Object.entries(process.env).filter(([n]) => n.startsWith('MSE_'))),
    },
    faelle: [],
  }
}
