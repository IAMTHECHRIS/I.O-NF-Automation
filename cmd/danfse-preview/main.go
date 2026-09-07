package main

import (
	"fmt"
	"os"

	"io-nf-automation/internal/danfse"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "uso: danfse-preview <entrada.xml> <saida.pdf>")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ler XML: %v\n", err)
		os.Exit(1)
	}
	pdfBytes, err := danfse.GerarDeXML(raw, "teste", "I.O NF Automation")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gerar DANFSe: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(os.Args[2], pdfBytes, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "gravar PDF: %v\n", err)
		os.Exit(1)
	}
}
