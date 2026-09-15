// Package wintask cadastra o próprio executável no Agendador de Tarefas do
// Windows, pra rodar sozinho todo dia — sem depender de alguém (TI,
// programador) configurar isso manualmente depois. Só funciona no Windows;
// em outros sistemas, EnsureDailyTask não faz nada (retorna nil).
package wintask

import "runtime"

const nomeTarefa = "ColetaNotasFiscaisAutomatica"

// EnsureDailyTask garante que existe uma tarefa agendada rodando este
// executável todo dia nos horários informados (formato "HH:MM"), com "rodar
// assim que possível" ligado.
func EnsureDailyTask(horarios ...string) error {
	if runtime.GOOS != "windows" {
		return nil // no-op fora do Windows
	}
	return garantirTarefa(horarios...)
}

// EnsurePeriodicTask garante que existe uma tarefa agendada rodando em ciclo
// contínuo pelo intervalo informado, sem precisar escolher horários fixos.
func EnsurePeriodicTask(intervaloMinutos int) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	return garantirTarefaPeriodica(intervaloMinutos)
}

func Status() (string, bool, error) {
	if runtime.GOOS != "windows" {
		return "", false, nil
	}
	return statusTarefa()
}

// RemoverTarefa apaga a tarefa agendada, se existir — usada pelo botão
// "Desinstalar" do painel, pra quem quer parar de usar o programa de
// verdade (não só fechar a janela) e não deixar nada rodando sozinho no
// Agendador de Tarefas do Windows.
func RemoverTarefa() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	return removerTarefa()
}
