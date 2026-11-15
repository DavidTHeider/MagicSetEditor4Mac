# 0115 — Die Schleife ist eine Vorgabe und keine Mauer

**Stand:** angenommen · **Datum:** 2026-09-15
**Schreibt fort:** [0046](0046-doch-ein-dienst-auf-der-rueckschleife.md) — ein
Halbsatz daraus fällt
**Berührt:** [0010](0010-mcp-als-eigener-prozess.md),
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
[0044](0044-was-ein-werkzeug-schreiben-darf.md),
[0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)

## Warum es diesen Vermerk gibt

[0046](0046-doch-ein-dienst-auf-der-rueckschleife.md) sagt über den
REST-Dienst:

> Er startet nicht von selbst, **er bindet nirgendwo sonst hin**, und er
> bringt keine eigene Fachlogik mit.

**Der mittlere Halbsatz fällt.** Das ist eine Entscheidung mit Datum, Anlass
und Preis und kein Nebeneffekt eines Werkstücks (§16) — und 0046 selbst
verlangt es so: „Dieser Vermerk ist dann fortzuschreiben, nicht zu umgehen."

## Der Anlass

**Gemeldet von einem Menschen am 2026-09-14**, beim Durchsehen der fertigen
Anwendung:

> „Unklar ist mit Blick auf ein früher eröffnetes Issue auch, ob MCP-Server
> bzw. REST-Server nicht auch in der GUI einfach aktiviert und konfigurierbar
> gebunden werden können (je nachdem, ob der Nutzer vom lokalen Standard
> abweichen möchte)."

**Und am selben Tag beantwortet**, ebenfalls vom Auftraggeber: ja, beide Male
— REST **und** MCP, einschließlich einstellbarer Bindungen. Dazu der Satz, an
dem die Abwägung hängt: *„Der Nutzer darf ja der App die Berechtigungen
erteilen. Das ist in Ordnung."*

## Die Entscheidung

**Die Bindung wird einstellbar. Die Schleife bleibt die Vorgabe.**

Und drei Dinge kommen mit, ohne die die Fähigkeit eine Falle wäre:

1. **Ein Zugang, der jenseits der Schleife nicht optional ist.**
2. **TLS, sobald die Bindung die Schleife verlässt** — sonst ginge genau
   dieser Zugang im Klartext über das Netz. In
   #1533
   stand TLS ausdrücklich **nicht** im Umfang; mit dieser Entscheidung steht
   es darin.
3. **Eine sichtbare Zustimmung je Öffnung**, nicht einmal im Leben —
   dieselbe Gestalt wie der Horchposten aus N2: nur auf Aufruf, und ein
   Profilwechsel beendet es.

## Was es kostet — und das gehört an den Anfang

**`mse-rest` kann schreiben.** Karten anlegen, Felder ändern, Sets löschen.
Wer ihn auf `0.0.0.0` bindet, stellt seinen Kartenbestand ins Netz — und das
ist keine abstrakte Gefahr, sondern die Beschreibung dessen, was der Dienst
tut.

**Daraus folgt kein Nein, sondern ein Umfang.** Die drei Bedingungen oben sind
der Preis; ohne sie wäre die Fähigkeit eine Einladung, sich zu schaden.

**Und eine Grenze verschiebt sich wirklich.**
[0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md) hat drei
Vertrauensgrenzen benannt und **eine** gebaut. Diese Entscheidung baut eine
zweite: der Dienst bekommt Gegenüber, die nicht auf diesem Rechner sitzen.
Was ihn dort hält, ist der Zugang — und nicht mehr die Tatsache, dass niemand
ihn erreichen kann.

## Was ausdrücklich **nicht** dazugehört

**Ein Ausführungsweg für beliebige Befehle bleibt abgelehnt**, und der Grund
bleibt der, der im Kopf von `internal/app/technik.go` steht.

**Das ist die Trennung, die dort bisher fehlte**: „beliebige Shell" und „die
eigenen zwei Dienste schalten" sind zwei Fragen, und nur die erste war
entschieden. Der Kopf jener Datei wird entsprechend umgeschrieben — nicht,
weil er falsch war, sondern weil er mehr deckte, als er trug.

**Kein Zertifikat einer Zertifizierungsstelle.** Das Zeugnis ist selbst
unterschrieben und trägt die Peeridentität, wie in
[0113](0113-kein-libp2p-im-lan.md); wer sich verbindet, vergleicht den
Fingerabdruck, den das Fenster zeigt. Ein CA-Zertifikat für einen Dienst auf
einem Heimrechner wäre ein Konto bei einem Dritten für ein Problem, das
niemand hat.

**Kein Schalten über MCP.** Ein Sprachmodell soll die Erreichbarkeit dieses
Rechners nicht ändern — dieselbe Linie wie bei der Rotation der
Peeridentität ([0044](0044-was-ein-werkzeug-schreiben-darf.md)).

## Was verworfen wurde

**„Die Schleife bleibt eine Mauer."** Das ist der Stand von 0046, und er
hätte den Satz des Auftraggebers übergangen. Er hat außerdem einen Preis, der
selten genannt wird: wer die Anwendung von einem Tablet oder einem zweiten
Rechner im eigenen Netz erreichen will, kommt heute nur über einen
SSH-Tunnel dorthin — also über ein Werkzeug, das weit mehr öffnet als der
eine Port.

**„Einstellbar, aber ohne TLS."** Das wäre die Fassung, die am billigsten
aussieht und am meisten kostet: ein Zugangstoken, das über ein fremdes WLAN
im Klartext geht, ist kein Zugang, sondern eine Ankündigung.

**„Einstellbar, aber nur auf eine benannte Schnittstelle, nie `0.0.0.0`."**
Das klingt vorsichtig und ist es nicht: eine benannte Schnittstelle in einem
Café ist dasselbe Netz wie `0.0.0.0` zu Hause. Was schützt, ist der Zugang
und nicht die Gestalt der Adresse.
