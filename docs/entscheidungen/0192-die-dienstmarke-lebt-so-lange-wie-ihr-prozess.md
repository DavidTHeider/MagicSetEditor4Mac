# 0192 — Die Dienstmarke lebt so lange wie ihr Prozess

**Stand:** angenommen am 2026-10-16 · #2355 (SEC02, Befund A-16).

## Lage

`mse-rest` würfelt bei jedem Start eine Marke aus `crypto/rand`
(`internal/rest/marke.go`) und sagt sie auf der Standardausgabe an. Unter dem
LaunchAgent zeigten `StandardOutPath` **und** `StandardErrorPath` auf
`schnittstellen/dienst.txt`, und launchd **hängt an**: jede Ansage seit der
Einrichtung stand darin, samt aller Marken und aller Fehlerausgaben, ohne
Grenze und mit den Rechten, die launchd beim Anlegen vergibt.

**Gemessen am 2026-10-16** am Stand vor SEC02, mit dem gebauten `mse-rest` und
der Ausgabe so angehängt, wie launchd sie anhängt: nach zehn Starts **zehn**
Marken in der Datei, alle verschieden, die Datei **0644**
(`docs/audits/0.9.0/umsetzung/SEC02-2355.md`).

## Entscheidung

**Rotation je Start, kein zeitlicher Ablauf.**

1. **Eine Marke gilt genau so lange wie der Prozess, der sie gewürfelt hat.**
   Jeder Start würfelt neu; ein Neustart durch launchd, `kickstart -k`, das
   Ende über den eigenen Schalter oder eine Abmeldung machen die alte
   ungültig.
2. **Die Ansage schreibt das Programm selbst** — `mse-rest -ansage <datei>`
   ersetzt sie je Start atomar (Zwischendatei, `Sync`, `rename`) mit
   **0600** in einem **0700**-Ordner. Es steht immer nur die gültige Marke
   darin. **Auch ein bestehender Ordner wird privat**: einem Verweis an
   seiner Stelle wird nicht gefolgt, und was Gruppe und andere darin
   dürften, fällt (`service.PrivaterOrdner`, Zweitblick über Phase 5,
   Befund 3). **Und der Weg dorthin wird vorher geprüft** (Nachlauf, Befund
   3): jeder Vorfahr des aufgelösten Pfads gehört diesem Nutzer oder `root`,
   und niemand sonst darf darin schreiben, außer mit Sticky-Bit — sonst wird
   nichts angelegt. Ein Tausch durch denselben Nutzer bleibt möglich; er hat
   den Zugang ohnehin. Die Standardausgabe geht unter launchd ins Leere
   (`/dev/null`).
3. **Die Fehlerausgabe geht in eine eigene Datei**, `dienst.log`, und hat
   eine Grenze: über 1 MiB beginnt sie beim nächsten Start neu, mit einer
   Zeile, die das sagt. Eine Marke steht darin nicht — der Dienst schreibt
   sie nur in die Ansage.
4. **Ein Widerruf nimmt die Ansage mit** — erst nachdem der Dienst entladen
   ist. Lässt er sich nicht entladen, bleibt sie: dann läuft er, und seine
   Marke gilt.
5. **Das Diagnosepaket liest weder Ansage noch `dienst.log`**, und die
   Schwärzung erkennt eine Marke auch dann, wenn sie aus Ziffern allein oder
   aus `a` bis `f` allein besteht — bis dahin verlangte sie Ziffer **und**
   Buchstabe.

## Warum kein Ablauf

Ein Ablauf nach Zeit verlangte, dass ein Klient eine neue Marke holt, ohne
dass der Dienst neu startet — also einen zweiten Weg zur Marke, der selbst
geschützt sein muss, und eine Klientenmigration. Das wäre nach §16
**erweiternd oder brechend** und ist nicht Teil dieses Pakets. Die Rotation je
Start ist **verträglich**: ein Skript, das die Marke heute aus der Ansage
liest (`read -r _ marke`), liest sie nach jedem Start wieder dort.

**Was das kostet**: ein Dienst, der wochenlang ohne Neustart läuft, behält
wochenlang dieselbe Marke. Wer sie erneuern will, startet ihn neu — mit
`launchctl kickstart -k gui/<UID>/io.github.davidtheider.magicseteditor4mac.dienst`
oder auf der Seite *Schnittstellen* mit Widerrufen und Einrichten; einen
eigenen Knopf zum Neustart hat das Fenster nicht. Wer die Datei lesen kann,
hat den Zugang; dagegen stehen 0600 und 0700 und sonst nichts.

## Was bleibt

- **Eine schon eingerichtete plist** zeigt weiter auf die alte Datei, bis sie
  neu eingerichtet wird. Die Seite *Schnittstellen* nennt sie **veraltet**
  (der Inhalt weicht ab), und „Einrichten" ersetzt sie; die alte
  `dienst.txt` wird beim ersten Start ersetzt, die alten Marken darin gehen
  mit.
- **Der Kennungsumzug** ([0191](0191-die-kennung-ist-io-github-davidtheider.md))
  übernimmt die plist mit neuem Etikett und sonst unverändert — auch sie gilt
  danach als veraltet.
- **Dass launchd selbst die Ausgabe so leitet**, belegt der Gastlauf
  `scripts/sec02gast.mjs` mit echten Neustarts; ein Einheitstest sieht
  launchd nicht.

## Verworfen

- **Die Marke im Schlüsselbund.** Ein Skript läse sie dann nur über
  `security`, und der Dienst bräuchte unter launchd Zugriff auf den
  Anmeldebund — ein zweiter Ort mit eigenem Lebenszyklus für einen Wert, der
  nur so lange gilt wie ein Prozess.
- **Die Ausgabe weiter von launchd schreiben lassen, aber kürzen.** launchd
  öffnet die Datei selbst und hängt an; ein Kürzen zwischen zwei Starts ließe
  die Marken des letzten Laufs trotzdem stehen.
