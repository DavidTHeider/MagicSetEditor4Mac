#!/usr/bin/env node
// **Quick Look für Setdateien — machbar?** — G06
// (#2253).
//
//     node scripts/quicklookgast.mjs [<vm>]
//
// Baut am Wirt ein Probebündel mit der Erweiterung aus
// `scripts/huellenprobe/gast/quicklook/Vorschaubild.swift`, signiert beides
// **ad hoc** wie die Anwendung — die Erweiterung mit Sandbox, wie macOS sie
// für Erweiterungen verlangt —, und fährt `quicklook.pl` im Gast. **Nie am
// Wirt**: eine registrierte Erweiterung ändert, was der Finder zeigt. Der
// Beleg geht nach `docs/belege/g06-quicklook.md`.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { fahre, gastziel, imGeteilten, wurzel } from './huellenprobe/gastlauf.mjs'

// **Nie am Wirt** — `gastziel` weist `--wirt` und jeden Namen ab, der kein Gast ist.
const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })

const quelle = join(wurzel, 'scripts/huellenprobe/gast/quicklook')
const beginn = new Date()
// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung,
// nach einer Ausnahme und nach einem Signal (`imGeteilten`; Befund 2 des
// Zweitblicks und Befund 1 des Nachlaufs über Phase I).
const { lauf, e } = await imGeteilten('g06', async (geteilt) => {
  const ordner = join(geteilt, 'ql')
  const app = join(ordner, 'MSEProbe.app')
  const appex = join(app, 'Contents/PlugIns/MSEVorschau.appex')
  const messer = join(ordner, 'Messer.app')
  for (const pfad of [join(app, 'Contents/MacOS'), join(appex, 'Contents/MacOS'), join(messer, 'Contents/MacOS')]) mkdirSync(pfad, { recursive: true })

  const ziel = ['-target', 'arm64-apple-macos15.2']
  execFileSync('swiftc', ['-O', '-parse-as-library', '-application-extension', ...ziel, '-module-name', 'MSEVorschau',
    join(quelle, 'Vorschaubild.swift'), '-framework', 'QuickLookThumbnailing',
    '-Xlinker', '-e', '-Xlinker', '_NSExtensionMain', '-o', join(appex, 'Contents/MacOS/MSEVorschau')], { stdio: 'inherit' })
  writeFileSync(join(geteilt, 'main.swift'), 'print("MSEProbe")\n')
  execFileSync('swiftc', ['-O', ...ziel, join(geteilt, 'main.swift'), '-o', join(app, 'Contents/MacOS/MSEProbe')], { stdio: 'inherit' })
  execFileSync('swiftc', ['-O', ...ziel, join(quelle, 'Messer.swift'), '-o', join(messer, 'Contents/MacOS/Messer')], { stdio: 'inherit' })

  const plist = (inhalt) => `<?xml version="1.0" encoding="UTF-8"?>
  <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
  <plist version="1.0"><dict>${inhalt}</dict></plist>
  `
  writeFileSync(join(appex, 'Contents/Info.plist'), plist(`
  <key>CFBundleIdentifier</key><string>de.mse4mac.probe.quicklook.vorschaubild</string>
  <key>CFBundleExecutable</key><string>MSEVorschau</string>
  <key>CFBundleName</key><string>MSEVorschau</string>
  <key>CFBundlePackageType</key><string>XPC!</string>
  <key>CFBundleVersion</key><string>1</string>
  <key>CFBundleShortVersionString</key><string>1.0</string>
  <key>LSMinimumSystemVersion</key><string>15.2</string>
  <key>NSExtension</key><dict>
    <key>NSExtensionPointIdentifier</key><string>com.apple.quicklook.thumbnail</string>
    <key>NSExtensionPrincipalClass</key><string>MSEVorschau.Vorschaubild</string>
    <key>NSExtensionAttributes</key><dict>
      <key>QLSupportedContentTypes</key><array><string>de.mse4mac.probe.mse-set</string></array>
      <key>QLThumbnailMinimumDimension</key><integer>0</integer>
    </dict>
  </dict>`))
  writeFileSync(join(app, 'Contents/Info.plist'), plist(`
  <key>CFBundleIdentifier</key><string>de.mse4mac.probe.quicklook</string>
  <key>CFBundleExecutable</key><string>MSEProbe</string>
  <key>CFBundleName</key><string>MSEProbe</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleVersion</key><string>1</string>
  <key>CFBundleShortVersionString</key><string>1.0</string>
  <key>LSMinimumSystemVersion</key><string>15.2</string>
  <key>CFBundleDocumentTypes</key><array><dict>
    <key>CFBundleTypeName</key><string>MSE-Set</string>
    <key>CFBundleTypeRole</key><string>Viewer</string>
    <key>LSItemContentTypes</key><array><string>de.mse4mac.probe.mse-set</string></array>
  </dict></array>
  <key>UTImportedTypeDeclarations</key><array><dict>
    <key>UTTypeIdentifier</key><string>de.mse4mac.probe.mse-set</string>
    <key>UTTypeDescription</key><string>MSE-Set (Probe)</string>
    <key>UTTypeConformsTo</key><array><string>public.zip-archive</string><string>public.data</string></array>
    <key>UTTypeTagSpecification</key><dict><key>public.filename-extension</key><array><string>mse-set</string></array></dict>
  </dict></array>`))
  writeFileSync(join(messer, 'Contents/Info.plist'), plist(`
  <key>CFBundleIdentifier</key><string>de.mse4mac.probe.messer</string>
  <key>CFBundleExecutable</key><string>Messer</string>
  <key>CFBundleName</key><string>Messer</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleVersion</key><string>1</string>
  <key>LSMinimumSystemVersion</key><string>15.2</string>
  <key>LSUIElement</key><true/>`))
  writeFileSync(join(geteilt, 'sandbox.plist'), plist('<key>com.apple.security.app-sandbox</key><true/>'))
  execFileSync('codesign', ['--force', '--sign', '-', '--entitlements', join(geteilt, 'sandbox.plist'), appex])
  execFileSync('codesign', ['--force', '--sign', '-', app])
  execFileSync('codesign', ['--force', '--sign', '-', messer])

  // **Ein Set als Zip, wie MSE2 es schreibt** — die Dateien liegen an der Wurzel.
  execFileSync('zip', ['-q', '-r', join(ordner, 'probe.mse-set'), '.'], { cwd: join(wurzel, 'saat/probe-werkstatt.mse-set') })
  execFileSync('cp', ['-c', join(quelle, 'quicklook.pl'), join(ordner, 'quicklook.pl')])
  // **Nur auf ausdrücklichen Wunsch bricht der Treiber ab** — für die Probe,
  // ob er danach aufgeräumt hat (`MSE_PROBE_ABBRECHEN=1`).
  if (process.env.MSE_PROBE_ABBRECHEN === '1') writeFileSync(join(geteilt, 'abbrechen'), '')
  const lauf = await fahre({ vm, geteilt, name: 'ql', treiber: 'quicklook.pl' })
  return { lauf, e: existsSync(join(geteilt, 'g06-ergebnis.json')) ? JSON.parse(readFileSync(join(geteilt, 'g06-ergebnis.json'), 'utf8')) : null }
})
const { rueckgabe, ausgabe } = lauf
if (!e) {
  console.error(`Kein Ergebnis (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}

const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const bild = (b) => (b?.art ? `${b.art}, ${String(b.breite)} × ${String(b.hoehe)}, Mitte ${b.farbe}` : `kein Bild — ${String(b?.fehler ?? '?')}`)
const lief = e.nachher?.art === 'vorschaubild' && e.nachher.breite < e.nachher.hoehe
const text = [
  '# Beleg: G06 — Quick Look für Setdateien, die Machbarkeit',
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`${process.env.MSE_PROBE_ABBRECHEN === '1' ? 'MSE_PROBE_ABBRECHEN=1 ' : ''}node scripts/quicklookgast.mjs ${vm}\` · **Rückgabe:** ${String(rueckgabe)}`,
  '',
  `**Maschine:** Lume-VM \`${vm}\` — macOS ${e.maschine.system} (${e.maschine.bau}). Das Probebündel lag nur dort, wurde wieder abgemeldet und gelöscht (aufgeräumt: ${e.aufgeraeumt}).`,
  '',
  `**Ergebnis:** die ad hoc signierte Erweiterung ${lief ? '**lief**' : '**lief nicht**'}; ${e.nachher?.farbe === 'magenta' ? 'aus der Sandbox **waren** die Vorlagenpakete lesbar' : e.nachher?.farbe === 'cyan' ? 'aus der Sandbox waren die Vorlagenpakete **nicht** lesbar' : 'die Farbe sagt nichts über die Pakete'}.`,
  '',
  '| | Antwort von `QLThumbnailGenerator` für `probe.mse-set` (256 × 256 verlangt) |',
  '| --- | --- |',
  `| vorher, ohne Erweiterung (Gegenprobe) | ${bild(e.vorher)} |`,
  `| nachher, mit Erweiterung | ${bild(e.nachher)} |`,
  '',
  `- **Registrierung** (\`pluginkit -m -v\`): \`${String(e.registriert ?? '—').replace(/\s+/g, ' ')}\``,
  `- **Signatur der Erweiterung**: \`${String(e.signatur ?? '—').slice(0, 300)}\``,
  `- **Berechtigungen**: \`${String(e.berechtigungen ?? '—').replace(/\s+/g, ' ')}\``,
  `- **Größe der Erweiterung**: ${String(e.groesse ?? '—')} (ohne Renderer — er zeichnet eine Farbfläche).`,
  ...(e.abbruch ? [`- **Abbruch**: ${e.abbruch} — **der Lauf ist unvollständig**`] : []),
  '',
  '## Das Protokoll',
  '',
  '```text',
  ...e.protokoll,
  String(e.protokollDesSystems ?? ''),
  '```',
  '',
]
writeFileSync(join(wurzel, 'docs/belege/g06-quicklook.md'), text.join('\n'))
console.log(text.slice(0, 20).join('\n'))

// **Vollständig ist der Lauf, wenn beide Antworten kamen und aufgeräumt ist**
// — eine sauber gemessene negative Machbarkeit ist ein Ergebnis, ein Abbruch
// nicht (Befund 7 des Zweitblicks über Phase I).
const unvollstaendig = [
  e.abbruch && `abgebrochen: ${e.abbruch}`,
  !e.vorher?.art && 'keine Antwort vorher',
  !e.nachher?.art && 'keine Antwort nachher',
  e.aufgeraeumt !== 'ja' && 'nicht aufgeräumt',
].filter(Boolean)
if (unvollstaendig.length) console.error(`unvollständig: ${unvollstaendig.join(', ')}`)
process.exit(unvollstaendig.length ? 1 : 0)
