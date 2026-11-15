# 0181 — Der Schlüsselbund bleibt ungebunden, solange ad hoc signiert wird

**Stand:** angenommen am 2026-10-07, **ergänzt durch
[0185](0185-ein-eigener-herausgeber-truege-die-bindung.md)** am 2026-10-11:
ein eigener Herausgeber trüge die Bindung über Fassungen — gemessen; der
erste Grund unten gilt nur für ad hoc signierte Programme · SEC-05, SEC-06, SEC-23
(#2146,
#2147,
#2154).

## Die Frage

Diese Anwendung legt Geheimnisse in den Schlüsselbund von macOS: die
Zugänge fremder Dienste (KI-Anbieter, Kartendienste, Marktplatz), die
Geheimnisse der Anmeldeanbieter, den zweiten Faktor eines Profils und den
privaten Schlüssel der Peeridentität.
**Sie schreibt und liest über `/usr/bin/security`** (`internal/schluesselbund`,
kein cgo). Der Audit fand: ein so angelegter Eintrag vertraut genau diesem
Werkzeug — und damit **jedem Programm des Nutzers**, das es startet.

SEC-05 sollte das an einem künstlichen Bund nachweisen und ein Zugriffs- und
Migrationskonzept liefern; SEC-06 die Bindung an die vier Programme dieser
Anwendung bauen, SEC-23 die vorhandenen Einträge umziehen.

## Was gemessen ist

Am 2026-10-07 in einem Lume-Gast, an einem künstlichen Bund, ohne
Oberfläche — [`belege/sec05-schluesselbund.md`](../belege/sec05-schluesselbund.md):

- **Heute** (E1): die ACL vertraut `/usr/bin/security`; ein fremder Aufruf
  von `security find-generic-password -w` **liest still**.
- **Gebunden** (E2): mit `-T` an ein ad hoc signiertes Programm trägt die
  ACL dessen **cdhash**; der fremde Aufruf bekommt nichts, das Programm liest.
- **Eine neue Fassung** (E3) desselben Programms hat einen anderen cdhash —
  und **müsste fragen**.
- **Gesperrt** (E4) fragen beide.

## Die Entscheidung

**Für 0.7.0 bleibt der Weg über `/usr/bin/security`. Eine Bindung an die
Programme dieser Anwendung wird nicht gebaut, solange sie ad hoc signiert
ist** ([0003](0003-keine-notarisierung.md)).

- **Weil eine ad hoc signierte Bindung keine Fassung überlebt.** Der cdhash
  ist der einzige Anker, den eine Ad-hoc-Signatur hat, und er ändert sich mit
  jeder Fassung. Nach jedem Update müsste jedes der **vier** Programme —
  Fenster, `mse`, `mse-mcp`, `mse-rest` — für **jedes** Geheimnis einmal
  fragen, mit dem Kennwort des Anmeldebunds.
- **Weil zwei davon ohne Fenster laufen.** `mse-mcp` startet ein Assistent,
  `mse-rest` läuft als Dienst; eine Rückfrage dort hält einen Werkzeugruf oder
  eine Anfrage an, bis jemand am Rechner antwortet — oder sie scheitert.
- **Weil eine stille Migration nicht trägt** (SEC-23): umziehen ließen sich
  die heutigen Einträge still — `security` darf sie lesen —, aber nur einmal;
  das nächste Update bräche dieselbe Bindung wieder.
- **Weil eine eigene, nur ad hoc signierte Kennung keine Grenze wäre.** Eine
  Ad-hoc-Signatur kann eine ausdrückliche Anforderung tragen
  (`identifier "…"`), aber ohne Anker erfüllt sie jedes andere Programm mit
  demselben Bezeichner.

**Was die Bindung bräuchte, ist eine beständige Signaturidentität** — eine
Team-ID, mit der die ACL auf den Anker statt auf den cdhash zeigt. Ändert
sich [0003](0003-keine-notarisierung.md), ist SEC-06 der nächste Schritt und
SEC-23 die Migration dazu; beide sind dann einzeln zu planen.

## Was bis dahin gilt

- **Das Risiko ist benannt** — in `docs/HAERTUNG.md`: ein Programm, das unter
  demselben Nutzer läuft, kann die Geheimnisse dieser Anwendung über
  `security` lesen, ohne dass macOS fragt. Gegen ein solches Programm schützt
  der Bund heute nicht.
- **Ein gesperrter Bund schützt** (E4): wer den Anmeldebund sperrt, sperrt
  auch diese Einträge.
- **Tests fassen keinen echten Bund an** — `internal/schluesselbund` sperrt
  Testläufe für den Anmeldebund (`imLauf`).

## Verworfen

- **Binden und nach jedem Update fragen lassen.** Sicher gegen fremde
  Programme, aber bis zu vier Rückfragen je Geheimnis und Fassung, zwei davon
  an Programmen ohne Fenster — und wer „Ablehnen" drückt, verliert seine
  Zugänge, bis er sie neu einträgt.
- **Ein einziger Hauptschlüssel im Bund, die Geheimnisse verschlüsselt im
  Bestand.** Weniger Rückfragen (eine je Programm und Fassung), aber dieselbe
  Lage an jedem Update — und eine eigene Kryptographie mehr, wo §1 eine
  zweite Stelle dafür vermeidet.
- **Ein eigener, selbst signierter Zertifikatsanker auf dem Rechner.** Er
  müsste bei der Installation entstehen und das Bündel neu signieren; das ist
  ein Bau- und Freigabeweg, den diese Anwendung nicht hat.
