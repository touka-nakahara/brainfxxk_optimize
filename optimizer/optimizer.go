package optimizer

import (
	"github.com/rosylilly/brainfxxk/ast"
)

type Optimizer struct {
}

func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

func (o *Optimizer) Optimize(p *ast.Program) (*ast.Program, error) {
	exprs, err := o.optimizeExpressions(p.Expressions)
	if err != nil {
		return nil, err
	}

	prog := &ast.Program{
		Expressions: exprs,
	}

	return prog, nil
}

func (o *Optimizer) optimizeExpressions(exprs []ast.Expression) ([]ast.Expression, error) {
	// optimizeツリーで再構築する
	optimized := []ast.Expression{}
	for _, expr := range exprs {
		// 単体でのOptimize?
		optExpr, err := o.optimizeExpression(expr)
		if err != nil {
			return nil, err
		}

		switch optExpr.(type) {
		case *ast.PointerIncrementExpression:
			if len(optimized) > 0 {
				// 記号を集めている最中かつそれが連続ならそれに追加
				if last, ok := optimized[len(optimized)-1].(*ast.MOVE); ok {
					last.Count += 1
					last.Expressions = append(last.Expressions, optExpr)
					continue
				}
			}

			// 新しい記号なら新しく作る
			optExpr = &ast.MOVE{
				Count:       1,
				Expressions: []ast.Expression{optExpr},
			}

		case *ast.PointerDecrementExpression:
			if len(optimized) > 0 {
				if last, ok := optimized[len(optimized)-1].(*ast.MOVE); ok {
					last.Count -= 1
					last.Expressions = append(last.Expressions, optExpr)
					continue
				}
			}

			optExpr = &ast.MOVE{
				Count:       -1,
				Expressions: []ast.Expression{optExpr},
			}

		case *ast.ValueIncrementExpression:
			if len(optimized) > 0 {
				if last, ok := optimized[len(optimized)-1].(*ast.CALC); ok {
					last.Value += 1
					last.Expressions = append(last.Expressions, optExpr)
					continue
				}
			}

			optExpr = &ast.CALC{
				Value:       1,
				Expressions: []ast.Expression{optExpr},
			}

		case *ast.ValueDecrementExpression:
			if len(optimized) > 0 {
				if last, ok := optimized[len(optimized)-1].(*ast.CALC); ok {
					last.Value -= 1
					last.Expressions = append(last.Expressions, optExpr)
					continue
				}
			}

			optExpr = &ast.CALC{
				Value:       -1,
				Expressions: []ast.Expression{optExpr},
			}

		case *ast.WhileExpression:
			exprs := optExpr.(*ast.WhileExpression)
			// 中に入って最適化
			opBody, err := o.optimizeExpressions(exprs.Body)
			if err != nil {
				return nil, err
			}
			// ZERO RESET
			if len(opBody) == 1 {
				// - オペランドが入っていることをチェック
				if calc, ok := opBody[0].(*ast.CALC); ok {
					if calc.Value == -1 {
						optExpr = &ast.ZERORESET{
							Pos: exprs.StartPosition,
						}
					}
					break
				}

				// ZERO SHIFT
				// if mv, ok := opBody[0].(*ast.MOVE); ok {
				// 	optExpr = &ast.ZEROSHIFT{
				// 		Pos:  exprs.StartPosition,
				// 		Leap: mv.Count,
				// 	}
				// 	break
				// }

			}

			// COPY [->+<]
			if len(opBody) >= 4 {
				var plusPlace []int
				// 最初がマイナスから始まっている
				if calc, ok := opBody[0].(*ast.CALC); ok {
					if calc.Value == -1 {
						// > と　< の数が同じ
						// ついでに+の位置を覚えておく
						incptcnt := 0
						descptcnt := 0
						for _, exp := range opBody[1:] {
							if ex, ok := exp.(*ast.MOVE); ok {
								if ex.Count > 0 {
									incptcnt += ex.Count
									plusPlace = append(plusPlace, incptcnt)
								} else {
									descptcnt += ex.Count
								}
							}
						}
						if incptcnt == descptcnt {
							
						}
						optExpr = &ast.COPY{
							CopyPlace: plusPlace,
						}
					}
				}
			}

			optExpr.(*ast.WhileExpression).Body = opBody
		}

		optimized = append(optimized, optExpr)
	}

	return optimized, nil
}

func (o *Optimizer) optimizeExpression(expr ast.Expression) (ast.Expression, error) {
	return expr, nil
}
