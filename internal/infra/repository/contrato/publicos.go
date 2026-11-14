package contrato

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"perfila/internal/domain/entity"
	"perfila/internal/shared/errs"
)

// Spec: 3.6, RN-17, RN-18 — só os perfis do segmento, em ordem de ClienteID.
func testarPerfisPorSegmento(t *testing.T, r Repositorios) {
	ctx := context.Background()
	vip := entity.Segmento{ID: uuid.New(), Nome: "vip", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true}
	fiel := entity.Segmento{ID: uuid.New(), Nome: "fiel", Criterio: "frequenciaCompras >= 3", Prioridade: 2, Ativo: true}
	r.AdicionarSegmento(t, vip)
	r.AdicionarSegmento(t, fiel)
	var esperados []entity.PerfilCliente
	for i, seg := range []*uuid.UUID{&vip.ID, &fiel.ID, &vip.ID, nil} {
		c := cliente(t, r)
		p := entity.PerfilCliente{ID: uuid.New(), ClienteID: c.ID, FrequenciaCompras: i, TicketMedio: 600, CategoriaPreferida: "livros", SegmentoID: seg}
		_, err := r.Perfis.Salvar(ctx, p)
		ok(t, err)
		if seg == &vip.ID {
			esperados = append(esperados, p)
		}
	}
	if esperados[1].ClienteID.String() < esperados[0].ClienteID.String() {
		esperados[0], esperados[1] = esperados[1], esperados[0]
	}
	obtido, err := r.Perfis.ListarPorSegmento(ctx, vip.ID)
	ok(t, err)
	igual(t, "perfis do vip", obtido, esperados)
	vazio, err := r.Perfis.ListarPorSegmento(ctx, uuid.New())
	ok(t, err)
	igual(t, "segmento sem perfis", vazio, []entity.PerfilCliente{})
}

// Spec: RN-17, 3.4.3, Q-51 — exportação atômica, membros por estado, arquivo com prazo.
func testarPublico(t *testing.T, r Repositorios) {
	ctx := context.Background()
	seg := entity.Segmento{ID: uuid.New(), Nome: "vip", Criterio: "ticketMedio >= 500", Prioridade: 1, Ativo: true}
	r.AdicionarSegmento(t, seg)
	a, b := cliente(t, r), cliente(t, r)
	exportacao := func(tipo entity.TipoExportacao, quando time.Time, n int) entity.ExportacaoPublico {
		return entity.ExportacaoPublico{ID: uuid.New(), SegmentoID: seg.ID, SegmentoNome: seg.Nome, Tipo: tipo,
			Destino: entity.DestinoArquivoCSV, Operador: "analista", Data: quando, Quantidade: n,
			HashArquivo: "h-" + quando.Format(time.RFC3339), NomeArquivo: "publico.csv", ArquivoExpiraEm: quando.Add(24 * time.Hour)}
	}
	membro := func(e entity.ExportacaoPublico, c entity.Cliente, hash string) entity.MembroPublico {
		m, err := entity.NovoMembro(uuid.New(), e.ID, seg.ID, c.ID, hash, e.Data)
		ok(t, err)
		return m
	}

	inc := exportacao(entity.ExportacaoInclusao, agora, 2)
	ma, mb := membro(inc, a, "aaaa"), membro(inc, b, "bbbb")
	ok(t, r.Publicos.RegistrarExportacao(ctx, inc, []byte("email_sha256,segmento\n"), []entity.MembroPublico{mb, ma}))
	exportados, err := r.Publicos.ListarMembrosPorSegmento(ctx, seg.ID, entity.MembroExportado)
	ok(t, err)
	igual(t, "membros por hash", exportados, []entity.MembroPublico{ma, mb})
	e, conteudo, err := r.Publicos.BuscarArquivo(ctx, inc.ID)
	ok(t, err)
	igual(t, "exportação lida", e, inc)
	igual(t, "conteúdo", string(conteudo), "email_sha256,segmento\n")

	// Repetir a exportação falha sem gravar nada.
	if err := r.Publicos.RegistrarExportacao(ctx, inc, nil, []entity.MembroPublico{membro(inc, a, "cccc")}); err == nil {
		t.Fatal("exportação repetida deveria falhar")
	}
	exportados, err = r.Publicos.ListarMembrosPorSegmento(ctx, seg.ID, entity.MembroExportado)
	ok(t, err)
	igual(t, "membros depois da falha", len(exportados), 2)

	// Propagação: o membro de "a" fica pendente; ID inexistente desfaz tudo.
	ok(t, ma.MarcarRemocao(agora.Add(time.Hour)))
	fantasma := mb
	fantasma.ID = uuid.New()
	if err := r.Publicos.AtualizarMembros(ctx, []entity.MembroPublico{ma, fantasma}); err == nil {
		t.Fatal("membro inexistente deveria falhar")
	}
	pendentes, err := r.Publicos.ListarMembrosPorCliente(ctx, a.ID, entity.MembroRemocaoPendente)
	ok(t, err)
	igual(t, "nada pendente depois da falha", pendentes, []entity.MembroPublico{})
	ok(t, r.Publicos.AtualizarMembros(ctx, []entity.MembroPublico{ma}))
	pendentes, err = r.Publicos.ListarMembrosPorSegmento(ctx, seg.ID, entity.MembroRemocaoPendente)
	ok(t, err)
	igual(t, "pendente", pendentes, []entity.MembroPublico{ma})

	// Remoção: nova exportação que passa o membro a REMOVIDO na mesma operação.
	rem := exportacao(entity.ExportacaoRemocao, agora.Add(2*time.Hour), 1)
	ok(t, ma.Remover(rem.ID, rem.Data))
	ok(t, r.Publicos.RegistrarExportacao(ctx, rem, []byte("x"), []entity.MembroPublico{ma}))
	removidos, err := r.Publicos.ListarMembrosPorCliente(ctx, a.ID, entity.MembroRemovido)
	ok(t, err)
	igual(t, "removido", removidos, []entity.MembroPublico{ma})
	hist, err := r.Publicos.ListarExportacoes(ctx, seg.ID)
	ok(t, err)
	igual(t, "histórico, mais recente primeiro", hist, []entity.ExportacaoPublico{rem, inc})

	// Retenção (Q-51): no prazo exato o conteúdo vencido é apagado e o registro fica.
	n, err := r.Publicos.ApagarArquivosVencidos(ctx, inc.ArquivoExpiraEm)
	ok(t, err)
	igual(t, "arquivos apagados", n, 1)
	_, _, err = r.Publicos.BuscarArquivo(ctx, inc.ID)
	erroEsperado(t, err, errs.ErrArquivoIndisponivel)
	_, _, err = r.Publicos.BuscarArquivo(ctx, rem.ID)
	ok(t, err)
	_, _, err = r.Publicos.BuscarArquivo(ctx, uuid.New())
	erroEsperado(t, err, errs.ErrArquivoIndisponivel)
	hist, err = r.Publicos.ListarExportacoes(ctx, seg.ID)
	ok(t, err)
	igual(t, "registros ficam", len(hist), 2)
}
