package external

import (
	"context"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata" // fuso de São Paulo mesmo em imagens sem zoneinfo (UML8)

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/domain/port"
)

var _ port.Notificador = (*NotificadorSimulado)(nil)

// RetencaoCaixaSaida é quanto tempo uma mensagem simulada fica guardada. Spec: Q-57.
const RetencaoCaixaSaida = 24 * time.Hour

// NotificadorSimulado não envia e-mail: grava a mensagem na caixa de saída simulada, com o
// destinatário mascarado, e apaga as de mais de 24 horas a cada envio. Nada vai para o log: o
// texto leva o link, que é uma credencial (C-09). Spec: v1.10, Q-57.
type NotificadorSimulado struct {
	caixa port.CaixaSaidaSimulada
	clock port.Clock
}

func NovoNotificadorSimulado(caixa port.CaixaSaidaSimulada, clock port.Clock) *NotificadorSimulado {
	return &NotificadorSimulado{caixa: caixa, clock: clock}
}

// EnviarLinkAcesso monta a mensagem de `docs/design/03-textos.md` e a guarda. Spec: RN-19, Q-57.
func (n *NotificadorSimulado) EnviarLinkAcesso(ctx context.Context, para, link string, expiraEm, dataCadastro time.Time) error {
	agora := n.clock.Agora()
	if _, err := n.caixa.RemoverAte(ctx, agora.Add(-RetencaoCaixaSaida)); err != nil {
		return err
	}
	loc := saoPaulo()
	// O link vai no fim, sem pontuação depois dele, para nenhum leitor de e-mail colar um ponto no token.
	corpo := fmt.Sprintf("Olá. Recebemos um pedido de acesso aos dados do cadastro feito em %s. "+
		"Use o link abaixo até %s; ele só pode ser usado uma vez. Se não foi você, ignore esta mensagem. %s",
		dataCadastro.In(loc).Format("02/01/2006"), expiraEm.In(loc).Format("02/01/2006 15:04"), link)
	return n.caixa.Guardar(ctx, entity.MensagemSimulada{ID: uuid.New(), Para: MascararEmail(para),
		Assunto: "Seu link de acesso ao Perfila", Corpo: corpo, EnviadaEm: agora.UTC()})
}

// MascararEmail deixa a primeira letra e o domínio: c***@exemplo.test (03-textos.md). Spec: Q-57.
func MascararEmail(e string) string {
	local, dominio, ok := strings.Cut(strings.TrimSpace(e), "@")
	if !ok || local == "" {
		return "***"
	}
	return string([]rune(local)[0]) + "***@" + strings.ToLower(dominio)
}

func saoPaulo() *time.Location {
	if loc, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
		return loc
	}
	return time.FixedZone("BRT", -3*60*60)
}
