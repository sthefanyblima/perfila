package web

import "html/template"

// icone devolve um ícone SVG de traço 1,5px em grade de 16px, com currentColor. Os estados têm
// forma própria, nunca só cor (docs/design/README.md, Iconografia). Decorativo: aria-hidden.
func icone(nome string) template.HTML {
	const abre = `<svg class="pf-ico" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">`
	corpo, ok := map[string]string{
		// estados
		"ok":        `<circle cx="8" cy="8" r="6.25" fill="currentColor" stroke="none"/><path class="pf-ico__visto" d="M5.2 8.2l1.9 1.9 3.7-3.9"/>`,
		"tracejado": `<circle cx="8" cy="8" r="6" stroke-dasharray="2.4 2.2"/><path d="M5.5 8h5"/>`,
		"barra":     `<circle cx="8" cy="8" r="6"/><path d="M3.8 12.2l8.4-8.4"/>`,
		"atencao":   `<path d="M8 2.2l6.2 11H1.8z"/><path d="M8 6.5v3"/><path d="M8 11.4v.1"/>`,
		"vazado":    `<circle cx="8" cy="8" r="6"/>`,
		"pendente":  `<circle cx="8" cy="8" r="5.5" stroke-dasharray="1 0"/>`,
		"falha":     `<circle cx="8" cy="8" r="6"/><path d="M5.8 5.8l4.4 4.4M10.2 5.8l-4.4 4.4"/>`,
		// ações e navegação
		"copiar":    `<rect x="5.5" y="5.5" width="8" height="8" rx="2"/><path d="M10.5 3.5v-.5a1.5 1.5 0 0 0-1.5-1.5H4a1.5 1.5 0 0 0-1.5 1.5v5A1.5 1.5 0 0 0 4 9.5h.5"/>`,
		"mais":      `<path d="M8 3v10M3 8h10"/>`,
		"busca":     `<circle cx="7" cy="7" r="4.5"/><path d="M10.5 10.5l3 3"/>`,
		"processar": `<path d="M3 8a5 5 0 0 1 8.6-3.5L13 6"/><path d="M13 2.5V6H9.5"/><path d="M13 8a5 5 0 0 1-8.6 3.5L3 10"/><path d="M3 13.5V10h3.5"/>`,
		"fechar":    `<path d="M4 4l8 8M12 4l-8 8"/>`,
		"menu":      `<path d="M2.5 4.5h11M2.5 8h11M2.5 11.5h11"/>`,
		"info":      `<circle cx="8" cy="8" r="6"/><path d="M8 7.2v4"/><path d="M8 4.8v.1"/>`,
		"sair":      `<path d="M6 2.5H3.5a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1H6"/><path d="M10 11l3-3-3-3"/><path d="M13 8H6"/>`,
		"chave":     `<circle cx="5.5" cy="10.5" r="3"/><path d="M7.6 8.4l5.4-5.4"/><path d="M11 5l1.5 1.5"/>`,
	}[nome]
	if !ok {
		return ""
	}
	return template.HTML(abre + corpo + `</svg>`) //nolint:gosec // conteúdo fixo, sem entrada do usuário
}

// iconeDoEstado escolhe a forma do estado do domínio (tabela de iconografia).
func iconeDoEstado(estado string) string {
	switch tomDoEstado(estado) {
	case "ok":
		return "ok"
	case "critical":
		return "barra"
	case "warn":
		return "atencao"
	}
	if estado == "NAO_CONCEDIDO" {
		return "tracejado"
	}
	return "vazado"
}
