// **Das Alter der Sicherung und der freiwillige Plan, am nativen Fenster** —
// OPS01 (#2347),
// Zielartboard `Betrieb09`, Fläche „Sicherung".
//
// # Was hier geschieht
//
// 1. **Vorher**: unter *Einstellungen › Daten und Sicherung* steht
//    „unbekannt" — ohne Datum und nie „heute".
// 2. **Von Hand sichern** über den Dialog: danach „heute" mit Fläche und
//    Datei, und dieselbe Antwort an der Kommandozeile.
// 3. **Der Plan, und zwei Prozesse**: hält ein anderer Prozess die Sperre
//    neben dem Bestand, sichert der Plan nicht; ist sie frei, sichert er nach
//    der Ordnerwahl sofort — eingeschaltet **mit der Tastatur**. Eine fremde
//    Datei und eine unter dem Namen des Plans, die niemand vermerkt hat,
//    bleiben bytegleich.
// 4. **Neustart**: der Plan ist nicht mehr fällig, nichts geht weg.
// 5. **Zurückspielen als ganzer Weg**: die Sicherung von Hand kommt zurück,
//    und das Alter heißt danach wieder „unbekannt" — der zurückgespielte
//    Bestand kennt die Sicherung nicht, aus der er stammt.
// 6. **Volle Platte**: ein eigenes Abbild (`hdiutil`, 64 MB). Voll beim
//    Einrichten weist die Ordnerprüfung ab; voll, als der Plan wieder fällig
//    wird, scheitert er, sagt es, und nichts darauf wird angefasst.
// 7. **Dunkel**: ein Vermerk von vor zwanzig Tagen, hervorgehoben.
//
// # Was es nicht belegt
//
// **Eine Woche** — die Fälligkeit nach sieben Tagen und das Ausdünnen auf
// sieben eigene belegen `internal/domain` und `internal/service`
// (`TestDerPlanDuenntNurEigeneAus`, mit Verweis und geänderter Datei).
// **VoiceOver** — die Sätze sind gelesen, gesprochen hat sie hier niemand.
import { execFileSync, spawn } from 'node:child_process'
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

// zurDatengruppe öffnet *Einstellungen › Daten und Sicherung*.
async function zurDatengruppe(p) {
  await p.taste(',', 'befehl')
  const gruppe = "[...document.querySelectorAll('button, a, [role=tab]')].find((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))"
  await p.warte('die Gruppe »Daten« steht zur Wahl', () => p.seite(`return !!(${gruppe})`), 20000)
  await p.seite(`(${gruppe}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(gruppe, 'die Gruppe »Daten«')
  await p.ruhe()
}

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}\\p{N}]+/u, '') === ${JSON.stringify(text)})`

async function klicke(p, text) {
  await p.warte(`»${text}« steht da`, () => p.seite(`return !!(${knopf(text)}) && !(${knopf(text)}).disabled`), 20000)
  await p.seite(`(${knopf(text)}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(text), `»${text}«`)
}

// abschnitt ist die Fläche „Letzte Sicherung"; ihr Text ist, was dort steht.
const abschnitt = "[...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText.trim() === 'Letzte Sicherung')"
const alterstext = (p) => p.seite(`return (${abschnitt})?.innerText ?? ''`)
const kaestchen = `(${abschnitt})?.querySelector('input[type=checkbox]')`

// zeige rückt das Ende der Fläche ins Bild — dort stehen Plan und Fehler.
async function zeige(p, name) {
  await p.seite(`(${abschnitt})?.scrollIntoView({ block: 'end' }); return true`)
  await p.ruhe()
  await p.bild(name)
}

// stehtDa wartet, bis jeder Satz in der Fläche steht.
async function stehtDa(p, was, saetze, frist = 30000) {
  return p.warte(was, async () => {
    const text = await alterstext(p)

    return saetze.every((satz) => text.includes(satz)) ? text : false
  }, frist)
}

// inhalt gibt die Dateien eines Ordners mit ihrem Inhalt, sortiert.
const inhalt = (ordner) => readdirSync(ordner).sort().map((name) => `${name}:${readFileSync(join(ordner, name)).toString('base64').slice(0, 40)}`)

// sperreHalten hält die Sperre des Plans aus einem **anderen Prozess** —
// `flock` über perl, das auf jedem Mac liegt.
async function sperreHalten(p, pfad) {
  const kind = spawn('perl', ['-e', 'use Fcntl qw(:flock); open(my $f, ">>", $ARGV[0]) or die; flock($f, LOCK_EX) or die; $| = 1; print "gesperrt\\n"; sleep 300', pfad])
  let gesagt = ''
  kind.stdout.on('data', (teil) => { gesagt += teil })
  await p.warte('der andere Prozess hält die Sperre', async () => gesagt.includes('gesperrt'), 10000)

  return kind
}

export default async function sicherung09(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })
  await zurDatengruppe(p)

  p.notiz('1 — vorher: unbekannt heißt unbekannt')
  const vorher = await stehtDa(p, 'die Fläche »Letzte Sicherung« steht da', ['Letzte Sicherung', 'unbekannt'], 20000)
  // **Der Satz selbst nennt „heute"** („hier steht nie ‚heute', ohne dass …") —
  // gefragt wird deshalb nach der Kennzeichnung oben und nach einem Datum.
  const kennzeichen = await p.seite(`return (${abschnitt})?.querySelector('h3 + *')?.innerText ?? ''`)
  p.pruefe('ohne Vermerk kein Datum und kein »heute«', vorher.includes('noch nie gesichert') && !vorher.includes('Zuletzt gesichert am') && !kennzeichen.includes('heute'), `${kennzeichen} — ${vorher.replace(/\n/g, ' · ')}`)
  await zeige(p, 'vorher')

  p.notiz('2 — von Hand sichern, über den Dialog')
  const hand = join(ort.ort, 'hand.msesicherung')
  await p.dialogantwort(hand)
  await klicke(p, 'Bestand sichern …')
  const nachHand = await stehtDa(p, 'das Alter nennt die Sicherung von eben', ['heute', 'über das Fenster in hand.msesicherung, zurückgelesen und geprüft'])
  p.pruefe('die Sicherung liegt da', existsSync(hand), nachHand.replace(/\n/g, ' · '))
  const draussen = JSON.parse(p.mse(ort, '--json', 'bestand-sicherungsstand'))
  p.pruefe('die Kommandozeile sagt dasselbe', draussen.stufe === 'frisch' && draussen.flaeche === 'fenster' && draussen.datei === 'hand.msesicherung', JSON.stringify(draussen))
  await zeige(p, 'nach-der-sicherung')

  p.notiz('3 — der Plan: zwei Prozesse, dann nach der Ordnerwahl')
  const plan = join(ort.ort, 'plan')
  mkdirSync(plan)
  writeFileSync(join(plan, 'notizen.txt'), 'meine Notizen')
  writeFileSync(join(plan, 'bestand-plan-20200101-000000.msesicherung'), 'fremd, unter dem Namen des Plans')
  const fremdVorher = inhalt(plan)

  const anderer = await sperreHalten(p, `${ort.bestand}.plan`)
  await p.dialogantwort(plan)
  await p.klickeAuf(kaestchen, 'das Kästchen des Plans')
  await stehtDa(p, 'der Plan ist eingerichtet', [`Ordner: ${plan}`, 'Nach Plan noch nie gesichert'])
  await new Promise((los) => setTimeout(los, 5000))
  p.pruefe('mit der Sperre bei einem anderen sichert der Plan nicht', readdirSync(plan).length === 2 && (await alterstext(p)).includes('Nach Plan noch nie gesichert'), readdirSync(plan).join(', '))
  anderer.kill()
  await zeige(p, 'plan-gesperrt')

  // **Mit der Tastatur**: aus, und wieder an — ein neues Einrichten fragt
  // den Plan sofort.
  await p.seite(`${kaestchen}.focus(); return true`)
  await p.taste(' ')
  await stehtDa(p, 'der Plan ist aus', ['Wöchentlich in einen Ordner sichern'])
  await p.warte('das Kästchen ist leer', () => p.seite(`return ${kaestchen}?.checked === false`))
  await p.dialogantwort(plan)
  await p.seite(`${kaestchen}.focus(); return true`)
  await p.taste(' ')
  const nachPlan = await stehtDa(p, 'der Plan hat gesichert', ['Zuletzt nach Plan am', 'nach Plan in bestand-plan-'], 60000)
  const eigene = readdirSync(plan).filter((name) => /^bestand-plan-\d{8}-\d{6}\.msesicherung$/.test(name) && name !== 'bestand-plan-20200101-000000.msesicherung')
  p.pruefe('eine Sicherung des Plans liegt im Ordner', eigene.length === 1, eigene.join(', '))
  p.pruefe('die fremden Dateien sind bytegleich', JSON.stringify(inhalt(plan).filter((z) => !z.startsWith(eigene[0]))) === JSON.stringify(fremdVorher))
  p.pruefe('kein Fehler des Plans', !nachPlan.includes('konnte zuletzt nicht sichern'))
  await zeige(p, 'plan-gesichert')

  p.notiz('4 — Neustart: nichts ist fällig, nichts geht weg')
  const vorDemNeustart = inhalt(plan)
  p.pruefe('⌘Q beendet', await p.beende())
  await p.starte({ ort })
  await zurDatengruppe(p)
  await stehtDa(p, 'nach dem Neustart steht der Plan da', [`Ordner: ${plan}`, 'Zuletzt nach Plan am'])
  await new Promise((los) => setTimeout(los, 5000))
  p.pruefe('der Ordner ist unverändert', JSON.stringify(inhalt(plan)) === JSON.stringify(vorDemNeustart))

  p.notiz('5 — die Sicherung von Hand zurückspielen')
  await p.dialogantwort(hand)
  await klicke(p, 'Sicherung zurückspielen …')
  await p.wartetext('Diese Sicherung zurückspielen?', 20000)
  await klicke(p, 'Zurückspielen')
  await p.wartetext('Zurückgespielt.', 30000)
  const zurueck = await stehtDa(p, 'das Alter nach dem Zurückspielen', ['Letzte Sicherung'])
  p.pruefe('der zurückgespielte Bestand kennt seine eigene Sicherung nicht — unbekannt', zurueck.includes('unbekannt'), zurueck.replace(/\n/g, ' · '))
  p.pruefe('der Ordner des Plans ist auch danach unverändert', JSON.stringify(inhalt(plan)) === JSON.stringify(vorDemNeustart))
  await zeige(p, 'zurueckgespielt')
  p.pruefe('⌘Q beendet', await p.beende())

  await vollePlatte(p)
  await dunkel(p)
}

// vollePlatte prüft beide Lagen einer vollen Platte: **voll beim Einrichten**
// — dann weist die Ordnerprüfung ab (16 MB Mindestplatz) — und **voll, als der
// Plan wieder fällig wird**: dann scheitert er, sagt es, und nichts auf dem
// Abbild wird angefasst. Fällig wird er, weil der Vermerk seiner ersten
// Sicherung acht Tage zurückgestellt wird; gefragt wird nach dem Start erst
// nach zwei Minuten (`planverzug`).
async function vollePlatte(p) {
  p.notiz('6 — volle Platte')
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const abbild = join(ort.ort, 'voll.dmg')
  const halt = join(ort.ort, 'voll')
  mkdirSync(halt)
  execFileSync('hdiutil', ['create', '-size', '64m', '-fs', 'HFS+', '-volname', 'mse-voll', abbild], { stdio: 'ignore' })
  execFileSync('hdiutil', ['attach', '-nobrowse', '-mountpoint', halt, abbild], { stdio: 'ignore' })
  const sichtbar = () => readdirSync(halt).filter((name) => !name.startsWith('.')).sort()
  const frei = () => Number(execFileSync('df', ['-k', halt]).toString().trim().split('\n').at(-1).split(/\s+/)[3]) * 1024
  const fuellen = () => writeFileSync(join(halt, 'fuellung'), Buffer.alloc(Math.max(frei() - 4096, 0)))
  try {
    writeFileSync(join(halt, 'fremde.txt'), 'gehört jemand anderem')

    p.notiz('6a — voll beim Einrichten: abgewiesen, nichts entsteht')
    fuellen()
    p.zeige(`frei auf dem Abbild: ${String(Math.round(frei() / 1024))} KB`)
    const vorEinrichten = sichtbar()
    await p.starte({ ort })
    await zurDatengruppe(p)
    await stehtDa(p, 'die Fläche steht da', ['Letzte Sicherung', 'unbekannt'], 20000)
    await p.dialogantwort(halt)
    await p.klickeAuf(kaestchen, 'das Kästchen des Plans')
    const abgewiesen = await stehtDa(p, 'die Wahl wird abgewiesen', ['sind nur noch'], 20000)
    p.pruefe('kein Plan über einem vollen Ordner', !abgewiesen.includes('Ordner: ') && (await p.seite(`return ${kaestchen}?.checked === false`)), abgewiesen.replace(/\n/g, ' · '))
    p.pruefe('auf dem Abbild liegt, was vorher dalag', JSON.stringify(sichtbar()) === JSON.stringify(vorEinrichten), sichtbar().join(', '))
    await zeige(p, 'voll-beim-einrichten')

    p.notiz('6b — voll, als der Plan wieder fällig wird')
    rmSync(join(halt, 'fuellung'))
    await p.dialogantwort(halt)
    await p.klickeAuf(kaestchen, 'das Kästchen des Plans')
    await stehtDa(p, 'die erste Sicherung nach Plan', ['Zuletzt nach Plan am', 'nach Plan in bestand-plan-'], 60000)
    p.pruefe('⌘Q beendet', await p.beende())

    const damals = new Date(Date.now() - 8 * 24 * 3600 * 1000).toISOString().replace(/\.\d{3}Z$/, '.000Z')
    execFileSync('sqlite3', [ort.bestand, `UPDATE sicherungsvermerke SET gesichert = '${damals}' WHERE anlass = 'plan';`])
    fuellen()
    const vorher = sichtbar()
    p.zeige(`vorher: ${vorher.join(', ')} — frei ${String(Math.round(frei() / 1024))} KB`)

    await p.starte({ ort })
    await zurDatengruppe(p)
    const gemeldet = await stehtDa(p, 'der Plan sagt, dass er nicht sichern konnte', ['Der Plan konnte zuletzt nicht sichern'], 240000)
    p.pruefe('die Meldung nennt die volle Platte', /voll|Platz|space/i.test(gemeldet), gemeldet.replace(/\n/g, ' · '))
    p.pruefe('auf dem Abbild liegt, was vorher dalag — die eigene Sicherung bleibt, nichts dazu', JSON.stringify(sichtbar()) === JSON.stringify(vorher), sichtbar().join(', '))
    p.pruefe('keine halbe Sicherung blieb zurück', !readdirSync(halt).some((name) => name.includes('.unfertig') || name.startsWith('.sicherung-')), readdirSync(halt).join(', '))
    p.pruefe('die fremde Datei ist bytegleich', readFileSync(join(halt, 'fremde.txt'), 'utf8') === 'gehört jemand anderem')
    await zeige(p, 'voll')
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    for (let versuch = 0; versuch < 3; versuch++) {
      try {
        execFileSync('hdiutil', ['detach', halt], { stdio: 'ignore' })
        break
      } catch {
        execFileSync('sleep', ['1'])
      }
    }
  }
}

// dunkel zeigt eine Sicherung von vor zwanzig Tagen im dunklen Bild.
async function dunkel(p) {
  p.notiz('7 — dunkel, eine Sicherung von vor zwanzig Tagen')
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const damals = new Date(Date.now() - 20 * 24 * 3600 * 1000).toISOString().replace(/\.\d{3}Z$/, '.000Z')
  execFileSync('sqlite3', [ort.bestand, `INSERT INTO sicherungsvermerke (gesichert, anlass, flaeche, datei, groesse, pruefsumme) VALUES ('${damals}', 'hand', 'cli', 'damals.msesicherung', 1, '${'0'.repeat(64)}');`])
  await p.starte({ ort })
  await zurDatengruppe(p)
  const alt = await stehtDa(p, 'die alte Sicherung ist hervorgehoben', ['vor 20 Tagen · älter als 14 Tage', 'über die Kommandozeile in damals.msesicherung'], 20000)
  p.pruefe('Mantine steht dunkel', (await p.seite('return document.documentElement.dataset.mantineColorScheme')) === 'dark', alt.replace(/\n/g, ' · '))
  await zeige(p, 'alt-dunkel')
  p.pruefe('⌘Q beendet', await p.beende())
}
