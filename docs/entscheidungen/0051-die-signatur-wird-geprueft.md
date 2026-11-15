# 0051 — Die Signatur über den Prüfsummen wird geprüft

**Datum:** 2026-08-26 · **Status:** angenommen ·
**Anlass:** das Gesamtreview vor 0.2.0, Befund 2

---

## Die Lücke

Stufe 2 der Aktualisierung
(#1100) holte
`pruefsummen.txt`, nahm daraus die Zeile des Abbilds und verglich sie mit dem
Geladenen.

**Das prüft, ob die beiden Dateien zueinander passen — nicht, woher sie
kommen.** Eine Gegenstelle, die beide Antworten beherrscht, liefert ein eigenes
Abbild samt eigener Prüfsumme, und der Vergleich geht auf.

**Und die Abhilfe lag daneben.** `scripts/release.sh` schreibt seit dem
2026-08-25 `pruefsummen.txt.asc` und prüft sie im selben Lauf
(#892).
`docs/RELEASE.md` und `README.md` nennen sie ausdrücklich als das, was den
gemeinsamen Austausch verhindert. Die Anwendung sah sie nie an.

Das ist die Gestalt, vor der §10 warnt: **ein Schloss, das dagelegen hat.**

## Die Entscheidung

**Der Ladeweg holt `pruefsummen.txt.asc` und prüft sie, bevor er die Liste
deutet.** Eine Freigabe ohne Signatur wird abgewiesen — vor dem Laden, wie eine
ohne Prüfsummen.

**Der öffentliche Schlüssel liegt im Programm** (`go:embed`), nicht in einer
Einstellung und nicht im Schlüsselbund: er ist eine Zusage über den
Herausgeber. Ein Schlüssel, den die Anwendung beim ersten Lauf von irgendwo
holte, könnte derselbe setzen, gegen den er schützen soll.

Es ist derselbe Schlüssel, der unter jedem Commit steht (AGENTS.md §5):
`4D31AE0276CAA61C5D35581EBF81923EA5D38635`. Ein Wächter hält den eingebetteten
gegen diese Zahl — ein Programm, das gegen einen Schlüssel prüft, den niemand
kennt, sieht aus wie Sicherheit und ist keine.

## Die Reihenfolge

1. **Prüfsummen holen** — ein halbes Kilobyte.
2. **Signatur holen und prüfen.** Passt sie nicht, endet es hier.
3. **Die Zeile des Abbilds suchen.** Erst jetzt wird die Liste gedeutet.
4. **Abbild laden**, in eine Nachbardatei, mit laufender Prüfsumme.
5. **Vergleichen.** Erst die geprüfte Datei bekommt ihren Namen.

**Schritt 2 steht vor Schritt 3**, und das ist kein Zufall: wer die Liste erst
deutete und danach prüfte, hätte einen Weg, auf dem eine ungeprüfte Zeile schon
etwas entschieden hat.

## Die Bibliothek

`github.com/ProtonMail/go-crypto/openpgp` (BSD-3-Clause). **Sie lag schon im
Baum** — Wails bringt sie mit; sie ist am 2026-08-26 nur in `go.mod` sichtbar
geworden, weil dieses Programm sie jetzt selbst benutzt. Der Bau bekommt damit
keine neue Abhängigkeit, `DRITTANBIETER.md` einen neuen Eintrag.

**`golang.org/x/crypto/openpgp` wäre die andere Wahl gewesen** und ist seit
2021 eingefroren; ProtonMails Gabel ist ihre gepflegte Fortsetzung und
dieselbe Herkunft.

## Was das kostet

**Eine dritte Anfrage je Ladevorgang.** Sie wiegt 230 Byte.

**Und ein Freigabeweg, der die Signatur wirklich schreiben muss.** Tat er
vorher auch; ab jetzt merkt es jemand, wenn nicht — eine Freigabe ohne `.asc`
lässt sich nicht mehr laden. Das ist beabsichtigt.

**Für Tests ist der Bund einsetzbar** (`MitFreigabebund`). Ohne diese Naht
bräuchte jeder Ladeversuch den **privaten** Schlüssel des Herausgebers, und
dann prüfte er, ob `gpg` da ist. **Was sie nicht tut, ist die Prüfung
abschalten**: ein Test setzt einen echten Schlüssel ein und signiert damit
wirklich (§3). Die Läufe über die Flächen benutzen den echten Schlüssel — sie
starten eigene Programme, und die tragen den eingebetteten.

## Was sie weiterhin nicht leistet

**Sie macht die Anwendung nicht notariell beglaubigt**
([0003](0003-keine-notarisierung.md)). Gatekeeper sieht diesen Schlüssel nicht
an; der Erststart-Weg bleibt, wie er ist. Was sie leistet, ist die Zusage über
die **Herkunft der geladenen Datei** — und die hing bis heute an der Annahme,
dass niemand beide Antworten beherrscht.

**Und sie ersetzt die Selbstinstallation nicht**
([0050](0050-die-anwendung-ersetzt-sich-nicht-selbst.md)). Geladen und geprüft
heißt weiterhin: bitte installieren.


## Nachtrag vom 2026-11-15: ein eigener Freigabeschlüssel

**Seit 1.0.0 prüft die Anwendung gegen den Freigabeschlüssel
`44C59428DE1F766E`** (Fingerabdruck
`97721629ABDC00BA1711488244C59428DE1F766E`), nicht mehr gegen den
Commitschlüssel, den dieser Vermerk nennt. Der Vertrag bleibt: genau ein
benannter Anker, eingebettet, und ohne Signatur wird nichts geladen
([0205](0205-1-0-0-ist-ein-neuanfang.md)).
