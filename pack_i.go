package znets

type IPack interface {
	Input(string) int
	Pack([]byte) []byte
	UnPack([]byte) []byte
}
