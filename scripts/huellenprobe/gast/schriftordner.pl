#!/usr/bin/perl
# **Aktiviert macOS Schriften in Unterordnern von `~/Library/Fonts`?** —
# die Klärung #2411.
#
# Nachgelesen war nur, was Font Book beim Hinzufügen tut; was macOS mit einem
# **von Hand** dorthin gezogenen Ordner macht, sagt keine Quelle. Hier wird es
# gefragt: eine Schrift aus dem Pack liegt erst in einem Unterordner, dann
# flach — und jedes Mal fragt AppKit (`NSFontManager.availableFonts`, über
# JXA, denn der Gast hat weder Node noch Python), ob sie da ist.
#
#     perl schriftordner.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;
use Time::HiRes qw(time sleep);

binmode STDOUT, ':encoding(UTF-8)';

`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/schrift";
my $J = JSON::PP->new->utf8->canonical;
my $H = $ENV{HOME};
my $UNTER = "$H/Library/Fonts/R2411 Probe";

# **Zwei verschiedene Schriften**, je eine Lage: der erste Lauf nahm dieselbe
# für Unterordner und Gegenprobe, und flach blieb sie unsichtbar — fontd
# merkte sich die eben entfernte Datei. Eine Schrift je Frage, und keine, die
# in diesem Gast schon einmal lag.
my @faelle = (
  { lage => 'im Unterordner', datei => 'vanilla whale.ttf', name => 'VanillaWhale-Regular', ziel => $UNTER },
  { lage => 'flach (Gegenprobe)', datei => 'MagicMedieval.ttf', name => 'MagicMedieval', ziel => "$H/Library/Fonts" },
);

my %e = (maschine => { system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r });
my @schritte;

# da fragt AppKit nach einer Schrift — mehrmals, denn fontd meldet eine neue
# Datei nicht sofort.
sub da {
  my ($name) = @_;
  my $js = qq{ObjC.import("AppKit"); ObjC.deepUnwrap(\$.NSFontManager.sharedFontManager.availableFonts).filter((f) => f === "$name").length};
  for (1 .. 20) {
    my $zahl = `osascript -l JavaScript -e '$js' 2>/dev/null`;
    chomp $zahl;
    return 1 if ($zahl // 0) > 0;
    sleep 1;
  }
  return 0;
}

sub schritt {
  my ($was, $tun) = @_;
  my $ergebnis = eval { $tun->() };
  push @schritte, { teil => 'Schriftorte', was => $was, ok => $@ ? JSON::PP::false : JSON::PP::true, ms => 0,
    ($@ ? (fehler => $@ =~ s/\s+$//r) : (wert => "$ergebnis")) };
  return $ergebnis;
}

for my $f (@faelle) {
  schritt("$f->{name}: vorher nicht da", sub { da($f->{name}) ? die "schon da\n" : 'nicht da' });
  schritt("$f->{name}: $f->{lage}", sub {
    system('mkdir', '-p', $f->{ziel});
    system('cp', "$D/$f->{datei}", "$f->{ziel}/$f->{datei}") == 0 or die "kopieren\n";
    my $ja = da($f->{name});
    $e{unterordner} = $ja ? JSON::PP::true : JSON::PP::false if $f->{ziel} eq $UNTER;
    $f->{ziel} eq $UNTER || $ja or die "auch flach nicht da — die Probe misst nichts\n";
    return $ja ? 'macOS aktiviert sie' : 'macOS aktiviert sie nicht';
  });
}
system('rm', '-rf', $UNTER, map { "$H/Library/Fonts/$_->{datei}" } @faelle);

$e{schritte} = \@schritte;
open my $aus, '>', "$T/schrift-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
print "fertig\n";
