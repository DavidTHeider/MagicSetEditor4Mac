// **Paniken in den Rückrufen der Hülle, am nativen Fenster** — F02
// (#2341),
// Befund A-02.
//
// Die Abnahme: Panik in **Finder, Ablegen, Menü, Schluss und Start**
// einspeisen; der Prozess endet nicht unbemerkt, der Schluss bleibt in 5 s
// und hinterlässt keinen halben Bestand. Die Panik kommt über die Naht des
// Abnahmebaus (`MSE_ABNAHME_PANIK`, `internal/fenster/probepanik_abnahme_darwin.go`):
// der genannte Rückruf wirft in seinem Riegel, ohne dass ein echter Fehler
// im Baum liegt.
//
// # Was es prüft
//
// Je Rückruf: die Zeile mit der Panik steht mit Vorgangskennung im
// Protokoll, und die Anwendung antwortet danach noch. Beim Schluss: ⌘Q
// beendet in der Frist, der Bestand ist danach derselbe und lesbar — **und
// keine Startmarke bleibt liegen**: der Schluss lief nach der Panik ein
// zweites Mal und räumte geordnet ab (Zweitblick über Phase 1, Befund 3;
// bis dahin belegte dieser Lauf nur, dass der Prozess weg war). Der Start
// bekommt einen eigenen Lauf — wirft er, gibt es keinen zweiten.
//
// # Was es nicht belegt
//
// Eine Panik in Wails' eigenem Code — sie geht an `panikMelden`, und den
// prüft `riegel_test.go`, nicht dieser Lauf. Und die Bindungen: dort fängt
// Wails, und was das Fenster dann zeigt, ist F03.
import { execFileSync } from 'node:child_process'
import { mkdirSync, readdirSync, readFileSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')

export default async function panik09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const bestand = () => p.mse(ort, 'sets') + p.mse(ort, 'karten', '1')
  const vorher = bestand()
  const protokoll = () => readFileSync(join(dirname(ort.bestand), 'mse.log'), 'utf8')
  const panik = (was) => protokoll().split('\n').find((z) => z.includes(`Probepanik in ${was}`)) ?? ''
  const lebt = () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3")
  const ablage = join(ort.ort, 'ablage')
  mkdirSync(ablage, { recursive: true })
  const kopie = (name) => {
    const ziel = join(ablage, name)
    execFileSync('cp', ['-R', resolve(wurzel, 'saat/probe-werkstatt.mse-set'), ziel])
    return ziel
  }
  const belegt = async (was) => {
    const zeile = await p.warte(`${was}: die Panik steht im Protokoll`, () => Promise.resolve(panik(was) || false), 15000)
    p.zeige(zeile.slice(0, 260))
    p.pruefe(`${was}: die Panik steht mit Vorgangskennung im Protokoll`, /kennung=[0-9a-f]{8}/.test(zeile))
    p.pruefe(`${was}: die Anwendung antwortet danach`, await lebt())
  }

  p.notiz('Menü, Finder, Ablegen und Schluss werfen')
  await p.starte({ ort, umgebung: { MSE_ABNAHME_PANIK: 'ein Befehl aus dem Menü|das Öffnen aus dem Finder|das Ablegen|der Schluss' } })
  await p.groesse(1440, 900)

  await p.klick('MagicSetEditor4Mac', 'Einstellungen …')
  await belegt('ein Befehl aus dem Menü')

  execFileSync('open', ['-a', buendel, kopie('finder.mse-set')])
  await belegt('das Öffnen aus dem Finder')

  const mitte = await p.seite('return { x: innerWidth / 2, y: innerHeight / 2 }')
  await p.frage({ was: 'ablegen', pfad: [kopie('abgelegt.mse-set')], x: mitte.x, y: mitte.y })
  await belegt('das Ablegen')
  await p.bild('f02-nach-drei-paniken')

  const anfang = Date.now()
  p.pruefe('⌘Q beendet trotz Panik im Schluss', await p.beende())
  const dauer = Date.now() - anfang
  p.zeige(`bis der Prozess weg war: ${dauer} ms`)
  p.pruefe('in der Frist des Schlusses (5 s) und einer Sekunde Spiel', dauer < 6000)
  const schluss = panik('der Schluss')
  p.zeige(schluss.slice(0, 260))
  p.pruefe('der Schluss: die Panik steht mit Vorgangskennung im Protokoll', /kennung=[0-9a-f]{8}/.test(schluss))
  p.pruefe('der Bestand ist derselbe und lesbar', bestand() === vorher)
  const marken = readdirSync(join(dirname(ort.bestand), 'Sitzungen')).filter((n) => /^fenster-/.test(n))
  p.pruefe('der Schluss räumte trotzdem ab: keine Marke liegt', marken.length === 0, marken.join(', '))

  p.notiz('Der Start wirft')
  await p.starte({ ort, umgebung: { MSE_ABNAHME_PANIK: 'der Start' } })
  await belegt('der Start')
  await p.bild('f02-nach-panik-im-start')
  p.pruefe('⌘Q beendet', await p.beende())
  p.pruefe('der Bestand ist auch danach derselbe', bestand() === vorher)
}
