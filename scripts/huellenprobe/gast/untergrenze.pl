#!/usr/bin/perl
# **Läuft das unveränderte Bündel auf der zugesagten Untergrenze?** — I02
# (#2260).
#
# `Info.plist` sagt macOS 13, und alle vier Programme tragen `minos 13.0`
# (`vtool -show-build`). **Ein Bau ist kein Lauf**: ob die Go-Laufzeit, Wails
# und die Aufrufe in WebKit und AppKit dort tragen, sagt nur ein Start. Dieser
# Treiber fährt den Kernweg zweimal — über die Kommandozeile und am Fenster —
# und schreibt jeden Schritt mit seinem Ergebnis auf, auch den, der scheitert.
#
#     perl untergrenze.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use File::Path qw(make_path);
use File::Temp qw(tempdir);
use FindBin;
use lib $FindBin::Bin;
use JSON::PP;
use POSIX qw(WNOHANG);
use Time::HiRes qw(time sleep);

use Gast qw(lesen jetzt buendel messungen ort mse taste seite warte starte beende textziel aufraeumen laeuft);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $J = JSON::PP->new->utf8->canonical;
my $W = JSON::PP->new->allow_nonref;
my $m = $J->decode(lesen("$T/i02/messungen.json"));
messungen($m);

# **Ein eigenes Verzeichnis je Lauf** — zwei Läufe nehmen einander nichts weg.
my $bau = tempdir('/tmp/i02-abnahme-XXXXXX', CLEANUP => 1);
my $APP = "$bau/Abnahme.app";
system('ditto', "$T/i02/Abnahme.app", $APP) == 0 or die "Bündel\n";
buendel($APP);
my $saat = "$T/i02/probe-werkstatt.mse-set";

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r,
  speicher => int(`sysctl -n hw.memsize` / 1024 / 1024), last => scalar `sysctl -n vm.loadavg` =~ s/\s+$//r });
my (@schritte, @grenzen);

# schritt fährt einen Schritt und schreibt ihn auf — **auch den, der scheitert**:
# eine Untergrenze, an der der dritte Schritt bricht, ist ein anderes Ergebnis
# als eine, an der schon der erste bricht.
sub schritt {
  my ($flaeche, $was, $tun) = @_;
  my $beginn = jetzt();
  my $ergebnis = eval { $tun->() };
  my $fehler = $@ ? $@ =~ s/\s+$//r : undef;
  push @schritte, { flaeche => $flaeche, was => $was, ok => $fehler ? JSON::PP::false : JSON::PP::true,
    ms => int(jetzt() - $beginn), ($fehler ? (fehler => $fehler) : ()), (defined $ergebnis && !ref $ergebnis && !$fehler ? (wert => "$ergebnis") : ()) };
  die "$was: $fehler\n" if $fehler;
  return $ergebnis;
}

system('caffeinate -du -t 3600 >/dev/null 2>&1 &');

# **Ein frisch gestarteter Gast rechnet minutenlang an sich selbst** — wie bei
# B02: erst messen, wenn die Last einer Minute unter 1,5 liegt, höchstens nach
# 15 Minuten.
# Am Wirt (`--wirt`, die Positivkontrolle) entfällt das.
my $bis = time() + ($ENV{MSE_UNTERGRENZE_OHNE_RUHE} ? 0 : 900);
while (time() < $bis) { last if (split ' ', `sysctl -n vm.loadavg` =~ s/[{}]//gr)[0] < 1.5; sleep 10 }
$e{maschine}{lastBeimMessen} = `sysctl -n vm.loadavg` =~ s/\s+$//r;

# --- die Kommandozeile: einfahren, ändern, zeichnen, sichern
eval {
  my $o = ort();
  schritt('Kommandozeile', 'Hilfe', sub { mse($o, 'hilfe') =~ /mse/ or die "keine Hilfe\n"; 1 });
  my $set = schritt('Kommandozeile', 'Set einfahren', sub { (mse($o, 'oeffnen', $saat, '--json') =~ /"setID"\s*:\s*"?(\d+)/)[0] // die "keine Setkennung\n" });
  my $karte = schritt('Kommandozeile', 'Karten lesen', sub { (mse($o, 'karten', $set, '--json') =~ /"id"\s*:\s*"?(\d+)/)[0] // die "keine Karte\n" });
  schritt('Kommandozeile', 'Feld setzen', sub { mse($o, 'feld', $set, $karte, 'name', 'Untergrenze'); mse($o, 'karten', $set, '--json') =~ /Untergrenze/ or die "der Wert steht nicht im Bestand\n"; 1 });
  schritt('Kommandozeile', 'Karte zeichnen', sub {
    my $bild = "$o->{ort}/karte.png";
    mse($o, 'zeichnen', $karte, $bild);
    open my $f, '<:raw', $bild or die "kein Bild\n"; read $f, my $kopf, 8; close $f;
    $kopf eq "\x89PNG\r\n\x1a\n" or die "kein PNG\n";
    return -s $bild;
  });
  schritt('Kommandozeile', 'Set sichern', sub { mse($o, 'sichern', $set, "$o->{ort}/wieder.mse-set"); -e "$o->{ort}/wieder.mse-set/set" or die "nichts geschrieben\n"; 1 });
  1;
} or push @grenzen, 'die Kommandozeile brach ab — die folgenden Schritte fehlen';

# --- das Fenster: starten, Set aufschlagen, ein Feld ändern, Vorschau, Ende
eval {
  my $o = ort();
  mse($o, 'einstellungen', 'setzen', 'ersteschritte', 'aus');
  mse($o, 'einstellungen', 'setzen', 'sprache', 'de');
  my $set = (mse($o, 'oeffnen', $saat, '--json') =~ /"setID"\s*:\s*"?(\d+)/)[0] // die "keine Setkennung\n";
  my ($ms) = schritt('Fenster', 'Start bis stehende Oberfläche', sub { my ($d) = starte($o); return $d });
  $e{start} = $ms;
  $e{bilder} = seite("return await new Promise((fertig) => { const t0 = performance.now(); let n = 0; const uhr = setTimeout(() => fertig({ bilder: n, ms: performance.now() - t0 }), 2000); const schlag = () => { if (++n >= 30) { clearTimeout(uhr); fertig({ bilder: n, ms: performance.now() - t0 }) } else requestAnimationFrame(schlag) }; requestAnimationFrame(schlag) })");
  $e{sichtbar} = seite('return document.visibilityState');

  # **Über die Seite und nicht über die Ereignisschlange**: hinter dem
  # Sperrbildschirm des Gastes ist das Fenster nicht das vordere, und ein
  # nativer Klick kann ins Leere gehen (B02). Geprüft wird der Weg durch
  # React, die Bindungen und Go — nicht die Maus.
  schritt('Fenster', 'Set aufschlagen', sub {
    taste('2', 'befehl');
    my $ziel = textziel('Werkstatt-Testkarten');
    warte('das Set in der Liste', sub { seite("return !!($ziel)") }, 20);
    seite("($ziel).click(); return true");
    warte('das Namensfeld', sub { seite("return !!($m->{namensfeld})") }, 20);
    return 1;
  });
  schritt('Fenster', 'Feld ändern und Vorschau neu', sub {
    my $f = $m->{namensfeld};
    my $r = seite("const f = ($f); const bild = () => document.querySelector('img[alt=\"Vorschau der Karte\"]'); const vorher = bild()?.src; f.focus(); const setze = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(f), 'value').set; setze.call(f, 'Untergrenze am Fenster'); f.dispatchEvent(new Event('input', { bubbles: true })); f.blur(); const bis = performance.now() + 15000; while (performance.now() < bis) { const b = bild(); if (b && b.src !== vorher) { await b.decode(); return { breite: b.naturalWidth, hoehe: b.naturalHeight } } await new Promise((w) => setTimeout(w, 50)) } return { fehler: 'das Bild blieb' }");
    die "$r->{fehler}\n" if $r->{fehler};
    return "$r->{breite}×$r->{hoehe}";
  });
  schritt('Fenster', 'Wert im Bestand', sub {
    warte('der Wert im Bestand', sub { mse($o, 'karten', $set, '--json') =~ /Untergrenze am Fenster/ }, 10);
    return 1;
  });
  schritt('Fenster', '⌘Q beendet', sub { beende() or die "läuft nach 15 s noch\n"; 1 });
  1;
} or push @grenzen, 'das Fenster brach ab — die folgenden Schritte fehlen';

aufraeumen();
$e{schritte} = \@schritte;
$e{grenzen} = \@grenzen;
open my $aus, '>', "$T/i02-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
my @rot = grep { !$_->{ok} } @schritte;
print @rot ? "mit Fehlern: " . join('; ', map { "$_->{was}: $_->{fehler}" } @rot) . "\n" : "fertig\n";
