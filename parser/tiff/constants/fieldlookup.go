package constants

var FieldValueLookup = map[uint16]map[uint32]string{
	259: { // Compression
		1:     "Uncompressed",
		2:     "CCITT 1D",
		3:     "CCITT Group 3",
		4:     "CCITT Group 4",
		5:     "LZW",
		6:     "JPEG",
		32771: "Uncompressed (deprecated)",
		32773: "PackBits",
	},
	262: { // Photometric interpretation.
		0: "WhiteIsZero",
		1: "BlackIsZero",
		2: "RGB",
		3: "RGB Palette",
		4: "Transparency Mask",
		5: "CMYK",
		6: "YCbCr", // yes there is a gap, no I don't know why
		8: "CIELab",
	},
	284: { // PlanarConfiguration
		0: "Unknown",
		1: "Chunky",
		2: "Planar",
	},
	339 : { // SampleFormat
		0: "Unknown",
		1: "Unsigned Int",
		2: "Signed Int",
		3: "Float",
		4: "Undefined",
		5: "Complex Int",
		6: "Complex Float",
	},
	// TODO: more of these
}
