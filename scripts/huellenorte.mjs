// **Die echten Orte dieses Rechners, vor und nach einem Probelauf** — T01
// (#2095)
// und A01 (#2215).
//
// Jede Probe läuft in einem eigenen Zuhause (`Probe.ort`). Ob das hält, sagt
// nicht die Absicht, sondern ein Abdruck der Orte, an die die Anwendung ohne
// eigenes Zuhause schriebe — vorher und nachher.
//
// # Was er sieht
//
// Name, Größe und Änderungszeit **jedes Eintrags in voller Tiefe**, auch der
// versteckten — unter allen Orten, die zur Anwendung dieses Rechners gehören:
// ihre eigenen Verzeichnisse **und** die, die macOS unter ihrer Kennung
// führt (Vorgaben, Webansicht, HTTP-Speicher, Fensterzustand). Bis A01 waren
// es zwei Orte bis zur zweiten Ebene; der Audit zu 0.8.0 sah die Vorgaben des
// Dateidialogs wachsen, die keiner davon enthielt.
//
// # Was er nicht sieht
//
// **Inhalte**: verglichen wird Größe und Zeit, nicht eine Prüfsumme — eine
// Datei, die mit gleicher Größe und gleicher Zeit umgeschrieben wird, bleibt
// unsichtbar. **Den Rest des Rechners**: was nicht zur Anwendung gehört,
// liegt außerhalb. Und eine Anwendung, die nebenher läuft, verändert die
// Orte auch — dann sagt [anwendungLaeuft] es, und ein Unterschied ist kein
// Beleg gegen die Probe.
import { execFileSync } from 'node:child_process'
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, utimesSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

// Produktkennung ist die Kennung des ausgelieferten Bündels — unter ihr
// legt macOS ab, was nicht `HOME` folgt. **Altkennung** ist die vor 0.9.0
// (Vermerk 0191, R02): der Umzug liest dort, und auch dort schreibt keine
// Probe.
export const Produktkennung = 'io.github.davidtheider.magicseteditor4mac'
export const Altkennung = 'com.wails.magicseteditor4mac'

// echteOrte sind die Orte der Anwendung unter einem Zuhause —
// `os.UserConfigDir` und `os.UserCacheDir` auf macOS (`internal/app/orte.go`,
// `internal/app/scryfall.go`) und dazu die unter der Kennung des Bündels.
export const echteOrte = (heim) => [
  join(heim, 'Library/Application Support/MagicSetEditor4Mac'),
  join(heim, 'Library/Caches/MagicSetEditor4Mac'),
  ...[Produktkennung, Altkennung].flatMap((kennung) => [
    join(heim, 'Library/Caches', kennung),
    join(heim, 'Library/Preferences', `${kennung}.plist`),
    join(heim, 'Library/WebKit', kennung),
    join(heim, 'Library/HTTPStorages', kennung),
    join(heim, 'Library/Saved Application State', `${kennung}.savedState`),
  ]),
]

// abdruck liest die Orte in voller Tiefe — Pfad → Größe:Zeit. **`lstat`**:
// ein Verweis wird als Verweis abgedruckt und nicht verfolgt; die Schriften
// eines Testzuhauses zeigen sonst auf den ganzen Schriftordner.
//
// **Ein Lesefehler ist ein Eintrag und kein Loch** (Zweitblick über Phase A):
// was fehlt, fehlt nicht im Abdruck, aber was sich nicht lesen ließ, steht
// als `unlesbar:<Grund>` darin — [unlesbare] nennt es, und ein Vergleich
// darüber ist **nicht prüfbar**. Bis dahin stand ein gesperrtes Verzeichnis
// mit seiner Zeit da, und eine überschriebene Datei darin blieb unsichtbar.
export function abdruck(orte, tiefe = Infinity) {
  const eintraege = new Map()
  const lesen = (pfad, ebene) => {
    let angaben
    try {
      angaben = lstatSync(pfad)
    } catch (fehler) {
      if (fehler.code !== 'ENOENT') eintraege.set(pfad, `unlesbar:${String(fehler.code)}`)
      return
    }
    eintraege.set(pfad, `${angaben.isDirectory() ? 'd' : String(angaben.size)}:${String(angaben.mtimeMs)}`)
    if (!angaben.isDirectory() || ebene >= tiefe) return
    let namen
    try {
      namen = readdirSync(pfad)
    } catch (fehler) {
      eintraege.set(pfad, `unlesbar:${String(fehler.code)}`)
      return
    }
    for (const name of namen) lesen(join(pfad, name), ebene + 1)
  }
  // **Ohne `existsSync`** (Nachlauf über Phase A): es sagt auch dann „nicht
  // da", wenn der Weg dorthin gesperrt ist; `lstat` unterscheidet beides.
  for (const ort of orte) lesen(ort, 0)

  return eintraege
}

// unlesbare nennt die Pfade eines Abdrucks, die sich nicht lesen ließen.
export const unlesbare = (eintraege) => [...eintraege].filter(([, wert]) => wert.startsWith('unlesbar:')).map(([pfad]) => pfad).sort()

// abweichungen nennt, was zwischen zwei Abdrücken neu ist, fehlt oder sich
// verändert hat — leer, wenn nichts.
export function abweichungen(vorher, nachher) {
  const gefunden = []
  for (const [pfad, wert] of nachher) {
    if (!vorher.has(pfad)) gefunden.push(`neu: ${pfad}`)
    else if (vorher.get(pfad) !== wert) gefunden.push(`verändert: ${pfad}`)
  }
  for (const pfad of vorher.keys()) if (!nachher.has(pfad)) gefunden.push(`fehlt: ${pfad}`)

  return gefunden.sort()
}

// kontrollprobe ist die Positivkontrolle des Abdrucks im Lauf — A01: ein
// eigener Ort, eine Datei darin, ein Abdruck, dann wird sie **absichtlich**
// verändert. Erkennt der zweite Abdruck sie nicht, sagt ein leerer
// Vergleich der echten Orte nichts.
export function kontrollprobe() {
  const ort = mkdtempSync(join(tmpdir(), 'huellenorte-kontrolle-'))
  const datei = join(ort, 'tief', 'kontrolle')
  try {
    mkdirSync(join(ort, 'tief'))
    writeFileSync(datei, 'vorher')
    const vorher = abdruck([ort])
    const spaeter = new Date(Date.now() + 60_000)
    writeFileSync(datei, 'nachher')
    utimesSync(datei, spaeter, spaeter)
    const gefunden = abweichungen(vorher, abdruck([ort]))

    return { erkannt: gefunden.some((zeile) => zeile.endsWith('/tief/kontrolle')), gefunden }
  } finally {
    rmSync(ort, { recursive: true, force: true })
  }
}

// anwendungLaeuft sagt, ob die Anwendung dieses Rechners gerade läuft — an
// ihrer **Kennung** über LaunchServices, nicht am Programmnamen: das
// Abnahmebündel heißt genauso und hat eine eigene.
export function anwendungLaeuft(frage = () => execFileSync('lsappinfo', ['find', `bundleid=${Produktkennung}`]).toString()) {
  try {
    return frage().trim() !== ''
  } catch {
    return false
  }
}

// kindumgebung ist die Umgebung eines Kindprozesses an einem Ort der Probe —
// Zuhause, Bestand, Bund. **An einer Stelle** (T01, #2095): drei Drehbücher
// setzten sie selbst und vergaßen `HOME`.
export const kindumgebung = (ort) => ({
  ...process.env,
  ...(ort.heim ? { HOME: ort.heim } : {}),
  MSE_BESTAND: ort.bestand,
  MSE_SCHLUESSELBUND: ort.bund,
  // **Kein verändernder Ruf an launchd** — ein gebautes `mse` ist kein
  // `*.test` (Nachlauf über Phase 2 von 0.9.0, Befund 6).
  MSE_DIENST_GESPERRT: 'an',
})

// rechnerabdruck ist der Abdruck der echten Orte unter `heim` und dazu die
// Suchliste der Schlüsselbunde — die Proben legen eigene Bunde an, und
// `security create-keychain` dürfte sie dort eintragen.
export function rechnerabdruck(heim) {
  const eintraege = abdruck(echteOrte(heim))
  // **Mit diesem Zuhause gefragt**: `security` liest die Suchliste aus `HOME`,
  // und ein Aufrufer unter dem Probe-Zuhause sähe sonst dessen Liste — die
  // WebKit-Abnahme meldete so beim ersten Lauf eine Veränderung, die keine war.
  eintraege.set('security list-keychains -d user', execFileSync('security', ['list-keychains', '-d', 'user'], {
    env: { ...process.env, HOME: heim },
  }).toString().trim())

  return eintraege
}

// Als Programm für `scripts/webkitabnahme.sh`, `scripts/vorpruefung.sh` und
// `scripts/ci.sh`:
//
//     node scripts/huellenorte.mjs abdruck <heim> > vorher.json
//     node scripts/huellenorte.mjs vergleich <heim> vorher.json
//
// **Der Vergleich gibt 0 bei unberührten Orten, 1 bei einem Unterschied und
// 3, wenn er nichts sagen kann** — weil die Anwendung nebenher lief oder die
// Positivkontrolle nichts sah.
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const [befehl, heim, datei] = process.argv.slice(2)
  if (befehl === 'abdruck' && heim) {
    process.stdout.write(JSON.stringify([...rechnerabdruck(heim)]))
  } else if (befehl === 'vergleich' && heim && datei) {
    const vorher = new Map(JSON.parse(readFileSync(datei, 'utf8')))
    const nachher = rechnerabdruck(heim)
    const gefunden = abweichungen(vorher, nachher)
    for (const zeile of gefunden) console.log(zeile)
    const ungelesen = [...new Set([...unlesbare(vorher), ...unlesbare(nachher)])]
    if (ungelesen.length > 0) {
      console.log(`nicht prüfbar: ${String(ungelesen.length)} Ort(e) ließen sich nicht lesen — ${ungelesen.slice(0, 3).join(', ')}`)
      process.exit(3)
    }
    const kontrolle = kontrollprobe()
    if (!kontrolle.erkannt) {
      console.log('nicht prüfbar: die Positivkontrolle sah ihre veränderte Datei nicht')
      process.exit(3)
    }
    if (gefunden.length > 0 && anwendungLaeuft()) {
      console.log(`nicht prüfbar: ${Produktkennung} lief nebenher und verändert dieselben Orte`)
      process.exit(3)
    }
    process.exit(gefunden.length === 0 ? 0 : 1)
  } else {
    console.error('Aufruf: node scripts/huellenorte.mjs abdruck <heim> | vergleich <heim> <datei>')
    process.exit(2)
  }
}
