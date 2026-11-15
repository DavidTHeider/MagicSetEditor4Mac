#!/usr/bin/perl
# **Alles widerrufen unter launchd, wenn das Programm schon fehlt** — OPS03
# (#2349),
# der Lauf, den der Arbeitsrechner nicht tragen darf: ein LaunchAgent ist
# dauerhafte Einrichtung, und hier ist er Wegwerfware.
#
#  1. **Eingerichtet wie am Fenster**: das Zuhause aus
#     `scripts/ops03einrichtung.go` — Verweise, Shellprofil, Codex, Claude
#     Desktop, plist des Dienstes, je mit einem fremden Nachbarn —, das
#     Bündel unter `~/Applications`, der Dienst über `launchctl bootstrap`.
#  2. **Das Programm ist fort**: das Bündel liegt im Papierkorb, ein neues
#     unter anderem Namen in `~/Downloads`.
#  3. **`mse schnittstellen widerrufen` aus dem neuen Bündel**: jedes eigene
#     Ziel zurückgenommen, der Dienst entladen, seine plist weg, jede
#     Nachbardatei bytegleich.
#  4. **Ein zweites Mal**: nichts zu tun, Rückgabe 0.
#
#     perl ops03.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;
use File::Find;
use File::Path qw(make_path);
use File::Basename qw(dirname);
use Time::HiRes qw(time);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/ops03";
my $J = JSON::PP->new->utf8->canonical;
my $H = $ENV{HOME};
my $uid = $<;
my $ALT = lesen("$D/heim.txt") =~ s/\s+$//r;
my $ETIKETT = 'io.github.davidtheider.magicseteditor4mac.dienst';
my $PLIST = "$H/Library/LaunchAgents/$ETIKETT.plist";
my $A = "$H/Applications/MagicSetEditor4Mac.app";
my $B = "$H/Downloads/MagicSetEditor4Mac 2.app";
my %NACHBARN = %{ JSON::PP->new->utf8->decode(lesen("$D/heim/nachbarn.json")) };

# **Der Werkzeugort kam schon als Pfad dieses Gasts** — das Gedächtnis von
# Claude Desktop trägt eine Prüfsumme über die Datei, und in ihr steht er;
# hier ersetzt, hieße die Datei »seither geändert«.
lesen("$D/werkzeug.txt") =~ s/\s+$//r eq "$A/Contents/MacOS" or die "der Werkzeugort passt nicht zu $H\n";

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r });
my (@schritte, @ausgaben);

sub lesen { local $/; open my $f, '<:raw', $_[0] or die "$_[0]: $!\n"; my $x = <$f>; close $f; return $x }
sub schreiben { my ($p, $x) = @_; make_path(dirname($p)); open my $f, '>:raw', $p or die "$p: $!\n"; print $f $x; close $f }
sub jetzt { return time() * 1000 }

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

# geladen fragt launchd — die Rückgabe von `launchctl print`, 113 heißt nicht geladen.
sub geladen { system("launchctl print gui/$uid/$ETIKETT >/dev/null 2>&1"); return $? >> 8 }

# widerrufen ruft die Kommandozeile des neuen Bündels.
sub widerrufen {
  my $text = `'$B/Contents/MacOS/mse' schnittstellen widerrufen 2>&1`;
  my $code = $? >> 8;
  push @ausgaben, "\$ mse schnittstellen widerrufen  → $code\n$text";
  return ($code, $text);
}

eval {
  schritt(1, 'das Zuhause wie nach »Alles einrichten«', sub {
    my $n = 0;
    find({ no_chdir => 1, wanted => sub {
      my $rel = substr($File::Find::name, length("$D/heim"));
      return if $rel eq '' || $rel eq '/nachbarn.json' || $rel eq '/befehle.txt' || -d $File::Find::name && !-l $File::Find::name;
      my $ziel = "$H$rel";
      make_path(dirname($ziel));
      if (-l $File::Find::name) {
        symlink(readlink($File::Find::name) =~ s/\Q$ALT\E/$H/gr, $ziel) or die "$ziel: $!\n";
      } else {
        schreiben($ziel, lesen($File::Find::name) =~ s/\Q$ALT\E/$H/gr);
      }
      $n++;
    } }, "$D/heim");
    return "$n Einträge";
  });

  schritt(1, 'das Bündel unter ~/Applications', sub {
    system('ditto', '--noqtn', "$D/MagicSetEditor4Mac.app", $A) == 0 or die "einsetzen\n";
    return $A;
  });

  schritt(1, 'der Dienst geladen wie vom Fenster', sub {
    for my $zeile (split /\n/, lesen("$D/heim/befehle.txt")) {
      next if $zeile eq '';
      my @befehl = map { s/\Q$ALT\E/$H/gr =~ s{^gui/\d+}{gui/$uid}r } split /\t/, $zeile;
      $befehl[0] eq 'launchctl' or die "unerwarteter Befehl: @befehl\n";
      system('/bin/launchctl', @befehl[1 .. $#befehl]) == 0 or die "@befehl: $?\n";
    }
    geladen() == 0 or die "nicht geladen\n";
    return 'geladen';
  });

  schritt(1, 'die eigenen Einträge stehen', sub {
    readlink("$H/.local/bin/mse") eq "$A/Contents/MacOS/mse" or die "der Verweis auf mse\n";
    lesen("$H/.zprofile") =~ /\.local\/bin/ or die "das Shellprofil\n";
    lesen("$H/.codex/config.toml") =~ /magicseteditor4mac/ or die "Codex\n";
    lesen("$H/Library/Application Support/Claude/claude_desktop_config.json") =~ /magicseteditor4mac/ or die "Claude Desktop\n";
    -f $PLIST or die "die plist\n";
    return 'fünf Ziele';
  });

  schritt(2, 'das Programm in den Papierkorb, ein neues unter anderem Namen', sub {
    make_path("$H/.Trash");
    system('mv', $A, "$H/.Trash/MagicSetEditor4Mac.app") == 0 or die "verschieben\n";
    make_path("$H/Downloads");
    system('ditto', '--noqtn', "$D/MagicSetEditor4Mac.app", $B) == 0 or die "einsetzen\n";
    return $B;
  });

  schritt(3, 'mse schnittstellen widerrufen aus dem neuen Bündel', sub {
    my ($code, $text) = widerrufen();
    $code == 0 or die "Rückgabe $code: $text\n";
    for my $ziel (qw(befehle befehlszeile codex claudedesktop dienst)) {
      $text =~ /^\Q$ziel\E\twiderrufen$/m or die "$ziel nicht widerrufen: $text\n";
    }
    return 'fünf widerrufen';
  });

  schritt(3, 'der Dienst ist entladen, seine plist fort', sub {
    my $g = geladen();
    $g == 113 or die "launchctl print gibt $g\n";
    !-e $PLIST or die "die plist steht noch\n";
    return 'entladen';
  });

  schritt(3, 'die Verweise sind fort, jede Nachbardatei bytegleich', sub {
    !-l "$H/.local/bin/mse" && !-l "$H/.local/bin/mse-mcp" or die "ein Verweis steht noch\n";
    for my $name (sort keys %NACHBARN) {
      lesen("$H/$name") eq $NACHBARN{$name} or die "$name ist nicht wie vorher\n";
    }
    return scalar(keys %NACHBARN) . ' Nachbarn';
  });

  schritt(4, 'ein zweites Mal: nichts zu tun', sub {
    my ($code, $text) = widerrufen();
    $code == 0 or die "Rückgabe $code: $text\n";
    for my $zeile (grep { $_ ne '' } split /\n/, $text) {
      $zeile =~ /\tunveraendert$/ or die "nicht unverändert: $zeile\n";
    }
    return 'unverändert';
  });
};

# **Aufräumen, was der Lauf am Gast hinterließ** — auch nach einem Fehler.
system("launchctl bootout gui/$uid/$ETIKETT >/dev/null 2>&1");
$e{schritte} = \@schritte;
$e{ausgaben} = \@ausgaben;
schreiben("$T/ops03-ergebnis.json", $J->encode(\%e));
system('sync');
