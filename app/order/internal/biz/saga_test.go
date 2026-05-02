package biz_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/order/internal/biz"
)

func TestSaga_NoForbiddenRuntimeCalls(t *testing.T) {
	// Arrange
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "saga.go", nil, parser.AllErrors)
	require.NoError(t, err)

	forbidden := map[string]bool{
		"time.Now":        true,
		"time.Since":      true,
		"time.Until":      true,
		"time.After":      true,
		"time.NewTicker":  true,
		"time.NewTimer":   true,
		"rand":            true,
		"os.Getenv":       true,
	}

	// Act + Assert
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				key := id.Name + "." + x.Sel.Name
				if forbidden[key] {
					pos := fset.Position(x.Pos())
					t.Errorf("forbidden call %s at %s", key, pos)
				}
				if id.Name == "rand" {
					pos := fset.Position(x.Pos())
					t.Errorf("forbidden rand.* call at %s", pos)
				}
			}
		case *ast.GoStmt:
			pos := fset.Position(x.Pos())
			t.Errorf("forbidden 'go func' at %s", pos)
		}
		return true
	})
}

func TestNewOrderSagaWorkflow_returnsNonNil(t *testing.T) {
	// Arrange
	cfg := biz.SagaConfig{
		MaxPaymentAttempts:  3,
		PerAttemptTimeout:   30 * time.Second,
		PaymentInitialDelay: 2 * time.Second,
		PaymentBackoffMax:   30 * time.Second,
		MarkPaidRetryMax:    5,
		MarkPaidBudget:      60 * time.Second,
	}

	// Act
	wf := biz.NewOrderSagaWorkflow(cfg)

	// Assert
	assert.NotNil(t, wf, "NewOrderSagaWorkflow should return a non-nil workflow function")
}
