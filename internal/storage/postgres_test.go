package storage

import "testing"

func Test_newInsertProductQuery(t *testing.T) {
	type args struct {
		batchSize int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"TestCase 1", args{4}, "INSERT INTO product(name, protein, fat, carbohydrate, kcal) VALUES ($1,$2,$3,$4,$5), ($6,$7,$8,$9,$10), ($11,$12,$13,$14,$15), ($16,$17,$18,$19,$20)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newInsertProductQuery(tt.args.batchSize); got != tt.want {
				t.Errorf("newInsertProductQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
