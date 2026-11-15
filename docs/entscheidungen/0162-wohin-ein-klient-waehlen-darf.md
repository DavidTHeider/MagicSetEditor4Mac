# 0162 — Wohin ein Klient wählen darf, und wer das Ziel setzt

**Stand:** angenommen am 2026-10-02 · SEC-P4
(#2000),
Befund P4 der Sicherheitsprüfung QA-09 (`docs/PENTEST.md`). Marke `new`,
und gesucht wurde: MSE2 holt seine Installerliste und die Installer über
`wxURL` (`src/gui/packages_window.cpp`, Fassung f9d9356 vom 2021-07-06 — die
Liste von `settings.installer_list_url`, jeder Installer von der Adresse, die
die Liste nennt) und den Versionsstand ebenso (`src/gui/update_checker.cpp`).
**Eine Prüfung des Ziels gibt es dort nicht**, und auch keine KI und keine
Gegenüber; es gibt nichts zu portieren.

## Die Frage

Drei Wege wählten eine Adresse, die ein Rufer mit einem Token `schreiben`
bestimmen konnte, **ohne Prüfung auf die Schleife, das eigene Netz oder
`169.254.0.0/16` und ohne eigene Weiterleitungspolitik**: der Paketbezug
(`POST /pakete`), die Adressen der KI und der Proxy (über die Einstellungen,
angewählt danach von einem **lesenden** Weg wie `GET /ki/modelle`) und die
Annahme einer Einladung, die sich bei der Adresse meldet, die ihr Aussteller
selbst hineinschreibt. Zu entscheiden war, **wo** die Sperre sitzt und was
sie trifft, ohne die beschriebenen Wege zu nehmen: die lokale KI auf der
Schleife, ein Ollama auf einem zweiten Rechner im eigenen Netz
(`docs/NETZEIGENSCHAFTEN.md`), ein Firmenproxy und die Läufe gegen
`httptest` auf `127.0.0.1`.

## Die Entscheidung

1. **Eine Grundsperre für jeden Klienten** aus `internal/netz`: `0.0.0.0/8`,
   `::`, `169.254.0.0/16`, `fe80::/10`, Gruppenadressen und `240.0.0.0/4`.
   Kein Dienst braucht sie, und `0.0.0.0` landet auf macOS bei diesem
   Rechner (gemessen am 2026-10-02). Sie gilt auch der Leitung zum Proxy.
2. **Nur öffentlich für den Paketbezug** — dazu gesperrt die Schleife und das
   eigene Netz. Seine Adresse kommt **mit der Anfrage**; wer sie nennt, ist
   der Rufer, und gegen ihn hilft nur eine Sperre am Ziel.
3. **Geprüft an drei Stellen**: vor der Anfrage (Schema und Ziffern — die
   einzige Prüfung hinter einem Proxy), nach der Auflösung
   (`net.Dialer.Control`, bevor ein Paket hinausgeht) und bei jeder
   Weiterleitung (`CheckRedirect`, höchstens zehn). Ein Wächter hält, dass
   jeder Wähler und jeder Klient in `internal/netz` beides trägt
   (`internal/schichten/netzziel_test.go`).
4. **Wohin gesendet wird, setzt nur, wer senden darf.** `ki.adresse`,
   `dienst.<ki>.adresse` und `netz.proxy.*` fragen im Einstellungsverwalter
   und in der Einfuhr nach `externSenden` — und ein Token reicht nie dorthin
   ([0078](0078-der-wirksame-akteur.md)). **Die Sperre sitzt hier am
   Schreiben und nicht am Ziel**, weil das Ziel im eigenen Netz liegen
   darf; wer es setzt, ist ein Mensch.
5. **Eine einlösbare Einladung nimmt nur an, wer senden darf** — geprüft im
   Einladungswerk **vor** der Annahme, ganz oder gar nicht, und noch einmal
   vor der Meldung. Eine Einladung ohne Ort fragt nicht; sie wählt niemanden
   an. **Keine Bereichssperre auf ihre Adresse**: das eigene Netz ist dort
   der Zweck, und die Verbindung ist auf die Kennung des Einladenden
   festgenagelt.
6. **Der benannte Schalter ist `MSE_NETZ_SCHLEIFE=an`** — er öffnet für den
   Paketbezug die Schleife, und nur sie. **Eine Umgebungsangabe, keine
   Einstellung**: eine Einstellung setzt, wer schreiben darf — also genau der
   Rufer, gegen den die Sperre steht. Die Umgebung setzt, wer den Prozess
   startet. Gebraucht wird er von Läufen; im Betrieb setzt ihn niemand.

## Einordnung (§16)

**Brechend für Token**, und mit Absicht: ein Token setzt `ki.adresse`, die
Adresse eines KI-Anbieters und `netz.proxy.*` nicht mehr (`403`), und es
nimmt keine einlösbare Einladung mehr an. **Brechend für den Paketbezug aus
dem eigenen Netz**: ein Git-Server auf `10.…` oder `192.168.…` ist keine
Paketquelle mehr. **Verträglich** für jeden anderen Weg: Fenster,
Kommandozeile und Prozessmarke setzen und nehmen an wie bisher, die KI auf
der Schleife und im eigenen Netz bleibt erreichbar, ein Firmenproxy ebenso.

## Verworfen

- **Das eigene Netz auch für die KI sperren**, wie der Vorschlag im Issue es
  las. Es nähme das Ollama auf einem zweiten Rechner, das
  `NETZEIGENSCHAFTEN.md` als Weg beschreibt — und der Angriff dort ging
  ohnehin über das **Setzen** der Adresse, das jetzt ein Token nicht mehr
  darf.
- **Die Zieladresse des Proxys wie ein Ziel prüfen.** Ein Firmenproxy steht
  im eigenen Netz; wer ihn sperrte, sperrte jeden Bezug dahinter. Wer den
  Proxy setzt, regelt Punkt 4.
- **Nur vor der Anfrage prüfen.** Ein Name, der auf `127.0.0.1` zeigt, käme
  durch, und zwischen Prüfung und Verbindung dürfte er sich ändern. Erst
  `Control` sieht die Adresse, mit der wirklich verbunden wird.
- **Ein Schalter für das eigene Netz beim Paketbezug.** Niemand hat einen
  Git-Server im eigenen Netz als Paketquelle verlangt; ein Schalter, den
  keiner braucht, ist eine Lücke mit Namen. Kommt der Fall, gehört er hierher.
- **Eine Bereichssperre für Gegenüber.** Zwei Rechner im selben Netz sind
  der Zweck der Peerleitung; die Sperre wäre die Abschaltung der Funktion.
