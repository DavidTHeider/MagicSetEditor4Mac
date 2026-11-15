# Das Ein-Mac-Labor — und was es nicht ersetzt

**Stand: 2026-09-23** — QA-10
(#1741).

**Zwei Dinge, getrennt geführt, wie das Issue es verlangt**: was auf **einem**
Mac reproduzierbar geprüft wird — in der Pipeline, bei jedem Lauf —, und was
nur **zwei Geräte in echten Netzen** zeigen können. Das Erste ersetzt das
Zweite nicht; lokale Simulation ist kein Beleg für mDNS über WLAN, einen
Router mit NAT, eine Firewall oder ein fremdes Gerät.

## Die Instanzen des Labors

**Jede Instanz ist ein eigener Prozess** (`mse` aus dem Bau des Laufs) mit
eigenem **Zuhause** — und damit eigenen Paketen und Zwischenspeichern, die dem
Rechner gehören —, eigenem **Bestand** samt Bildern und Protokoll, eigenem
Schlüsselbund, eigener Peeridentität und eigenem Anschluss
(`stelleAuf` in `internal/flaechen/handelsdraht_test.go`).

**Zwei Schalter, weil zwei Eigentümer**: `MSE_BESTAND` trennt, was dem
Bestand gehört; das Zuhause trennt, was dem Rechner gehört
(`docs/AUTORISIERUNG.md`). Ein Labor setzt beide.

**Die Trennung ist bewiesen und nicht behauptet** (QA-10.AK-01):
`TestDasLaborTrenntZweiInstanzen` lässt beide Instanzen eine Spur
hinterlassen — ein Set, eine Einstellung, eine Identität, einen horchenden
Anschluss — und sieht danach nach, **wo** sie steht: bei der eigenen Instanz
und bei der anderen **nicht**, bis hinunter zu den Dateien unter ihrem
Zuhause.

**Bis QA-10.NAT-01 folgten die Bilder dem Bestand nicht.** `MSE_BESTAND`
galt für Datenbank und Protokoll; die eingesetzten Bilder landeten am
Vorgabeort, neben einem anderen Bestand. Seitdem folgen sie ihm
(`TestWasDemBestandGehoertFolgtIhm`), und was dem Rechner gehört, bleibt
beim Zuhause (`TestWasDemRechnerGehoertBleibtBeimZuhause`).
`TestNurBenannteStellenFragenDenVorgabeort` lässt den Vorgabeort nur an
benannten Stellen zu, jede mit ihrem Grund — **die Drossel etwa gilt je
Rechner**, denn zwei Instanzen gehen über eine Adresse hinaus.

## Was zwischen ihnen schiefgehen kann

**Erzeugt über einen Vermittler** (`internal/flaechen/vermittler_test.go`): ein
TCP-Weg auf der Schleifenadresse, durch den TLS Ende zu Ende hindurchgeht —
genau das, was ein Relais im Netz ist. Er lässt sich sperren, verzögern und
kappen (QA-10.AK-02).

| Fall | Lauf | beobachtet an |
| --- | --- | --- |
| Direktverbindung | `TestEinHandelGehtUeberEinenEchtenDraht` | beiden Enden |
| Weg über einen Vermittler | `TestEinHandelGehtUeberEinenVermittler` | beiden Enden und am Vermittler |
| Ausfall | `TestBeiAusfallBleibtDerHandelImAusgang` | Berts Ausgang: offen; Annas Bestand: nichts — und nach der Rückkehr des Weges dieselbe Nachricht |
| Verlust und Wiederholung | `TestEinVerlorenerWegWirdWiederholt` | die erste Zustellung scheitert, die zweite trägt |
| Latenz | `TestLatenzHaeltDieZustellungNichtAuf` | die Dauer — sonst prüfte der Fall den schnellen Weg |
| Neustart | `TestNachDemNeustartDesHorchersGehtEsWeiter` | kein Horcher, dann ein neuer Anschluss |
| Schlüsselwechsel | `TestNachEinemSchluesselwechselGehtNichtsZumFalschen` | der Handschlag ist auf die alte Kennung festgenagelt; die Nachricht bleibt im Ausgang |

**Einen Proxyweg für Gegenstellen gibt es in diesem Bau nicht** — und damit
ist der Relaisweg der einzige Mittler, den das Labor erzeugen kann
(QA-10.AK-02 nennt »Proxy-/Relaypfad«). `internal/leitung` wählt über einen
`net.Dialer` ohne Proxy, und tailcat nähme `HTTP_PROXY` für das DERP-Relais
nur, wenn `tailscale.com/feature/useproxy` mitgebaut wäre: `derphttp.dialNode`
fragt einen Haken, den erst dieses Paket setzt, und `go list -deps
./internal/fernleitung` führt es nicht — nachgesehen am 2026-09-23, bei
tailcat v0.6.0 ([`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md)). **Der Proxy der
HTTP-Klienten** ist ein anderer Weg und steht unter eigenen Läufen
(`TestEineAnfrageGehtDurchDenProxyDerUmgebung`,
`TestDieEinstellungSchlaegtDieUmgebung`). **Ein Gerät hinter einem Proxy**
bleibt deshalb ein Fall für zwei Geräte — unten.

## Was beobachtbar ist

| Gegenstand | wo es sich zeigt | Lauf | Ebene |
| --- | --- | --- | --- |
| **Handel** | `handel stand` und `handelspost` an beiden Enden — offen, Versuche | die Tabelle darüber, `TestDerPoststandUeberJedeFlaeche` | **zwei Prozesse**, zwei Bestände |
| **Spiel** | `Partiediagnose` — Leitungen, Rolle, Runde, und kein Geheimnis | `TestDieDiagnoseNenntDieLeitungenUndKeinGeheimnis` | ein Prozess, zwei Anwendungen |
| **Wiederverbindung** am Tisch | dito, nach einer Rückkehr — und wenn sie scheitert, endet die Sitzung | `TestDieRundeErscheintErstNachDerRueckkehr`, `TestEineGescheiterteRueckkehrEndetDieSitzung`, `TestEinAbgerissenerTischRaeumtSichSelbst` | ein Prozess, zwei Anwendungen |
| **Rollenwechsel** | ein entzogenes Recht hält einen laufenden Wunsch auf | `TestEinEntzogenesRechtHaeltDenWunschAuf` | ein Prozess, zwei Anwendungen |
| **Einladung** | das Format und seine Fassung | `TestEineNeuereEinladungWirdBenanntAbgewiesen` | **nur das Format** |

**QA-10.AK-03 ist damit zur Hälfte eingelöst, und so wird es geführt.**
Beobachtbar sind alle fünf; **über zwei Prozesse** beobachtet ist nur der
Handel. Spiel, Wiederverbindung und Rollenwechsel laufen eine Ebene tiefer,
weil es eine Partie über das Netz nur am Fenster gibt — zwei Fenster trägt
S12.NAT-04 in Phase 6.11. Die Einladung über fremde Netze braucht das
Internet und bleibt Handabnahme (unten). Hier stand bis zum Zweitblick über
Phase 6.0 AK-03 als erfüllt.

**Das Spiel läuft im Labor in einem Prozess, nicht in zweien.** Eine Partie
über das Netz gibt es heute nur am Fenster; die Läufe darüber setzen zwei
Anwendungen mit getrennten Beständen in **einen** Testprozess. Das ist eine
Ebene tiefer als der Handel, und es steht hier, damit es nicht für mehr
gehalten wird.

**Am nativen Fenster aber in zwei Instanzen** — seit S12.NAT-04 (#1672) die
Hüllenprobe `nat04`: zwei Programme über LaunchServices, je mit eigenem
Bestand, Bund und eigener Identität, Horchposten und TLS über die Schleife.
Belegt sind die verdeckten Hände in beiden Fenstern, der Zug des Gastes im
Fenster des Wirts ohne dessen Zutun, die Besetzung beim Wirt und ein Satz
vom Gast zum Wirt (`docs/belege/huellenprobe/nat04.md`). **Sie fand drei
Fehler bei 100 % Abdeckung**: die Führung des Wirts am Tisch des Gastes,
Kartenbilder aus dem falschen Bestand, eine Besetzung, die nur auf
„Nachsehen" kam. **Nicht belegt**: zwei Geräte in zwei Netzen (#1998).

**Und der Handel ebenso** — die Hüllenprobe `nat04handel`: Einrichtung über
die Kommandozeile, gehandelt in beiden Fenstern, Schritt für Schritt vom
Zustimmen bis zum beidseitigen Abschluss, übertragen mit „In den Ausgang
legen" und „Jetzt übertragen" (`docs/belege/huellenprobe/nat04handel.md`).
Belegt ist, dass jedes Fenster **von selbst** zeigt, wer dran ist, dass die
Gegenseite ihr Exemplar in „Zusagen" festhält, bevor sie die Übergabe
bestätigt, und dass die Karte am Ende aus Annas Bestand in Berts gewandert
ist. **Sie fand vier Fehler bei 100 % Abdeckung**: das Fenster bemerkte
nichts, was über die Leitung kam; die Gegenseite konnte die Übergabe ohne
Deckung bestätigen und danach nicht mehr verbuchen; der Beleg stand nur als
Antwort auf den eigenen Klick da; die leere Liste sagte „Noch kein Handel"
neben einem abgeschlossenen.

**Die Einladung über fremde Netze geht über tailcat und damit über das
Internet** — ein Relais, das dieses Labor nicht nachbildet. Belegt ist hier
das Format; die Verbindung ist Handabnahme.

**Private Bestände und echte Konten bleiben unberührt**: jede Instanz hat ihr
eigenes Zuhause, ihren eigenen Bestand und ihren eigenen Bund, `internal/schluesselbund` weist den
Anmeldebund einem Testprogramm ab, und kein Lauf geht ins Netz hinaus.

## Was nur zwei Geräte zeigen

**Diese Abnahmen bleiben ausdrücklich erforderlich** (QA-10.AK-04) und
stehen in keinem Lauf:

| Fall | warum das Labor ihn nicht zeigt |
| --- | --- |
| Zwei Geräte im selben WLAN, gefunden über mDNS | Multicast über eine Funkstrecke und einen Router; das Labor hat nur die Schleifenadresse und, hinter `netzlauf`, eine Schnittstelle dieser Maschine |
| Zwei Geräte in getrennten Netzen, über tailcat | NAT, DERP als Relais, das Internet |
| Ein Gerät hinter einem Proxy oder VPN | ein Weg, den dieses Programm nicht selbst baut |
| Eine Firewall, die den Anschluss verweigert | eine Regel eines fremden Systems |

**Ihr Stand steht in der Abnahme der Freigabe**, mit Datum, Geräten und
Netzen — und bis dahin als **offen**, nicht als bestanden.
