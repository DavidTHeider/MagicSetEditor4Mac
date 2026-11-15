# 0157 — Ein markierter Block, der byte-gleich wieder geht

**Stand:** angenommen am 2026-10-01 · OPS-09, Schnitt 1
(#1907);
Schnitt 2 in #1961

## Die Frage

OPS-03 legte `mse` und `mse-mcp` nach `~/.local/bin` und fasste das
Shellprofil ausdrücklich **nicht** an: wer das Verzeichnis nicht im `PATH`
hatte, bekam eine Zeile zum Kopieren. OPS-09 verlangt je Schnittstelle einen
Knopf zum Einrichten und einen zum Widerrufen — in Dateien, die Menschen und
andere Programme ebenfalls schreiben (`~/.zprofile`,
`~/.codex/config.toml`) —, eine Vorschau mit jeder Zeile (AK-01) und ein
Widerrufen, das die Datei **byte-gleich** hinterlässt (AK-02).

## Was MSE2 tut

Nichts — MSE2 richtet außerhalb seines Programms nichts ein (gesucht nach
`zprofile`, `PATH`, `LaunchAgents` in `/Users/david/mse2-quelle/src`,
Fassung f9d9356 vom 2021-07-06: keine Fundstelle). Marke **new**.

## Was gilt

1. **Ein markierter Block** (`domain.BlockEinsetzen`): zwei Kommentarzeilen
   umschließen, was dieses Programm geschrieben hat — so hält es auch
   `conda init`. Was außerhalb steht, gehört dem Menschen.
2. **Byte-gleich**: der Block wird ans Ende gehängt; fehlte der Datei das
   letzte Zeilenende, wird eines ergänzt, und die Anfangsmarke sagt es —
   widerrufen nimmt es mit. Wird eine Datei dabei leer, geht sie: sie entstand
   mit dem Einrichten.
3. **Idempotent**: zweimal einrichten ändert beim zweiten Mal nichts; ein
   eigener Block mit anderem Inhalt ist **veraltet** (die Anwendung liegt
   woanders) und wird an seiner Stelle ersetzt.
4. **Fremd heißt: nicht anfassen** — derselbe Eintrag außerhalb des Blocks
   (`[mcp_servers.magicseteditor4mac]` von Hand), ein halber Block, eine
   Datei, die sich nicht lesen lässt.
5. **Ein Symlink bleibt ein Symlink**: geschrieben wird dorthin, wohin er
   zeigt — wer sein Profil aus einem Dotfiles-Verzeichnis holt, behält es.
6. **Geschrieben wird nur am Fenster**; lesend steht der Stand an allen vier
   Flächen (`Schnittstellen`, `lesen.schnittstellen`, `mse schnittstellen`,
   `GET /schnittstellen`) — dieselbe Entscheidung wie bei OPS-03.
7. **Die Formate sind nachgelesen**: Codex an `codex mcp add --help`
   (CLI 0.159.2, 2026-10-01: `[mcp_servers.<name>]` mit `command`); das
   Anmeldeprofil von zsh ist `~/.zprofile`.

## Was es nicht ist

**Kein JSON**: Claude Desktop und Claude Code halten ihre Konfiguration in
JSON, das keinen Kommentar kennt — und Claude Code schreibt seine Datei
laufend selbst um. Dafür, und für den Dienst als LaunchAgent, gibt es
Schnitt 2 (#1961) mit je eigener Antwort auf „byte-gleich".

## Verworfen

- **Die Datei vorher sichern und beim Widerrufen zurücklegen**: das nähme
  jede Änderung mit, die ein Mensch seither gemacht hat.
- **Eine Zeile ohne Marken**: dieselbe Zeile kann ein Mensch geschrieben
  haben; widerrufen nähme dann seine.

## Gemessen

- Domäne: hinein und spurlos hinaus für leere Dateien, mit und ohne
  Zeilenende, mit fremden Blöcken; veraltet an seiner Stelle ersetzt; ein
  halber Block bleibt.
- Dienst an echten Dateien in einem eigenen Verzeichnis: umkehrbar,
  idempotent, fremd bleibt, veraltet repariert, nicht da und unnötig
  schreiben nichts, ein Symlink bleibt.
- Anwendung mit eigenem Zuhause: einrichten und widerrufen beider Ziele, die
  Vorschau nennt jede Zeile.
