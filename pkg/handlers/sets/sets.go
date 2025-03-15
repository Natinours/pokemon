package sets

// SetInfo représente les informations d'un set
type SetInfo struct {
	Name      string
	CardCount int
}

// Map des informations des sets
var setInfos = map[string]SetInfo{
	"swsh1":  {"Épée et Bouclier", 202},
	"swsh2":  {"Clash des Rebelles", 192},
	"swsh3":  {"Ténèbres Embrasées", 189},
	"swsh4":  {"Voltage Éclatant", 185},
	"swsh5":  {"Styles de Combat", 163},
	"swsh6":  {"Règne de Glace", 198},
	"swsh7":  {"Évolution Céleste", 203},
	"swsh8":  {"Poing de Fusion", 264},
	"swsh9":  {"Étoiles Brillantes", 172},
	"swsh10": {"Origine Perdue", 196},
}

// GetSetName retourne le nom d'un set à partir de son ID
func GetSetName(series string) string {
	if info, ok := setInfos[series]; ok {
		return info.Name
	}
	return series
}

// GetSetInfo retourne les informations d'un set à partir de son ID
func GetSetInfo(series string) (SetInfo, bool) {
	info, ok := setInfos[series]
	return info, ok
}
