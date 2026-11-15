# Was die Fassung 0.7.0 geschrieben hat

**Geschrieben von 0.7.0 selbst und nicht nachgebaut** — J02
(#2266,
Vermerk [0189](../../docs/entscheidungen/0189-nur-die-grundlinie.md)). Seit es
nur die Grundlinie gibt, ist 0.7.0 die Brücke für jeden älteren Bestand; diese
Sicherung belegt, dass was dort steht, hier ohne einen Schritt aufgeht.
Gelesen wird sie von
[`internal/app/grundlinienaufstieg_test.go`](../../internal/app/grundlinienaufstieg_test.go).

| Datei | Was | SHA-256 |
| --- | --- | --- |
| `sicherung.zip` | eine Sicherung des ganzen Bestands, Schema `0097_faktorhaushalt.sql` — zwei Sets, vier Karten, ein Deck, Sprache `en` | `5fe76a90791dcef7a1786c38fd30c3a98e9ece03bad4d760348b6b209b365a5b` |

## Wie sie entstanden ist

Am 2026-10-12, mit der Kommandozeile aus dem Tag `v0.7.0` (Commit
`ad53f5e3c`), gebaut mit Go 1.27.2 — **in einem eigenen Bestand und einem
leeren Zuhause**:

```bash
git worktree add --detach ../v070 v0.7.0
go build -o mse070 ./cmd/mse                       # im Arbeitsbaum ../v070

export HOME=$ORT/home CFFIXED_USER_HOME=$ORT/home MSE_BESTAND=$ORT/bestand.db
mse070 paket-installieren saat/probe.mse-game
mse070 paket-installieren saat/probe-schlicht.mse-style
mse070 set-neu "Altes Probeset" probe probe-schlicht
mse070 karte 1 anlegen
mse070 oeffnen testdaten/probe.mse-set             # »Mein Probeset«, drei Karten
mse070 einstellungen setzen sprache en
mse070 einstellungen setzen erscheinungsbild dunkel
mse070 deck neu "Altes Deck" probe "aus 0.7.0"
mse070 bestand-sichern sicherung.zip
```

## Was hier nicht liegt

**Einstellungsdatei und Umschlag.** Sie gehören nicht zum Schema und hängen
nicht an der Grundlinie; was eine frühere Fassung davon schrieb, belegt
[`testdaten/fassung-0.5.0`](../fassung-0.5.0/LIESMICH.md).
