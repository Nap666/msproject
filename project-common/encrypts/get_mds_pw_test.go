package encrypts

import (
	"fmt"
	"testing"
)

func TestMd5(t *testing.T) {
	str := "@LDzp5201314"
	//89812519dbbe46a97b795750903d4171
	fmt.Println(Md5(str))
}
