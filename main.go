package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const budgetsFile = "budgets.json"

func main() {
	l := NewLedger()

	fmt.Println("== initial budgets via SetBudget ==")
	mustSetBudget(l, Budget{Category: "еда", Limit: 5000})
	mustSetBudget(l, Budget{Category: "транспорт", Limit: 2000})
	printBudgets(l)

	fmt.Println("\n== adding transactions ==")
	tryAdd(l, Transaction{Amount: 3000, Category: "еда", Description: "продукты на неделю"})
	tryAdd(l, Transaction{Amount: 1500, Category: "еда", Description: "ресторан"})
	tryAdd(l, Transaction{Amount: 1000, Category: "еда", Description: "доставка пиццы"})
	tryAdd(l, Transaction{Amount: 500, Category: "транспорт", Description: "такси"})
	tryAdd(l, Transaction{Amount: 2500, Category: "транспорт", Description: "проездной"})
	tryAdd(l, Transaction{Amount: 10000, Category: "одежда", Description: "куртка (нет бюджета)"})
	tryAdd(l, Transaction{Amount: -50, Category: "еда", Description: "отрицательная сумма"})
	printTransactions(l)

	fmt.Printf("\n== loading budgets from %s ==\n", budgetsFile)
	if err := loadBudgetsFromFile(l, budgetsFile); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
	printBudgets(l)

	fmt.Println("\n== retry after limit increase ==")
	tryAdd(l, Transaction{Amount: 1000, Category: "еда", Description: "доставка пиццы"})
	printTransactions(l)

	fmt.Println("\n== I/O and parsing errors ==")
	if err := loadBudgetsFromFile(l, "missing.json"); err != nil {
		fmt.Println("load error:", err)
	}
	if err := l.LoadBudgets(strings.NewReader(`[{"category": "еда", "limit": "много"}]`)); err != nil {
		fmt.Println("load error:", err)
	}
	if err := l.LoadBudgets(strings.NewReader(`[{"category": "еда", "limit": 100`)); err != nil {
		fmt.Println("load error:", err)
	}
	if err := l.LoadBudgets(strings.NewReader(`[{"category": "", "limit": 100}]`)); err != nil {
		fmt.Println("load error:", err)
	}
}

func loadBudgetsFromFile(l *Ledger, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open budgets file: %w", err)
	}
	defer f.Close()

	if err := l.LoadBudgets(bufio.NewReader(f)); err != nil {
		return fmt.Errorf("load budgets from %s: %w", path, err)
	}
	return nil
}

func mustSetBudget(l *Ledger, b Budget) {
	if err := l.SetBudget(b); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func tryAdd(l *Ledger, tx Transaction) {
	err := l.AddTransaction(tx)
	switch {
	case err == nil:
		fmt.Printf("OK       %-10s %8.2f  %s\n", tx.Category, tx.Amount, tx.Description)
	case errors.Is(err, ErrBudgetExceeded):
		fmt.Printf("REJECTED %-10s %8.2f  %s: %v\n", tx.Category, tx.Amount, tx.Description, err)
	default:
		fmt.Printf("INVALID  %-10s %8.2f  %s: %v\n", tx.Category, tx.Amount, tx.Description, err)
	}
}

func printBudgets(l *Ledger) {
	for _, b := range l.Budgets() {
		fmt.Printf("  %-12s limit=%.2f spent=%.2f\n", b.Category, b.Limit, l.CategoryTotal(b.Category))
	}
}

func printTransactions(l *Ledger) {
	fmt.Println("transactions:")
	for _, tx := range l.Transactions() {
		fmt.Printf("  #%d %-10s %8.2f  %s\n", tx.ID, tx.Category, tx.Amount, tx.Description)
	}
}
