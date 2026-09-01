# ADR-011 - Identité git par répertoire via `includeIf`

- **Statut** : accepté
- **Date** : 2026-08
- **Commits** : `b286cad` (initial commit)

## Contexte

Deux identités git coexistent : personnelle (`wallforfry@gmail.com`) et
professionnelle. Le profil de machine ne permet pas de trancher
([ADR-004](004-profils-perso-pro.md)) : les dépôts des deux mondes cohabitent sur
la même machine `pro`, et une identité choisie à l'apply serait fausse la moitié du
temps.

Une identité git erronée n'échoue pas : elle produit des commits signés du mauvais
nom, découverts plus tard, dans l'historique d'autrui.

## Décision

`~/.gitconfig` porte l'identité personnelle par défaut et délègue la
professionnelle à git lui-même :

```ini
[includeIf "gitdir:~/Projects/<employeur>/"]
    path = ~/Projects/<employeur>/.gitconfig
```

L'identité est donc fonction du **répertoire du dépôt**, résolue par git à chaque
commande, et non du profil résolu à l'apply.

Le fragment réel est chiffré ([ADR-016](016-depot-public-sensible-chiffre.md)) : il
porte les
informations d'employeur que [ADR-016](016-depot-public-sensible-chiffre.md) tient hors du
dépôt.

## Conséquences

- L'identité est correcte sans intervention, y compris dans un dépôt cloné le jour
  même.
- La convention devient une contrainte de rangement : un dépôt professionnel
  **doit** vivre sous le répertoire déclaré. Ailleurs, il hérite silencieusement de
  l'identité personnelle.
- Le fichier inclus est un prérequis non versionné : sur une machine neuve, son
  absence est silencieuse - git ignore un `includeIf` dont le chemin manque.
- `~/.gitconfig` reste un template pour une seule autre raison, le chemin de `gh`
  résolu par `lookPath` ([ADR-017](017-clone-anonyme-gh-en-ecriture.md)).

## Alternatives écartées

- **Identité templatée par profil** : fausse dans les dépôts personnels d'une
  machine `pro`, ce qui est le cas courant.
- **`user.useConfigOnly = true`**, forçant une déclaration explicite par dépôt :
  correct mais bruyant, une erreur à chaque premier commit dans un dépôt neuf.
- **Un hook `pre-commit` vérifiant l'identité** : détecte au lieu de prévenir, et
  doit être installé dans chaque dépôt.
- **Versionner en clair le `.gitconfig` de l'employeur** ici : y ferait entrer les
  informations d'employeur que le dépôt tient à l'écart.
