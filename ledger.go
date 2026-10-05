package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

var ErrInvalidTransaction = errors.New("invalid transaction")

var ErrInvalidBudget = errors.New("invalid budget")

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
}

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
	nextID       int
}

func NewLedger() *Ledger {
	return &Ledger{
		transactions: []Transaction{},
		budgets:      map[string]Budget{},
		nextID:       1,
	}
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive, got %.2f", ErrInvalidTransaction, tx.Amount)
	}
	if tx.Category == "" {
		return fmt.Errorf("%w: category must not be empty", ErrInvalidTransaction)
	}

	if budget, ok := l.budgets[tx.Category]; ok {
		spent := l.CategoryTotal(tx.Category)
		if spent+tx.Amount > budget.Limit {
			return fmt.Errorf(
				"%w: category=%q limit=%.2f spent=%.2f amount=%.2f",
				ErrBudgetExceeded, tx.Category, budget.Limit, spent, tx.Amount,
			)
		}
	}

	tx.ID = l.nextID
	l.nextID++
	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *Ledger) SetBudget(b Budget) error {
	if b.Category == "" {
		return fmt.Errorf("%w: category must not be empty", ErrInvalidBudget)
	}
	if b.Limit < 0 {
		return fmt.Errorf("%w: category=%q limit must be non-negative, got %.2f", ErrInvalidBudget, b.Category, b.Limit)
	}
	l.budgets[b.Category] = b
	return nil
}

func (l *Ledger) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	if err := json.NewDecoder(r).Decode(&budgets); err != nil {
		return fmt.Errorf("decode budgets json: %w", err)
	}
	for i, b := range budgets {
		if err := l.SetBudget(b); err != nil {
			return fmt.Errorf("budget at index %d: %w", i, err)
		}
	}
	return nil
}

func (l *Ledger) CategoryTotal(category string) float64 {
	var total float64
	for _, tx := range l.transactions {
		if tx.Category == category {
			total += tx.Amount
		}
	}
	return total
}

func (l *Ledger) Transactions() []Transaction {
	return slices.Clone(l.transactions)
}

func (l *Ledger) Budgets() []Budget {
	result := make([]Budget, 0, len(l.budgets))
	for _, category := range slices.Sorted(maps.Keys(l.budgets)) {
		result = append(result, l.budgets[category])
	}
	return result
}
