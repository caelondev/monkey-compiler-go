package code

type Tag byte

const (
	CONSTANT_NIL        Tag = 1
	CONSTANT_NUMBER     Tag = 2
	CONSTANT_BOOL_FALSE Tag = 3
	CONSTANT_BOOL_TRUE  Tag = 4
)
