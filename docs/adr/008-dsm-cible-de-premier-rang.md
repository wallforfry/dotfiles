# ADR-008 - Synology DSM comme cible de premier rang

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `4a7d1e8` (`scriptTempDir`), `41a9fb9` puis `410b4c6` (`.profile`, posé puis retiré), `cdf79e1` (fragment ssh), `f29360b` (zsh)

## Contexte

Le NAS Synology n'est pas une machine d'appoint : c'est un poste sur lequel on
travaille par ssh. Or DSM enfreint à peu près toutes les hypothèses usuelles d'un
poste Unix, et chacune s'est manifestée par une panne :

- `/tmp` est monté `noexec`, d'où un « permission denied » sur le script
  d'installation de starship (`4a7d1e8`).
- Ni `apt` ni `apk` : zsh vient d'Entware, `opkg install zsh` (`f29360b`).
- `chsh` est indisponible, et DSM réinitialise le shell de `/etc/passwd` à chaque
  mise à jour (`41a9fb9`).
- Le shell ouvert par ssh **n'est pas un shell de login** : `~/.profile` n'est
  jamais lu (`410b4c6`).

Ce dernier point a coûté un aller-retour : `.profile` a été posé pour y `exec zsh`,
puis retiré une fois constaté qu'il n'était jamais lu.

## Décision

DSM est traité comme une cible de premier rang. Quatre règles en découlent :

1. **Rien d'exécutable sous `/tmp`.** `scriptTempDir` pointe sur
   `~/.cache/chezmoi`, et les scripts créent leurs répertoires temporaires sous
   `$HOME/.cache`.
2. **Aucun gestionnaire de paquets supposé.** Le script d'outillage dégrade vers
   des binaires statiques dans `~/bin`, et signale explicitement ce qu'il ne sait
   pas installer ([ADR-007](007-outillage-run-onchange.md)).
3. **La bascule vers zsh se fait côté client**, par `RemoteCommand` dans un
   fragment ssh versionné ([ADR-009](009-fragments-ssh-versionnes.md)) - aucun
   point d'accroche n'existe côté serveur.
4. **`sh` POSIX pour tout script de bootstrap**, aucune dépendance à bash 4 ni aux
   options GNU.

## Conséquences

- Toute décision d'installation se juge d'abord sur DSM : ce qui y passe passe
  partout.
- `RemoteCommand` a un coût direct : `ssh nas <commande>`, `scp` et `rsync`
  échouent, et doivent passer `-o RemoteCommand=none`. Le fragment le documente en
  commentaire.
- Le shell non interactif reste servi par `.zshenv`, jamais par `.zprofile`
  ([ADR-007](007-outillage-run-onchange.md)).
- Une vérification faite sur macOS ne dit rien de DSM. Nommer la plateforme
  exercée fait partie de la barre de vérification (`AGENTS.md`).

## Alternatives écartées

- **`~/.profile` avec `exec zsh`** : posé (`41a9fb9`) puis retiré (`410b4c6`), le
  fichier n'étant jamais lu par un shell ssh non-login sur DSM. C'est l'alternative
  la plus tentante, et elle est réfutée par la mesure.
- **`chsh`** : indisponible, et écrasé par les mises à jour de DSM.
- **Traiter le NAS comme non géré**, configuré à la main : c'est l'état de départ,
  et la raison des erreurs au login.
- **Installer Entware pour tout** : disponible, mais ajoute un gestionnaire de
  paquets non standard au chemin critique de l'installation. Réservé à zsh, qui ne
  peut pas s'en passer.
