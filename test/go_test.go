package test

import (
	"fmt"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	token := "bearer dsahidfdsfdshnvihhdnvdivdusnuiidnsivnsid"
	token = strings.ReplaceAll(token, "bearer", "")
	fmt.Println(token)
}
