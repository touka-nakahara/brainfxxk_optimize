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
				if mv, ok := opBody[0].(*ast.MOVE); ok {
					optExpr = &ast.ZEROSHIFT{
						Pos:  exprs.StartPosition,
						Leap: mv.Count,
					}
					break
				}

			}

			// COPY [->+<]
			// whileの中身が4つの命令でできているかをチェックする
			if len(opBody) == 4 {
				// 4つの中身がすべてCOPYと同じかをチェックする
				if calc, ok := (opBody[0]).(*ast.CALC); ok {
					if calc.Value == -1 {
						// ><の移動数が同じかどうかをチェックする
						if moveF, ok := (opBody[1]).(*ast.MOVE); ok {
							var moveValue int
							var copyPlace int
							if moveF.Count >= 1 {
								moveValue += moveF.Count
								copyPlace += moveF.Count
								if moveB, ok := (opBody[3]).(*ast.MOVE); ok {
									moveValue += moveB.Count
									if moveValue == 0 {
										var multiplier int
										if calcM, ok := (opBody[2]).(*ast.CALC); ok {
											multiplier = calcM.Value
											// COPYに置き換える
											optExpr = &ast.COPY{
												Pos:        exprs.StartPosition,
												CopyPlace:  copyPlace,
												Multiplier: multiplier,
											}
											break
										}
									}
								}
							}
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
