package utils

import (
	"time"
)

// BrasiliaLocation é o timezone de Brasília/São Paulo
var BrasiliaLocation *time.Location

func init() {
	// Carrega o timezone de Brasília
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		// Fallback para UTC se não conseguir carregar
		BrasiliaLocation = time.UTC
	} else {
		BrasiliaLocation = loc
	}
}

// Now retorna o tempo atual no timezone de Brasília
func Now() time.Time {
	return time.Now().In(BrasiliaLocation)
}

// NowUTC retorna o tempo atual em UTC (para banco de dados)
func NowUTC() time.Time {
	return time.Now().UTC()
}

// FormatBrasilia formata um time no timezone de Brasília
func FormatBrasilia(t time.Time, layout string) string {
	return t.In(BrasiliaLocation).Format(layout)
}

// FormatBrasiliaDefault formata um time no formato padrão brasileiro
func FormatBrasiliaDefault(t time.Time) string {
	return t.In(BrasiliaLocation).Format("02/01/2006 às 15:04:05")
}
