package object

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var NativeFunctions = []struct {
	Name     string
	NativeFn *CompiledNativeFunction
}{
	{"len", &CompiledNativeFunction{NATIVE_LEN_FUNCTION}},
	{"print", &CompiledNativeFunction{NATIVE_PRINT_FUNCTION}},
	{"type", &CompiledNativeFunction{NATIVE_TYPE_FUNCTION}},
	{"to_number", &CompiledNativeFunction{NATIVE_TO_NUMBER_FUNCTION}},
	{"to_string", &CompiledNativeFunction{NATIVE_TO_STRING_FUNCTION}},
	{"time", &CompiledNativeFunction{NATIVE_TIME_FUNCTION}},
	{"prompt", &CompiledNativeFunction{NATIVE_PROMPT_FUNCTION}},
	{"random", &CompiledNativeFunction{NATIVE_RANDOM_FUNCTION}},
}

func NewError(format string, a ...interface{}) *CompiledError {
	return &CompiledError{Message: fmt.Sprint(format, a)}
}

func GetNativeFnByName(name string) *CompiledNativeFunction {
	for _, fn := range NativeFunctions {
		if fn.Name == name {
			return fn.NativeFn
		}
	}

	return nil
}

func NATIVE_LEN_FUNCTION(args []Object) Object {
	if len(args) != 1 {
		return NewError("Invalid argument count. Expected 1, got %d instead", len(args))
	}

	switch arg := args[0].(type) {
	case *String:
		return &Number{Value: float64(len(arg.Value))}
	case *Array:
		return &Number{Value: float64(len(arg.Elements))}
	}

	return NewError("Cannot use 'len' on an unsupported argument type '%s'", args[0].Type())
}

func NATIVE_PRINT_FUNCTION(args []Object) Object {
	for i, msg := range args {
		strMsg, ok := msg.(*String)
		if !ok {
			fmt.Printf("%s", msg.Inspect())
		} else {
			trimmedMsg := strMsg.Value[0:len(strMsg.Value)]
			fmt.Printf("%s", trimmedMsg)
		}

		if i != len(args)-1 {
			fmt.Printf(", ")
		}
	}

	fmt.Println()
	return NIL
}

func NATIVE_TYPE_FUNCTION(args []Object) Object {
	if len(args) != 1 {
		return NewError("Invalid argument count. Expected 1, got %d instead", len(args))
	}

	arg := args[0]
	typeStr := arg.Type()
	return &String{Value: string(typeStr)}
}

func NATIVE_TO_NUMBER_FUNCTION(args []Object) Object {
	if len(args) != 1 {
		return NewError("Invalid argument count. Expected 1, got %d instead", len(args))
	}

	switch obj := args[0].(type) {
	case *Number, *NaN, *Infinity:
		return obj
	case *String:
		v, err := strconv.ParseFloat(obj.Value, 64)
		if err != nil {
			return NAN
		}
		return &Number{Value: v}
	case *Boolean:
		if obj.Value {
			return &Number{Value: 1}
		}
		return &Number{Value: 0}
	default:
		return NAN
	}
}

func NATIVE_TO_STRING_FUNCTION(args []Object) Object {
	if len(args) != 1 {
		return NewError("Invalid argument count. Expected 1, got %d instead", len(args))
	}
	return &String{Value: args[0].Inspect()}
}

func NATIVE_TIME_FUNCTION(args []Object) Object {
	if len(args) != 0 {
		return NewError("Invalid argument count. Expected 0, got %d instead", len(args))
	}
	t := float64(time.Now().UnixNano()) / 1e6 // milliseconds
	return &Number{Value: t}
}

func NATIVE_PROMPT_FUNCTION(args []Object) Object {
	if len(args) != 1 {
		return NewError("Invalid argument count. Expected 1, got %d instead", len(args))
	}
	message, ok := args[0].(*String)
	if !ok {
		return NewError("Cannot use prompt message type '%s' as prompt message", args[0].Type())
	}
	fmt.Print(message.Value)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return &String{Value: scanner.Text()}
	}
	return NewError("I/O error")
}

func NATIVE_RANDOM_FUNCTION(args []Object) Object {
	if len(args) != 0 {
		return NewError("Invalid argument count. Expected 0, got %d instead", len(args))
	}

	return &Number{Value: rand.Float64()}
}
