# 0114 — tailcat, gepinnt, und die Einladung trägt die Fassung

**Stand:** angenommen · **Datum:** 2026-09-15
**Schreibt fort:** [0112](0112-tailcat-statt-eines-eigenen-dienstes.md) —
bestätigt ihn und beantwortet seine offene Frage
**Berührt:** [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
[0113](0113-kein-libp2p-im-lan.md), #1456

## Warum es diesen Vermerk gibt

#1456 verlangt
ihn **vor der ersten Zeile Adapter**:

> tailcat auf einen Commit pinnen, und festlegen, **was geschieht, wenn sich
> sein Drahtformat ändert** — welche Fassungen dann noch miteinander reden, wie
> das erkannt wird und was Nutzende dann sehen. Ein Adapter ohne diese Antwort
> ist eine Zusage, die niemand einlösen kann.

## Was gemessen wurde

Am 2026-09-14/15, in einem eigenen Modul neben dem Projekt, gegen einen
Graphen, der vorher **127** Module führte.

| | tailcat v0.6.0 |
| --- | --- |
| Neue Module im Graphen | **525** |
| Fremde Pakete, die wirklich **übersetzt** werden | **301** — 163 aus `tailscale.com`, 42 aus `gvisor.dev` |
| Binärprogramm, das nichts tut als eine Konstante zu drucken | **23,5 MB** gegen 2,4 MB leer |
| Lizenz | BSD-3-Clause |
| Toolchain | verlangt **Go ≥ 1.27.1** |

**Zwei Zahlen sind unterwegs korrigiert worden, und beide gehören hierher.**

Im Graphen steht der **vollständige AWS-SDK v2**. Übersetzt wird davon **kein
einziges Paket** — er hängt an einem Zweig von `tailscale.com`, den dieser
Gebrauch nicht betritt. 525 ist die Zahl des Graphen, 301 die des Baus, und nur
die zweite steht im Binärprogramm (§10).

Und die Toolchain war **kein** Hindernis: golangci-lint **v2.13.0** brachte
„go1.27 support". Ein erster Entwurf dieses Vermerks hat daraus einen
Ablehnungsgrund gebaut; es war keiner, und der Hinweis kam vom Auftraggeber.
Wer eine Ablehnung auf eine Unmöglichkeit stützt, prüft zuerst, ob sie eine
ist.

## Die Entscheidung

**tailcat wird eingebettet, auf `v0.6.0` gepinnt.** Go steht seit dem
2026-09-15 auf 1.27.1 und golangci-lint auf v2.13.2.

## Die Antwort auf die Fassungsfrage

**Das Problem ist nicht unser Protokoll.** Der Handschlag aus N1 trägt eine
`Handschlagfassung` und **weist eine Abweichung ab statt sie
herunterzuhandeln** (`internal/peer/handschlag.go`). Das trägt weiter.

**Das Problem liegt darunter.** Ändert tailcat sein Drahtformat, kommt gar
keine Verbindung zustande — der Handschlag läuft nie, und was ein Mensch sähe,
wäre „nicht erreichbar". Ein Verbindungsfehler, der in Wahrheit ein
Fassungsfehler ist, ist die stille Form, vor der §8 warnt.

**Deshalb trägt die Einladung die Fassung, nicht der Transport.**

Eine Einladung ist nicht die rohe `tailcat.Addr`, sondern:

```
mse4mac:1:<tailcat-Addr>
```

- **Vor dem Wählen geprüft.** Wer eine Einladung mit einer höheren Zahl
  bekommt, erfährt es **bevor** irgendetwas ins Netz geht — und liest einen
  Satz, der sagt, dass die Gegenseite neuer ist.
- **Die Zahl steigt, wenn tailcats Drahtformat bricht**, und nur dann. Sie ist
  keine zweite Protokollfassung neben der des Handschlags; sie sagt einzig:
  *diese beiden Fassungen können einander erreichen.*
- **Kein Herunterhandeln.** Dieselbe Linie wie im Handschlag: wer zwei kann,
  bietet zwei an.

**Was das nicht leistet, gehört dazu:** eine Einladung derselben Fassung, deren
tailcat sich trotzdem anders verhält, endet weiter in einem Verbindungsfehler.
Die Zahl schützt vor dem **bekannten** Bruch, nicht vor dem unbemerkten. Wer
tailcat hebt, sieht in dessen Änderungsprotokoll nach und hebt die Zahl mit —
ein Wächter kann das nicht.

## Was das kostet, und es gehört ausgesprochen

1. **Keine Stabilitätszusage.** *„the Go API, the CLI flags and output, and the
   wire format may all change."* Wir pinnen; ein Aufstieg ist jedes Mal eine
   Entscheidung mit Blick ins Änderungsprotokoll.
2. **Öffentliches DERP ohne Zusage** — *„no uptime SLAs or throughput
   targets"*, gedrosselt im Relaisfall, unkomprimiert. Für N4 ist der Durchsatz
   eine reale Größe.
3. **Die Betriebsverantwortung wandert, sie verschwindet nicht.** Aus „wir
   betreiben einen Dienst" wird „wir hängen an `tailcat.dev/derpmap.json`". Das
   gehört in `docs/NETZEIGENSCHAFTEN.md` und in die Release-Notiz.
4. **301 übersetzte Fremdpakete und 21 MB Binärzuwachs.** Das ist der Preis
   dafür, dass NAT-Durchquerung und Verschlüsselung nicht selbst gebaut werden.

## Gehoben auf v0.7.0 — am 2026-10-07

**K01** (#2187): erst v0.7.0 nennt in `Server.Status()` die
Gegenstellen (`StatusBuilder{WantPeers: true}`, in der Quelle nachgelesen),
und erst damit sagt der Fernweg, ob er direkt oder über das Relais geht.

**Die Fassungsfrage, beantwortet wie oben verlangt** — erst im
Änderungsprotokoll, dann gemessen: `wire.go` ist zwischen beiden Fassungen
unverändert, und zwischen zwei VM-Gästen trägt eine Einladung von v0.7.0 zu
v0.6.0 und umgekehrt (Echo in 94 und 95 ms, `docs/belege/sec22-vmlauf.md`).
**Die Einladungsfassung bleibt deshalb bei 1.** Mitgehoben hat MVS
`tailscale.com` auf die Fassung, die tailcat verlangt, dazu `x/net`, `x/text`
und `x/crypto` — keine davon ist ein Bruch des Drahtformats.

## Was verworfen bleibt

**Pion mit einem eigenen `mse-connect`** — wir müssten einen Dienst
**betreiben**, und das ist der Grund aus
[0112](0112-tailcat-statt-eines-eigenen-dienstes.md).

**Und der Weg „Adresse von Hand" bleibt daneben stehen**, nicht als Rückfall,
sondern als eigener Weg: wer Tailscale, WireGuard, ZeroTier oder eine
Portweiterleitung schon betreibt, braucht keine Einladung.
