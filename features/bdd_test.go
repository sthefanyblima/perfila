package features_test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"

	"perfila/features/steps"
)

// Só @implantacao fica fora até a T-24 (Docker/Nginx), quando ganha job próprio no CI.
// Sobrescreva com BDD_TAGS, por exemplo: BDD_TAGS="@lgpd" go test ./features/...
const tagsPadrao = "~@implantacao && ~@wip"

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
