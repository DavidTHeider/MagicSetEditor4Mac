# Was die Fassung 0.5.0 geschrieben hat

**Geschrieben von 0.5.0 selbst und nicht nachgebaut** — ENG-06.AK-03
(#1731).
Ein Nachbau mit dem heutigen Code prüfte nur, ob diese Fassung sich selbst
lesen kann. Gelesen werden die Dateien von
[`internal/app/vorfassung_test.go`](../../internal/app/vorfassung_test.go).

| Datei | Was | SHA-256 |
| --- | --- | --- |
| `sicherung.zip` | eine Sicherung des ganzen Bestands, Schema `0077_handelsfremdbelege.sql` — zwei Sets, vier Karten, ein Deck, Einstellungen | `d1dbee5103ea0933da9c43ff889a7d32d829c11171008025852738282262f34b` |
| `einstellungen.json` | die Einstellungen als Datei, Abschnitte `nutzer`, `profil` und `arbeitsplatz`, Austauschfassung `1` | `521b27c8eb2ac0680b5eaa742cc94ae328a72ae07a1189f12ef0a68ef1675a63` |
| `probeset.mseumschlag` | ein unterschriebener Umschlag mit dem Set »Mein Probeset« — Urheber `bafzaajaiaejcbbdwuyds5lqlzi3gg3yx5ooj3nowx3kci5asu5ani6sevore3zsd` | `d20e9ebd0f17c21c69646ce5742f2c44b1026dfb883898df0d921403f1301ad7` |

## Wie sie entstanden sind

Am 2026-09-22, mit der Kommandozeile aus dem Tag `v0.5.0` (Commit `f498fb4a`),
gebaut mit Go 1.27.1 — **in einem eigenen Bestand, einem leeren Zuhause und
einem Wegwerfschlüsselbund**, der danach gelöscht wurde:

```bash
git worktree add --detach ../v050 v0.5.0
go build -o mse050 ./cmd/mse                       # im Arbeitsbaum ../v050

export HOME=$ORT/home MSE_BESTAND=$ORT/bestand.db
mse050 paket-installieren saat/probe.mse-game
mse050 paket-installieren saat/probe-schlicht.mse-style
mse050 set-neu "Altes Probeset" probe probe-schlicht
mse050 karte 1 anlegen
mse050 oeffnen testdaten/probe.mse-set             # »Mein Probeset«, drei Karten
mse050 einstellungen setzen sprache en
mse050 einstellungen setzen erscheinungsbild dunkel
mse050 einstellungen setzen arbeitsplatz.navigatorbreite 280
mse050 einstellungen setzen arbeitsplatz.inspektor verborgen
mse050 deck neu "Altes Deck" probe "aus 0.5.0"
mse050 einstellungen ausfuhr nutzer profil arbeitsplatz > einstellungen.json
mse050 bestand-sichern sicherung.zip

security create-keychain -p leer $ORT/bund.keychain-db
MSE_SCHLUESSELBUND=$ORT/bund.keychain-db mse050 peer-anlegen
MSE_SCHLUESSELBUND=$ORT/bund.keychain-db mse050 umschlag-schnueren 2 probeset.mseumschlag
security delete-keychain $ORT/bund.keychain-db
```

**Der Schlüssel des Urhebers ist mit dem Bund gelöscht.** Das ist Absicht:
geprüft wird, dass eine Unterschrift von damals heute trägt — nicht, dass
jemand sie heute noch einmal leisten kann.

## Was hier nicht liegt

- **Eine Partie als Datei.** Das Format gibt es erst seit 0.6.0
  (`domain.PartieaustauschFassung`); 0.5.0 konnte keine schreiben.
- **Eine Sammlungsliste mit Druckverweis.** Sie bräuchte Drucke aus Scryfall
  und damit das Netz; die Liste **ohne** Verweis, wie ältere Fassungen sie
  schrieben, liest `TestEinAltesCSVOhneReferenzLiestSich`.
- **Ein Mitschnitt des Handschlags.** Er ist an seine Leitung und an zwei
  frisch gewürfelte Zahlen gebunden und lässt sich absichtlich nicht wieder
  vorlegen — `docs/FASSUNGEN.md` nennt, was stattdessen gilt.
