package features_test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"

	"perfila/features/steps"
)

// @slow (os cenários @implantacao, que sobem contêineres) fica fora do make verify e roda
// em make bdd-implantacao e num job próprio do CI (Q-13).
// Sobrescreva com BDD_TAGS, por exemplo: BDD_TAGS="@lgpd" go test ./features/...
const tagsPadrao = "~@slow && ~@wip"

func TestFeatures(t *testing.T) {
	tags := os.Getenv("BDD_TAGS")
	if tags == "" {
		tags = tagsPadrao
	}
	formato := os.Getenv("BDD_FORMAT")
	if formato == "" {
		formato = "progress"
	}
	suite := godog.TestSuite{
		ScenarioInitializer: steps.InitializeScenario,
		Options: &godog.Options{
			Format:   formato,
			Paths:    []string{"."},
			Tags:     tags,
			Strict:   true, // passo indefinido ou pendente reprova o CI
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("cenários BDD falharam")
	}
}
