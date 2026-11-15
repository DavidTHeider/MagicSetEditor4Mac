# Das Drahtformat — was beide Enden teilen

**D05 (#2236).**
Zwei Fassungen dieses Programms sprechen miteinander, ohne voneinander zu
wissen. Diese Datei sagt, was zwischen ihnen gilt; die Fassungen stehen mit
Aufstieg und Abstieg in [`FASSUNGEN.md`](FASSUNGEN.md), die Eigenschaften des
Transports in [`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md).

## Der Rahmen (`internal/leitung`)

| Regel | Wert | Gehalten von |
| --- | --- | --- |
| Kopf | 4 Byte, Länge **big-endian** | `leitung.Rahmenlaenge` |
| erlaubte Länge | 1 bis 1 MiB (`peer.Hoechstlaenge`) — **geprüft, bevor zugeteilt wird** | `TestEinBehaupteterRahmenStelltKeinenSpeicher` |
| Rumpf in Stücken | wird zusammengesetzt, gleich wie er ankommt | `TestEinRahmenInEinzelnenBytesKommtGanzAn` |
| abgeschnittener Rumpf | endet mit einem Konflikt, nicht mit einer halben Nachricht | `TestEinAbgeschnittenerRahmenEndetMitEinemFehler` |
| **langsamer Sender** | ein angekündigter Rumpf muss binnen **einer Minute** da sein (`rahmenfrist`), auch wenn die Sitzung keine Frist hat; die Frist des Kontexts gilt, wenn sie früher endet | `TestEinLangsamerSenderEndetMitDerRahmenfrist` |
| Abbruch | geht jeder Frist vor, auch zwischen Kopf und Rumpf | `TestEinAbbruchZwischenKopfUndRumpfGehtVor` |
| Gegenüber | TLS 1.3, Zeugnis **festgenagelt** auf die erwartete Peerkennung; ein anderes wird abgewiesen | `TestEinAnderesGegenueberAlsErwartetWirdAbgewiesen` |

## Die Nachrichten (`internal/domain`)

Jede Familie hat einen **gepinnten Vektor** unter
`internal/domain/testdata/draht/<familie>.vektor` — das Byte für Byte, was
diese Fassung sendet — und **negative Vektoren** unter
`testdata/draht/negativ/<familie>--<fall>.vektor`, die abgewiesen werden
(`internal/domain/drahtvektoren_test.go`).

| Familie | Fassung | Was der Vektor festhält | Negative Fälle |
| --- | --- | --- | --- |
| Einlösung | `Einloesefassung` 1 | JSON mit `fassung`, `nonce`, `adresse`, `name`; höchstens 4 KiB | andere Fassung, ohne Nonce, Überlänge |
| Antwort auf die Einlösung | `Einloesefassung` 1 | JSON mit `fassung` und `ergebnis` — **`Kodiere` schreibt die eigene Fassung**, auch bei einer von Hand gebauten Antwort | andere Fassung, fremdes Ergebnis |
| Einladung über fremde Netze | `Einladungsfassung` 1 | `mse4mac:1:<adresse>` — **vor dem Wählen** geprüft | andere Fassung, fremdes Schema |
| Kontrolle einer Partie | über `ProtokollSpiel` | `{"art":…,"inhalt":…}`; nur die erwarteten Arten | unerwartete Art, erfundene Art |
| Partienachricht | über `ProtokollSpiel` | Go-Feldnamen (`Art`, `Partie`, `Epoche`, `Folge`, `Von`, `Inhalt`, `Kette`) | ohne Epoche, zu kurzes Kettenglied |
| Handelsumschlag | `Handelsprotokollfassung` 1 | Go-Feldnamen | andere Fassung, ohne Vorgang, erfundene Art |
| Unterschriftsform des Umschlags | `Handelsprotokollfassung` 1 | die Bytes, die unterschrieben werden, durch `\x00` getrennt | — |

**Warum die Unterschriftsform eigens:** ändert sie sich, bricht jede
Unterschrift der Gegenseite — auch wenn das JSON gleich bleibt.

## Wiederholung

| Familie | Was eine Wiederholung tut | Gehalten von |
| --- | --- | --- |
| Partienachricht | `Empfaengerstand` nimmt nur die nächste Folge einer Epoche; eine schon gesehene ist `FolgeSchonDa` und wird verworfen, eine aus einer alten Epoche `FolgeAlteEpoche` | `FuzzEmpfaengerstand`, `TestDerEigeneZugKommtNichtZweimalAn` |
| Spielbefehl | dieselbe Kennung erzeugt keinen zweiten Zug | `TestDerselbeBefehlZweimalErzeugtKeinenZweitenZug` |
| Handelsnachricht | der Briefkasten verbucht einen Schlüssel einmal; ein zweiter Abschluss ist ein Konflikt, bevor der Bestand sich bewegt | `TestEinZweiterAbschlussBewegtNichtsBeiJedemZuschnitt` |
| Einlösung | eine eingelöste Einladung antwortet `verbraucht` | `ErrEinladungVerbraucht` |

## Wer einen Vektor ändert

**Ein geänderter Vektor ist ein Bruch.** Er wird nur mit
`DRAHT_SCHREIBEN=1` neu geschrieben, die Fassung der Familie steigt, und
`FASSUNGEN.md` bekommt die Zeile mit Aufstieg und Abstieg. Eine Gegenseite mit
der alten Fassung muss dann **vor** dem ersten Byte abgewiesen werden, nicht
mittendrin.

## Was diese Datei nicht sagt

Ob tailcat selbst sein Drahtformat hält — es sagt das ausdrücklich nicht zu
([0114](entscheidungen/0114-tailcat-gepinnt.md)); deshalb trägt die Einladung
eine Fassung. Und wie sich ein Rumpf über ein echtes Relais verhält: das misst
D03, nicht dieser Vertrag.
