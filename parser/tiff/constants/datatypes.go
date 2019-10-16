package constants


var DataTypeNames = map[uint16]string {
	1: "byte",
	2: "ascii",
	3: "short",
	4: "long",
	5: "rational",
	6: "sbyte",
	7: "undefine",
	8: "sshort",
	9: "slong",
	10: "srational",
	11: "float",
	12: "double",
}

var DataTypeSize = map[uint16]uint32 {
	1: 1,
	2: 1,
	3: 2,
	4: 4,
	5: 8,
	6: 1,
	7: 1,
	8: 2,
	9: 4,
	10: 8,
	11: 4,
	12: 8,
}
