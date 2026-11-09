// Package integrations provides opt-in setup for the Manager's reviewed catalog.
package integrations

// Integration describes public setup data. It never contains secret values.
type Integration struct {
	ID, Name, Purpose, GuideURL, Instructions string
	URL, TokenEnv, Command                    string
	Args                                      []string
}

// Catalog returns independent values so callers cannot alter shared defaults.
func Catalog() []Integration {
	return []Integration{
		{ID: "context7", Name: "Context7", Purpose: "Documentación de librerías · recomendada", URL: "https://mcp.context7.com/mcp", TokenEnv: "CONTEXT7_API_KEY",
			GuideURL:     "https://context7.com/docs/resources/all-clients",
			Instructions: "Obtené una API key desde https://context7.com/dashboard y cargala como CONTEXT7_API_KEY en el entorno local. Reiniciá Manager y Codex después de cambiar el entorno. El archivo sólo guarda el nombre de la variable. El servicio puede aplicar límites o costos según tu cuenta."},
		{ID: "engram", Name: "Engram", Purpose: "Memoria del proyecto · recomendada", Command: "engram", Args: []string{"mcp"},
			GuideURL:     "https://github.com/Gentleman-Programming/engram/releases",
			Instructions: "Descargá el binario oficial para tu sistema, verificá sus checksums y agregá engram al PATH. Reabrí el Manager y elegí Configurar. Esta conexión usa engram mcp; no instala plugins, hooks ni configura sincronización. Elegí el proyecto en Codex antes de usar memoria. Comprobación local: disponibilidad del comando, no acceso a memorias."},
		{ID: "exa", Name: "Exa", Purpose: "Investigación web · opcional", URL: "https://mcp.exa.ai/mcp",
			GuideURL:     "https://exa.ai/docs/get-started/exa-mcp",
			Instructions: "La conexión remota permite probar búsqueda y fetch con límites del proveedor. Para cuenta, autenticación y otras funciones seguí la guía oficial. Si ya usás el plugin de Exa en la app, conservá esa conexión y evitá agregar otra. La comprobación no hace búsquedas ni inicia trabajos pagos."},
		{ID: "github", Name: "GitHub", Purpose: "Issues y PR · opcional", URL: "https://api.githubcopilot.com/mcp/", TokenEnv: "GITHUB_PAT_TOKEN",
			GuideURL:     "https://github.com/github/github-mcp-server/blob/main/docs/installation-guides/install-codex.md",
			Instructions: "Creá un token con los repositorios y permisos mínimos que necesites: https://github.com/settings/personal-access-tokens . Cargalo como GITHUB_PAT_TOKEN en el entorno local y reabrí Manager y Codex. Nunca pegues el token en el chat o en Git. Si ya usás el conector/plugin de la app, mantenelo y evitá duplicarlo. Comprobar no lista repositorios ni modifica issues/PR."},
		{ID: "egolite", Name: "Egolite", Purpose: "Validación visual e interacción · navegador, no MCP", Command: "ego-browser",
			GuideURL:     "https://lite.ego.app/",
			Instructions: "Descargá ego(lite) desde el sitio oficial para tu sistema y completá su onboarding. El onboarding registra ego-browser en PATH. Reabrí el Manager y comprobá disponibilidad. En Codex usá la skill ego-browser y su guía instalada. No se instala un MCP ni se importan datos del navegador desde Chuli. Si no hay build para tu sistema, esta integración no está disponible."},
	}
}

func Find(id string) (Integration, bool) {
	for _, item := range Catalog() {
		if item.ID == id {
			return item, true
		}
	}
	return Integration{}, false
}

func (i Integration) entry() map[string]any {
	if i.URL != "" {
		entry := map[string]any{"url": i.URL}
		if i.TokenEnv != "" {
			entry["bearer_token_env_var"] = i.TokenEnv
		}
		return entry
	}
	return map[string]any{"command": i.Command, "args": i.Args}
}
