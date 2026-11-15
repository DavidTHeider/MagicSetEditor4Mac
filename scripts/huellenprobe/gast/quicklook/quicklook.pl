#!/usr/bin/perl
# **Lädt macOS eine ad hoc signierte Quick-Look-Erweiterung, und was sieht
# sie?** — G06 (#2253).
#
# Im Wegwerf-Gast: ein Probebündel mit der Erweiterung aus `Vorschaubild.swift`
# wird in `~/Applications` gelegt und registriert. Das Messprogramm
# (`Messer.swift`) fragt `QLThumbnailGenerator` nach einem Vorschaubild des
# Probesets — **vorher** ohne die Erweiterung als Gegenprobe, **nachher** mit
# ihr. Es läuft über `open` in der Anmeldesitzung: `qlmanage -t` blockierte
# über SSH. Die Art der Antwort sagt, ob die Erweiterung lief, die Farbe, ob
# sie aus der Sandbox die Vorlagenpakete lesen konnte. Danach wird alles
# wieder abgemeldet und gelöscht.
#
#     perl quicklook.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;

binmode STDOUT, ':encoding(UTF-8)';

# **Nur in einem Gast** — die zweite Sicherung neben `gastziel` (Befund 1 des
# Zweitblicks über Phase I): dieser Treiber richtet ein, was nie auf einen
# Wirt gehört, und bricht deshalb ab, bevor er etwas anfasst.
`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/ql";
my $W = "$ENV{HOME}/g06";
my $A = "$ENV{HOME}/Applications/MSEProbe.app";
my $X = "$A/Contents/PlugIns/MSEVorschau.appex";
my $pakete = "$ENV{HOME}/Library/Application Support/MagicSetEditor4Mac/pakete";
my $LS = '/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister';
my $J = JSON::PP->new->utf8->canonical;
system('rm', '-rf', $W);
mkdir $W;

my %e = (maschine => { system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r });
my @protokoll;
sub laufe { my $b = shift; my $t = `$b 2>&1`; push @protokoll, "\$ $b", ($t =~ s/\s+$//r), 'rc=' . ($? >> 8); return $t }

# muss ist ein Schritt der Einrichtung, der gelingen muss — sonst ist der Lauf
# **unvollständig** und keine gemessene Ablehnung (Befund 3 des Nachlaufs
# über Phase I: ein gescheitertes `ditto` gäbe zwei Systemsymbole, und das
# sähe aus wie »die Erweiterung lief nicht«).
sub muss { my $b = shift; my $t = laufe($b); die "$b: Rückgabe " . ($? >> 8) . "\n" if $?; return $t }

# mit_frist fährt einen Befehl und beendet ihn nach einer Frist.
sub mit_frist {
  my ($frist, @befehl) = @_;
  my $pid = fork // die "fork\n";
  if (!$pid) { exec @befehl; exit 127 }
  my $bis = time() + $frist;
  while (time() < $bis) { return $? >> 8 if waitpid($pid, 1) > 0; select(undef, undef, undef, 0.2) }
  kill 'TERM', $pid;
  waitpid $pid, 0;
  # **Und das Messprogramm selbst**, nicht nur das wartende `open` (Befund 1
  # des Nachlaufs) — sein Pfad liegt im geteilten Verzeichnis dieses Laufs.
  system('pkill', '-f', "$D/Messer.app/Contents/MacOS/Messer");
  return -1;
}

# bild fragt das Messprogramm in der Anmeldesitzung nach einem Vorschaubild.
sub bild {
  my ($wohin) = @_;
  my $aus = "$W/$wohin.json";
  my $rc = mit_frist(90, 'open', '-n', '-W', '-a', "$D/Messer.app", '--env', "MSE_G06_DATEI=$D/probe.mse-set", '--env', "MSE_G06_AUS=$aus");
  push @protokoll, "\$ open -W -a Messer.app ($wohin)", "rc=$rc";
  return { fehler => "keine Antwort (open: $rc)" } unless -s $aus;
  open my $f, '<', $aus or return { fehler => "$aus: $!" };
  local $/;
  return $J->decode(<$f>);
}

# **Erst aufräumen**, was ein abgebrochener Lauf hinterließ.
laufe("pluginkit -r '$X'") if -d $X;
system('rm', '-rf', $A);

# **Nur wegräumen, was dieser Lauf anlegt** — auch im Gast.
my $anwendungsordner = "$ENV{HOME}/Library/Application Support/MagicSetEditor4Mac";
my $ordnerNeu = !-e $anwendungsordner;

# aufraeumen meldet ab und löscht, was dieser Lauf anlegte — **auch nach
# einem Abbruch** (`END`, `INT`, `TERM`; Befund 1 des Nachlaufs über Phase I)
# und mit geprüften Rückgaben: wer nicht abmelden konnte, hat nicht
# aufgeräumt (Befund 3). **»no plugin at« ist kein Fehler** — dann war nichts
# mehr zu melden; geurteilt wird am Endzustand. Ein zweiter Ruf tut nichts.
my $geraeumt = 0;
sub aufraeumen {
  return if $geraeumt++;
  my $fehler = 0;
  if (-d $X) { my $t = laufe("pluginkit -r '$X'"); $fehler ||= ($? && $t !~ /no plugin at/) }
  if (-d $A) { laufe("'$LS' -u '$A'"); $fehler ||= $? }
  system('rm', '-rf', $A, $W);
  unlink "$pakete/g06-probe.mse-style";
  system('rm', '-rf', $anwendungsordner) if $ordnerNeu;
  my $registriert = laufe('pluginkit -m -i de.mse4mac.probe.quicklook.vorschaubild') =~ /de\.mse4mac/;
  $e{aufgeraeumt} = ($fehler || $registriert || -e $A) ? 'nein' : 'ja';
  return;
}
# schreibe legt das Ergebnis samt Protokoll ins geteilte Verzeichnis — **auch
# nach einem Signal**: sonst fehlte gerade das Protokoll des Abbruchs.
sub schreibe {
  $e{protokoll} = \@protokoll;
  open my $aus, '>', "$T/g06-ergebnis.json" or die "$!\n";
  print {$aus} $J->encode(\%e);
  close $aus;
  return;
}
$SIG{INT} = $SIG{TERM} = sub { $e{abbruch} //= "Signal $_[0]"; aufraeumen(); schreibe(); exit 130 };
END { aufraeumen() }

eval {
  # Die Vorlagenpakete, wie die Anwendung sie anlegt.
  system('mkdir', '-p', $pakete);
  system('touch', "$pakete/g06-probe.mse-style");

  $e{vorher} = bild('vorher');

  system('mkdir', '-p', "$ENV{HOME}/Applications");
  muss("ditto '$D/MSEProbe.app' '$A'");
  $e{signatur} = laufe("codesign -dvv '$X'") =~ s/\s+/ /gr;
  $e{berechtigungen} = laufe("codesign -d --entitlements - --xml '$X'") =~ s/.*(<dict>.*<\/dict>).*/$1/sr;
  $e{groesse} = (split ' ', laufe("du -sk '$X'"))[0] . ' KB';
  muss("'$LS' -f -R '$A'");
  muss("pluginkit -a '$X'");
  # **Registriert ist, was `pluginkit -m` nennt** — bis zu 20 s gefragt. Ohne
  # das mäße »nachher« ohne Erweiterung und sähe aus wie »lief nicht«: eine
  # gescheiterte Einrichtung und keine gemessene Ablehnung (Befund 3).
  for (1 .. 20) {
    $e{registriert} = laufe('pluginkit -m -v -i de.mse4mac.probe.quicklook.vorschaubild') =~ s/\s+$//r;
    last if $e{registriert} =~ /de\.mse4mac/;
    sleep 1;
  }
  die "die Erweiterung ist nicht registriert\n" unless $e{registriert} =~ /de\.mse4mac/;
  # **Die Probe des Abbruchs** (Befund 1 des Nachlaufs über Phase I): liegt
  # `abbrechen` im geteilten Verzeichnis, schickt sich der Treiber hier — mit
  # registrierter Erweiterung — selbst TERM. Derselbe Weg wie ein Signal von
  # außen, aber an einer festen Stelle; ein Signal über `lume ssh` kam zu spät.
  kill 'TERM', $$ if -e "$T/abbrechen";
  sleep 2;
  $e{nachher} = bild('nachher');
  $e{protokollDesSystems} = laufe("log show --last 3m --style compact --predicate 'eventMessage CONTAINS[c] \"MSEVorschau\" OR eventMessage CONTAINS[c] \"de.mse4mac.probe\"' | tail -8");
  1;
} or $e{abbruch} = "$@" =~ s/\s+$//r;

# --- aufräumen: abmelden, löschen
aufraeumen();
schreibe();
# **Ein Abbruch ist kein Ergebnis** (Befund 7 des Zweitblicks über Phase I).
print $e{abbruch} ? "abgebrochen: $e{abbruch}\n" : "fertig\n";
exit($e{abbruch} ? 1 : 0);
