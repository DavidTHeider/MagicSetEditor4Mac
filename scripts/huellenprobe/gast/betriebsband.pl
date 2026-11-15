#!/usr/bin/perl
# **Ein Dienst, der endet, meldet sich im Betriebsband** — OPS02
# (#2348),
# der Teil, den der Arbeitsrechner nicht tragen darf: ein LaunchAgent ist
# dauerhafte Einrichtung, und hier ist er Wegwerfware.
#
#  1. **Ein Wegwerfort, das Fenster gestartet**: kein Hinweis auf den Dienst.
#  2. **Der Dienst geladen** (`launchctl bootstrap`) über einem Bestand in
#     einem Ordner ohne Rechte: `mse-rest` endet mit 1, und launchd startet
#     ihn nach seiner Drossel neu — `KeepAlive` mit `SuccessfulExit` falsch,
#     wie die plist der Anwendung.
#  3. **Das Fenster nennt ihn** — Art `dienstBeendet`, Ausgang 1 —, und der
#     Hinweis bleibt über vier Abrufe stehen, obwohl launchd ihn dazwischen
#     neu startet: ein Abruf in seinem kurzen Lauf nimmt ihn nicht zurück.
#  4. **Der Ordner bekommt seine Rechte zurück**: der Dienst läuft durch, und
#     nach zwei Abrufen ohne Neustart geht der Hinweis.
#
# **Gefragt wird die Bindung, nicht die Seite.** Hinter dem Sperrbildschirm
# des Gastes ist die Seite verborgen (`document.visibilityState`), und das
# Fenster fragt dann nicht — so ist es gebaut. Die Seite ruft deshalb
# `Betriebsband` selbst, über dieselbe Naht und mit derselben Gestalt wie
# ihr Abruf alle 30 Sekunden; wie der Hinweis aussieht, belegt
# `betriebsband09` am Wirt.
#
#     perl betriebsband.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use File::Path qw(make_path);
use File::Temp qw(tempdir);
use FindBin;
use lib $FindBin::Bin;
use JSON::PP;
use Time::HiRes qw(time sleep);

use Gast qw(lesen jetzt buendel messungen ort mse seite warte starte beende aufraeumen);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $J = JSON::PP->new->utf8->canonical;
my $m = $J->decode(lesen("$T/ops02/messungen.json"));
messungen($m);

my $bau = tempdir('/tmp/ops02-abnahme-XXXXXX', CLEANUP => 1);
my $APP = "$bau/Abnahme.app";
system('ditto', '--noqtn', "$T/ops02/Abnahme.app", $APP) == 0 or die "Bündel\n";
buendel($APP);
my $saat = "$T/ops02/probe-werkstatt.mse-set";

my $uid = $<;
my $ETIKETT = 'io.github.davidtheider.magicseteditor4mac.dienst';
my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r });
my (@schritte, @ausgaben);

sub schritt {
  my ($teil, $was, $tun) = @_;
  my $beginn = jetzt();
  my $ergebnis = eval { $tun->() };
  my $fehler = $@ ? $@ =~ s/\s+$//r : undef;
  push @schritte, { teil => $teil, was => $was, ok => $fehler ? JSON::PP::false : JSON::PP::true,
    ms => int(jetzt() - $beginn), ($fehler ? (fehler => $fehler) : ()), (defined $ergebnis && !$fehler ? (wert => "$ergebnis") : ()) };
  die "$was: $fehler\n" if $fehler;
  return $ergebnis;
}

# lage ist, was launchd über den Dienst sagt — Zustand, Läufe, letzter Ausgang.
sub lage {
  my $text = `launchctl print gui/$uid/$ETIKETT 2>&1`;
  my %l = (code => $? >> 8);
  $l{zustand} = $1 if $text =~ /^\tstate = (.+)$/m;
  $l{laeufe} = $1 if $text =~ /^\truns = (\d+)$/m;
  $l{ausgang} = $1 if $text =~ /^\tlast exit code = (\d+)/m;
  return \%l;
}
sub lagetext { my $l = lage(); return join(', ', map { "$_ $l->{$_}" } sort keys %$l) }

# band ruft `Betriebsband` über die Naht des Fensters und gibt die Antwort.
my $RUF = q{const antwort = await fetch(window.location.origin + '/wails/runtime', { method: 'POST',
  headers: { 'x-wails-client-id': 'ops02gast', 'Content-Type': 'application/json' },
  body: JSON.stringify({ object: 0, method: 0, args: { 'call-id': 'ops02-' + String(Date.now()),
    methodName: 'github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.Betriebsband', args: [] } }) });
  return await antwort.json()};
sub band { return seite($RUF) }
sub diensthinweis {
  my ($b) = @_;
  my ($h) = grep { $_->{art} eq 'dienstBeendet' } @{ $b->{hinweise} // [] };
  return $h;
}

eval {
  my $o = ort();
  my $d = ort();
  my $plist = "$o->{heim}/Library/LaunchAgents/$ETIKETT.plist";
  my $gesperrt = "$d->{ort}/bestand";

  schritt(1, 'ein Wegwerfort, das Fenster steht, kein Hinweis auf den Dienst', sub {
    mse($o, 'einstellungen', 'setzen', 'sprache', 'de');
    mse($o, 'oeffnen', $saat);
    starte($o);
    my $b = band();
    !diensthinweis($b) or die "ein Hinweis ohne Dienst\n";
    grep({ $_ eq 'dienst' } @{ $b->{ungemessen} }) and die "der Dienst ließ sich nicht fragen\n";
    return 'kein Hinweis, ' . scalar(@{ $b->{hinweise} }) . ' Hinweise sonst';
  });

  schritt(2, 'der Dienst geladen, über einem Bestand ohne Rechte', sub {
    make_path($gesperrt);
    mse({ %$d, bestand => "$gesperrt/bestand.db" }, 'oeffnen', $saat);
    chmod 0000, $gesperrt or die "chmod: $!\n";
    make_path("$o->{heim}/Library/LaunchAgents");
    open my $f, '>:raw', $plist or die "$plist: $!\n";
    print $f <<"PLIST";
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$ETIKETT</string>
	<key>ProgramArguments</key>
	<array>
		<string>$APP/Contents/MacOS/mse-rest</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>HOME</key>
		<string>$d->{heim}</string>
		<key>MSE_BESTAND</key>
		<string>$gesperrt/bestand.db</string>
		<key>MSE_SCHLUESSELBUND</key>
		<string>$d->{bund}</string>
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>StandardOutPath</key>
	<string>$o->{ort}/dienst.log</string>
	<key>StandardErrorPath</key>
	<string>$o->{ort}/dienst.log</string>
</dict>
</plist>
PLIST
    close $f;
    system('/bin/launchctl', 'bootstrap', "gui/$uid", $plist) == 0 or die "bootstrap: $?\n";
    my $l = warte('launchd nennt einen Fehler', sub { my $l = lage(); ($l->{ausgang} // 0) != 0 ? $l : 0 }, 30, 0.5);
    push @ausgaben, "launchctl print nach dem Laden: " . lagetext();
    return "Ausgang $l->{ausgang}";
  });

  schritt(3, 'das Fenster nennt den Dienst, und der Hinweis bleibt über vier Abrufe', sub {
    my $h = warte('der Hinweis steht da', sub { diensthinweis(band()) }, 45, 1);
    $h->{ziel} eq 'einstellungen/diagnose' && $h->{werte}{ausgang} eq '1' or die "Ziel $h->{ziel}, Ausgang $h->{werte}{ausgang}\n";
    my %gesehen;
    for my $runde (1 .. 4) {
      sleep 30;
      diensthinweis(band()) or die "in Runde $runde fort\n";
      my $l = lage();
      $gesehen{$l->{laeufe} // 0} = 1;
      push @ausgaben, "Runde $runde: " . lagetext();
    }
    keys(%gesehen) > 1 or die "launchd hat ihn nicht neu gestartet — der Lauf prüfte kein Flattern\n";
    return scalar(keys %gesehen) . ' verschiedene Läufe gesehen';
  });

  schritt(4, 'die Rechte zurück: der Dienst läuft durch, der Hinweis geht', sub {
    chmod 0755, $gesperrt or die "chmod: $!\n";
    warte('der Dienst läuft', sub { (lage()->{zustand} // '') eq 'running' }, 30, 0.5);
    warte('der Hinweis ist fort', sub { diensthinweis(band()) ? 0 : 1 }, 100, 5);
    push @ausgaben, "danach: " . lagetext();
    return 'fort';
  });

  beende();
};
my $fehler = $@;

# **Aufräumen, was der Lauf am Gast hinterließ** — auch nach einem Fehler.
system("launchctl bootout gui/$uid/$ETIKETT >/dev/null 2>&1");
$e{schritte} = \@schritte;
$e{ausgaben} = \@ausgaben;
$e{fehler} = $fehler =~ s/\s+$//r if $fehler;
open my $aus, '>:raw', "$T/ops02-ergebnis.json" or die "Ergebnis: $!\n";
print $aus $J->encode(\%e);
close $aus;
system('sync');
