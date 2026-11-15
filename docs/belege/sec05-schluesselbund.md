# Beleg: wer ein Geheimnis im Schlüsselbund liest — SEC-05

**Gefahren am 2026-10-07** im Lume-Gast `mse070-a` (macOS 15.3.1) über
`lume ssh`, mit der Freigabe des Auftraggebers für die zwei Gäste. **Nur ein
künstlicher Bund** mit eigenem Kennwort in `~/sec05`, kurz in die Suchliste
des Gasts gestellt und danach gelöscht; der Anmeldebund des Gasts wurde nicht
beschrieben, der dieses Rechners nicht berührt. **Kein Geheimnis steht hier**:
das Skript gibt Längen und Status aus.

**Warum über SSH**: ohne Oberfläche kann macOS nicht fragen. Wo ein Zugriff
eine Bestätigung bräuchte, antwortet die native API mit `-25308` („User
interaction is not allowed") und `security` mit einem Fehler — so trennt der
Lauf „liest still" von „müsste fragen", ohne dass ein Dialog erscheint.

**Die Leser**:

- **der fremde security-Starter** — ein eigener Aufruf von
  `/usr/bin/security find-generic-password -w`, wie ihn jedes Programm des
  Nutzers starten kann,
- **ein eigener Leser über die native API** (`SecItemCopyMatching`), ein
  kleines Swift-Programm, ad hoc signiert wie die vier Programme dieser
  Anwendung — in zwei Fassungen mit verschiedenem cdhash
  (`90c5e276…` und `7c65fc6f…`).

| Fall | ACL des Eintrags | fremder security-Starter | eigener Leser |
| --- | --- | --- | --- |
| **E1** — wie das Produkt heute schreibt (`security add-generic-password`, ohne `-T`) | `/usr/bin/security`, Anforderung `identifier "com.apple.security" and anchor apple` | **liest still** | müsste fragen (`-25308`) |
| **E2** — gebunden an den Leser (`-T leser-v1`), ohne `security` | `leser-v1`, Anforderung `cdhash H"90c5e276…"` | **bekommt nichts** | liest |
| **E2b** — dazu die Partitionsliste `cdhash:90c5e276…` | unverändert | bekommt nichts | liest |
| **E3** — eine neue Fassung desselben Lesers (anderer cdhash) | unverändert | — | **müsste fragen** (`-25308`) |
| **E4** — der Bund gesperrt | — | bekommt nichts | müsste fragen |

**Aufgeräumt**: der künstliche Bund gelöscht (`delete rc=0`), die Suchliste
des Gasts wie vorher, `~/sec05` entfernt.

## Das Protokoll

```text
### E1 — wie das Produkt heute: add-generic-password ohne -T
description: dienst1
applications (1):
0: /usr/bin/security (OK)
requirement: identifier "com.apple.security" and anchor apple
description: dienst1
applications: <null>
description: dienst1
applications (0):
fremder security-Starter: GELESEN 10 Zeichen
leser v1: STATUS -25308 User interaction is not allowed.
### E2 — gebunden an leser-v1, ohne security (-T leser-v1)
description: dienst2
applications (1):
0: /Users/lume/sec05/leser-v1 (OK)
requirement: cdhash H"90c5e2768301a7f19461f1ae8ae1219126cd40e7"
description: dienst2
applications: <null>
description: dienst2
applications (0):
fremder security-Starter: rc=36 
leser v1: GELESEN 7 Bytes
### E2b — dazu die Partitionsliste mit dem cdhash von leser-v1
set-partition rc=0
description: dienst2
applications (1):
0: /Users/lume/sec05/leser-v1 (OK)
requirement: cdhash H"90c5e2768301a7f19461f1ae8ae1219126cd40e7"
description: dienst2
applications: <null>
description: dienst2
applications (0):
fremder security-Starter: rc=36 
leser v1: GELESEN 7 Bytes
### E3 — eine neue Fassung desselben Lesers (leser-v2, anderer cdhash)
leser v2: STATUS -25308 User interaction is not allowed.
### E4 — gesperrt
fremder security-Starter: rc=36 
leser v2: STATUS -25308 User interaction is not allowed.
### Aufräumen
delete rc=0
Ordner entfernt
    "/Users/lume/Library/Keychains/login.keychain-db"
```

## Was er nicht belegt

- **Die vier Programme dieser Anwendung selbst** als Leser — gemessen ist ein
  Leser mit derselben Art Signatur (ad hoc). Wie die Programme lesen, steht in
  `internal/schluesselbund`: über `/usr/bin/security`, also wie E1.
- **Einen Dialog**: was ein Mensch auf eine Rückfrage täte, sieht der Lauf
  nicht; er sieht nur, dass eine nötig wäre.
- **Den Anmeldebund** eines echten Nutzers und dessen vorhandene Einträge.
