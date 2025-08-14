package shared

type Data struct {
	Num1, Num2 float64
}

type RPCInterface interface {
	Mul(args Data, res *float64) error
	Power(args Data, res *float64) error
}
