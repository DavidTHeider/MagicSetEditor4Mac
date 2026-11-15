# 0194 — Der Freigabeschlüssel wechselt über benannte Anker, und die Fassung ist Teil der Signatur

**Stand:** angenommen am 2026-10-16 · #2356 (SEC03, Befunde A-17 und A-18). **Der Schlüssel selbst ist nicht gewechselt** — ihn erzeugt und verwahrt der Auftraggeber; kein Agent erzeugt oder veröffentlicht ein echtes Geheimnis.

## Lage

Am Stand `d5712b648`, nachgelesen und gemessen:

1. **Signiert wird mit dem Commitschlüssel** `BF81923EA5D38635` (AGENTS.md §5), und eingebettet war **genau einer** (`eingebetteterBund` wies zwei ab, #892). Ein Wechsel — Verlust, Kompromittierung oder die Trennung von Commit- und Freigabeschlüssel — hätte jede installierte Fassung von der nächsten abgeschnitten: sie kennt nur den alten, die neue ist mit dem neuen signiert. Übrig bliebe ein Abbild von Hand.
2. **Ein Widerruf erreichte nichts.** Der eingebettete Bund trug keinen, und eine Fassung, die mit einem widerrufenen Schlüssel signiert ist, hätte jede installierte angenommen.
3. **Die Fassung war nicht Teil dessen, was signiert ist.** Sie kam aus dem Tag. Unter `v9.9.9` lud die Anwendung die echten Prüfsummen, die echte Signatur und das echte Abbild von 0.7.0 — **rot gesehen** in `rueckfall_test.go`; das Manifest einer anderen Fassung nahm sie ebenso an.
4. **Echte Schlüssel lagen in der automatischen Probe.** Zwei Testdateien signierten mit `gpg` und dem Freigabeschlüssel aus dem Schlüsselbund dieses Rechners — in jedem `make ci` auf dem Rechner, der auch freigibt.

## Entscheidung

1. **Ein eigener Freigabeschlüssel**, getrennt vom Commitschlüssel. Ihn erzeugt der Auftraggeber, offline, und hinterlegt den öffentlichen Teil samt **vorab erzeugtem Widerrufszertifikat** an einem Ort außerhalb dieses Rechners. Bis dahin signiert der heutige weiter.
2. **Der Bund führt genau die benannten Anker** (`freigabeanker` in `internal/aktualisierung/signatur.go`): einen heute, **zwei im Übergang**. Ein Schlüssel, der kein benannter Anker ist, macht den ganzen Bund ungültig — die Zusage aus #892, dass niemand still einen zweiten dazulegt, bleibt; sie steht jetzt als Liste und nicht als Zahl.
3. **Der Übergang in zwei Freigaben.**
   - **Freigabe N** trägt beide Anker im Bund, ist mit dem **alten** signiert. Jede installierte Fassung nimmt sie an.
   - **Freigabe N+1** ist mit dem **neuen** signiert und trägt nur ihn — und den alten, wenn er widerrufen wird, mit seinem Widerruf. N nimmt sie an, weil N den neuen kennt.
   - Wer N überspringt, bleibt bei N−1 und nimmt N+1 **nicht** an: er braucht N oder ein Abbild von Hand. Das steht in der Releasenotiz von N+1.
4. **Ein Widerruf reist im Bund.** Ein widerrufener Schlüssel darf darin stehen; go-crypto v1.3.0 weist eine Signatur von ihm ab (`ErrKeyRevoked`, `openpgp/read.go`, `checkMessageSignatureDetails`, nachgelesen am 2026-10-16) — **jede** Signatur, auch eine ältere, denn geprüft wird gegen die heutige Zeit und nicht gegen die der Signatur. Ein Widerruf mit dem Grund `KeyCompromised` gilt unbedingt; jeder andere, solange er selbst nicht abgelaufen ist (`keys.go`, `revoked`, `SigExpired`; präzisiert nach dem Zweitblick über Phase 5). Ein benannter Anker, der widerrufen ist, macht den Bund ungültig.
5. **Die Fassung ist Teil der Signatur.** Das Abbild heißt, wie `release.sh` es für die behauptete Fassung ablegt (`MagicSetEditor4Mac-<fassung>.dmg`), und dieser Name steht in der signierten Liste; das Manifest nennt dieselbe Fassung wie der Tag. Eine alte Liste nennt den neuen Namen nicht.
6. **Echte Schlüssel bleiben außerhalb der automatischen Probe.** Im Prozess signiert ein Probeschlüssel; die Programme der Flächenläufe tragen das Tag `probeanker` und lesen ihren Anker aus `MSE_PROBEANKER`. Kein Freigabeweg baut mit dem Tag — ein Wächter hält es, und einer hält, dass kein Test `gpg` zum Signieren ruft. Dass der **eingebettete** Schlüssel der richtige ist, prüft die veröffentlichte Signatur einer echten Freigabe, ohne privaten Schlüssel.

## Was das nicht leistet

- **Eine installierte Fassung kennt nur ihren Bund.** Wer eine Fassung von **vor** dem Widerruf hat, nimmt eine Signatur des widerrufenen Schlüssels weiter an. Ein Widerruf wirkt ab der Fassung, die ihn trägt — deshalb gehört ein kompromittierter Schlüssel sofort in eine Freigabe mit Widerruf und in die Notiz an die Nutzenden.
- **Ein Angreifer mit dem gültigen Schlüssel** signiert, was er will; die Anker schützen gegen einen fremden und einen widerrufenen, nicht gegen den richtigen in falscher Hand.
- **Die Gatekeeper-Frage** ist eine andere: diese Signatur entscheidet, was die Anwendung lädt, nicht ob macOS das Bündel startet ([0003](0003-keine-notarisierung.md)).


## Nachtrag vom 2026-11-15: gewechselt zu 1.0.0, ohne die zwei Freigaben

**Der eigene Freigabeschlüssel aus Punkt 1 ist da** — `44C59428DE1F766E`,
getrennt vom Commitschlüssel `1503F87D4C128638`, beide nur mit der
Noreply-Adresse von GitHub. **Erzeugt hat ihn der Agent**, auf ausdrückliche
Bitte des Auftraggebers und ohne Passphrase wie den bisherigen — anders, als
die Lage oben es vorsah. Die Widerrufszertifikate legte `gpg` unter
`~/.gnupg/openpgp-revocs.d/` ab; außerhalb dieses Rechners verwahrt sie der
Auftraggeber.

**Den Übergang aus Punkt 3 gab es nicht.** 1.0.0 führt nur den neuen Anker und
ist mit ihm signiert. Einen öffentlich installierten Bestand, den der alte
Anker hätte mitnehmen müssen, gab es nicht; wer eine Entwicklungsfassung
installiert hat, nimmt 1.0.0 als Abbild von Hand
([0205](0205-1-0-0-ist-ein-neuanfang.md)). **Ab hier gilt Punkt 3 wieder**:
der nächste Wechsel geht über zwei Freigaben.
