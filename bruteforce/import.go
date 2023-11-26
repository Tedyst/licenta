package bruteforce

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/tedyst/licenta/db"
)

func ImportFromReader(ctx context.Context, reader io.Reader, database db.TransactionQuerier) error {
	scanner := bufio.NewScanner(reader)

	batch := []string{}
	for scanner.Scan() {
		batch = append(batch, scanner.Text())
		if len(batch) == 100000 {
			err := database.InsertBruteforcePasswords(ctx, batch)
			if err != nil {
				return err
			}
			fmt.Printf("Inserted %d passwords\n", len(batch))
			batch = []string{}
		}
	}

	err := database.InsertBruteforcePasswords(ctx, batch)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted %d passwords\n", len(batch))

	return nil
}
