#!/usr/bin/perl
# **B02 im Gast** — dieselben Messungen wie `scripts/huellenprobe/budgets.mjs`,
# in einer VM gefahren
# (#2221).
#
# **Perl und nicht Node**, weil der Gast keins hat und keins bekommen soll:
# Perl 5.34 liegt jedem macOS bei, samt `JSON::PP`, `HTTP::Tiny` und
# `IO::Socket::UNIX`. **Die JavaScript-Stücke sind nicht abgeschrieben** — sie
# kommen aus `messungen.json`, die `scripts/budgetsgast.mjs` aus dem Drehbuch
# und der Hüllenprobe schreibt; gemessen wird in der Seite mit denselben
# Zeichen.
#
# Gestartet wird wie am Wirt über LaunchServices, mit eigenem Zuhause,
# Bestand und Bund je Ort — der Bund mit dem Probekennwort der Hüllenprobe,
# **kein Kennwort des Gastes**. Was am Wirt die Abnahmeleitung und der MCP
# von Wails tun, tun hier `IO::Socket::UNIX` und `HTTP::Tiny`.
#
#     perl budgets.pl <geteiltes Verzeichnis im Gast>
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

use Gast qw(lesen jetzt buendel messungen ort mse frage taste seite warte starte beende klicke textziel aufraeumen laeuft);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $J = JSON::PP->new->utf8->canonical;
my $W = JSON::PP->new->allow_nonref;

my $m = $J->decode(lesen("$T/b02/messungen.json"));
messungen($m);

# **Das Bündel liegt auf der Platte des Gastes**, nicht im geteilten
# Verzeichnis: ein Start über das geteilte Dateisystem mäße dessen Weg mit.
# **Ein eigenes Verzeichnis je Lauf** — zwei Läufe nehmen einander nichts weg.
my $bau = tempdir('/tmp/b02-abnahme-XXXXXX', CLEANUP => 1);
my $APP = "$bau/Abnahme.app";
system('ditto', "$T/b02/Abnahme.app", $APP) == 0 or die "Bündel\n";
buendel($APP);

sub prozesse {
  my %p;
  for (`ps -Ao pid=,rss=,comm=`) { my ($pid, $rss, $name) = /^\s*(\d+)\s+(\d+)\s+(.*)$/ or next; $p{$pid} = { rss => $rss, name => $name } }
  return \%p;
}
my $vorher = prozesse();
sub speicher {
  my ($pid) = @_;
  my $p = prozesse();
  my ($web, $gpu) = (0, 0);
  for my $k (keys %$p) {
    next if $vorher->{$k};
    $web += $p->{$k}{rss} if $p->{$k}{name} =~ /WebKit\.WebContent/;
    $gpu += $p->{$k}{rss} if $p->{$k}{name} =~ /WebKit\.GPU/;
  }
  return { anwendung => int(($p->{$pid}{rss} // 0) / 1024), web => int($web / 1024), gpu => int($gpu / 1024) };
}

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r,
  speicher => int(`sysctl -n hw.memsize` / 1024 / 1024), last => scalar `sysctl -n vm.loadavg` =~ s/\s+$//r });
my @fehler;
my $schritt = 'Beginn';

# **Der Bildschirm des Gastes bleibt wach** (`caffeinate -du`, keine
# Einstellung): ohne Betrachter schläft er, und WebKit drosselt dann
# `requestAnimationFrame` und die Zeitgeber — gerade das, woran die Vorschau
# und das Rollen gemessen werden.
system('caffeinate -du -t 3600 >/dev/null 2>&1 &');

# **Ein frisch gestarteter Gast rechnet minutenlang an sich selbst** — der
# erste Lauf traf eine Last von 58, und die Oberfläche stand nach 30 s noch
# nicht. Gemessen wird erst, wenn die Last einer Minute unter 1,5 liegt,
# höchstens nach 15 Minuten; der Wert steht im Beleg.
my $bis = time() + 900;
while (time() < $bis) { last if (split ' ', `sysctl -n vm.loadavg` =~ s/[{}]//gr)[0] < 1.5; sleep 10 }
$e{maschine}{lastBeimMessen} = `sysctl -n vm.loadavg` =~ s/\s+$//r;

eval {
  my $saat = "$T/b02/probe-werkstatt.mse-set";
  my (@kalt, @warm);
  $schritt = 'Kaltstarts';
  for my $i (1 .. 5) {
    my $o = ort();
    mse($o, 'einstellungen', 'setzen', 'ersteschritte', 'aus');
    mse($o, 'oeffnen', $saat);
    my ($ms) = starte($o);
    push @kalt, $ms;
    push @fehler, "Kaltstart $i: ⌘Q beendet nicht" unless beende();
  }
  $schritt = 'warme Starts';
  my $o = ort();
  mse($o, 'einstellungen', 'setzen', 'sprache', 'de');
  mse($o, 'einstellungen', 'setzen', 'ersteschritte', 'aus');
  mse($o, 'oeffnen', $saat);
  my $pid;
  for my $i (0 .. 6) {
    my ($ms, $p) = starte($o);
    $pid = $p;
    push @warm, $ms if $i > 0;
    last if $i == 6;
    push @fehler, "Start $i: ⌘Q beendet nicht" unless beende();
  }
  @e{qw(kalt warm)} = (\@kalt, \@warm);
  $e{speicherNachStart} = speicher($pid);

  # **Zeichnet die Seite überhaupt?** Ohne Bilder gibt es keine Vorschau- und
  # keine Rollzeit. Gemessen am 2026-10-09: der Gast steht nach dem Start
  # hinter seinem **Sperrbildschirm**, und dahinter zeichnet WebKit nicht —
  # null Bilder, mit und ohne Anzeige. Entsperren hieße ein Kennwort eingeben,
  # und das geschieht nicht (§17). Dann endet der Lauf hier, als Grenze.
  $schritt = 'die Probe auf requestAnimationFrame';
  $e{bilder} = seite("return await new Promise((fertig) => { const t0 = performance.now(); let n = 0; const uhr = setTimeout(() => fertig({ bilder: n, ms: performance.now() - t0 }), 2000); const schlag = () => { if (++n >= 30) { clearTimeout(uhr); fertig({ bilder: n, ms: performance.now() - t0 }) } else requestAnimationFrame(schlag) }; requestAnimationFrame(schlag) })");
  $e{sichtbar} = seite('return document.visibilityState');
  if ($e{bilder}{bilder} < 30) {
    push @{ $e{grenzen} }, "die Seite zeichnet nicht ($e{bilder}{bilder} Bilder in 2 s, Sichtbarkeit »$e{sichtbar}«) — Vorschau, Seitenwechsel und Rollen bleiben ungemessen";
    beende();
    return 1;
  }

  # Feld → Vorschaubild, wie am Wirt: abwechselnd ein Zeichen und sein Löschen.
  $schritt = 'Feld → Vorschaubild';
  taste('2', 'befehl');
  warte('das Set', sub { seite("return document.body.innerText.includes('Werkstatt-Testkarten')") }, 20);
  klicke(textziel('Werkstatt-Testkarten'));
  my $feld = $m->{namensfeld};
  warte('das Namensfeld', sub { seite("return !!($feld)") }, 20);
  klicke($feld);
  my (@vorschau, @gruende);
  my $daneben = 0;
  for (my $i = 0; @vorschau < 24 && $i < 40; $i++) {
    klicke($feld) unless seite("return document.activeElement === ($feld)");
    my $alt = seite("return ($feld).value ?? ($feld).innerText");
    seite($m->{vorschauzeit});
    taste($i % 2 == 0 ? 'x' : 'Backspace');
    my $r = seite('return window.__vorschaumessung');
    my $neu = seite("return ($feld).value ?? ($feld).innerText");
    if ($neu eq $alt) { $daneben++; next }
    push @gruende, $r->{grund} if $r->{ms} < 0;
    push @vorschau, $r->{ms};
  }
  @e{qw(vorschau gruende daneben)} = (\@vorschau, \@gruende, $daneben);
  $e{speicherNachVorschau} = speicher($pid);

  # Seitenwechsel über die Befehlspalette.
  $schritt = 'Seitenwechsel';
  my @ziele = qw(start decks sammlung sets turniere einstellungen);
  my @wechsel;
  for my $i (0 .. 23) {
    my $ziel = $ziele[$i % scalar @ziele];
    my $zeile = $m->{paletteZeile} =~ s/__SEITE__/$ziel/r;
    taste('k', 'befehl');
    warte('die Palette', sub { seite($m->{paletteOffen}) }, 15);
    seite("($zeile).scrollIntoView({ block: 'center' }); return true");
    seite($m->{seitenzeit});
    klicke($zeile);
    push @wechsel, seite('return window.__seitenmessung');
  }
  $e{wechsel} = \@wechsel;
  push @fehler, 'nach den Seitenwechseln: ⌘Q beendet nicht' unless beende();

  # Rollen durch Black, mit dem ganzen Pack im Zuhause dieses Ortes.
  $schritt = 'Black';
  if (-d "$T/b02/pack/data" && -e "$T/b02/Black.mse-set") {
    my $g = ort();
    my $pakete = "$g->{heim}/Library/Application Support/MagicSetEditor4Mac/pakete";
    make_path($pakete);
    system('cp', '-R', "$T/b02/pack/data/.", $pakete) == 0 or die "Pack\n";
    system('cp', '-R', "$T/b02/Black.mse-set", "$g->{ort}/Black.mse-set") == 0 or die "Black\n";
    mse($g, 'einstellungen', 'setzen', 'ersteschritte', 'aus');
    mse($g, 'oeffnen', "$g->{ort}/Black.mse-set");
    my $setname = (split /\t/, (split /\n/, mse($g, 'sets'))[0])[1];
    $e{setname} = $setname;
    starte($g);
    taste('2', 'befehl');
    my $q = $W->encode($setname);
    warte('das große Set', sub { seite("return document.body.innerText.includes($q)") }, 60);
    klicke(textziel($setname));
    warte('die lange Liste', sub { seite($m->{langeListe}) }, 120);
    $e{rollen} = seite($m->{rollzeit});
    push @fehler, 'Black: ⌘Q beendet nicht' unless beende();
  } else {
    push @fehler, 'Pack oder Black fehlt im geteilten Verzeichnis';
  }
  1;
} or push @fehler, "abgebrochen bei »$schritt« — $@" =~ s/\s+$//r;

aufraeumen();
$e{fehler} = \@fehler;
open my $aus, '>', "$T/b02-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
print scalar(@fehler) ? "mit Fehlern: @fehler\n" : "fertig\n";
