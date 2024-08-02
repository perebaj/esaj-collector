package esaj

import (
	"testing"
)

func Test_numeroDigitoAnoUnificado(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"
	want := "1029989-06.2022"

	got, err := numeroDigitoAnoUnificado(processID)
	if err != nil {
		t.Errorf("numeroDigitoAnoUnificado() error = %v", err)
		return
	}
	if got != want {
		t.Errorf("numeroDigitoAnoUnificado() = %v, want %v", got, want)
	}
}

func Test_foroNumeroUnificado(t *testing.T) {
	processID := "1029989-06.2022.8.26.0053"
	want := "0053"

	got, err := foroNumeroUnificado(processID)
	if err != nil {
		t.Errorf("foroNumeroUnificado() error = %v", err)
		return
	}
	if got != want {
		t.Errorf("foroNumeroUnificado() = %v, want %v", got, want)
	}
}
