# Collection Pokémon TCG - Épée et Bouclier

Une application web moderne pour explorer et gérer votre collection de cartes Pokémon de la série Épée et Bouclier (SWSH1 à SWSH10).

![Pokémon TCG](static/img/pokemon-logo.png)

## 🌟 Fonctionnalités

- **Collection complète** : Accès à toutes les cartes des séries SWSH1 à SWSH10
- **Recherche avancée** : Recherchez des cartes par nom
- **Filtrage** : Filtrez les cartes par type (Feu, Eau, etc.)
- **Favoris** : Créez et gérez votre collection personnelle
- **Images HD** : Visualisez les cartes en haute qualité
- **Interface responsive** : Fonctionne sur tous les appareils

## 📦 Séries disponibles

- SWSH1 : Épée et Bouclier
- SWSH2 : Clash des Rebelles
- SWSH3 : Ténèbres Embrasées
- SWSH4 : Voltage Éclatant
- SWSH5 : Styles de Combat
- SWSH6 : Règne de Glace
- SWSH7 : Évolution Céleste
- SWSH8 : Poing de Fusion
- SWSH9 : Étoiles Brillantes
- SWSH10 : Origine Perdue

## 🚀 Installation

### Prérequis

- Go 1.16 ou supérieur
- Un navigateur web moderne

### Installation

1. Clonez le dépôt :
```bash
git clone https://github.com/Natinours/pokemon.git
cd pokemon
```

2. Installez les dépendances :
```bash
go mod download
```

3. Lancez l'application :
```bash
go run main.go
```

4. Ouvrez votre navigateur et accédez à :
```
http://localhost:8080
```

## 🛠️ Technologies utilisées

- **Backend** : Go
- **Frontend** : HTML, CSS, JavaScript
- **API** : TCGdex API
- **Cache** : Système de cache intégré
- **Stockage** : LocalStorage pour les favoris

## 📱 Utilisation

1. **Page d'accueil** : Vue d'ensemble des séries disponibles
2. **Collection** : Parcourez toutes les cartes par série
3. **Recherche** : Utilisez la barre de recherche pour trouver des cartes spécifiques
4. **Favoris** : Gérez votre collection personnelle
5. **Filtres** : Utilisez les filtres pour affiner votre recherche

## 🔄 Mise à jour des données

Les données des cartes sont automatiquement mises à jour via l'API TCGdex. Le système de cache intégré assure des performances optimales tout en maintenant les données à jour.

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à :

1. Fork le projet
2. Créer une branche pour votre fonctionnalité
3. Commiter vos changements
4. Pousser vers la branche
5. Ouvrir une Pull Request

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 🙏 Remerciements

- [TCGdex API](https://www.tcgdex.net/) pour les données des cartes
- La communauté Pokémon TCG pour son soutien
- Tous les contributeurs du projet

## 📞 Contact

Pour toute question ou suggestion, n'hésitez pas à ouvrir une issue sur GitHub. 