package db

import "time"

// Company represents a company in the database
type Company struct {
	ID        int64
	UserID    int64
	Name      string
	Money     int64 // Stored in thousandths (e.g., 50,000.000 = 50000000)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MoneyToFloat converts money from integer (thousandths) to float with 3 decimals
func (c *Company) MoneyToFloat() float64 {
	return float64(c.Money) / 1000.0
}

// MoneyFromFloat converts money from float to integer (thousandths)
func MoneyFromFloat(amount float64) int64 {
	return int64(amount * 1000)
}
